package cmd

import (
	"fmt"
	"strings"

	"github.com/ligato/vpp-agent/cmd/agentctl2/utils"

	"github.com/ligato/vpp-agent/cmd/agentctl2/restapi"
	"github.com/spf13/cobra"
)

// RootCmd represents the base command when called without any subcommands.
var log = &cobra.Command{
	Use:     "log",
	Aliases: []string{"l"},
	Short:   "Log",
	Long: `
	Listening log
`,
}

var logList = &cobra.Command{
	Use:   "list",
	Short: "List <logger>",
	Long: `
	Listening log
`,
	Args: cobra.MaximumNArgs(1),
	Run:  logFunction,
}

var logSet = &cobra.Command{
	Use:   "set",
	Short: "set <looger> <debug|info|warning|error|fatal|panic>",
	Long: `
	Set log
`,
	Args: cobra.RangeArgs(2, 2),
	Run:  setFunction,
}

var verbose bool

func init() {
	RootCmd.AddCommand(log)
	log.AddCommand(logList)
	log.AddCommand(logSet)
	logList.Flags().BoolVar(&verbose, "v", false, "verbose")
}

func logFunction(cmd *cobra.Command, args []string) {
	msg := restapi.GetLog(globalFlags.Endpoints, "/log/list")

	if verbose {
		fmt.Printf("%s\n", msg)
		return
	}

	data := utils.ConvertToLogList(msg)

	if 0 == len(data) {
		fmt.Print("No data found.\n")
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

	if 0 == len(tmpData) {
		fmt.Print("No data found.\n")
		return
	}

	printLogList(tmpData)
}

func printLogList(data utils.LogList) {
	buffer, err := data.PrintLogList()
	if nil == err {
		fmt.Print(buffer.String())
	} else {
		fmt.Printf("Error: %v\n", err)
	}
}

func setFunction(cmd *cobra.Command, args []string) {
	logger := args[0]
	level := args[1]

	restapi.SetLog(globalFlags.Endpoints, "/log/"+logger+"/"+level)
}
