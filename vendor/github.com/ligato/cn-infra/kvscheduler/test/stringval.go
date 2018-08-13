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

package test

import (
	. "github.com/ligato/cn-infra/kvscheduler/api"
)

// StringValue is used in the UTs.
type StringValue struct {
	valueType ValueType
	label     string
	str       string
}

// NewStringValue creates a new instance of StringValue.
func NewStringValue(valueType ValueType, label, str string) Value {
	return &StringValue{
		valueType: valueType,
		label:     label,
		str:       str,
		}
}

// Label returns value label as passed to the NewStringValue constructor.
func (sv *StringValue) Label() string {
	return sv.label
}

// Equivalent compares label, string and type.
func (sv *StringValue) Equivalent(v2 Value) bool {
	sv2, isStrVal := v2.(*StringValue)
	if !isStrVal {
		return false
	}
	return sv.valueType == sv2.valueType && sv.label == sv2.label && sv.str == sv2.str
}

// Label returns string representation as passed to the NewStringValue constructor.
func (sv *StringValue) String() string {
	return sv.str
}

// Type returns value type as chosen in the NewStringValue constructor.
func (sv *StringValue) Type() ValueType {
	return sv.valueType
}