package main

import (
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/vpp-agent/examples/tutorials/05_kv-scheduler/adapter"
	"github.com/ligato/vpp-agent/examples/tutorials/05_kv-scheduler/model"
	"github.com/ligato/vpp-agent/plugins/kvscheduler/api"
)

/* Interface Descriptor */

const (
	ifDescriptorName = "if-descriptor"
	ifPrefix         = "/interface/"
)

// NewIfDescriptor creates a new instance of the descriptor
func NewIfDescriptor(logger logging.PluginLogger) *api.KVDescriptor {
	// convert typed descriptor into generic descriptor API using adapter
	typedDescriptor := &adapter.InterfaceDescriptor{
		// Descriptor name, must be unique across all descriptors
		Name: ifDescriptorName,
		// Prefix for the descriptor-specific configuration
		NBKeyPrefix: ifPrefix,
		// A string value defining descriptor type
		ValueTypeName: proto.MessageName(&model.Interface{}),
		// A unique identifier of the configuration (name, label)
		KeyLabel: func(key string) string {
			return strings.TrimPrefix(key, ifPrefix)
		},
		// Returns true if the provided key is relevant for this descriptor is some way
		KeySelector: func(key string) bool {
			if strings.HasPrefix(key, ifPrefix) {
				return true
			}
			return false
		},
		// Enables metadata in the KV scheduler
		WithMetadata: true,
		// Add a new configuration item
		Create: func(key string, value *model.Interface) (metadata interface{}, err error) {
			logger.Infof("Interface %s created", value.Name)
			// Return interface name so the scheduler remembers it
			return value.Name, nil
		},
	}
	return adapter.NewInterfaceDescriptor(typedDescriptor)
}

/* Route Descriptor */

const (
	routeDescriptorName    = "route-descriptor"
	routePrefix            = "/route/"
	routeInterfaceDepLabel = "route-interface"
)

// RouteDescriptor is a descriptor object with
type RouteDescriptor struct {
	// dependencies
	log logging.PluginLogger
}

// GetDescriptor returns type safe descriptor structure
func NewRouteDescriptor(logger logging.PluginLogger) *api.KVDescriptor {
	descriptorCtx := &RouteDescriptor{
		log: logger,
	}
	typedDescriptor := &adapter.RouteDescriptor{
		// Descriptor name, must be unique across all descriptors
		Name: routeDescriptorName,
		// Prefix for the descriptor-specific configuration
		NBKeyPrefix: routePrefix,
		// A string value defining descriptor type
		ValueTypeName: proto.MessageName(&model.Route{}),
		// A unique identifier of the configuration (name, label)
		KeyLabel: descriptorCtx.KeyLabel,
		// Returns true if the provided key is relevant for this descriptor is some way
		KeySelector: descriptorCtx.KeySelector,
		// All other keys that must exist before the item is configured
		Dependencies: descriptorCtx.Dependencies,
		// A list of descriptors expected to handle dependencies
		RetrieveDependencies: []string{ifDescriptorName},
		// Add a new configuration item
		Create: descriptorCtx.Create,
	}
	return adapter.NewRouteDescriptor(typedDescriptor)
}

func (d *RouteDescriptor) KeyLabel(key string) string {
	return strings.TrimPrefix(key, routePrefix)
}

func (d *RouteDescriptor) KeySelector(key string) bool {
	if strings.HasPrefix(key, routePrefix) {
		return true
	}
	return false
}

func (d *RouteDescriptor) Dependencies(key string, value *model.Route) []api.Dependency {
	return []api.Dependency{
		{
			Label: routeInterfaceDepLabel,
			Key:   ifPrefix + value.InterfaceName,
		},
	}
}

func (d *RouteDescriptor) Create(key string, value *model.Route) (metadata interface{}, err error) {
	d.log.Infof("Created route %s dependent on interface %s", value.Name, value.InterfaceName)
	return nil, nil
}
