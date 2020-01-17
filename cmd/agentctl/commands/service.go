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
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/jhump/protoreflect/grpcreflect"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	ref "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
	"google.golang.org/grpc/status"

	agentcli "go.ligato.io/vpp-agent/v3/cmd/agentctl/cli"
)

func NewServiceCommand(cli agentcli.Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "service",
		Short: "Manage agent services",
	}
	cmd.AddCommand(
		NewServiceListCommand(cli),
		NewServiceCallCommand(cli),
	)
	return cmd
}

func NewServiceListCommand(cli agentcli.Cli) *cobra.Command {
	var (
		opts ServiceListOptions
	)
	cmd := &cobra.Command{
		Use:     "list [SERVICE]",
		Aliases: []string{"ls", "l"},
		Short:   "List remote services",
		Args:    cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServiceList(cli, opts)
		},
	}
	flags := cmd.Flags()
	flags.BoolVarP(&opts.Methods, "methods", "m", false, "Show service methods")
	return cmd
}

type ServiceListOptions struct {
	Methods bool
}

func runServiceList(cli agentcli.Cli, opts ServiceListOptions) error {
	grpcClient, err := cli.Client().GRPCConn()
	if err != nil {
		return err
	}
	ctx := context.Background()
	c := grpcreflect.NewClient(ctx, ref.NewServerReflectionClient(grpcClient))

	services, err := c.ListServices()
	if err != nil {
		msg := status.Convert(err).Message()
		return errors.Wrapf(err, "listing services failed: %v", msg)
	}

	for _, srv := range services {
		if srv == reflectionServiceName {
			continue
		}
		s, err := c.ResolveService(srv)
		if err != nil {
			return fmt.Errorf("resolving service %s failed: %v", srv, err)
		}
		fmt.Fprintf(cli.Out(), "service %s (%v)\n", s.GetFullyQualifiedName(), s.GetFile().GetName())
		if opts.Methods {
			for _, m := range s.GetMethods() {
				req := m.GetInputType().GetName()
				if m.IsClientStreaming() {
					req = fmt.Sprintf("stream %s", req)
				}
				resp := m.GetOutputType().GetName()
				if m.IsServerStreaming() {
					resp = fmt.Sprintf("stream %s", resp)
				}
				fmt.Fprintf(cli.Out(), " - rpc %s (%v) returns (%v)\n", m.GetName(), req, resp)
			}
			fmt.Fprintln(cli.Out())
		}
	}

	return nil
}

const reflectionServiceName = "grpc.reflection.v1alpha.ServerReflection"

func NewServiceCallCommand(cli agentcli.Cli) *cobra.Command {
	var (
		opts ServiceCallOptions
	)
	cmd := &cobra.Command{
		Use:     "call SERVICE METHOD",
		Aliases: []string{"c"},
		Short:   "Call methods on services",
		Args:    cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Service = args[0]
			opts.Method = args[1]
			return runServiceCall(cli, opts)
		},
	}
	return cmd
}

type ServiceCallOptions struct {
	Service string
	Method  string
}

func runServiceCall(cli agentcli.Cli, opts ServiceCallOptions) error {
	grpcClient, err := cli.Client().GRPCConn()
	if err != nil {
		return err
	}
	c := grpcreflect.NewClient(context.Background(), ref.NewServerReflectionClient(grpcClient))

	svc, err := c.ResolveService(opts.Service)
	if err != nil {
		msg := status.Convert(err).Message()
		return errors.Wrapf(err, "resolving service failed: %v", msg)
	}

	m := svc.FindMethodByName(opts.Method)
	if m == nil {
		return fmt.Errorf("method %s not found for service %s", opts.Method, svc.GetName())
	}

	endpoint, err := fqrnToEndpoint(m.GetFullyQualifiedName())
	if err != nil {
		return err
	}

	req := dynamic.NewMessage(m.GetInputType())
	reply := dynamic.NewMessage(m.GetOutputType())

	if len(m.GetInputType().GetFields()) > 0 {
		fmt.Fprintf(cli.Out(), "Enter input request %s:\n", m.GetInputType().GetFullyQualifiedName())

		var buf bytes.Buffer
		_, err = buf.ReadFrom(cli.In())
		if err != nil {
			return err
		}
		b, err := yaml.YAMLToJSON(buf.Bytes())
		if err != nil {
			return err
		}
		if err = json.Unmarshal(b, req); err != nil {
			return err
		}
		input, err := req.MarshalTextIndent()
		if err != nil {
			return err
		}
		fmt.Fprintf(cli.Out(), "Request (%s):\n%s\n", m.GetInputType().GetName(), input)
	}

	fmt.Fprintf(cli.Out(), "Calling %s\n", m.GetFullyQualifiedName())

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*7)
	defer cancel()
	if err := grpcClient.Invoke(ctx, endpoint, req, reply); err != nil {
		return err
	}

	output, err := reply.MarshalTextIndent()
	if err != nil {
		return err
	}
	fmt.Fprintf(cli.Out(), "Response (%s):\n%s\n", m.GetOutputType().GetName(), output)

	return nil
}

// fqrnToEndpoint converts FullQualifiedRPCName to endpoint
//
// e.g.
//	pkg_name.svc_name.rpc_name -> /pkg_name.svc_name/rpc_name
func fqrnToEndpoint(fqrn string) (string, error) {
	sp := strings.Split(fqrn, ".")
	if len(sp) < 3 {
		return "", errors.New("invalid FQRN format")
	}

	return fmt.Sprintf("/%s/%s", strings.Join(sp[:len(sp)-1], "."), sp[len(sp)-1]), nil
}
