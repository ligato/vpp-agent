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

//go:generate protoc --proto_path=../model/interfaces --gogo_out=../model/interfaces ../model/interfaces/interfaces.proto

package ifplugin

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/ligato/vpp-agent/plugins/linux/nsplugin"
	"github.com/vishvananda/netlink"

	"bytes"

	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/ligato/cn-infra/utils/addrs"
	"github.com/ligato/vpp-agent/plugins/linux/ifplugin/ifaceidx"
	"github.com/ligato/vpp-agent/plugins/linux/ifplugin/linuxcalls"
	"github.com/ligato/vpp-agent/plugins/linux/model/interfaces"
)

// LinuxInterfaceConfig is used to cache the configuration of Linux interfaces.
type LinuxInterfaceConfig struct {
	config *interfaces.LinuxInterfaces_Interface
	peer   *LinuxInterfaceConfig
}

// LinuxInterfaceConfigurator runs in the background in its own goroutine where it watches for any changes
// in the configuration of interfaces as modelled by the proto file "model/interfaces/interfaces.proto"
// and stored in ETCD under the key "/vnf-agent/{vnf-agent}/linux/config/v1/interface".
// Updates received from the northbound API are compared with the Linux network configuration and differences
// are applied through the Netlink API.
type LinuxInterfaceConfigurator struct {
	log logging.Logger

	// In-memory mappings
	ifIndexes ifaceidx.LinuxIfIndexRW
	ifIdxSeq  uint32
	ifByName  map[string]*LinuxInterfaceConfig   // interface name -> interface configuration
	ifsByMs   map[string][]*LinuxInterfaceConfig // microservice label -> list of interfaces attached to this microservice

	// Channels
	ifMsNotif   chan *nsplugin.MicroserviceEvent
	ifStateChan chan *LinuxInterfaceStateNotification

	// Go routine management
	cfgLock sync.Mutex
	ctx     context.Context    // Context within which all goroutines are running
	cancel  context.CancelFunc // Cancel can be used to cancel all goroutines and their jobs inside of the plugin.
	wg      sync.WaitGroup     // Wait group allows to wait until all goroutines of the plugin have finished.

	// Linux namespace/calls handler
	ifHandler linuxcalls.NetlinkAPI
	nsHandler nsplugin.NamespaceAPI

	//Timer used to measure and store time
	stopwatch *measure.Stopwatch
}

// Init linux plugin and start go routines.
func (plugin *LinuxInterfaceConfigurator) Init(logging logging.PluginLogger, ifHandler linuxcalls.NetlinkAPI, nsHandler nsplugin.NamespaceAPI,
	ifIndexes ifaceidx.LinuxIfIndexRW, stateChan chan *LinuxInterfaceStateNotification, ifNotif chan *nsplugin.MicroserviceEvent, stopwatch *measure.Stopwatch) (err error) {
	// Logger
	plugin.log = logging.NewLogger("-if-conf")
	plugin.log.Debug("Initializing Linux Interface configurator")

	// Mappings
	plugin.ifIndexes = ifIndexes
	plugin.ifByName = make(map[string]*LinuxInterfaceConfig)
	plugin.ifsByMs = make(map[string][]*LinuxInterfaceConfig)
	plugin.ifIdxSeq = 1

	// Init channel
	plugin.ifStateChan = stateChan
	plugin.ifMsNotif = ifNotif

	plugin.ctx, plugin.cancel = context.WithCancel(context.Background())

	// Interface and namespace handlers
	plugin.ifHandler = ifHandler
	plugin.nsHandler = nsHandler

	// Configurator-wide stopwatch instance
	plugin.stopwatch = stopwatch

	// Start watching on microservice events
	go plugin.watchMicroservices(plugin.ctx)

	// Start watching on state updater events
	go plugin.watchLinuxStateUpdater()

	return err
}

// Close does nothing for linux interface configurator. State and notification channels are closed in linux plugin.
func (plugin *LinuxInterfaceConfigurator) Close() error {
	return nil
}

// GetLinuxInterfaceIndexes returns in-memory mapping of linux inerfaces
func (plugin *LinuxInterfaceConfigurator) GetLinuxInterfaceIndexes() ifaceidx.LinuxIfIndex {
	return plugin.ifIndexes
}

// GetInterfaceByNameCache returns cache of interface <-> config entries
func (plugin *LinuxInterfaceConfigurator) GetInterfaceByNameCache() map[string]*LinuxInterfaceConfig {
	return plugin.ifByName
}

// GetInterfaceByMsCache returns cache of microservice <-> interface list
func (plugin *LinuxInterfaceConfigurator) GetInterfaceByMsCache() map[string][]*LinuxInterfaceConfig {
	return plugin.ifsByMs
}

