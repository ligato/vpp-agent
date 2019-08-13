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
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ligato/cn-infra/agent"
	"github.com/ligato/cn-infra/logging"

	"github.com/ligato/vpp-agent/cmd/vpp-agent/app"
)

const logo = `                                      __
 _  _____  ___ _______ ____ ____ ___ / /_
| |/ / _ \/ _ /___/ _ '/ _ '/ -_/ _ / __/
|___/ .__/ .__/   \_'_/\_' /\__/_//_\__/  %s
   /_/  /_/           /___/

`

var (
	startTimeout = agent.DefaultStartTimeout
	stopTimeout  = agent.DefaultStopTimeout
	// FIXME: global flags are not currently compatible with config package (flags are parsed in NewAgent)
	//startTimeout = flag.Duration("start-timeout", agent.DefaultStartTimeout, "Timeout duration for starting agent.")
	//stopTimeout  = flag.Duration("stop-timeout", agent.DefaultStopTimeout, "Timeout duration for stopping agent.")
)

var debugging func() func()

func main() {
	fmt.Fprintf(os.Stdout, logo, agent.BuildVersion)

	if debugging != nil {
		defer debugging()()
	}

	vppAgent := app.New()
	a := agent.NewAgent(
		agent.AllPlugins(vppAgent),
		agent.StartTimeout(startTimeout),
		agent.StopTimeout(stopTimeout),
	)

	if err := a.Run(); err != nil {
		logging.DefaultLogger.Fatal(err)
	}
}

func init() {
	logging.DefaultLogger.SetOutput(os.Stdout)
	logging.DefaultLogger.SetLevel(logging.DebugLevel)
	// Setup start/stop timeouts for agent
	if t := os.Getenv("START_TIMEOUT"); t != "" {
		dur, err := time.ParseDuration(t)
		if err != nil {
			log.Fatalf("Invalid duration (%s) for start timeout!", t)
		} else {
			startTimeout = dur
		}
	}
	if t := os.Getenv("STOP_TIMEOUT"); t != "" {
		dur, err := time.ParseDuration(t)
		if err != nil {
			log.Fatalf("Invalid duration (%s) for stop timeout!", t)
		} else {
			stopTimeout = dur
		}
	}
}
