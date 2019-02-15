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
	"strings"

	srv6 "github.com/ligato/vpp-agent/api/models/vpp/srv6"
)

// PolicyIndexCache is storage for Policy indexes assigned inside VPP
type PolicyIndexCache struct {
	internal map[string]uint32
}

// NewPolicyIndexCache creates Policy index storage
func NewPolicyIndexCache() *PolicyIndexCache {
	return &PolicyIndexCache{
		internal: make(map[string]uint32),
	}
}

// Get retrieves VPP index for policy <policy>
func (c *PolicyIndexCache) Get(policy *srv6.Policy) (uint32, bool) {
	index, ok := c.internal[c.id(policy)]
	return index, ok
}

// Put stores VPP index for policy <policy>
func (c *PolicyIndexCache) Put(policy *srv6.Policy, index uint32) {
	c.internal[c.id(policy)] = index
}

// Remove removes VPP index for policy <policy> from storage
func (c *PolicyIndexCache) Remove(policy *srv6.Policy) {
	delete(c.internal, c.id(policy))
}

func (c *PolicyIndexCache) id(policy *srv6.Policy) string {
	return strings.TrimSpace(strings.ToLower(policy.Bsid))
}
