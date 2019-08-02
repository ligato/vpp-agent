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
)

const defaultLabel = "vpp1"

var (
	// globalFlags defines a single type to hold all cobra global flags.
	globalFlags struct {
		Endpoints []string
		Label     string
	}
)

// NewRootCmd returns new base command.
func NewRootCmd(name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   name,
		Short: "A CLI tool for managing agents",
		Example: `Specify the etcd to connect to and list all agents that it knows about:
 $ export ETCD_ENDPOINTS=172.17.0.1:2379

or with a command line flag:
 $ agentctl --endpoints 172.17.0.1:2379 show
`,
		Version: fmt.Sprintf("%s", agent.BuildVersion),
	}

	label := defaultLabel
	if l := os.Getenv("MICROSERVICE_LABEL"); l != "" {
		label = l
	}

	// define flags
	cmd.PersistentFlags().StringVarP(&globalFlags.Label, "label", "l", label,
		"Microservice label identiying agent instance")
	cmd.PersistentFlags().StringSliceVarP(&globalFlags.Endpoints, "endpoints", "e", nil,
		"Etcd endpoints to connect to (comma-separated)")

	addCommands(cmd)
	return cmd
}

func addCommands(cmd *cobra.Command) {
	cmd.AddCommand(
		showCmd(),
		generateCmd(),
		putCmd(),
		delCmd(),
		importCmd(),
		dumpCmd(),
		vppcliCmd(),
		logCmd(),
	)
}
