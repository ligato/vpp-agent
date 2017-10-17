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
	"errors"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	intf "github.com/ligato/vpp-agent/plugins/linuxplugin/model/interfaces"

	"fmt"
	"net"

	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

const (
	netnsMountDir = "/var/run/netns"
)

// NamespaceMgmtCtx represents context of an ongoing management of Linux namespaces.
// The same context should not be used concurrently.
type NamespaceMgmtCtx struct {
	lockedOsThread bool
}

var defaultNs = netns.None()

func init() {
	// Save the network namespace used at the start of the application
	defaultNs, _ = netns.Get()
}

// NewNamespaceMgmtCtx creates and returns a new context for management of Linux namespaces.
func NewNamespaceMgmtCtx() *NamespaceMgmtCtx {
	return &NamespaceMgmtCtx{lockedOsThread: false}
}

// CompareNamespaces is a comparison function for "intf.Interfaces_Interface_Namespace" type.
func CompareNamespaces(ns1 *intf.LinuxInterfaces_Interface_Namespace, ns2 *intf.LinuxInterfaces_Interface_Namespace) int {
	if ns1 == nil || ns2 == nil {
		if ns1 == ns2 {
			return 0
		}
		return -1
	}
	if ns1.Type != ns2.Type {
		return int(ns1.Type) - int(ns2.Type)
	}
	switch ns1.Type {
	case intf.LinuxInterfaces_Interface_Namespace_PID_REF_NS:
		return int(ns1.Pid) - int(ns2.Pid)
	case intf.LinuxInterfaces_Interface_Namespace_MICROSERVICE_REF_NS:
		return strings.Compare(ns1.Microservice, ns2.Microservice)
	case intf.LinuxInterfaces_Interface_Namespace_NAMED_NS:
		return strings.Compare(ns1.Name, ns2.Name)
	case intf.LinuxInterfaces_Interface_Namespace_FILE_REF_NS:
		return strings.Compare(ns1.Filepath, ns2.Filepath)
	}
	return 0
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

// GetDefaultNamespace returns an instance of the proto message referencing default namespace.
func GetDefaultNamespace() *intf.LinuxInterfaces_Interface_Namespace {
	return &intf.LinuxInterfaces_Interface_Namespace{Type: intf.LinuxInterfaces_Interface_Namespace_NAMED_NS, Name: ""}
}

// SetInterfaceNamespace moves a given Linux interface into a specified namespace.
func SetInterfaceNamespace(ctx *NamespaceMgmtCtx, ifName string, namespace *intf.LinuxInterfaces_Interface_Namespace,
	log logging.Logger, stopwatch *measure.Stopwatch) error {
	// Get network namespace file descriptor
	ns, err := GetOrCreateNs(namespace, log)
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
		"dest-namespace-fd": int(ns)}).Debug("Moved Linux interface across namespaces")

	// re-configure interface in its new namespace
	revertNs, err := SwitchNamespace(ctx, namespace, log)
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
		log.WithFields(logging.Fields{"ifName": ifName}).Debug("Linux interface was re-enabled")
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
		log.WithFields(logging.Fields{"ifName": ifName, "addr": network}).Debug("IP address was re-assigned to Linux interface")
	}

	// revert back the MTU config
	err = SetInterfaceMTU(ifName, netIntf.MTU, measure.GetTimeLog("set_iface_mtu", stopwatch))
	if nil != err {
		return fmt.Errorf("failed to set MTU of a Linux interface `%s`: %v", ifName, err)
	}
	log.WithFields(logging.Fields{"ifName": ifName, "mtu": netIntf.MTU}).Debug("MTU was reconfigured for Linux interface")

	return nil
}

