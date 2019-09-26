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
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/gogo/protobuf/proto"
	"github.com/spf13/cobra"

	"github.com/ligato/vpp-agent/api/types"
	agentcli "github.com/ligato/vpp-agent/cmd/agentctl/cli"
	"github.com/ligato/vpp-agent/plugins/kvscheduler/api"
)

func NewDumpCommand(cli agentcli.Cli) *cobra.Command {
	var opts DumpOptions

	cmd := &cobra.Command{
		Use:     "dump MODEL",
		Aliases: []string{"d"},
		Short:   "Dump running state",
		Example: `
 To dump VPP interfaces run:
  $ agentctl dump vpp.interfaces

 To use different dump view use --view flag:
  $ agentctl dump --view=NB vpp.interfaces

 For a list of all supported models that can be dumped run:
  $ agentctl model list

 To specify the HTTP address of the agent use --host flag:
  $ agentctl --host 172.17.0.3 dump vpp.interfaces
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Models = args
			return runDump(cli, opts)
		},
	}
	cmd.Flags().StringVarP(&opts.View, "view", "v", "cached", "Dump view type: cached, NB, SB")
	return cmd
}

type DumpOptions struct {
	Models []string
	View   string
}

func runDump(cli agentcli.Cli, opts DumpOptions) error {
	dumpView := opts.View
	model := opts.Models[0]

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
	if modelKeyPrefix == "" {
		fmt.Fprintf(os.Stderr, "No model found for: %q\n", model)
		return fmt.Errorf("no such model")
	}

	dump, err := cli.Client().SchedulerDump(ctx, types.SchedulerDumpOptions{
		KeyPrefix: modelKeyPrefix,
		View:      dumpView,
	})
	if err != nil {
		return err
	}

	sort.Sort(dumpByKey(dump))
	printDumpTable(cli.Out(), dump)

	return nil
}

// printDumpTable prints dump data using table format
//
// KEY                                        VALUE                        ORIGIN    METADATA
// config/vpp/v2/interfaces/UNTAGGED-local0   [vpp.interfaces.Interface]   from-SB   map[IPAddresses:<nil> SwIfIndex:0 TAPHostIfName: Vrf:0]
// name: "UNTAGGED-local0"
// type: SOFTWARE_LOOPBACK
//
// config/vpp/v2/interfaces/loop1             [vpp.interfaces.Interface]   from-NB   map[IPAddresses:<nil> SwIfIndex:1 TAPHostIfName: Vrf:0]
// name: "loop1"
// type: SOFTWARE_LOOPBACK
//
func printDumpTable(out io.Writer, dump []api.KVWithMetadata) {
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 3, ' ', 0)
	fmt.Fprintf(w, "KEY\tVALUE\tORIGIN\tMETADATA\t\n")

	for _, d := range dump {
		val := proto.MarshalTextString(d.Value)
		val = strings.ReplaceAll(val, "\n", "\t\t\t\n\t")
		var meta string
		if d.Metadata != nil {
			meta = fmt.Sprintf("%+v", d.Metadata)
		}

		fmt.Fprintf(w, "%s\t[%s]\t%s\t%s\t\n",
			d.Key, proto.MessageName(d.Value), d.Origin, meta)
		fmt.Fprintf(w, "\t%s\t\t\n", val)
	}
	if err := w.Flush(); err != nil {
		return
	}
	fmt.Fprint(out, buf.String())
}

/*func dumpKeyPrefix(cli agentcli.Cli, keyPrefix string, dumpView string) ([]api.KVWithMetadata, error) {
	type ProtoWithName struct {
		ProtoMsgName string
		ProtoMsgData string
	}
	type KVWithMetadata struct {
		api.KVWithMetadata
		Value ProtoWithName
	}
	var kvdump []KVWithMetadata

	q := fmt.Sprintf(`/scheduler/dump?key-prefix=%s&view=%s`,
		url.QueryEscape(keyPrefix), url.QueryEscape(dumpView))

	resp, err := cli.GET(q)
	if err != nil {
		return nil, err
	}

	Debugf("dump respo: %s\n", resp)

	if err := json.Unmarshal(resp, &kvdump); err != nil {
		return nil, fmt.Errorf("decoding reply failed: %v", err)
	}

}*/

type dumpByKey []api.KVWithMetadata

func (s dumpByKey) Len() int {
	return len(s)
}

func (s dumpByKey) Less(i, j int) bool {
	return s[i].Key < s[j].Key
}

func (s dumpByKey) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
