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

// ArrayValue is used in the UTs.
type ArrayValue struct {
	valueType ValueType
	label     string
	items     []string
}

// NewArrayValue creates a new instance of ArrayValue.
func NewArrayValue(valueType ValueType, label string, items ...string) Value {
	return &ArrayValue{
		valueType: valueType,
		label:     label,
		items:     items,
	}
}

// Label returns value label as passed to the NewArrayValue constructor.
func (av *ArrayValue) Label() string {
	return av.label
}

// Equivalent compares label, array of items and type.
func (av *ArrayValue) Equivalent(v2 Value) bool {
	av2, isArrayVal := v2.(*ArrayValue)
	if !isArrayVal {
		return false
	}
	if av.valueType != av2.valueType {
		return false
	}
	if av.label != av2.label {
		return false
	}
	if len(av.items) != len(av2.items) {
		return false
	}
	for idx, item := range av.items {
		if item != av2.items[idx] {
			return false
		}
	}
	return true
}

// Label returns string representation of the array of string items.
func (av *ArrayValue) String() string {
	str := "["
	for idx, item := range av.items {
		str += item
		if idx < len(av.items)-1 {
			str += ","
		}
	}
	str += "]"
	return str
}

// Type returns value type as chosen in the NewArrayValue constructor.
func (av *ArrayValue) Type() ValueType {
	return av.valueType
}

// GetItems returns the array of items the value represents.
func (av *ArrayValue) GetItems() []string {
	return av.items
}

// ArrayValueDerBuilder can be used as DerValuesBuilder in MockDescriptorArgs
// to derive one StringValue for every item in the array.
func ArrayValueDerBuilder(key string, value Value) (derivedVals []KeyValuePair) {
	arrayVal, isArrayVal := value.(*ArrayValue)
	if isArrayVal {
		for _, item := range arrayVal.GetItems() {
			derivedVals = append(derivedVals, KeyValuePair{
				Key:   key + "/" + item,
				Value: NewStringValue(Object, item, item),
			})
		}
	}
	return derivedVals
}