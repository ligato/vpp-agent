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

	"github.com/ligato/cn-infra/logging"
	"github.com/vishvananda/netns"
)

const (
	netnsMountDir = "/var/run/netns"
)

// Namespace types
const (
	PidRefNs          = 0
	MicroserviceRefNs = 1
	NamedNs           = 2
	FileRefNs         = 3
)

var defaultNs = netns.None()

func init() {
	// Save the network namespace used at the start of the application.
	defaultNs, _ = netns.Get()
}

// Namespace is a generic representation of typed namespace (interface, arp, etc...)
type Namespace struct {
	Type         int32
	Pid          uint32
	Microservice string
	Name         string
	Filepath     string
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
		return strings.Compare(ns.Filepath, nsToCompare.Filepath)
	}
	return 0
}

// NamespaceToStr returns a string representation of a namespace suitable for logging purposes.
func (ns *Namespace) NamespaceToStr() string {
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
		return "FILE:" + ns.Filepath
	default:
		return "unknown namespace type"
	}
}

// GetDefaultNamespace returns a generic default namespace
func GetDefaultNamespace() *Namespace {
	return &Namespace{Type: NamedNs, Name: ""}
}

// SwitchNamespace switches the network namespace of the current thread.
// Caller should eventually call the returned "revert" function in order to get back to the original
// network namespace (for example using "defer revert()").
func (ns *Namespace) SwitchNamespace(ctx *NamespaceMgmtCtx, log logging.Logger) (revert func(), err error) {
	var nsHandle netns.NsHandle

	// Save the current network namespace.
	origns, err := netns.Get()
	if err != nil {
		return func() {}, err
	}

	// Get network namespace file descriptor.
	nsHandle, err = ns.GetOrCreateNs(log)
	if err != nil {
		return func() {}, err
	}
	defer nsHandle.Close()

	alreadyLocked := ctx.lockedOsThread
	if !alreadyLocked {
		// Lock the OS Thread so we don't accidentally switch namespaces later.
		runtime.LockOSThread()
		ctx.lockedOsThread = true
		log.Debug("Locked OS thread")
	}

	// Switch the namespace.
	l := log.WithFields(logging.Fields{"ns": nsHandle.String(), "ns-fd": int(nsHandle)})
	if err := netns.Set(nsHandle); err != nil {
		l.Errorf("Failed to switch Linux network namespace (%v): %v", ns.NamespaceToStr(), err)
	} else {
		l.Debugf("Switched Linux network namespace (%v)", ns.NamespaceToStr())
	}

	return func() {
		l := log.WithFields(logging.Fields{"orig-ns": origns.String(), "orig-ns-fd": int(origns)})
		if err := netns.Set(origns); err != nil {
			l.Errorf("Failed to switch Linux network namespace: %v", err)
		} else {
			l.Debugf("Switched back to the original Linux network namespace")
		}
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
func (ns *Namespace) GetOrCreateNs(log logging.Logger) (netns.NsHandle, error) {
	var nsHandle netns.NsHandle
	var err error

	if ns == nil {
		return dupNsHandle(defaultNs)
	}

	switch ns.Type {
	case PidRefNs:
		if ns.Pid == 0 {
			// We consider scheduler's PID as the representation of the default namespace.
			return dupNsHandle(defaultNs)
		}
		nsHandle, err = netns.GetFromPid(int(ns.Pid))
		if err != nil {
			return netns.None(), err
		}
	case NamedNs:
		if ns.Name == "" {
			return dupNsHandle(defaultNs)
		}
		nsHandle, err = netns.GetFromName(ns.Name)
		if err != nil {
			// Create named namespace if it doesn't exist yet.
			_, _, err = CreateNamedNetNs(ns.Name, log)
			if err != nil {
				return netns.None(), err
			}
			nsHandle, err = netns.GetFromName(ns.Name)
			if err != nil {
				return netns.None(), errors.New("unable to get namespace by name")
			}
		}
	case FileRefNs:
		if ns.Filepath == "" {
			return dupNsHandle(defaultNs)
		}
		nsHandle, err = netns.GetFromPath(ns.Filepath)
		if err != nil {
			return netns.None(), err
		}
	case MicroserviceRefNs:
		return netns.None(), errors.New("unable to convert microservice label to PID at this level")
	}

	return nsHandle, nil
}

// CreateNamedNetNs creates a new named Linux network namespace.
// It does exactly the same thing as the command "ip netns add NAMESPACE".
func CreateNamedNetNs(namespace string, log logging.Logger) (netns.NsHandle, *Namespace, error) {
	log.WithFields(logging.Fields{"namespace": namespace}).
		Debug("Creating new named Linux namespace")
	// Prepare namespace proto object.
	nsObj := &Namespace{
		Type: NamedNs,
		Name: namespace,
	}

	// Lock the OS Thread so we don't accidentally switch namespaces.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// Save the current network namespace.
	origns, err := netns.Get()
	if err != nil {
		log.WithFields(logging.Fields{"namespace": nsObj.Name}).
			Error("Failed to get the original namespace")
		return netns.None(), nsObj, err
	}
	defer origns.Close()

	// Create directory for namespace mounts.
	err = os.MkdirAll(netnsMountDir, 0755)
	if err != nil {
		log.WithFields(logging.Fields{"namespace": nsObj.Name}).
			Error("Failed to create directory for namespace mounts")
		return netns.None(), nsObj, err
	}

	/* Make it possible for network namespace mounts to propagate between
	   mount namespaces.  This makes it likely that unmounting a network
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
			log.WithFields(logging.Fields{"namespace": nsObj.Name}).
				Error("mount --make-shared failed")
			return netns.None(), nsObj, err
		}
		/* Upgrade netnsMountDir to a mount point */
		err = syscall.Mount(netnsMountDir, netnsMountDir, "none", syscall.MS_BIND, "")
		if err != nil {
			log.WithFields(logging.Fields{"namespace": nsObj.Name}).
				Error("mount --bind failed")
			return netns.None(), nsObj, err
		}
		mountedNetnsDir = true
	}

	// Create file path for the mount.
	netnsMountFile := path.Join(netnsMountDir, nsObj.Name)
	file, err := os.OpenFile(netnsMountFile, os.O_RDONLY|os.O_CREATE|os.O_EXCL, 0444)
	if err != nil {
		log.WithFields(logging.Fields{"namespace": nsObj.Name}).
			Error("failed to create destination path for the namespace mount")
		return netns.None(), nsObj, err
	}
	file.Close()

	// Create and switch to a new namespace.
	newNsHandle, err := netns.New()
	if err != nil {
		log.WithFields(logging.Fields{"namespace": nsObj.Name}).
			Error("failed to create namespace")
		return netns.None(), nsObj, err
	}
	netns.Set(newNsHandle)

	// Create a bind-mount for the namespace.
	tid := syscall.Gettid()
	err = syscall.Mount("/proc/self/task/"+strconv.Itoa(tid)+"/ns/net", netnsMountFile, "none", syscall.MS_BIND, "")

	// Switch back to the original namespace.
	netns.Set(origns)

	if err != nil {
		newNsHandle.Close()
		log.WithFields(logging.Fields{"namespace": nsObj.Name}).
			Error("failed to create namespace bind-mount")
		return netns.None(), nsObj, err
	}

	return newNsHandle, nsObj, nil
}

// DeleteNamedNetNs deletes an existing named Linux network namespace.
// It does exactly the same thing as the command "ip netns del NAMESPACE".
func DeleteNamedNetNs(namespace string, log logging.Logger) error {
	log.WithFields(logging.Fields{"namespace": namespace}).
		Debug("Deleting named Linux namespace")

	// Unmount the namespace.
	netnsMountFile := path.Join(netnsMountDir, namespace)
	err := syscall.Unmount(netnsMountFile, syscall.MNT_DETACH)
	if err != nil {
		log.WithFields(logging.Fields{"namespace": namespace}).
			Error("failed to unmount namespace")
	}

	// Remove file path used for the mount.
	err = os.Remove(netnsMountFile)
	if err != nil {
		log.WithFields(logging.Fields{"namespace": namespace}).
			Error("failed to remove namespace file")
	}

	return err
}

// NamedNetNsExists checks whether namespace exists.
func NamedNetNsExists(namespace string, log logging.Logger) (bool, error) {
	netnsMountFile := path.Join(netnsMountDir, namespace)
	if _, err := os.Stat(netnsMountFile); err != nil {
		if os.IsNotExist(err) {
			log.WithFields(logging.Fields{"namespace": namespace}).
				Debug("namespace not found")
			return false, nil
		}
		log.WithFields(logging.Fields{"namespace": namespace}).
			Error("failed to read namespace")
		return false, err
	}
	log.WithFields(logging.Fields{"namespace": namespace}).
		Debug("namespace found")
	return true, nil
}
