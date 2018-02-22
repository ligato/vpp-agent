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

//go:generate protoc --proto_path=../common/model/interfaces --gogo_out=../common/model/interfaces ../common/model/interfaces/interfaces.proto

package ifplugin

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fsouza/go-dockerclient"
	"github.com/ligato/cn-infra/utils/safeclose"
	vppIfIdx "github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/ifaceidx"
	"github.com/vishvananda/netlink"

	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/cn-infra/servicelabel"
	"github.com/ligato/cn-infra/utils/addrs"
	"github.com/ligato/vpp-agent/plugins/linuxplugin/common/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/linuxplugin/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/linuxplugin/ifplugin/linuxcalls"
)

/* how often in seconds to refresh the microservice label -> docker container PID map */
const (
	dockerRefreshPeriod = 3 * time.Second
	dockerRetryPeriod   = 5 * time.Second
	vethConfigNamespace = "veth-cfg-ns"
)

// Interface type string representations
const (
	veth   = "veth"
	tapTun = "tun"
)

// MicroserviceCtx contains all data required to handle microservice changes
type MicroserviceCtx struct {
	nsMgmtCtx     *linuxcalls.NamespaceMgmtCtx
	created       []string
	since         string
	lastInspected int64
}

// LinuxInterfaceConfig is used to cache the configuration of Linux interfaces.
type LinuxInterfaceConfig struct {
	config *interfaces.LinuxInterfaces_Interface
	peer   *LinuxInterfaceConfig
}

// Microservice is used to store PID and ID of the container running a given microservice.
type Microservice struct {
	label string
	pid   int
	id    string
}

// unavailableMicroserviceErr is error implementation used when a given microservice is not deployed.
type unavailableMicroserviceErr struct {
	label string
}

func (e *unavailableMicroserviceErr) Error() string {
	return fmt.Sprintf("Microservice '%s' is not available", e.label)
}

// LinuxInterfaceConfigurator runs in the background in its own goroutine where it watches for any changes
// in the configuration of interfaces as modelled by the proto file "model/interfaces/interfaces.proto"
// and stored in ETCD under the key "/vnf-agent/{vnf-agent}/linux/config/v1/interface".
// Updates received from the northbound API are compared with the Linux network configuration and differences
// are applied through the Netlink API.
type LinuxInterfaceConfigurator struct {
	Log logging.Logger

	cfgLock sync.Mutex

	/* logical interface name -> Linux interface index (both managed and unmanaged interfaces) */
	ifIndexes ifaceidx.LinuxIfIndexRW

	/* interface caches (managed interfaces only) */
	intfByName          map[string]*LinuxInterfaceConfig   /* interface name -> interface configuration */
	intfsByMicroservice map[string][]*LinuxInterfaceConfig /* microservice label -> list of interfaces attached to this microservice */

	/* microservice caches */
	microserviceByLabel map[string]*Microservice /* microservice label -> microservice info */
	microserviceByID    map[string]*Microservice /* microservice container ID -> microservice info */

	/* docker client - used to convert microservice label into the PID and ID of the container */
	dockerClient *docker.Client

	/* management of go routines */
	ctx    context.Context    // Context within which all goroutines are running
	cancel context.CancelFunc // Cancel can be used to cancel all goroutines and their jobs inside of the plugin.
	wg     sync.WaitGroup     // Wait group allows to wait until all goroutines of the plugin have finished.

	/* veth pre-configure namespace */
	vethCfgNamespace *interfaces.LinuxInterfaces_Interface_Namespace

	/* channel to send microservice updates */
	microserviceChan chan *MicroserviceCtx

	/* interface state */
	ifStateChan chan *LinuxInterfaceStateNotification

	/* VPP interface indices */
	VppIfIndices vppIfIdx.SwIfIndex

	/* time measurement */
	Stopwatch *measure.Stopwatch // timer used to measure and store time
}

// Init linuxplugin and start go routines.
func (plugin *LinuxInterfaceConfigurator) Init(stateChan chan *LinuxInterfaceStateNotification,
	ifIndexes ifaceidx.LinuxIfIndexRW, msChan chan *MicroserviceCtx) (err error) {
	plugin.Log.Debug("Initializing Linux Interface configurator")
	plugin.ifIndexes = ifIndexes

	// Init channel
	plugin.ifStateChan = stateChan
	plugin.microserviceChan = msChan

	// Allocate caches.
	plugin.intfByName = make(map[string]*LinuxInterfaceConfig)
	plugin.intfsByMicroservice = make(map[string][]*LinuxInterfaceConfig)
	plugin.microserviceByLabel = make(map[string]*Microservice)
	plugin.microserviceByID = make(map[string]*Microservice)

	plugin.dockerClient, err = docker.NewClientFromEnv()
	if err != nil {
		plugin.Log.WithFields(logging.Fields{
			"DOCKER_HOST":       os.Getenv("DOCKER_HOST"),
			"DOCKER_TLS_VERIFY": os.Getenv("DOCKER_TLS_VERIFY"),
			"DOCKER_CERT_PATH":  os.Getenv("DOCKER_CERT_PATH"),
		}).Errorf("Failed to get docker client instance from the environment variables: %v", err)
		return err
	}
	plugin.Log.Debugf("Using docker client endpoint: %+v", plugin.dockerClient.Endpoint())

	plugin.ctx, plugin.cancel = context.WithCancel(context.Background())
	go plugin.trackMicroservices(plugin.ctx)

	// Create cfg namespace
	err = plugin.prepareVethConfigNamespace()

	// Start watching on state updater events
	go plugin.watchLinuxStateUpdater()

	return err
}

// Close stops all goroutines started by linuxplugin
func (plugin *LinuxInterfaceConfigurator) Close() error {
	// remove veth pre-configure namespace
	wasErr := linuxcalls.DeleteNamedNetNs(plugin.vethCfgNamespace.Name, plugin.Log)
	plugin.cancel()
	plugin.wg.Wait()

	safeclose.Close(plugin.ifStateChan)

	return wasErr
}

// LookupLinuxInterfaces looks up all currently unmanaged Linux interfaces in the current namespace and registers them into
// the name-to-index mapping. Furthermore, it triggers goroutine that will watch for newly added interfaces (by another party)
// unless it is already running.
func (plugin *LinuxInterfaceConfigurator) LookupLinuxInterfaces() error {
	plugin.cfgLock.Lock()
	defer plugin.cfgLock.Unlock()

	intfs, err := net.Interfaces()
	if err != nil {
		return err
	}
	for _, inter := range intfs {
		idx := GetLinuxInterfaceIndex(inter.Name)
		if idx < 0 {
			continue
		}
		res := plugin.ifIndexes.LookupNameByHostIfName(inter.Name)
		if len(res) == 1 {
			continue
		}
		plugin.Log.WithFields(logging.Fields{"name": inter.Name, "idx": idx}).Debug("Found new Linux interface")
		plugin.ifIndexes.RegisterName(inter.Name, uint32(idx), &interfaces.LinuxInterfaces_Interface{
			Name: inter.Name, HostIfName: inter.Name})
	}
	return nil
}

