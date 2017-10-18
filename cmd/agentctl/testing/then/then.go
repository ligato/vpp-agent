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

// Package then contains methods for verification of agentctl unit test outcomes.
package then

import (
	"strings"

	"github.com/onsi/gomega"
)

// ContainsItems can be used to verify if the provided item(s) is present in the table. It could be an agent label, an interface or
// a header
func ContainsItems(data string, item ...string) {
	for _, header := range item {
		itemExists := strings.Contains(data, header)
		gomega.Expect(itemExists).To(gomega.BeTrue())
	}
}

// DoesNotContainItems can be used to verify if the provided item(s) is missing in the table. It could be an agent label, an interface or
// a header
func DoesNotContainItems(data string, item ...string) {
	for _, header := range item {
		itemExists := strings.Contains(data, header)
		gomega.Expect(itemExists).To(gomega.BeFalse())
	}
}
