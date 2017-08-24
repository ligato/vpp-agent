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

package idxmap

import (
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logroot"
	"time"
)

// DefaultNotifTimeout for delivery of notification
const DefaultNotifTimeout = 2 * time.Second

// ToChan creates a callback that can be passed to the Watch function in order to receive
// notifications through a channel. If the notification can not be delivered until timeout it is dropped.
func ToChan(ch chan NamedMappingGenericEvent, opts ...interface{}) func(dto NamedMappingGenericEvent) {

	timeout := DefaultNotifTimeout
	logger := logroot.StandardLogger()

	for _, opt := range opts {
		switch opt.(type) {
		case *WithLoggerOpt:
			logger = opt.(*WithLoggerOpt).logger
		case *WithTimeoutOpt:
			timeout = opt.(*WithTimeoutOpt).timeout
		}
	}

	return func(dto NamedMappingGenericEvent) {
		select {
		case ch <- dto:
		case <-time.After(timeout):
			logger.Warn("Unable to deliver notification")
		}
	}
}

// WithTimeoutOpt defines the maximum time that is attempted to deliver notification.
type WithTimeoutOpt struct {
	timeout time.Duration
}

// WithTimeout creates an option for ToChan function that defines a timeout for notification delivery.
func WithTimeout(timeout time.Duration) *WithTimeoutOpt {
	return &WithTimeoutOpt{timeout: timeout}
}

// WithLoggerOpt defines a logger that logs if delivery of notification is unsuccessful.
type WithLoggerOpt struct {
	logger logging.Logger
}

// WithLogger creates an option for ToChan function that specifies a logger to be used.
func WithLogger(logger logging.Logger) *WithLoggerOpt {
	return &WithLoggerOpt{logger: logger}
}