// ConfigureLinuxInterface reacts to a new northbound Linux interface config by creating and configuring
// the interface in the host network stack through Netlink API.
func (plugin *LinuxInterfaceConfigurator) ConfigureLinuxInterface(linuxIf *interfaces.LinuxInterfaces_Interface) error {
	plugin.cfgLock.Lock()
	defer plugin.cfgLock.Unlock()

	plugin.handleOptionalHostIfName(linuxIf)
	plugin.Log.Infof("Configuring new Linux interface %v", linuxIf.HostIfName)

	// Linux interface type resolution
	switch linuxIf.Type {
	case interfaces.LinuxInterfaces_VETH:
		// Get peer interface config if exists and cache the original configuration with peer
		if linuxIf.Veth == nil {
			return fmt.Errorf("VETH interface %v has no peer defined", linuxIf.HostIfName)
		}
		peerConfig := plugin.getInterfaceConfig(linuxIf.Veth.PeerIfName)
		ifConfig := plugin.addToCache(linuxIf, peerConfig)

		return plugin.configureVethInterface(ifConfig, peerConfig)
	case interfaces.LinuxInterfaces_AUTO_TAP:
		// TAP (auto) interface looks for existing interface with the same host name or temp name (cached without peer)
		ifConfig := plugin.addToCache(linuxIf, nil)

		return plugin.configureTapInterface(ifConfig)
	default:
		return fmt.Errorf("unknown linux interface type: %v", linuxIf.Type)
	}
}

// ModifyLinuxInterface applies changes in the NB configuration of a Linux interface into the host network stack
// through Netlink API.
func (plugin *LinuxInterfaceConfigurator) ModifyLinuxInterface(newLinuxIf, oldLinuxIf *interfaces.LinuxInterfaces_Interface) (err error) {
	// If host names are not defined, name == host name
	plugin.handleOptionalHostIfName(newLinuxIf)
	plugin.handleOptionalHostIfName(oldLinuxIf)
	plugin.Log.Infof("Modifying Linux interface %v", newLinuxIf.HostIfName)

	if oldLinuxIf.Type != newLinuxIf.Type {
		return fmt.Errorf("%v: linux interface type change not allowed", newLinuxIf.Name)
	}

	// Get old and new peer/host interfaces (peers for VETH, host for TAP)
	var oldPeer, newPeer string
	if oldLinuxIf.Type == interfaces.LinuxInterfaces_VETH && oldLinuxIf.Veth != nil {
		oldPeer = oldLinuxIf.Veth.PeerIfName
	} else if oldLinuxIf.Type == interfaces.LinuxInterfaces_AUTO_TAP {
		oldPeer = oldLinuxIf.HostIfName
	}
	if newLinuxIf.Type == interfaces.LinuxInterfaces_VETH && newLinuxIf.Veth != nil {
		newPeer = newLinuxIf.Veth.PeerIfName
	} else if newLinuxIf.Type == interfaces.LinuxInterfaces_AUTO_TAP {
		newPeer = newLinuxIf.HostIfName
	}

	// Prepare namespace objects of new and old interfaces
	newIfaceNs := linuxcalls.ToGenericNs(newLinuxIf.Namespace)
	oldIfaceNs := linuxcalls.ToGenericNs(oldLinuxIf.Namespace)
	if newPeer != oldPeer || newLinuxIf.HostIfName != oldLinuxIf.HostIfName || newIfaceNs.CompareNamespaces(oldIfaceNs) != 0 {
		// Change of the peer interface (VETH) or host (TAP) or the namespace requires to create the interface from the scratch.
		err := plugin.DeleteLinuxInterface(oldLinuxIf)
		if err == nil {
			err = plugin.ConfigureLinuxInterface(newLinuxIf)
		}
		return err
	}

	plugin.cfgLock.Lock()
	defer plugin.cfgLock.Unlock()

	// Update the cached configuration.
	plugin.removeFromCache(oldLinuxIf)
	peer := plugin.getInterfaceConfig(newPeer)
	plugin.addToCache(newLinuxIf, peer)

	// Verify required namespace
	if !plugin.isNamespaceAvailable(newLinuxIf.Namespace) {
		plugin.Log.Errorf("unable to configure linux interface %v: interface namespace is not available",
			newLinuxIf.HostIfName)
		return nil
	}

	// Validate configuration/namespace according to interface type
	if newLinuxIf.Type == interfaces.LinuxInterfaces_VETH {
		if peer == nil {
			// Interface doesn't actually exist physically.
			plugin.Log.Infof("cannot configure linux interface %v: peer interface %v is not configured yet",
				newLinuxIf.HostIfName, newPeer)
			return nil
		}
		if !plugin.isNamespaceAvailable(oldLinuxIf.Namespace) {
			plugin.Log.Warnf("unable to modify linux interface %v: peer interface namespace is not available",
				oldLinuxIf.HostIfName)
			return nil
		}
		if !plugin.isNamespaceAvailable(newLinuxIf.Namespace) {
			plugin.Log.Warnf("unable to modify linux interface %v: interface namespace is not available",
				newLinuxIf.HostIfName)
			return nil
		}
	} else if newLinuxIf.Type == interfaces.LinuxInterfaces_AUTO_TAP {
		if !plugin.isNamespaceAvailable(newLinuxIf.Namespace) {
			// Interface doesn't actually exist physically.
			plugin.Log.WithField("ifName", newLinuxIf.Name).Debug("Linux interface is not ready to be re-configured")
			return nil
		}
	} else {
		plugin.Log.Warnf("Unknown interface type %v", newLinuxIf.Type)

		return nil
	}

	// The namespace was not changed so interface can be reconfigured
	nsMgmtCtx := linuxcalls.NewNamespaceMgmtCtx()

	return plugin.modifyLinuxInterface(nsMgmtCtx, oldLinuxIf, newLinuxIf)
}

// DeleteLinuxInterface reacts to a removed NB configuration of a Linux interface.
func (plugin *LinuxInterfaceConfigurator) DeleteLinuxInterface(linuxIf *interfaces.LinuxInterfaces_Interface) error {
	plugin.cfgLock.Lock()
	defer plugin.cfgLock.Unlock()

	plugin.handleOptionalHostIfName(linuxIf)
	plugin.Log.Infof("Removing Linux interface %v", linuxIf.HostIfName)

	oldConfig := plugin.removeFromCache(linuxIf)
	var peerConfig *LinuxInterfaceConfig
	if oldConfig != nil {
		peerConfig = oldConfig.peer
	}

	if linuxIf.Type == interfaces.LinuxInterfaces_AUTO_TAP {
		return plugin.deleteTapInterface(oldConfig)
	} else if linuxIf.Type == interfaces.LinuxInterfaces_VETH {
		return plugin.deleteVethInterface(oldConfig, peerConfig)
	}
	plugin.Log.Warnf("Unknown type of interface: %v", linuxIf.Type)
	return nil
}

