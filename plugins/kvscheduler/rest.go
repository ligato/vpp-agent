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
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"github.com/unrolled/render"

	"github.com/ligato/cn-infra/rpc/rest"
	kvs "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/kvscheduler/internal/graph"
)

const (
	// prefix used for REST urls of the scheduler.
	urlPrefix = "/scheduler/"

	// txnHistoryURL is URL used to obtain the transaction history.
	txnHistoryURL = urlPrefix + "txn-history"

	// sinceArg is the name of the argument used to define the start of the time
	// window for the transaction history to display.
	sinceArg = "since"

	// untilArg is the name of the argument used to define the end of the time
	// window for the transaction history to display.
	untilArg = "until"

	// seqNumArg is the name of the argument used to define the sequence number
	// of the transaction to display (txnHistoryURL).
	seqNumArg = "seq-num"

	// formatArg is the name of the argument used to set the output format
	// for the transaction history API.
	formatArg = "format"

	// recognized formats:
	formatJSON = "json"
	formatText = "text"

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

	// downstreamResyncURL is URL used to trigger downstream-resync.
	downstreamResyncURL = urlPrefix + "downstream-resync"

	// retryArg is the name of the argument used for "downstream-resync" API to tell whether
	// to retry failed operations or not.
	retryArg = "retry"

	// verboseArg is the name of the argument used for "downstream-resync" API
	// to tell whether the refreshed graph should be printed to stdout or not.
	verboseArg = "verbose"

	// dumpURL is URL used to dump either SB or scheduler's internal state of kv-pairs
	// under the given descriptor.
	dumpURL = urlPrefix + "dump"

	// descriptorArg is the name of the argument used to define descriptor for "dump" API.
	descriptorArg = "descriptor"

	// stateArg is the name of the argument used for "dump" API to tell whether
	// to dump "SB" (what there really is), "internal" state (what scheduler thinks
	// there is) or "NB" (the requested state). Default is to dump SB.
	stateArg = "state"

	/* recognized system states: */

	// SB = southbound (what there really is)
	SB = "SB"
	// internalState (scheduler's view of SB)
	internalState = "internal"
	// NB = northbound (the requested state)
	NB = "NB"
)

// errorString wraps string representation of an error that, unlike the original
// error, can be marshalled.
type errorString struct {
	Error string
}

// kvWithMetaForJSON is an internal extension to KVWithMetadata, with proto Message
// customized to implement MarshalJSON using jsonpb Marshaller.
// The jsonpb package produces a different output than the standard "encoding/json"
// package, which does not operate correctly on protocol buffers.
// On the other hand, the marshaller from jsonpb cannot handle anything other
// than proto messages.
type kvWithMetaForJSON struct {
	Key      string
	Value    protoMsgForJSON
	Metadata kvs.Metadata
	Origin   kvs.ValueOrigin
}

// protoMsgForJSON customizes proto.Message to implement MarshalJSON using
// the marshaller from jsonpb.
type protoMsgForJSON struct {
	proto.Message
}

// MarshalJSON marshalls proto message using the marshaller from jsonpb.
func (p *protoMsgForJSON) MarshalJSON() ([]byte, error) {
	marshaller := &jsonpb.Marshaler{}
	str, err := marshaller.MarshalToString(p.Message)
	if err != nil {
		return nil, err
	}
	return []byte(str), nil
}

// kvPairsForJSON converts a list of key-value pairs with metadata into an equivalent
// list of kvWithMetaForJSON.
func kvPairsForJSON(pairs []kvs.KVWithMetadata) (out []kvWithMetaForJSON) {
	for _, kv := range pairs {
		out = append(out, kvWithMetaForJSON{
			Key: kv.Key,
			Value: protoMsgForJSON{
				Message: kv.Value,
			},
			Metadata: kv.Metadata,
			Origin:   kv.Origin,
		})
	}
	return out
}

// registerHandlers registers all supported REST APIs.
func (scheduler *Scheduler) registerHandlers(http rest.HTTPHandlers) {
	if http == nil {
		scheduler.Log.Warn("No http handler provided, skipping registration of KVScheduler REST handlers")
		return
	}
	http.RegisterHTTPHandler(txnHistoryURL, scheduler.txnHistoryGetHandler, "GET")
	http.RegisterHTTPHandler(keyTimelineURL, scheduler.keyTimelineGetHandler, "GET")
	http.RegisterHTTPHandler(graphSnapshotURL, scheduler.graphSnapshotGetHandler, "GET")
	http.RegisterHTTPHandler(flagStatsURL, scheduler.flagStatsGetHandler, "GET")
	http.RegisterHTTPHandler(downstreamResyncURL, scheduler.downstreamResyncPostHandler, "POST")
	http.RegisterHTTPHandler(dumpURL, scheduler.dumpGetHandler, "GET")
}

