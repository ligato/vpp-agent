// Copyright (c) 2019 Bell Canada, Pantheon Technologies and/or its affiliates.
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

package cache

import (
	"bytes"
	"fmt"
	"strings"

	srv6 "github.com/ligato/vpp-agent/api/models/vpp/srv6"
)

// PolicySegmentListIndexCache is storage for PolicySegmentList indexes assigned inside VPP
type PolicySegmentListIndexCache struct {
	internal map[string]uint32
}

// NewPolicySegmentListIndexCache creates PolicySegmentList index storage
func NewPolicySegmentListIndexCache() *PolicySegmentListIndexCache {
	return &PolicySegmentListIndexCache{
		internal: make(map[string]uint32),
	}
}

// Get retrieves VPP index for PolicySegmentList <sl>
func (c *PolicySegmentListIndexCache) Get(sl *srv6.PolicySegmentList) (uint32, bool) {
	index, ok := c.internal[c.id(sl)]
	return index, ok
}

// Put stores VPP index for PolicySegmentList <sl>
func (c *PolicySegmentListIndexCache) Put(sl *srv6.PolicySegmentList, index uint32) {
	c.internal[c.id(sl)] = index
}

// Remove removes VPP index for PolicySegmentList <sl> from storage
func (c *PolicySegmentListIndexCache) Remove(sl *srv6.PolicySegmentList) {
	delete(c.internal, c.id(sl))
}

func (c *PolicySegmentListIndexCache) id(sl *srv6.PolicySegmentList) string {
	var segmentStrBuf bytes.Buffer
	for _, segment := range sl.Segments {
		segmentStrBuf.WriteString(strings.TrimSpace(strings.ToLower(segment)))
		segmentStrBuf.WriteString(",")
	}
	return strings.TrimSpace(strings.ToLower(sl.GetPolicyBsid())) + "#" + fmt.Sprint(sl.GetWeight()) + "#" + segmentStrBuf.String()
}
