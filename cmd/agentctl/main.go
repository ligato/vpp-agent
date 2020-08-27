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

package main

import (
	"fmt"
	"os"

	agentcli "go.ligato.io/vpp-agent/v3/cmd/agentctl/cli"
	"go.ligato.io/vpp-agent/v3/cmd/agentctl/commands"
	"go.ligato.io/vpp-agent/v3/pkg/version"
)

const logo = `
                      __      __  __
  ___ ____ ____ ___  / /_____/ /_/ /
 / _ '/ _ '/ -_) _ \/ __/ __/ __/ / 
 \_,_/\_, /\__/_//_/\__/\__/\__/_/  
     /___/
`

func runAgentctl(cli *agentcli.AgentCli) error {
	cmd, err := commands.NewRootCommand(cli)
	if err != nil {
		return err
	}
	cmd.Long = logo
	cmd.Version = version.Version()
	return cmd.Execute()
}

func main() {
	cli := commands.NewAgentCli()

	if err := runAgentctl(cli); err != nil {
		fmt.Fprintf(cli.Err(), "\nERROR: %v\n", err)
		os.Exit(commands.ExitCode(err))
	}
}
