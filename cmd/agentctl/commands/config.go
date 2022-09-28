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
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	yaml2 "github.com/ghodss/yaml"
	"github.com/olekukonko/tablewriter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"go.ligato.io/vpp-agent/v3/client"
	"go.ligato.io/vpp-agent/v3/cmd/agentctl/api/types"
	agentcli "go.ligato.io/vpp-agent/v3/cmd/agentctl/cli"
	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/proto/ligato/configurator"
	"go.ligato.io/vpp-agent/v3/proto/ligato/kvscheduler"
)

func NewConfigCommand(cli agentcli.Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage agent configuration",
	}
	cmd.AddCommand(
		newConfigGetCommand(cli),
		newConfigUpdateCommand(cli),
		newConfigDeleteCommand(cli),
		newConfigRetrieveCommand(cli),
		newConfigWatchCommand(cli),
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
	flags.StringSliceVar(&opts.Labels, "labels", []string{}, "Output only config items that have given labels. "+
		"Format of labels is: \"<string>=<string>\" key-value pairs separated by comma. "+
		"Empty keys and duplicated keys are not allowed. "+
		"If value of label is empty, equals sign can be omitted. "+
		"For example: --labels=\"foo=bar\",\"baz=\",\"qux\"")
	return cmd
}

type ConfigGetOptions struct {
	Format string
	Labels []string
}

func runConfigGet(cli agentcli.Cli, opts ConfigGetOptions) error {
	// get generic client
	c, err := cli.Client().GenericClient()
	if err != nil {
		return err
	}

	// create dynamically config that can hold all remote known models
	// (not using local model registry that gives only locally available models)
	knownModels, err := c.KnownModels("config")
	if err != nil {
		return fmt.Errorf("getting registered models: %w", err)
	}
	config, err := client.NewDynamicConfig(knownModels)
	if err != nil {
		return fmt.Errorf("can't create all-config proto message dynamically due to: %w", err)
	}

	// fill labels map
	labels, err := parseLabels(opts.Labels)
	if err != nil {
		return fmt.Errorf("parsing labels failed: %w", err)
	}

	// retrieve data into config
	err = c.GetFilteredConfig(client.Filter{Labels: labels}, config)
	if err != nil {
		return fmt.Errorf("can't retrieve configuration due to: %w", err)
	}

	// handle data output
	format := opts.Format
	if len(format) == 0 {
		format = `yaml`
	}
	if err := formatAsTemplate(cli.Out(), format, config); err != nil {
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
		Long:  "Update configuration in agent from file",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigUpdate(cli, opts, args)
		},
	}
	flags := cmd.Flags()
	flags.StringVarP(&opts.Format, "format", "f", "", "Format output")
	flags.BoolVar(&opts.Replace, "replace", false, "Replaces all existing config")
	// TODO implement waitdone also for generic client
	// flags.BoolVar(&opts.WaitDone, "waitdone", false, "Waits until config update is done")
	// TODO implement transaction output when verbose is used
	// flags.BoolVarP(&opts.Verbose, "verbose", "v", false, "Show verbose output")
	flags.DurationVarP(&opts.Timeout, "timeout", "t",
		5*time.Minute, "Timeout for sending updated data")
	flags.StringSliceVar(&opts.Labels, "labels", []string{}, "Labels associated with updated config items. "+
		"Format of labels is: \"<string>=<string>\" key-value pairs separated by comma. "+
		"Empty keys and duplicated keys are not allowed. "+
		"If value of label is empty, equals sign can be omitted. "+
		"For example: --labels=\"foo=bar\",\"baz=\",\"qux\"")
	return cmd
}

type ConfigUpdateOptions struct {
	Format  string
	Replace bool
	// WaitDone bool
	// Verbose  bool
	Timeout time.Duration
	Labels  []string
}

