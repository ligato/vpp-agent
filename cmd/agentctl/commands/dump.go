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
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/golang/protobuf/proto"
	"github.com/ligato/cn-infra/logging"
	"github.com/spf13/cobra"

	"go.ligato.io/vpp-agent/v3/cmd/agentctl/api/types"
	agentcli "go.ligato.io/vpp-agent/v3/cmd/agentctl/cli"
	"go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
)

func NewDumpCommand(cli agentcli.Cli) *cobra.Command {
	var opts DumpOptions

	cmd := &cobra.Command{
		Use:   "dump MODEL",
		Short: "Dump running state",
		Example: `
 To dump all data:
  $ {{.CommandPath}} all

 To dump all VPP data in json format run:
  $ {{.CommandPath}} -f json vpp.*

 To use different dump view use --view flag:
  $ {{.CommandPath}} --view=NB vpp.interfaces`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Models = args
			return runDump(cli, opts)
		},
	}
	flags := cmd.Flags()
	flags.StringVar(&opts.View, "view", "cached", "Dump view type: cached, NB, SB")
	flags.StringVarP(&opts.Format, "format", "f", "", "Format output")
	return cmd
}

type DumpOptions struct {
	Models []string
	View   string
	Format string
}

func runDump(cli agentcli.Cli, opts DumpOptions) error {
	dumpView := opts.View

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	allModels, err := cli.Client().ModelList(ctx, types.ModelListOptions{
		Class: "config",
	})
	if err != nil {
		return err
	}

	refs := opts.Models
	if opts.Models[0] == "all" {
		refs = []string{"*"}
	}

	var keyPrefixes []string
	for _, m := range filterModelsByRefs(allModels, refs) {
		keyPrefixes = append(keyPrefixes, m.KeyPrefix)
	}
	if len(keyPrefixes) == 0 {
		return fmt.Errorf("no models found for %q", opts.Models)
	}

	var dumps []api.KVWithMetadata
	for _, keyPrefix := range keyPrefixes {
		dump, err := cli.Client().SchedulerDump(ctx, types.SchedulerDumpOptions{
			KeyPrefix: keyPrefix,
			View:      dumpView,
		})
		if err != nil {
			logging.Debug(fmt.Errorf("dump for %s failed: %v", keyPrefix, err))
			continue
		}
		dumps = append(dumps, dump...)
	}

	sort.Sort(dumpByKey(dumps))

	format := opts.Format
	if len(format) == 0 {
		printDumpTable(cli.Out(), dumps)
	} else {
		if err := formatAsTemplate(cli.Out(), format, dumps); err != nil {
			return err
		}
	}

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
		panic(err)
	}
	fmt.Fprint(out, buf.String())
}

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
