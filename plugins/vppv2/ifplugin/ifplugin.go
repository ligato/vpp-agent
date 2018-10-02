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

//go:generate descriptor-adapter --descriptor-name Interface  --value-type *interfaces.Interface --meta-type *ifaceidx.IfaceMetadata --import "../model/interfaces" --import "ifaceidx" --output-dir "descriptor"

package ifplugin

import (
	"os"
	"errors"

	"github.com/ligato/cn-infra/infra"
	"github.com/ligato/cn-infra/datasync"

	scheduler "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/vppv2/ifplugin/descriptor"
	"github.com/ligato/vpp-agent/plugins/vppv2/ifplugin/descriptor/adapter"
	"github.com/ligato/vpp-agent/plugins/vppv2/ifplugin/ifaceidx"
	linux_ifplugin "github.com/ligato/vpp-agent/plugins/linuxv2/ifplugin"
	"github.com/ligato/vpp-agent/plugins/govppmux"
)

const (
	// vppStatusPublishersEnv is the name of the environment variable used to
	// override state publishers from the configuration file.
	vppStatusPublishersEnv = "VPP_STATUS_PUBLISHERS"
)

// IfPlugin configures VPP interfaces using GoVPP.
type IfPlugin struct {
	Deps

	// descriptors
	ifDescriptor *descriptor.InterfaceDescriptor

	// index map
	intfIndex ifaceidx.IfaceMetadataIndex

	// From config file
	enableStopwatch bool
	ifMtu           uint32
}

// Deps lists dependencies of the interface plugin.
type Deps struct {
	infra.PluginDeps
	Scheduler         scheduler.KVScheduler
	GoVppmux          govppmux.API
	PublishStatistics datasync.KeyProtoValWriter
	DataSyncs         map[string]datasync.KeyProtoValWriter
	LinuxIfPlugin     linux_ifplugin.API /* optional, provide if TAP+AUTO_TAP interfaces are used */
}

// Config holds the vpp-plugin configuration.
type Config struct {
	Mtu              uint32   `json:"mtu"`
	Stopwatch        bool     `json:"stopwatch"`
	StatusPublishers []string `json:"status-publishers"`
}

// Init loads configuration file and registers interface-related descriptors.
func (p *IfPlugin) Init() error {
	// Read config file and set all related fields
	p.fromConfigFile()

	// init & register descriptors
	p.ifDescriptor = descriptor.NewInterfaceDescriptor(p.ifMtu, p.LinuxIfPlugin, p.Log)
	ifDescriptor := adapter.NewInterfaceDescriptor(p.ifDescriptor.GetDescriptor())
	p.Deps.Scheduler.RegisterKVDescriptor(ifDescriptor)

	// obtain read-only reference to index map
	var withIndex bool
	metadataMap := p.Deps.Scheduler.GetMetadataMap(ifDescriptor.Name)
	p.intfIndex, withIndex = metadataMap.(ifaceidx.IfaceMetadataIndex)
	if !withIndex {
		return errors.New("missing index with interface metadata")
	}
	return nil
}

// GetInterfaceIndex gives read-only access to map with metadata of all configured
// VPP interfaces.
func (p *IfPlugin) GetInterfaceIndex() ifaceidx.IfaceMetadataIndex {
	return p.intfIndex
}

// fromConfigFile loads plugin attributes from the configuration file.
func (p *IfPlugin) fromConfigFile() {
	config, err := p.loadConfig()
	if err != nil {
		p.Log.Errorf("Error reading %v config file: %v", p.PluginName, err)
		return
	}
	if config != nil {
		publishers := datasync.KVProtoWriters{}
		for _, pub := range config.StatusPublishers {
			db, found := p.Deps.DataSyncs[pub]
			if !found {
				p.Log.Warnf("Unknown status publisher %q from config", pub)
				continue
			}
			publishers = append(publishers, db)
			p.Log.Infof("Added status publisher %q from config", pub)
		}
		p.Deps.PublishStatistics = publishers
		if config.Mtu != 0 {
			p.ifMtu = config.Mtu
			p.Log.Infof("Default MTU set to %v", p.ifMtu)
		}

		if config.Stopwatch {
			p.enableStopwatch = true
			p.Log.Info("stopwatch enabled for %v", p.PluginName)
		} else {
			p.Log.Info("stopwatch disabled for %v", p.PluginName)
		}
	} else {
		p.Log.Infof("stopwatch disabled for %v", p.PluginName)
	}
}

// loadConfig loads configuration file.
func (p *IfPlugin) loadConfig() (*Config, error) {
	config := &Config{}

	found, err := p.Cfg.LoadValue(config)
	if err != nil {
		return nil, err
	} else if !found {
		p.Log.Debugf("%v config not found", p.PluginName)
		return nil, nil
	}
	p.Log.Debugf("%v config found: %+v", p.PluginName, config)

	if pubs := os.Getenv(vppStatusPublishersEnv); pubs != "" {
		p.Log.Debugf("status publishers from env: %v", pubs)
		config.StatusPublishers = append(config.StatusPublishers, pubs)
	}

	return config, err
}
