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
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/ligato/vpp-agent/api/types"
	agentcli "github.com/ligato/vpp-agent/cmd/agentctl/cli"
	"github.com/ligato/vpp-agent/pkg/models"
	"github.com/ligato/vpp-agent/plugins/kvscheduler/api"
)

func NewStatusCommand(cli agentcli.Cli) *cobra.Command {
	var opts StatusOptions

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Retrieve agent status",
		Args:  cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Models = args
			return runStatus(cli, opts)
		},
	}
	return cmd
}

type StatusOptions struct {
	Models []string
}

func runStatus(cli agentcli.Cli, opts StatusOptions) error {
	var model string
	if len(opts.Models) > 0 {
		model = opts.Models[0]
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	allModels, err := cli.Client().ModelList(ctx, types.ModelListOptions{})
	if err != nil {
		return err
	}

	var modelKeyPrefix string
	for _, m := range allModels {
		if (m.Alias != "" && model == m.Alias) || model == m.Name {
			modelKeyPrefix = m.KeyPrefix
			break
		}
	}

	status, err := cli.Client().SchedulerStatus(ctx, types.SchedulerStatusOptions{
		KeyPrefix: modelKeyPrefix,
	})
	if err != nil {
		return err
	}

	printStatusTable(status)
	return nil
}

// printStatusTable prints status data using table format
func printStatusTable(status []*api.BaseValueStatus) {
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 10, 0, 3, ' ', 0)
	fmt.Fprintf(w, "MODEL\tNAME\tSTATE\tDETAILS\tLAST OP\tERROR\t\n")

	var printVal = func(val *api.ValueStatus) {
		var (
			model string
			name  string
		)

		m, err := models.GetModelForKey(val.Key)
		if err != nil {
			name = val.Key
		} else {
			model = fmt.Sprintf("%s.%s", m.Module, m.Type)
			name = m.StripKeyPrefix(val.Key)
		}

		var lastOp string
		if val.LastOperation != api.TxnOperation_UNDEFINED {
			lastOp = val.LastOperation.String()
		}
		state := val.State.String()
		if val.State == api.ValueState_OBTAINED {
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
		return
	}
	fmt.Fprint(os.Stdout, buf.String())
}