// Validate, create and configure VETH type linux interface
func (plugin *LinuxInterfaceConfigurator) configureVethInterface(ifConfig, peerConfig *LinuxInterfaceConfig) error {
	plugin.Log.WithFields(logging.Fields{"name": ifConfig.config.Name, "hostName": ifConfig.config.HostIfName,
		"peer": ifConfig.config.Veth.PeerIfName}).Debug("Configuring new Veth interface")
	// Create VETH after both end's configs and target namespaces are available.
	if peerConfig == nil {
		plugin.Log.Infof("cannot configure linux interface %v: peer interface %v is not configured yet",
			ifConfig.config.HostIfName, ifConfig.config.Veth.PeerIfName)
		return nil
	}
	if !plugin.isNamespaceAvailable(ifConfig.config.Namespace) {
		plugin.Log.Warnf("unable to configure linux interface %v: interface namespace is not available",
			ifConfig.config.HostIfName)
		return nil
	}
	if !plugin.isNamespaceAvailable(peerConfig.config.Namespace) {
		plugin.Log.Warnf("unable to configure linux interface %v: peer namespace is not available",
			ifConfig.config.HostIfName)
		return nil
	}

	nsMgmtCtx := linuxcalls.NewNamespaceMgmtCtx()

	// Prepare generic veth config namespace object
	vethNs := linuxcalls.ToGenericNs(plugin.vethCfgNamespace)

	// Switch to veth cfg namespace
	revertNs, err := vethNs.SwitchNamespace(nsMgmtCtx, plugin.Log)
	if err != nil {
		return err
	}
	defer revertNs()

	if err := plugin.addVethInterfacePair(nsMgmtCtx, ifConfig.config, peerConfig.config); err != nil {
		return err
	}

	if err := plugin.configureLinuxInterface(nsMgmtCtx, ifConfig.config); err != nil {
		return err
	}
	if err := plugin.configureLinuxInterface(nsMgmtCtx, peerConfig.config); err != nil {
		return err
	}

	plugin.Log.Infof("Linux interface %v with hostIfName %v configured", ifConfig.config.Name, ifConfig.config.HostIfName)

	return nil
}

// Validate and apply linux TAP configuration to the interface. The interface is not created here, it is added
// to the default namespace when it's VPP end is configured
func (plugin *LinuxInterfaceConfigurator) configureTapInterface(ifConfig *LinuxInterfaceConfig) error {
	plugin.Log.WithFields(logging.Fields{"name": ifConfig.config.Name,
		"hostName": ifConfig.config.HostIfName}).Debug("Applying new Linux TAP interface configuration")

	// Tap interfaces can be processed directly using config and also via linux interface events. This check
	// should prevent to process the same interface multiple times.
	_, _, exists := plugin.ifIndexes.LookupIdx(ifConfig.config.Name)
	if exists {
		plugin.Log.Debugf("TAP interface %v already processed", ifConfig.config.Name)
		return nil
	}

	// Search default namespace for appropriate interface
	linuxIfs, err := netlink.LinkList()
	if err != nil {
		return fmt.Errorf("failed to read linux interfaces: %v", err)
	}

	nsMgmtCtx := linuxcalls.NewNamespaceMgmtCtx()

	// Verify availability of namespace from configuration
	if !plugin.isNamespaceAvailable(ifConfig.config.Namespace) {
		plugin.Log.Errorf("unable to apply linux TAP configuration %v: destination namespace is not available",
			ifConfig.config.Name, ifConfig.config.HostIfName)
		return nil
	}

	// Check if TAP temporary name is defined
	if ifConfig.config.Tap == nil || ifConfig.config.Tap.TempIfName == "" {
		plugin.Log.Debugf("Tap interface %v temporary name not defined", ifConfig.config.HostIfName)
		// In such a case, set temp name as host (look for interface named as host name)
		ifConfig.config.Tap = &interfaces.LinuxInterfaces_Interface_Tap{
			TempIfName: ifConfig.config.HostIfName,
		}
	}

	// Try to find temp interface in default namespace.
	var found bool
	for _, linuxIf := range linuxIfs {
		if ifConfig.config.Tap.TempIfName == linuxIf.Attrs().Name {
			if linuxIf.Type() == tapTun {
				found = true
				break
			}
			plugin.Log.Debugf("Linux TAP config %v found linux interface %v, but it is not the TAP interface type",
				ifConfig.config.Name, ifConfig.config.HostIfName)
		}
	}
	if !found {
		plugin.Log.Debugf("Linux TAP config %v did not found the linux interface with name %v", ifConfig.config.Name,
			ifConfig.config.Tap.TempIfName)
		return nil
	}

	return plugin.configureLinuxInterface(nsMgmtCtx, ifConfig.config)
}

