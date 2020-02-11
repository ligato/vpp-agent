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

//go:generate descriptor-adapter --descriptor-name Interface  --value-type *linux_interfaces.Interface --meta-type *ifaceidx.LinuxIfMetadata --import "go.ligato.io/vpp-agent/v3/proto/ligato/linux/interfaces" --import "go.ligato.io/vpp-agent/v3/plugins/linux/ifplugin/ifaceidx" --output-dir "descriptor"

package ifplugin

import (
	"github.com/pkg/errors"

	"go.ligato.io/cn-infra/v2/infra"
	"go.ligato.io/cn-infra/v2/servicelabel"

	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/plugins/linux/ifplugin/descriptor"
	"go.ligato.io/vpp-agent/v3/plugins/linux/ifplugin/ifaceidx"
	"go.ligato.io/vpp-agent/v3/plugins/linux/ifplugin/linuxcalls"
	"go.ligato.io/vpp-agent/v3/plugins/linux/nsplugin"
	"go.ligato.io/vpp-agent/v3/plugins/netalloc"
)

const (
	// by default, at most 10 go routines will split the configured namespaces
	// to execute the Retrieve operation in parallel.
	defaultGoRoutinesCnt = 10
)

// IfPlugin configures Linux VETH and TAP interfaces using Netlink API.
type IfPlugin struct {
	Deps

	// From configuration file
	disabled bool

	// system handlers
	ifHandler linuxcalls.NetlinkAPI

	// descriptors
	ifDescriptor     *descriptor.InterfaceDescriptor
	ifWatcher        *descriptor.InterfaceWatcher
	ifAddrDescriptor *descriptor.InterfaceAddressDescriptor

	// index map
	ifIndex ifaceidx.LinuxIfMetadataIndex
}

// Deps lists dependencies of the interface plugin.
type Deps struct {
	infra.PluginDeps
	ServiceLabel servicelabel.ReaderAPI
	KVScheduler  kvs.KVScheduler
	NsPlugin     nsplugin.API
	AddrAlloc    netalloc.AddressAllocator
	VppIfPlugin  descriptor.VPPIfPluginAPI /* mandatory if TAP_TO_VPP interfaces are used */
}

// Config holds the ifplugin configuration.
type Config struct {
	Disabled      bool `json:"disabled"`
	GoRoutinesCnt int  `json:"go-routines-count"`
}

// Init registers interface-related descriptors and starts watching of the default
// network namespace for interface changes.
func (p *IfPlugin) Init() error {
	// parse configuration file
	config, err := p.retrieveConfig()
	if err != nil {
		return err
	}
	p.Log.Debugf("Linux interface plugin config: %+v", config)
	if config.Disabled {
		p.disabled = true
		p.Log.Infof("Disabling Linux Interface plugin")
		return nil
	}

	// init & register interface descriptor
	var ifDescriptor *kvs.KVDescriptor
	ifDescriptor, p.ifDescriptor = descriptor.NewInterfaceDescriptor(p.ServiceLabel, p.NsPlugin, p.VppIfPlugin,
		p.AddrAlloc, p.Log)
	err = p.Deps.KVScheduler.RegisterKVDescriptor(ifDescriptor)
	if err != nil {
		return err
	}

	// obtain read-only reference to index map
	var withIndex bool
	metadataMap := p.Deps.KVScheduler.GetMetadataMap(ifDescriptor.Name)
	p.ifIndex, withIndex = metadataMap.(ifaceidx.LinuxIfMetadataIndex)
	if !withIndex {
		return errors.New("missing index with interface metadata")
	}

	// init handler and pass it to the interface descriptor
	p.ifHandler = linuxcalls.NewNetLinkHandler(p.NsPlugin, p.ifIndex, p.ServiceLabel.GetAgentPrefix(),
		config.GoRoutinesCnt, p.Log)
	p.ifDescriptor.SetInterfaceHandler(p.ifHandler)

	var addrDescriptor *kvs.KVDescriptor
	addrDescriptor, p.ifAddrDescriptor = descriptor.NewInterfaceAddressDescriptor(p.NsPlugin,
		p.AddrAlloc, p.ifHandler, p.Log)
	err = p.Deps.KVScheduler.RegisterKVDescriptor(addrDescriptor)
	if err != nil {
		return err
	}

	p.ifWatcher = descriptor.NewInterfaceWatcher(p.KVScheduler, p.ifHandler, p.Log)
	err = p.Deps.KVScheduler.RegisterKVDescriptor(p.ifWatcher.GetDescriptor())
	if err != nil {
		return err
	}

	// pass read-only index map to descriptors
	p.ifDescriptor.SetInterfaceIndex(p.ifIndex)
	p.ifAddrDescriptor.SetInterfaceIndex(p.ifIndex)

	// start interface watching
	if err = p.ifWatcher.StartWatching(); err != nil {
		return err
	}

	return nil
}

// Close stops watching of the default network namespace.
func (p *IfPlugin) Close() error {
	if p.disabled {
		return nil
	}
	p.ifWatcher.StopWatching()
	return nil
}

// GetInterfaceIndex gives read-only access to map with metadata of all configured
// linux interfaces.
func (p *IfPlugin) GetInterfaceIndex() ifaceidx.LinuxIfMetadataIndex {
	return p.ifIndex
}

// retrieveConfig loads IfPlugin configuration file.
func (p *IfPlugin) retrieveConfig() (*Config, error) {
	config := &Config{
		// default configuration
		GoRoutinesCnt: defaultGoRoutinesCnt,
	}
	found, err := p.Cfg.LoadValue(config)
	if !found {
		p.Log.Debug("Linux IfPlugin config not found")
		return config, nil
	}
	if err != nil {
		return nil, err
	}
	p.Log.Debug("Linux IfPlugin config found")
	return config, err
}
