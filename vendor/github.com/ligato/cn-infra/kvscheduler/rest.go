// Copyright (c) 2018 Cisco and/or its affiliates.
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

package kvscheduler

import (
	"errors"
	"net/http"
	"time"

	"github.com/ligato/cn-infra/rpc/rest"
	"github.com/unrolled/render"
	"strings"
)

const (
	// prefix used for REST urls of the scheduler.
	urlPrefix = "/scheduler/"

	// txnHistoryURL is URL used to obtain the transaction history.
	txnHistoryURL = urlPrefix + "txn-history"

	// verboseArg is the name of the argument used to enable/disable verbose
	// output for transaction history.
	verboseArg = "verbose"

	// sinceArg is the name of the argument used to define the start of the time
	// window for the transaction history to display.
	sinceArg = "since"

	// untilArg is the name of the argument used to define the end of the time
	// window for the transaction history to display.
	untilArg = "until"

	// keyTimelineURL is URL used to obtain timeline of value changes for a given key.
	keyTimelineURL = urlPrefix + "key-timeline"

	// keyArg is the name of the argument used to define key for "key-timeline" API.
	keyArg = "key"

	// graphSnapshotURL is URL used to obtain graph snapshot from a given point in time.
	graphSnapshotURL = urlPrefix + "graph-snapshot"

	// flagStatsURL is URL used to obtain flag statistics.
	flagStatsURL = urlPrefix + "flag-stats"

	// flagArg is the name of the argument used to define flag for "flag-stats" API.
	flagArg = "flag"

	// prefixArg is the name of the argument used to define prefix to filter keys
	// for "flag-stats" API.
	prefixArg = "prefix"

	// time is the name of the argument used to define point in time for a graph snapshot
	// to retrieve.
	timeArg = "time"
)

// registerHandlers registers all supported REST APIs.
func (scheduler *Scheduler) registerHandlers(http rest.HTTPHandlers) {
	if http == nil {
		scheduler.Log.Warn("No http handler provided, skipping registration of KVScheduler REST handlers")
		return
	}
	http.RegisterHTTPHandler(txnHistoryURL, scheduler.txnHistoryGetHandler, "GET")
	scheduler.Log.Infof("KVScheduler REST handler registered: GET %v", txnHistoryURL)
	http.RegisterHTTPHandler(keyTimelineURL, scheduler.keyTimelineGetHandler, "GET")
	scheduler.Log.Infof("KVScheduler REST handler registered: GET %v", keyTimelineURL)
	http.RegisterHTTPHandler(graphSnapshotURL, scheduler.graphSnapshotGetHandler, "GET")
	scheduler.Log.Infof("KVScheduler REST handler registered: GET %v", graphSnapshotURL)
	http.RegisterHTTPHandler(flagStatsURL, scheduler.flagStatsGetHandler, "GET")
	scheduler.Log.Infof("KVScheduler REST handler registered: GET %v", flagStatsURL)
}

// txnHistoryGetHandler is the GET handler for "txn-history" API.
func (scheduler *Scheduler) txnHistoryGetHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var since, until time.Time
		var verbose bool
		args := req.URL.Query()

		// parse optional *verbose* argument
		if verboseStr, withVerbose := args[verboseArg]; withVerbose && len(verboseStr) == 1 {
			verboseVal := verboseStr[0]
			if verboseVal == "true" || verboseVal == "1" {
				verbose = true
			}
		}

		// parse optional *until* argument
		if untilStr, withUntil := args[untilArg]; withUntil && len(untilStr) == 1 {
			var err error
			until, err = stringToTime(untilStr[0])
			if err != nil {
				formatter.JSON(w, http.StatusInternalServerError, err)
				return
			}
		}

		// parse optional *since* argument
		if sinceStr, withSince := args[sinceArg]; withSince && len(sinceStr) == 1 {
			var err error
			since, err = stringToTime(sinceStr[0])
			if err != nil {
				formatter.JSON(w, http.StatusInternalServerError, err)
				return
			}
		}

		txnHistory := scheduler.getTransactionHistory(since, until)
		formatter.Text(w, http.StatusOK, txnHistory.StringWithOpts(false, 0, verbose))
	}
}

// keyTimelineGetHandler is the GET handler for "key-timeline" API.
func (scheduler *Scheduler) keyTimelineGetHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		args := req.URL.Query()

		// parse mandatory *key* argument
		if keys, withKey := args[keyArg]; withKey && len(keys) == 1 {
			graphR := scheduler.graph.Read()
			defer graphR.Release()

			timeline := graphR.GetNodeTimeline(keys[0])
			formatter.JSON(w, http.StatusOK, timeline)
			return
		}

		err := errors.New("missing key argument")
		formatter.JSON(w, http.StatusInternalServerError, err)
		return
	}
}

// graphSnapshotGetHandler is the GET handler for "graph-snapshot" API.
func (scheduler *Scheduler) graphSnapshotGetHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		timeVal := time.Now()
		args := req.URL.Query()

		// parse optional *time* argument
		if timeStr, withTime := args[timeArg]; withTime && len(timeStr) == 1 {
			var err error
			timeVal, err = stringToTime(timeStr[0])
			if err != nil {
				formatter.JSON(w, http.StatusInternalServerError, err)
				return
			}
		}

		graphR := scheduler.graph.Read()
		defer graphR.Release()

		snapshot := graphR.GetSnapshot(timeVal)
		formatter.JSON(w, http.StatusOK, snapshot)
	}
}

// flagStatsGetHandler is the GET handler for "flag-stats" API.
func (scheduler *Scheduler) flagStatsGetHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		args := req.URL.Query()
		var prefixes []string

		// parse repeated *prefix* argument
		prefixes, _ = args[prefixArg]

		if flags, withFlag := args[flagArg]; withFlag && len(flags) == 1 {
			graphR := scheduler.graph.Read()
			defer graphR.Release()

			stats := graphR.GetFlagStats(flags[0], func(key string) bool {
				if len(prefixes) == 0 {
					return true
				}
				for _, prefix := range prefixes {
					if strings.HasPrefix(key, prefix) {
						return true
					}
				}
				return false
			})
			formatter.JSON(w, http.StatusOK, stats)
			return
		}

		err := errors.New("missing flag argument")
		formatter.JSON(w, http.StatusInternalServerError, err)
		return
	}
}