// Set linux interface to proper namespace and configure attributes
func (plugin *LinuxInterfaceConfigurator) configureLinuxInterface(nsMgmtCtx *linuxcalls.NamespaceMgmtCtx, ifConfig *interfaces.LinuxInterfaces_Interface) (err error) {
	if ifConfig.HostIfName == "" {
		return fmt.Errorf("host interface not specified for %v", ifConfig.Name)
	}

	// Move interface to the proper namespace.
	ns := ifConfig.Namespace
	if ns != nil && ns.Type == interfaces.LinuxInterfaces_Interface_Namespace_MICROSERVICE_REF_NS {
		ns = plugin.convertMicroserviceNsToPidNs(ns.Microservice)
		if ns == nil {
			return &unavailableMicroserviceErr{}
		}
	}

	// Use temporary/host name (according to type) to set interface to different namespace
	if ifConfig.Type == interfaces.LinuxInterfaces_AUTO_TAP {
		err = linuxcalls.SetInterfaceNamespace(nsMgmtCtx, ifConfig.Tap.TempIfName, ifConfig.Namespace, plugin.Log, plugin.Stopwatch)
		if err != nil {
			return fmt.Errorf("failed to set TAP interface %s to namespace %s: %v", ifConfig.Tap.TempIfName, ifConfig.Namespace, err)
		}
	} else {
		err = linuxcalls.SetInterfaceNamespace(nsMgmtCtx, ifConfig.HostIfName, ifConfig.Namespace, plugin.Log, plugin.Stopwatch)
		if err != nil {
			return fmt.Errorf("failed to set interface %s to namespace %s: %v", ifConfig.HostIfName, ifConfig.Namespace, err)
		}
	}
	// Continue configuring interface in its namespace.
	revertNs, err := plugin.switchToNamespace(nsMgmtCtx, ifConfig.Namespace)
	if err != nil {
		return fmt.Errorf("failed to switch network namespace: %v", err)
	}
	defer revertNs()

	// For TAP interfaces only - rename interface to the actual host name if needed
	if ifConfig.Type == interfaces.LinuxInterfaces_AUTO_TAP {
		if ifConfig.HostIfName != ifConfig.Tap.TempIfName {
			if err := linuxcalls.RenameInterface(ifConfig.Tap.TempIfName, ifConfig.HostIfName,
				measure.GetTimeLog("rename-linux-interface", plugin.Stopwatch)); err != nil {
				plugin.Log.Errorf("Failed to rename TAP interface from %s to %s: %v", ifConfig.Tap.TempIfName,
					ifConfig.HostIfName, err)
				return err
			}
		} else {
			plugin.Log.Debugf("Renaming of TAP interface %v skipped, host name is the same as temporary", ifConfig.HostIfName)
		}
	}

	var wasErr error

	// Set interface up.
	if ifConfig.Enabled {
		err := linuxcalls.InterfaceAdminUp(ifConfig.HostIfName, measure.GetTimeLog("iface_admin_up", plugin.Stopwatch))
		if nil != err {
			wasErr = fmt.Errorf("failed to enable Linux interface: %v", err)
			plugin.Log.Error(wasErr)
		}
	}

	// Set interface MAC address
	if ifConfig.PhysAddress != "" {
		err = linuxcalls.SetInterfaceMac(ifConfig.HostIfName, ifConfig.PhysAddress, nil)
		if err != nil {
			wasErr = fmt.Errorf("cannot assign MAC '%s': %v", ifConfig.PhysAddress, err)
			plugin.Log.Error(wasErr)
		}
		plugin.Log.Debugf("MAC '%s' set to interface %s", ifConfig.PhysAddress, ifConfig.HostIfName)
	}

	// Set interface IP addresses
	ipAddresses, err := addrs.StrAddrsToStruct(ifConfig.IpAddresses)
	if err != nil {
		plugin.Log.Error(err)
		wasErr = err
	}
	for i, ipAddress := range ipAddresses {
		err = linuxcalls.AddInterfaceIP(plugin.Log, ifConfig.HostIfName, ipAddresses[i], nil)
		if err != nil {
			err = fmt.Errorf("cannot assign IP address '%s': %v", ipAddress, err)
			plugin.Log.Error(err)
			wasErr = err
		} else {
			plugin.Log.Debugf("IP address '%s' set to interface %s", ipAddress, ifConfig.HostIfName)
		}
	}

	if ifConfig.Mtu != 0 {
		linuxcalls.SetInterfaceMTU(ifConfig.HostIfName, int(ifConfig.Mtu), nil)
		plugin.Log.Debugf("MTU %d set to interface %s", ifConfig.Mtu, ifConfig.HostIfName)
	}

	idx := GetLinuxInterfaceIndex(ifConfig.HostIfName)
	if idx < 0 {
		return fmt.Errorf("failed to get index of the Linux interface %s", ifConfig.HostIfName)
	}

	// Register interface with its original name and store host name in metadata
	plugin.ifIndexes.RegisterName(ifConfig.Name, uint32(idx), ifConfig)
	plugin.Log.WithFields(logging.Fields{"ifName": ifConfig.Name, "ifIdx": idx}).
		Info("An entry added into ifState.")

	return wasErr
}

// Update linux interface attributes in it's namespace
func (plugin *LinuxInterfaceConfigurator) modifyLinuxInterface(nsMgmtCtx *linuxcalls.NamespaceMgmtCtx,
	oldIfConfig, newIfConfig *interfaces.LinuxInterfaces_Interface) error {
	// Switch to required namespace
	revertNs, err := plugin.switchToNamespace(nsMgmtCtx, oldIfConfig.Namespace)
	if err != nil {
		return fmt.Errorf("failed to switch network namespace: %v", err)
	}
	defer revertNs()

	// Verify that the interface already exists in the Linux namespace.
	idx := GetLinuxInterfaceIndex(oldIfConfig.HostIfName)
	if idx < 0 {
		plugin.Log.Debugf("Host interface %v was not found", oldIfConfig.HostIfName)
		// If host does not exist, configure new setup as a new one
		return plugin.ConfigureLinuxInterface(newIfConfig)
	}

	var wasErr error

	// Set admin status.
	if newIfConfig.Enabled != oldIfConfig.Enabled {
		if newIfConfig.Enabled {
			err = linuxcalls.InterfaceAdminUp(newIfConfig.HostIfName, measure.GetTimeLog("iface_admin_up", plugin.Stopwatch))
		} else {
			err = linuxcalls.InterfaceAdminDown(newIfConfig.HostIfName, measure.GetTimeLog("iface_admin_down", plugin.Stopwatch))
		}
		if nil != err {
			wasErr = fmt.Errorf("failed to enable/disable Linux interface: %v", err)
		}
	}

	// Configure new MAC address if set.
	if newIfConfig.PhysAddress != "" && newIfConfig.PhysAddress != oldIfConfig.PhysAddress {
		plugin.Log.WithFields(logging.Fields{"PhysAddress": newIfConfig.PhysAddress, "hostIfName": newIfConfig.HostIfName}).
			Debug("MAC address re-configured.")
		err := linuxcalls.SetInterfaceMac(newIfConfig.HostIfName, newIfConfig.PhysAddress, measure.GetTimeLog("set_iface_mac", plugin.Stopwatch))
		if err != nil {
			wasErr = fmt.Errorf("failed to assign physical address to a Linux interface: %v", err)
			plugin.Log.Error(wasErr)
		}
	}

	// IP addresses
	newAddrs, err := addrs.StrAddrsToStruct(newIfConfig.IpAddresses)
	if err != nil {
		plugin.Log.Error(err)
		wasErr = err
	}
	oldAddrs, err := addrs.StrAddrsToStruct(oldIfConfig.IpAddresses)
	if err != nil {
		plugin.Log.Error(err)
		wasErr = err
	}
	var del, add []*net.IPNet
	del, add = addrs.DiffAddr(newAddrs, oldAddrs)

	for i := range del {
		plugin.Log.WithFields(logging.Fields{"IP address": del[i], "hostIfName": newIfConfig.HostIfName}).Debug("IP address deleted.")
		err := linuxcalls.DelInterfaceIP(newIfConfig.HostIfName, del[i], measure.GetTimeLog("del_iface_ip", plugin.Stopwatch))
		if nil != err {
			wasErr = fmt.Errorf("failed to unassign IPv4 address from a Linux interface: %v", err)
			plugin.Log.Error(wasErr)
		}
	}

	for i := range add {
		plugin.Log.WithFields(logging.Fields{"IP address": add[i], "hostIfName": newIfConfig.HostIfName}).Debug("IP address added.")
		err := linuxcalls.AddInterfaceIP(plugin.Log, newIfConfig.HostIfName, add[i], measure.GetTimeLog("add_iface_ip", plugin.Stopwatch))
		if nil != err {
			wasErr = fmt.Errorf("failed to assign IPv4 address to a Linux interface: %v", err)
			plugin.Log.Error(wasErr)
		}
	}

	// MTU
	if newIfConfig.Mtu != oldIfConfig.Mtu {
		mtu := newIfConfig.Mtu
		if mtu > 0 {
			plugin.Log.WithFields(logging.Fields{"MTU": mtu, "hostIfName": newIfConfig.HostIfName}).Debug("MTU re-configured.")
			err := linuxcalls.SetInterfaceMTU(newIfConfig.HostIfName, int(mtu), measure.GetTimeLog("set_iface_mtu", plugin.Stopwatch))
			if nil != err {
				wasErr = fmt.Errorf("failed to set MTU of a Linux interface: %v", err)
				plugin.Log.Error(wasErr)
			}
		}
	}

	plugin.Log.Infof("Linux interface %v modified", newIfConfig.Name)

	return wasErr
}

