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
	"strings"

	"github.com/ligato/vpp-agent/plugins/linuxv2/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/linuxv2/model/interfaces"
	nslinuxcalls "github.com/ligato/vpp-agent/plugins/linuxv2/nsplugin/linuxcalls"
)

// addVETH creates a new VETH pair if neither of VETH-ends are configured, or just
// applies configuration to the unfinished VETH-end with a temporary host name.
func (intfd *InterfaceDescriptor) addVETH(key string, linuxIf *interfaces.LinuxInterface) (metadata *ifaceidx.LinuxIfMetadata, err error) {
	// determine host/logical/temporary interface names
	hostName := getHostIfName(linuxIf)
	peerName := getVethPeerName(linuxIf)
	tempHostName := getVethTemporaryHostName(linuxIf.Name)
	tempPeerHostName := getVethTemporaryHostName(peerName)

	// context
	nsCtx := nslinuxcalls.NewNamespaceMgmtCtx()
	ifIndex := intfd.scheduler.GetMetadataMap(InterfaceDescriptorName)
	agentPrefix := intfd.serviceLabel.GetAgentPrefix()

	// validate configuration
	if peerName == "" {
		err = ErrVETHWithoutPeer
		intfd.log.Error(err)
		return nil, err
	}

	// check if this VETH-end was already created by the other end
	_, peerExists := ifIndex.GetValue(peerName)
	if !peerExists {
		// delete obsolete/invalid unfinished VETH (ignore errors)
		intfd.ifHandler.DeleteInterface(tempHostName)
		intfd.ifHandler.DeleteInterface(tempPeerHostName)

		// create a new VETH pair
		err = intfd.ifHandler.AddVethInterfacePair(tempHostName, tempPeerHostName)
		if err != nil {
			intfd.log.Error(err)
			return nil, err
		}

		// add alias to both VETH ends
		err = intfd.ifHandler.SetInterfaceAlias(tempHostName, agentPrefix+getVethAlias(linuxIf.Name, peerName))
		if err != nil {
			intfd.log.Error(err)
			return nil, err
		}
		err = intfd.ifHandler.SetInterfaceAlias(tempPeerHostName, agentPrefix+getVethAlias(peerName, linuxIf.Name))
		if err != nil {
			intfd.log.Error(err)
			return nil, err
		}
	}

	// move the VETH-end to the right namespace
	err = intfd.setInterfaceNamespace(nsCtx, tempHostName, linuxIf.Namespace)
	if err != nil {
		intfd.log.Error(err)
		return nil, err
	}

	// move to the namespace with the interface
	revert, err := intfd.nsPlugin.SwitchToNamespace(nsCtx, linuxIf.Namespace)
	if err != nil {
		intfd.log.Error(err)
		return nil, err
	}
	defer revert()

	// rename from the temporary host name to the requested host name
	intfd.ifHandler.RenameInterface(tempHostName, hostName)
	if err != nil {
		intfd.log.Error(err)
		return nil, err
	}

	// build metadata
	link, err := intfd.ifHandler.GetLinkByName(hostName)
	if err != nil {
		intfd.log.Error(err)
		return nil, err
	}
	metadata = &ifaceidx.LinuxIfMetadata{
		Namespace:    linuxIf.Namespace,
		LinuxIfIndex: link.Attrs().Index,
	}

	return metadata, nil
}

// deleteVETH either un-configures one VETH-end if the other end is still configured, or
// removes the entire VETH pair.
func (intfd *InterfaceDescriptor) deleteVETH(nsCtx nslinuxcalls.NamespaceMgmtCtx, key string, linuxIf *interfaces.LinuxInterface, metadata *ifaceidx.LinuxIfMetadata) error {
	// determine host/logical/temporary interface names
	hostName := getHostIfName(linuxIf)
	peerName := getVethPeerName(linuxIf)
	tempHostName := getVethTemporaryHostName(linuxIf.Name)
	tempPeerHostName := getVethTemporaryHostName(peerName)

	// check if the other end is still configured
	ifIndex := intfd.scheduler.GetMetadataMap(InterfaceDescriptorName)
	_, peerExists := ifIndex.GetValue(peerName)
	if peerExists {
		// just un-configure this VETH-end, but do not delete the pair

		// rename to the temporary host name
		err := intfd.ifHandler.RenameInterface(hostName, tempHostName)
		if err != nil {
			intfd.log.Error(err)
			return err
		}

		// move this VETH-end to the default namespace
		err = intfd.setInterfaceNamespace(nsCtx, tempHostName, nil)
		if err != nil {
			intfd.log.Error(err)
			return err
		}
	} else {
		// remove the VETH pair completely now
		err := intfd.ifHandler.DeleteInterface(hostName)
		if err != nil {
			intfd.log.Error(err)
			return err
		}
		// peer should be automatically removed as well, but just in case...
		intfd.ifHandler.DeleteInterface(tempPeerHostName) // ignore errors
	}

	return nil
}

// getVethAlias returns alias for Linux VETH interface managed by the agent.
// The alias stores the VETH logical name together with the peer (logical) name.
func getVethAlias(vethName, peerName string) string {
	return vethName + "/" + peerName
}

// parseVethAlias parses out VETH logical name together with the peer name from the alias.
func parseVethAlias(alias string) (vethName, peerName string) {
	aliasParts := strings.Split(alias, "/")
	vethName = aliasParts[0]
	if len(aliasParts) > 0 {
		peerName = aliasParts[1]
	}
	return
}

// getVethPeerName returns the name of the peer interface from the configuration.
func getVethPeerName(linuxIf *interfaces.LinuxInterface) string {
	ref, ok := linuxIf.Link.(*interfaces.LinuxInterface_VethPeerIfName)
	if ok {
		return ref.VethPeerIfName
	}
	return ""
}

// getVethTemporaryHostName (deterministically) generates a temporary host name
// for a VETH interface.
func getVethTemporaryHostName(vethName string) string {
	return fmt.Sprintf("veth-%d", fnvHash(vethName))
}

// fnvHash hashes string using fnv32a algorithm.
func fnvHash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}