func runConfigUpdate(cli agentcli.Cli, opts ConfigUpdateOptions, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), opts.Timeout)
	defer cancel()

	// get input file
	if len(args) == 0 {
		return fmt.Errorf("missing file argument")
	}
	file := args[0]
	b, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("reading file %s: %w", file, err)
	}

	// get generic client
	c, err := cli.Client().GenericClient()
	if err != nil {
		return err
	}

	// create dynamically config that can hold all remote known models
	// (not using local model registry that gives only locally available models)
	knownModels, err := c.KnownModels("config")
	if err != nil {
		return fmt.Errorf("getting registered models: %w", err)
	}
	config, err := client.NewDynamicConfig(knownModels)
	if err != nil {
		return fmt.Errorf("can't create all-config proto message dynamically due to: %w", err)
	}

	// filling dynamically created config with data from input file
	bj, err := yaml2.YAMLToJSON(b)
	if err != nil {
		return fmt.Errorf("converting to JSON: %w", err)
	}
	err = protojson.Unmarshal(bj, config)
	if err != nil {
		return fmt.Errorf("can't unmarshall input file data "+
			"into dynamically created config due to: %v", err)
	}
	logrus.Infof("loaded config :\n%s", config)

	// extracting proto messages from dynamically created config structure
	// (generic client wants single proto messages and not one big hierarchical config)
	configMessages, err := client.DynamicConfigExport(config)
	if err != nil {
		return fmt.Errorf("can't extract single configuration proto messages "+
			"from one big configuration proto message due to: %v", err)
	}

	// fill labels map
	labels, err := parseLabels(opts.Labels)
	if err != nil {
		return fmt.Errorf("parsing labels failed: %w", err)
	}

	// update/resync configuration
	_, err = c.UpdateItems(ctx, createUpdateItems(configMessages, labels), opts.Replace)
	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	// handle configuration update result and command output
	format := opts.Format
	if len(format) == 0 {
		format = `{{.}}`
	}
	if err := formatAsTemplate(cli.Out(), format, "OK"); err != nil {
		return err
	}

	return nil
}

func newConfigDeleteCommand(cli agentcli.Cli) *cobra.Command {
	var (
		opts ConfigDeleteOptions
	)
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete config in agent",
		Long:  "Delete configuration in agent",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigDelete(cli, opts, args)
		},
	}
	flags := cmd.Flags()
	flags.StringVarP(&opts.Format, "format", "f", "", "Format output")
	flags.BoolVar(&opts.WaitDone, "waitdone", false, "Waits until config update is done")
	flags.BoolVarP(&opts.Verbose, "verbose", "v", false, "Show verbose output")
	return cmd
}

type ConfigDeleteOptions struct {
	Format   string
	WaitDone bool
	Verbose  bool
}

func runConfigDelete(cli agentcli.Cli, opts ConfigDeleteOptions, args []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c, err := cli.Client().ConfiguratorClient()
	if err != nil {
		return err
	}

	if len(args) == 0 {
		return fmt.Errorf("missing file argument")
	}
	file := args[0]
	b, err := os.ReadFile(file)
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
	logrus.Infof("loaded config delete:\n%s", update)

	var data interface{}

	var header metadata.MD
	resp, err := c.Delete(ctx, &configurator.DeleteRequest{
		Delete:   update,
		WaitDone: opts.WaitDone,
	}, grpc.Header(&header))
	if err != nil {
		logrus.Warnf("delete failed: %v", err)
		data = err
	} else {
		data = resp
	}

	if opts.Verbose {
		logrus.Debugf("grpc header: %+v", header)
		if seqNum, ok := header["seqnum"]; ok {
			ref, _ := strconv.Atoi(seqNum[0])
			txns, err := cli.Client().SchedulerHistory(ctx, types.SchedulerHistoryOptions{
				SeqNum: ref,
			})
			if err != nil {
				logrus.Warnf("getting history for seqNum %d failed: %v", ref, err)
			} else {
				data = txns
			}
		}
	}

	format := opts.Format
	if len(format) == 0 {
		format = `{{.}}`
	}
	if err := formatAsTemplate(cli.Out(), format, data); err != nil {
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
		Aliases: []string{"ret", "read", "dump"},
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

	c, err := cli.Client().ConfiguratorClient()
	if err != nil {
		return err
	}
	resp, err := c.Dump(ctx, &configurator.DumpRequest{})
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

func newConfigWatchCommand(cli agentcli.Cli) *cobra.Command {
	var (
		opts ConfigWatchOptions
	)
	cmd := &cobra.Command{
		Use:     "watch",
		Aliases: []string{"notify", "subscribe"},
		Short:   "Watch events",
		Example: "Filter events by VPP interface name 'loop1'" +
			`{{.CommandPath}} config watch --filter='{"vpp_notification":{"interface":{"state":{"name":"loop1"}}}}'` +
			"" +
			"Filter events by VPP interface UPDOWN type" +
			`{{.CommandPath}} config watch --filter='{"vpp_notification":{"interface":{"type":"UPDOWN"}}}'`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigWatch(cli, opts)
		},
	}
	flags := cmd.Flags()
	flags.StringArrayVar(&opts.Filters, "filter", nil, "Filter(s) for notifications (multiple filters are used with AND operator). Value should be JSON data of configurator.Notification.")
	flags.StringVarP(&opts.Format, "format", "f", "", "Format output")
	return cmd
}

type ConfigWatchOptions struct {
	Format  string
	Filters []string
}

func runConfigWatch(cli agentcli.Cli, opts ConfigWatchOptions) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c, err := cli.Client().ConfiguratorClient()
	if err != nil {
		return err
	}

	filters, err := prepareNotifyFilters(opts.Filters)
	if err != nil {
		return fmt.Errorf("filters error: %w", err)
	}

	var nextIdx uint32
	stream, err := c.Notify(ctx, &configurator.NotifyRequest{
		Idx:     nextIdx,
		Filters: filters,
	})
	if err != nil {
		return err
	}

	format := opts.Format
	if len(format) == 0 {
		format = `------------------
 NOTIFICATION #{{.NextIdx}}
------------------
{{if .Notification.GetVppNotification}}Source: VPP
Value: {{protomulti .Notification.GetVppNotification}}
{{else if .Notification.GetLinuxNotification}}Source: LINUX
Value:  {{protomulti .Notification.GetLinuxNotification}}
{{else}}Source: {{printf "%T" .Notification.GetNotification}}
Value:  {{protomulti .Notification.GetNotification}}
{{end}}`
	}

	for {
		notif, err := stream.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		logrus.Debugf("Notification[%d]: %v",
			notif.NextIdx-1, notif.Notification)
		nextIdx = notif.NextIdx

		if err := formatAsTemplate(cli.Out(), format, notif); err != nil {
			return err
		}
	}

	return nil
}