// Remove VETH type interface
func (plugin *LinuxInterfaceConfigurator) deleteVethInterface(ifConfig, peerConfig *LinuxInterfaceConfig) error {
	plugin.Log.Debugf("Removing VETH interface %v ", ifConfig.config.HostIfName)
	// Veth interface removal
	if ifConfig == nil || ifConfig.config == nil || !plugin.isNamespaceAvailable(ifConfig.config.Namespace) ||
		peerConfig == nil || peerConfig.config == nil || !plugin.isNamespaceAvailable(peerConfig.config.Namespace) {
		name := "<unknown>"
		if ifConfig != nil && ifConfig.config != nil {
			name = ifConfig.config.Name
		}
		plugin.Log.WithField("ifName", name).Debug("VETH interface doesn't exist")
		return nil
	}

	// Move to the namespace with the interface.
	nsMgmtCtx := linuxcalls.NewNamespaceMgmtCtx()
	revertNs, err := plugin.switchToNamespace(nsMgmtCtx, ifConfig.config.Namespace)
	if err != nil {
		return fmt.Errorf("failed to switch network namespace: %v", err)
	}
	defer revertNs()

	err = linuxcalls.DelVethInterfacePair(ifConfig.config.HostIfName, peerConfig.config.HostIfName,
		plugin.Log, measure.GetTimeLog("del_veth_iface", plugin.Stopwatch))
	if err != nil {
		return fmt.Errorf("failed to delete VETH interface: %v", err)
	}

	// Unregister both VETH ends from the in-memory map (following triggers notifications for all subscribers).
	plugin.ifIndexes.UnregisterName(ifConfig.config.Name)
	plugin.ifIndexes.UnregisterName(peerConfig.config.Name)

	plugin.Log.Infof("Linux Interface %v removed", ifConfig.config.Name)

	return nil
}

// Un-configure TAP interface, set original name and return it to the default namespace (do not delete,
// the interface will be removed together with the peer (VPP TAP))
func (plugin *LinuxInterfaceConfigurator) deleteTapInterface(ifConfig *LinuxInterfaceConfig) error {
	plugin.Log.Debugf("Removing Linux TAP configuration %v from interface %v ", ifConfig.config.Name, ifConfig.config.HostIfName)
	if ifConfig == nil || ifConfig.config == nil {
		plugin.Log.Warn("Unable to remove linux TAP configuration: no data available")
		return nil
	}
	if !plugin.isNamespaceAvailable(ifConfig.config.Namespace) {
		plugin.Log.Warnf("Unable to remove linux TAP configuration: cannot access namespace %v", ifConfig.config.Namespace.Name)
		return nil
	}

	// Move to the namespace with the interface.
	nsMgmtCtx := linuxcalls.NewNamespaceMgmtCtx()
	revertNs, err := plugin.switchToNamespace(nsMgmtCtx, ifConfig.config.Namespace)
	if err != nil {
		return fmt.Errorf("failed to switch network namespace: %v", err)
	}
	defer revertNs()

	// Get all IP addresses currently configured on the interface. It is not enough to just remove all IP addresses
	// present in the ifConfig object, there can be default IP address which needs to be removed as well.
	var ipAddresses []*net.IPNet
	link, err := netlink.LinkList()
	for _, linuxIf := range link {
		if linuxIf.Attrs().Name == ifConfig.config.HostIfName {
			IPlist, err := netlink.AddrList(linuxIf, netlink.FAMILY_ALL)
			if err != nil {
				return err
			}
			for _, address := range IPlist {
				ipAddresses = append(ipAddresses, address.IPNet)
			}
			break
		}
	}
	// Remove all IP addresses from the TAP
	var wasErr error
	for _, ipAddress := range ipAddresses {
		if err := linuxcalls.DelInterfaceIP(ifConfig.config.HostIfName, ipAddress, nil); err != nil {
			plugin.Log.Error(err)
			wasErr = err
		}
	}

	// Move back to default namespace
	if ifConfig.config.Type == interfaces.LinuxInterfaces_AUTO_TAP {
		// Rename to its original name (if possible)
		if ifConfig.config.Tap == nil || ifConfig.config.Tap.TempIfName == "" {
			plugin.Log.Warnf("Cannot restore linux TAP %v interface state, original name (temp) is not available", ifConfig.config.HostIfName)
			ifConfig.config.Tap = &interfaces.LinuxInterfaces_Interface_Tap{
				TempIfName: ifConfig.config.HostIfName,
			}
		}
		if ifConfig.config.Tap.TempIfName == ifConfig.config.HostIfName {
			plugin.Log.Debugf("Renaming of TAP interface %v skipped, host name is the same as temporary", ifConfig.config.HostIfName)
		} else {
			if err := linuxcalls.RenameInterface(ifConfig.config.HostIfName, ifConfig.config.Tap.TempIfName,
				measure.GetTimeLog("rename-linux-interface", plugin.Stopwatch)); err != nil {

				plugin.Log.Errorf("Failed to rename TAP interface from %s to %s: %v", ifConfig.config.HostIfName,
					ifConfig.config.Tap.TempIfName, err)
				wasErr = err
			}
		}
		err = linuxcalls.SetInterfaceNamespace(nsMgmtCtx, ifConfig.config.Tap.TempIfName, &interfaces.LinuxInterfaces_Interface_Namespace{},
			plugin.Log, plugin.Stopwatch)
		if err != nil {
			return fmt.Errorf("failed to set Linux TAP interface %s to namespace %s: %v", ifConfig.config.Tap.TempIfName, "default", err)
		}
	} else {
		err = linuxcalls.SetInterfaceNamespace(nsMgmtCtx, ifConfig.config.HostIfName, &interfaces.LinuxInterfaces_Interface_Namespace{},
			plugin.Log, plugin.Stopwatch)
		if err != nil {
			return fmt.Errorf("failed to set Linux TAP interface %s to namespace %s: %v", ifConfig.config.HostIfName, "default", err)
		}
	}

	// Unregister TAP from the in-memory map
	plugin.ifIndexes.UnregisterName(ifConfig.config.Name)

	return wasErr
}

