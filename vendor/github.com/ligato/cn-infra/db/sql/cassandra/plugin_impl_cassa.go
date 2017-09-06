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

package cassandra

import (
	"github.com/ligato/cn-infra/db/sql"
	"github.com/ligato/cn-infra/flavors/local"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/willfaught/gockle"
)

// Plugin implements Plugin interface therefore can be loaded with other plugins
type Plugin struct {
	Deps // inject

	clientConfig *ClientConfig
	session      gockle.Session
}

// Deps is here to group injected dependencies of plugin
// to not mix with other plugin fields.
type Deps struct {
	local.PluginInfraDeps // inject
}

// Init is called at plugin startup. The session to etcd is established.
func (p *Plugin) Init() (err error) {
	if p.session != nil {
		return nil // skip initialization
	}

	// Retrieve config
	var cfg Config
	found, err := p.PluginConfig.GetValue(&cfg)
	// need to be strict about config presence for ETCD
	if !found {
		p.Log.Info("cassandra client config not found ", p.PluginConfig.GetConfigName(),
			" - skip loading this plugin")
		return nil
	}
	if err != nil {
		return err
	}

	// Init session
	p.clientConfig, err = ConfigToClientConfig(&cfg)
	if err != nil {
		return err
	}

	return nil
}

// AfterInit is called by the Agent Core after all plugins have been initialized.
func (p *Plugin) AfterInit() error {
	if p.session == nil && p.clientConfig != nil {
		session, err := CreateSessionFromConfig(p.clientConfig)
		if err != nil {
			return err
		}

		p.session = gockle.NewSession(session)
	}

	/* TODO Register for providing status reports (polling mode)
	if p.StatusCheck != nil && p.session != nil {
		p.StatusCheck.Register(core.PluginName(p.String()), func() (statuscheck.PluginState, error) {
			_, _, err := p.Skeleton.NewBroker("/").GetValue(healthCheckProbeKey, nil)
			if err == nil {
				return statuscheck.OK, nil
			}
			return statuscheck.Error, err
		})
	} else {
		p.Log.Warnf("Unable to start status check for etcd")
	}*/

	return nil
}

// FromExistingSession is used mainly for testing
func FromExistingSession(session gockle.Session) *Plugin {
	return &Plugin{session: session}
}

// NewBroker returns a Broker instance to work with Cassandra Data Base
func (p *Plugin) NewBroker() sql.Broker {
	return NewBrokerUsingSession(p.session)
}

// Close resources
func (p *Plugin) Close() error {
	_, err := safeclose.CloseAll(p.session)
	return err
}

// String returns if set Deps.PluginName or "cassa-client" otherwise
func (p *Plugin) String() string {
	if len(p.Deps.PluginName) == 0 {
		return "cassa-client"
	}
	return string(p.Deps.PluginName)
}
