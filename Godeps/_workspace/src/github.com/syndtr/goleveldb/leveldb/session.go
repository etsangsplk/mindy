// Copyright (c) 2012, Suryandaru Triandana <syndtr@gmail.com>
// All rights reserved.
//
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package leveldb

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/eris-ltd/mindy/Godeps/_workspace/src/github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/eris-ltd/mindy/Godeps/_workspace/src/github.com/syndtr/goleveldb/leveldb/journal"
	"github.com/eris-ltd/mindy/Godeps/_workspace/src/github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/eris-ltd/mindy/Godeps/_workspace/src/github.com/syndtr/goleveldb/leveldb/storage"
	"github.com/eris-ltd/mindy/Godeps/_workspace/src/github.com/syndtr/goleveldb/leveldb/util"
)

type ErrManifestCorrupted struct {
	Field  string
	Reason string
}

func (e *ErrManifestCorrupted) Error() string {
	return fmt.Sprintf("leveldb: manifest corrupted (field '%s'): %s", e.Field, e.Reason)
}

func newErrManifestCorrupted(f storage.File, field, reason string) error {
	return errors.NewErrCorrupted(f, &ErrManifestCorrupted{field, reason})
}

// session represent a persistent database session.
type session struct {
	// Need 64-bit alignment.
	stNextFileNum    uint64 // current unused file number
	stJournalNum     uint64 // current journal file number; need external synchronization
	stPrevJournalNum uint64 // prev journal file number; no longer used; for compatibility with older version of leveldb
	stSeqNum         uint64 // last mem compacted seq; need external synchronization
	stTempFileNum    uint64

	stor     storage.Storage
	storLock util.Releaser
	o        *cachedOptions
	icmp     *iComparer
	tops     *tOps

	manifest       *journal.Writer
	manifestWriter storage.Writer
	manifestFile   storage.File

	stCompPtrs []iKey   // compaction pointers; need external synchronization
	stVersion  *version // current version
	vmu        sync.Mutex
}

// Creates new initialized session instance.
func newSession(stor storage.Storage, o *opt.Options) (s *session, err error) {
	if stor == nil {
		return nil, os.ErrInvalid
	}
	storLock, err := stor.Lock()
	if err != nil {
		return
	}
	s = &session{
		stor:       stor,
		storLock:   storLock,
		stCompPtrs: make([]iKey, o.GetNumLevel()),
	}
	s.setOptions(o)
	s.tops = newTableOps(s)
	s.setVersion(newVersion(s))
	s.log("log@legend F·NumFile S·FileSize N·Entry C·BadEntry B·BadBlock Ke·KeyError D·DroppedEntry L·Level Q·SeqNum T·TimeElapsed")
	return
}

// Close session.
func (s *session) close() {
	s.tops.close()
	if s.manifest != nil {
		s.manifest.Close()
	}
	if s.manifestWriter != nil {
		s.manifestWriter.Close()
	}
	s.manifest = nil
	s.manifestWriter = nil
	s.manifestFile = nil
	s.stVersion = nil
}

// Release session lock.
func (s *session) release() {
	s.storLock.Release()
}

// Create a new database session; need external synchronization.
func (s *session) create() error {
	// create manifest
	return s.newManifest(nil, nil)
}

// Recover a database session; need external synchronization.
func (s *session) recover() (err error) {
	defer func() {
		if os.IsNotExist(err) {
			// Don't return os.ErrNotExist if the underlying storage contains
			// other files that belong to LevelDB. So the DB won't get trashed.
			if files, _ := s.stor.GetFiles(storage.TypeAll); len(files) > 0 {
				err = &errors.ErrCorrupted{File: &storage.FileInfo{Type: storage.TypeManifest}, Err: &errors.ErrMissingFiles{}}
			}
		}
	}()

	m, err := s.stor.GetManifest()
	if err != nil {
		return
	}

	reader, err := m.Open()
	if err != nil {
		return
	}
	defer reader.Close()

	var (
		// Options.
		numLevel = s.o.GetNumLevel()
		strict   = s.o.GetStrict(opt.StrictManifest)

		jr      = journal.NewReader(reader, dropper{s, m}, strict, true)
		rec     = &sessionRecord{}
		staging = s.stVersion.newStaging()
	)
	for {
		var r io.Reader
		r, err = jr.Next()
		if err != nil {
			if err == io.EOF {
				err = nil
				break
			}
			return errors.SetFile(err, m)
		}

		err = rec.decode(r, numLevel)
		if err == nil {
			// save compact pointers
			for _, r := range rec.compPtrs {
				s.stCompPtrs[r.level] = iKey(r.ikey)
			}
			// commit record to version staging
			staging.commit(rec)
		} else {
			err = errors.SetFile(err, m)
			if strict || !errors.IsCorrupted(err) {
				return
			} else {
				s.logf("manifest error: %v (skipped)", errors.SetFile(err, m))
			}
		}
		rec.resetCompPtrs()
		rec.resetAddedTables()
		rec.resetDeletedTables()
	}

	switch {
	case !rec.has(recComparer):
		return newErrManifestCorrupted(m, "comparer", "missing")
	case rec.comparer != s.icmp.uName():
		return newErrManifestCorrupted(m, "comparer", fmt.Sprintf("mismatch: want '%s', got '%s'", s.icmp.uName(), rec.comparer))
	case !rec.has(recNextFileNum):
		return newErrManifestCorrupted(m, "next-file-num", "missing")
	case !rec.has(recJournalNum):
		return newErrManifestCorrupted(m, "journal-file-num", "missing")
	case !rec.has(recSeqNum):
		return newErrManifestCorrupted(m, "seq-num", "missing")
	}

	s.manifestFile = m
	s.setVersion(staging.finish())
	s.setNextFileNum(rec.nextFileNum)
	s.recordCommited(rec)
	return nil
}

// Commit session; need external synchronization.
func (s *session) commit(r *sessionRecord) (err error) {
	v := s.version()
	defer v.release()

	// spawn new version based on current version
	nv := v.spawn(r)

	if s.manifest == nil {
		// manifest journal writer not yet created, create one
		err = s.newManifest(r, nv)
	} else {
		err = s.flushManifest(r)
	}

	// finally, apply new version if no error rise
	if err == nil {
		s.setVersion(nv)
	}

	return
}
