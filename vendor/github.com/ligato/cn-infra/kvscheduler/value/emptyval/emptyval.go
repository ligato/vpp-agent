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

package emptyval

import (
	. "github.com/ligato/cn-infra/kvscheduler/api"
)

// emptyValue can be used whenever the mere existence of the value is the only
// information needed (typically Property values).
type emptyValue struct {
	valueType ValueType
}

// NewEmptyValue creates a new instance of empty value.
func NewEmptyValue(valueType ValueType) Value {
	return &emptyValue{valueType: valueType}
}

// Label returns empty string.
func (ev *emptyValue) Label() string {
	return ""
}

// Equivalent returns true for two empty values.
func (ev *emptyValue) Equivalent(v2 Value) bool {
	_, isEmpty := v2.(*emptyValue)
	if !isEmpty {
		return false
	}
	return true
}

// String returns empty string.
func (ev *emptyValue) String() string {
	return ""
}

// Type returns the type selected in NewEmptyValue constructor.
func (ev *emptyValue) Type() ValueType {
	return ev.valueType
}