// ConfigureLinuxInterface reacts to a new northbound Linux interface config by creating and configuring
// the interface in the host network stack through Netlink API.
func (plugin *LinuxInterfaceConfigurator) ConfigureLinuxInterface(linuxIf *interfaces.LinuxInterfaces_Interface) error {
	plugin.cfgLock.Lock()
	defer plugin.cfgLock.Unlock()

	plugin.handleOptionalHostIfName(linuxIf)
	plugin.log.Infof("Configuring new Linux interface %v", linuxIf.HostIfName)

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
	plugin.log.Infof("Modifying Linux interface %v", newLinuxIf.HostIfName)

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
	newIfaceNs := plugin.nsHandler.IfNsToGeneric(newLinuxIf.Namespace)
	oldIfaceNs := plugin.nsHandler.IfNsToGeneric(oldLinuxIf.Namespace)
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
	if !plugin.nsHandler.IsNamespaceAvailable(newLinuxIf.Namespace) {
		plugin.log.Errorf("unable to configure linux interface %v: interface namespace is not available",
			newLinuxIf.HostIfName)
		return nil
	}

	// Validate configuration/namespace according to interface type
	if newLinuxIf.Type == interfaces.LinuxInterfaces_VETH {
		if peer == nil {
			// Interface doesn't actually exist physically.
			plugin.log.Infof("cannot configure linux interface %v: peer interface %v is not configured yet",
				newLinuxIf.HostIfName, newPeer)
			return nil
		}
		if !plugin.nsHandler.IsNamespaceAvailable(oldLinuxIf.Namespace) {
			plugin.log.Warnf("unable to modify linux interface %v: peer interface namespace is not available",
				oldLinuxIf.HostIfName)
			return nil
		}
		if !plugin.nsHandler.IsNamespaceAvailable(newLinuxIf.Namespace) {
			plugin.log.Warnf("unable to modify linux interface %v: interface namespace is not available",
				newLinuxIf.HostIfName)
			return nil
		}
	} else if newLinuxIf.Type == interfaces.LinuxInterfaces_AUTO_TAP {
		if !plugin.nsHandler.IsNamespaceAvailable(newLinuxIf.Namespace) {
			// Interface doesn't actually exist physically.
			plugin.log.WithField("ifName", newLinuxIf.Name).Debug("Linux interface is not ready to be re-configured")
			return nil
		}
	} else {
		plugin.log.Warnf("Unknown interface type %v", newLinuxIf.Type)

		return nil
	}

	// The namespace was not changed so interface can be reconfigured
	nsMgmtCtx := nsplugin.NewNamespaceMgmtCtx()

	return plugin.modifyLinuxInterface(nsMgmtCtx, oldLinuxIf, newLinuxIf)
}

// DeleteLinuxInterface reacts to a removed NB configuration of a Linux interface.
func (plugin *LinuxInterfaceConfigurator) DeleteLinuxInterface(linuxIf *interfaces.LinuxInterfaces_Interface) error {
	plugin.cfgLock.Lock()
	defer plugin.cfgLock.Unlock()

	plugin.handleOptionalHostIfName(linuxIf)
	plugin.log.Infof("Removing Linux interface %v", linuxIf.HostIfName)

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
	plugin.log.Warnf("Unknown type of interface: %v", linuxIf.Type)
	return nil
}

// Validate, create and configure VETH type linux interface
func (plugin *LinuxInterfaceConfigurator) configureVethInterface(ifConfig, peerConfig *LinuxInterfaceConfig) error {
	plugin.log.WithFields(logging.Fields{"name": ifConfig.config.Name, "hostName": ifConfig.config.HostIfName,
		"peer": ifConfig.config.Veth.PeerIfName}).Debug("Configuring new Veth interface")
	// Create VETH after both end's configs and target namespaces are available.
	if peerConfig == nil {
		plugin.log.Infof("cannot configure linux interface %v: peer interface %v is not configured yet",
			ifConfig.config.HostIfName, ifConfig.config.Veth.PeerIfName)
		return nil
	}
	if !plugin.nsHandler.IsNamespaceAvailable(ifConfig.config.Namespace) {
		plugin.log.Warnf("unable to configure linux interface %v: interface namespace is not available",
			ifConfig.config.HostIfName)
		return nil
	}
	if !plugin.nsHandler.IsNamespaceAvailable(peerConfig.config.Namespace) {
		plugin.log.Warnf("unable to configure linux interface %v: peer namespace is not available",
			ifConfig.config.HostIfName)
		return nil
	}

	nsMgmtCtx := nsplugin.NewNamespaceMgmtCtx()

	// Prepare generic veth config namespace object
	vethNs := plugin.nsHandler.IfNsToGeneric(plugin.nsHandler.GetConfigNamespace())

	// Switch to veth cfg namespace
	revertNs, err := plugin.nsHandler.SwitchNamespace(vethNs, nsMgmtCtx)
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

	plugin.log.Infof("Linux interface %v with hostIfName %v configured", ifConfig.config.Name, ifConfig.config.HostIfName)

	return nil
}

