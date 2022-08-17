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
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"go.ligato.io/cn-infra/v2/logging"

	"go.ligato.io/vpp-agent/v3/cmd/agentctl/api/types"
	agentcli "go.ligato.io/vpp-agent/v3/cmd/agentctl/cli"
	"go.ligato.io/vpp-agent/v3/pkg/models"
	"go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
)

func NewDumpCommand(cli agentcli.Cli) *cobra.Command {
	var (
		opts DumpOptions
	)
	cmd := &cobra.Command{
		Use:   "dump MODEL [MODEL...]",
		Short: "Dump running state",
		Long:  "Dumps actual running state",
		Example: `
# Dump everything
{{.CommandPath}} all

# Dump VPP interfaces & routes
{{.CommandPath}} vpp.interfaces vpp.l3.routes

# Dump all VPP data in JSON format
{{.CommandPath}} -f json vpp.*

# Dump only VPP memif interfaces
{{.CommandPath}} -f '{{` + "`{{range .}}{{if eq .Value.Type.String \"MEMIF\" }}{{json .}}{{end}}{{end}}`" + `}}' vpp.interfaces

# Dump everything currently defined at northbound
{{.CommandPath}} --view=NB all

# Dump all VPP & Linux data directly from southband
{{.CommandPath}} --view=SB vpp.* linux.*

# Dump all VPP & Linux data directly from southband
{{.CommandPath}} --view=SB all
`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				fmt.Fprintf(cli.Err(), "You must specify models to dump. Use \"%s models\" for a complete list of known models.\n", cmd.Root().Name())
				return fmt.Errorf("no models specified")
			}
			opts.Models = args
			return opts.Validate()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDump(cli, opts)
		},
	}
	flags := cmd.Flags()
	flags.StringVar(&opts.View, "view", "cached", "Dump view type: cached, NB, SB")
	flags.StringVar(&opts.Origin, "origin", "", "Show only data with specific origin: NB, SB, unknown")
	flags.StringVarP(&opts.Format, "format", "f", "", "Format output (json|yaml|go-template|proto)")
	return cmd
}

type DumpOptions struct {
	Models []string
	View   string
	Origin string
	Format string
}

func (opts *DumpOptions) Validate() error {
	// models
	if opts.Models[0] == "all" {
		opts.Models = []string{"*"}
	}
	// view
	switch strings.ToLower(opts.View) {
	case "cached", "cache", "":
		opts.View = "cached"
	case "nb", "north", "northbound":
		opts.View = "NB"
	case "sb", "south", "southbound":
		opts.View = "SB"
	default:
		return fmt.Errorf("invalid view type: %q", opts.View)
	}
	// origin
	switch strings.ToLower(opts.Origin) {
	case "":
	case "unknown":
		opts.Origin = api.UnknownOrigin.String()
	case "from-nb", "nb", "north", "northbound":
		opts.Origin = api.FromNB.String()
	case "from-sb", "sb", "south", "southbound":
		opts.Origin = api.FromSB.String()
	default:
		return fmt.Errorf("invalid origin: %q", opts.Origin)
	}
	return nil
}

func runDump(cli agentcli.Cli, opts DumpOptions) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	allModels, err := cli.Client().ModelList(ctx, types.ModelListOptions{
		Class: "config",
	})
	if err != nil {
		return err
	}
	var keyPrefixes []string
	for _, m := range filterModelsByRefs(allModels, opts.Models) {
		keyPrefixes = append(keyPrefixes, m.KeyPrefix)
	}
	if len(keyPrefixes) == 0 {
		return fmt.Errorf("no matching models found for %q", opts.Models)
	}
	var (
		errs  Errors
		dumps []api.RecordedKVWithMetadata
	)
	for _, keyPrefix := range keyPrefixes {
		dump, err := cli.Client().SchedulerDump(ctx, types.SchedulerDumpOptions{
			KeyPrefix: keyPrefix,
			View:      opts.View,
		})
		if err != nil {
			errs = append(errs, fmt.Errorf("dump for %s failed: %v", keyPrefix, err))
			continue
		}
		dumps = append(dumps, dump...)
	}
	if errs != nil {
		logging.Debugf("dump finished with %d errors\n%v", len(errs), errs)
	}
	if len(errs) == len(keyPrefixes) {
		return fmt.Errorf("dump failed:\n%v", errs)
	}

	dumps = filterDumpByOrigin(dumps, opts.Origin)
	sort.Slice(dumps, func(i, j int) bool {
		return dumps[i].Key < dumps[j].Key
	})

	format := opts.Format
	if len(format) == 0 {
		printDumpTable(cli.Out(), dumps)
	} else {
		fdumps, err := convertDumps(dumps)
		if err != nil {
			return err
		}
		if err := formatAsTemplate(cli.Out(), format, fdumps); err != nil {
			return err
		}
	}
	return nil
}

func filterDumpByOrigin(dumps []api.RecordedKVWithMetadata, origin string) []api.RecordedKVWithMetadata {
	if origin == "" {
		return dumps
	}
	var filtered []api.RecordedKVWithMetadata
	for _, d := range dumps {
		if !strings.EqualFold(d.Origin.String(), origin) {
			continue
		}
		filtered = append(filtered, d)
	}
	return filtered
}

func printDumpTable(out io.Writer, dump []api.RecordedKVWithMetadata) {
	table := tablewriter.NewWriter(out)
	table.SetHeader([]string{
		"Model", "Origin", "Value", "Metadata", "Key",
	})
	table.SetAutoMergeCells(true)
	table.SetAutoWrapText(false)
	table.SetRowLine(true)

	for _, d := range dump {
		val := yamlTmpl(d.Value)
		var meta string
		if d.Metadata != nil {
			meta = yamlTmpl(d.Metadata)
		}
		var (
			name  = "-"
			model string
			orig  = d.Origin
		)
		if m, err := models.GetModelForKey(d.Key); err == nil {
			name, _ = m.ParseKey(d.Key)
			model = m.Name()
			if name == "" {
				name = d.Key
			}
		}
		val = fmt.Sprintf("# %s\n%s", d.Value.ProtoReflect().Descriptor().FullName(), val)
		var row []string
		row = []string{
			model,
			orig.String(),
			val,
			meta,
			name,
		}
		table.Append(row)
	}
	table.Render()
}

// formatDump is a helper type that can be used with user defined custom dump formats
type formatDump struct {
	Key      string
	Value    map[string]interface{}
	Metadata api.Metadata
	Origin   api.ValueOrigin
}

func convertDumps(in []api.RecordedKVWithMetadata) (out []formatDump, err error) {
	for _, d := range in {
		b, err := d.Value.MarshalJSON()
		if err != nil {
			return nil, err
		}
		var values map[string]interface{}
		if err = json.Unmarshal(b, &values); err != nil {
			return nil, err
		}
		// TODO: this "ProtoMsgData" string key has to be the same as the field name of
		// the ProtoWithName struct that contains the message data. ProtoWithName struct
		// is a part of kvschedulers internal utils package. Perhaps we could make this
		// field name a part of the public kvscheduler API so we do not have to rely
		// on string key here.
		if val, ok := values["ProtoMsgData"]; ok {
			out = append(out, formatDump{
				Key:      d.Key,
				Value:    val.(map[string]interface{}),
				Metadata: d.Metadata,
				Origin:   d.Origin,
			})
		}
	}
	return out, nil
}