// txnHistoryGetHandler is the GET handler for "txn-history" API.
func (scheduler *Scheduler) txnHistoryGetHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var since, until time.Time
		var seqNum int
		args := req.URL.Query()

		// parse optional *format* argument (default = JSON)
		format := formatJSON
		if formatStr, withFormat := args[formatArg]; withFormat && len(formatStr) == 1 {
			format = formatStr[0]
			if format != formatJSON && format != formatText {
				err := errors.New("unrecognized output format")
				formatter.JSON(w, http.StatusInternalServerError, errorString{err.Error()})
				return
			}
		}

		// parse optional *seq-num* argument
		if seqNumStr, withSeqNum := args[seqNumArg]; withSeqNum && len(seqNumStr) == 1 {
			var err error
			seqNum, err = strconv.Atoi(seqNumStr[0])
			if err != nil {
				scheduler.logError(formatter.JSON(w, http.StatusInternalServerError, errorString{err.Error()}))
				return
			}

			// sequence number takes precedence over the since-until time window
			txn := scheduler.getRecordedTransaction(uint(seqNum))
			if txn == nil {
				err := errors.New("transaction with such sequence is not recorded")
				scheduler.logError(formatter.JSON(w, http.StatusNotFound, errorString{err.Error()}))
				return
			}

			if format == formatJSON {
				scheduler.logError(formatter.JSON(w, http.StatusOK, txn))
			} else {
				scheduler.logError(formatter.Text(w, http.StatusOK, txn.StringWithOpts(false, 0)))
			}
			return
		}

		// parse optional *until* argument
		if untilStr, withUntil := args[untilArg]; withUntil && len(untilStr) == 1 {
			var err error
			until, err = stringToTime(untilStr[0])
			if err != nil {
				scheduler.logError(formatter.JSON(w, http.StatusInternalServerError, errorString{err.Error()}))
				return
			}
		}

		// parse optional *since* argument
		if sinceStr, withSince := args[sinceArg]; withSince && len(sinceStr) == 1 {
			var err error
			since, err = stringToTime(sinceStr[0])
			if err != nil {
				scheduler.logError(formatter.JSON(w, http.StatusInternalServerError, errorString{err.Error()}))
				return
			}
		}

		txnHistory := scheduler.getTransactionHistory(since, until)
		if format == formatJSON {
			scheduler.logError(formatter.JSON(w, http.StatusOK, txnHistory))
		} else {
			scheduler.logError(formatter.Text(w, http.StatusOK, txnHistory.StringWithOpts(false, 0)))
		}
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
			scheduler.logError(formatter.JSON(w, http.StatusOK, timeline))
			return
		}

		err := errors.New("missing key argument")
		scheduler.logError(formatter.JSON(w, http.StatusInternalServerError, errorString{err.Error()}))
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
				scheduler.logError(formatter.JSON(w, http.StatusInternalServerError, errorString{err.Error()}))
				return
			}
		}

		graphR := scheduler.graph.Read()
		defer graphR.Release()

		snapshot := graphR.GetSnapshot(timeVal)
		scheduler.logError(formatter.JSON(w, http.StatusOK, snapshot))
	}
}

// flagStatsGetHandler is the GET handler for "flag-stats" API.
func (scheduler *Scheduler) flagStatsGetHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		args := req.URL.Query()

		// parse repeated *prefix* argument
		prefixes := args[prefixArg]

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
			scheduler.logError(formatter.JSON(w, http.StatusOK, stats))
			return
		}

		err := errors.New("missing flag argument")
		scheduler.logError(formatter.JSON(w, http.StatusInternalServerError, errorString{err.Error()}))
	}
}

// downstreamResyncPostHandler is the POST handler for "downstream-resync" API.
func (scheduler *Scheduler) downstreamResyncPostHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		// parse optional *retry* argument
		args := req.URL.Query()
		retry := false
		if retryStr, withRetry := args[retryArg]; withRetry && len(retryStr) == 1 {
			retryVal := retryStr[0]
			if retryVal == "true" || retryVal == "1" {
				retry = true
			}
		}

		// parse optional *verbose* argument
		verbose := false
		if verboseStr, withVerbose := args[verboseArg]; withVerbose && len(verboseStr) == 1 {
			verboseVal := verboseStr[0]
			if verboseVal == "true" || verboseVal == "1" {
				verbose = true
			}
		}

		ctx := context.Background()
		ctx = kvs.WithResync(ctx, kvs.DownstreamResync, verbose)
		if retry {
			ctx = kvs.WithRetry(ctx, time.Second, true)
		}
		kvErrors, txnError := scheduler.StartNBTransaction().Commit(ctx)
		if txnError != nil {
			scheduler.logError(formatter.JSON(w, http.StatusInternalServerError, errorString{txnError.Error()}))
			return
		}
		if len(kvErrors) > 0 {
			kvErrorMap := make(map[string]errorString)
			for _, keyWithError := range kvErrors {
				kvErrorMap[keyWithError.Key] = errorString{keyWithError.Error.Error()}
			}
			scheduler.logError(formatter.JSON(w, http.StatusInternalServerError, kvErrorMap))
			return
		}
		scheduler.logError(formatter.Text(w, http.StatusOK, "SB was successfully synchronized with KVScheduler\n"))
	}
}

