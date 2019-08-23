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
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func NewModelCommand(cli *AgentCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "model",
		Short: "Manage known models",
	}
	cmd.AddCommand(
		newModelListCommand(cli),
		newModelInspectCommand(cli),
	)
	return cmd
}

func newModelInspectCommand(cli *AgentCli) *cobra.Command {
	var opts ModelInspectOptions

	cmd := &cobra.Command{
		Use:     "inspect MODEL [MODEL...]",
		Aliases: []string{"i"},
		Short:   "Display detailed information on one or more models",
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			opts.Names = args
			runModelInspect(cli, opts)
		},
		DisableFlagsInUseLine: true,
	}
	// TODO: add support for custom formatting instead of json
	//cmd.Flags().StringVar(&opts.Format, "format", "", "Format for the output")
	return cmd
}

type ModelInspectOptions struct {
	Names  []string
	Format string
}

func runModelInspect(cli *AgentCli, opts ModelInspectOptions) {
	models := filterModelsByPrefix(cli.AllModels(), opts.Names)

	Debugf("models: %+v", models)

	b, err := json.MarshalIndent(models, "", "  ")
	if err != nil {
		ExitWithError(fmt.Errorf("Encoding data failed: %v", err))
	}

	fmt.Fprintf(os.Stdout, "%s\n", b)
}

func newModelListCommand(cli *AgentCli) *cobra.Command {
	var opts ModelListOptions

	cmd := &cobra.Command{
		Use:     "ls [PATTERN]",
		Aliases: []string{"list", "l"},
		Short:   "List models",
		Args:    cobra.ArbitraryArgs,
		Run: func(cmd *cobra.Command, args []string) {
			opts.Refs = args
			runModelList(cli, opts)
		},
		DisableFlagsInUseLine: true,
	}
	return cmd
}

type ModelListOptions struct {
	Refs    []string
	NoTrunc bool
}

func runModelList(cli *AgentCli, opts ModelListOptions) {
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "MODEL\tKEY PREFIX\tPROTO NAME\t\n")

	models := filterModelsByRefs(cli.AllModels(), opts.Refs)

	for _, model := range models {
		fmt.Fprintf(w, "%s\t%s\t%s\t\n",
			model.Name, model.KeyPrefix, model.ProtoName)
	}
	if err := w.Flush(); err != nil {
		return
	}
	fmt.Fprint(os.Stdout, buf.String())
}

func filterModelsByPrefix(models []ModelDetail, prefixes []string) []ModelDetail {
	if len(prefixes) == 0 {
		return models
	}
	var filtered []ModelDetail
	for _, pref := range prefixes {
		var model ModelDetail
		for _, m := range models {
			if !strings.HasPrefix(m.Name, pref) {
				continue
			}
			if model.Name != "" {
				ExitWithError(fmt.Errorf("Multiple models found with provided prefix: %s", pref))
				return nil
			}
			model = m
		}
		if model.Name == "" {
			ExitWithError(fmt.Errorf("No model found for provided prefix: %s", pref))
			return nil
		}
		filtered = append(filtered, model)
	}
	return filtered
}

func filterModelsByRefs(models []ModelDetail, refs []string) []ModelDetail {
	var filtered []ModelDetail
	for _, model := range models {
		if !matchAnyRef(model, refs) {
			continue
		}
		filtered = append(filtered, model)
	}
	return filtered
}

func matchAnyRef(model ModelDetail, refs []string) bool {
	if len(refs) == 0 {
		return true
	}
	for _, ref := range refs {
		if ok, _ := path.Match(ref, model.Name); ok {
			return true
		}
	}
	return false
}