// SwitchNamespace switches the network namespace of the current thread.
// Caller should eventually call the returned "revert" function in order to get back to the original
// network namespace (for example using "defer revert()").
func SwitchNamespace(ctx *NamespaceMgmtCtx, namespace *intf.LinuxInterfaces_Interface_Namespace, log logging.Logger) (revert func(), err error) {
	var ns netns.NsHandle

	// Save the current network namespace
	origns, err := netns.Get()
	if err != nil {
		return func() {}, err
	}

	// Get network namespace file descriptor
	ns, err = GetOrCreateNs(namespace, log)
	if err != nil {
		return func() {}, err
	}
	defer ns.Close()

	alreadyLocked := ctx.lockedOsThread
	if !alreadyLocked {
		// Lock the OS Thread so we don't accidentally switch namespaces later.
		runtime.LockOSThread()
		ctx.lockedOsThread = true
		log.Debug("Locked OS thread")
	}

	// Switch the namespace.
	netns.Set(ns)
	log.WithFields(logging.Fields{"dest-namespace": NamespaceToStr(namespace), "dest-namespace-fd": int(ns)}).Debug(
		"Switched Linux network namespace")

	return func() {
		netns.Set(origns)
		log.WithFields(logging.Fields{"namespace-fd": int(origns)}).Debug(
			"Switched back to the original Linux network namespace")
		origns.Close()
		if !alreadyLocked {
			runtime.UnlockOSThread()
			ctx.lockedOsThread = false
			log.Debug("Unlocked OS thread")
		}
	}, nil
}

// dupNsHandle duplicates namespace handle.
func dupNsHandle(ns netns.NsHandle) (netns.NsHandle, error) {
	dup, err := syscall.Dup(int(ns))
	return netns.NsHandle(dup), err
}

// GetOrCreateNs returns an existing Linux network namespace or creates a new one if it doesn't exist yet.
// It is, however, only possible to create "named" namespaces. For PID-based namespaces, process with
// the given PID must exists, otherwise the function returns an error.
func GetOrCreateNs(namespace *intf.LinuxInterfaces_Interface_Namespace, log logging.Logger) (netns.NsHandle, error) {
	var ns netns.NsHandle
	var err error

	if namespace == nil {
		return dupNsHandle(defaultNs)
	}

	switch namespace.Type {
	case intf.LinuxInterfaces_Interface_Namespace_PID_REF_NS:
		if namespace.Pid == 0 {
			// We consider scheduler's PID as the representation of the default namespace
			return dupNsHandle(defaultNs)
		}
		ns, err = netns.GetFromPid(int(namespace.Pid))
		if err != nil {
			return netns.None(), err
		}
	case intf.LinuxInterfaces_Interface_Namespace_NAMED_NS:
		if namespace.Name == "" {
			return dupNsHandle(defaultNs)
		}
		ns, err = netns.GetFromName(namespace.Name)
		if err != nil {
			// Create named namespace if it doesn't exist yet.
			_, _, err = CreateNamedNetNs(namespace.Name, log)
			if err != nil {
				return netns.None(), err
			}
			ns, err = netns.GetFromName(namespace.Name)
			if err != nil {
				return netns.None(), errors.New("failed to get namespace by name")
			}
		}
	case intf.LinuxInterfaces_Interface_Namespace_FILE_REF_NS:
		if namespace.Filepath == "" {
			return dupNsHandle(defaultNs)
		}
		ns, err = netns.GetFromPath(namespace.Filepath)
		if err != nil {
			return netns.None(), err
		}
	case intf.LinuxInterfaces_Interface_Namespace_MICROSERVICE_REF_NS:
		return netns.None(), errors.New("don't know how to convert microservice label to PID at this level")
	}

	return ns, nil
}

