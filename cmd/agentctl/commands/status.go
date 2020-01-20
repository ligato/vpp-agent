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

	"github.com/spf13/cobra"

	agentcli "go.ligato.io/vpp-agent/v3/cmd/agentctl/cli"
)

func NewStatusCommand(cli agentcli.Cli) *cobra.Command {
	var (
		opts StatusOptions
	)
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Retrieve agent status",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(cli, opts)
		},
	}
	flags := cmd.Flags()
	flags.StringVarP(&opts.Format, "format", "f", "", "Format output")
	return cmd
}

type StatusOptions struct {
	Format string
}

func runStatus(cli agentcli.Cli, opts StatusOptions) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	status, err := cli.Client().Status(ctx)
	if err != nil {
		return err
	}

	format := opts.Format
	if len(format) == 0 {
		format = defaultFormatStatus
	}

	if err := formatAsTemplate(cli.Out(), format, status); err != nil {
		return err
	}

	return nil
}

const defaultFormatStatus = `AGENT
       State: {{.AgentStatus.State}}
     Version: {{.AgentStatus.BuildVersion}}
     Started: {{epoch .AgentStatus.StartTime}} ({{ago (epoch .AgentStatus.StartTime)}} ago)
  
 Last change: {{ago (epoch .AgentStatus.LastChange)}}
 Last update: {{ago (epoch .AgentStatus.LastUpdate)}}

PLUGINS
{{- range $name, $plugin := .PluginStatus}}
   {{$name}}: {{$plugin.State}}
{{- end}}
`
