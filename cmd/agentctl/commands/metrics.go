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
	"fmt"

	agentcli "github.com/ligato/vpp-agent/cmd/agentctl/cli"
	"github.com/spf13/cobra"
)

func NewMetricsCommand(cli agentcli.Cli) *cobra.Command {
	var opts MetricsOptions

	cmd := &cobra.Command{
		Use:   "metrics",
		Short: "Retrieve runtime metrics",
		Args:  cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Patterns = args
			return runMetrics(cli, opts)
		},
		DisableFlagsInUseLine: true,
	}
	return cmd
}

type MetricsOptions struct {
	Patterns []string
}

func runMetrics(cli agentcli.Cli, opts MetricsOptions) error {
	metrics, err := retrieveMetrics(cli, opts.Patterns)
	if err != nil {
		return err
	}

	fmt.Fprintf(cli.Out(), "%s\n", metrics)
	return nil
}

func retrieveMetrics(cli agentcli.Cli, patterns []string) (string, error) {
	/*q := fmt.Sprintf(`/debug/exp/var`)

	resp, err := cli.Client().
	if err != nil {
		return "", err
	}
	logging.Debugf("metrics resp: %s\n", resp)*/

	/*var out bytes.Buffer
	if err := json.Indent(&out, resp, "", "  "); err != nil {
		return "", err
	}
	metrics := fmt.Sprintf("[%s]", out.Bytes())*/
	//metrics := string(resp)

	/*var metrics []*api.BaseValueMetrics
	if err := json.Unmarshal(resp, &metrics); err != nil {
		return nil, fmt.Errorf("decoding reply failed: %v", err)
	}**/

	return "metrics", nil
}
