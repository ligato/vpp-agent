// Copyright (c) 2017 Cisco and/or its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/common-nighthawk/go-figure"
	"github.com/ligato/cn-infra/agent"
	"github.com/ligato/cn-infra/logging"
	"github.com/spf13/cobra"
)

// RootName defines default name used for root command
var RootName = "agentctl"

var global struct {
	AgentHost    string
	PortGRPC     string
	PortHTTP     string
	ServiceLabel string
	Endpoints    []string

	Debug bool
}

// NewRootCommand returns new root command.
func NewRootCommand(cli *AgentCli) *cobra.Command {
	return newRootCommand(cli, RootName)
}

func newRootCommand(cli *AgentCli, name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     fmt.Sprintf("agentctl [OPTIONS] "),
		Short:   fmt.Sprintf("%s manages Ligato agents", name),
		Long:    figure.NewFigure(name, "", false).String(),
		Version: fmt.Sprintf("%s (%s)", agent.BuildVersion, agent.CommitHash),
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			cli.Init()
			Debugf("running command: %s\n\n", cmd.CommandPath())
		},
		DisableFlagsInUseLine: true,
		TraverseChildren:      true,
	}
	cmd.SetUsageTemplate(usageTemplate)
	AddFlags(cmd)
	AddCommands(cmd, cli)
	return cmd
}

func AddFlags(cmd *cobra.Command) {
	var (
		serviceLabel  = os.Getenv("MICROSERVICE_LABEL")
		agentHost     = os.Getenv("AGENT_HOST")
		etcdEndpoints = strings.Split(os.Getenv("ETCD_ENDPOINTS"), ",")
	)
	flags := cmd.PersistentFlags()
	flags.StringVarP(&global.AgentHost, "host", "H", agentHost, "Address on which agent is reachable")
	flags.StringVar(&global.PortGRPC, "grpc-port", "9111", "gRPC server port")
	flags.StringVar(&global.PortHTTP, "http-port", "9191", "HTTP server port")
	flags.StringVarP(&global.ServiceLabel, "service-label", "l", serviceLabel, "Service label for agent instance")
	flags.StringSliceVarP(&global.Endpoints, "etcd-endpoints", "e", etcdEndpoints, "Etcd endpoints to connect to")
	flags.BoolVarP(&global.Debug, "debug", "D", false, "Enable debug mode")
}

func AddCommands(cmd *cobra.Command, cli *AgentCli) {
	cmd.AddCommand(
		NewModelCommand(cli),
		NewLogCommand(cli),
		NewImportCommand(cli),
		NewVppCommand(cli),
		NewDumpCommand(cli),
		NewKvdbCommand(cli),
		NewGenerateCommand(cli),
		showCmd(),
	)
}

func Debugf(f string, a ...interface{}) {
	if global.Debug || logging.DefaultLogger.GetLevel() >= logging.DebugLevel {
		if !strings.HasSuffix(f, "\n") {
			f = f + "\n"
		}
		fmt.Printf("[DEBUG] "+f, a...)
	}
}

var usageTemplate = `Usage:

{{- if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}}{{- if .HasAvailableLocalFlags}} [OPTIONS]{{end}} COMMAND{{end}}

{{- if gt .Aliases 0}}

Aliases:
  {{.NameAndAliases}}

{{- end}}
{{- if .HasExample}}

Examples:
{{.Example}}

{{- end}}
{{- if .HasAvailableSubCommands}}

Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Options:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Run "{{.CommandPath}} COMMAND --help" for more information about a command.{{end}}
`