// Validate and apply linux TAP configuration to the interface. The interface is not created here, it is added
// to the default namespace when it's VPP end is configured
func (plugin *LinuxInterfaceConfigurator) configureTapInterface(ifConfig *LinuxInterfaceConfig) error {
	plugin.log.WithFields(logging.Fields{"name": ifConfig.config.Name,
		"hostName": ifConfig.config.HostIfName}).Debug("Applying new Linux TAP interface configuration")

	// Tap interfaces can be processed directly using config and also via linux interface events. This check
	// should prevent to process the same interface multiple times.
	_, _, exists := plugin.ifIndexes.LookupIdx(ifConfig.config.Name)
	if exists {
		plugin.log.Debugf("TAP interface %v already processed", ifConfig.config.Name)
		return nil
	}

	// Search default namespace for appropriate interface
	linuxIfs, err := plugin.ifHandler.GetLinkList()
	if err != nil {
		return fmt.Errorf("failed to read linux interfaces: %v", err)
	}

	nsMgmtCtx := nsplugin.NewNamespaceMgmtCtx()

	// Verify availability of namespace from configuration
	if !plugin.nsHandler.IsNamespaceAvailable(ifConfig.config.Namespace) {
		plugin.log.Errorf("unable to apply linux TAP configuration %v: destination namespace is not available",
			ifConfig.config.Name, ifConfig.config.HostIfName)
		return nil
	}

	// Check if TAP temporary name is defined
	if ifConfig.config.Tap == nil || ifConfig.config.Tap.TempIfName == "" {
		plugin.log.Debugf("Tap interface %v temporary name not defined", ifConfig.config.HostIfName)
		// In such a case, set temp name as host (look for interface named as host name)
		ifConfig.config.Tap = &interfaces.LinuxInterfaces_Interface_Tap{
			TempIfName: ifConfig.config.HostIfName,
		}
	}

	// Try to find temp interface in default namespace.
	var found bool
	for _, linuxIf := range linuxIfs {
		if ifConfig.config.Tap.TempIfName == linuxIf.Attrs().Name {
			if linuxIf.Type() == tap {
				found = true
				break
			}
			plugin.log.Debugf("Linux TAP config %v found linux interface %v, but it is not the TAP interface type",
				ifConfig.config.Name, ifConfig.config.HostIfName)
		}
	}
	if !found {
		plugin.log.Debugf("Linux TAP config %v did not found the linux interface with name %v", ifConfig.config.Name,
			ifConfig.config.Tap.TempIfName)
		return nil
	}

	return plugin.configureLinuxInterface(nsMgmtCtx, ifConfig.config)
}

