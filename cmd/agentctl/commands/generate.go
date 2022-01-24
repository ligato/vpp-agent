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
	"strings"

	"github.com/ghodss/yaml"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/prototext"

	"go.ligato.io/vpp-agent/v3/cmd/agentctl/api/types"
	agentcli "go.ligato.io/vpp-agent/v3/cmd/agentctl/cli"
)

const defaultIndent = "  "

func NewGenerateCommand(cli agentcli.Cli) *cobra.Command {
	var (
		opts GenerateOptions
	)
	cmd := &cobra.Command{
		Use:     "generate MODEL",
		Aliases: []string{"gen"},
		Short:   "Generate config samples",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Model = args[0]
			return runGenerate(cli, opts)
		},
	}
	flags := cmd.Flags()
	flags.StringVarP(&opts.Format, "format", "f", "json",
		"Output formats: json, yaml")
	flags.BoolVar(&opts.OneLine, "oneline", false,
		"Print output as single line (only json format)")
	return cmd
}

type GenerateOptions struct {
	Model   string
	Format  string
	OneLine bool
}

func runGenerate(cli agentcli.Cli, opts GenerateOptions) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	allModels, err := cli.Client().ModelList(ctx, types.ModelListOptions{
		Class: "config",
	})
	if err != nil {
		return err
	}

	modelList := filterModelsByRefs(allModels, []string{opts.Model})

	if len(modelList) == 0 {
		return fmt.Errorf("no model found for: %s", opts.Model)
	}

	logrus.Debugf("models: %+v", modelList)
	model := modelList[0]

	valueType := protoMessageType(model.ProtoName)
	if valueType == nil {
		return fmt.Errorf("unknown proto message defined for: %s", model.ProtoName)
	}
	modelInstance := valueType.New().Interface()

	var out string

	switch strings.ToLower(opts.Format) {
	case "j", "json":
		m := protojson.MarshalOptions{
			UseEnumNumbers:  false,
			EmitUnpopulated: true,
			AllowPartial:    true,
			Indent:          defaultIndent,
			UseProtoNames:   true,
			Resolver:        nil,
		}
		if opts.OneLine {
			m.Indent = ""
		}
		b, err := m.Marshal(modelInstance)
		if err != nil {
			return fmt.Errorf("encoding to json failed: %v", err)
		}
		out = string(b)
	case "y", "yaml":
		m := protojson.MarshalOptions{
			UseEnumNumbers:  false,
			AllowPartial:    true,
			EmitUnpopulated: true,
			Indent:          defaultIndent,
			UseProtoNames:   true,
			Resolver:        nil,
		}
		if opts.OneLine {
			m.Indent = ""
		}
		b, err := m.Marshal(modelInstance)
		if err != nil {
			return fmt.Errorf("encoding to json failed: %v", err)
		}
		b, err = yaml.JSONToYAML(b)
		if err != nil {
			return fmt.Errorf("encoding to yaml failed: %v", err)
		}
		out = string(b)
	case "p", "proto":
		m := prototext.MarshalOptions{
			AllowPartial: true,
			Indent:       "  ",
			Resolver:     nil,
		}
		b, err := m.Marshal(modelInstance)
		if err != nil {
			return fmt.Errorf("encoding to proto text failed: %v", err)
		}
		out = string(b)
	default:
		return fmt.Errorf("unknown format: %s", opts.Format)
	}

	fmt.Fprintf(cli.Out(), "%s\n", out)
	return nil
}
