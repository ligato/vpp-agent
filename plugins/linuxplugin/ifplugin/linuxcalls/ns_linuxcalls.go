// +build !windows,!darwin

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

package linuxcalls

import (
	"strconv"

	intf "github.com/ligato/vpp-agent/plugins/linuxplugin/common/model/interfaces"

	"fmt"
	"net"

	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/vishvananda/netlink"
)

// ToGenericNs converts interface-type namespace to generic type namespace. Such an object can be used to call common
// namespace-related methods
func ToGenericNs(ns *intf.LinuxInterfaces_Interface_Namespace) *Namespace {
	if ns == nil {
		return &Namespace{}
	}
	return &Namespace{Type: int32(ns.Type), Pid: ns.Pid, Microservice: ns.Microservice, Name: ns.Name, Filepath: ns.Filepath}
}

// ToInterfaceNs converts generic namespace to interface-type namespace
func ToInterfaceNs(ns *Namespace) (*intf.LinuxInterfaces_Interface_Namespace, error) {
	if ns == nil {
		return nil, fmt.Errorf("provided namespace is nil")
	}
	var namespaceType intf.LinuxInterfaces_Interface_Namespace_NamespaceType
	switch ns.Type {
	case 0:
		namespaceType = intf.LinuxInterfaces_Interface_Namespace_PID_REF_NS
	case 1:
		namespaceType = intf.LinuxInterfaces_Interface_Namespace_MICROSERVICE_REF_NS
	case 2:
		namespaceType = intf.LinuxInterfaces_Interface_Namespace_NAMED_NS
	case 3:
		namespaceType = intf.LinuxInterfaces_Interface_Namespace_FILE_REF_NS
	}
	return &intf.LinuxInterfaces_Interface_Namespace{Type: namespaceType, Pid: ns.Pid, Microservice: ns.Microservice, Name: ns.Name, Filepath: ns.Filepath}, nil
}

// NamespaceToStr returns a string representation of a namespace suitable for logging purposes.
func NamespaceToStr(namespace *intf.LinuxInterfaces_Interface_Namespace) string {
	if namespace != nil {
		switch namespace.Type {
		case intf.LinuxInterfaces_Interface_Namespace_PID_REF_NS:
			return "PID:" + strconv.Itoa(int(namespace.Pid))
		case intf.LinuxInterfaces_Interface_Namespace_MICROSERVICE_REF_NS:
			return "MICROSERVICE:" + namespace.Microservice
		case intf.LinuxInterfaces_Interface_Namespace_NAMED_NS:
			return namespace.Name
		case intf.LinuxInterfaces_Interface_Namespace_FILE_REF_NS:
			return "FILE:" + namespace.Filepath
		}
	}
	return "<nil>"
}

// SetInterfaceNamespace moves a given Linux interface into a specified namespace.
func SetInterfaceNamespace(ctx *NamespaceMgmtCtx, ifName string, namespace *intf.LinuxInterfaces_Interface_Namespace,
	log logging.Logger, stopwatch *measure.Stopwatch) error {
	ifaceNs := ToGenericNs(namespace)

	// Get network namespace file descriptor
	ns, err := ifaceNs.GetOrCreateNs(log)
	if err != nil {
		return err
	}
	defer ns.Close()

	// Get the link handler.
	link, err := netlink.LinkByName(ifName)
	if err != nil {
		return err
	}

	// When interface moves from one namespace to another, it loses all its IP addresses, admin status
	// and MTU configuration -- we need to remember the interface configuration before the move
	// and re-configure the interface in the new namespace.

	netIntf, err := net.InterfaceByName(ifName)
	if err != nil {
		return err
	}

	addrs, err := netIntf.Addrs()
	if err != nil {
		return err
	}

	// Move the interface into the namespace.
	err = netlink.LinkSetNsFd(link, int(ns))
	if err != nil {
		return err
	}
	log.WithFields(logging.Fields{"ifName": ifName, "dest-namespace": NamespaceToStr(namespace),
		"dest-namespace-fd": int(ns)}).
		Debug("Moved Linux interface across namespaces")

	// re-configure interface in its new namespace
	revertNs, err := ifaceNs.SwitchNamespace(ctx, log)
	if err != nil {
		return err
	}
	defer revertNs()

	if netIntf.Flags&net.FlagUp == 1 {
		// re-enable interface
		err = InterfaceAdminUp(ifName, measure.GetTimeLog("iface_admin_up", stopwatch))
		if nil != err {
			return fmt.Errorf("failed to enable Linux interface `%s`: %v", ifName, err)
		}
		log.WithFields(logging.Fields{"ifName": ifName}).
			Debug("Linux interface was re-enabled")
	}

	// re-add IP addresses
	for i := range addrs {
		ip, network, err := net.ParseCIDR(addrs[i].String())
		network.IP = ip /* combine IP address with netmask */
		if err != nil {
			return fmt.Errorf("failed to parse IPv4 address of a Linux interface `%s`: %v", ifName, err)
		}
		err = AddInterfaceIP(ifName, network, measure.GetTimeLog("add_iface_ip", stopwatch))
		if err != nil {
			if err.Error() == "file exists" {
				continue
			}
			return fmt.Errorf("failed to assign IPv4 address to a Linux interface `%s`: %v", ifName, err)
		}
		log.WithFields(logging.Fields{"ifName": ifName, "addr": network}).
			Debug("IP address was re-assigned to Linux interface")
	}

	// revert back the MTU config
	err = SetInterfaceMTU(ifName, netIntf.MTU, measure.GetTimeLog("set_iface_mtu", stopwatch))
	if nil != err {
		return fmt.Errorf("failed to set MTU of a Linux interface `%s`: %v", ifName, err)
	}
	log.WithFields(logging.Fields{"ifName": ifName, "mtu": netIntf.MTU}).
		Debug("MTU was reconfigured for Linux interface")

	return nil
}
