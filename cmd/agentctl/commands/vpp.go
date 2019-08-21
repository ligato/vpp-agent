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
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func NewVppCommand(cli *AgentCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vpp",
		Short: "Manage VPP instance",
	}
	cmd.AddCommand(
		newVppCliCommand(cli),
		newVppInfoCommand(cli),
	)
	return cmd
}

func newVppCliCommand(cli *AgentCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cli",
		Short: "Execute VPP CLI command",
		Example: `
 To run a VPP CLI command:
  $ agentctl vpp cli show version

 Do same as above, but specify the HTTP address of the agent:
  $ agentctl --httpaddr 172.17.0.3:9191 vpp cli show version
`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			vppcmd := strings.Join(args, " ")
			return runVppCli(cli, vppcmd)
		},
		SilenceUsage: true,
	}
	return cmd
}

func runVppCli(cli *AgentCli, vppcmd string) error {
	fmt.Fprintf(os.Stdout, "# %s\n", vppcmd)

	data := map[string]interface{}{
		"vppclicommand": vppcmd,
	}
	resp, err := cli.HttpRestPOST("/vpp/command", data)
	if err != nil {
		return fmt.Errorf("HTTP POST request failed: %v", err)
	}

	var reply string
	if err := json.Unmarshal(resp, &reply); err != nil {
		return fmt.Errorf("decoding reply failed: %v", err)
	}

	fmt.Fprintf(os.Stdout, "%s", reply)
	return nil
}

func newVppInfoCommand(cli *AgentCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "info",
		Aliases: []string{"i"},
		Short:   "Retrieve info about VPP",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVppInfo(cli)
		},
		SilenceUsage: true,
	}
	return cmd
}

func runVppInfo(cli *AgentCli) error {
	data := map[string]interface{}{
		"vppclicommand": "show version verbose",
	}
	resp, err := cli.HttpRestPOST("/vpp/command", data)
	if err != nil {
		return fmt.Errorf("HTTP POST request failed: %v", err)
	}

	var reply string
	if err := json.Unmarshal(resp, &reply); err != nil {
		return fmt.Errorf("decoding reply failed: %v", err)
	}

	fmt.Fprintf(os.Stdout, "%s", reply)
	return nil
}
