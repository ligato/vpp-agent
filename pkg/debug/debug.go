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

package debug

import (
	_ "expvar"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strings"

	"github.com/pkg/profile"
)

const defaultServerAddr = ":1234"

var (
	profileMode     = os.Getenv("DEBUG_PROFILE_MODE")
	profilePath     = os.Getenv("DEBUG_PROFILE_PATH")
	debugServerAddr = os.Getenv("DEBUG_SERVER_ADDR")
)

type Debug struct {
	closer func()
}

func Start() interface {
	Stop()
} {
	var d Debug

	d.runProfiling()

	d.runServer()

	return &d
}

func (d *Debug) Stop() {
	if d.closer != nil {
		d.closer()
	}
}

func (d *Debug) runProfiling() {
	var profiling func(*profile.Profile)

	switch strings.ToLower(profileMode) {
	case "cpu":
		profiling = profile.CPUProfile
	case "mem":
		profiling = profile.MemProfile
	case "mutex":
		profiling = profile.MutexProfile
	case "block":
		profiling = profile.BlockProfile
	case "trace":
		profiling = profile.TraceProfile
	default:
		// do nothing
		return
	}

	opts := []func(*profile.Profile){
		profiling,
		profile.ProfilePath(profilePath),
		profile.NoShutdownHook,
	}

	d.closer = profile.Start(opts...).Stop
}

func (d *Debug) runServer() {
	addr := debugServerAddr
	if addr == "" {
		addr = defaultServerAddr
	}

	log.Printf("debug server listening on: %s", addr)

	go func() {
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Printf("debug server error: %v", err)
		}
	}()
}
