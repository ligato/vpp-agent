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

	"github.com/ligato/cn-infra/agent"
	"github.com/ligato/cn-infra/logging"
	"github.com/spf13/cobra"
)

var (
	// globalFlags defines all global flags.
	globalFlags struct {
		AgentAddr    string
		GrpcPort     string
		HttpPort     string
		ServiceLabel string
		Endpoints    []string

		Debug bool
	}
	agentLabel string
	agentAddr  = "127.0.0.1"
	endpoints  []string
)

func init() {
	if l := os.Getenv("MICROSERVICE_LABEL"); l != "" {
		agentLabel = l
	}
	if a := os.Getenv("AGENT_ADDR"); a != "" {
		agentAddr = a
	}
	if e := os.Getenv("ETCD_ENDPOINTS"); e != "" {
		endpoints = strings.Split(e, ",")
	}
}

// NewAgentctlCommand returns new root command.
func NewAgentctlCommand() *cobra.Command {
	cli := &AgentCli{}

	cmd := &cobra.Command{
		Use:     "agentctl",
		Short:   "agentctl manages vpp-agent instances",
		Version: fmt.Sprintf("%s", agent.BuildVersion),
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			cli.Initialize()
		},
	}

	// define global flags
	flags := cmd.PersistentFlags()
	flags.StringVar(&globalFlags.AgentAddr, "addr", agentAddr, "Address on which agent is reachable")
	flags.StringVar(&globalFlags.GrpcPort, "grpcport", "9111", "gRPC server port")
	flags.StringVar(&globalFlags.HttpPort, "httpport", "9191", "HTTP server port")
	flags.StringVar(&globalFlags.ServiceLabel, "label", agentLabel, "Service label for agent instance")
	flags.StringSliceVarP(&globalFlags.Endpoints, "endpoints", "e", endpoints, "Etcd endpoints to connect to")
	flags.BoolVarP(&globalFlags.Debug, "debug", "D", false, "Enable debug mode")

	addCommands(cmd, cli)

	return cmd
}

func addCommands(cmd *cobra.Command, cli *AgentCli) {
	cmd.AddCommand(
		NewDumpCommand(cli),
		NewLogCommand(cli),
		NewImportCommand(cli),
		NewVppcliCommand(cli),
		NewConfigCommand(cli),
		showCmd(),
		generateCmd(),
	)
}

func Debugf(f string, a ...interface{}) {
	if globalFlags.Debug || logging.DefaultLogger.GetLevel() >= logging.DebugLevel {
		if !strings.HasSuffix(f, "\n") {
			f = f + "\n"
		}
		fmt.Printf(f, a...)
	}
}
