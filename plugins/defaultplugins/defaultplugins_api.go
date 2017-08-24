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

// GetSwIfIndexes gives access to mapping of logical names (used in ETCD configuration) to sw_if_index.
// This mapping is helpful if other plugins need to configure VPP by the Binary API that uses sw_if_index input.
//
// Example of is_sw_index lookup by logical name of the port "vswitch_ingres" of the network interface
//
//   func Init() error {
//      swIfIndexes := defaultplugins.GetSwIfIndexes()
//      swIfIndexes.LookupByName("vswitch_ingres")
//
func GetSwIfIndexes() ifaceidx.SwIfIndex {
	return plugin().swIfIndexes
}

// TODO dump interface

// GetBfdSessionIndexes gives access to mapping of logical names (used in ETCD configuration) to bfd_session_indexes.
// The mapping consists of the interface (its name), generated index and the BFDSessionMeta with an authentication key
// used for the particular session.
func GetBfdSessionIndexes() idxvpp.NameToIdx {
	return plugin().bfdSessionIndexes
}

// GetBfdAuthKeyIndexes gives access to mapping of logical names (used in ETCD configuration) to bfd_auth_keys.
// The authentication key has its own unique ID - the value is as a string stored in the mapping. Unique index is generated
// uint32 number
func GetBfdAuthKeyIndexes() idxvpp.NameToIdx {
	return plugin().bfdAuthKeysIndexes
}

// GetBfdEchoFunctionIndexes gives access to mapping of logical names (used in ETCD configuration) to bfd_echo_function
// The echo function uses the interface name as an unique ID - this value is as a string stored in the mapping. The index
// is generated uint32 number
func GetBfdEchoFunctionIndexes() idxvpp.NameToIdx {
	return plugin().bfdEchoFunctionIndex
}

// GetBDIndexes gives access to mapping of logical names (used in ETCD configuration) as bd_indexes. The mapping consists
// from the unique Bridge domain name and the bridge domain ID
func GetBDIndexes() bdidx.BDIndex {
	return plugin().bdIndexes
}

// GetFIBIndexes gives access to mapping of logical names (used in ETCD configuration) as fib_indexes. The FIB's physical
// address is the name in the mapping. The key is generated. The FIB mapping also contains a metadata, FIBMeta with various
// info about the Interface/Bridge domain where this fib belongs to:
// - InterfaceName
// - Bridge domain name
// - BVI (bool flag for interface)
// - Static config
func GetFIBIndexes() idxvpp.NameToIdx {
	return plugin().fibIndexes
}

// GetFIBDesIndexes gives access to mapping of logical names (used in ETCD configuration) as fib_des_indexes. The mapping
// reflects the desired state. FIBs that have been created, but cannot be configured because of missing interface or
// bridge domain are stored here. If both, interface and bridge domain are created later, stored FIBs will be configured
// as well. The mapping uses FIBMeta as metadata (the same as above)
func GetFIBDesIndexes() idxvpp.NameToIdx {
	return plugin().fibDesIndexes
}

// GetXConnectIndexes gives access to mapping of logical names (used in ETCD configuration) as xc_indexes. The mapping
// uses the name and the index of receive interface (the one all packets are received on). XConnectMeta is a container
// for the transmit interface name.
func GetXConnectIndexes() idxvpp.NameToIdx {
	return plugin().xcIndexes
}

// GetRegisteredInterfaceToBridgeDomainIndexes gives access to mapping of logical names (used in ETCD configuration).
// The mapping holds all interface to bridge domain pairs which should be configured on the VPP. The mapping uses the
// interface name, unique index and a metadata with information about whether the interface is BVI
// and which bridge domain it belongs to
func GetRegisteredInterfaceToBridgeDomainIndexes() idxvpp.NameToIdx {
	return plugin().ifToBdDesIndexes
}

// GetConfiguredInterfaceToBridgeDomainIndexes gives access to mapping of logical names (used in ETCD configuration)
// The mapping holds all interface to bridge domain pairs which are currently configured on the VPP. The mapping uses
// interface name, unique index and a metadata with information about whether the interface is BVI
// and which bridge domain it belongs to
func GetConfiguredInterfaceToBridgeDomainIndexes() idxvpp.NameToIdx {
	return plugin().ifToBdRealIndexes
}
