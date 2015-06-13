package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/eris-ltd/mindy/Godeps/_workspace/src/github.com/tendermint/tendermint/types"
)

type TinyDNSData map[string]*ResourceRecord

// the json we expect to see in blockchain's namereg
type ResourceRecord struct {
	Type    string `json:"type"`
	FQDN    string `json:"fqdn"`
	Address string `json:"address"`
}

// read tinydns data from file
func TinyDNSDataFromFile(file string) (TinyDNSData, error) {
	// read tinydns file
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	tinydnsData := make(map[string]*ResourceRecord)
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
		tinydnsData[name] = &ResourceRecord{typ, name, ip}
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

// checks if entry is json ResourceRecord and has a valid address
func validateDNSEntryRR(entry *types.NameRegEntry) (*ResourceRecord, error) {
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

	rr := new(ResourceRecord)
	if err := json.Unmarshal([]byte(entry.Data), rr); err != nil {
		return nil, err
	}

	spl = strings.Split(rr.Address, ".")
	if len(spl) != 4 {
		return nil, fmt.Errorf("Address must be a valid ipv4 address")
	}
	return rr, nil
}

// grab all dns records from the blockchain
func getDNSRecords() ([]*ResourceRecord, error) {
	r, err := client.ListNames()
	if err != nil {
		return nil, err
	}
	dnsEntries := []*ResourceRecord{}
	for _, entry := range r.Names {
		if rr, err := validateDNSEntryRR(entry); err == nil {
			dnsEntries = append(dnsEntries, rr)
		} else {
			fmt.Println("... invalid dns entry", entry.Name, entry.Data, err)
		}
	}
	return dnsEntries, nil
}

// grab all records from chain and make any updates
func fetchAndUpdateRecords(dnsData TinyDNSData) {
	// get all dns entries from chain
	dnsRecords, err := getDNSRecords()
	if err != nil {
		fmt.Println("Error getting dns entries", err)
		return
	}

	anyUpdates := false
	for _, rr := range dnsRecords {
		name, addr := rr.FQDN, rr.Address
		record, ok := dnsData[name]

		toUpdate := true
		// if we have it and nothings changed, don't update
		if ok && record.FQDN == name && record.Address == addr {
			toUpdate = false
		}

		if toUpdate {
			anyUpdates = true
			switch rr.Type {
			case "NS":
				addTinyDNSNSRecord(name, addr)
			case "A":
				addTinyDNSARecord(name, addr)
			default:
				fmt.Println("Found Resource Record with unknown type", rr.Type)
				continue
			}
			dnsData[name] = rr
		}
	}

	if anyUpdates {
		// done adding entries. commit them
		if err = makeTinyDNSRecords(); err != nil {
			fmt.Println("Error rebuilding data.cdb", err)
		}
	} else {
		fmt.Println("No new updates")
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