// removeObsoleteVeth deletes VETH interface which should no longer exist.
func (plugin *LinuxInterfaceConfigurator) removeObsoleteVeth(nsMgmtCtx *linuxcalls.NamespaceMgmtCtx, vethName string, hostIfName string, ns *interfaces.LinuxInterfaces_Interface_Namespace) error {
	plugin.Log.WithFields(logging.Fields{"vethName": vethName, "hostIfName": hostIfName, "ns": linuxcalls.NamespaceToStr(ns)}).
		Debug("Attempting to remove obsolete VETH")

	revertNs, err := plugin.switchToNamespace(nsMgmtCtx, ns)
	defer revertNs()
	if err != nil {
		// Already removed as namespace no longer exists.
		plugin.ifIndexes.UnregisterName(vethName)
		return nil
	}
	exists, err := linuxcalls.InterfaceExists(hostIfName, measure.GetTimeLog("iface_exists", plugin.Stopwatch))
	if err != nil {
		plugin.Log.Error(err)
		return err
	}
	if !exists {
		// already removed
		plugin.ifIndexes.UnregisterName(vethName)
		return nil
	}
	ifType, err := linuxcalls.GetInterfaceType(hostIfName, measure.GetTimeLog("get_iface_type", plugin.Stopwatch))
	if err != nil {
		plugin.Log.Error(err)
		return err
	}
	if ifType != veth {
		return fmt.Errorf("interface '%s' already exists and is not VETH", vethName)
	}
	peerName, err := linuxcalls.GetVethPeerName(hostIfName, measure.GetTimeLog("get_veth_peer", plugin.Stopwatch))
	if err != nil {
		plugin.Log.Error(err)
		return err
	}
	plugin.Log.WithFields(logging.Fields{"ifName": vethName, "peerName": peerName}).
		Debug("Removing obsolete VETH interface")
	err = linuxcalls.DelVethInterfacePair(hostIfName, peerName, plugin.Log, measure.GetTimeLog("del_veth_iface", plugin.Stopwatch))
	if err != nil {
		plugin.Log.Error(err)
		return err
	}
	plugin.ifIndexes.UnregisterName(vethName)
	return nil
}

// addVethInterfacePair creates a new VETH interface with a "clean" configuration.
func (plugin *LinuxInterfaceConfigurator) addVethInterfacePair(nsMgmtCtx *linuxcalls.NamespaceMgmtCtx,
	iface, peer *interfaces.LinuxInterfaces_Interface) error {
	err := plugin.removeObsoleteVeth(nsMgmtCtx, iface.Name, iface.HostIfName, iface.Namespace)
	if err != nil {
		return err
	}
	err = plugin.removeObsoleteVeth(nsMgmtCtx, peer.Name, peer.HostIfName, peer.Namespace)
	if err != nil {
		return err
	}
	// VETH is first created in its own cfg namespace so it has to be removed there as well.
	err = plugin.removeObsoleteVeth(nsMgmtCtx, iface.Name, iface.HostIfName, plugin.vethCfgNamespace)
	if err != nil {
		return err
	}
	err = plugin.removeObsoleteVeth(nsMgmtCtx, peer.Name, peer.HostIfName, plugin.vethCfgNamespace)
	if err != nil {
		return err
	}
	err = linuxcalls.AddVethInterfacePair(iface.HostIfName, peer.HostIfName, plugin.Log, measure.GetTimeLog("add_veth_iface", plugin.Stopwatch))
	if err != nil {
		return fmt.Errorf("failed to create new VETH: %v", err)
	}

	return nil
}

// Watcher receives events from state updater about created/removed linux interfaces and performs appropriate actions
func (plugin *LinuxInterfaceConfigurator) watchLinuxStateUpdater() {
	plugin.Log.Debugf("Linux interface state watcher started")

	for {
		linuxIf := <-plugin.ifStateChan
		linuxIfName := linuxIf.attributes.Name

		switch {
		case linuxIf.interfaceType == tapTun:
			if linuxIf.interfaceState == netlink.OperDown {
				// Find whether it is a registered tap interface and un-register it. Otherwise the change is ignored.
				for _, indexedName := range plugin.ifIndexes.GetMapping().ListNames() {
					_, ifConfigMeta, found := plugin.ifIndexes.LookupIdx(indexedName)
					if !found {
						// Should not happen
						plugin.Log.Warnf("Interface %v not found in the mapping", indexedName)
						continue
					}
					if ifConfigMeta.HostIfName == "" {
						plugin.Log.Warnf("No info about host name for %v", indexedName)
						continue
					}
					if ifConfigMeta.HostIfName == linuxIfName {
						// Registered Linux TAP interface was removed. However it should not be removed from cache
						// because the configuration still exists in the VPP
						plugin.ifIndexes.UnregisterName(linuxIfName)
					}
				}
			} else {
				// Event that TAP interface was created.
				plugin.Log.Debugf("Received data about linux TAP interface %s", linuxIfName)

				// Look for TAP which is using this interface as the other end
				for _, ifConfig := range plugin.intfByName {
					if ifConfig == nil || ifConfig.config == nil {
						plugin.Log.Warnf("Cached config for interface %v is empty", linuxIfName)
						continue
					}
					if ifConfig.config.Tap.TempIfName == linuxIfName {
						// Skip processed interfaces
						_, _, exists := plugin.ifIndexes.LookupIdx(ifConfig.config.Name)
						if exists {
							continue
						}
						// Host interface was found, configure linux TAP
						err := plugin.ConfigureLinuxInterface(ifConfig.config)
						if err != nil {
							plugin.Log.Error(err)
						}
					}
				}
			}
		default:
			plugin.Log.Debugf("Linux interface type %v state processing skipped", linuxIf.interfaceType)
		}
	}
}

// getInterfaceConfig returns cached configuration of a given interface.
func (plugin *LinuxInterfaceConfigurator) getInterfaceConfig(ifName string) *LinuxInterfaceConfig {
	config, ok := plugin.intfByName[ifName]
	if ok {
		return config
	}
	return nil
}

