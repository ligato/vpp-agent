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

package linuxlocal

import (
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/datasync/resync"
	"github.com/ligato/vpp-agent/clientv1/linux/localclient"
	"github.com/ligato/vpp-agent/plugins/defaultplugins"
	"github.com/ligato/vpp-agent/plugins/govppmux"
	"github.com/ligato/vpp-agent/plugins/linuxplugin"
	"github.com/ligato/cn-infra/flavors/generic"
	"github.com/ligato/cn-infra/db/keyval/etcdv3"
)

// Flavor glues together multiple plugins to mange VPP and linux interfaces configuration using local client.
type Flavor struct {
	injected         bool
	Generic 		 generic.Flavor
	Etcd    		 etcdv3.Plugin
	LinuxLocalClient localclient.Plugin
	Resync           resync.Plugin
	GoVPP            govppmux.GOVPPPlugin
	Linux            linuxplugin.Plugin
	VPP              defaultplugins.Plugin
}

// Inject sets object references
func (f *Flavor) Inject() error {
	if f.injected {
		return nil
	}
	f.injected = true
	f.Generic.Inject()

	f.Etcd.LogFactory = &f.Generic.Logrus
	f.Etcd.ServiceLabel = &f.Generic.ServiceLabel
	f.Etcd.StatusCheck = &f.Generic.StatusCheck


	f.GoVPP.StatusCheck = &f.Generic.StatusCheck
	f.GoVPP.LogFactory = &f.Generic.Logrus
	f.VPP.ServiceLabel = &f.Generic.ServiceLabel
	f.VPP.Linux = &f.Linux
	f.VPP.GoVppmux = &f.GoVPP

	return nil
}

// Plugins combines Generic Plugins and Standard VPP Plugins + (their ETCD Connector/Adapter with RESYNC)
func (f *Flavor) Plugins() []*core.NamedPlugin {
	f.Inject()
	return core.ListPluginsInFlavor(f)
}
