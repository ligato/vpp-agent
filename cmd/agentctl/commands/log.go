//  Copyright (c) 2019 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

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
