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
	"text/tabwriter"

	"github.com/spf13/cobra"

	"go.ligato.io/vpp-agent/v3/cmd/agentctl/api/types"
	agentcli "go.ligato.io/vpp-agent/v3/cmd/agentctl/cli"
)

func NewMetricsCommand(cli agentcli.Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "metrics",
		Short: "Get runtime metrics",
	}
	cmd.AddCommand(
		newMetricsListCommand(cli),
		newMetricsGetCommand(cli),
	)
	return cmd
}

func newMetricsListCommand(cli agentcli.Cli) *cobra.Command {
	var opts MetricsListOptions

	cmd := &cobra.Command{
		Use:     "list [PATTERN]",
		Aliases: []string{"list", "l"},
		Short:   "List metrics",
		Args:    cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Refs = args
			return runMetricsList(cli, opts)
		},
		DisableFlagsInUseLine: true,
	}
	return cmd
}

type MetricsListOptions struct {
	Refs []string
}

func runMetricsList(cli agentcli.Cli, opts MetricsListOptions) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	allModels, err := cli.Client().ModelList(ctx, types.ModelListOptions{
		Class: "metrics",
	})
	if err != nil {
		return err
	}

	models := filterModelsByRefs(allModels, opts.Refs)

	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "METRIC\tPROTO MESSAGE\t\n")
	for _, model := range models {
		fmt.Fprintf(w, "%s\t%s\t\n", model.Name, model.ProtoName)
	}
	if err := w.Flush(); err != nil {
		return err
	}

	fmt.Fprint(cli.Out(), buf.String())
	return nil
}

func newMetricsGetCommand(cli agentcli.Cli) *cobra.Command {
	var opts MetricsGetOptions

	cmd := &cobra.Command{
		Use:   "get METRIC",
		Short: "Get metrics data",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Metrics = args
			return runMetricsGet(cli, opts)
		},
	}
	return cmd
}

type MetricsGetOptions struct {
	Metrics []string
}

func runMetricsGet(cli agentcli.Cli, opts MetricsGetOptions) error {
	metric := opts.Metrics[0]

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	data, err := cli.Client().GetMetricData(ctx, metric)
	if err != nil {
		return err
	}

	if err := formatAsTemplate(cli.Out(), "json", data); err != nil {
		return err
	}

	return nil
}
