// Copyright (c) 2017 Cisco and/or its affiliates.
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

package main

import (
	"os"
	"time"

	"github.com/ligato/cn-infra/datasync/grpcsync/examplesple/benchmark"

	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/examples/simple-agent/generic"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logroot"
)

func main() {
	logroot.StandardLogger().SetLevel(logging.DebugLevel)

	err := benchmark.Run()

	f := generic.Flavour{}
	agent := core.NewAgent(logroot.StandardLogger(), 15*time.Second, f.Plugins()...)

	err := core.EventLoopWithInterrupt(agent, nil)
	if err != nil {
		os.Exit(1)
	}
}
