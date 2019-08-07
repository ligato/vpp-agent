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
	"bytes"
	"fmt"
	"os"
	"path"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func NewModelCommand(cli *AgentCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "model",
		Short: "Manage registered models",
	}
	cmd.AddCommand(
		newModelListCommand(cli),
	)
	return cmd
}

func newModelListCommand(cli *AgentCli) *cobra.Command {
	var opts ModelsListOptions

	cmd := &cobra.Command{
		Use:     "ls",
		Aliases: []string{"list"},
		Short:   "List models",
		Args:    cobra.ArbitraryArgs,
		Run: func(cmd *cobra.Command, args []string) {
			opts.Refs = args
			runModelList(cli, opts)
		},
	}
	cmd.Flags().BoolVar(&opts.NoTrunc, "no-trunc", false, "Disable truncing output")
	return cmd
}

type ModelsListOptions struct {
	Refs    []string
	NoTrunc bool
}

func runModelList(cli *AgentCli, opts ModelsListOptions) {
	models := cli.AllModels()

	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)

	fmt.Fprintf(w, "MODEL\tVERSION\tPROTOBUF\tKEY PREFIX\tNAME TEMPLATE\t\n")
	for _, model := range models {
		if !matchAnyRef(model, opts.Refs) {
			continue
		}

		nameTemplate := model.NameTemplate
		if !opts.NoTrunc && len(nameTemplate) > 51 {
			nameTemplate = fmt.Sprintf("%s…", nameTemplate[:50])
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t\n",
			model.Name, model.Version, model.ProtoName, model.KeyPrefix, nameTemplate)
	}
	if err := w.Flush(); err != nil {
		return
	}
	fmt.Fprint(os.Stdout, buf.String())
}

func matchAnyRef(model modelDetail, refs []string) bool {
	if len(refs) == 0 {
		return true
	}
	for _, ref := range refs {
		if ok, _ := path.Match(ref, model.Name); ok {
			return ok
		}
	}
	return false
}