// Set linux interface to proper namespace and configure attributes
func (plugin *LinuxInterfaceConfigurator) configureLinuxInterface(nsMgmtCtx *nsplugin.NamespaceMgmtCtx, ifConfig *interfaces.LinuxInterfaces_Interface) (err error) {
	if ifConfig.HostIfName == "" {
		return fmt.Errorf("host interface not specified for %v", ifConfig.Name)
	}

	// Use temporary/host name (according to type) to set interface to different namespace
	if ifConfig.Type == interfaces.LinuxInterfaces_AUTO_TAP {
		err = plugin.nsHandler.SetInterfaceNamespace(nsMgmtCtx, ifConfig.Tap.TempIfName, ifConfig.Namespace)
		if err != nil {
			return fmt.Errorf("failed to set TAP interface %s to namespace %s: %v", ifConfig.Tap.TempIfName, ifConfig.Namespace, err)
		}
	} else {
		err = plugin.nsHandler.SetInterfaceNamespace(nsMgmtCtx, ifConfig.HostIfName, ifConfig.Namespace)
		if err != nil {
			return fmt.Errorf("failed to set interface %s to namespace %s: %v", ifConfig.HostIfName, ifConfig.Namespace, err)
		}
	}
	// Continue configuring interface in its namespace.
	revertNs, err := plugin.nsHandler.SwitchToNamespace(nsMgmtCtx, ifConfig.Namespace)
	if err != nil {
		return fmt.Errorf("failed to switch network namespace: %v", err)
	}
	defer revertNs()

	// For TAP interfaces only - rename interface to the actual host name if needed
	if ifConfig.Type == interfaces.LinuxInterfaces_AUTO_TAP {
		if ifConfig.HostIfName != ifConfig.Tap.TempIfName {
			if err := plugin.ifHandler.RenameInterface(ifConfig.Tap.TempIfName, ifConfig.HostIfName); err != nil {
				plugin.log.Errorf("Failed to rename TAP interface from %s to %s: %v", ifConfig.Tap.TempIfName,
					ifConfig.HostIfName, err)
				return err
			}
		} else {
			plugin.log.Debugf("Renaming of TAP interface %v skipped, host name is the same as temporary", ifConfig.HostIfName)
		}
	}

	var wasErr error

	// Set interface up.
	if ifConfig.Enabled {
		err := plugin.ifHandler.SetInterfaceUp(ifConfig.HostIfName)
		if nil != err {
			wasErr = fmt.Errorf("failed to enable Linux interface: %v", err)
			plugin.log.Error(wasErr)
		}
	}

	// Set interface MAC address
	if ifConfig.PhysAddress != "" {
		err = plugin.ifHandler.SetInterfaceMac(ifConfig.HostIfName, ifConfig.PhysAddress)
		if err != nil {
			wasErr = fmt.Errorf("cannot assign MAC '%s': %v", ifConfig.PhysAddress, err)
			plugin.log.Error(wasErr)
		}
		plugin.log.Debugf("MAC '%s' set to interface %s", ifConfig.PhysAddress, ifConfig.HostIfName)
	}

	// Set interface IP addresses
	ipAddresses, err := addrs.StrAddrsToStruct(ifConfig.IpAddresses)
	if err != nil {
		plugin.log.Error(err)
		wasErr = err
	}
	// Get all configured interface addresses
	confAddresses, err := plugin.ifHandler.GetAddressList(ifConfig.HostIfName)
	if err != nil {
		plugin.log.Error(err)
		wasErr = err
	}
	for i, ipAddress := range ipAddresses {
		// Check link local addresses which cannot be reassigned
		if addressExists(confAddresses, ipAddresses[i]) {
			plugin.log.Debugf("Cannot assign %s to interface %s, IP already exists",
				ipAddresses[i].IP.String(), ifConfig.HostIfName)
			continue
		}
		err = plugin.ifHandler.AddInterfaceIP(ifConfig.HostIfName, ipAddresses[i])
		if err != nil {
			err = fmt.Errorf("cannot assign IP address '%s': %v", ipAddress, err)
			plugin.log.Error(err)
			wasErr = err
		} else {
			plugin.log.Debugf("IP address '%s' set to interface %s", ipAddress, ifConfig.HostIfName)
		}
	}

	if ifConfig.Mtu != 0 {
		plugin.ifHandler.SetInterfaceMTU(ifConfig.HostIfName, int(ifConfig.Mtu))
		plugin.log.Debugf("MTU %d set to interface %s", ifConfig.Mtu, ifConfig.HostIfName)
	}

	netIf, err := plugin.ifHandler.GetInterfaceByName(ifConfig.HostIfName)
	if err != nil {
		return fmt.Errorf("failed to get index of the Linux interface %s: %v", ifConfig.HostIfName, err)
	}

	// Register interface with its original name and store host name in metadata
	plugin.ifIndexes.RegisterName(ifConfig.Name, plugin.ifIdxSeq, &ifaceidx.IndexedLinuxInterface{
		Index: uint32(netIf.Index),
		Data:  ifConfig,
	})
	plugin.ifIdxSeq++
	plugin.log.WithFields(logging.Fields{"ifName": ifConfig.Name, "ifIdx": netIf.Index}).
		Info("An entry added into ifState.")

	return wasErr
}

