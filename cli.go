package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
	"encoding/json"

	"github.com/eris-ltd/mint-client/Godeps/_workspace/src/github.com/spf13/cobra"
	"github.com/eris-ltd/mint-client/Godeps/_workspace/src/github.com/tendermint/tendermint/types"
)

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

func validateDNSEntryRR(entry *types.NameRegEntry) (*ResourceRecord, error) {
	spl := strings.Split(entry.Name, ".")
	if len(spl) < 2 {
		return nil, fmt.Errorf("A valid name must have at least a host name, and tld")
	}

	// data should be a jsonEncoded(jsonEncoded(ResourceRecord))
	/*var jsonString string
	if err := json.Unmarshal([]byte(entry.Data), &jsonString); err != nil{
		return nil, err
	}*/

	rr := new(ResourceRecord)
	if err := json.Unmarshal([]byte(entry.Data), rr); err != nil{
		return nil, err
	}
	
	spl = strings.Split(rr.Address, ".")
	if len(spl) != 4 {
		return nil, fmt.Errorf("Address must be a valid ipv4 address")
	}
	return rr, nil
}

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

func cliListNames(cmd *cobra.Command, args []string) {
	dnsEntries, err := getDNSRecords()
	ifExit(err)
	s, err := formatOutput(args, 1, dnsEntries)
	ifExit(err)
	fmt.Println(s)
}

func cliCatchup(cobraCmd *cobra.Command, args []string) {
	dnsEntries, err := getDNSRecords()
	ifExit(err)

	err = os.Chdir(DefaultTinyDNSDir)
	ifExit(err)

	// KISS
	// for each entry, try add-host
	// if fails, run add-alias
	// TODO: parse/update the data file manually?
	for _, entry := range dnsEntries {
		ifExit(addTinyDNSARecord(entry.FQDN, entry.Address))
	}

	// done adding entries. commit them
	ifExit(makeTinyDNSRecords())
}

func makeTinyDNSRecords() error {
	cmd := exec.Command("make")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

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


func addTinyDNSNSRecord(fqdn, addr string) error {
	fmt.Println("Running add ns", fqdn, addr, "...")
	cmd := exec.Command("./add-ns", fqdn, addr)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func cliRun(cmd *cobra.Command, args []string) {
	// parse the tinydns data
	dnsData, err := TinyDNSDataFromFile(tinydnsDataFileFlag)
	ifExit(err)

	fetchAndUpdateRecords(dnsData)

	ticker := time.Tick(time.Second * time.Duration(updateEveryFlag))
	for {
		select {
		case <-ticker:
			fetchAndUpdateRecords(dnsData)
		}
	}
}

func fetchAndUpdateRecords(dnsData TinyDNSData) {
	// get all dns entries from chain
	dnsRecords, err := getDNSRecords()
	if err != nil {
		fmt.Println("Error getting dns entries", err)
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
			switch rr.Type{
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
