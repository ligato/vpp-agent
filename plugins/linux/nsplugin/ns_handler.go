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
	"context"
	"net"
	"runtime"
	"sync"
	"syscall"

	"github.com/fsouza/go-dockerclient"
	"github.com/go-errors/errors"
	"github.com/ligato/cn-infra/logging"
	ipAddrs "github.com/ligato/cn-infra/utils/addrs"
	"github.com/ligato/vpp-agent/plugins/linux/ifplugin/linuxcalls"
	intf "github.com/ligato/vpp-agent/plugins/linux/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/linux/model/l3"
	"github.com/vishvananda/netns"
)

// NsHandler is a plugin to handle namespaces and microservices for other linux plugins (ifplugin, l3plugin ...).
// It does not follow the standard concept of CRUD, but provides a set of methods other plugins can use to manage
// namespaces
type NsHandler struct {
	log logging.Logger

	cfgLock sync.Mutex

	// Default namespace
	defaultNs netns.NsHandle

	// docker client - used to convert microservice label into the PID and ID of the container
	dockerClient *docker.Client
	// Microservice label -> Microservice info
	microServiceByLabel map[string]*Microservice //todo
	// Microservice container ID -> Microservice info
	microServiceByID map[string]*Microservice //todo
	// channel to send microservice updates
	microserviceChan chan *MicroserviceCtx

	ifMicroserviceNotif chan *MicroserviceEvent

	// config namespace, serves as a temporary namespace for VETH type interfaces where they are created and then
	// moved to proper namespace
	configNs *intf.LinuxInterfaces_Interface_Namespace

	// Handlers
	ifHandler  linuxcalls.NetlinkAPI
	sysHandler SystemAPI

	// Context within which all goroutines are running
	ctx context.Context
	// Cancel can be used to cancel all goroutines and their jobs inside of the plugin.
	cancel context.CancelFunc
	// Wait group allows to wait until all goroutines of the plugin have finished.
	wg sync.WaitGroup
}

// Init namespace handler caches and create config namespace
func (h *NsHandler) Init(logger logging.PluginLogger, ifHandler linuxcalls.NetlinkAPI, sysHandler SystemAPI,
	msChan chan *MicroserviceCtx, ifNotif chan *MicroserviceEvent) error {
	// Logger
	h.log = logger.NewLogger("-ns-handler")

	// Init channels
	h.microserviceChan = msChan
	h.ifMicroserviceNotif = ifNotif

	h.ctx, h.cancel = context.WithCancel(context.Background())

	h.microServiceByLabel = make(map[string]*Microservice)
	h.microServiceByID = make(map[string]*Microservice)

	// Handlers
	h.ifHandler = ifHandler
	h.sysHandler = sysHandler

	// Default namespace
	var err error
	h.defaultNs, err = netns.Get()
	if err != nil {
		return errors.Errorf("failed to init default namespace: %v", err)
	}

	// Docker client
	h.dockerClient, err = docker.NewClientFromEnv()
	if err != nil {
		return errors.Errorf("failed to get docker client instance from the environment variables: %v", err)
	}
	h.log.Debugf("Using docker client endpoint: %+v", h.dockerClient.Endpoint())

	// Create config namespace (for VETHs)
	if err = h.prepareConfigNamespace(); err != nil {
		return errors.Errorf("failed to prepare config namespace: %v", err)
	}

	// Start microservice tracker
	go h.trackMicroservices(h.ctx)

	h.log.Infof("Namespace handler plugin initialized")

	return nil
}

// Close pre-configured namespace
func (h *NsHandler) Close() error {
	if h.configNs != nil {
		// Remove veth pre-configure namespace
		ns := h.IfNsToGeneric(h.configNs)
		if err := ns.deleteNamedNetNs(h.sysHandler, h.log); err != nil {
			return errors.Errorf("failed to delete named namspace: %v", err)
		}
		h.cancel()
		h.wg.Wait()
	}

	return nil
}

// GetConfigNamespace return configuration namespace object
func (h *NsHandler) GetConfigNamespace() *intf.LinuxInterfaces_Interface_Namespace {
	return h.configNs
}

// GetMicroserviceByLabel returns internal microservice-by-label mapping
func (h *NsHandler) GetMicroserviceByLabel() map[string]*Microservice {
	return h.microServiceByLabel
}

