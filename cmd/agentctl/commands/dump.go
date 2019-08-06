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
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/vpp-agent/cmd/agentctl/cli"
	models "github.com/ligato/vpp-agent/pkg/models"
)

func NewDumpCommand(cli *cli.AgentCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "dump",
		Aliases: []string{"d"},
		Short:   "Dump commands",
		Args:    cobra.MinimumNArgs(2),
	}

	for _, m := range models.RegisteredModels() {
		protoName := m.Info["protoName"]
		keyPrefix := m.Info["keyPrefix"]
		module := strings.Split(m.Model.Module, ".")
		typ := m.Model.Type

		use := fmt.Sprintf("%s.%s", module[0], typ)

		logging.Debugf("add dump command %q", use)

		c := &cobra.Command{
			Use:   use,
			Short: fmt.Sprintf("Dump %s", protoName),
			Args:  cobra.NoArgs,
			Run: func(cmd *cobra.Command, args []string) {
				query := fmt.Sprintf(`key-prefix=%s&view=cached`, url.QueryEscape(keyPrefix))
				resp, err := cli.HttpRestGET("/scheduler/dump?" + query)
				if err != nil {
					ExitWithError(err)
				}
				fmt.Fprintf(os.Stdout, "%s\n", resp)
			},
		}
		cmd.AddCommand(c)
	}

	return cmd
}
