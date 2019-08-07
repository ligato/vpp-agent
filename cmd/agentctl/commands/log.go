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

	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/vpp-agent/cmd/agentctl/utils"
)

func NewLogCommand(cli *AgentCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "log",
		Short: "Manage agent logging",
	}
	cmd.AddCommand(
		newLogListCommand(cli),
		newLogSetCommand(cli),
	)
	return cmd
}

func newLogListCommand(cli *AgentCli) *cobra.Command {
	opts := LogListOptions{}

	cmd := &cobra.Command{
		Use:     "list <logger>",
		Aliases: []string{"ls"},
		Short:   "Show vppagent logs",
		Long: `
A CLI tool to connect to vppagent and show vppagent logs.
`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.Logger = args[0]
			}
			return RunLogList(cli, opts)
		},
	}

	return cmd
}

type LogListOptions struct {
	Logger string
}

func RunLogList(cli *AgentCli, opts LogListOptions) error {
	resp, err := cli.HttpRestGET("/log/list")
	if err != nil {
		return fmt.Errorf("HTTP GET request failed: %v", err)
	}
	msg := string(resp)

	logging.Debugf("%s\n", msg)

	if strings.Compare(msg, "404 page not found") == 0 {
		fmt.Println(msg)
		return fmt.Errorf("not found")
	}

	data := utils.ConvertToLogList(msg)

	if len(data) == 0 {
		return fmt.Errorf("no data found")
	}

	if opts.Logger == "" {
		printLogList(data)
		return nil
	}

	tmpData := make(utils.LogList, 0)

	for _, value := range data {
		if strings.Contains(value.Logger, opts.Logger) {
			tmpData = append(tmpData, value)
		}
	}

	if len(tmpData) == 0 {
		return fmt.Errorf("No data found")
	}

	printLogList(tmpData)
	return nil
}

func printLogList(data utils.LogList) {
	buffer, err := data.PrintLogList()
	if err == nil {
		fmt.Fprintf(os.Stdout, buffer.String())
	} else {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
}

type LogSetOptions struct {
	Logger string
	Level  string
}

func newLogSetCommand(cli *AgentCli) *cobra.Command {
	opts := LogSetOptions{}
	cmd := &cobra.Command{
		Use:   "set <logger> <debug|info|warning|error|fatal|panic>",
		Short: "Set vppagent logger type",
		Long: `
A CLI tool to connect to vppagent and set vppagent logger type.
`,
		Args: cobra.RangeArgs(2, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Logger = args[0]
			opts.Level = args[1]
			return RunLogSet(cli, opts)
		},
	}
	return cmd
}

func RunLogSet(cli *AgentCli, opts LogSetOptions) error {
	resp, err := cli.HttpRestPUT("/log/"+opts.Logger+"/"+opts.Level, nil)
	if err != nil {
		return fmt.Errorf("HTTP PUT request failed: %v", err)
	}

	logging.Debugf("response: %s\n", resp)
	return nil
}
