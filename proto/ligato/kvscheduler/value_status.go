//  Copyright (c) 2019 Cisco and/or its affiliates.
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
	"encoding/json"
)

// MarshalJSON ensures data is correctly marshaled
func (x ValueState) MarshalJSON() ([]byte, error) {
	return json.Marshal(x.String())
}

// UnmarshalJSON ensures that data is correctly unmarshaled
func (x *ValueState) UnmarshalJSON(b []byte) error {
	if b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		*x = ValueState(ValueState_value[s])
	} else {
		var n int
		if err := json.Unmarshal(b, &n); err != nil {
			return err
		}
		*x = ValueState(n)
	}
	return nil
}

// MarshalJSON ensures data is correctly marshaled
func (x TxnOperation) MarshalJSON() ([]byte, error) {
	return json.Marshal(x.String())
}

// UnmarshalJSON ensures that data is correctly unmarshaled
func (x *TxnOperation) UnmarshalJSON(b []byte) error {
	if b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		*x = TxnOperation(TxnOperation_value[s])
	} else {
		var n int
		if err := json.Unmarshal(b, &n); err != nil {
			return err
		}
		*x = TxnOperation(n)
	}
	return nil
}
