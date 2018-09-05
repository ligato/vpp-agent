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

//go:generate protoc --proto_path=../model/namespace --gogo_out=../model/namespace namespace.proto

package nsplugin

import (
	"fmt"

	"github.com/go-errors/errors"
	"github.com/vishvananda/netns"

	"github.com/ligato/cn-infra/infra"
	scheduler "github.com/ligato/cn-infra/kvscheduler/api"
	"github.com/ligato/cn-infra/logging"

	"github.com/ligato/cn-infra/logging/measure"
	nsmodel "github.com/ligato/vpp-agent/plugins/linuxv2/model/namespace"
	"github.com/ligato/vpp-agent/plugins/linuxv2/nsplugin/descriptor"
	nsLinuxcalls "github.com/ligato/vpp-agent/plugins/linuxv2/nsplugin/linuxcalls"
)

// NsPlugin is a plugin to handle namespaces and microservices for other linux
// plugins (ifplugin, l3plugin ...).
// It does not follow the standard concept of CRUD, but provides a set of methods
// other plugins can use to manage namespaces.
type NsPlugin struct {
	Deps

	// From configuration file
	disabled  bool
	stopwatch *measure.Stopwatch

	// Default namespace
	defaultNs netns.NsHandle

	// Handlers
	sysHandler     nsLinuxcalls.SystemAPI
	namedNsHandler nsLinuxcalls.NamedNetNsAPI

	// Descriptor
	msDescriptor *descriptor.MicroserviceDescriptor
}

// Deps lists dependencies of the NsPlugin.
type Deps struct {
	infra.PluginDeps
	Scheduler scheduler.KVScheduler
}

// Config holds the nsplugin configuration.
type Config struct {
	Stopwatch bool `json:"stopwatch"`
	Disabled  bool `json:"disabled"`
}

// unavailableMicroserviceErr is error implementation used when a given microservice is not deployed.
type unavailableMicroserviceErr struct {
	label string
}

func (e *unavailableMicroserviceErr) Error() string {
	return fmt.Sprintf("Microservice '%s' is not available", e.label)
}

// Init namespace handler caches and create config namespace
func (plugin *NsPlugin) Init() error {
	// Parse configuration file
	config, err := plugin.retrieveConfig()
	if err != nil {
		return err
	}
	if config != nil {
		if config.Disabled {
			plugin.disabled = true
			plugin.Log.Infof("Disabling Linux Namespace plugin")
			return nil
		}
		if config.Stopwatch {
			plugin.Log.Infof("stopwatch enabled for %v", plugin.PluginName)
			plugin.stopwatch = measure.NewStopwatch("Linux-NsPlugin", plugin.Log)
		} else {
			plugin.Log.Infof("stopwatch disabled for %v", plugin.PluginName)
		}
	} else {
		plugin.Log.Infof("stopwatch disabled for %v", plugin.PluginName)
	}

	// Handlers
	plugin.sysHandler = nsLinuxcalls.NewSystemHandler(plugin.stopwatch)
	plugin.namedNsHandler = nsLinuxcalls.NewNamedNetNsHandler(plugin.sysHandler, plugin.Log, plugin.stopwatch)

	// Default namespace
	plugin.defaultNs, err = plugin.sysHandler.GetCurrentNamespace()
	if err != nil {
		return errors.Errorf("failed to init default namespace: %v", err)
	}

	// Microservice descriptor
	plugin.msDescriptor, err = descriptor.NewMicroserviceDescriptor(plugin.Scheduler, plugin.Log)
	if err != nil {
		return err
	}
	plugin.Scheduler.RegisterKVDescriptor(plugin.msDescriptor)
	plugin.msDescriptor.StartTracker()

	plugin.Log.Infof("Namespace plugin initialized")

	return nil
}

// Close stops microservice tracker
func (plugin *NsPlugin) Close() error {
	if plugin.disabled {
		return nil
	}
	plugin.msDescriptor.StopTracker()

	return nil
}

// GetNamespaceHandle returns low-level run-time handle for the given namespace
// to be used with Netlink API. Do not forget to eventually close the handle using
// the netns.NsHandle.Close() method.
func (plugin *NsPlugin) GetNamespaceHandle(ctx nsLinuxcalls.NamespaceMgmtCtx, namespace *nsmodel.Namespace) (handle netns.NsHandle, err error) {
	if plugin.disabled {
		return 0, errors.New("NsPlugin is disabled")
	}
	// Convert microservice namespace
	if namespace != nil && namespace.Type == nsmodel.Namespace_MICROSERVICE_REF_NS {
		// Convert namespace
		namespace = plugin.convertMicroserviceNsToPidNs(namespace.Microservice)
		if namespace == nil {
			return 0, &unavailableMicroserviceErr{}
		}
	}

	// Get network namespace file descriptor
	ns, err := plugin.getOrCreateNs(ctx, namespace)
	if err != nil {
		return 0, errors.Errorf("failed to get or create namespace %s: %v", namespace.Name, err)
	}

	return ns, nil
}

