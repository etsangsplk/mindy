package main

import (
	"fmt"
	"os"
	"time"

	"github.com/eris-ltd/mindy/Godeps/_workspace/src/github.com/spf13/cobra"
)

func cliListNames(cmd *cobra.Command, args []string) {
	dnsEntries, err := getDNSRecords()
	ifExit(err)
	// tendermint/wire can't handl maps so
	// write them one at a time
	var outputs []string
	for _, d := range dnsEntries {
		s, err := formatOutput(args, 1, d)
		ifExit(err)
		outputs = append(outputs, s)
	}
	for _, d := range outputs {
		fmt.Println(d)
	}
}

func cliCatchup(cobraCmd *cobra.Command, args []string) {
	ifExit(os.Chdir(DefaultTinyDNSDir))

	fetchAndUpdateRecords()
}

func cliRun(cmd *cobra.Command, args []string) {
	err := os.Chdir(DefaultTinyDNSDir)
	ifExit(err)

	fetchAndUpdateRecords()

	ticker := time.Tick(time.Second * time.Duration(updateEveryFlag))
	for {
		select {
		case <-ticker:
			fetchAndUpdateRecords()
		}
	}
}
