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

//go:generate protoc --proto_path=model/l4 --gogo_out=model/l4 model/l4/l4.proto
//go:generate binapi-generator --input-file=/usr/share/vpp/api/session.api.json --output-dir=bin_api

package l4plugin

import (
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/servicelabel"
	"github.com/ligato/cn-infra/logging/measure"
	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l4plugin/model/l4"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l4plugin/vppcalls"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l4plugin/nsidx"
	"github.com/ligato/vpp-agent/plugins/govppmux"
	"fmt"
)

// L4Configurator runs in the background in its own goroutine where it watches for any changes
// in the configuration of interfaces as modelled by the proto file "../model/l4/l4.proto"
// and stored in ETCD under the keys "/vnf-agent/{vnf-agent}/vpp/config/v1/l4/features"
// and "/vnf-agent/{vnf-agent}/vpp/config/v1/l4/namespaces/{namespace_id}".
// Updates received from the northbound API are compared with the VPP run-time configuration and differences
// are applied through the VPP binary API.
type L4Configurator struct {
	Log logging.Logger

	ServiceLabel servicelabel.ReaderAPI
	GoVppmux      govppmux.API

	// Indexes
	SwIfIndexes ifaceidx.SwIfIndex
	AppNsIndexs nsidx.AppNsIndexRW

	// timer used to measure and store time
	Stopwatch *measure.Stopwatch

	// channel to communicate with the vpp
	vppCh *govppapi.Channel
}

// Init members (channels...) and start go routines
func (plugin *L4Configurator) Init() error {
	plugin.Log.Debugf("Init L4Configurator plugin")
	var err error

	// init vpp channel
	plugin.vppCh, err = plugin.GoVppmux.NewAPIChannel()
	if err != nil {
		return err
	}

	return nil
}

// Close members, channels
func (plugin *L4Configurator) Close() error {
	return nil
}

// ConfigureAppNamespace process the NB AppNamespace config and propagates it to bin api calls
func (plugin *L4Configurator) ConfigureAppNamespace(ns *l4.AppNamespaces_AppNamespace) error {
	plugin.Log.Infof("Configuring new AppNamespace with ID %v", ns.NamespaceId)

	// Find interface
	ifIdx, _, found := plugin.SwIfIndexes.LookupIdx(ns.Interface)
	if !found {
		return fmt.Errorf("cannot create AppNamespace with index %v, required interface %v not found in the mapping", ns.NamespaceId, ns.Interface)
	}

	// Namespace ID
	nsId := []byte(ns.NamespaceId)


	err := vppcalls.AddAppNamespace(ns.Secret, ifIdx, ns.Ipv4FibId, ns.Ipv6FibId, nsId, 0, plugin.Log, plugin.vppCh)
	if err != nil {
		return err
	}

	return nil
}

// ModifyAppNamespace process the NB AppNamespace config and propagates it to bin api calls
func (plugin *L4Configurator) ModifyAppNamespace(newNs *l4.AppNamespaces_AppNamespace, oldNs *l4.AppNamespaces_AppNamespace) error {
	return nil
}

// DeleteAppNamespace process the NB AppNamespace config and propagates it to bin api calls
func (plugin *L4Configurator) DeleteAppNamespace(ns *l4.AppNamespaces_AppNamespace) error {
	return nil
}