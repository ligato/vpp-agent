package main

import (
	"os"

	"github.com/ligato/vpp-agent/cmd/agentctl/commands"
)

func main() {
	rootCmd := commands.NewRootCmd("agentctl")
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
