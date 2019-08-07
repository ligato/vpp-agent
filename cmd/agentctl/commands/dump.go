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

	"github.com/spf13/cobra"
)

func NewDumpCommand(cli *AgentCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "dump",
		Aliases: []string{"d"},
		Short:   "Dump current state",
	}

	for _, model := range cli.AllModels() {
		c := &cobra.Command{
			Use: model.Alias,
			Aliases: []string{
				model.Name,
				model.ProtoName,
				model.KeyPrefix,
			},
			Short: fmt.Sprintf("Dump for %s model (%s)", model.Name, model.ProtoName),
			Args:  cobra.NoArgs,
			Run: func(cmd *cobra.Command, args []string) {
				runDump(cli, model)
			},
		}
		cmd.AddCommand(c)
	}

	return cmd
}

func runDump(cli *AgentCli, model modelDetail) {
	q := fmt.Sprintf(`key-prefix=%s&view=cached`, url.QueryEscape(model.KeyPrefix))
	resp, err := cli.HttpRestGET("/scheduler/dump?" + q)
	if err != nil {
		ExitWithError(err)
	}

	fmt.Fprintf(os.Stdout, "%s\n", resp)
}
