// Copyright (c) 2020 Pantheon.tech
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

package vpp_dns

import "go.ligato.io/vpp-agent/v3/pkg/models"

const (
	ModuleName = "vpp.dns"
)

var (
	// DNSCache is registered NB model of DNSCache
	ModelDNSCache = models.Register(&DNSCache{}, models.Spec{
		Module:  ModuleName,
		Type:    "dnscache",
		Version: "v1",
	})
)