func prepareNotifyFilters(filters []string) ([]*configurator.Notification, error) {
	var list []*configurator.Notification
	for _, filter := range filters {
		notif := &configurator.Notification{}
		err := protojson.Unmarshal([]byte(filter), notif)
		if err != nil {
			return nil, err
		}
		list = append(list, notif)
	}
	return list, nil
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
specifying a reference to sequence number (txn ID).

Type can be one of:
 - config change  (NB - full resync)
 - status update  (SB)
 - config sync    (NB - upstream resync)
 - status sync    (NB - downstream resync)
 - retry #X for Y (retry of TX)
`,
		Example: `
# Show entire history
{{.CommandPath}} config history

# Show entire history with details
{{.CommandPath}} config history --details

# Show entire history in transaction log format
{{.CommandPath}} config history -f log

# Show entire history in classic log format
{{.CommandPath}} config history -f log

# Show history point with sequence number 3
{{.CommandPath}} config history 3

# Show history point with seq. number 3 in log format
{{.CommandPath}} config history -f log 3
`,
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
	flags.BoolVar(&opts.Details, "details", false, "Include details")
	return cmd
}

type ConfigHistoryOptions struct {
	Format  string
	Details bool
	TxnRef  string
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

	// register remote models into the default registry
	_, err = cli.Client().ModelList(ctx, types.ModelListOptions{
		Class: "config",
	})
	if err != nil {
		return err
	}

	txns, err := cli.Client().SchedulerHistory(ctx, types.SchedulerHistoryOptions{
		SeqNum: ref,
	})
	if err != nil {
		return err
	}

	format := opts.Format
	if len(format) == 0 {
		printHistoryTable(cli.Out(), txns, opts.Details)
	} else if format == "log" {
		format = "{{.}}"
	}
	if err := formatAsTemplate(cli.Out(), format, txns); err != nil {
		return err
	}

	return nil
}

func printHistoryTable(out io.Writer, txns kvs.RecordedTxns, withDetails bool) {
	table := tablewriter.NewWriter(out)
	header := []string{
		"Seq", "Type", "Start", "Input", "Operations", "Result", "Summary",
	}
	if withDetails {
		header = append(header, "Details")
	}
	table.SetHeader(header)
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
		age := shortHumanDuration(time.Since(txn.Start))
		var result string
		var resClr int
		var detail string
		var summary string
		var input string
		if len(txn.Values) > 0 {
			input = fmt.Sprintf("%-2d values", len(txn.Values))
		} else {
			input = "<none>"
		}
		var operation string
		if len(txn.Executed) > 0 {
			operation = txnOperations(txn)
			summary = txnValueStates(txn)
		} else {
			operation = "<none>"
			summary = "<none>"
		}
		errs := txnErrors(txn)
		if errs != nil {
			result = "error"
			resClr = tablewriter.FgHiRedColor
			if len(errs) > 1 {
				result = fmt.Sprintf("%d errors", len(errs))
			}
		} else if len(txn.Executed) > 0 {
			result = "ok"
			resClr = tablewriter.FgGreenColor
		}
		if withDetails {
			if errs != nil {
				for _, e := range errs {
					if detail != "" {
						detail += "\n"
					}
					detail += fmt.Sprintf("%v", e.Error())
				}
			}
			if reasons := txnPendingReasons(txn); reasons != "" {
				if detail != "" {
					detail += "\n"
				}
				detail += reasons
			}
		}
		row := []string{
			fmt.Sprint(txn.SeqNum),
			typ,
			age,
			input,
			operation,
			result,
			summary,
		}
		if withDetails {
			row = append(row, detail)
		}
		clrs := []tablewriter.Colors{
			{},
			{tablewriter.Normal, clr},
			{},
			{},
			{},
			{resClr},
			{},
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
			return "status sync"
		}
		return "config change"
	case kvs.RetryFailedOps:
		return fmt.Sprintf("retry #%d for %d", txn.RetryAttempt, txn.RetryForTxn)
	}
	return "?"
}

func txnValueStates(txn *kvs.RecordedTxn) string {
	opermap := map[string]int{}
	for _, r := range txn.Executed {
		opermap[r.NewState.String()]++
	}
	var opers []string
	for k, v := range opermap {
		opers = append(opers, fmt.Sprintf("%s:%v", k, v))
	}
	sort.Strings(opers)
	return strings.Join(opers, ", ")
}

func txnOperations(txn *kvs.RecordedTxn) string {
	opermap := map[string]int{}
	for _, r := range txn.Executed {
		opermap[r.Operation.String()]++
	}
	var opers []string
	for k, v := range opermap {
		opers = append(opers, fmt.Sprintf("%s:%v", k, v))
	}
	sort.Strings(opers)
	return strings.Join(opers, ", ")
}

func txnPendingReasons(txn *kvs.RecordedTxn) string {
	var details []string
	for _, r := range txn.Executed {
		if r.NewState == kvscheduler.ValueState_PENDING {
			// TODO: include pending resons in details
			detail := fmt.Sprintf("[%s] %s -> %s", r.Operation, r.Key, r.NewState)
			details = append(details, detail)
		}
	}
	return strings.Join(details, "\n")
}

func txnErrors(txn *kvs.RecordedTxn) Errors {
	var errs Errors
	for _, r := range txn.Executed {
		if r.NewErrMsg != "" {
			r.NewErr = fmt.Errorf("[%s] %s -> %s: %v", r.Operation, r.Key, r.NewState, r.NewErrMsg)
			errs = append(errs, r.NewErr)
		}
	}
	return errs
}

// parseLabels parses labels obtained from command line flags
// This function does not allow duplicate or empty ("") keys
func parseLabels(rawLabels []string) (map[string]string, error) {
	if len(rawLabels) == 0 {
		return nil, nil
	}
	labels := make(map[string]string)
	var lkey, lval string
	for _, rawLabel := range rawLabels {
		i := strings.IndexByte(rawLabel, '=')
		if i == -1 {
			lkey, lval = rawLabel, ""
		} else {
			lkey, lval = rawLabel[:i], rawLabel[i+1:]
		}
		if lkey == "" {
			return nil, fmt.Errorf("key of label %s is empty", rawLabel)
		}
		if _, ok := labels[lkey]; ok {
			return nil, fmt.Errorf("label key %s is duplicated", lkey)
		}
		labels[lkey] = lval
	}
	return labels, nil
}

func createUpdateItems(msgs []proto.Message, labels map[string]string) []client.UpdateItem {
	var result []client.UpdateItem
	for _, msg := range msgs {
		result = append(result, client.UpdateItem{Message: msg, Labels: labels})
	}
	return result
}
