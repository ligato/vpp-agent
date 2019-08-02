package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ligato/vpp-agent/cmd/agentctl/restapi"
	"github.com/ligato/vpp-agent/cmd/agentctl/utils"
)

func logCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "log",
		Aliases: []string{"l"},
		Short:   "Show/Set vppagent logs",
		Long: `
A CLI tool to connect to vppagent and show/set vppagent logs.
Use the 'ETCD_ENDPOINTS'' environment variable or the 'endpoints'
flag in the command line to specify vppagent instances to
connect to.
`,
	}

	cmd.AddCommand(logList)
	cmd.AddCommand(logSet)

	return cmd
}

var logList = &cobra.Command{
	Use:   "list <logget>",
	Short: "Show vppagent logs",
	Long: `
A CLI tool to connect to vppagent and show vppagent logs.
Use the 'ETCD_ENDPOINTS'' environment variable or the 'endpoints'
flag in the command line to specify vppagent instances to
connect to.
`,
	Example: `Specify the vppagent to connect to and show vppagent logs:
	$ export ETCD_ENDPOINTS=172.17.0.3:9191
	$ ./agentctl log list

Do as above, but with a command line flag:
  $ ./agentctl --endpoints 172.17.0.3:9191 log list
`,

	Args: cobra.MaximumNArgs(1),
	Run:  logFunction,
}

var logSet = &cobra.Command{
	Use:   "set <logger> <debug|info|warning|error|fatal|panic>",
	Short: "Set vppagent logger type",
	Long: `
A CLI tool to connect to vppagent and set vppagent logger type.
Use the 'ETCD_ENDPOINTS'' environment variable or the 'endpoints'
flag in the command line to specify vppagent instances to
connect to.
`,
	Example: `Specify the vppagent to connect to and show vppagent logs:
	$ export ETCD_ENDPOINTS=172.17.0.3:9191
	$ ./agentctl log set agent info

Do as above, but with a command line flag:
  $ ./agentctl --endpoints 172.17.0.3:9191 log set agent info
`,

	Args: cobra.RangeArgs(2, 2),
	Run:  setFunction,
}

var verbose bool

func init() {
	logList.Flags().BoolVar(&verbose, "v", false, "verbose")
}

func logFunction(cmd *cobra.Command, args []string) {
	msg := restapi.GetMsg(globalFlags.Endpoints, "/log/list")

	if verbose {
		fmt.Fprintf(os.Stdout, "%s\n", msg)
		return
	}

	if strings.Compare(msg, "404 page not found") == 0 {
		fmt.Println(msg)
		return
	}

	data := utils.ConvertToLogList(msg)

	if 0 == len(data) {
		fmt.Fprintf(os.Stdout, "No data found.\n")
		return
	}

	if 1 != len(args) {
		printLogList(data)
		return
	}

	logger := args[0]

	tmpData := make(utils.LogList, 0)

	for _, value := range data {
		if strings.Contains(value.Logger, logger) {
			tmpData = append(tmpData, value)
		}
	}

	if len(tmpData) == 0 {
		fmt.Fprintf(os.Stdout, "No data found.\n")
		return
	}

	printLogList(tmpData)
}

func printLogList(data utils.LogList) {
	buffer, err := data.PrintLogList()
	if err == nil {
		fmt.Fprintf(os.Stdout, buffer.String())
		fmt.Printf("\n")
	} else {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
}

func setFunction(cmd *cobra.Command, args []string) {
	logger := args[0]
	level := args[1]

	restapi.SetMsg(globalFlags.Endpoints, "/log/"+logger+"/"+level)
}