// dumpGetHandler is the GET handler for "dump" API.
func (scheduler *Scheduler) dumpGetHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		args := req.URL.Query()

		// parse mandatory *descriptor* argument
		descriptors, withDescriptor := args[descriptorArg]
		if !withDescriptor {
			err := errors.New("missing descriptor argument")
			scheduler.logError(formatter.JSON(w, http.StatusInternalServerError, errorString{err.Error()}))
			return
		}
		if len(descriptors) != 1 {
			err := errors.New("descriptor argument listed more than once")
			scheduler.logError(formatter.JSON(w, http.StatusInternalServerError, errorString{err.Error()}))
			return
		}
		descriptor := descriptors[0]

		// parse optional *state* argument (default = SB)
		state := SB
		if stateStr, withState := args[stateArg]; withState && len(stateStr) == 1 {
			state = stateStr[0]
			if state != SB && state != NB && state != internalState {
				err := errors.New("unrecognized system state")
				scheduler.logError(formatter.JSON(w, http.StatusInternalServerError, errorString{err.Error()}))
				return
			}
		}

		// pause transaction processing
		if state == SB {
			scheduler.txnLock.Lock()
			defer scheduler.txnLock.Unlock()
		}

		graphR := scheduler.graph.Read()
		defer graphR.Release()

		if state == NB {
			// dump the requested state
			var kvPairs []kvWithMetaForJSON
			nbNodes := graphR.GetNodes(nil,
				graph.WithFlags(&DescriptorFlag{descriptor}, &OriginFlag{kvs.FromNB}),
				graph.WithoutFlags(&DerivedFlag{}))

			for _, node := range nbNodes {
				lastChange := getNodeLastChange(node)
				if lastChange.value == nil {
					// value requested to be deleted
					continue
				}
				kvPairs = append(kvPairs, kvWithMetaForJSON{
					Key:    node.GetKey(),
					Value:  protoMsgForJSON{Message: lastChange.value},
					Origin: kvs.FromNB,
				})
			}
			scheduler.logError(formatter.JSON(w, http.StatusOK, kvPairs))
			return
		}

		/* internal/SB: */

		// dump from the in-memory graph first (for SB Dump it is used for correlation)
		inMemNodes := nodesToKVPairsWithMetadata(
			graphR.GetNodes(nil,
				graph.WithFlags(&DescriptorFlag{descriptor}),
				graph.WithoutFlags(&PendingFlag{}, &DerivedFlag{})))

		if state == internalState {
			// return the scheduler's view of SB for the given descriptor
			scheduler.logError(formatter.JSON(w, http.StatusOK, kvPairsForJSON(inMemNodes)))
			return
		}

		// obtain Dump handler from the descriptor
		kvDescriptor := scheduler.registry.GetDescriptor(descriptor)
		if kvDescriptor == nil {
			err := errors.New("descriptor is not registered")
			scheduler.logError(formatter.JSON(w, http.StatusInternalServerError, errorString{err.Error()}))
			return
		}
		if kvDescriptor.Dump == nil {
			err := errors.New("descriptor does not support Dump operation")
			scheduler.logError(formatter.JSON(w, http.StatusInternalServerError, errorString{err.Error()}))
			return
		}

		// dump the state directly from SB via descriptor
		dump, err := kvDescriptor.Dump(inMemNodes)
		if err != nil {
			scheduler.logError(formatter.JSON(w, http.StatusInternalServerError, errorString{err.Error()}))
			return
		}

		scheduler.logError(formatter.JSON(w, http.StatusOK, kvPairsForJSON(dump)))
		return
	}
}

// logError logs non-nil errors from JSON formatter
func (scheduler *Scheduler) logError(err error) {
	if err != nil {
		scheduler.Log.Error(err)
	}
}

// stringToTime converts Unix timestamp from string to time.Time.
func stringToTime(s string) (time.Time, error) {
	sec, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(sec, 0), nil
}
