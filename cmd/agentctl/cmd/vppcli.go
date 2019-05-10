package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/ligato/vpp-agent/cmd/agentctl/restapi"
	"github.com/spf13/cobra"
)

// RootCmd represents the base command when called without any subcommands.
var cliConfig = &cobra.Command{
	Use:   "vppcli",
	Short: "CLI command for vppagent",
	Long: `
A CLI tool to connect to vppagent and run VPP CLI command.
Use the 'ETCD_ENDPOINTS'' environment variable or the 'endpoints'
flag in the command line to specify vppagent instances to
connect to.
`,
	Example: `Specify the vppagent to connect to and run VPP CLI command:
	$ export ETCD_ENDPOINTS=172.17.0.3:9191
	$ ./agentctl vppcli 'show int'

Do as above, but with a command line flag:
  $ ./agentctl --endpoints 172.17.0.3:9191 vppcli 'show int'
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

	tmp := strings.Replace(resp, "\\n", "\n", -1)
	fmt.Fprintf(os.Stdout, "%s\n", tmp)
}
