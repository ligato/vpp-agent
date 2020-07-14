//  Copyright (c) 2020 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package descriptor

import (
	"context"

	"github.com/go-errors/errors"
	"go.ligato.io/cn-infra/v2/logging"

	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/descriptor/adapter"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/ifaceidx"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/vppcalls"
	interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/interfaces"
)

const (
	// IP6ndDescriptorName is the name of the descriptor.
	IP6ndDescriptorName = "vpp-ip6nd"
)

// IP6ndDescriptor instructs KVScheduler how to configure VPP IP6ND entries.
type IP6ndDescriptor struct {
	log       logging.Logger
	handler   vppcalls.IP6ndVppAPI
	scheduler kvs.KVScheduler
	ifIndex   ifaceidx.IfaceMetadataIndex
}

// NewIP6ndDescriptor creates a new instance of the IP6ndDescriptor.
func NewIP6ndDescriptor(scheduler kvs.KVScheduler, handler vppcalls.IP6ndVppAPI,
	ifIndex ifaceidx.IfaceMetadataIndex, log logging.PluginLogger) *kvs.KVDescriptor {

	ctx := &IP6ndDescriptor{
		scheduler: scheduler,
		handler:   handler,
		ifIndex:   ifIndex,
		log:       log.NewLogger("ip6nd-descriptor"),
	}

	typedDescr := &adapter.IP6NDDescriptor{
		Name:        IP6ndDescriptorName,
		KeySelector: ctx.IsIP6NDRelatedKey,
		KeyLabel:    ctx.InterfaceNameFromKey,
		Create:      ctx.Create,
		Delete:      ctx.Delete,
		//Retrieve:             ctx.Retrieve,
		RetrieveDependencies: []string{InterfaceDescriptorName},
	}
	return adapter.NewIP6NDDescriptor(typedDescr)
}

// IsIP6NDRelatedKey returns true if the key is identifying IP6ND config (derived value)
func (d *IP6ndDescriptor) IsIP6NDRelatedKey(key string) bool {
	if _, isValid := interfaces.ParseNameFromIP6NDKey(key); isValid {
		return true
	}
	return false
}

// InterfaceNameFromKey returns interface name from IP6ND-related key.
func (d *IP6ndDescriptor) InterfaceNameFromKey(key string) string {
	if iface, isValid := interfaces.ParseNameFromIP6NDKey(key); isValid {
		return iface
	}
	return key
}

// Create adds a VPP IP6ND entry.
func (d *IP6ndDescriptor) Create(key string, entry *interfaces.Interface_IP6ND) (metadata interface{}, err error) {
	ifName, _ := interfaces.ParseNameFromIP6NDKey(key)
	ifMeta, found := d.ifIndex.LookupByName(ifName)
	if !found {
		err = errors.Errorf("failed to find IP6ND-enabled interface %s", ifName)
		d.log.Error(err)
		return nil, err
	}

	if err := d.handler.SetIP6ndAutoconfig(context.Background(), ifMeta.SwIfIndex, entry.AddressAutoconfig, entry.InstallDefaultRoutes); err != nil {
		err = errors.Errorf("failed to enable IP6ND for interface %s", ifName)
		d.log.Error(err)
		return nil, err
	}

	return nil, err
}

// Delete removes a VPP IP6ND entry.
func (d *IP6ndDescriptor) Delete(key string, entry *interfaces.Interface_IP6ND, metadata interface{}) (err error) {
	ifName, _ := interfaces.ParseNameFromIP6NDKey(key)
	ifMeta, found := d.ifIndex.LookupByName(ifName)
	if !found {
		err = errors.Errorf("failed to find IP6ND-enabled interface %s", ifName)
		d.log.Error(err)
		return err
	}

	if err := d.handler.SetIP6ndAutoconfig(context.Background(), ifMeta.SwIfIndex, false, false); err != nil {
		err = errors.Errorf("failed to disable IP6ND for interface %s", ifName)
		d.log.Error(err)
		return err
	}

	return nil
}

// Retrieve returns all IP6ND entries.
// TODO: implement retrieve
/*func (d *IP6ndDescriptor) Retrieve(correlate []adapter.IP6NDKVWithMetadata) (
	retrieved []adapter.IP6NDKVWithMetadata, err error,
) {
	entries, err := d.handler.DumpIP6ND()
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		retrieved = append(retrieved, adapter.IP6NDKVWithMetadata{
			Key:    models.Key(entry),
			Value:  entry,
			Origin: kvs.UnknownOrigin,
		})
	}
	return
}*/
