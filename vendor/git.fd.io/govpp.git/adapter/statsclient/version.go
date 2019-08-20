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

package statsclient

import (
	"fmt"
)

const (
	MinVersion = 0
	MaxVersion = 1
)

func checkVersion(ver uint64) error {
	if ver < MinVersion {
		return fmt.Errorf("stat segment version is too old: %v (minimal version: %v)", ver, MinVersion)
	} else if ver > MaxVersion {
		return fmt.Errorf("stat segment version is not supported: %v (minimal version: %v)", ver, MaxVersion)
	}
	return nil
}
