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

package local

import (
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/datasync/resync"
	"github.com/ligato/cn-infra/flavors/etcdkafka"
	"github.com/ligato/vpp-agent/clientv1/defaultplugins/localclient"
	"github.com/ligato/vpp-agent/plugins/defaultplugins"
	"github.com/ligato/vpp-agent/plugins/govppmux"
)

// Flavor glues together multiple plugins to mange VPP configuration using local client.
type Flavor struct {
	Base        etcdkafka.FlavorEtcdKafka
	LocalClient localclient.Plugin
	Resync      resync.Plugin
	GoVPP       govppmux.GOVPPPlugin
	VPP         defaultplugins.Plugin

	injected bool
}

// Inject sets object references
func (f *Flavor) Inject() error {
	if f.injected {
		return nil
	}
	f.injected = true

	f.Base.Inject(nil)

	f.GoVPP.Deps.PluginInfraDeps = *f.Base.FlavorLocal.InfraDeps("govpp")
	f.VPP.Deps.PluginInfraDeps = *f.Base.FlavorLocal.InfraDeps("default-plugins")
	f.VPP.GoVppmux = &f.GoVPP

	return nil
}

// Plugins combines Generic Plugins and Standard VPP Plugins
func (f *Flavor) Plugins() []*core.NamedPlugin {
	f.Inject()
	return core.ListPluginsInFlavor(f)
}
