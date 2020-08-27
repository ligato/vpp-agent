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
	if err := formatAsTemplate(cli.Out(), format, resp); err != nil {
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
		Use:   "history",
		Short: "Retrieve config history",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigHistory(cli, opts)
		},
	}
	flags := cmd.Flags()
	flags.StringVarP(&opts.Format, "format", "f", "", "Format output")
	return cmd
}

type ConfigHistoryOptions struct {
	Format  string
	Verbose bool
	Retry   bool
}

func runConfigHistory(cli agentcli.Cli, opts ConfigHistoryOptions) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	txns, err := cli.Client().SchedulerHistory(ctx, types.SchedulerHistoryOptions{})
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
		"Seq", "", "Type", "", "Age", "Summary", "Result",
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
		typ := kvs.TxnTypeToString(txn.TxnType)
		info := txn.Description
		if txn.TxnType == kvs.NBTransaction && txn.ResyncType != kvs.NotResync {
			info = fmt.Sprintf("%s", kvs.ResyncTypeToString(txn.ResyncType))
		}
		elapsed := txn.Stop.Sub(txn.Start).Round(time.Millisecond / 10)
		took := elapsed.String()
		if elapsed < time.Millisecond/10 {
			took = "<.1ms"
		} else if elapsed > time.Millisecond*100 {
			took = elapsed.Round(time.Millisecond).String()
		}
		_ = took
		result := txnErrors(txn)
		resClr := tablewriter.FgGreenColor
		if result != "" {
			resClr = tablewriter.FgHiRedColor
		} else {
			result = "ok"
		}
		var typClr int
		switch txn.TxnType {
		case kvs.NBTransaction:
			typClr = tablewriter.FgYellowColor
		case kvs.SBNotification:
			typClr = tablewriter.FgCyanColor
		case kvs.RetryFailedOps:
			typClr = tablewriter.FgMagentaColor
		}
		age := shortHumanDuration(time.Since(txn.Start))
		summary := fmt.Sprintf("%d executed", len(txn.Executed))
		row := []string{
			fmt.Sprintf("%3v", txn.SeqNum),
			fmt.Sprintf("%v", txnIcon(txn)),
			typ,
			info,
			fmt.Sprintf("%-3s", age),
			//fmt.Sprintf("%-3s (took %v)", age, took),
			fmt.Sprintf("values: %2d -> %s", len(txn.Values), summary),
			result,
		}
		clrs := []tablewriter.Colors{
			{tablewriter.Normal, typClr},
			{tablewriter.Bold, typClr + 60},
			{tablewriter.Normal, typClr},
			{},
			{},
			{},
			{resClr},
		}
		table.Rich(row, clrs)
	}
	table.Render()
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
		return "⟱"
	case kvs.RetryFailedOps:
		return "↻"
	}
	return "?"
}
