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
	"fmt"
	"github.com/pkg/errors"
	"go.ligato.io/vpp-agent/v3/plugins/linux/ifplugin/linuxcalls"

	"go.ligato.io/vpp-agent/v3/plugins/linux/ifplugin/ifaceidx"
	nslinuxcalls "go.ligato.io/vpp-agent/v3/plugins/linux/nsplugin/linuxcalls"
	interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/linux/interfaces"
)

const (
	// Enabling this option allows a “global” listen socket to work across L3 master domains (e.g., VRFs).
	sysctlL3MDevVar = "net.ipv4.tcp_l3mdev_accept"
)

// createVRF creates a new VRF network device.
func (d *InterfaceDescriptor) createVRF(
	nsCtx nslinuxcalls.NamespaceMgmtCtx, linuxIf *interfaces.Interface,
) (md *ifaceidx.LinuxIfMetadata, err error) {
	hostName := getHostIfName(linuxIf)
	rt := linuxIf.GetVrfDev().GetRoutingTable()
	agentPrefix := d.serviceLabel.GetAgentPrefix()

	// move to the namespace with the interface
	revert, err := d.nsPlugin.SwitchToNamespace(nsCtx, linuxIf.Namespace)
	if err != nil {
		d.log.Error("switch to namespace failed:", err)
		return nil, err
	}
	defer revert()

	// Enable child sockets to inherit the L3 master device index.
	// Without this VRF cannot really be used.
	err = d.enableL3MasterDev()
	if err != nil {
		return nil, err
	}

	// create a new VRF device
	err = d.ifHandler.AddVRFDevice(hostName, rt)
	if err != nil {
		return nil, errors.WithMessagef(err, "failed to add VRF device %s (rt: %d)", hostName, rt)
	}

	// add alias
	err = d.ifHandler.SetInterfaceAlias(hostName, agentPrefix+linuxcalls.GetVRFAlias(linuxIf))
	if err != nil {
		return nil, errors.WithMessagef(err, "error setting VRF %s alias", hostName)
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
		VrfDevRT:     rt,
	}, nil
}

// deleteVRF removes VRF network device.
func (d *InterfaceDescriptor) deleteVRF(linuxIf *interfaces.Interface) error {
	hostName := getHostIfName(linuxIf)
	err := d.ifHandler.DeleteInterface(hostName)
	if err != nil {
		d.log.Error(err)
		return err
	}
	return nil
}

func (d *InterfaceDescriptor) enableL3MasterDev() error {
	value, err := getSysctl(sysctlL3MDevVar)
	if err != nil {
		err = fmt.Errorf("could not read sysctl value for %s: %w",
			sysctlL3MDevVar, err)
		return err
	}
	if value == "0" {
		_, err = setSysctl(sysctlL3MDevVar, "1")
		if err != nil {
			err = fmt.Errorf("failed to enable %s: %w", sysctlL3MDevVar, err)
			return err
		}
	}
	return nil
}