package main

import (
	"fmt"
	"os"
	"path"

	"github.com/eris-ltd/mindy/Godeps/_workspace/src/github.com/spf13/cobra"
	cclient "github.com/eris-ltd/mindy/Godeps/_workspace/src/github.com/tendermint/tendermint/rpc/core_client"
)

var (
	DefaultNodeRPCHost = "mint" // expect to be docker linked
	DefaultNodeRPCPort = "46657"
	DefaultNodeRPCAddr = "http://" + DefaultNodeRPCHost + ":" + DefaultNodeRPCPort

	DefaultChainID string

	DefaultTinyDNSDir  = "/etc/service/tinydns/root"
	DefaultTinyDNSData = path.Join(DefaultTinyDNSDir, "data")

	REQUEST_TYPE = "JSONRPC"
	client       cclient.Client
)

func init() {
	if nodeRPCHost := os.Getenv("NODE_RPC_HOST"); nodeRPCHost != "" {
		DefaultNodeRPCHost = nodeRPCHost
		DefaultNodeRPCAddr = "http://" + DefaultNodeRPCHost + ":" + DefaultNodeRPCPort
	}
}

// flags
var (
	nodeAddrFlag        string
	tinydnsDataFileFlag string
	updateEveryFlag     int
)

func main() {
	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Display the mindy version",
		Run: func(c *cobra.Command, args []string) {
			fmt.Println("0.0.1")
		},
	}

	var listNamesCmd = &cobra.Command{
		Use:   "list",
		Short: "List all dns entries in the name reg",
		Long:  "",
		Run:   cliListNames,
	}

	var catchupCmd = &cobra.Command{
		Use:   "catchup",
		Short: "Add each dns entry from the name reg to the tinydns data file. Expects tinydns to be installed.",
		Long:  "",
		Run:   cliCatchup,
	}
	catchupCmd.Flags().StringVarP(&tinydnsDataFileFlag, "tinydns-data", "d", DefaultTinyDNSData, "path to tinydns data file")

	var runCmd = &cobra.Command{
		Use:   "run",
		Short: "Listen for NameTxs and update accordingly. Expects tinydns to be installed.",
		Long:  "",
		Run:   cliRun,
	}
	runCmd.Flags().StringVarP(&tinydnsDataFileFlag, "tinydns-data", "d", DefaultTinyDNSData, "path to tinydns data file")
	runCmd.Flags().IntVarP(&updateEveryFlag, "update-every", "u", 60, "number of seconds to wait before updating from chain")

	var rootCmd = &cobra.Command{Use: "mindy"}
	rootCmd.PersistentFlags().StringVarP(&nodeAddrFlag, "node-addr", "a", DefaultNodeRPCAddr, "full address of rpc host")
	rootCmd.PersistentPreRun = before
	rootCmd.AddCommand(listNamesCmd, catchupCmd, runCmd, versionCmd)
	rootCmd.Execute()
}

func before(cmd *cobra.Command, args []string) {
	client = cclient.NewClient(nodeAddrFlag, REQUEST_TYPE)
}

func exit(err error) {
	fmt.Println(err)
	os.Exit(1)
}

func ifExit(err error) {
	if err != nil {
		exit(err)
	}
}
