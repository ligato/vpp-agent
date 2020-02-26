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

package kvscheduler

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"

	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
)

// Set of raw Prometheus metrics.
// Labels
// * txn_type
// * slice
// Do not increment directly, use Report* methods.
var (
	transactionsProcessed = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "ligato",
		Subsystem: "kvscheduler",
		Name:      "txn_processed",
		Help:      "The total number of transactions processed.",
	})
	transactionsDropped = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "ligato",
		Subsystem: "kvscheduler",
		Name:      "txn_dropped",
		Help:      "The total number of transactions dropped.",
	})
	queueCapacity = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "ligato",
		Subsystem: "kvscheduler",
		Name:      "queue_capacity",
		Help:      "The capacity of the transactions queue.",
	})
	queueLength = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "ligato",
		Subsystem: "kvscheduler",
		Name:      "queue_length",
		Help:      "The number of transactions in the queue.",
	})
	queueWaitSeconds = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: "ligato",
		Subsystem: "kvscheduler",
		Name:      "queue_wait_seconds",
		Help:      "Wait time in queue for transactions.",
		MaxAge:    time.Second * 30,
	},
		[]string{"txn_type"},
	)
	txnProcessDurationSeconds = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: "ligato",
		Subsystem: "kvscheduler",
		Name:      "txn_process_duration_seconds",
		Help:      "Processing time of transactions.",
		MaxAge:    time.Second * 30,
	},
		[]string{"slice"},
	)
	txnDurationSeconds = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "ligato",
		Subsystem: "kvscheduler",
		Name:      "txn_duration_seconds",
		Help:      "Bucketed histogram of processing time of transactions by type.",
	},
		[]string{"txn_type"},
	)
)

func init() {
	prometheus.MustRegister(transactionsProcessed)
	prometheus.MustRegister(transactionsDropped)
	prometheus.MustRegister(queueCapacity)
	prometheus.MustRegister(queueLength)
	prometheus.MustRegister(queueWaitSeconds)
	prometheus.MustRegister(txnProcessDurationSeconds)
	prometheus.MustRegister(txnDurationSeconds)
}

func reportTxnProcessed(typ kvs.TxnType, sec float64) {
	transactionsProcessed.Inc()
	txnDurationSeconds.WithLabelValues(typ.String()).Observe(sec)
}

func reportTxnDropped() {
	transactionsDropped.Inc()
}

func reportQueueCap(c int) {
	queueCapacity.Set(float64(c))
}

func reportQueued(n int) {
	queueLength.Add(float64(n))
}

func reportQueueWait(typ kvs.TxnType, sec float64) {
	queueWaitSeconds.WithLabelValues(typ.String()).Observe(sec)
}

func reportTxnProcessDuration(slice string, sec float64) {
	txnProcessDurationSeconds.WithLabelValues(slice).Observe(sec)
}
