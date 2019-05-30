// Copyright (c) 2019 Cisco and/or its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package vpp-agent-init starts supervisor plugin managing other processes
// (VPP-Agent, VPP, ...)

package main

import (
	"github.com/ligato/cn-infra/agent"
	sv "github.com/ligato/cn-infra/exec/supervisor"
)

func main() {
	a := agent.NewAgent(agent.AllPlugins(&sv.DefaultPlugin))
	if err := a.Run(); err != nil {
		panic(err)
	}
}
