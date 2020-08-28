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
	"io/ioutil"
	"strconv"
	"time"

	yaml2 "github.com/ghodss/yaml"
	"github.com/olekukonko/tablewriter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"

	"go.ligato.io/vpp-agent/v3/cmd/agentctl/api/types"
	agentcli "go.ligato.io/vpp-agent/v3/cmd/agentctl/cli"
	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/proto/ligato/configurator"
)

func NewConfigCommand(cli agentcli.Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage agent configuration",
	}
	cmd.AddCommand(
		newConfigGetCommand(cli),
		newConfigRetrieveCommand(cli),
		newConfigUpdateCommand(cli),
		newConfigResyncCommand(cli),
		newConfigHistoryCommand(cli),
	)
	return cmd
}

func newConfigGetCommand(cli agentcli.Cli) *cobra.Command {
	var (
		opts ConfigGetOptions
	)
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get config from agent",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigGet(cli, opts)
		},
	}
	flags := cmd.Flags()
	flags.StringVarP(&opts.Format, "format", "f", "", "Format output")
	return cmd
}

type ConfigGetOptions struct {
	Format string
}

func runConfigGet(cli agentcli.Cli, opts ConfigGetOptions) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, err := cli.Client().ConfiguratorClient()
	if err != nil {
		return err
	}
	resp, err := client.Get(ctx, &configurator.GetRequest{})
	if err != nil {
		return err
	}

	format := opts.Format
	if len(format) == 0 {
		format = `yaml`
	}
	if err := formatAsTemplate(cli.Out(), format, resp.Config); err != nil {
		return err
	}

	return nil
}

func newConfigUpdateCommand(cli agentcli.Cli) *cobra.Command {
	var (
		opts ConfigUpdateOptions
	)
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update config in agent",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigUpdate(cli, opts, args)
		},
	}
	flags := cmd.Flags()
	flags.StringVarP(&opts.Format, "format", "f", "", "Format output")
	flags.BoolVar(&opts.Replace, "replace", false, "Replaces entire config in agent")
	return cmd
}

type ConfigUpdateOptions struct {
	Format  string
	Replace bool
}

func runConfigUpdate(cli agentcli.Cli, opts ConfigUpdateOptions, args []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, err := cli.Client().ConfiguratorClient()
	if err != nil {
		return err
	}

	if len(args) == 0 {
		return fmt.Errorf("missing file argument")
	}
	file := args[0]
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return fmt.Errorf("reading file %s: %w", file, err)
	}

	var update = &configurator.Config{}
	bj, err := yaml2.YAMLToJSON(b)
	if err != nil {
		return fmt.Errorf("converting to JSON: %w", err)
	}
	err = protojson.Unmarshal(bj, update)
	if err != nil {
		return err
	}
	logrus.Infof("loaded config update:\n%s", update)

	if _, err := client.Update(ctx, &configurator.UpdateRequest{
		Update:     update,
		FullResync: opts.Replace,
	}); err != nil {
		return err
	}

	return nil
}

func newConfigRetrieveCommand(cli agentcli.Cli) *cobra.Command {
	var (
		opts ConfigRetrieveOptions
	)
	cmd := &cobra.Command{
		Use:     "retrieve",
		Aliases: []string{"ret", "read"},
		Short:   "Retrieve currently running config",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigRetrieve(cli, opts)
		},
	}
	flags := cmd.Flags()
	flags.StringVarP(&opts.Format, "format", "f", "", "Format output")
	return cmd
}

type ConfigRetrieveOptions struct {
	Format string
}

func runConfigRetrieve(cli agentcli.Cli, opts ConfigRetrieveOptions) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, err := cli.Client().ConfiguratorClient()
	if err != nil {
		return err
	}
	resp, err := client.Dump(ctx, &configurator.DumpRequest{})
	if err != nil {
		return err
	}

	format := opts.Format
	if len(format) == 0 {
		format = `yaml`
	}
	if err := formatAsTemplate(cli.Out(), format, resp.Dump); err != nil {
		return err
	}

	return nil
}

func newConfigResyncCommand(cli agentcli.Cli) *cobra.Command {
	var (
		opts ConfigResyncOptions
	)
	cmd := &cobra.Command{
		Use:   "resync",
		Short: "Run config resync",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigResync(cli, opts)
		},
	}
	flags := cmd.Flags()
	flags.StringVarP(&opts.Format, "format", "f", "", "Format output")
	flags.BoolVar(&opts.Verbose, "verbose", false, "Run resync in verbose mode")
	flags.BoolVar(&opts.Retry, "retry", false, "Run resync with retries")
	return cmd
}

type ConfigResyncOptions struct {
	Format  string
	Verbose bool
	Retry   bool
}

// TODO: define default format with go template
const defaultFormatConfigResync = `json`

func runConfigResync(cli agentcli.Cli, opts ConfigResyncOptions) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rectxn, err := cli.Client().SchedulerResync(ctx, types.SchedulerResyncOptions{
		Retry:   opts.Retry,
		Verbose: opts.Verbose,
	})
	if err != nil {
		return err
	}
	format := opts.Format
	if len(format) == 0 {
		format = defaultFormatConfigResync
	}
	if err := formatAsTemplate(cli.Out(), format, rectxn); err != nil {
		return err
	}

	return nil
}

