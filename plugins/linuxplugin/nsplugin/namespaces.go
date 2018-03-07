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

package nsplugin

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"github.com/ligato/cn-infra/logging"
	intf "github.com/ligato/vpp-agent/plugins/linuxplugin/common/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/linuxplugin/common/model/l3"
	"github.com/vishvananda/netns"
)

// Namespace-related constants
const (
	// Configuration namespace for veth interfaces
	configNamespace = "veth-cfg-ns"
	// Namespace mount directory
	netNsMountDir = "/var/run/netns"
	// Namespace types
	PidRefNs          = 0
	MicroserviceRefNs = 1
	NamedNs           = 2
	FileRefNs         = 3
)

// Namespace is a generic representation of typed namespace (interface, arp, etc...)
type Namespace struct {
	Type         int32
	Pid          uint32
	Microservice string
	Name         string
	FilePath     string
}

// NamespaceMgmtCtx represents context of an ongoing management of Linux namespaces.
// The same context should not be used concurrently.
type NamespaceMgmtCtx struct {
	lockedOsThread bool
}

// NewNamespaceMgmtCtx creates and returns a new context for management of Linux namespaces.
func NewNamespaceMgmtCtx() *NamespaceMgmtCtx {
	return &NamespaceMgmtCtx{lockedOsThread: false}
}

// CompareNamespaces is a comparison function for "Namespace" type.
func (ns *Namespace) CompareNamespaces(nsToCompare *Namespace) int {
	if ns == nil || nsToCompare == nil {
		if ns == nsToCompare {
			return 0
		}
		return -1
	}
	if ns.Type != nsToCompare.Type {
		return int(ns.Type) - int(nsToCompare.Type)
	}
	switch ns.Type {
	case PidRefNs:
		return int(ns.Pid) - int(ns.Pid)
	case MicroserviceRefNs:
		return strings.Compare(ns.Microservice, nsToCompare.Microservice)
	case NamedNs:
		return strings.Compare(ns.Name, nsToCompare.Name)
	case FileRefNs:
		return strings.Compare(ns.FilePath, nsToCompare.FilePath)
	}
	return 0
}

// GenericToIfaceNs converts generic namespace to interface-type namespace
func (ns *Namespace) GenericToIfaceNs() (*intf.LinuxInterfaces_Interface_Namespace, error) {
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
	return &intf.LinuxInterfaces_Interface_Namespace{Type: namespaceType, Pid: ns.Pid, Microservice: ns.Microservice, Name: ns.Name, Filepath: ns.FilePath}, nil
}

// GenericToRouteNs converts generic namespace to arp-type namespace
func (plugin *NsHandler) GenericToRouteNs(ns *Namespace) (*l3.LinuxStaticRoutes_Route_Namespace, error) {
	if ns == nil {
		return nil, fmt.Errorf("provided namespace is nil")
	}
	var namespaceType l3.LinuxStaticRoutes_Route_Namespace_NamespaceType
	switch ns.Type {
	case 0:
		namespaceType = l3.LinuxStaticRoutes_Route_Namespace_PID_REF_NS
	case 1:
		namespaceType = l3.LinuxStaticRoutes_Route_Namespace_MICROSERVICE_REF_NS
	case 2:
		namespaceType = l3.LinuxStaticRoutes_Route_Namespace_NAMED_NS
	case 3:
		namespaceType = l3.LinuxStaticRoutes_Route_Namespace_FILE_REF_NS
	}
	return &l3.LinuxStaticRoutes_Route_Namespace{Type: namespaceType, Pid: ns.Pid, Microservice: ns.Microservice, Name: ns.Name, Filepath: ns.FilePath}, nil
}

// GenericToArpNs converts generic namespace to arp-type namespace
func (plugin *NsHandler) GenericToArpNs(ns *Namespace) (*l3.LinuxStaticArpEntries_ArpEntry_Namespace, error) {
	if ns == nil {
		return nil, fmt.Errorf("provided namespace is nil")
	}
	var namespaceType l3.LinuxStaticArpEntries_ArpEntry_Namespace_NamespaceType
	switch ns.Type {
	case 0:
		namespaceType = l3.LinuxStaticArpEntries_ArpEntry_Namespace_PID_REF_NS
	case 1:
		namespaceType = l3.LinuxStaticArpEntries_ArpEntry_Namespace_MICROSERVICE_REF_NS
	case 2:
		namespaceType = l3.LinuxStaticArpEntries_ArpEntry_Namespace_NAMED_NS
	case 3:
		namespaceType = l3.LinuxStaticArpEntries_ArpEntry_Namespace_FILE_REF_NS
	}
	return &l3.LinuxStaticArpEntries_ArpEntry_Namespace{Type: namespaceType, Pid: ns.Pid, Microservice: ns.Microservice, Name: ns.Name, Filepath: ns.FilePath}, nil
}

// GenericNsToString returns a string representation of a namespace suitable for logging purposes.
func (ns *Namespace) GenericNsToString() string {
	if ns == nil {
		return "invalid namespace"
	}
	switch ns.Type {
	case PidRefNs:
		return "PID:" + strconv.Itoa(int(ns.Pid))
	case MicroserviceRefNs:
		return "MICROSERVICE:" + ns.Microservice
	case NamedNs:
		return ns.Name
	case FileRefNs:
		return "FILE:" + ns.FilePath
	default:
		return "unknown namespace type"
	}
}

