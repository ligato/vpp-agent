package descriptor

import (
	"github.com/ligato/cn-infra/logging"
	l3 "github.com/ligato/vpp-agent/api/models/vpp/l3"
	kvs "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/vpp/l3plugin/descriptor/adapter"
	"github.com/ligato/vpp-agent/plugins/vpp/l3plugin/vppcalls"
	"github.com/pkg/errors"
)

const (
	// ProxyArpInterfaceDescriptorName is the name of the descriptor.
	DHCPProxyDescriptorName = "dhcp-proxy"
)

// DHCPProxyDescriptor teaches KVScheduler how to configure VPP DHCP proxy.
type DHCPProxyDescriptor struct {
	log             logging.Logger
	dhcpProxyHandler vppcalls.DHCPProxyAPI
	scheduler       kvs.KVScheduler
}

// NewDHCPProxyDescriptor creates a new instance of the DHCPProxyDescriptor.
func NewDHCPProxyDescriptor(scheduler kvs.KVScheduler,
	dhcpProxyHandler vppcalls.DHCPProxyAPI, log logging.PluginLogger) *kvs.KVDescriptor {

	ctx := &DHCPProxyDescriptor{
		scheduler:       scheduler,
		dhcpProxyHandler: dhcpProxyHandler,
		log:             log.NewLogger("dhcp-proxy-descriptor"),
	}

	typedDescr := &adapter.DHCPProxyDescriptor{
		Name: 			 DHCPProxyDescriptorName,
		KeySelector: 	 l3.ModelDHCPProxy.IsKeyValid,
		KeyLabel: 		 l3.ModelDHCPProxy.StripKeyPrefix,
		NBKeyPrefix:     l3.ModelDHCPProxy.KeyPrefix(),
		ValueTypeName:   l3.ModelDHCPProxy.ProtoName(),
		Create:          ctx.Create,
		Delete:          ctx.Delete,
		Retrieve:        ctx.Retrieve,
		Dependencies: 	 ctx.Dependencies,
	}
	return adapter.NewDHCPProxyDescriptor(typedDescr)
}

// Dependencies lists dependencies for a VPP DHCP proxy.
func (d *DHCPProxyDescriptor) Dependencies(key string, value *l3.DHCPProxy)  (deps []kvs.Dependency) {
	//todo implement method
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

	retrieved = append(retrieved, adapter.DHCPProxyKVWithMetadata{
		Key:    l3.DHCPProxyKey(),
		Value:  dhcpProxyDetails.DHCPProxy,
		Origin: kvs.FromNB,
	})

	return retrieved, nil
}