// GetMicroserviceByID returns internal microservice-by-id mapping
func (h *NsHandler) GetMicroserviceByID() map[string]*Microservice {
	return h.microServiceByID
}

// SetInterfaceNamespace moves a given Linux interface into a specified namespace.
func (h *NsHandler) SetInterfaceNamespace(ctx *NamespaceMgmtCtx, ifName string, namespace *intf.LinuxInterfaces_Interface_Namespace) error {
	// Convert microservice namespace
	var err error
	if namespace != nil && namespace.Type == intf.LinuxInterfaces_Interface_Namespace_MICROSERVICE_REF_NS {
		// Convert namespace
		ifNs := h.convertMicroserviceNsToPidNs(namespace.Microservice)
		// Back to interface ns type
		namespace, err = ifNs.GenericToIfaceNs()
		if err != nil {
			return errors.Errorf("failed to convert generic interface namespace: %v", err)
		}
		if namespace == nil {
			return &unavailableMicroserviceErr{}
		}
	}

	ifaceNs := h.IfNsToGeneric(namespace)

	// Get network namespace file descriptor
	ns, err := h.getOrCreateNs(ifaceNs)
	if err != nil {
		return errors.Errorf("faield to get or create namespace %s: %v", namespace.Name, err)
	}
	defer ns.Close()

	// Get the link plugin.
	link, err := h.ifHandler.GetLinkByName(ifName)
	if err != nil {
		return errors.Errorf("failed to get link for interface %s: %v", ifName, err)
	}

	// When interface moves from one namespace to another, it loses all its IP addresses, admin status
	// and MTU configuration -- we need to remember the interface configuration before the move
	// and re-configure the interface in the new namespace.
	addresses, isIPv6, err := h.getLinuxIfAddrs(link.Attrs().Name)
	if err != nil {
		return errors.Errorf("failed to get IP address list from interface %s: %v", link.Attrs().Name, err)
	}

	// Move the interface into the namespace.
	err = h.sysHandler.LinkSetNsFd(link, int(ns))
	if err != nil {
		return errors.Errorf("failed to set interface %s file descriptor: %v", link.Attrs().Name, err)
	}

	// Re-configure interface in its new namespace
	revertNs, err := h.SwitchNamespace(ifaceNs, ctx)
	if err != nil {
		return errors.Errorf("failed to switch namespace: %v", err)
	}
	defer revertNs()

	if link.Attrs().Flags&net.FlagUp == 1 {
		// Re-enable interface
		err = h.ifHandler.SetInterfaceUp(ifName)
		if nil != err {
			return errors.Errorf("failed to re-enable Linux interface `%s`: %v", ifName, err)
		}
	}

	// Re-add IP addresses
	for _, address := range addresses {
		// Skip IPv6 link local address if there is no other IPv6 address
		if !isIPv6 && address.IP.IsLinkLocalUnicast() {
			continue
		}
		err = h.ifHandler.AddInterfaceIP(ifName, address)
		if err != nil {
			if err.Error() == "file exists" {
				continue
			}
			return errors.Errorf("failed to re-assign IP address to a Linux interface `%s`: %v", ifName, err)
		}
	}

	// Revert back the MTU config
	err = h.ifHandler.SetInterfaceMTU(ifName, link.Attrs().MTU)
	if nil != err {
		return errors.Errorf("failed to re-assign MTU of a Linux interface `%s`: %v", ifName, err)
	}

	return nil
}

// SwitchToNamespace switches the network namespace of the current thread.
func (h *NsHandler) SwitchToNamespace(nsMgmtCtx *NamespaceMgmtCtx, ns *intf.LinuxInterfaces_Interface_Namespace) (revert func(), err error) {
	if ns != nil && ns.Type == intf.LinuxInterfaces_Interface_Namespace_MICROSERVICE_REF_NS {
		// Convert namespace
		ifNs := h.convertMicroserviceNsToPidNs(ns.Microservice)
		// Back to interface ns type
		ns, err = ifNs.GenericToIfaceNs()
		if err != nil {
			return func() {}, err
		}
		if ns == nil {
			return func() {}, &unavailableMicroserviceErr{}
		}
	}

	// Prepare generic namespace object
	ifaceNs := h.IfNsToGeneric(ns)

	return h.SwitchNamespace(ifaceNs, nsMgmtCtx)
}

