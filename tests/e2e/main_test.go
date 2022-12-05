//  Copyright (c) 2020 Cisco and/or its affiliates.
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

package e2e

import (
	"flag"
	"log"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	govppcore "go.fd.io/govpp/core"

	"go.ligato.io/vpp-agent/v3/tests/e2e/e2etest"
	"go.ligato.io/vpp-agent/v3/tests/testutils"
)

func TestMain(m *testing.M) {
	debugFlag := flag.Bool("debug", false, "Turn on debug mode.")
	// Detect if GitHub Workflow debug flag is set.
	debugEnv := os.Getenv("RUNNER_DEBUG") != "" && os.Getenv("GITHUB_WORKFLOW") != ""
	log.SetFlags(log.Lmicroseconds | log.Lshortfile)
	flag.Parse()
	e2etest.Debug = *debugFlag || debugEnv
	if e2etest.Debug {
		logrus.SetLevel(logrus.DebugLevel)
		govppcore.SetLogLevel(logrus.DebugLevel)
	}
	if testutils.RunTestSuite("e2e") {
		result := m.Run()
		os.Exit(result)
	}
}
