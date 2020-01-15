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

package descriptor

import (
	"fmt"
	"hash/fnv"

	"github.com/pkg/errors"

	"go.ligato.io/vpp-agent/v3/plugins/linux/ifplugin/linuxcalls"

	"go.ligato.io/vpp-agent/v3/plugins/linux/ifplugin/ifaceidx"
	nslinuxcalls "go.ligato.io/vpp-agent/v3/plugins/linux/nsplugin/linuxcalls"
	interfaces "go.ligato.io/vpp-agent/v3/proto/ligato/linux/interfaces"
)

// createVETH creates a new VETH pair if neither of VETH-ends are configured, or just
// applies configuration to the unfinished VETH-end with a temporary host name.
func (d *InterfaceDescriptor) createVETH(
	nsCtx nslinuxcalls.NamespaceMgmtCtx, key string, linuxIf *interfaces.Interface,
) (md *ifaceidx.LinuxIfMetadata, err error) {
	// determine host/logical/temporary interface names
	hostName := getHostIfName(linuxIf)
	peerName := linuxIf.GetVeth().GetPeerIfName()
	tempHostName := getVethTemporaryHostName(linuxIf.GetName())
	tempPeerHostName := getVethTemporaryHostName(peerName)

	// context
	agentPrefix := d.serviceLabel.GetAgentPrefix()

	// check if this VETH-end was already created by the other end
	_, peerExists := d.intfIndex.LookupByName(peerName)
	if !peerExists {
		// delete obsolete/invalid unfinished VETH (ignore errors)
		d.ifHandler.DeleteInterface(tempHostName)
		d.ifHandler.DeleteInterface(tempPeerHostName)

		// create a new VETH pair
		err = d.ifHandler.AddVethInterfacePair(tempHostName, tempPeerHostName)
		if err != nil {
			return nil, errors.WithMessagef(err, "error adding veth interface pair %s, %s", tempHostName, tempPeerHostName)
		}

		// add alias to both VETH ends
		err = d.ifHandler.SetInterfaceAlias(tempHostName, agentPrefix+linuxcalls.GetVethAlias(linuxIf.Name, peerName))
		if err != nil {
			return nil, errors.WithMessagef(err, "error setting interface %s alias", tempHostName)
		}
		err = d.ifHandler.SetInterfaceAlias(tempPeerHostName, agentPrefix+linuxcalls.GetVethAlias(peerName, linuxIf.Name))
		if err != nil {
			return nil, errors.WithMessagef(err, "error setting peer interface %s alias", tempPeerHostName)
		}
	}

	// move the VETH-end to the right namespace
	err = d.setInterfaceNamespace(nsCtx, tempHostName, linuxIf.Namespace)
	if err != nil {
		return nil, errors.WithMessagef(err, "error setting interface %s to namespace %v", tempHostName, linuxIf.Namespace)
	}

	// move to the namespace with the interface
	revert, err := d.nsPlugin.SwitchToNamespace(nsCtx, linuxIf.Namespace)
	if err != nil {
		return nil, errors.WithMessagef(err, "error switching to namespace %v", linuxIf.Namespace)
	}
	defer revert()

	// rename from the temporary host name to the requested host name
	if err = d.ifHandler.RenameInterface(tempHostName, hostName); err != nil {
		return nil, errors.WithMessagef(err, "error renaming %s to %s", tempHostName, hostName)
	}

	// build metadata
	link, err := d.ifHandler.GetLinkByName(hostName)
	if err != nil {
		return nil, errors.WithMessagef(err, "error getting link %s", hostName)
	}

	return &ifaceidx.LinuxIfMetadata{
		Namespace:    linuxIf.Namespace,
		LinuxIfIndex: link.Attrs().Index,
	}, nil
}

// deleteVETH either un-configures one VETH-end if the other end is still configured, or
// removes the entire VETH pair.
func (d *InterfaceDescriptor) deleteVETH(nsCtx nslinuxcalls.NamespaceMgmtCtx, key string, linuxIf *interfaces.Interface, metadata *ifaceidx.LinuxIfMetadata) error {
	// determine host/logical/temporary interface names
	hostName := getHostIfName(linuxIf)
	peerName := linuxIf.GetVeth().GetPeerIfName()
	tempHostName := getVethTemporaryHostName(linuxIf.Name)
	tempPeerHostName := getVethTemporaryHostName(peerName)

	// check if the other end is still configured
	_, peerExists := d.intfIndex.LookupByName(peerName)
	if peerExists {
		// just un-configure this VETH-end, but do not delete the pair

		// rename to the temporary host name
		err := d.ifHandler.RenameInterface(hostName, tempHostName)
		if err != nil {
			d.log.Error(err)
			return err
		}

		// move this VETH-end to the default namespace
		err = d.setInterfaceNamespace(nsCtx, tempHostName, nil)
		if err != nil {
			d.log.Error(err)
			return err
		}
	} else {
		// remove the VETH pair completely now
		err := d.ifHandler.DeleteInterface(hostName)
		if err != nil {
			d.log.Error(err)
			return err
		}
		if tempPeerHostName != "" {
			// peer should be automatically removed as well, but just in case...
			_ = d.ifHandler.DeleteInterface(tempPeerHostName) // ignore errors
		}
	}

	return nil
}

// getVethTemporaryHostName (deterministically) generates a temporary host name
// for a VETH interface.
func getVethTemporaryHostName(vethName string) string {
	if vethName == "" {
		return ""
	}
	return fmt.Sprintf("veth-%d", fnvHash(vethName))
}

// fnvHash hashes string using fnv32a algorithm.
func fnvHash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}
