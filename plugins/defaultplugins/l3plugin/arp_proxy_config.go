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

package l3plugin

import (
	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/l3"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l3plugin/l3idx"
	"github.com/ligato/vpp-agent/plugins/govppmux"
)

// ProxyArpConfigurator runs in the background in its own goroutine where it watches for any changes
// in the configuration of L3 proxy arp entries as modelled by the proto file "../model/l3/l3.proto" and stored
// in ETCD under the key "/vnf-agent/{vnf-agent}/vpp/config/v1/proxyarp". Configuration uses separate keys
// for proxy arp range and interfaces. Updates received from the northbound API are compared with the VPP
// run-time configuration and differences are applied through the VPP binary API.
type ProxyArpConfigurator struct {
	Log logging.Logger

	GoVppmux govppmux.API

	// ProxyArpIndices is a list of proxy ARP interface entries which are successfully configured on the VPP
	ProxyArpIfIndices l3idx.ARPIndexRW
	// ProxyArpRngIndices is a list of proxy ARP range entries which are successfully configured on the VPP
	ProxyArpRngIndices l3idx.ARPIndexRW

	ARPIndexSeq uint32
	SwIfIndexes ifaceidx.SwIfIndex
	vppChan     *govppapi.Channel

	Stopwatch *measure.Stopwatch
}

func (plugin *ProxyArpConfigurator) Init() error {
	return nil
}

func (plugin *ProxyArpConfigurator) Close() error {
	return nil
}

func (plugin *ProxyArpConfigurator) AddInterface(pArpIf *l3.ProxyArpInterfaces_Interface) error {
	return nil
}

func (plugin *ProxyArpConfigurator) ModifyInterface(newPArpIf, oldPArpIf *l3.ProxyArpInterfaces_Interface) error {
	return nil
}

func (plugin *ProxyArpConfigurator) DeleteInterface(pArpIf *l3.ProxyArpInterfaces_Interface) error {
	return nil
}

func (plugin *ProxyArpConfigurator) AddRange(pArpRng *l3.ProxyArpRanges_Range) error {
	return nil
}

func (plugin *ProxyArpConfigurator) ModifyRange(newPArpRng, oldPArpRng *l3.ProxyArpRanges_Range) error {
	return nil
}

func (plugin *ProxyArpConfigurator) DeleteRange(pArpRng *l3.ProxyArpRanges_Range) error {
	return nil
}
