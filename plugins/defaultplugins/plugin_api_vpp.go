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

//go:generate generic github.com/ligato/cn-infra/datasync/chngapi apitypes/iftypes Types->GetInterfaces Type->github.cisco.com/ligato/vpp-agent/defaultplugins/ifplugin/model/interfaces:interfaces.Interfaces_Interface

package defaultplugins

import (
	"github.com/ligato/vpp-agent/idxvpp"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/bdidx"
)

// API of VPP Plugin
type API interface {

	// DisableResync for one or more VPP plugins. Use in Init() phase.
	DisableResync(keyPrefix ...string)

	// GetSwIfIndexes gives access to mapping of logical names (used in ETCD configuration) to sw_if_index.
	// This mapping is helpful if other plugins need to configure VPP by the Binary API that uses sw_if_index input.
	//
	// Example of is_sw_index lookup by logical name of the port "vswitch_ingres" of the network interface:
	//
	//   func Init() error {
	//      swIfIndexes := defaultplugins.GetSwIfIndexes()
	//      swIfIndexes.LookupByName("vswitch_ingres")
	//
	GetSwIfIndexes() ifaceidx.SwIfIndex

	// GetBfdSessionIndexes gives access to mapping of logical names (used in ETCD configuration) to bfd_session_indexes.
	// The mapping consists of the interface (its name), generated index and the BFDSessionMeta with an authentication key
	// used for the particular session.
	GetBfdSessionIndexes() idxvpp.NameToIdx

	// GetBfdAuthKeyIndexes gives access to mapping of logical names (used in ETCD configuration) to bfd_auth_keys.
	// The authentication key has its own unique ID - the value is as a string stored in the mapping. Unique index is generated
	// uint32 number.
	GetBfdAuthKeyIndexes() idxvpp.NameToIdx

	// GetBfdEchoFunctionIndexes gives access to mapping of logical names (used in ETCD configuration) to bfd_echo_function
	// The echo function uses the interface name as an unique ID - this value is as a string stored in the mapping. The index
	// is generated uint32 number.
	GetBfdEchoFunctionIndexes() idxvpp.NameToIdx

	// GetBDIndexes gives access to mapping of logical names (used in ETCD configuration) as bd_indexes. The mapping consists
	// from the unique Bridge domain name and the bridge domain ID.
	GetBDIndexes() bdidx.BDIndex

	// GetFIBIndexes gives access to mapping of logical names (used in ETCD configuration) as fib_indexes. The FIB's physical
	// address is the name in the mapping. The key is generated. The FIB mapping also contains a metadata, FIBMeta with various
	// info about the Interface/Bridge domain where this fib belongs to:
	// - InterfaceName
	// - Bridge domain name
	// - BVI (bool flag for interface)
	// - Static config
	GetFIBIndexes() idxvpp.NameToIdx

	// GetXConnectIndexes gives access to mapping of logical names (used in ETCD configuration) as xc_indexes. The mapping
	// uses the name and the index of receive interface (the one all packets are received on). XConnectMeta is a container
	// for the transmit interface name.
	GetXConnectIndexes() idxvpp.NameToIdx
}
