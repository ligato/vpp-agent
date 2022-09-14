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

package govppmux

import (
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	govppapi "go.fd.io/govpp/api"
)

// Set of raw Prometheus metrics.
// Labels
// * message
// * error
// Do not increment directly, use Report* methods.
var (
	channelsCreated = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "ligato",
		Subsystem: "govppmux",
		Name:      "channels_created_total",
		Help:      "The total number of created channels.",
	})
	channelsCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "ligato",
		Subsystem: "govppmux",
		Name:      "channels",
		Help:      "The current number of opened channels.",
	})
	requestsSent = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "ligato",
		Subsystem: "govppmux",
		Name:      "requests_total",
		Help:      "The total number of sent requests.",
	},
		[]string{"message"},
	)
	requestsCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "ligato",
		Subsystem: "govppmux",
		Name:      "requests",
		Help:      "The current number of in-flight requests.",
	})
	requestsFailed = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "ligato",
		Subsystem: "govppmux",
		Name:      "requests_failed_total",
		Help:      "The total number of failed requests.",
	},
		[]string{"message", "error"},
	)
	requestsDone = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "ligato",
		Subsystem: "govppmux",
		Name:      "requests_done_total",
		Help:      "The total number of done requests.",
	},
		[]string{"message"},
	)
	repliesReceived = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "ligato",
		Subsystem: "govppmux",
		Name:      "replies_received_total",
		Help:      "The total number of received replies.",
	},
		[]string{"message"},
	)
	successfulRequestHandlingSec = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "ligato",
		Subsystem: "govppmux",
		Name:      "successful_request_duration_seconds",
		Help:      "Bucketed histogram of processing time of successfully handled requests by message name.",
		// lowest bucket start of upper bound 0.0005 sec (0.5 ms) with factor 2
		// highest bucket start of 0.0005 sec * 2^12 == 2.048 sec
		Buckets: prometheus.ExponentialBuckets(0.0005, 2, 13),
	},
		[]string{"message"},
	)
)

func init() {
	prometheus.MustRegister(channelsCreated)
	prometheus.MustRegister(channelsCount)
	prometheus.MustRegister(requestsSent)
	prometheus.MustRegister(requestsCount)
	prometheus.MustRegister(requestsDone)
	prometheus.MustRegister(requestsFailed)
	prometheus.MustRegister(repliesReceived)
	prometheus.MustRegister(successfulRequestHandlingSec)
}

func reportChannelsOpened() {
	channelsCreated.Inc()
	channelsCount.Inc()

	if DisableOldStats {
		return
	}
	atomic.AddUint64(&stats.ChannelsCreated, 1)
	atomic.AddUint64(&stats.ChannelsOpen, 1)
}

func reportChannelsClosed() {
	channelsCount.Dec()

	if DisableOldStats {
		return
	}
	atomic.AddUint64(&stats.ChannelsOpen, ^uint64(0)) // decrement
}

func reportRequestSent(request govppapi.Message) {
	requestsCount.Inc()
	requestsSent.WithLabelValues(request.GetMessageName()).Inc()

	if DisableOldStats {
		return
	}
	atomic.AddUint64(&stats.RequestsSent, 1)
}

func reportRequestFailed(request govppapi.Message, err error) {
	requestsCount.Dec()
	requestsFailed.WithLabelValues(request.GetMessageName(), err.Error()).Inc()

	if DisableOldStats {
		return
	}
	trackError(err.Error())
	atomic.AddUint64(&stats.RequestsFail, 1)
}

func reportRequestSuccess(request govppapi.Message, startTime time.Time) {
	took := time.Since(startTime)
	requestsCount.Dec()
	requestsDone.WithLabelValues(request.GetMessageName()).Inc()
	successfulRequestHandlingSec.WithLabelValues(request.GetMessageName()).Observe(took.Seconds())

	if DisableOldStats {
		return
	}
	atomic.AddUint64(&stats.RequestsDone, 1)
	trackMsgRequestDur(request.GetMessageName(), took)
}

func reportRepliesReceived(reply govppapi.Message) {
	repliesReceived.WithLabelValues(reply.GetMessageName()).Inc()

	if DisableOldStats {
		return
	}
	atomic.AddUint64(&stats.RequestReplies, 1)
	trackMsgReply(reply.GetMessageName())
}
