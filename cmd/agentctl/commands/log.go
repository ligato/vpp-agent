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
	"context"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"go.ligato.io/vpp-agent/v3/cmd/agentctl/api/types"
	agentcli "go.ligato.io/vpp-agent/v3/cmd/agentctl/cli"
)

func NewLogCommand(cli agentcli.Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "log",
		Short: "Manage agent logging",
	}
	cmd.AddCommand(
		newLogListCommand(cli),
		newLogSetCommand(cli),
	)
	return cmd
}

func newLogListCommand(cli agentcli.Cli) *cobra.Command {
	var (
		opts LogListOptions
	)
	cmd := &cobra.Command{
		Use:     "list [logger]",
		Aliases: []string{"ls"},
		Short:   "List agent loggers",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.Name = args[0]
			}
			return RunLogList(cli, opts)
		},
	}
	return cmd
}

type LogListOptions struct {
	Name string
}

func RunLogList(cli agentcli.Cli, opts LogListOptions) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	loggers, err := cli.Client().LoggerList(ctx)
	if err != nil {
		return err
	}

	if len(loggers) == 0 {
		return fmt.Errorf("no logger found")
	}

	var filtered []types.Logger
	for _, value := range loggers {
		if opts.Name == "" || strings.Contains(value.Logger, opts.Name) {
			filtered = append(filtered, value)
		}
	}
	sort.Sort(sortedLoggers(filtered))
	printLoggerList(cli.Out(), filtered)

	return nil
}

func printLoggerList(out io.Writer, list []types.Logger) {
	w := tabwriter.NewWriter(out, 0, 0, 1, ' ', 0)
	fmt.Fprintf(w, "LOGGER\tLEVEL\t\n")
	for _, l := range list {
		fmt.Fprintf(w, "%s\t%s\t\n", l.Logger, l.Level)
	}
	if err := w.Flush(); err != nil {
		return
	}
}

type sortedLoggers []types.Logger

func (ll sortedLoggers) Len() int {
	return len(ll)
}

func (ll sortedLoggers) Less(i, j int) bool {
	return ll[i].Logger < ll[j].Logger
}

func (ll sortedLoggers) Swap(i, j int) {
	ll[i], ll[j] = ll[j], ll[i]
}

func newLogSetCommand(cli agentcli.Cli) *cobra.Command {
	opts := LogSetOptions{}
	cmd := &cobra.Command{
		Use:   "set <logger> <debug|info|warning|error|fatal|panic>",
		Short: "Set agent logger level",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Logger = args[0]
			opts.Level = args[1]
			return RunLogSet(cli, opts)
		},
	}
	return cmd
}

type LogSetOptions struct {
	Logger string
	Level  string
}

func RunLogSet(cli agentcli.Cli, opts LogSetOptions) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := cli.Client().LoggerSet(ctx, opts.Logger, opts.Level)
	if err != nil {
		return err
	}

	fmt.Fprintf(cli.Out(), "logger %s has been set to level %s\n", opts.Logger, opts.Level)

	return nil
}
