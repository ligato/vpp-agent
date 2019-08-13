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
	"github.com/ligato/cn-infra/datasync/kvdbsync"
	"github.com/ligato/cn-infra/datasync/resync"
	"github.com/ligato/cn-infra/logging"

	"github.com/ligato/vpp-agent/cmd/vpp-agent/app"
	"github.com/ligato/vpp-agent/pkg/debug"
)

const logo = `                                      __
 _  _____  ___ _______ ____ ____ ___ / /_
| |/ / _ \/ _ /___/ _ '/ _ '/ -_/ _ / __/
|___/ .__/ .__/   \_'_/\_' /\__/_//_\__/  %s
   /_/  /_/           /___/

`

var (
	debugEnabled = os.Getenv("DEBUG_ENABLED") != ""

	startTimeout = agent.DefaultStartTimeout
	stopTimeout  = agent.DefaultStopTimeout

	resyncTimeout = time.Second * 10
)

func main() {
	fmt.Fprintf(os.Stdout, logo, agent.BuildVersion)

	if debugEnabled {
		logging.DefaultLogger.SetLevel(logging.DebugLevel)
		logging.DefaultLogger.Debug("DEBUG ENABLED")
		defer debug.Start().Stop()
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
	// Overrides for start/stop timeouts of agent
	if t := os.Getenv("START_TIMEOUT"); t != "" {
		dur, err := time.ParseDuration(t)
		if err != nil {
			log.Fatalf("Invalid duration (%s) for start timeout!", t)
		} else {
			log.Printf("setting agent start timeout to: %v (via START_TIMEOUT)", dur)
			startTimeout = dur
		}
	}
	if t := os.Getenv("STOP_TIMEOUT"); t != "" {
		dur, err := time.ParseDuration(t)
		if err != nil {
			log.Fatalf("Invalid duration (%s) for stop timeout!", t)
		} else {
			log.Printf("setting agent stop timeout to: %v (via STOP_TIMEOUT)", dur)
			stopTimeout = dur
		}
	}

	// Override resync timeouts
	if t := os.Getenv("RESYNC_TIMEOUT"); t != "" {
		dur, err := time.ParseDuration(t)
		if err != nil {
			log.Fatalf("Invalid duration (%s) for resync timeout!", t)
		} else {
			log.Printf("setting resync timeout to: %v (via RESYNC_TIMEOUT)", dur)
			resyncTimeout = dur
		}
	}
	kvdbsync.ResyncDoneTimeout = resyncTimeout
	resync.SingleResyncAckTimeout = resyncTimeout
}
