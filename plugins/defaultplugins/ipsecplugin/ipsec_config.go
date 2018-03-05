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

//go:generate protoc --proto_path=../common/model/ipsec --gogo_out=../common/model/ipsec ../common/model/ipsec/ipsec.proto

//go:generate binapi-generator --input-file=/usr/share/vpp/api/ipsec.api.json --output-dir=../common/bin_api

// Package ipsecplugin implements the IPSec plugin that handles management of IPSec for VPP.
package ipsecplugin

import (
	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/ipsec"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ipsecplugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/govppmux"
)

// IPSecConfigurator runs in the background in its own goroutine where it watches for any changes
// in the configuration of interfaces as modelled by the proto file "../model/ipsec/ipsec.proto"
// and stored in ETCD under the key "/vnf-agent/{vnf-agent}/vpp/config/v1/ipsec".
// Updates received from the northbound API are compared with the VPP run-time configuration and differences
// are applied through the VPP binary API.
type IPSecConfigurator struct {
	Log       logging.Logger
	Stopwatch *measure.Stopwatch // timer used to measure and store time

	GoVppmux govppmux.API
	vppCh    *govppapi.Channel

	SwIfIndexes ifaceidx.SwIfIndexRW

	SpdIndexSeq uint32
	SaIndexSeq  uint32
}

// Init members (channels...) and start go routines
func (plugin *IPSecConfigurator) Init() (err error) {
	plugin.Log.Debug("Initializing IPSec configurator")

	plugin.vppCh, err = plugin.GoVppmux.NewAPIChannel()
	if err != nil {
		return err
	}
	if err := vppcalls.CheckMsgCompatibilityForIPSec(plugin.vppCh); err != nil {
		return err
	}

	return nil
}

// Close GOVPP channel
func (plugin *IPSecConfigurator) Close() error {
	return safeclose.Close(plugin.vppCh)
}

// ConfigureSPD configures SPD
func (plugin *IPSecConfigurator) ConfigureSPD(spd *ipsec.SecurityPolicyDatabases_SPD) error {
	plugin.Log.Infof("Configuring SPD %v", spd.Name)

	return nil
}

// ModifySPD
func (plugin *IPSecConfigurator) ModifySPD(oldSpd *ipsec.SecurityPolicyDatabases_SPD, spd *ipsec.SecurityPolicyDatabases_SPD) error {
	plugin.Log.Infof("Modifying SPD %v", spd.Name)

	return nil
}

// DeleteSPD
func (plugin *IPSecConfigurator) DeleteSPD(oldSpd *ipsec.SecurityPolicyDatabases_SPD) error {
	plugin.Log.Infof("Deleting SPD %v", oldSpd.Name)

	return nil
}

// ConfigureSA
func (plugin *IPSecConfigurator) ConfigureSA(sa *ipsec.SecurityAssociations_SA) error {
	plugin.Log.Infof("Configuring SA %v", sa.Name)

	return nil
}

// ModifySA
func (plugin *IPSecConfigurator) ModifySA(oldSa *ipsec.SecurityAssociations_SA, sa *ipsec.SecurityAssociations_SA) error {
	plugin.Log.Infof("Modifying SA %v", sa.Name)

	return nil
}

// DeleteSA
func (plugin *IPSecConfigurator) DeleteSA(oldSa *ipsec.SecurityAssociations_SA) error {
	plugin.Log.Infof("Deleting SA %v", oldSa.Name)

	return nil
}

// Resync
func (plugin *IPSecConfigurator) Resync(spds []*ipsec.SecurityPolicyDatabases_SPD, sas []*ipsec.SecurityAssociations_SA) error {
	return nil
}
