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
	"net"
	"strings"

	"github.com/pkg/errors"
	"go.ligato.io/cn-infra/v2/logging"

	"go.ligato.io/vpp-agent/v3/pkg/models"
	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	ifdescriptor "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/descriptor"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin/descriptor/adapter"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin/vppcalls"
	l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"
)

const (
	// ProxyArpDescriptorName is the name of the descriptor.
	ProxyArpDescriptorName = "vpp-proxy-arp"

	// Dependency labels:
	vrfTableProxyARPDep = "vrf-table-exists"
)

// Validation errors:
var (
	// ErrMissingIP returned when one of IP fields in ProxyARP range is not set.
	ErrMissingIP = errors.New("missing IP address")
	// ErrIPWithMask returned when one of IP fields in ProxyARP range is set with a subnet mask.
	ErrIPWithMask = errors.New("only one IP must be defined (e.g. \"192.0.2.1\"), not a subnet")
	// ErrInvalidIP returned when one of IP fields in ProxyARP range can not be parsed.
	ErrInvalidIP = errors.New("invalid IP address")
	// ErrIPv6NotSupported returned when one of IP fields in ProxyARP range is defined as IPv6.
	ErrIPv6NotSupported = errors.New("IP address must be IPv4, not IPv6")
)

// ProxyArpDescriptor teaches KVScheduler how to configure VPP proxy ARPs.
type ProxyArpDescriptor struct {
	log             logging.Logger
	proxyArpHandler vppcalls.ProxyArpVppAPI
	scheduler       kvs.KVScheduler
}

// NewProxyArpDescriptor creates a new instance of the ProxyArpDescriptor.
func NewProxyArpDescriptor(scheduler kvs.KVScheduler,
	proxyArpHandler vppcalls.ProxyArpVppAPI, log logging.PluginLogger) *kvs.KVDescriptor {

	ctx := &ProxyArpDescriptor{
		scheduler:       scheduler,
		proxyArpHandler: proxyArpHandler,
		log:             log.NewLogger("proxy-arp-descriptor"),
	}

	typedDescr := &adapter.ProxyARPDescriptor{
		Name:                 ProxyArpDescriptorName,
		NBKeyPrefix:          l3.ModelProxyARP.KeyPrefix(),
		ValueTypeName:        l3.ModelProxyARP.ProtoName(),
		KeySelector:          l3.ModelProxyARP.IsKeyValid,
		ValueComparator:      ctx.EquivalentProxyArps,
		Validate:             ctx.Validate,
		Create:               ctx.Create,
		Update:               ctx.Update,
		Delete:               ctx.Delete,
		Retrieve:             ctx.Retrieve,
		Dependencies:         ctx.Dependencies,
		DerivedValues:        ctx.DerivedValues,
		RetrieveDependencies: []string{ifdescriptor.InterfaceDescriptorName},
	}
	return adapter.NewProxyARPDescriptor(typedDescr)
}

// Validate validates ProxyARP setup.
func (d *ProxyArpDescriptor) Validate(key string, proxyArp *l3.ProxyARP) error {
	for _, r := range proxyArp.Ranges {
		if r.FirstIpAddr == "" {
			return kvs.NewInvalidValueError(ErrMissingIP, "ranges.first_ip_addr")
		}
		if r.LastIpAddr == "" {
			return kvs.NewInvalidValueError(ErrMissingIP, "ranges.last_ip_addr")
		}

		if strings.Contains(r.FirstIpAddr, "/") {
			return kvs.NewInvalidValueError(ErrIPWithMask, "ranges.first_ip_addr")
		}
		if strings.Contains(r.LastIpAddr, "/") {
			return kvs.NewInvalidValueError(ErrIPWithMask, "ranges.last_ip_addr")
		}

		firstIP := net.ParseIP(r.FirstIpAddr)
		if firstIP == nil {
			return kvs.NewInvalidValueError(ErrInvalidIP, "ranges.first_ip_addr")
		}
		lastIP := net.ParseIP(r.LastIpAddr)
		if lastIP == nil {
			return kvs.NewInvalidValueError(ErrInvalidIP, "ranges.last_ip_addr")
		}

		if firstIP.To4() == nil {
			return kvs.NewInvalidValueError(ErrIPv6NotSupported, "ranges.first_ip_addr")
		}
		if lastIP.To4() == nil {
			return kvs.NewInvalidValueError(ErrIPv6NotSupported, "ranges.last_ip_addr")
		}
	}
	return nil
}

// Dependencies lists dependencies for a VPP Proxy ARP.
func (d *ProxyArpDescriptor) Dependencies(key string, proxyArp *l3.ProxyARP) []kvs.Dependency {
	var dependencies []kvs.Dependency

	for _, r := range proxyArp.Ranges {
		if r.VrfId == 0 {
			continue
		}
		dependencies = append(dependencies, kvs.Dependency{
			Label: vrfTableProxyARPDep,
			Key:   l3.VrfTableKey(r.VrfId, l3.VrfTable_IPV4),
		})
	}

	return dependencies
}

// DerivedValues derives l3.ProxyARP_Interface for every interface.
func (d *ProxyArpDescriptor) DerivedValues(key string, proxyArp *l3.ProxyARP) (derValues []kvs.KeyValuePair) {
	for _, iface := range proxyArp.Interfaces {
		derValues = append(derValues, kvs.KeyValuePair{
			Key:   l3.ProxyARPInterfaceKey(iface.Name),
			Value: iface,
		})
	}
	return derValues
}