// SwitchToNamespace switches the network namespace of the current thread.
// Caller should eventually call the returned "revert" function in order to get back to the original
// network namespace (for example using "defer revert()").
func (plugin *NsPlugin) SwitchToNamespace(ctx nsLinuxcalls.NamespaceMgmtCtx, ns *nsmodel.Namespace) (revert func(), err error) {
	if plugin.disabled {
		return func() {}, errors.New("NsPlugin is disabled")
	}

	// Save the current network namespace.
	origns, err := netns.Get()
	if err != nil {
		return func() {}, err
	}

	// Get network namespace file descriptor.
	nsHandle, err := plugin.GetNamespaceHandle(ctx, ns)
	if err != nil {
		origns.Close()
		return func() {}, err
	}
	defer nsHandle.Close()

	// Lock the OS Thread so we don't accidentally switch namespaces later.
	ctx.LockOSThread()

	// Switch the namespace.
	l := plugin.Log.WithFields(logging.Fields{"ns": nsHandle.String(), "ns-fd": int(nsHandle)})
	if err := plugin.sysHandler.SetNamespace(nsHandle); err != nil {
		ctx.UnlockOSThread()
		origns.Close()
		l.Errorf("Failed to switch Linux network namespace (%v): %v", ns, err)
		return func() {}, err
	}

	return func() {
		l := plugin.Log.WithFields(logging.Fields{"orig-ns": origns.String(), "orig-ns-fd": int(origns)})
		if err := plugin.sysHandler.SetNamespace(origns); err != nil {
			l.Errorf("Failed to switch Linux network namespace: %v", err)
		}
		origns.Close()
		ctx.UnlockOSThread()
	}, nil
}

// retrieveConfig loads NsPlugin configuration file.
func (plugin *NsPlugin) retrieveConfig() (*Config, error) {
	config := &Config{}
	found, err := plugin.Cfg.LoadValue(config)
	if !found {
		plugin.Log.Debug("Linux NsPlugin config not found")
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	plugin.Log.Debug("Linux NsPlugin config found")
	return config, err
}

// getOrCreateNs returns an existing Linux network namespace or creates a new one if it doesn't exist yet.
// It is, however, only possible to create "named" namespaces. For PID-based namespaces, process with
// the given PID must exists, otherwise the function returns an error.
func (plugin *NsPlugin) getOrCreateNs(ctx nsLinuxcalls.NamespaceMgmtCtx, ns *nsmodel.Namespace) (netns.NsHandle, error) {
	var nsHandle netns.NsHandle
	var err error

	if ns == nil {
		return plugin.sysHandler.DuplicateNamespaceHandle(plugin.defaultNs)
	}

	switch ns.Type {
	case nsmodel.Namespace_PID_REF_NS:
		if ns.Pid == 0 {
			// We consider PID 0 as the representation of the default namespace.
			return plugin.sysHandler.DuplicateNamespaceHandle(plugin.defaultNs)
		}
		nsHandle, err = plugin.sysHandler.GetNamespaceFromPid(int(ns.Pid))
		if err != nil {
			return netns.None(), errors.Errorf("failed to get namespace handle from pid: %v", err)
		}
	case nsmodel.Namespace_NAMED_NS:
		if ns.Name == "" {
			return plugin.sysHandler.DuplicateNamespaceHandle(plugin.defaultNs)
		}
		nsHandle, err = plugin.sysHandler.GetNamespaceFromName(ns.Name)
		if err != nil {
			// Create named namespace if it doesn't exist yet.
			_, err = plugin.namedNsHandler.CreateNamedNetNs(ctx, ns.Name)
			if err != nil {
				return netns.None(), errors.Errorf("failed to create named net namspace: %v", err)
			}
			nsHandle, err = plugin.sysHandler.GetNamespaceFromName(ns.Name)
			if err != nil {
				return netns.None(), errors.Errorf("unable to get namespace by name")
			}
		}
	case nsmodel.Namespace_FILE_REF_NS:
		if ns.Filepath == "" {
			return plugin.sysHandler.DuplicateNamespaceHandle(plugin.defaultNs)
		}
		nsHandle, err = plugin.sysHandler.GetNamespaceFromPath(ns.Filepath)
		if err != nil {
			return netns.None(), errors.Errorf("failed to get file %s from path: %v", ns.Filepath, err)
		}
	case nsmodel.Namespace_MICROSERVICE_REF_NS:
		return netns.None(), errors.Errorf("unable to convert microservice label to PID at this level")
	}

	return nsHandle, nil
}

// convertMicroserviceNsToPidNs converts microservice-referenced namespace into the PID-referenced namespace.
func (plugin *NsPlugin) convertMicroserviceNsToPidNs(microserviceLabel string) (pidNs *nsmodel.Namespace) {
	if microservice, found := plugin.msDescriptor.GetMicroserviceStateData(microserviceLabel); found {
		pidNamespace := &nsmodel.Namespace{}
		pidNamespace.Type = nsmodel.Namespace_PID_REF_NS
		pidNamespace.Pid = uint32(microservice.PID)
		return pidNamespace
	}
	return nil
}