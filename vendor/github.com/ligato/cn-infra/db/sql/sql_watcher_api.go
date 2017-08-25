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

package sql

import "github.com/ligato/cn-infra/datasync"

// Watcher define API for monitoring changes in a datastore
type Watcher interface {
	// Watch starts to monitor changes in data store. Watch events will be delivered to the callback.
	Watch(callback func(WatchResp), statement ...string) error
}

// WatchResp represents a notification about change. It is sent through the watch resp channel.
type WatchResp interface {
	GetChangeType() datasync.PutDel
	// GetValue returns the value in the event
	GetValue(outBinding interface{}) error
}

// ToChan TODO
func ToChan(respChan chan WatchResp, options ...interface{}) func(event WatchResp) {
	return func(WatchResp) {
		/*select {
		case respChan <- resp:
		case <-time.After(defaultOpTimeout):
			log.Warn("Unable to deliver watch event before timeout.")
		}

		select {
		case wresp := <-recvChan:
			for _, ev := range wresp.Events {
				handleWatchEvent(respChan, ev)
			}
		case <-closeCh:
			log.WithField("key", key).Debug("Watch ended")
			return
		}*/
	}
}
