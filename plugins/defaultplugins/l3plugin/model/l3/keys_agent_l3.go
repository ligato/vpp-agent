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

package l3

// Prefixes
const (
	// RoutesPrefix is the relative key prefix for routes.
	RoutesPrefix = "vpp/config/v1/vrf/0/fib/" //TODO <VRF>
)

// RouteKey returns the key used in ETCD to store vpp route for vpp instance
func RouteKey(net string) string {
	return RoutesPrefix + net
}

// RouteKeyPrefix returns the prefix used in ETCD to store vpp routes for vpp instance
func RouteKeyPrefix() string {
	return RoutesPrefix
}
