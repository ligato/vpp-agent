package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ligato/vpp-agent/cmd/agentctl/restapi"
)

func vppcliCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vppcli",
		Short: "Execute VPP CLI command",
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
		RunE: vppcliFunction,
	}
	return cmd
}

func vppcliFunction(cmd *cobra.Command, args []string) error {
	cli := strings.Join(args, " ")
	fmt.Fprintf(os.Stdout, "VPP CLI: %s\n", cli)

	data := map[string]interface{}{
		"vppclicommand": cli,
	}

	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	msg := string(b)

	resp := restapi.PostMsg(globalFlags.Endpoints, "/vpp/command", msg)

	tmp := strings.Replace(resp, "\\n", "\n", -1)
	fmt.Fprintf(os.Stdout, "%s\n", tmp)

	return nil
}
