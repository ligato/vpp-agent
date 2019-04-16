package cmd

import (
	"fmt"
	"os"

	"github.com/ligato/cn-infra/agent"

	"github.com/spf13/cobra"
)

// RootCmd represents the base command when called without any subcommands.
var showVersion = &cobra.Command{
	Use:     "version",
	Aliases: []string{"V"},
	Short:   "Show agent version",
	Long: `
	Show agent version
`,
	Run: versionFunc,
}

func init() {
	RootCmd.AddCommand(showVersion)
}

func versionFunc(cmd *cobra.Command, args []string) {
	fmt.Fprintf(os.Stdout, "agentctl version %s\n", agent.BuildVersion)
}
