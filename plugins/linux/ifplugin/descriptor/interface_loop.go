// Copyright (c) 2019 Cisco and/or its affiliates.
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

package descriptor

import (
	interfaces "github.com/ligato/vpp-agent/api/models/linux/interfaces"
	"github.com/ligato/vpp-agent/plugins/linux/ifplugin/ifaceidx"

	nslinuxcalls "github.com/ligato/vpp-agent/plugins/linux/nsplugin/linuxcalls"
	"strings"
)

// createLoopback adds logical name as alias to linux loopback interface
func (d *InterfaceDescriptor) createLoopback(nsCtx nslinuxcalls.NamespaceMgmtCtx,
	linuxIf *interfaces.Interface) (metadata *ifaceidx.LinuxIfMetadata, err error) {

	hostName := getHostIfName(linuxIf)
	agentPrefix := d.serviceLabel.GetAgentPrefix()

	if linuxIf.Namespace != nil {
		revert, err := d.nsPlugin.SwitchToNamespace(nsCtx, linuxIf.Namespace)
		if err != nil {
			d.log.Error(err)
			return nil, err
		}
		defer revert()
	}

	// in each network namespace there is exactly one loopback interface,
	// multiple configurations for a single namespaces is not valid
	alias, err := d.getLoopbackAlias(nsCtx, linuxIf)
	if err != nil {
		d.log.Error(err)
		return nil, ErrLoopbackNotFound
	}
	if alias != "" && alias != linuxIf.Name {
		d.log.Errorf("loopback already configured using logical name '%v'", alias)
		return nil, ErrLoopbackAlreadyConfigured
	}

	// add alias in order to include loopback in retrieve output
	err = d.ifHandler.SetInterfaceAlias(hostName, agentPrefix+linuxIf.Name)
	if err != nil {
		d.log.Error(err)
		return nil, err
	}

	link, err := d.ifHandler.GetLinkByName(hostName)
	if err != nil {
		d.log.Error(err)
		return nil, err
	}

	metadata = &ifaceidx.LinuxIfMetadata{
		Namespace:    linuxIf.Namespace,
		LinuxIfIndex: link.Attrs().Index,
	}

	return metadata, nil
}

// deleteLoopback clear associated interface alias
func (d *InterfaceDescriptor) deleteLoopback(nsCtx nslinuxcalls.NamespaceMgmtCtx, linuxIf *interfaces.Interface) error {
	hostName := getHostIfName(linuxIf)

	if linuxIf.Namespace != nil {
		revert, err := d.nsPlugin.SwitchToNamespace(nsCtx, linuxIf.Namespace)
		if err != nil {
			d.log.Error(err)
			return err
		}
		defer revert()
	}

	// remove interface alias
	// - actually vishvananda/netlink does not support alias removal, so we just change
	//   it to a string which is not prefixed with agent label
	err := d.ifHandler.SetInterfaceAlias(hostName, "")
	if err != nil {
		d.log.Error(err)
		return err
	}

	return nil
}

// getLoopbackAlias returns alias associated with the loopback
func (d *InterfaceDescriptor) getLoopbackAlias(nsCtx nslinuxcalls.NamespaceMgmtCtx, linuxIf *interfaces.Interface) (string, error) {
	if linuxIf.Namespace != nil {
		revert, err := d.nsPlugin.SwitchToNamespace(nsCtx, linuxIf.Namespace)
		if err != nil {
			d.log.Error(err)
			return "", err
		}
		defer revert()
	}

	link, err := d.ifHandler.GetLinkByName(getHostIfName(linuxIf))
	if err != nil {
		d.log.Error(err)
		return "", err
	}
	return strings.TrimPrefix(link.Attrs().Alias, d.serviceLabel.GetAgentPrefix()), nil
}
