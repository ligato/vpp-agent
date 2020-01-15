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
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	agentcli "go.ligato.io/vpp-agent/v3/cmd/agentctl/cli"
)

func NewVppCommand(cli agentcli.Cli) *cobra.Command {
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

func newVppCliCommand(cli agentcli.Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cli",
		Aliases: []string{"c"},
		Short:   "Execute VPP CLI command",
		Example: `
 To run a VPP CLI command:
  $ {{.CommandPath}} vpp cli show version
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

func runVppCli(cli agentcli.Cli, vppcmd string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fmt.Fprintf(cli.Out(), "vpp# %s\n", vppcmd)

	reply, err := cli.Client().VppRunCli(ctx, vppcmd)
	if err != nil {
		return err
	}

	fmt.Fprintf(cli.Out(), "%s", reply)
	return nil
}

func newVppInfoCommand(cli agentcli.Cli) *cobra.Command {
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

func runVppInfo(cli agentcli.Cli) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	version, err := cli.Client().VppRunCli(ctx, "show version verbose")
	if err != nil {
		return err
	}
	fmt.Fprintf(cli.Out(), "VERSION:\n%s\n", version)

	config, err := cli.Client().VppRunCli(ctx, "show version cmdline")
	if err != nil {
		return err
	}
	fmt.Fprintf(cli.Out(), "CONFIG:\n%s\n", config)

	return nil
}