// IfaceNsToString returns a string representation of a namespace suitable for logging purposes.
func (plugin *NsHandler) IfaceNsToString(namespace *intf.LinuxInterfaces_Interface_Namespace) string {
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

// createNamedNetNs creates a new named Linux network namespace.
// It does exactly the same thing as the command "ip netns add NAMESPACE".
func (ns *Namespace) createNamedNetNs(log logging.Logger) (netns.NsHandle, error) {
	log.WithFields(logging.Fields{"namespace": ns.Name}).
		Debug("Creating new named Linux namespace")

	// Lock the OS Thread so we don't accidentally switch namespaces.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// Save the current network namespace.
	origns, err := netns.Get()
	if err != nil {
		log.WithFields(logging.Fields{"namespace": ns.Name}).
			Error("Failed to get the original namespace")
		return netns.None(), err
	}
	defer origns.Close()

	// Create directory for namespace mounts.
	err = os.MkdirAll(netNsMountDir, 0755)
	if err != nil {
		log.WithFields(logging.Fields{"namespace": ns.Name}).
			Error("Failed to create directory for namespace mounts")
		return netns.None(), err
	}

	/* Make it possible for network namespace mounts to propagate between
	   mount namespaces.  This makes it likely that unmounting a network
	   namespace file in one namespace will unmount the network namespace
	   file in all namespaces allowing the network namespace to be freed
	   sooner.
	*/
	mountedNetnsDir := false
	for {
		err = syscall.Mount("", netNsMountDir, "none", syscall.MS_SHARED|syscall.MS_REC, "")
		if err == nil {
			break
		}
		if e, ok := err.(syscall.Errno); !ok || e != syscall.EINVAL || mountedNetnsDir {
			log.WithFields(logging.Fields{"namespace": ns.Name}).
				Error("mount --make-shared failed")
			return netns.None(), err
		}
		/* Upgrade netNsMountDir to a mount point */
		err = syscall.Mount(netNsMountDir, netNsMountDir, "none", syscall.MS_BIND, "")
		if err != nil {
			log.WithFields(logging.Fields{"namespace": ns.Name}).
				Error("mount --bind failed")
			return netns.None(), err
		}
		mountedNetnsDir = true
	}

	// Create file path for the mount.
	netnsMountFile := path.Join(netNsMountDir, ns.Name)
	file, err := os.OpenFile(netnsMountFile, os.O_RDONLY|os.O_CREATE|os.O_EXCL, 0444)
	if err != nil {
		log.WithFields(logging.Fields{"namespace": ns.Name}).
			Error("failed to create destination path for the namespace mount")
		return netns.None(), err
	}
	file.Close()

	// Create and switch to a new namespace.
	newNsHandle, err := netns.New()
	if err != nil {
		log.WithFields(logging.Fields{"namespace": ns.Name}).
			Error("failed to create namespace")
		return netns.None(), err
	}
	netns.Set(newNsHandle)

	// Create a bind-mount for the namespace.
	tid := syscall.Gettid()
	err = syscall.Mount("/proc/self/task/"+strconv.Itoa(tid)+"/ns/net", netnsMountFile, "none", syscall.MS_BIND, "")

	// Switch back to the original namespace.
	netns.Set(origns)

	if err != nil {
		newNsHandle.Close()
		log.WithFields(logging.Fields{"namespace": ns.Name}).
			Error("failed to create namespace bind-mount")
		return netns.None(), err
	}

	return newNsHandle, nil
}

// deleteNamedNetNs deletes an existing named Linux network namespace.
// It does exactly the same thing as the command "ip netns del NAMESPACE".
func (ns *Namespace) deleteNamedNetNs(log logging.Logger) error {
	log.WithFields(logging.Fields{"namespace": ns.Name}).
		Debug("Deleting named Linux namespace")

	// Unmount the namespace.
	netnsMountFile := path.Join(netNsMountDir, ns.Name)
	err := syscall.Unmount(netnsMountFile, syscall.MNT_DETACH)
	if err != nil {
		log.WithFields(logging.Fields{"namespace": ns.Name}).
			Error("failed to unmount namespace")
	}

	// Remove file path used for the mount.
	err = os.Remove(netnsMountFile)
	if err != nil {
		log.WithFields(logging.Fields{"namespace": ns.Name}).
			Error("failed to remove namespace file")
	}

	return err
}

// namedNetNsExists checks whether namespace exists.
func (ns *Namespace) namedNetNsExists(log logging.Logger) (bool, error) {
	netnsMountFile := path.Join(netNsMountDir, ns.Name)
	if _, err := os.Stat(netnsMountFile); err != nil {
		if os.IsNotExist(err) {
			log.WithFields(logging.Fields{"namespace": ns.Name}).
				Debug("namespace not found")
			return false, nil
		}
		log.WithFields(logging.Fields{"namespace": ns.Name}).
			Error("failed to read namespace")
		return false, err
	}
	log.WithFields(logging.Fields{"namespace": ns.Name}).
		Debug("namespace found")
	return true, nil
}