// CreateNamedNetNs creates a new named Linux network namespace.
// It does exactly the same thing as the command "ip netns add NAMESPACE" .
func CreateNamedNetNs(namespace string, log logging.Logger) (netns.NsHandle, *intf.LinuxInterfaces_Interface_Namespace, error) {
	log.WithFields(logging.Fields{"namespace": namespace}).Debug("Creating new named Linux namespace")
	// Prepare namespace proto object
	nsObj := &intf.LinuxInterfaces_Interface_Namespace{
		Type: intf.LinuxInterfaces_Interface_Namespace_NAMED_NS,
		Name: namespace,
	}

	// Lock the OS Thread so we don't accidentally switch namespaces
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// Save the current network namespace
	origns, err := netns.Get()
	if err != nil {
		log.WithFields(logging.Fields{"namespace": nsObj.Name}).Error("Failed to get the original namespace")
		return netns.None(), nsObj, err
	}
	defer origns.Close()

	// Create directory for namespace mounts
	err = os.MkdirAll(netnsMountDir, 0755)
	if err != nil {
		log.WithFields(logging.Fields{"namespace": nsObj.Name}).Error("Failed to create directory for namespace mounts")
		return netns.None(), nsObj, err
	}

	/* Make it possible for network namespace mounts to propagate between
	   mount namespaces.  This makes it likely that a unmounting a network
	   namespace file in one namespace will unmount the network namespace
	   file in all namespaces allowing the network namespace to be freed
	   sooner.
	*/
	mountedNetnsDir := false
	for {
		err = syscall.Mount("", netnsMountDir, "none", syscall.MS_SHARED|syscall.MS_REC, "")
		if err == nil {
			break
		}
		if e, ok := err.(syscall.Errno); !ok || e != syscall.EINVAL || mountedNetnsDir {
			log.WithFields(logging.Fields{"namespace": nsObj.Name}).Error("mount --make-shared failed")
			return netns.None(), nsObj, err
		}
		/* Upgrade netnsMountDir to a mount point */
		err = syscall.Mount(netnsMountDir, netnsMountDir, "none", syscall.MS_BIND, "")
		if err != nil {
			log.WithFields(logging.Fields{"namespace": nsObj.Name}).Error("mount --bind failed")
			return netns.None(), nsObj, err
		}
		mountedNetnsDir = true
	}

	// Create file path for the mount
	netnsMountFile := path.Join(netnsMountDir, nsObj.Name)
	file, err := os.OpenFile(netnsMountFile, os.O_RDONLY|os.O_CREATE|os.O_EXCL, 0444)
	if err != nil {
		log.WithFields(logging.Fields{"namespace": nsObj.Name}).Error("failed to create destination path for the namespace mount")
		return netns.None(), nsObj, err
	}
	file.Close()

	// Create and switch to a new namespace
	newNsHandle, err := netns.New()
	if err != nil {
		log.WithFields(logging.Fields{"namespace": nsObj.Name}).Error("failed to create namespace")
		return netns.None(), nsObj, err
	}
	netns.Set(newNsHandle)

	// Create a bind-mount for the namespace
	tid := syscall.Gettid()
	err = syscall.Mount("/proc/self/task/"+strconv.Itoa(tid)+"/ns/net", netnsMountFile, "none", syscall.MS_BIND, "")

	// Switch back to the original namespace.
	netns.Set(origns)

	if err != nil {
		newNsHandle.Close()
		log.WithFields(logging.Fields{"namespace": nsObj.Name}).Error("failed to create namespace bind-mount")
		return netns.None(), nsObj, err
	}

	return newNsHandle, nsObj, nil
}

// DeleteNamedNetNs deletes an existing named Linux network namespace.
// It does exactly the same thing as the command "ip netns del NAMESPACE" .
func DeleteNamedNetNs(namespace string, log logging.Logger) error {
	log.WithFields(logging.Fields{"namespace": namespace}).Debug("Deleting named Linux namespace")

	// Unmount the namespace
	netnsMountFile := path.Join(netnsMountDir, namespace)
	err := syscall.Unmount(netnsMountFile, syscall.MNT_DETACH)
	if err != nil {
		log.WithFields(logging.Fields{"namespace": namespace}).Error("failed to unmount namespace")
	}

	// Remove file path used for the mount
	err = os.Remove(netnsMountFile)
	if err != nil {
		log.WithFields(logging.Fields{"namespace": namespace}).Error("failed to remove namespace file")
	}

	return err
}

// NamedNetNsExists checks whether namespace exists.
func NamedNetNsExists(namespace string, log logging.Logger) (bool, error) {
	netnsMountFile := path.Join(netnsMountDir, namespace)
	if _, err := os.Stat(netnsMountFile); err != nil {
		if os.IsNotExist(err) {
			log.WithFields(logging.Fields{"namespace": namespace}).Debug("namespace not found")
			return false, nil
		}
		log.WithFields(logging.Fields{"namespace": namespace}).Error("failed to read namespace")
		return false, err
	}
	log.WithFields(logging.Fields{"namespace": namespace}).Debug("namespace found")
	return true, nil
}
