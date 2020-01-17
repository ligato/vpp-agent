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

package cli

import (
	"io"

	"github.com/docker/cli/cli/streams"
	"github.com/docker/docker/pkg/term"

	"go.ligato.io/vpp-agent/v3/cmd/agentctl/client"
)

// AgentCliOption applies a modification on a AgentCli.
type AgentCliOption func(cli *AgentCli) error

// WithStandardStreams sets a cli in, out and err streams with the standard streams.
func WithStandardStreams() AgentCliOption {
	return func(cli *AgentCli) error {
		// Set terminal emulation based on platform as required.
		stdin, stdout, stderr := term.StdStreams()
		cli.in = streams.NewIn(stdin)
		cli.out = streams.NewOut(stdout)
		cli.err = stderr
		return nil
	}
}

// WithCombinedStreams uses the same stream for the output and error streams.
func WithCombinedStreams(combined io.Writer) AgentCliOption {
	return func(cli *AgentCli) error {
		cli.out = streams.NewOut(combined)
		cli.err = combined
		return nil
	}
}

// WithInputStream sets a cli input stream.
func WithInputStream(in io.ReadCloser) AgentCliOption {
	return func(cli *AgentCli) error {
		cli.in = streams.NewIn(in)
		return nil
	}
}

// WithOutputStream sets a cli output stream.
func WithOutputStream(out io.Writer) AgentCliOption {
	return func(cli *AgentCli) error {
		cli.out = streams.NewOut(out)
		return nil
	}
}

// WithErrorStream sets a cli error stream.
func WithErrorStream(err io.Writer) AgentCliOption {
	return func(cli *AgentCli) error {
		cli.err = err
		return nil
	}
}

// WithClient sets an APIClient.
func WithClient(c client.APIClient) AgentCliOption {
	return func(cli *AgentCli) error {
		cli.client = c
		return nil
	}
}
