// Copyright (c) 2018 Cisco and/or its affiliates.
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

//go:generate protoc --proto_path=../common/model/nat --gogo_out=../common/model/nat ../common/model/nat/nat.proto

//go:generate binapi-generator --input-file=/usr/share/vpp/api/nat.api.json --output-dir=../common/bin_api

package ifplugin

import (
	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/ligato/vpp-agent/idxvpp"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/nat"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/govppmux"
)

// NatConfigurator runs in the background in its own goroutine where it watches for any changes
// in the configuration of NAT address pools and static entries with or without a load ballance,
// as modelled by the proto file "../common/model/nat/nat.proto"
// and stored in ETCD under the keys:
// - "/vnf-agent/{agent-label}/vpp/config/v1/nat/{vrf}/addrpool/" for NAT address pool
// - "/vnf-agent/{agent-label}/vpp/config/v1/nat/{vrf}/static/" for NAT static mapping
// - "/vnf-agent/{agent-label}/vpp/config/v1/nat/{vrf}/staticlb/" for NAT static mapping with
//   load ballancer
// Updates received from the northbound API are compared with the VPP run-time configuration and differences
// are applied through the VPP binary API.
type NatConfigurator struct {
	Log logging.Logger

	GoVppmux    govppmux.API
	SwIfIndexes ifaceidx.SwIfIndex
	NatIndices  idxvpp.NameToIdxRW
	NatIndexSeq uint32
	vppChan     *govppapi.Channel

	Stopwatch *measure.Stopwatch
}

// Init NAT configurator
func (plugin *NatConfigurator) Init() (err error) {
	plugin.Log.Debug("Initializing STN configurator")

	// Init VPP API channel
	if plugin.vppChan, err = plugin.GoVppmux.NewAPIChannel(); err != nil {
		return
	}

	// Check VPP message compatibility
	if err = plugin.checkMsgCompatibility(); err != nil {
		return
	}

	return
}

// Close used resources
func (plugin *NatConfigurator) Close() error {
	return safeclose.Close(plugin.vppChan)
}

// SetNatGlobalConfig configures common setup for all NAT use cases
func (plugin *NatConfigurator) SetNatGlobalConfig(config *nat.NatGlobal) error {

	return nil
}

// ModifyNatGlobalConfig modifies common setup for all NAT use cases
func (plugin *NatConfigurator) ModifyNatGlobalConfig(oldConfig, newConfig *nat.NatGlobal) error {
	return nil
}

// DeleteNatGlobalConfig removes common setup for all NAT use cases
func (plugin *NatConfigurator) DeleteNatGlobalConfig(config *nat.NatGlobal) error {
	return nil
}

// ConfigureSNat configures new SNAT setup
func (plugin *NatConfigurator) ConfigureSNat(sNat *nat.Nat44SNat) error {
	return nil
}

// ModifySNat modifies existing SNAT setup
func (plugin *NatConfigurator) ModifySNat(oldSNat, newSNat *nat.Nat44SNat) error {
	return nil
}

// DeleteSNat removes existing SNAT setup
func (plugin *NatConfigurator) DeleteSNat(sNat *nat.Nat44SNat) error {
	return nil
}

// ConfigureDNat configures new DNAT setup
func (plugin *NatConfigurator) ConfigureDNat(sNat *nat.Nat44DNat) error {
	return nil
}

// ModifyDNat modifies existing DNAT setup
func (plugin *NatConfigurator) ModifyDNat(oldSNat, newSNat *nat.Nat44DNat) error {
	return nil
}

// DeleteDNat removes existing DNAT setup
func (plugin *NatConfigurator) DeleteDNat(sNat *nat.Nat44DNat) error {
	return nil
}

// checkMsgCompatibility verifies compatibility of used binary API calls
func (plugin *NatConfigurator) checkMsgCompatibility() error {
	msgs := []govppapi.Message{}
	err := plugin.vppChan.CheckMessageCompatibility(msgs...)
	if err != nil {
		plugin.Log.Error(err)
	}
	return err
}
