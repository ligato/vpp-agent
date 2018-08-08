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
	"errors"
	"fmt"
	"net"
	"os"
	"runtime"
	"sync"
	"syscall"

	"github.com/fsouza/go-dockerclient"
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
func (plugin *NsHandler) Init(logger logging.PluginLogger, ifHandler linuxcalls.NetlinkAPI, sysHandler SystemAPI,
	msChan chan *MicroserviceCtx, ifNotif chan *MicroserviceEvent) error {
	// Logger
	plugin.log = logger.NewLogger("-ns-handler")
	plugin.log.Infof("Initializing namespace handler plugin")

	// Init channels
	plugin.microserviceChan = msChan
	plugin.ifMicroserviceNotif = ifNotif

	plugin.ctx, plugin.cancel = context.WithCancel(context.Background())

	plugin.microServiceByLabel = make(map[string]*Microservice)
	plugin.microServiceByID = make(map[string]*Microservice)

	// Handlers
	plugin.ifHandler = ifHandler
	plugin.sysHandler = sysHandler

	// Default namespace
	var err error
	plugin.defaultNs, err = netns.Get()
	if err != nil {
		return fmt.Errorf("failed to init default namespace: %v", err)
	}

	// Docker client
	plugin.dockerClient, err = docker.NewClientFromEnv()
	if err != nil {
		plugin.log.WithFields(logging.Fields{
			"DOCKER_HOST":       os.Getenv("DOCKER_HOST"),
			"DOCKER_TLS_VERIFY": os.Getenv("DOCKER_TLS_VERIFY"),
			"DOCKER_CERT_PATH":  os.Getenv("DOCKER_CERT_PATH"),
		}).Errorf("Failed to get docker client instance from the environment variables: %v", err)
		return err
	}
	plugin.log.Debugf("Using docker client endpoint: %+v", plugin.dockerClient.Endpoint())

	// Create config namespace (for VETHs)
	err = plugin.prepareConfigNamespace()

	// Start microservice tracker
	go plugin.trackMicroservices(plugin.ctx)

	return err
}

// Close pre-configured namespace
func (plugin *NsHandler) Close() error {
	var wasErr error
	if plugin.configNs != nil {
		// Remove veth pre-configure namespace
		ns := plugin.IfNsToGeneric(plugin.configNs)
		wasErr = ns.deleteNamedNetNs(plugin.sysHandler, plugin.log)
		plugin.cancel()
		plugin.wg.Wait()
	}

	return wasErr
}

// GetConfigNamespace return configuration namespace object
func (plugin *NsHandler) GetConfigNamespace() *intf.LinuxInterfaces_Interface_Namespace {
	return plugin.configNs
}

// GetMicroserviceByLabel returns internal microservice-by-label mapping
func (plugin *NsHandler) GetMicroserviceByLabel() map[string]*Microservice {
	return plugin.microServiceByLabel
}

// GetMicroserviceByID returns internal microservice-by-id mapping
func (plugin *NsHandler) GetMicroserviceByID() map[string]*Microservice {
	return plugin.microServiceByID
}

// SetInterfaceNamespace moves a given Linux interface into a specified namespace.
func (plugin *NsHandler) SetInterfaceNamespace(ctx *NamespaceMgmtCtx, ifName string, namespace *intf.LinuxInterfaces_Interface_Namespace) error {
	// Convert microservice namespace
	var err error
	if namespace != nil && namespace.Type == intf.LinuxInterfaces_Interface_Namespace_MICROSERVICE_REF_NS {
		// Convert namespace
		ifNs := plugin.convertMicroserviceNsToPidNs(namespace.Microservice)
		// Back to interface ns type
		namespace, err = ifNs.GenericToIfaceNs()
		if err != nil {
			return err
		}
		if namespace == nil {
			return &unavailableMicroserviceErr{}
		}
	}

	ifaceNs := plugin.IfNsToGeneric(namespace)

	// Get network namespace file descriptor
	ns, err := plugin.getOrCreateNs(ifaceNs)
	if err != nil {
		return err
	}
	defer ns.Close()

	// Get the link plugin.
	link, err := plugin.ifHandler.GetLinkByName(ifName)
	if err != nil {
		return err
	}

	// When interface moves from one namespace to another, it loses all its IP addresses, admin status
	// and MTU configuration -- we need to remember the interface configuration before the move
	// and re-configure the interface in the new namespace.
	addresses, isIPv6, err := plugin.getLinuxIfAddrs(link.Attrs().Name)
	if err != nil {
		return err
	}

	// Move the interface into the namespace.
	err = plugin.sysHandler.LinkSetNsFd(link, int(ns))
	if err != nil {
		return err
	}
	plugin.log.WithFields(logging.Fields{"ifName": ifName, "dest-namespace": plugin.IfaceNsToString(namespace),
		"dest-namespace-fd": int(ns)}).
		Debug("Moved Linux interface across namespaces")

	// Re-configure interface in its new namespace
	revertNs, err := plugin.SwitchNamespace(ifaceNs, ctx)
	if err != nil {
		return err
	}
	defer revertNs()

	if link.Attrs().Flags&net.FlagUp == 1 {
		// Re-enable interface
		err = plugin.ifHandler.SetInterfaceUp(ifName)
		if nil != err {
			return fmt.Errorf("failed to enable Linux interface `%s`: %v", ifName, err)
		}
		plugin.log.WithFields(logging.Fields{"ifName": ifName}).
			Debug("Linux interface was re-enabled")
	}

	// Re-add IP addresses
	for _, address := range addresses {
		// Skip IPv6 link local address if there is no other IPv6 address
		if !isIPv6 && address.IP.IsLinkLocalUnicast(){
			continue
		}
		err = plugin.ifHandler.AddInterfaceIP(ifName, address)
		if err != nil {
			if err.Error() == "file exists" {
				continue
			}
			return fmt.Errorf("failed to assign IP address to a Linux interface `%s`: %v", ifName, err)
		}
		plugin.log.WithFields(logging.Fields{"ifName": ifName, "addr": address}).
			Debug("IP address was re-assigned to Linux interface")
	}

	// Revert back the MTU config
	err = plugin.ifHandler.SetInterfaceMTU(ifName, link.Attrs().MTU)
	if nil != err {
		return fmt.Errorf("failed to set MTU of a Linux interface `%s`: %v", ifName, err)
	}
	plugin.log.WithFields(logging.Fields{"ifName": ifName, "mtu": link.Attrs().MTU}).
		Debug("MTU was reconfigured for Linux interface")

	return nil
}

