package main

import (
	"fmt"
	"time"

	"github.com/eris-ltd/mindy/Godeps/_workspace/src/github.com/spf13/cobra"
)

func cliListNames(cmd *cobra.Command, args []string) {
	dnsEntries, err := getDNSRecords()
	ifExit(err)
	s, err := formatOutput(args, 1, dnsEntries)
	ifExit(err)
	fmt.Println(s)
}

func cliCatchup(cobraCmd *cobra.Command, args []string) {
	// parse the tinydns data
	dnsData, err := TinyDNSDataFromFile(tinydnsDataFileFlag)
	ifExit(err)

	fetchAndUpdateRecords(dnsData)
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
