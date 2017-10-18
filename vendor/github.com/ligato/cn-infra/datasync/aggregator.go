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

package datasync

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/ligato/cn-infra/utils/safeclose"
)

// CompositeKVProtoWatcher is an adapter that allows to aggregate multiple
// watchers (KeyValProtoWatcher) into one.
// Watch request is delegated into all of them.
type CompositeKVProtoWatcher struct {
	Adapters []KeyValProtoWatcher
}

// CompositeKVProtoWriter is an adapter that allows to aggregate multiple
// writers (KeyProtoValWriter) into one.
// Put request is delegated into all of them.
type CompositeKVProtoWriter struct {
	Adapters []KeyProtoValWriter
}

// AggregatedRegistration is adapter that allows to aggregate multiple
// registrations (WatchRegistration) into one.
// Close operation is applied collectively on all included registration.
type AggregatedRegistration struct {
	Registrations []WatchRegistration
}

// Watch subscribes to every transport available within transport aggregator.
// The function implements KeyValProtoWatcher.Watch().
func (ta *CompositeKVProtoWatcher) Watch(resyncName string, changeChan chan ChangeEvent,
	resyncChan chan ResyncEvent, keyPrefixes ...string) (WatchRegistration, error) {
	registrations := []WatchRegistration{}
	for _, transport := range ta.Adapters {
		watcherReg, err := transport.Watch(resyncName, changeChan, resyncChan, keyPrefixes...)
		if err != nil {
			return nil, err
		}

		if watcherReg != nil {
			registrations = append(registrations, watcherReg)
		}
	}

	return &AggregatedRegistration{
		Registrations: registrations,
	}, nil
}

// Put writes data to all aggregated transports.
// This function implements KeyProtoValWriter.Put().
func (ta *CompositeKVProtoWriter) Put(key string, data proto.Message, opts ...PutOption) error {
	if len(ta.Adapters) == 0 {
		return fmt.Errorf("No transport is available in aggregator")
	}
	var wasError error
	for _, transport := range ta.Adapters {
		err := transport.Put(key, data, opts...)
		if err != nil {
			wasError = err
		}
	}
	return wasError
}

// Close every registration under the aggregator.
// This function implements WatchRegistration.Close().
func (wa *AggregatedRegistration) Close() error {
	_, err := safeclose.CloseAll(wa.Registrations)
	return err
}
