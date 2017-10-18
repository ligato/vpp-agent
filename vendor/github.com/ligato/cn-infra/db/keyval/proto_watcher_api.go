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

package keyval

import (
	"time"

	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/logging/logroot"
)

// ProtoWatcher define API for monitoring changes in datastore.
// Changes are returned as protobuf/JSON-formatted data.
type ProtoWatcher interface {
	// Watch starts to monitor changes associated with the keys.
	// Watch events will be delivered to callback (not channel) <respChan>.
	Watch(respChan func(ProtoWatchResp), key ...string) error
}

// ProtoWatchResp represents a notification about data change.
// It is sent through the respChan callback.
type ProtoWatchResp interface {
	datasync.ChangeValue
	datasync.WithKey
}

// ToChanProto creates a callback that can be passed to the Watch function
// in order to receive JSON/protobuf-formatted notifications through a channel.
// If the notification can not be delivered until timeout, it is dropped.
func ToChanProto(ch chan ProtoWatchResp, opts ...interface{}) func(dto ProtoWatchResp) {

	timeout := datasync.DefaultNotifTimeout
	logger := logroot.StandardLogger()

	for _, opt := range opts {
		switch opt.(type) {
		case *core.WithLoggerOpt:
			logger = opt.(*core.WithLoggerOpt).Logger
		case *core.WithTimeoutOpt:
			timeout = opt.(*core.WithTimeoutOpt).Timeout
		}
	}

	return func(dto ProtoWatchResp) {
		select {
		case ch <- dto:
		case <-time.After(timeout):
			logger.Warn("Unable to deliver notification")
		}
	}
}
