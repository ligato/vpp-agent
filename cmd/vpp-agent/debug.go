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

// +build !nodebug

package main

import (
	_ "expvar"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
)

var (
	debugEnabled    = os.Getenv("DEBUG_ENABLED") != ""
	debugServerAddr = os.Getenv("DEBUG_SERVERADDR")
	cpuprofile      = os.Getenv("DEBUG_CPUPROFILE")
	memprofile      = os.Getenv("DEBUG_MEMPROFILE")
	traceprofile    = os.Getenv("DEBUG_TRACEPROFILE")
)

func init() {
	if debugEnabled {
		go debugServer()
		debugging = debug
	}
}

func debugServer() {
	addr := debugServerAddr
	if addr == "" {
		addr = ":1234"
	}
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Printf("debug server error: %v", err)
	}
}

func debug() func() {
	/*trace.AuthRequest = func(req *http.Request) (any, sensitive bool) {
		return true, true
	}*/
	var err error

	var cpuFile *os.File
	if cpuprofile != "" {
		cpuFile, err = os.Create(cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		if err := pprof.StartCPUProfile(cpuFile); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
	}

	var traceFile *os.File
	if traceprofile != "" {
		traceFile, err = os.Create(traceprofile)
		if err != nil {
			log.Fatalf("failed to create trace output file: %v", err)
		}
		if err := trace.Start(traceFile); err != nil {
			log.Fatalf("failed to start trace: %v", err)
		}
	}

	return func() {
		if traceFile != nil {
			trace.Stop()
			log.Printf("closing trace profile file: %s", traceFile.Name())
			if err := traceFile.Close(); err != nil {
				log.Fatalf("failed to close trace file: %v", err)
			}
		}
		if cpuFile != nil {
			pprof.StopCPUProfile()
			log.Printf("closing CPU profile file: %s", cpuFile.Name())
			if err := cpuFile.Close(); err != nil {
				log.Printf("closing failed: %v", err)
			}
		}
		if memprofile != "" {
			f, err := os.Create(memprofile)
			if err != nil {
				log.Fatal("could not create memory profile: ", err)
			}
			defer func() {
				log.Printf("closing memory profile: %s", memprofile)
				if err := f.Close(); err != nil {
					log.Printf("closing failed: %v", err)
				}
			}()
			runtime.GC() // get up-to-date statistics
			if err := pprof.WriteHeapProfile(f); err != nil {
				log.Fatal("could not write memory profile: ", err)
			}
		}
	}
}
