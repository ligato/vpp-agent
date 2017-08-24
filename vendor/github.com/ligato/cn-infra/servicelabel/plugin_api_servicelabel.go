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

package servicelabel

// default service key prefix, can be changed in the build time using ldflgs, e.g.
// -ldflags '-X github.com/ligato/cn-infra/servicelabel.agentPrefix=/xyz/'
var agentPrefix = "/vnf-agent/"

// MicroserviceLabelEnvVar label this is inferred from the flag name
const MicroserviceLabelEnvVar = "MICROSERVICE_LABEL"

// ReaderAPI allows to read agent micorservice label with prefix.
// Reason for doing this is to have prefix for all keys of the agent.
type ReaderAPI interface {
	// GetAgentLabel returns string that is supposed to be used to distinguish
	// (ETCD) key prefixes for particular VNF (particular VPP Agent configuration)
	GetAgentLabel() string

	// GetAgentPrefix returns the string that is supposed to be used as the prefix for configuration of current
	// MicroserviceLabel "subtree" of the particular VPP Agent instance (e.g. in ETCD).
	GetAgentPrefix() string
	// GetDifferentAgentPrefix returns the string that is supposed to be used as the prefix for configuration
	// "subtree" of the particular VPP Agent instance (e.g. in ETCD).
	GetDifferentAgentPrefix(microserviceLabel string) string

	// GetAllAgentsPrefix returns the string that is supposed to be used as the prefix for configuration
	// subtree of the particular VPP Agent instance (e.g. in ETCD).
	GetAllAgentsPrefix() string
}

// GetAllAgentsPrefix returns the string that is supposed to be used as the prefix for configuration
// subtree of the particular VPP Agent instance (e.g. in ETCD).
func GetAllAgentsPrefix() string {
	return agentPrefix
}

// GetDifferentAgentPrefix returns the string that is supposed to be used as the prefix for configuration
// "subtree" of the particular VPP Agent instance (e.g. in ETCD).
func GetDifferentAgentPrefix(microserviceLabel string) string {
	return agentPrefix + microserviceLabel + "/"
}
