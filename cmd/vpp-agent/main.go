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

// Package vpp-agent implements the main entry point into the VPP Agent
// and it is used to build the VPP Agent executable.
package main

import (
	"os"
	"time"

	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/logging"
	log "github.com/ligato/cn-infra/logging/logrus"
	vpp_flavor "github.com/ligato/vpp-agent/flavors/vpp"
)

// main is the main entry point into the VPP Agent. First, a new CN-Infra
// Agent (app) is created using the set of plugins defined in vpp_flavor
// (../../flavors/vpp). Second, the function calls EventLoopWithInterrupt()
// which initializes and starts all plugins and then waits for the user
// to terminate the VPP Agent process with SIGINT. All VPP Agent's work between
// the initialization and termination is performed by the plugins.
func main() {

	f := vpp_flavor.Flavor{}
	agent := core.NewAgent(log.DefaultLogger(), 15*time.Second, f.Plugins()...)

	err := core.EventLoopWithInterrupt(agent, nil)
	if err != nil {
		os.Exit(1)
	}
}

// init sets the Log output and Log level parameters for VPP Agent's default
// logger.
func init() {
	log.DefaultLogger().SetOutput(os.Stdout)
	log.DefaultLogger().SetLevel(logging.DebugLevel)
}
