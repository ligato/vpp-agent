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
	"os"

	"github.com/ligato/vpp-agent/cmd/agentctl/commands"
)

const logo = `
     ___                    __  ________  __
    /   | ____ ____  ____  / /_/ ____/ /_/ /
   / /| |/ __ '/ _ \/ __ \/ __/ /   / __/ / 
  / ___ / /_/ /  __/ / / / /_/ /___/ /_/ /  
 /_/  |_\__, /\___/_/ /_/\__/\____/\__/_/   
       /____/

`

func main() {
	cli := commands.NewAgentCli()

	cmd := commands.NewRootCommand(cli)
	cmd.Long = logo

	if err := cmd.Execute(); err != nil {
		os.Exit(-1)
	}
}
