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
	"context"
	"fmt"
	"io"
	"path"
	"strings"
	"text/tabwriter"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"go.ligato.io/vpp-agent/v3/cmd/agentctl/api/types"
	agentcli "go.ligato.io/vpp-agent/v3/cmd/agentctl/cli"
)

func NewModelCommand(cli agentcli.Cli) *cobra.Command {
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

func newModelListCommand(cli agentcli.Cli) *cobra.Command {
	var opts ModelListOptions
	cmd := &cobra.Command{
		Use:     "ls [PATTERN]",
		Aliases: []string{"list", "l"},
		Short:   "List models",
		Args:    cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Refs = args
			return runModelList(cli, opts)
		},
	}
	flags := cmd.Flags()
	flags.StringVar(&opts.Class, "class", "", "Filter by model class")
	flags.StringVarP(&opts.Format, "format", "f", "", "Format output")
	return cmd
}

type ModelListOptions struct {
	Class  string
	Refs   []string
	Format string
}

func runModelList(cli agentcli.Cli, opts ModelListOptions) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	allModels, err := cli.Client().ModelList(ctx, types.ModelListOptions{
		Class: opts.Class,
	})
	if err != nil {
		return err
	}

	models := filterModelsByRefs(allModels, opts.Refs)

	format := opts.Format
	if len(format) == 0 {
		printModelTable(cli.Out(), models)
	} else {
		if err := formatAsTemplate(cli.Out(), format, models); err != nil {
			return err
		}
	}

	return nil
}

func printModelTable(out io.Writer, models []types.Model) {
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "MODEL\tCLASS\tPROTO MESSAGE\tKEY PREFIX\t\n")
	for _, model := range models {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t\n",
			model.Name, model.Class, model.ProtoName, model.KeyPrefix)
	}
	if err := w.Flush(); err != nil {
		panic(err)
	}

	fmt.Fprint(out, buf.String())
}

func filterModelsByPrefix(models []types.Model, prefixes []string) ([]types.Model, error) {
	if len(prefixes) == 0 {
		return models, nil
	}
	var filtered []types.Model
	for _, pref := range prefixes {
		var model types.Model
		for _, m := range models {
			if !strings.HasPrefix(m.Name, pref) {
				continue
			}
			if model.Name != "" {
				return nil, fmt.Errorf("multiple models found with provided prefix: %s", pref)
			}
			model = m
		}
		if model.Name == "" {
			return nil, fmt.Errorf("no model found for provided prefix: %s", pref)
		}
		filtered = append(filtered, model)
	}
	return filtered, nil
}

func filterModelsByRefs(models []types.Model, refs []string) []types.Model {
	var filtered []types.Model
	for _, model := range models {
		if !matchAnyRef(model, refs) {
			continue
		}
		filtered = append(filtered, model)
	}
	return filtered
}

func matchAnyRef(model types.Model, refs []string) bool {
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

func newModelInspectCommand(cli agentcli.Cli) *cobra.Command {
	var (
		opts ModelInspectOptions
	)
	cmd := &cobra.Command{
		Use:     "inspect MODEL [MODEL...]",
		Aliases: []string{"i"},
		Short:   "Display detailed information on one or more models",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Names = args
			return runModelInspect(cli, opts)
		},
	}
	cmd.Flags().StringVarP(&opts.Format, "format", "f", "", "Format for the output")
	return cmd
}

type ModelInspectOptions struct {
	Names  []string
	Format string
}

func runModelInspect(cli agentcli.Cli, opts ModelInspectOptions) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	allModels, err := cli.Client().ModelList(ctx, types.ModelListOptions{})
	if err != nil {
		return err
	}

	models, err := filterModelsByPrefix(allModels, opts.Names)
	if err != nil {
		return err
	}

	logrus.Debugf("models: %+v", models)

	format := opts.Format
	if len(format) == 0 {
		format = "json"
	}

	if err := formatAsTemplate(cli.Out(), format, models); err != nil {
		return err
	}

	return nil
}