// Update linux interface attributes in it's namespace
func (plugin *LinuxInterfaceConfigurator) modifyLinuxInterface(nsMgmtCtx *nsplugin.NamespaceMgmtCtx,
	oldIfConfig, newIfConfig *interfaces.LinuxInterfaces_Interface) error {
	// Switch to required namespace
	revertNs, err := plugin.nsHandler.SwitchToNamespace(nsMgmtCtx, oldIfConfig.Namespace)
	if err != nil {
		return fmt.Errorf("failed to switch network namespace: %v", err)
	}
	defer revertNs()

	// Verify that the interface already exists in the Linux namespace.
	_, err = plugin.ifHandler.GetInterfaceByName(oldIfConfig.HostIfName)
	if err != nil {
		plugin.log.Debugf("Host interface %v was not found", oldIfConfig.HostIfName)
		// If host does not exist, configure new setup as a new one
		return plugin.ConfigureLinuxInterface(newIfConfig)
	}

	var wasErr error

	// Set admin status.
	if newIfConfig.Enabled != oldIfConfig.Enabled {
		if newIfConfig.Enabled {
			err = plugin.ifHandler.SetInterfaceUp(newIfConfig.HostIfName)
		} else {
			err = plugin.ifHandler.SetInterfaceDown(newIfConfig.HostIfName)
		}
		if nil != err {
			wasErr = fmt.Errorf("failed to enable/disable Linux interface: %v", err)
		}
	}

	// Configure new MAC address if set.
	if newIfConfig.PhysAddress != "" && newIfConfig.PhysAddress != oldIfConfig.PhysAddress {
		plugin.log.WithFields(logging.Fields{"PhysAddress": newIfConfig.PhysAddress, "hostIfName": newIfConfig.HostIfName}).
			Debug("MAC address re-configured.")
		err := plugin.ifHandler.SetInterfaceMac(newIfConfig.HostIfName, newIfConfig.PhysAddress)
		if err != nil {
			wasErr = fmt.Errorf("failed to assign physical address to a Linux interface: %v", err)
			plugin.log.Error(wasErr)
		}
	}

	// IP addresses
	newAddrs, err := addrs.StrAddrsToStruct(newIfConfig.IpAddresses)
	if err != nil {
		plugin.log.Error(err)
		wasErr = err
	}
	oldAddrs, err := addrs.StrAddrsToStruct(oldIfConfig.IpAddresses)
	if err != nil {
		plugin.log.Error(err)
		wasErr = err
	}
	var del, add []*net.IPNet
	del, add = addrs.DiffAddr(newAddrs, oldAddrs)

	for i := range del {
		plugin.log.WithFields(logging.Fields{"IP address": del[i], "hostIfName": newIfConfig.HostIfName}).Debug("IP address deleted.")
		err := plugin.ifHandler.DelInterfaceIP(newIfConfig.HostIfName, del[i])
		if nil != err {
			wasErr = fmt.Errorf("failed to unassign IPv4 address from a Linux interface: %v", err)
			plugin.log.Error(wasErr)
		}
	}

	// Get all configured interface addresses
	confAddresses, err := plugin.ifHandler.GetAddressList(newIfConfig.HostIfName)
	if err != nil {
		plugin.log.Error(err)
		wasErr = err
	}
	for i := range add {
		plugin.log.WithFields(logging.Fields{"IP address": add[i], "hostIfName": newIfConfig.HostIfName}).Debug("IP address added.")
		// Check link local addresses which cannot be reassigned
		if addressExists(confAddresses, add[i]) {
			plugin.log.Debugf("Cannot assign %s to interface %s, IP already exists",
				add[i].IP.String(), newIfConfig.HostIfName)
			continue
		}
		err := plugin.ifHandler.AddInterfaceIP(newIfConfig.HostIfName, add[i])
		if nil != err {
			wasErr = fmt.Errorf("failed to assign IPv4 address to a Linux interface: %v", err)
			plugin.log.Error(wasErr)
		}
	}

	// MTU
	if newIfConfig.Mtu != oldIfConfig.Mtu {
		mtu := newIfConfig.Mtu
		if mtu > 0 {
			plugin.log.WithFields(logging.Fields{"MTU": mtu, "hostIfName": newIfConfig.HostIfName}).Debug("MTU re-configured.")
			err := plugin.ifHandler.SetInterfaceMTU(newIfConfig.HostIfName, int(mtu))
			if nil != err {
				wasErr = fmt.Errorf("failed to set MTU of a Linux interface: %v", err)
				plugin.log.Error(wasErr)
			}
		}
	}

	plugin.log.Infof("Linux interface %v modified", newIfConfig.Name)

	return wasErr
}

// Remove VETH type interface
func (plugin *LinuxInterfaceConfigurator) deleteVethInterface(ifConfig, peerConfig *LinuxInterfaceConfig) error {
	plugin.log.Debugf("Removing VETH interface %v ", ifConfig.config.HostIfName)
	// Veth interface removal
	if ifConfig == nil || ifConfig.config == nil || !plugin.nsHandler.IsNamespaceAvailable(ifConfig.config.Namespace) ||
		peerConfig == nil || peerConfig.config == nil || !plugin.nsHandler.IsNamespaceAvailable(peerConfig.config.Namespace) {
		name := "<unknown>"
		if ifConfig != nil && ifConfig.config != nil {
			name = ifConfig.config.Name
		}
		plugin.log.WithField("ifName", name).Debug("VETH interface doesn't exist")
		return nil
	}

	// Move to the namespace with the interface.
	nsMgmtCtx := nsplugin.NewNamespaceMgmtCtx()
	revertNs, err := plugin.nsHandler.SwitchToNamespace(nsMgmtCtx, ifConfig.config.Namespace)
	if err != nil {
		return fmt.Errorf("failed to switch network namespace: %v", err)
	}
	defer revertNs()

	err = plugin.ifHandler.DelVethInterfacePair(ifConfig.config.HostIfName, peerConfig.config.HostIfName)
	if err != nil {
		return fmt.Errorf("failed to delete VETH interface: %v", err)
	}

	// Unregister both VETH ends from the in-memory map (following triggers notifications for all subscribers).
	plugin.ifIndexes.UnregisterName(ifConfig.config.Name)
	plugin.ifIndexes.UnregisterName(peerConfig.config.Name)

	plugin.log.Infof("Linux Interface %v removed", ifConfig.config.Name)

	return nil
}