// SwitchNamespace switches the network namespace of the current thread.
// Caller should eventually call the returned "revert" function in order to get back to the original
// network namespace (for example using "defer revert()").
func (h *NsHandler) SwitchNamespace(ns *Namespace, ctx *NamespaceMgmtCtx) (revert func(), err error) {
	var nsHandle netns.NsHandle
	if ns != nil && ns.Type == MicroserviceRefNs {
		ns = h.convertMicroserviceNsToPidNs(ns.Microservice)
		if ns == nil {
			return func() {}, &unavailableMicroserviceErr{}
		}
	}

	// Save the current network namespace.
	origns, err := netns.Get()
	if err != nil {
		return func() {}, err
	}

	// Get network namespace file descriptor.
	nsHandle, err = h.getOrCreateNs(ns)
	if err != nil {
		return func() {}, err
	}
	defer nsHandle.Close()

	alreadyLocked := ctx.lockedOsThread
	if !alreadyLocked {
		// Lock the OS Thread so we don't accidentally switch namespaces later.
		runtime.LockOSThread()
		ctx.lockedOsThread = true
	}

	// Switch the namespace.
	l := h.log.WithFields(logging.Fields{"ns": nsHandle.String(), "ns-fd": int(nsHandle)})
	if err := h.sysHandler.SetNamespace(nsHandle); err != nil {
		l.Errorf("Failed to switch Linux network namespace (%v): %v", ns.GenericNsToString(), err)
	}

	return func() {
		l := h.log.WithFields(logging.Fields{"orig-ns": origns.String(), "orig-ns-fd": int(origns)})
		if err := netns.Set(origns); err != nil {
			l.Errorf("Failed to switch Linux network namespace: %v", err)
		}
		origns.Close()
		if !alreadyLocked {
			runtime.UnlockOSThread()
			ctx.lockedOsThread = false
		}
	}, nil
}

// IsNamespaceAvailable returns true if the destination namespace is available.
func (h *NsHandler) IsNamespaceAvailable(ns *intf.LinuxInterfaces_Interface_Namespace) bool {
	if ns != nil && ns.Type == intf.LinuxInterfaces_Interface_Namespace_MICROSERVICE_REF_NS {
		if h.dockerClient == nil {
			return false
		}
		_, available := h.microServiceByLabel[ns.Microservice]
		return available
	}
	return true
}

// IfNsToGeneric converts interface-type namespace to generic type namespace. Such an object can be used to call common
// namespace-related methods
func (h *NsHandler) IfNsToGeneric(ns *intf.LinuxInterfaces_Interface_Namespace) *Namespace {
	if ns == nil {
		return &Namespace{}
	}
	return &Namespace{Type: int32(ns.Type), Pid: ns.Pid, Microservice: ns.Microservice, Name: ns.Name, FilePath: ns.Filepath}
}

// ArpNsToGeneric converts arp-type namespace to generic type namespace. Such an object can be used to call common
// namespace-related methods
func (h *NsHandler) ArpNsToGeneric(ns *l3.LinuxStaticArpEntries_ArpEntry_Namespace) *Namespace {
	if ns == nil {
		return &Namespace{}
	}
	return &Namespace{Type: int32(ns.Type), Pid: ns.Pid, Microservice: ns.Microservice, Name: ns.Name, FilePath: ns.Filepath}
}

// RouteNsToGeneric converts route-type namespace to generic type namespace. Such an object can be used to call common
// namespace-related methods
func (h *NsHandler) RouteNsToGeneric(ns *l3.LinuxStaticRoutes_Route_Namespace) *Namespace {
	if ns == nil {
		return &Namespace{}
	}
	return &Namespace{Type: int32(ns.Type), Pid: ns.Pid, Microservice: ns.Microservice, Name: ns.Name, FilePath: ns.Filepath}
}

