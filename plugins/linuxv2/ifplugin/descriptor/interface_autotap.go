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
	"errors"
	"strings"

	"github.com/ligato/vpp-agent/plugins/linuxv2/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/linuxv2/model/interfaces"

	nslinuxcalls "github.com/ligato/vpp-agent/plugins/linuxv2/nsplugin/linuxcalls"
)

// addAutoTAP moves Linux-side of the VPP-TAP interface to the destination namespace
// and sets the requested host name.
func (intfd *InterfaceDescriptor) addAutoTAP(key string, linuxIf *interfaces.LinuxInterface) (metadata *ifaceidx.LinuxIfMetadata, err error) {
	// determine host/logical/temporary interface names
	tempHostName := getTapTempHostName(linuxIf)
	hostName := getHostIfName(linuxIf)

	// context
	nsCtx := nslinuxcalls.NewNamespaceMgmtCtx()
	agentPrefix := intfd.serviceLabel.GetAgentPrefix()

	// add alias to associate TAP with the logical name of the AUTO-TAP
	err = intfd.ifHandler.SetInterfaceAlias(tempHostName, agentPrefix+getTapAlias(linuxIf))
	if err != nil {
		intfd.log.Error(err)
		return nil, err
	}

	// move the TAP to the right namespace
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

	// rename from temporary host name to the request host name
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
		TapTempName:  tempHostName,
		Namespace:    linuxIf.Namespace,
		LinuxIfIndex: link.Attrs().Index,
	}

	return metadata, nil
}

// deleteAutoTAP returns TAP interface back to the default namespace and renames
// the interface back to original name.
func (intfd *InterfaceDescriptor) deleteAutoTAP(nsCtx nslinuxcalls.NamespaceMgmtCtx, key string, linuxIf *interfaces.LinuxInterface, metadata *ifaceidx.LinuxIfMetadata) error {
	hostName := getHostIfName(linuxIf)
	agentPrefix := intfd.serviceLabel.GetAgentPrefix()

	// get original TAP name
	link, err := intfd.ifHandler.GetLinkByName(hostName)
	if err != nil {
		intfd.log.Error(err)
		return err
	}
	alias := strings.TrimPrefix(link.Attrs().Alias, agentPrefix)
	_, tempHostName := parseTapAlias(alias)
	if tempHostName == "" {
		err = errors.New("failed to obtain the original TAP host name")
		intfd.log.Error(err)
		return err
	}

	// rename back to the temporary name
	intfd.ifHandler.RenameInterface(hostName, tempHostName)
	if err != nil {
		intfd.log.Error(err)
		return err
	}

	// move TAP back to the default namespace
	err = intfd.setInterfaceNamespace(nsCtx, tempHostName, nil)
	if err != nil {
		intfd.log.Error(err)
		return err
	}

	// move to the default namespace
	revert, err := intfd.nsPlugin.SwitchToNamespace(nsCtx, nil)
	if err != nil {
		intfd.log.Error(err)
		return err
	}
	defer revert()

	// remove interface alias at last(!)
	err = intfd.ifHandler.SetInterfaceAlias(tempHostName, "")
	if err != nil {
		intfd.log.Error(err)
		return err
	}

	return nil
}

// getTapAlias returns alias for Linux TAP interface managed by the agent.
// The alias stores the AUTO-TAP logical name together with the original TAP name.
func getTapAlias(linuxIf *interfaces.LinuxInterface) string {
	return linuxIf.Name + "/" + getTapTempHostName(linuxIf)
}

// parseTapAlias parses out AUTO-TAP logical name together with the original TAP
// name from the alias.
func parseTapAlias(alias string) (tapName, tapTmpName string) {
	aliasParts := strings.Split(alias, "/")
	tapName = aliasParts[0]
	if len(aliasParts) > 0 {
		tapTmpName = aliasParts[1]
	}
	return
}

// getTapTempHostName returns host name of the TAP interface to which the AUTO-TAP
// configuration should apply.
func getTapTempHostName(linuxIf *interfaces.LinuxInterface) string {
	tempIfName := linuxIf.GetAutoTap().GetTempIfName()
	if tempIfName == "" {
		return getHostIfName(linuxIf)
	}
	return tempIfName
}
