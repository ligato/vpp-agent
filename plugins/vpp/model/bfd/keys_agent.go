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

package bfd

import "github.com/ligato/vpp-agent/plugins/vpp/model"

const (
	// restBfdKey is a REST path of a bfd
	restBfdKey = model.ProtoApiVersion + "bfd"
	// bfdSessionPrefix bfd-session/
	bfdSessionPrefix = "vpp/config" + model.ProtoApiVersion + "bfd/session/"
	// restBfdSessionKey is a REST path of a bfd sessions
	restBfdSessionKey = model.ProtoApiVersion + "bfd/sessions"
	// bfdAuthKeysPrefix bfd-key/
	bfdAuthKeysPrefix = "vpp/config" + model.ProtoApiVersion + "bfd/auth-key/"
	// restBfdAuthKey is a REST path of a bfd authentication keys
	restBfdAuthKey = model.ProtoApiVersion + "bfd/authkeys"
	// BfdEchoFunctionPrefix bfd-echo-function/
	bfdEchoFunctionPrefix = "vpp/config" + model.ProtoApiVersion + "bfd/echo-function"
)

// RestBfdKey returns prefix used in REST to dump bfd config
func RestBfdKey() string {
	return restBfdKey
}

// SessionKeyPrefix returns the prefix used in ETCD to store vpp bfd config.
func SessionKeyPrefix() string {
	return bfdSessionPrefix
}

// SessionKey returns the prefix used in ETCD to store vpp bfd config
// of a particular bfd session in selected vpp instance.
func SessionKey(bfdSessionIfaceLabel string) string {
	return bfdSessionPrefix + bfdSessionIfaceLabel
}

// RestSessionKey returns prefix used in REST to dump bfd session config
func RestSessionKey() string {
	return restBfdSessionKey
}

// AuthKeysKeyPrefix returns the prefix used in ETCD to store vpp bfd config.
func AuthKeysKeyPrefix() string {
	return bfdAuthKeysPrefix
}

// AuthKeysKey returns the prefix used in ETCD to store vpp bfd config
// of a particular bfd key in selected vpp instance.
func AuthKeysKey(bfdKeyIDLabel string) string {
	return bfdAuthKeysPrefix + bfdKeyIDLabel
}

// RestAuthKeysKey returns prefix used in REST to dump bfd authentication config
func RestAuthKeysKey() string {
	return restBfdAuthKey
}

// EchoFunctionKeyPrefix returns the prefix used in ETCD to store vpp bfd config.
func EchoFunctionKeyPrefix() string {
	return bfdEchoFunctionPrefix
}

// EchoFunctionKey returns the prefix used in ETCD to store vpp bfd config
// of a particular bfd echo function in selected vpp instance.
func EchoFunctionKey(bfdEchoIfaceLabel string) string {
	return bfdEchoFunctionPrefix + bfdEchoIfaceLabel
}
