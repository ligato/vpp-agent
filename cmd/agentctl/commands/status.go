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
	"go.ligato.io/cn-infra/v2/health/probe"

	"go.ligato.io/vpp-agent/v3/cmd/agentctl/api/types"
	agentcli "go.ligato.io/vpp-agent/v3/cmd/agentctl/cli"
)

func NewStatusCommand(cli agentcli.Cli) *cobra.Command {
	var (
		opts StatusOptions
	)
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Retrieve agent status and version info",
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

	s, err := cli.Client().Status(ctx)
	if err != nil {
		return err
	}
	v, err := cli.Client().AgentVersion(ctx)
	if err != nil {
		return err
	}

	format := opts.Format
	if len(format) == 0 {
		format = defaultFormatStatus
	}

	data := struct {
		Status  *probe.ExposedStatus
		Version *types.Version
	}{s, v}

	if err := formatAsTemplate(cli.Out(), format, data); err != nil {
		return err
	}

	return nil
}

const defaultFormatStatus = `AGENT
    App name:    {{.Version.App}}
    Version:     {{.Version.Version}}

    State:       {{.Status.AgentStatus.State}}
    Started:     {{epoch .Status.AgentStatus.StartTime}} ({{ago (epoch .Status.AgentStatus.StartTime)}} ago)
    Last change: {{ago (epoch .Status.AgentStatus.LastChange)}}
    Last update: {{ago (epoch .Status.AgentStatus.LastUpdate)}}

    Go version:  {{.Version.GoVersion}}
    OS/Arch:     {{.Version.OS}}/{{.Version.Arch}}

    Build Info:
        Git commit: {{.Version.GitCommit}}
        Git branch: {{.Version.GitBranch}}
        User:       {{.Version.BuildUser}}
        Host:       {{.Version.BuildHost}}
        Built:      {{epoch .Version.BuildTime}}

PLUGINS
{{- range $name, $plugin := .Status.PluginStatus}}
    {{$name}}: {{$plugin.State}}
{{- end}}
`
