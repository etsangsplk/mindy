package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"bytes"

	"github.com/eris-ltd/mindy/Godeps/_workspace/src/github.com/tendermint/tendermint/types"
)

// there may be multiple records for a single domain name
type TinyDNSData map[string][]*ResourceRecord

// the json we expect to see in blockchain's namereg
type ResourceRecord struct {
	Type    string `json:"type"`
	FQDN    string `json:"fqdn"`
	Address string `json:"address"`
}

func (rr *ResourceRecord) Equals (r *ResourceRecord) bool{
	return rr.Type == r.Type && rr.FQDN == r.FQDN && rr.Address == r.Address
}

// TODO: something better
func RecordsEqual (r1 []*ResourceRecord, r2 []*ResourceRecord) bool{
	
	for _, rr1 := range r1{
		eq := false
		for _, rr2 := range r2{
			if rr1.Equals(rr2){
				eq = true
				break
			}
		}
		if !eq{
			return false
		}
	}
	return true
}

// read tinydns data from file
func TinyDNSDataFromFile(file string) (TinyDNSData, error) {
	// read tinydns file
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	tinydnsData := make(map[string][]*ResourceRecord)
	dataLines := strings.Split(string(b), "\n")
	for _, line := range dataLines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		first := string(line[0])
		fields := strings.Split(line[1:], ":")
		name, ip := fields[0], fields[1]
		var typ string
		switch first {
		case ".":
			typ = "NS"
		case "=", "+":
			typ = "A"
		case "#":
			continue
		default:
			return nil, fmt.Errorf("Unknown first character in tinydns data: %s", first)
		}
		newRecord := &ResourceRecord{typ, name, ip}
		rrs, ok := tinydnsData[name]
		if ok {
			rrs = append(rrs, newRecord)
		} else {
			rrs = []*ResourceRecord{newRecord}
		}
		tinydnsData[name] = rrs
	}
	return tinydnsData, nil
}

// __deprecated__.
func validateDNSEntrySimple(entry *types.NameRegEntry) error {
	spl := strings.Split(entry.Name, ".")
	if len(spl) < 3 {
		return fmt.Errorf("A valid name must have at least a subdomain, host name, and tld")
	}
	spl = strings.Split(entry.Data, ".")
	if len(spl) != 4 {
		return fmt.Errorf("Data must be a valid ipv4 address")
	}
	return nil
}

// an entry is valid if it is a json encoded single or list of ResourceRecords
func validateDNSEntryRR(entry *types.NameRegEntry) ([]*ResourceRecord, error) {
	spl := strings.Split(entry.Name, ".")
	if len(spl) < 2 {
		return nil, fmt.Errorf("A valid name must have at least a host name, and tld")
	}

	/*
		// data should be a jsonEncoded(jsonEncoded(ResourceRecord))
		var jsonString string
		if err := json.Unmarshal([]byte(entry.Data), &jsonString); err != nil{
			return nil, err
		}*/

	rrl := []*ResourceRecord{}
	if err := json.Unmarshal([]byte(entry.Data), &rrl); err != nil {
		rr := new(ResourceRecord)
		if err2 := json.Unmarshal([]byte(entry.Data), rr); err2 != nil{
			return nil, err
		}
		rrl = []*ResourceRecord{rr}
	}

	for _, rr := range rrl{
		spl = strings.Split(rr.Address, ".")
		if len(spl) != 4 {
			return nil, fmt.Errorf("Address must be a valid ipv4 address. Got %s", rr.Address)
		}
	}
	return rrl, nil
}

// grab all dns records from the blockchain
func getDNSRecords() (map[string][]*ResourceRecord, error) {
	r, err := client.ListNames()
	if err != nil {
		return nil, err
	}
	dnsEntries := make(map[string][]*ResourceRecord)
	for _, entry := range r.Names {
		if rrl, err := validateDNSEntryRR(entry); err == nil {
			dnsEntries[entry.Name] = rrl
		} else {
			fmt.Println("... invalid dns entry", entry.Name, entry.Data, err)
		}
	}
	return dnsEntries, nil
}

// grab all records from chain and re-write the data file
// this is the simplest approach until we index the data file
// and subscribe to chain events
func fetchAndUpdateRecords() {
	// get all dns entries from chain
	dnsRecords, err := getDNSRecords()
	if err != nil {
		fmt.Println("Error getting dns entries", err)
		return
	}

	buf := new(bytes.Buffer)

	for _, chainRecord := range dnsRecords {
		addedHost := false
		// a single name may have multiple records
		for _, record := range chainRecord{
			name, addr := record.FQDN, record.Address
			switch record.Type {
			case "NS":
				fmt.Printf("adding NS record for %s:%s\n", name, addr)
				buf.WriteString(fmt.Sprintf(".%s:%s:86400\n", name, addr))
			case "A":
				if !addedHost{
					fmt.Printf("adding A record for %s:%s\n", name, addr)
					buf.WriteString(fmt.Sprintf("=%s:%s:86400\n", name, addr))
					addedHost = true
				} else {
					fmt.Printf("adding A record for %s:%s\n", name, addr)
					buf.WriteString(fmt.Sprintf("+%s:%s:86400\n", name, addr))
				}
			default:
				fmt.Println("Found Resource Record with unknown type", record.Type)
				continue
			}

		}
	}

	if err = ioutil.WriteFile("data", buf.Bytes(), 0644); err != nil{
		fmt.Println("Error writing data file")
	}

	// done adding entries. commit them
	if err = makeTinyDNSRecords(); err != nil {
		fmt.Println("Error rebuilding data.cdb", err)
	}
}

//------------------------------------------------------------
// tinydns commands

// add an A record (tinydns only lets you have one official, and then many aliases)
func addTinyDNSARecord(fqdn, addr string) error {
	fmt.Println("Running add host", fqdn, addr, "...")
	cmd := exec.Command("./add-host", fqdn, addr)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Println("\t ... running add-alias")
		cmd := exec.Command("./add-alias", fqdn, addr)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	return nil
}

// add a name-server record
func addTinyDNSNSRecord(fqdn, addr string) error {
	fmt.Println("Running add ns", fqdn, addr, "...")
	cmd := exec.Command("./add-ns", fqdn, addr)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// compile the tinydns data to binary
func makeTinyDNSRecords() error {
	cmd := exec.Command("make")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