// addToCache adds interface configuration into the cache.
func (plugin *LinuxInterfaceConfigurator) addToCache(iface *interfaces.LinuxInterfaces_Interface, peerIface *LinuxInterfaceConfig) *LinuxInterfaceConfig {
	config := &LinuxInterfaceConfig{config: iface, peer: peerIface}
	plugin.intfByName[iface.Name] = config
	if peerIface != nil {
		peerIface.peer = config
	}
	if iface.Namespace != nil && iface.Namespace.Type == interfaces.LinuxInterfaces_Interface_Namespace_MICROSERVICE_REF_NS {
		if _, ok := plugin.intfsByMicroservice[iface.Namespace.Microservice]; ok {
			plugin.intfsByMicroservice[iface.Namespace.Microservice] = append(plugin.intfsByMicroservice[iface.Namespace.Microservice], config)
		} else {
			plugin.intfsByMicroservice[iface.Namespace.Microservice] = []*LinuxInterfaceConfig{config}
		}
	}
	plugin.Log.Debugf("Linux interface with name %v added to cache (peer: %v)",
		iface.Name, peerIface)
	return config
}

// removeFromCache removes interfaces configuration from the cache.
func (plugin *LinuxInterfaceConfigurator) removeFromCache(iface *interfaces.LinuxInterfaces_Interface) *LinuxInterfaceConfig {
	if config, ok := plugin.intfByName[iface.Name]; ok {
		if config.peer != nil {
			config.peer.peer = nil
		}
		if iface.Namespace != nil && iface.Namespace.Type == interfaces.LinuxInterfaces_Interface_Namespace_MICROSERVICE_REF_NS {
			var filtered []*LinuxInterfaceConfig
			for _, intf := range plugin.intfsByMicroservice[iface.Namespace.Microservice] {
				if intf.config.Name != iface.Name {
					filtered = append(filtered, intf)
				}
			}
			plugin.intfsByMicroservice[iface.Namespace.Microservice] = filtered
		}
		delete(plugin.intfByName, iface.Name)
		plugin.Log.Debugf("Linux interface with name %v was removed from cache", iface.Name)
		return config
	}
	return nil
}

// isNamespaceAvailable returns true if the destination namespace is available.
func (plugin *LinuxInterfaceConfigurator) isNamespaceAvailable(ns *interfaces.LinuxInterfaces_Interface_Namespace) bool {
	if ns != nil && ns.Type == interfaces.LinuxInterfaces_Interface_Namespace_MICROSERVICE_REF_NS {
		if plugin.dockerClient == nil {
			return false
		}
		_, available := plugin.microserviceByLabel[ns.Microservice]
		return available
	}
	return true
}

// convertMicroserviceNsToPidNs converts microservice-referenced namespace into the PID-referenced namespace.
func (plugin *LinuxInterfaceConfigurator) convertMicroserviceNsToPidNs(microserviceLabel string) (pidNs *interfaces.LinuxInterfaces_Interface_Namespace) {

	if microservice, ok := plugin.microserviceByLabel[microserviceLabel]; ok {
		pidNamespace := &interfaces.LinuxInterfaces_Interface_Namespace{}
		pidNamespace.Type = interfaces.LinuxInterfaces_Interface_Namespace_PID_REF_NS
		pidNamespace.Pid = uint32(microservice.pid)
		return pidNamespace
	}
	return nil
}

// switchToNamespace switches the network namespace of the current thread.
func (plugin *LinuxInterfaceConfigurator) switchToNamespace(nsMgmtCtx *linuxcalls.NamespaceMgmtCtx, ns *interfaces.LinuxInterfaces_Interface_Namespace) (revert func(), err error) {

	if ns != nil && ns.Type == interfaces.LinuxInterfaces_Interface_Namespace_MICROSERVICE_REF_NS {
		ns = plugin.convertMicroserviceNsToPidNs(ns.Microservice)
		if ns == nil {
			return func() {}, &unavailableMicroserviceErr{}
		}
	}

	// Prepare generic namespace object
	ifaceNs := linuxcalls.ToGenericNs(ns)

	return ifaceNs.SwitchNamespace(nsMgmtCtx, plugin.Log)
}

// trackMicroservices is running in the background and maintains a map of microservice labels to container info.
func (plugin *LinuxInterfaceConfigurator) trackMicroservices(ctx context.Context) {
	plugin.wg.Add(1)
	defer plugin.wg.Done()

	msCtx := &MicroserviceCtx{
		nsMgmtCtx: linuxcalls.NewNamespaceMgmtCtx(),
	}

	var clientOk bool

	timer := time.NewTimer(0)
	for {
		select {
		case <-timer.C:
			if err := plugin.dockerClient.Ping(); err != nil {
				if clientOk {
					plugin.Log.Errorf("Docker ping check failed: %v", err)
				}
				clientOk = false

				// Sleep before another retry.
				timer.Reset(dockerRetryPeriod)
				continue
			}

			if !clientOk {
				plugin.Log.Infof("Docker ping check OK")
				/*if info, err := plugin.dockerClient.Info(); err != nil {
					plugin.Log.Errorf("Retrieving docker info failed: %v", err)
					timer.Reset(dockerRetryPeriod)
					continue
				} else {
					plugin.Log.Infof("Docker connection established: server version: %v (%v %v %v)",
						info.ServerVersion, info.OperatingSystem, info.Architecture, info.KernelVersion)
				}*/
			}
			clientOk = true

			plugin.microserviceChan <- msCtx

			// Sleep before another refresh.
			timer.Reset(dockerRefreshPeriod)
		case <-plugin.ctx.Done():
			return
		}
	}
}

// HandleMicroservices handles microservice changes
func (plugin *LinuxInterfaceConfigurator) HandleMicroservices(ctx *MicroserviceCtx) {
	var err error
	var newest int64
	var containers []docker.APIContainers
	var nextCreated []string

	// First check if any microservice has terminated.
	plugin.cfgLock.Lock()
	for container := range plugin.microserviceByID {
		details, err := plugin.dockerClient.InspectContainer(container)
		if err != nil || !details.State.Running {
			plugin.processTerminatedMicroservice(ctx.nsMgmtCtx, container)
		}
	}
	plugin.cfgLock.Unlock()

	// Now check if previously created containers have transitioned to the state "running".
	for _, container := range ctx.created {
		details, err := plugin.dockerClient.InspectContainer(container)
		if err == nil {
			if details.State.Running {
				plugin.detectMicroservice(ctx.nsMgmtCtx, details)
			} else if details.State.Status == "created" {
				nextCreated = append(nextCreated, container)
			}
		} else {
			plugin.Log.Debugf("Inspect container ID %v failed: %v", container, err)
		}
	}
	ctx.created = nextCreated

	// Finally inspect newly created containers.
	listOpts := docker.ListContainersOptions{
		All:     true,
		Filters: map[string][]string{},
	}
	if ctx.since != "" {
		listOpts.Filters["since"] = []string{ctx.since}
	}

	containers, err = plugin.dockerClient.ListContainers(listOpts)
	if err != nil {
		plugin.Log.Errorf("Error listing docker containers: %v", err)
		if err, ok := err.(*docker.Error); ok &&
			(err.Status == 500 || err.Status == 404) { // 404 is required to support older docker version
			plugin.Log.Debugf("clearing since: %v", ctx.since)
			ctx.since = ""
		}
		return
	}

	for _, container := range containers {
		plugin.Log.Debugf("processing new container %v with state %v", container.ID, container.State)
		if container.State == "running" && container.Created > ctx.lastInspected {
			// Inspect the container to get the list of defined environment variables.
			details, err := plugin.dockerClient.InspectContainer(container.ID)
			if err != nil {
				plugin.Log.Debugf("Inspect container %v failed: %v", container.ID, err)
				continue
			}
			plugin.detectMicroservice(ctx.nsMgmtCtx, details)
		}
		if container.State == "created" {
			ctx.created = append(ctx.created, container.ID)
		}
		if container.Created > newest {
			newest = container.Created
			ctx.since = container.ID
		}
	}

	if newest > ctx.lastInspected {
		ctx.lastInspected = newest
	}
}