func newConfigHistoryCommand(cli agentcli.Cli) *cobra.Command {
	var (
		opts ConfigHistoryOptions
	)
	cmd := &cobra.Command{
		Use:   "history [REF]",
		Short: "Show config history",
		Long: `Show history of config changes and status updates

 Prints a table of most important information about the history of changes to 
 config and status updates that have occurred. You can filter the output by
 specifying a reference to sequence number (txn ID).`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.TxnRef = args[0]
			}
			return runConfigHistory(cli, opts)
		},
	}
	flags := cmd.Flags()
	flags.StringVarP(&opts.Format, "format", "f", "", "Format output")
	return cmd
}

type ConfigHistoryOptions struct {
	Format string
	TxnRef string
}

func runConfigHistory(cli agentcli.Cli, opts ConfigHistoryOptions) (err error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ref := -1
	if opts.TxnRef != "" {
		ref, err = strconv.Atoi(opts.TxnRef)
		if err != nil {
			return fmt.Errorf("invalid reference: %q, use number > 0", opts.TxnRef)
		}
	}

	txns, err := cli.Client().SchedulerHistory(ctx, types.SchedulerHistoryOptions{
		SeqNum: ref,
	})
	if err != nil {
		return err
	}
	format := opts.Format
	if len(format) == 0 {
		printHistoryTable(cli.Out(), txns)
	}
	if err := formatAsTemplate(cli.Out(), format, txns); err != nil {
		return err
	}

	return nil
}

func printHistoryTable(out io.Writer, txns kvs.RecordedTxns) {
	table := tablewriter.NewWriter(out)
	table.SetHeader([]string{
		"Seq", "", "Type", "Start", "Input", "Operations", "Result", "Summary",
	})
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t")

	for _, txn := range txns {
		typ := getTxnType(txn)
		clr := getTxnColor(txn)

		result := txnErrors(txn)
		resClr := tablewriter.FgGreenColor
		if result != "" {
			resClr = tablewriter.FgHiRedColor
		} else {
			result = "ok"
		}
		age := shortHumanDuration(time.Since(txn.Start))
		var input string
		if len(txn.Values) > 0 {
			input = fmt.Sprintf("%-2d values", len(txn.Values))
		} else {
			input = "-"
		}
		var operation string
		if len(txn.Executed) > 0 {
			operation = fmt.Sprintf("%-2d executed", len(txn.Executed))
		} else {
			operation = "-"
		}
		summary := txn.Description
		row := []string{
			fmt.Sprintf("%3v", txn.SeqNum),
			txnIcon(txn),
			typ,
			fmt.Sprintf("%-3s", age),
			input,
			operation,
			result,
			summary,
		}
		clrs := []tablewriter.Colors{
			{},
			{tablewriter.Bold, clr},
			{tablewriter.Normal, clr},
			{},
			{},
			{},
			{resClr},
			{},
		}
		table.Rich(row, clrs)
	}
	table.Render()
}

func getTxnColor(txn *kvs.RecordedTxn) int {
	var clr int
	switch txn.TxnType {
	case kvs.NBTransaction:
		if txn.ResyncType == kvs.NotResync {
			clr = tablewriter.FgYellowColor
		} else if txn.ResyncType == kvs.FullResync {
			clr = tablewriter.FgHiYellowColor
		} else {
			clr = tablewriter.FgYellowColor
		}
	case kvs.SBNotification:
		clr = tablewriter.FgCyanColor
	case kvs.RetryFailedOps:
		clr = tablewriter.FgMagentaColor
	}
	return clr
}

func getTxnType(txn *kvs.RecordedTxn) string {
	switch txn.TxnType {
	case kvs.SBNotification:
		return "status update"
	case kvs.NBTransaction:
		if txn.ResyncType == kvs.FullResync {
			return "config replace"
		} else if txn.ResyncType == kvs.UpstreamResync {
			return "config sync"
		} else if txn.ResyncType == kvs.DownstreamResync {
			return "config check"
		}
		return "config change"
	case kvs.RetryFailedOps:
		return fmt.Sprintf("retry #%d for %d", txn.RetryAttempt, txn.RetryForTxn)
	}
	return "?"
}

func txnErrors(txn *kvs.RecordedTxn) string {
	var errs Errors
	for _, r := range txn.Executed {
		if r.NewErrMsg != "" {
			r.NewErr = fmt.Errorf("%v", r.NewErrMsg)
			errs = append(errs, r.NewErr)
		}
	}
	if errs != nil {
		word := "error"
		if len(errs) > 1 {
			word = fmt.Sprintf("%d errors", len(errs))
		}
		return fmt.Sprintf("%s: %v", word, errs.Error())
	}
	return ""
}

func txnIcon(txn *kvs.RecordedTxn) string {
	switch txn.TxnType {
	case kvs.SBNotification:
		return "⇧"
	case kvs.NBTransaction:
		if txn.ResyncType == kvs.NotResync {
			return "⇩"
		} else if txn.ResyncType == kvs.FullResync {
			return "⟱"
		}
		return "⇅"
	case kvs.RetryFailedOps:
		return "↻"
	}
	return "?"
}
