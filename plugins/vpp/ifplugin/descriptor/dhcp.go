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

package descriptor

import (
	"context"
	"strings"
	"sync"

	"github.com/golang/protobuf/proto"
	prototypes "github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"
	"go.ligato.io/cn-infra/v2/logging"

	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/vppcalls"
	"go.ligato.io/vpp-agent/v3/proto/ligato/netalloc"
	interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
)

const (
	// DHCPDescriptorName is the name of the descriptor configuring DHCP for VPP
	// interfaces.
	DHCPDescriptorName = "vpp-dhcp"
)

// DHCPDescriptor enables/disables DHCP for VPP interfaces and notifies about
// new DHCP leases.
type DHCPDescriptor struct {
	// provided by the plugin
	log         logging.Logger
	ifHandler   vppcalls.InterfaceVppAPI
	kvscheduler kvs.KVScheduler
	ifIndex     ifaceidx.IfaceMetadataIndex

	// DHCP notification watching
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewDHCPDescriptor creates a new instance of DHCPDescriptor.
func NewDHCPDescriptor(kvscheduler kvs.KVScheduler, ifHandler vppcalls.InterfaceVppAPI,
	ifIndex ifaceidx.IfaceMetadataIndex, log logging.PluginLogger,
) (*kvs.KVDescriptor, *DHCPDescriptor) {
	ctx := &DHCPDescriptor{
		kvscheduler: kvscheduler,
		ifHandler:   ifHandler,
		ifIndex:     ifIndex,
		log:         log.NewLogger("dhcp-descriptor"),
	}
	descr := &kvs.KVDescriptor{
		Name:                 DHCPDescriptorName,
		KeySelector:          ctx.IsDHCPRelatedKey,
		KeyLabel:             ctx.InterfaceNameFromKey,
		WithMetadata:         true,              // DHCP leases
		Create:               ctx.Create,        // DHCP client
		Delete:               ctx.Delete,        // DHCP client
		Retrieve:             ctx.Retrieve,      // DHCP leases
		DerivedValues:        ctx.DerivedValues, // IP address from DHCP lease
		RetrieveDependencies: []string{InterfaceDescriptorName},
	}
	return descr, ctx
}

// WatchDHCPNotifications starts watching for DHCP notifications.
func (d *DHCPDescriptor) WatchDHCPNotifications(ctx context.Context) {
	// Create child context
	var childCtx context.Context
	childCtx, d.cancel = context.WithCancel(ctx)

	d.wg.Add(1)
	go d.watchDHCPNotifications(childCtx)
}

// Close stops watching of DHCP notifications.
func (d *DHCPDescriptor) Close() error {
	d.cancel()
	d.wg.Wait()
	return nil
}

// IsDHCPRelatedKey returns true if the key is identifying DHCP client (derived value)
// or DHCP lease (notification).
func (d *DHCPDescriptor) IsDHCPRelatedKey(key string) bool {
	if _, isValid := interfaces.ParseNameFromDHCPClientKey(key); isValid {
		return true
	}
	if _, isValid := interfaces.ParseNameFromDHCPLeaseKey(key); isValid {
		return true
	}
	return false
}

// InterfaceNameFromKey returns interface name from DHCP-related key.
func (d *DHCPDescriptor) InterfaceNameFromKey(key string) string {
	if iface, isValid := interfaces.ParseNameFromDHCPClientKey(key); isValid {
		return iface
	}
	if iface, isValid := interfaces.ParseNameFromDHCPLeaseKey(key); isValid {
		return iface
	}
	return key
}

// Create enables DHCP client.
func (d *DHCPDescriptor) Create(key string, emptyVal proto.Message) (metadata kvs.Metadata, err error) {
	ifName, _ := interfaces.ParseNameFromDHCPClientKey(key)
	ifMeta, found := d.ifIndex.LookupByName(ifName)
	if !found {
		err = errors.Errorf("failed to find DHCP-enabled interface %s", ifName)
		d.log.Error(err)
		return nil, err
	}

	if err := d.ifHandler.SetInterfaceAsDHCPClient(ifMeta.SwIfIndex, ifName); err != nil {
		err = errors.Errorf("failed to enable DHCP client for interface %s", ifName)
		d.log.Error(err)
		return nil, err
	}

	return nil, err
}

// Delete disables DHCP client.
func (d *DHCPDescriptor) Delete(key string, emptyVal proto.Message, metadata kvs.Metadata) error {
	ifName, _ := interfaces.ParseNameFromDHCPClientKey(key)
	ifMeta, found := d.ifIndex.LookupByName(ifName)
	if !found {
		err := errors.Errorf("failed to find DHCP-enabled interface %s", ifName)
		d.log.Error(err)
		return err
	}

	if err := d.ifHandler.UnsetInterfaceAsDHCPClient(ifMeta.SwIfIndex, ifName); err != nil {
		err = errors.Errorf("failed to disable DHCP client for interface %s", ifName)
		d.log.Error(err)
		return err
	}

	// notify about the unconfigured client by removing the lease notification
	return d.kvscheduler.PushSBNotification(kvs.KVWithMetadata{
		Key:      interfaces.DHCPLeaseKey(ifName),
		Value:    nil,
		Metadata: nil,
	})
}

// Retrieve returns all existing DHCP leases.
func (d *DHCPDescriptor) Retrieve(correlate []kvs.KVWithMetadata) (
	leases []kvs.KVWithMetadata, err error,
) {
	// Retrieve VPP configuration.
	dhcpDump, err := d.ifHandler.DumpDhcpClients()
	if err != nil {
		d.log.Error(err)
		return leases, err
	}

	for ifIdx, dhcpData := range dhcpDump {
		ifName, _, found := d.ifIndex.LookupBySwIfIndex(ifIdx)
		if !found {
			d.log.Warnf("failed to find interface sw_if_index=%d with DHCP lease", ifIdx)
			return leases, err
		}
		// Store lease under both value (for visibility & to derive interface IP address)
		// and metadata (for watching).
		lease := &interfaces.DHCPLease{
			InterfaceName:   ifName,
			HostName:        dhcpData.Lease.Hostname,
			HostPhysAddress: dhcpData.Lease.HostMac,
			IsIpv6:          dhcpData.Lease.IsIPv6,
			HostIpAddress:   dhcpData.Lease.HostAddress,
			RouterIpAddress: dhcpData.Lease.RouterAddress,
		}
		leases = append(leases, kvs.KVWithMetadata{
			Key:      interfaces.DHCPLeaseKey(ifName),
			Value:    lease,
			Metadata: lease,
			Origin:   kvs.FromSB,
		})
	}

	return leases, nil
}

// DerivedValues derives empty value for leased IP address.
func (d *DHCPDescriptor) DerivedValues(key string, dhcpData proto.Message) (derValues []kvs.KeyValuePair) {
	if strings.HasPrefix(key, interfaces.DHCPLeaseKeyPrefix) {
		dhcpLease, ok := dhcpData.(*interfaces.DHCPLease)
		if ok && dhcpLease.HostIpAddress != "" {
			return []kvs.KeyValuePair{
				{
					Key: interfaces.InterfaceAddressKey(dhcpLease.InterfaceName, dhcpLease.HostIpAddress,
						netalloc.IPAddressSource_FROM_DHCP),
					Value: &prototypes.Empty{},
				},
			}
		}
	}
	return derValues
}

// watchDHCPNotifications watches and processes DHCP notifications.
func (d *DHCPDescriptor) watchDHCPNotifications(ctx context.Context) {
	defer d.wg.Done()
	d.log.Debug("Started watcher on DHCP notifications")

	dhcpChan := make(chan *vppcalls.Lease)
	if err := d.ifHandler.WatchDHCPLeases(dhcpChan); err != nil {
		d.log.Errorf("watching dhcp leases failed: %v", err)
		return
	}

	for {
		select {
		case lease := <-dhcpChan:
			// interface logical name
			ifName, _, found := d.ifIndex.LookupBySwIfIndex(lease.SwIfIndex)
			if !found {
				d.log.Warnf("Interface sw_if_index=%d with DHCP lease was not found in the mapping", lease.SwIfIndex)
				continue
			}

			d.log.Debugf("DHCP assigned %v to interface %q (router address %v)", lease.HostAddress, ifName, lease.RouterAddress)

			// notify about the new lease
			dhcpLease := &interfaces.DHCPLease{
				InterfaceName:   ifName,
				HostName:        lease.Hostname,
				HostPhysAddress: lease.HostMac,
				HostIpAddress:   lease.HostAddress,
				RouterIpAddress: lease.RouterAddress,
			}
			if err := d.kvscheduler.PushSBNotification(kvs.KVWithMetadata{
				Key:      interfaces.DHCPLeaseKey(ifName),
				Value:    dhcpLease,
				Metadata: dhcpLease,
			}); err != nil {
				d.log.Error(err)
			}
		case <-ctx.Done():
			return
		}
	}
}
