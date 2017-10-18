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

// Plugin implements etcdv3 plugin.
type Plugin struct {
	Deps // inject
	*plugin.Skeleton
	disabled   bool
	connection keyval.CoreBrokerWatcher
}

// Deps lists dependencies of the etcdv3 plugin.
// If injected, etcd plugin will use StatusCheck to signal the connection status.
type Deps struct {
	local.PluginInfraDeps // inject
}

// Init retrieves etcd configuration and establishes a new connection
// with the etcd data store.
// If the configuration file doesn't exist or cannot be read, the returned error
// will be of type os.PathError. An untyped error is returned in case the file
// doesn't contain a valid YAML configuration.
// The function may also return error if TLS connection is selected and the
// CA or client certificate is not accessible(os.PathError)/valid(untyped).
// Check clientv3.New from coreos/etcd for possible errors returned when
// the connection cannot be established.
func (p *Plugin) Init() (err error) {
	// Init connection
	if p.Skeleton == nil {
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
		etcdConfig, err := ConfigToClientv3(&cfg)
		if err != nil {
			return err
		}

		p.connection, err = NewEtcdConnectionWithBytes(*etcdConfig, p.Log)
		if err != nil {
			return err
		}

		p.Skeleton = plugin.NewSkeleton(p.String(),
			p.ServiceLabel,
			p.connection,
		)
	}
	err = p.Skeleton.Init()
	if err != nil {
		return err
	}

	// Register for providing status reports (polling mode)
	if p.StatusCheck != nil {
		p.StatusCheck.Register(core.PluginName(p.String()), func() (statuscheck.PluginState, error) {
			_, _, _, err := p.connection.GetValue(healthCheckProbeKey)
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

// AfterInit registers status polling function with StatusCheck plugin
// (if injected).
func (p *Plugin) AfterInit() error {
	if p.disabled {
		return nil
	}

	return nil
}

// FromExistingConnection is used mainly for testing to inject existing
// connection into the plugin.
// Note, need to set Deps for returned value!
func FromExistingConnection(connection keyval.CoreBrokerWatcher, sl servicelabel.ReaderAPI) *Plugin {
	skel := plugin.NewSkeleton("testing", sl, connection)
	return &Plugin{Skeleton: skel, connection: connection}
}

// Close shutdowns the connection.
func (p *Plugin) Close() error {
	_, err := safeclose.CloseAll(p.Skeleton)
	return err
}

// String returns the plugin name from dependencies if injected,
// "kvdbsync" otherwise.
func (p *Plugin) String() string {
	if len(p.Deps.PluginName) == 0 {
		return "kvdbsync"
	}
	return string(p.Deps.PluginName)
}

// Disabled returns *true* if the plugin is not in use due to missing
// etcd configuration.
func (p *Plugin) Disabled() (disabled bool) {
	return p.disabled
}