// getOrCreateNs returns an existing Linux network namespace or creates a new one if it doesn't exist yet.
// It is, however, only possible to create "named" namespaces. For PID-based namespaces, process with
// the given PID must exists, otherwise the function returns an error.
func (h *NsHandler) getOrCreateNs(ns *Namespace) (netns.NsHandle, error) {
	var nsHandle netns.NsHandle
	var err error

	if ns == nil {
		return dupNsHandle(h.defaultNs)
	}

	switch ns.Type {
	case PidRefNs:
		if ns.Pid == 0 {
			// We consider scheduler's PID as the representation of the default namespace.
			return dupNsHandle(h.defaultNs)
		}
		nsHandle, err = netns.GetFromPid(int(ns.Pid))
		if err != nil {
			return netns.None(), errors.Errorf("failed to get namespace handle from pid: %v", err)
		}
	case NamedNs:
		if ns.Name == "" {
			return dupNsHandle(h.defaultNs)
		}
		nsHandle, err = h.sysHandler.GetNamespaceFromName(ns.Name)
		if err != nil {
			// Create named namespace if it doesn't exist yet.
			_, err = ns.createNamedNetNs(h.sysHandler, h.log)
			if err != nil {
				return netns.None(), errors.Errorf("failed to create named net namspace: %v", err)
			}
			nsHandle, err = netns.GetFromName(ns.Name)
			if err != nil {
				return netns.None(), errors.Errorf("unable to get namespace by name")
			}
		}
	case FileRefNs:
		if ns.FilePath == "" {
			return dupNsHandle(h.defaultNs)
		}
		nsHandle, err = netns.GetFromPath(ns.FilePath)
		if err != nil {
			return netns.None(), errors.Errorf("failed to get file %s from path: %v", ns.FilePath, err)
		}
	case MicroserviceRefNs:
		return netns.None(), errors.Errorf("unable to convert microservice label to PID at this level")
	}

	return nsHandle, nil
}

// Create named namespace used for VETH interface creation instead of the default one.
func (h *NsHandler) prepareConfigNamespace() error {
	// Prepare namespace proto object.
	ns := &Namespace{
		Type: NamedNs,
		Name: configNamespace,
	}
	// Check if namespace exists.
	found, err := ns.namedNetNsExists(h.log)
	if err != nil {
		return errors.Errorf("failed to evaluate namespace %s presence: %v", ns.Name, err)
	}
	// Remove namespace if exists.
	if found {
		err := ns.deleteNamedNetNs(h.sysHandler, h.log)
		if err != nil {
			return errors.Errorf("failed to delete namespace %s: %v", ns.Name, err)
		}
	}
	_, err = ns.createNamedNetNs(h.sysHandler, h.log)
	if err != nil {
		return errors.Errorf("failed to create namespace %s: %v", ns.Name, err)
	}
	h.configNs, err = ns.GenericToIfaceNs()
	if err != nil {
		return errors.Errorf("failed to convert generic namespace %s to interface-type namespace: %v",
			ns.Name, err)
	}
	return nil
}

// convertMicroserviceNsToPidNs converts microservice-referenced namespace into the PID-referenced namespace.
func (h *NsHandler) convertMicroserviceNsToPidNs(microserviceLabel string) (pidNs *Namespace) {
	if microservice, ok := h.microServiceByLabel[microserviceLabel]; ok {
		pidNamespace := &Namespace{}
		pidNamespace.Type = PidRefNs
		pidNamespace.Pid = uint32(microservice.PID)
		return pidNamespace
	}
	return nil
}

// getLinuxIfAddrs returns a list of IP addresses for given linux interface with info whether there is IPv6 address
// (except default link local)
func (h *NsHandler) getLinuxIfAddrs(ifName string) ([]*net.IPNet, bool, error) {
	var networks []*net.IPNet
	addrs, err := h.ifHandler.GetAddressList(ifName)
	if err != nil {
		return nil, false, errors.Errorf("failed to get IP address set from linux interface %s", ifName)
	}
	var containsIPv6 bool
	for _, ipAddr := range addrs {
		network, ipv6, err := ipAddrs.ParseIPWithPrefix(ipAddr.String())
		if err != nil {
			return nil, false, errors.Errorf("failed to parse IP address %s", ipAddr.String())
		}
		// Set once if IP address is version 6 and not a link local address
		if !containsIPv6 && ipv6 && !ipAddr.IP.IsLinkLocalUnicast() {
			containsIPv6 = true
		}
		networks = append(networks, network)
	}

	return networks, containsIPv6, nil
}

// dupNsHandle duplicates namespace handle.
func dupNsHandle(ns netns.NsHandle) (netns.NsHandle, error) {
	dup, err := syscall.Dup(int(ns))
	return netns.NsHandle(dup), err
}