var microserviceContainerCreated = make(map[string]time.Time)

// detectMicroservice inspects container to see if it is a microservice.
// If microservice is detected, processNewMicroservice() is called to process it.
func (plugin *LinuxInterfaceConfigurator) detectMicroservice(nsMgmtCtx *linuxcalls.NamespaceMgmtCtx, container *docker.Container) {
	// Search for the microservice label.
	var label string
	for _, env := range container.Config.Env {
		if strings.HasPrefix(env, servicelabel.MicroserviceLabelEnvVar+"=") {
			label = env[len(servicelabel.MicroserviceLabelEnvVar)+1:]
			if label != "" {
				plugin.Log.Debugf("detected container as microservice: Name=%v ID=%v Created=%v State.StartedAt=%v", container.Name, container.ID, container.Created, container.State.StartedAt)
				last := microserviceContainerCreated[label]
				if last.After(container.Created) {
					plugin.Log.Debugf("ignoring older container created at %v as microservice: %+v", last, container)
					continue
				}
				microserviceContainerCreated[label] = container.Created
				plugin.processNewMicroservice(nsMgmtCtx, label, container.ID, container.State.Pid)
			}
		}
	}
}

// processNewMicroservice is triggered every time a new microservice gets freshly started. All pending interfaces are moved
// to its namespace.
func (plugin *LinuxInterfaceConfigurator) processNewMicroservice(nsMgmtCtx *linuxcalls.NamespaceMgmtCtx, microserviceLabel string, id string, pid int) {
	plugin.cfgLock.Lock()
	defer plugin.cfgLock.Unlock()

	microservice, restarted := plugin.microserviceByLabel[microserviceLabel]
	if restarted {
		plugin.processTerminatedMicroservice(nsMgmtCtx, microservice.id)
		plugin.Log.WithFields(logging.Fields{"label": microserviceLabel, "new-pid": pid, "new-id": id}).
			Warn("Microservice has been restarted")
	} else {
		plugin.Log.WithFields(logging.Fields{"label": microserviceLabel, "pid": pid, "id": id}).
			Debug("Discovered new microservice")
	}

	microservice = &Microservice{label: microserviceLabel, pid: pid, id: id}
	plugin.microserviceByLabel[microserviceLabel] = microservice
	plugin.microserviceByID[id] = microservice

	skip := make(map[string]struct{}) /* interfaces to be skipped in subsequent iterations */
	for _, iface := range plugin.intfsByMicroservice[microserviceLabel] {
		if _, toSkip := skip[iface.config.Name]; toSkip {
			continue
		}
		peer := iface.peer
		if peer != nil {
			// Peer will be processed in this iteration and skipped in the subsequent ones.
			skip[peer.config.Name] = struct{}{}
		}
		if peer != nil && plugin.isNamespaceAvailable(peer.config.Namespace) {
			// Prepare generic vet cfg namespace object
			ifaceNs := linuxcalls.ToGenericNs(plugin.vethCfgNamespace)

			// Switch to veth cfg namespace
			revertNs, err := ifaceNs.SwitchNamespace(nsMgmtCtx, plugin.Log)
			if err != nil {
				return
			}

			// VETH is ready to be created and configured
			err = plugin.addVethInterfacePair(nsMgmtCtx, iface.config, peer.config)
			if err != nil {
				plugin.Log.Error(err.Error())
				continue
			}

			if err := plugin.configureLinuxInterface(nsMgmtCtx, iface.config); err != nil {
				plugin.Log.Warnf("failed to configure VETH interface %s: %v", iface.config.Name, err)
			} else if err := plugin.configureLinuxInterface(nsMgmtCtx, peer.config); err != nil {
				plugin.Log.Warnf("failed to configure VETH interface %s: %v", peer.config.Name, err)
			}
			revertNs()
		} else {
			plugin.Log.Debugf("peer VETH %v is not ready yet, microservice: %+v", iface.config.Name, microservice)
		}
	}
}

// processTerminatedMicroservice is triggered every time a known microservice has terminated. All associated interfaces
// become obsolete and are thus removed.
func (plugin *LinuxInterfaceConfigurator) processTerminatedMicroservice(nsMgmtCtx *linuxcalls.NamespaceMgmtCtx, id string) {
	microservice, exists := plugin.microserviceByID[id]
	if !exists {
		plugin.Log.WithFields(logging.Fields{"id": id}).
			Warn("Detected removal of an unknown microservice")
		return
	}
	plugin.Log.WithFields(logging.Fields{"label": microservice.label, "pid": microservice.pid, "id": microservice.id}).
		Debug("Microservice has terminated")

	delete(plugin.microserviceByLabel, microservice.label)
	delete(plugin.microserviceByID, microservice.id)

	for _, iface := range plugin.intfsByMicroservice[microservice.label] {
		plugin.removeObsoleteVeth(nsMgmtCtx, iface.config.Name, iface.config.HostIfName, iface.config.Namespace)
		plugin.removeObsoleteVeth(nsMgmtCtx, iface.peer.config.Name, iface.peer.config.HostIfName, iface.peer.config.Namespace)
	}
}

// If hostIfName is not set, symbolic name will be used.
func (plugin *LinuxInterfaceConfigurator) handleOptionalHostIfName(config *interfaces.LinuxInterfaces_Interface) {
	if config.HostIfName == "" {
		config.HostIfName = config.Name
	}
}

// Create named namespace used for veth interface creation instead of the default one.
func (plugin *LinuxInterfaceConfigurator) prepareVethConfigNamespace() error {
	// Check if namespace exists.
	found, err := linuxcalls.NamedNetNsExists(vethConfigNamespace, plugin.Log)
	if err != nil {
		return err
	}
	// Remove namespace if exists.
	if found {
		err := linuxcalls.DeleteNamedNetNs(vethConfigNamespace, plugin.Log)
		if err != nil {
			return err
		}
	}

	_, ns, err := linuxcalls.CreateNamedNetNs(vethConfigNamespace, plugin.Log)
	if err != nil {
		return err
	}
	plugin.vethCfgNamespace, err = linuxcalls.ToInterfaceNs(ns)
	return err
}
