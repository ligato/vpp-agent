package main

import (
	"fmt"
	"github.com/ligato/vpp-agent/cmd/agentctl/cmd"
	"os"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
