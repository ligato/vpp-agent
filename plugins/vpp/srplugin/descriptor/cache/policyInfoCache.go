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

// Package cache contains global caches helping with implementation of descriptors
package cache

import srv6 "github.com/ligato/vpp-agent/api/models/vpp/srv6"

// PolicyInfoCache stores global state need to provide special handling in Policy/PolicySegmentList
// creation and removal related to the fact that VPP doesn't allow to have Policy without PolicySegmentLists.
// Such a limitation of VPP implicates additional implementation challenges in case of handling first/last PolicySegmentLists
// by creation/removal.
// PolicyInfoCache is mapping from Policy BSID (string form) to global info about given policy.
type PolicyInfoCache map[string]*PolicyInfo

// PolicyInfo stores global state for one Policy. It is part of PolicyInfoCache and is used for special Policy/PolicySegmentList handling
type PolicyInfo struct {
	// (used for creation) policy segment list choosen to be used as first policy segment list when policy is created
	PolicyCreationSL *srv6.PolicySegmentList

	// (used for creation) ignoring creation of policyCreationSL in creation method (Policy creates handles this) can be done only once per
	// existence of policy -> this boolean locks further creation ignoring (policyCreationSL alone is not enough because
	// we can remove it and add it again, equals would say it is the choosen one created by policy and it would not be created)
	PolicyCreationSLCreated bool

	// (used for delete) only the delete of last policy segment list can be ignored (it is handled by Policy delete that
	// is trigger in the same transaction due to dependencies)
	SLCounter int
}
