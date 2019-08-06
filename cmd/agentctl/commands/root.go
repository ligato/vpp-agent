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

	"github.com/ligato/cn-infra/agent"
	"github.com/spf13/cobra"

	"github.com/ligato/vpp-agent/cmd/agentctl/cli"
)

var (
	// globalFlags defines all global flags.
	globalFlags struct {
		AgentAddr    string
		GrpcAddr     string
		HttpAddr     string
		ServiceLabel string
		Endpoints    []string

		Debug bool
	}
	agentLabel string
	agentAddr  = "127.0.0.1"
)

func init() {
	if l := os.Getenv("MICROSERVICE_LABEL"); l != "" {
		agentLabel = l
	}
	if a := os.Getenv("AGENT_ADDR"); a != "" {
		agentAddr = a
	}
}

// NewAgentctlCommand returns new root command.
func NewAgentctlCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "agentctl",
		Short:   "agentctl manages vpp-agent instances",
		Version: fmt.Sprintf("%s", agent.BuildVersion),
	}

	// define global flags
	flags := cmd.PersistentFlags()

	flags.StringVar(&globalFlags.AgentAddr, "addr", agentAddr, "Address on which agent is reachable")
	flags.StringVar(&globalFlags.GrpcAddr, "grpcaddr", agentAddr+":9111", "gRPC server address")
	flags.StringVar(&globalFlags.HttpAddr, "httpaddr", agentAddr+":9191", "HTTP server address")
	flags.StringVar(&globalFlags.ServiceLabel, "label", agentLabel, "Service label for agent instance")
	flags.StringSliceVar(&globalFlags.Endpoints, "endpoints", nil, "Etcd endpoints to connect to")

	flags.BoolVarP(&globalFlags.Debug, "debug", "D", false, "Enable debug mode")

	cli := cli.NewAgentCli()
	cli.HttpAddr = globalFlags.HttpAddr

	addCommands(cli, cmd)

	return cmd
}

func addCommands(cli *cli.AgentCli, cmd *cobra.Command) {
	cmd.AddCommand(
		configCmd(),
		NewDumpCommand(cli),
		NewLogCommand(cli),
		showCmd(),
		generateCmd(),
		putCmd(),
		delCmd(),
		importCmd(),
		NewVppcliCommand(cli),
	)
}

func Debugf(f string, a ...interface{}) {
	if globalFlags.Debug {
		fmt.Fprintf(os.Stderr, "DEBUG: "+f, a...)
	}
}
