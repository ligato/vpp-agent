package cmd

import (
	"fmt"
	"os"

	"github.com/ligato/vpp-agent/cmd/agentctl2/restapi"
	"github.com/spf13/cobra"
)

// RootCmd represents the base command when called without any subcommands.
var cliConfig = &cobra.Command{
	Use:   "vppcli",
	Short: "CLI command for VPP",
	Long: `
	Run CLI command for VPP
`,
	Args: cobra.MinimumNArgs(1),
	Run:  cliFunction,
}

func init() {
	RootCmd.AddCommand(cliConfig)
}

func cliFunction(cmd *cobra.Command, args []string) {
	var cli string

	for _, str := range args {
		cli = cli + " " + str
	}

	msg := fmt.Sprintf("{\"vppclicommand\":\"%v\"}", cli)

	fmt.Fprintf(os.Stdout, "%s\n", msg)

	resp := restapi.PostMsg(globalFlags.Endpoints, "/vpp/command", msg)

	//TODO: Need format
	fmt.Fprintf(os.Stdout, "%s\n", resp)
}
