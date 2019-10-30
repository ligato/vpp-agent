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
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"go.ligato.io/vpp-agent/v2/cmd/agentctl/api/types"
	agentcli "go.ligato.io/vpp-agent/v2/cmd/agentctl/cli"
	"go.ligato.io/vpp-agent/v2/pkg/models"
	"go.ligato.io/vpp-agent/v2/proto/ligato/kvscheduler"
)

func NewValuesCommand(cli agentcli.Cli) *cobra.Command {
	var opts ValuesOptions

	cmd := &cobra.Command{
		Use:   "values [MODEL]",
		Short: "Retrieve values from scheduler",
		Args:  cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Models = args
			return runValues(cli, opts)
		},
	}
	return cmd
}

type ValuesOptions struct {
	Models []string
}

func runValues(cli agentcli.Cli, opts ValuesOptions) error {
	var model string
	if len(opts.Models) > 0 {
		model = opts.Models[0]
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	allModels, err := cli.Client().ModelList(ctx, types.ModelListOptions{
		Class: "config",
	})
	if err != nil {
		return err
	}

	var modelKeyPrefix string
	for _, m := range allModels {
		if model == m.Name {
			modelKeyPrefix = m.KeyPrefix
			break
		}
	}

	status, err := cli.Client().SchedulerValues(ctx, types.SchedulerValuesOptions{
		KeyPrefix: modelKeyPrefix,
	})
	if err != nil {
		return err
	}

	if err := printValuesTable(cli.Out(), status); err != nil {
		return err
	}

	return nil
}

// printValuesTable prints values data using table format
func printValuesTable(out io.Writer, status []*kvscheduler.BaseValueStatus) error {
	var buf bytes.Buffer

	w := tabwriter.NewWriter(&buf, 10, 0, 3, ' ', 0)
	fmt.Fprintf(w, "MODEL\tNAME\tSTATE\tDETAILS\tLAST OP\tERROR\t\n")

	var printVal = func(val *kvscheduler.ValueStatus) {
		var (
			model string
			name  string
		)

		m, err := models.GetModelForKey(val.Key)
		if err != nil {
			name = val.Key
		} else {
			model = m.Spec().ModelName()
			name = m.StripKeyPrefix(val.Key)
		}

		var lastOp string
		if val.LastOperation != kvscheduler.TxnOperation_UNDEFINED {
			lastOp = val.LastOperation.String()
		}
		state := val.State.String()
		if val.State == kvscheduler.ValueState_OBTAINED {
			state = strings.ToLower(state)
		}

		var details string
		if len(val.Details) > 0 {
			details = strings.Join(val.Details, ", ")
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t\n", model, name, state, details, lastOp, val.Error)
	}

	for _, d := range status {
		printVal(d.Value)
		for _, v := range d.DerivedValues {
			printVal(v)
		}
	}
	if err := w.Flush(); err != nil {
		return err
	}
	_, err := buf.WriteTo(out)
	return err
}
