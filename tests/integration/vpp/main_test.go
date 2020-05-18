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

package vpp

import (
	"flag"
	"log"
	"os"
	"testing"

	govppcore "git.fd.io/govpp.git/core"
	"github.com/sirupsen/logrus"
)

var (
	vppPath     = flag.String("vpp-path", "/usr/bin/vpp", "VPP program path")
	vppConfig   = flag.String("vpp-config", "", "VPP config file")
	vppSockAddr = flag.String("vpp-sock-addr", "/run/vpp/api.sock", "VPP binapi socket address")
	vppRetry    = flag.Uint("retry", 3, "Number of VPP connect retries")

	debug = flag.Bool("debug", false, "Turn on debug mode.")
)

func TestMain(m *testing.M) {
	log.SetFlags(log.Lmicroseconds | log.Lshortfile)
	flag.Parse()
	if *debug {
		govppcore.SetLogLevel(logrus.DebugLevel)
	}
	result := m.Run()
	os.Exit(result)
}
