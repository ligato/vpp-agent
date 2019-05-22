// Copyright (c) 2019 PANTHEON.tech
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

package descriptor

import (
	"net"

	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/utils/addrs"
	l3 "github.com/ligato/vpp-agent/api/models/vpp/l3"
	kvs "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/vpp/l3plugin/descriptor/adapter"
	"github.com/ligato/vpp-agent/plugins/vpp/l3plugin/vppcalls"
	"github.com/pkg/errors"
)

const (
	// ProxyArpInterfaceDescriptorName is the name of the descriptor.
	DHCPProxyDescriptorName = "dhcp-proxy"

	//dependecy labels
	vrfTableDependency = "vrf-table-exists"
)

// DHCPProxyDescriptor teaches KVScheduler how to configure VPP DHCP proxy.
type DHCPProxyDescriptor struct {
	log              logging.Logger
	dhcpProxyHandler vppcalls.DHCPProxyAPI
	scheduler        kvs.KVScheduler
}

// NewDHCPProxyDescriptor creates a new instance of the DHCPProxyDescriptor.
func NewDHCPProxyDescriptor(scheduler kvs.KVScheduler,
	dhcpProxyHandler vppcalls.DHCPProxyAPI, log logging.PluginLogger) *kvs.KVDescriptor {

	ctx := &DHCPProxyDescriptor{
		scheduler:        scheduler,
		dhcpProxyHandler: dhcpProxyHandler,
		log:              log.NewLogger("dhcp-proxy-descriptor"),
	}

	typedDescr := &adapter.DHCPProxyDescriptor{
		Name:          DHCPProxyDescriptorName,
		KeySelector:   l3.ModelDHCPProxy.IsKeyValid,
		KeyLabel:      l3.ModelDHCPProxy.StripKeyPrefix,
		NBKeyPrefix:   l3.ModelDHCPProxy.KeyPrefix(),
		ValueTypeName: l3.ModelDHCPProxy.ProtoName(),
		Create:        ctx.Create,
		Delete:        ctx.Delete,
		Retrieve:      ctx.Retrieve,
		Dependencies:  ctx.Dependencies,
		Validate:      ctx.Validate,
	}
	return adapter.NewDHCPProxyDescriptor(typedDescr)
}

func (d *DHCPProxyDescriptor) Validate(key string, value *l3.DHCPProxy) error {

	ipAddr := net.ParseIP(value.SourceIpAddress)
	if ipAddr == nil {
		return errors.Errorf("invalid source IP address: %q", value.SourceIpAddress)
	}

	for _, server := range value.Servers {
		serverIPAddr := net.ParseIP(server.IpAddress)
		if serverIPAddr == nil {
			return errors.Errorf("invalid server IP address: %q", server.IpAddress)
		}

		if ipAddr.To4() != nil && serverIPAddr.To4() == nil || serverIPAddr.To4() != nil && ipAddr.To4() == nil {
			return errors.Errorf("Server address %s must use same IP protocol as source address %s",
				serverIPAddr, ipAddr)
		}
	}
	return nil
}

// Dependencies lists dependencies for a VPP DHCP proxy.
func (d *DHCPProxyDescriptor) Dependencies(key string, value *l3.DHCPProxy) (deps []kvs.Dependency) {
	// non-zero VRFs
	var protocol l3.VrfTable_Protocol
	_, isIPv6, _ := addrs.ParseIPWithPrefix(value.SourceIpAddress)
	if isIPv6 {
		protocol = l3.VrfTable_IPV6
	}

	if value.RxVrfId != 0 {
		deps = append(deps, kvs.Dependency{
			Label: vrfTableDependency,
			Key:   l3.VrfTableKey(value.RxVrfId, protocol),
		})
	}

	for _, server := range value.Servers {
		if server.VrfId != 0 {
			deps = append(deps, kvs.Dependency{
				Label: vrfTableDependency,
				Key:   l3.VrfTableKey(server.VrfId, protocol),
			})
		}
	}

	return deps
}

// Create enables VPP DHCP proxy.
func (d *DHCPProxyDescriptor) Create(key string, value *l3.DHCPProxy) (metadata interface{}, err error) {
	if err := d.dhcpProxyHandler.CreateDHCPProxy(value); err != nil {
		return nil, errors.Errorf("failed to create DHCP proxy %v", err)
	}
	return nil, nil
}

// Delete disables VPP DHCP proxy.
func (d *DHCPProxyDescriptor) Delete(key string, value *l3.DHCPProxy, metadata interface{}) error {
	if err := d.dhcpProxyHandler.DeleteDHCPProxy(value); err != nil {
		return errors.Errorf("failed to delete DHCP proxy %v", err)
	}
	return nil
}

// Retrieve returns current VPP DHCP proxy configuration.
func (d *DHCPProxyDescriptor) Retrieve(correlate []adapter.DHCPProxyKVWithMetadata) (retrieved []adapter.DHCPProxyKVWithMetadata, err error) {
	// Retrieve VPP configuration
	dhcpProxyDetails, err := d.dhcpProxyHandler.DumpDHCPProxy()

	if err != nil {
		return nil, err
	}

	if dhcpProxyDetails == nil {
		return nil, nil
	}

	for _, detail := range dhcpProxyDetails {
		retrieved = append(retrieved, adapter.DHCPProxyKVWithMetadata{
			Key:    l3.DHCPProxyKey(detail.DHCPProxy.SourceIpAddress),
			Value:  detail.DHCPProxy,
			Origin: kvs.FromNB,
		})
	}

	return retrieved, nil
}
