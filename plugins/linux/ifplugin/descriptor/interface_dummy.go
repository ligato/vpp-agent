// Copyright (c) 2020 Pantheon.tech
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
	"github.com/pkg/errors"
	"go.ligato.io/vpp-agent/v3/plugins/linux/ifplugin/linuxcalls"

	"go.ligato.io/vpp-agent/v3/plugins/linux/ifplugin/ifaceidx"
	nslinuxcalls "go.ligato.io/vpp-agent/v3/plugins/linux/nsplugin/linuxcalls"
	interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/linux/interfaces"
)

// createDummyIf creates dummy interface.
func (d *InterfaceDescriptor) createDummyIf(
	nsCtx nslinuxcalls.NamespaceMgmtCtx, linuxIf *interfaces.Interface,
) (md *ifaceidx.LinuxIfMetadata, err error) {
	hostName := getHostIfName(linuxIf)
	agentPrefix := d.serviceLabel.GetAgentPrefix()

	// move to the namespace with the interface
	revert, err := d.nsPlugin.SwitchToNamespace(nsCtx, linuxIf.Namespace)
	if err != nil {
		d.log.Error("switch to namespace failed:", err)
		return nil, err
	}
	defer revert()

	// create a new Dummy interface
	err = d.ifHandler.AddDummyInterface(hostName)
	if err != nil {
		return nil, errors.WithMessagef(err,
			"failed to create dummy interface %s", hostName)
	}

	// add alias
	err = d.ifHandler.SetInterfaceAlias(hostName, agentPrefix+linuxcalls.GetDummyIfAlias(linuxIf))
	if err != nil {
		return nil, errors.WithMessagef(err,
			"error setting alias for Dummy interface %s", hostName)
	}

	// build metadata
	link, err := d.ifHandler.GetLinkByName(hostName)
	if err != nil {
		return nil, errors.WithMessagef(err, "error getting link %s", hostName)
	}

	return &ifaceidx.LinuxIfMetadata{
		Namespace:    linuxIf.Namespace,
		LinuxIfIndex: link.Attrs().Index,
		HostIfName:   hostName,
	}, nil
}

// deleteDummyIf removes dummy interface.
func (d *InterfaceDescriptor) deleteDummyIf(linuxIf *interfaces.Interface) error {
	hostName := getHostIfName(linuxIf)
	err := d.ifHandler.DeleteInterface(hostName)
	if err != nil {
		d.log.Error(err)
		return err
	}
	return nil
}