// Copyright (c) 2017 Cisco and/or its affiliates.
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

package etcdv3

import (
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/cn-infra/db/keyval/plugin"
	"github.com/ligato/cn-infra/flavors/local"
	"github.com/ligato/cn-infra/health/statuscheck"
	"github.com/ligato/cn-infra/servicelabel"
	"github.com/ligato/cn-infra/utils/safeclose"
)

const (
	// healthCheckProbeKey is a key used to probe Etcd state
	healthCheckProbeKey string = "/probe-etcd-connection"
)

// Plugin implements Plugin interface therefore can be loaded with other plugins
type Plugin struct {
	Deps // inject
	*plugin.Skeleton
	disabled bool
}

// Deps is here to group injected dependencies of plugin
// to not mix with other plugin fields.
type Deps struct {
	local.PluginInfraDeps // inject
}

// Init is called at plugin startup. The connection to etcd is established.
func (p *Plugin) Init() (err error) {
	// Retrieve config
	var cfg Config
	found, err := p.PluginConfig.GetValue(&cfg)
	if !found {
		p.Log.Info("etcd config not found ", p.PluginConfig.GetConfigName(), " - skip loading this plugin")
		p.disabled = true
		return nil
	}
	if err != nil {
		return err
	}

	// Init connection
	etcdConfig, err := ConfigToClientv3(&cfg)
	if err != nil {
		return err
	}

	if p.Skeleton == nil {
		con, err := NewEtcdConnectionWithBytes(*etcdConfig, p.Log)
		if err != nil {
			return err
		}

		p.Skeleton = plugin.NewSkeleton(p.String(),
			p.ServiceLabel,
			con,
		)
	}
	err = p.Skeleton.Init()
	if err != nil {
		return err
	}

	return nil
}

// AfterInit is called by the Agent Core after all plugins have been initialized.
func (p *Plugin) AfterInit() error {
	if p.disabled {
		return nil
	}

	// Register for providing status reports (polling mode)
	if p.StatusCheck != nil {
		p.StatusCheck.Register(core.PluginName(p.String()), func() (statuscheck.PluginState, error) {
			_, _, err := p.Skeleton.NewBroker("/").GetValue(healthCheckProbeKey, nil)
			if err == nil {
				return statuscheck.OK, nil
			}
			return statuscheck.Error, err
		})
	} else {
		p.Log.Warnf("Unable to start status check for etcd")
	}

	return nil
}

// FromExistingConnection is used mainly for testing
func FromExistingConnection(connection keyval.CoreBrokerWatcher, sl servicelabel.ReaderAPI) *Plugin {
	skel := plugin.NewSkeleton("testing", sl, connection)
	return &Plugin{Skeleton: skel}
}

// Close resources
func (p *Plugin) Close() error {
	_, err := safeclose.CloseAll(p.Skeleton)
	return err
}

// String returns if set Deps.PluginName or "kvdbsync" otherwise
func (p *Plugin) String() string {
	if len(p.Deps.PluginName) == 0 {
		return "kvdbsync"
	}
	return string(p.Deps.PluginName)
}

// Disabled if the plugin was not found
func (p *Plugin) Disabled() (disabled bool) {
	return p.disabled
}