// EquivalentProxyArps compares VPP Proxy ARPs.
func (d *ProxyArpDescriptor) EquivalentProxyArps(key string, oldValue, newValue *l3.ProxyARP) bool {
	if len(newValue.Ranges) != len(oldValue.Ranges) {
		return false
	}
	toAdd, toDelete := calculateRngDiff(newValue.Ranges, oldValue.Ranges)
	return len(toAdd) == 0 && len(toDelete) == 0
}

// Create adds VPP Proxy ARP.
func (d *ProxyArpDescriptor) Create(key string, value *l3.ProxyARP) (metadata interface{}, err error) {
	for _, r := range value.Ranges {
		firstIP := net.ParseIP(r.FirstIpAddr).To4()
		lastIP := net.ParseIP(r.LastIpAddr).To4()

		if err := d.proxyArpHandler.AddProxyArpRange(firstIP, lastIP, r.VrfId); err != nil {
			return nil, errors.Errorf("failed to add proxy ARP address range %s - %s (VRF: %d): %v", firstIP, lastIP, r.VrfId, err)
		}
	}
	return nil, nil
}

// Update modifies VPP Proxy ARP.
func (d *ProxyArpDescriptor) Update(key string, oldValue, newValue *l3.ProxyARP, oldMetadata interface{}) (newMetadata interface{}, err error) {
	toAdd, toDelete := calculateRngDiff(newValue.Ranges, oldValue.Ranges)

	for _, r := range toDelete {
		firstIP := net.ParseIP(r.FirstIpAddr).To4()
		lastIP := net.ParseIP(r.LastIpAddr).To4()

		if err := d.proxyArpHandler.DeleteProxyArpRange(firstIP, lastIP, r.VrfId); err != nil {
			return nil, errors.Errorf("failed to delete proxy ARP address range %s - %s (VRF: %d): %v", firstIP, lastIP, r.VrfId, err)
		}
	}

	for _, r := range toAdd {
		firstIP := net.ParseIP(r.FirstIpAddr).To4()
		lastIP := net.ParseIP(r.LastIpAddr).To4()

		if err := d.proxyArpHandler.AddProxyArpRange(firstIP, lastIP, r.VrfId); err != nil {
			return nil, errors.Errorf("failed to add proxy ARP address range %s - %s (VRF: %d): %v", firstIP, lastIP, r.VrfId, err)
		}
	}

	return nil, nil
}

// Delete deletes VPP Proxy ARP.
func (d *ProxyArpDescriptor) Delete(key string, value *l3.ProxyARP, metadata interface{}) error {
	for _, r := range value.Ranges {
		firstIP := net.ParseIP(r.FirstIpAddr).To4()
		lastIP := net.ParseIP(r.LastIpAddr).To4()

		if err := d.proxyArpHandler.DeleteProxyArpRange(firstIP, lastIP, r.VrfId); err != nil {
			return errors.Errorf("failed to delete proxy ARP address range %s - %s (VRF: %d): %v", firstIP, lastIP, r.VrfId, err)
		}
	}
	return nil
}

// Retrieve returns VPP Proxy ARP configuration.
func (d *ProxyArpDescriptor) Retrieve(correlate []adapter.ProxyARPKVWithMetadata) (
	retrieved []adapter.ProxyARPKVWithMetadata, err error) {

	// Retrieve VPP configuration
	rangesDetails, err := d.proxyArpHandler.DumpProxyArpRanges()
	if err != nil {
		return nil, err
	}
	ifacesDetails, err := d.proxyArpHandler.DumpProxyArpInterfaces()
	if err != nil {
		return nil, err
	}

	proxyArp := &l3.ProxyARP{}
	for _, rangeDetail := range rangesDetails {
		proxyArp.Ranges = append(proxyArp.Ranges, rangeDetail.Range)
	}
	for _, ifaceDetail := range ifacesDetails {
		proxyArp.Interfaces = append(proxyArp.Interfaces, ifaceDetail.Interface)
	}

	retrieved = append(retrieved, adapter.ProxyARPKVWithMetadata{
		Key:    models.Key(proxyArp),
		Value:  proxyArp,
		Origin: kvs.UnknownOrigin,
	})

	return retrieved, nil
}

// calculateRngDiff calculates difference between old and new ranges.
func calculateRngDiff(newRngs, oldRngs []*l3.ProxyARP_Range) (toAdd, toDelete []*l3.ProxyARP_Range) {
	// Find missing ranges.
	for _, newRng := range newRngs {
		var found bool
		for _, oldRng := range oldRngs {
			if newRng.VrfId == oldRng.VrfId &&
				newRng.FirstIpAddr == oldRng.FirstIpAddr &&
				newRng.LastIpAddr == oldRng.LastIpAddr {
				found = true
				break
			}
		}
		if !found {
			toAdd = append(toAdd, newRng)
		}
	}
	// Find obsolete ranges.
	for _, oldRng := range oldRngs {
		var found bool
		for _, newRng := range newRngs {
			if oldRng.VrfId == newRng.VrfId &&
				oldRng.FirstIpAddr == newRng.FirstIpAddr &&
				oldRng.LastIpAddr == newRng.LastIpAddr {
				found = true
				break
			}
		}
		if !found {
			toDelete = append(toDelete, oldRng)
		}
	}
	return
}