// switchToNamespace switches the network namespace of the current thread.
func (plugin *NsHandler) SwitchToNamespace(nsMgmtCtx *NamespaceMgmtCtx, ns *intf.LinuxInterfaces_Interface_Namespace) (revert func(), err error) {
	if ns != nil && ns.Type == intf.LinuxInterfaces_Interface_Namespace_MICROSERVICE_REF_NS {
		// Convert namespace
		ifNs := plugin.convertMicroserviceNsToPidNs(ns.Microservice)
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
	ifaceNs := plugin.IfNsToGeneric(ns)

	return plugin.SwitchNamespace(ifaceNs, nsMgmtCtx)
}

// SwitchNamespace switches the network namespace of the current thread.
// Caller should eventually call the returned "revert" function in order to get back to the original
// network namespace (for example using "defer revert()").
func (plugin *NsHandler) SwitchNamespace(ns *Namespace, ctx *NamespaceMgmtCtx) (revert func(), err error) {
	var nsHandle netns.NsHandle
	if ns != nil && ns.Type == MicroserviceRefNs {
		ns = plugin.convertMicroserviceNsToPidNs(ns.Microservice)
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
	nsHandle, err = plugin.getOrCreateNs(ns)
	if err != nil {
		return func() {}, err
	}
	defer nsHandle.Close()

	alreadyLocked := ctx.lockedOsThread
	if !alreadyLocked {
		// Lock the OS Thread so we don't accidentally switch namespaces later.
		runtime.LockOSThread()
		ctx.lockedOsThread = true
		plugin.log.Debug("Locked OS thread")
	}

	// Switch the namespace.
	l := plugin.log.WithFields(logging.Fields{"ns": nsHandle.String(), "ns-fd": int(nsHandle)})
	if err := plugin.sysHandler.SetNamespace(nsHandle); err != nil {
		l.Errorf("Failed to switch Linux network namespace (%v): %v", ns.GenericNsToString(), err)
	} else {
		l.Debugf("Switched Linux network namespace (%v)", ns.GenericNsToString())
	}

	return func() {
		l := plugin.log.WithFields(logging.Fields{"orig-ns": origns.String(), "orig-ns-fd": int(origns)})
		if err := netns.Set(origns); err != nil {
			l.Errorf("Failed to switch Linux network namespace: %v", err)
		} else {
			l.Debugf("Switched back to the original Linux network namespace")
		}
		origns.Close()
		if !alreadyLocked {
			runtime.UnlockOSThread()
			ctx.lockedOsThread = false
			plugin.log.Debug("Unlocked OS thread")
		}
	}, nil
}

// IsNamespaceAvailable returns true if the destination namespace is available.
func (plugin *NsHandler) IsNamespaceAvailable(ns *intf.LinuxInterfaces_Interface_Namespace) bool {
	if ns != nil && ns.Type == intf.LinuxInterfaces_Interface_Namespace_MICROSERVICE_REF_NS {
		if plugin.dockerClient == nil {
			return false
		}
		_, available := plugin.microServiceByLabel[ns.Microservice]
		return available
	}
	return true
}

// IfNsToGeneric converts interface-type namespace to generic type namespace. Such an object can be used to call common
// namespace-related methods
func (plugin *NsHandler) IfNsToGeneric(ns *intf.LinuxInterfaces_Interface_Namespace) *Namespace {
	if ns == nil {
		return &Namespace{}
	}
	return &Namespace{Type: int32(ns.Type), Pid: ns.Pid, Microservice: ns.Microservice, Name: ns.Name, FilePath: ns.Filepath}
}

// ArpNsToGeneric converts arp-type namespace to generic type namespace. Such an object can be used to call common
// namespace-related methods
func (plugin *NsHandler) ArpNsToGeneric(ns *l3.LinuxStaticArpEntries_ArpEntry_Namespace) *Namespace {
	if ns == nil {
		return &Namespace{}
	}
	return &Namespace{Type: int32(ns.Type), Pid: ns.Pid, Microservice: ns.Microservice, Name: ns.Name, FilePath: ns.Filepath}
}

// RouteNsToGeneric converts route-type namespace to generic type namespace. Such an object can be used to call common
// namespace-related methods
func (plugin *NsHandler) RouteNsToGeneric(ns *l3.LinuxStaticRoutes_Route_Namespace) *Namespace {
	if ns == nil {
		return &Namespace{}
	}
	return &Namespace{Type: int32(ns.Type), Pid: ns.Pid, Microservice: ns.Microservice, Name: ns.Name, FilePath: ns.Filepath}
}

// getOrCreateNs returns an existing Linux network namespace or creates a new one if it doesn't exist yet.
// It is, however, only possible to create "named" namespaces. For PID-based namespaces, process with
// the given PID must exists, otherwise the function returns an error.
func (plugin *NsHandler) getOrCreateNs(ns *Namespace) (netns.NsHandle, error) {
	var nsHandle netns.NsHandle
	var err error

	if ns == nil {
		return dupNsHandle(plugin.defaultNs)
	}

	switch ns.Type {
	case PidRefNs:
		if ns.Pid == 0 {
			// We consider scheduler's PID as the representation of the default namespace.
			return dupNsHandle(plugin.defaultNs)
		}
		nsHandle, err = netns.GetFromPid(int(ns.Pid))
		if err != nil {
			return netns.None(), err
		}
	case NamedNs:
		if ns.Name == "" {
			return dupNsHandle(plugin.defaultNs)
		}
		nsHandle, err = plugin.sysHandler.GetNamespaceFromName(ns.Name)
		if err != nil {
			// Create named namespace if it doesn't exist yet.
			_, err = ns.createNamedNetNs(plugin.sysHandler, plugin.log)
			if err != nil {
				return netns.None(), err
			}
			nsHandle, err = netns.GetFromName(ns.Name)
			if err != nil {
				return netns.None(), errors.New("unable to get namespace by name")
			}
		}
	case FileRefNs:
		if ns.FilePath == "" {
			return dupNsHandle(plugin.defaultNs)
		}
		nsHandle, err = netns.GetFromPath(ns.FilePath)
		if err != nil {
			return netns.None(), err
		}
	case MicroserviceRefNs:
		return netns.None(), errors.New("unable to convert microservice label to PID at this level")
	}

	return nsHandle, nil
}

// Create named namespace used for VETH interface creation instead of the default one.
func (plugin *NsHandler) prepareConfigNamespace() error {
	// Prepare namespace proto object.
	ns := &Namespace{
		Type: NamedNs,
		Name: configNamespace,
	}
	// Check if namespace exists.
	found, err := ns.namedNetNsExists(plugin.log)
	if err != nil {
		return err
	}
	// Remove namespace if exists.
	if found {
		err := ns.deleteNamedNetNs(plugin.sysHandler, plugin.log)
		if err != nil {
			return err
		}
	}
	_, err = ns.createNamedNetNs(plugin.sysHandler, plugin.log)
	if err != nil {
		return err
	}
	plugin.configNs, err = ns.GenericToIfaceNs()
	return err
}

// convertMicroserviceNsToPidNs converts microservice-referenced namespace into the PID-referenced namespace.
func (plugin *NsHandler) convertMicroserviceNsToPidNs(microserviceLabel string) (pidNs *Namespace) {
	if microservice, ok := plugin.microServiceByLabel[microserviceLabel]; ok {
		pidNamespace := &Namespace{}
		pidNamespace.Type = PidRefNs
		pidNamespace.Pid = uint32(microservice.Pid)
		return pidNamespace
	}
	return nil
}

// getLinuxIfAddrs returns a list of IP addresses for given linux interface with info whether there is IPv6 address
// (except default link local)
func (plugin *NsHandler) getLinuxIfAddrs(ifName string) ([]*net.IPNet, bool, error) {
	var networks []*net.IPNet
	addrs, err := plugin.ifHandler.GetAddressList(ifName)
	if err != nil {
		return nil, false, fmt.Errorf("failed to get IP address set from linux interface %s", ifName)
	}
	var containsIPv6 bool
	for _, ipAddr := range addrs {
		network, ipv6, err := ipAddrs.ParseIPWithPrefix(ipAddr.String())
		if err != nil {
			return nil, false, fmt.Errorf("failed to parse IP address %s", ipAddr.String())
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
