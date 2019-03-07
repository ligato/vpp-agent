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

// IfDescriptor defines all dependencies used in descriptor methods (vpp handlers, loggers, ...)
type IfDescriptor struct {
	// dependencies
	log logging.PluginLogger
}

// NewIfDescriptor creates a new instance of the descriptor
func NewIfDescriptor(logger logging.PluginLogger) *IfDescriptor {
	return &IfDescriptor{
		log: logger,
	}
}

// GetDescriptor returns the type-safe descriptor
func (d *IfDescriptor) GetDescriptor() *adapter.InterfaceDescriptor {
	return &adapter.InterfaceDescriptor{
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
			d.log.Infof("Interface %s created", value.Name)
			// Return interface name so the scheduler remembers it
			return value.Name, nil
		},
	}
}

/* Route Descriptor */

const (
	routeDescriptorName    = "route-descriptor"
	routePrefix            = "/route/"
	routeInterfaceDepLabel = "route-interface"
)

type RouteDescriptor struct {
	// dependencies
	log logging.PluginLogger
}

func NewRouteDescriptor(logger logging.PluginLogger) *RouteDescriptor {
	return &RouteDescriptor{
		log: logger,
	}
}

func (d *RouteDescriptor) GetDescriptor() *adapter.RouteDescriptor {
	return &adapter.RouteDescriptor{
		// Descriptor name, must be unique across all descriptors
		Name: routeDescriptorName,
		// Prefix for the descriptor-specific configuration
		NBKeyPrefix: routePrefix,
		// A string value defining descriptor type
		ValueTypeName: proto.MessageName(&model.Route{}),
		// A unique identifier of the configuration (name, label)
		KeyLabel: func(key string) string {
			return strings.TrimPrefix(key, routePrefix)
		},
		// Returns true if the provided key is relevant for this descriptor is some way
		KeySelector: func(key string) bool {
			if strings.HasPrefix(key, routePrefix) {
				return true
			}
			return false
		},
		// All other keys that must exist before the item is configured
		Dependencies: func(key string, value *model.Route) []api.Dependency {
			return []api.Dependency{
				{
					Label: routeInterfaceDepLabel,
					Key:   ifPrefix + value.InterfaceName,
				},
			}
		},
		// A list of descriptors expected to handle dependencies
		RetrieveDependencies: []string{ifDescriptorName},
		// Add a new configuration item
		Create: func(key string, value *model.Route) (metadata interface{}, err error) {
			d.log.Infof("Created route %s dependent on interface %s", value.Name, value.InterfaceName)
			return nil, nil
		},
	}
}
