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

	"github.com/ligato/vpp-agent/cmd/agentctl/cli"
)

func NewVppcliCommand(cli *cli.AgentCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vppcli",
		Short: "Execute VPP CLI command",
		Long: `
A CLI tool to connect to vppagent and run VPP CLI command.
`,
		Example: `Run a VPP CLI command:
  $ agentctl vppcli show version

Do same as above, but specify the HTTP address of the agent:
  $ agentctl --httpaddr 172.17.0.3:9191 vppcli show version
`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			vppcmd := strings.Join(args, " ")
			return runVppcli(cli, vppcmd)
		},
	}
	return cmd
}

func runVppcli(cli *cli.AgentCli, vppcmd string) error {
	fmt.Fprintf(os.Stdout, "vpp# %s\n", vppcmd)

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
