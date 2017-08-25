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
	"time"
)

// PutOption defines options for Put operation. The particular options can be found below.
type PutOption interface {
	//PutOptionMark is just for marking implementation that it implements this interface
	PutOptionMark()
}

// DelOption defines options for Del operation. The particular options can be found below.
type DelOption interface {
	//DelOptionMark is just for marking implementation that it implements this interface
	DelOptionMark()
}

// WithTTLOpt defines a TTL for data being put. Once TTL elapses the data is removed from data store.
type WithTTLOpt struct {
	PutOptionMarker
	TTL time.Duration
}

// WithTTL creates new instance of TTL option. Once TTL elapses data is removed.
// Beware: some implementation might be using TTL with lower precision.
func WithTTL(TTL time.Duration) *WithTTLOpt {
	return &WithTTLOpt{TTL: TTL}
}

// WithPrefixOpt applies an operation to all items with the specified prefix.
type WithPrefixOpt struct {
	DelOptionMarker
}

// WithPrefix creates new instance of WithPrefixOpt.
func WithPrefix() *WithPrefixOpt {
	return &WithPrefixOpt{}
}

// PutOptionMarker is meant for anonymous composition in With*Opt structs
type PutOptionMarker struct{}

// PutOptionMark  is just for marking implementation that it implements this interface
func (marker *PutOptionMarker) PutOptionMark() {}

// DelOptionMarker is meant for anonymous composition in With*Opt structs
type DelOptionMarker struct{}

// DelOptionMark is just for marking implementation that it implements this interface
func (marker *DelOptionMarker) DelOptionMark() {}
