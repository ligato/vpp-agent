//  Copyright (c) 2018 Cisco and/or its affiliates.
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

package acl

import "github.com/ligato/vpp-agent/plugins/vpp/model"

const (
	// DB key prefix
	aclPrefix = "vpp/config" + model.ProtoApiVersion + "acl/"
	// REST Acl IP prefix
	restAclIP = model.ProtoApiVersion + "acl/ip"
	// REST Acl IP example prefix
	restAclIPExample = model.ProtoApiVersion + "acl/ip/example"
	// REST Acl MACIP prefix
	restAclMACIP = model.ProtoApiVersion + "acl/macip"
	// REST Acl MACIP example prefix
	restAclMACIPExample = model.ProtoApiVersion + "acl/macip/example"
)

// KeyPrefix returns the prefix used in ETCD to store vpp ACLs config.
func KeyPrefix() string {
	return aclPrefix
}

// Key returns the prefix used in ETCD to store vpp ACL config
// of a particular ACL in selected vpp instance.
func Key(aclName string) string {
	return aclPrefix + aclName
}

// RestIPKey returns prefix used in REST to dump ACL IP config
func RestIPKey() string {
	return restAclIP
}

// RestIPExampleKey returns prefix used in REST to dump ACL IP example config
func RestIPExampleKey() string {
	return restAclIPExample
}

// RestMACIPKey returns prefix used in REST to dump ACL MACIP config
func RestMACIPKey() string {
	return restAclMACIP
}

// RestMACIPExampleKey returns prefix used in REST to dump ACL MACIP example config
func RestMACIPExampleKey() string {
	return restAclMACIPExample
}
