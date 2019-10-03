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

	"github.com/jhump/protoreflect/grpcreflect"
	evans_grpc "github.com/ktr0731/evans/grpc"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	ref "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
	"google.golang.org/grpc/status"

	agentcli "github.com/ligato/vpp-agent/cmd/agentctl/cli"
)

func NewServiceCommand(cli agentcli.Cli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "service",
		Short: "Manage agent services",
	}
	cmd.AddCommand(
		NewServiceListCommand(cli),
		//NewServiceCallCommand(cli),
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
			opts.Services = args
			return runServiceList(cli, opts)
		},
	}
	flags := cmd.Flags()
	flags.BoolVarP(&opts.Methods, "methods", "m", false, "Show service methods")
	return cmd
}

type ServiceListOptions struct {
	Services []string
	Methods  bool
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
		return errors.Wrapf(err, "failed to list services: %v", msg)
	}

	for _, srv := range services {
		if srv == reflectionServiceName {
			continue
		}
		s, err := c.ResolveService(srv)
		if err != nil {
			return fmt.Errorf("resolving service failed: %v", err)
		}
		fmt.Fprintf(cli.Out(), "%s (%v)\n", s.GetFullyQualifiedName(), s.GetFile().GetName())
		if opts.Methods {
			for _, m := range s.GetMethods() {
				fmt.Fprintf(cli.Out(), " - %s (%v) %v\n", m.GetName(), m.GetInputType().GetName(), m.GetOutputType().GetName())
			}
		}
		fmt.Fprintln(cli.Out())
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
		Short:   "Call remote services",
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
	/*grpcClient, err := cli.Client().GRPCConn()
	if err != nil {
		return err
	}*/
	addr, _ := cli.Client().GRPCAddr()

	c, err := evans_grpc.NewClient(addr, "", true, false, "", "", "")
	if err != nil {
		return err
	}

	pkgs, err := c.ListPackages()
	if err != nil {
		return err
	}

	for _, pkg := range pkgs {
		fmt.Fprintf(cli.Out(), " - %v", pkg)
	}

	//c.Invoke()

	/*ctx := context.Background()
	c := grpcreflect.NewClient(ctx, ref.NewServerReflectionClient(grpcClient))

	svc, err := c.ResolveService(opts.Service)
	if err != nil {
		msg := status.Convert(err).Message()
		return errors.Wrapf(err, "resolving service failed: %v", msg)
	}

	m := svc.FindMethodByName(opts.Method)
	if m == nil {
		return fmt.Errorf("method %s not found for service %s", opts.Method, svc.GetName())
	}*/

	return nil
}