// Un-configure TAP interface, set original name and return it to the default namespace (do not delete,
// the interface will be removed together with the peer (VPP TAP))
func (plugin *LinuxInterfaceConfigurator) deleteTapInterface(ifConfig *LinuxInterfaceConfig) error {
	if ifConfig == nil || ifConfig.config == nil {
		plugin.log.Warn("Unable to remove linux TAP configuration: no data available")
		return nil
	}
	plugin.log.Debugf("Removing Linux TAP configuration %v from interface %v ", ifConfig.config.Name, ifConfig.config.HostIfName)
	if !plugin.nsHandler.IsNamespaceAvailable(ifConfig.config.Namespace) {
		plugin.log.Warnf("Unable to remove linux TAP configuration: cannot access namespace %v", ifConfig.config.Namespace.Name)
		return nil
	}

	// Move to the namespace with the interface.
	nsMgmtCtx := nsplugin.NewNamespaceMgmtCtx()
	revertNs, err := plugin.nsHandler.SwitchToNamespace(nsMgmtCtx, ifConfig.config.Namespace)
	if err != nil {
		return fmt.Errorf("failed to switch network namespace: %v", err)
	}
	defer revertNs()

	// Get all IP addresses currently configured on the interface. It is not enough to just remove all IP addresses
	// present in the ifConfig object, there can be default IP address which needs to be removed as well.
	var ipAddresses []*net.IPNet
	link, err := plugin.ifHandler.GetLinkList()
	for _, linuxIf := range link {
		if linuxIf.Attrs().Name == ifConfig.config.HostIfName {
			IPlist, err := plugin.ifHandler.GetAddressList(linuxIf.Attrs().Name)
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
		if err := plugin.ifHandler.DelInterfaceIP(ifConfig.config.HostIfName, ipAddress); err != nil {
			plugin.log.Error(err)
			wasErr = err
		}
	}

	// Move back to default namespace
	if ifConfig.config.Type == interfaces.LinuxInterfaces_AUTO_TAP {
		// Rename to its original name (if possible)
		if ifConfig.config.Tap == nil || ifConfig.config.Tap.TempIfName == "" {
			plugin.log.Warnf("Cannot restore linux TAP %v interface state, original name (temp) is not available", ifConfig.config.HostIfName)
			ifConfig.config.Tap = &interfaces.LinuxInterfaces_Interface_Tap{
				TempIfName: ifConfig.config.HostIfName,
			}
		}
		if ifConfig.config.Tap.TempIfName == ifConfig.config.HostIfName {
			plugin.log.Debugf("Renaming of TAP interface %v skipped, host name is the same as temporary", ifConfig.config.HostIfName)
		} else {
			if err := plugin.ifHandler.RenameInterface(ifConfig.config.HostIfName, ifConfig.config.Tap.TempIfName); err != nil {

				plugin.log.Errorf("Failed to rename TAP interface from %s to %s: %v", ifConfig.config.HostIfName,
					ifConfig.config.Tap.TempIfName, err)
				wasErr = err
			}
		}
		err = plugin.nsHandler.SetInterfaceNamespace(nsMgmtCtx, ifConfig.config.Tap.TempIfName, &interfaces.LinuxInterfaces_Interface_Namespace{})
		if err != nil {
			return fmt.Errorf("failed to set Linux TAP interface %s to namespace %s: %v", ifConfig.config.Tap.TempIfName, "default", err)
		}
	} else {
		err = plugin.nsHandler.SetInterfaceNamespace(nsMgmtCtx, ifConfig.config.HostIfName, &interfaces.LinuxInterfaces_Interface_Namespace{})
		if err != nil {
			return fmt.Errorf("failed to set Linux TAP interface %s to namespace %s: %v", ifConfig.config.HostIfName, "default", err)
		}
	}

	// Unregister TAP from the in-memory map
	plugin.ifIndexes.UnregisterName(ifConfig.config.Name)

	return wasErr
}

// removeObsoleteVeth deletes VETH interface which should no longer exist.
func (plugin *LinuxInterfaceConfigurator) removeObsoleteVeth(nsMgmtCtx *nsplugin.NamespaceMgmtCtx, vethName string, hostIfName string, ns *interfaces.LinuxInterfaces_Interface_Namespace) error {
	plugin.log.WithFields(logging.Fields{"vethName": vethName, "hostIfName": hostIfName, "ns": plugin.nsHandler.IfaceNsToString(ns)}).
		Debug("Attempting to remove obsolete VETH")

	revertNs, err := plugin.nsHandler.SwitchToNamespace(nsMgmtCtx, ns)
	defer revertNs()
	if err != nil {
		// Already removed as namespace no longer exists.
		plugin.ifIndexes.UnregisterName(vethName)
		return nil
	}
	exists, err := plugin.ifHandler.InterfaceExists(hostIfName)
	if err != nil {
		plugin.log.Error(err)
		return err
	}
	if !exists {
		// already removed
		plugin.ifIndexes.UnregisterName(vethName)
		return nil
	}
	ifType, err := plugin.ifHandler.GetInterfaceType(hostIfName)
	if err != nil {
		plugin.log.Error(err)
		return err
	}
	if ifType != veth {
		return fmt.Errorf("interface '%s' already exists and is not VETH", vethName)
	}
	peerName, err := plugin.ifHandler.GetVethPeerName(hostIfName)
	if err != nil {
		plugin.log.Error(err)
		return err
	}
	plugin.log.WithFields(logging.Fields{"ifName": vethName, "peerName": peerName}).
		Debug("Removing obsolete VETH interface")
	err = plugin.ifHandler.DelVethInterfacePair(hostIfName, peerName)
	if err != nil {
		plugin.log.Error(err)
		return err
	}
	plugin.ifIndexes.UnregisterName(vethName)
	return nil
}

// addVethInterfacePair creates a new VETH interface with a "clean" configuration.
func (plugin *LinuxInterfaceConfigurator) addVethInterfacePair(nsMgmtCtx *nsplugin.NamespaceMgmtCtx,
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
	err = plugin.removeObsoleteVeth(nsMgmtCtx, iface.Name, iface.HostIfName, plugin.nsHandler.GetConfigNamespace())
	if err != nil {
		return err
	}
	err = plugin.removeObsoleteVeth(nsMgmtCtx, peer.Name, peer.HostIfName, plugin.nsHandler.GetConfigNamespace())
	if err != nil {
		return err
	}
	err = plugin.ifHandler.AddVethInterfacePair(iface.HostIfName, peer.HostIfName)
	if err != nil {
		return fmt.Errorf("failed to create new VETH: %v", err)
	}

	return nil
}

// Watcher receives events from state updater about created/removed linux interfaces and performs appropriate actions
func (plugin *LinuxInterfaceConfigurator) watchLinuxStateUpdater() {
	plugin.log.Debugf("Linux interface state watcher started")

	for {
		linuxIf := <-plugin.ifStateChan
		linuxIfName := linuxIf.attributes.Name

		switch {
		case linuxIf.interfaceType == tap:
			if linuxIf.interfaceState == netlink.OperDown {
				// Find whether it is a registered tap interface and un-register it. Otherwise the change is ignored.
				for _, indexedName := range plugin.ifIndexes.GetMapping().ListNames() {
					_, ifConfigMeta, found := plugin.ifIndexes.LookupIdx(indexedName)
					if !found {
						// Should not happen
						plugin.log.Warnf("Interface %v not found in the mapping", indexedName)
						continue
					}
					if ifConfigMeta == nil {
						// Should not happen
						plugin.log.Warnf("Interface %v metadata does not exist", indexedName)
						continue
					}
					if ifConfigMeta.Data.HostIfName == "" {
						plugin.log.Warnf("No info about host name for %v", indexedName)
						continue
					}
					if ifConfigMeta.Data.HostIfName == linuxIfName {
						// Registered Linux TAP interface was removed. However it should not be removed from cache
						// because the configuration still exists in the VPP
						plugin.ifIndexes.UnregisterName(linuxIfName)
					}
				}
			} else {
				// Event that TAP interface was created.
				plugin.log.Debugf("Received data about linux TAP interface %s", linuxIfName)

				// Look for TAP which is using this interface as the other end
				for _, ifConfig := range plugin.ifByName {
					if ifConfig == nil || ifConfig.config == nil {
						plugin.log.Warnf("Cached config for interface %v is empty", linuxIfName)
						continue
					}

					if (ifConfig.config.Tap != nil && ifConfig.config.Tap.TempIfName == linuxIfName) ||
						ifConfig.config.HostIfName == linuxIfName {
						// Skip processed interfaces
						_, _, exists := plugin.ifIndexes.LookupIdx(ifConfig.config.Name)
						if exists {
							continue
						}
						// Host interface was found, configure linux TAP
						err := plugin.ConfigureLinuxInterface(ifConfig.config)
						if err != nil {
							plugin.log.Error(err)
						}
					}
				}
			}
		default:
			plugin.log.Debugf("Linux interface type %v state processing skipped", linuxIf.interfaceType)
		}
	}
}

// getInterfaceConfig returns cached configuration of a given interface.
func (plugin *LinuxInterfaceConfigurator) getInterfaceConfig(ifName string) *LinuxInterfaceConfig {
	config, ok := plugin.ifByName[ifName]
	if ok {
		return config
	}
	return nil
}

// addToCache adds interface configuration into the cache.
func (plugin *LinuxInterfaceConfigurator) addToCache(iface *interfaces.LinuxInterfaces_Interface, peerIface *LinuxInterfaceConfig) *LinuxInterfaceConfig {
	config := &LinuxInterfaceConfig{config: iface, peer: peerIface}
	plugin.ifByName[iface.Name] = config
	if peerIface != nil {
		peerIface.peer = config
	}
	if iface.Namespace != nil && iface.Namespace.Type == interfaces.LinuxInterfaces_Interface_Namespace_MICROSERVICE_REF_NS {
		if _, ok := plugin.ifsByMs[iface.Namespace.Microservice]; ok {
			plugin.ifsByMs[iface.Namespace.Microservice] = append(plugin.ifsByMs[iface.Namespace.Microservice], config)
		} else {
			plugin.ifsByMs[iface.Namespace.Microservice] = []*LinuxInterfaceConfig{config}
		}
	}
	plugin.log.Debugf("Linux interface with name %v added to cache (peer: %v)",
		iface.Name, peerIface)
	return config
}

// removeFromCache removes interfaces configuration from the cache.
func (plugin *LinuxInterfaceConfigurator) removeFromCache(iface *interfaces.LinuxInterfaces_Interface) *LinuxInterfaceConfig {
	if config, ok := plugin.ifByName[iface.Name]; ok {
		if config.peer != nil {
			config.peer.peer = nil
		}
		if iface.Namespace != nil && iface.Namespace.Type == interfaces.LinuxInterfaces_Interface_Namespace_MICROSERVICE_REF_NS {
			var filtered []*LinuxInterfaceConfig
			for _, intf := range plugin.ifsByMs[iface.Namespace.Microservice] {
				if intf.config.Name != iface.Name {
					filtered = append(filtered, intf)
				}
			}
			plugin.ifsByMs[iface.Namespace.Microservice] = filtered
		}
		delete(plugin.ifByName, iface.Name)
		plugin.log.Debugf("Linux interface with name %v was removed from cache", iface.Name)
		return config
	}
	return nil
}

// watchMicroservices handles events from namespace plugin
func (plugin *LinuxInterfaceConfigurator) watchMicroservices(ctx context.Context) {
	plugin.wg.Add(1)
	defer plugin.wg.Done()

	nsMgmtCtx := nsplugin.NewNamespaceMgmtCtx()

	for {
		select {
		case msEvent := <-plugin.ifMsNotif:
			microservice := msEvent.Microservice
			if microservice == nil {
				plugin.log.Error("Empty microservice event")
				continue
			}
			if msEvent.EventType == nsplugin.NewMicroservice {
				skip := make(map[string]struct{}) /* interfaces to be skipped in subsequent iterations */
				for _, iface := range plugin.ifsByMs[microservice.Label] {
					if _, toSkip := skip[iface.config.Name]; toSkip {
						continue
					}
					peer := iface.peer
					if peer != nil {
						// peer will be processed in this iteration and skipped in the subsequent ones.
						skip[peer.config.Name] = struct{}{}
					}
					if peer != nil && plugin.nsHandler.IsNamespaceAvailable(peer.config.Namespace) {
						// Prepare generic vet cfg namespace object
						ifaceNs := plugin.nsHandler.IfNsToGeneric(plugin.nsHandler.GetConfigNamespace())

						// Switch to veth cfg namespace
						revertNs, err := plugin.nsHandler.SwitchNamespace(ifaceNs, nsMgmtCtx)
						if err != nil {
							return
						}

						// VETH is ready to be created and configured
						err = plugin.addVethInterfacePair(nsMgmtCtx, iface.config, peer.config)
						if err != nil {
							plugin.log.Error(err.Error())
							continue
						}

						if err := plugin.configureLinuxInterface(nsMgmtCtx, iface.config); err != nil {
							plugin.log.Warnf("failed to configure VETH interface %s: %v", iface.config.Name, err)
						} else if err := plugin.configureLinuxInterface(nsMgmtCtx, peer.config); err != nil {
							plugin.log.Warnf("failed to configure VETH interface %s: %v", peer.config.Name, err)
						}
						revertNs()
					} else {
						plugin.log.Debugf("peer VETH %v is not ready yet, microservice: %+v", iface.config.Name, microservice)
					}
				}
			} else if msEvent.EventType == nsplugin.TerminatedMicroservice {
				for _, iface := range plugin.ifsByMs[microservice.Label] {
					plugin.removeObsoleteVeth(nsMgmtCtx, iface.config.Name, iface.config.HostIfName, iface.config.Namespace)
					if iface.peer != nil && iface.peer.config != nil {
						plugin.removeObsoleteVeth(nsMgmtCtx, iface.peer.config.Name, iface.peer.config.HostIfName, iface.peer.config.Namespace)
					} else {
						plugin.log.Warnf("Obsolete peer for %s not removed, no peer data", iface.config.Name)
					}
				}
			} else {
				plugin.log.Errorf("Unknown microservice event type: %s", msEvent.EventType)
			}
		case <-plugin.ctx.Done():
			return
		}
	}
}

// If hostIfName is not set, symbolic name will be used.
func (plugin *LinuxInterfaceConfigurator) handleOptionalHostIfName(config *interfaces.LinuxInterfaces_Interface) {
	if config.HostIfName == "" {
		config.HostIfName = config.Name
	}
}

func addressExists(configured []netlink.Addr, provided *net.IPNet) bool {
	for _, confAddr := range configured {
		if bytes.Equal(confAddr.IP, provided.IP) {
			return true
		}
	}
	return false
}
