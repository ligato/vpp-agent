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

package defaultplugins

import (
	"strconv"
	"strings"

	"time"

	"fmt"
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/aclplugin/model/acl"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/model/bfd"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/model/l2"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l3plugin/model/l3"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l4plugin/model/l4"
)

// DataResyncReq is used to transfer expected configuration of the VPP to the plugins
type DataResyncReq struct {
	// ACLs is a list af all access lists that are expected to be in VPP after RESYNC
	ACLs []*acl.AccessLists_Acl
	// Interfaces is a list af all interfaces that are expected to be in VPP after RESYNC
	Interfaces []*interfaces.Interfaces_Interface
	// SingleHopBFDSession is a list af all BFD sessions that are expected to be in VPP after RESYNC
	SingleHopBFDSession []*bfd.SingleHopBFD_Session
	// SingleHopBFDKey is a list af all BFD authentication keys that are expected to be in VPP after RESYNC
	SingleHopBFDKey []*bfd.SingleHopBFD_Key
	// SingleHopBFDEcho is a list af all BFD echo functions that are expected to be in VPP after RESYNC
	SingleHopBFDEcho []*bfd.SingleHopBFD_EchoFunction
	// BridgeDomains is a list af all BDs that are expected to be in VPP after RESYNC
	BridgeDomains []*l2.BridgeDomains_BridgeDomain
	// FibTableEntries is a list af all FIBs that are expected to be in VPP after RESYNC
	FibTableEntries []*l2.FibTableEntries_FibTableEntry
	// XConnects is a list af all XCons that are expected to be in VPP after RESYNC
	XConnects []*l2.XConnectPairs_XConnectPair
	// StaticRoutes is a list af all Static Routes that are expected to be in VPP after RESYNC
	StaticRoutes []*l3.StaticRoutes_Route
	// ArpEntries is a list af all ARP entries that are expected to be in VPP after RESYNC
	ArpEntries []*l3.ArpTable_ArpTableEntry
	// L4Features is a bool flag that is expected to be set in VPP after RESYNC
	L4Features *l4.L4Features
	// AppNamespaces is a list af all App Namespaces that are expected to be in VPP after RESYNC
	AppNamespaces []*l4.AppNamespaces_AppNamespace
}

// NewDataResyncReq is a constructor
func NewDataResyncReq() *DataResyncReq {
	return &DataResyncReq{
		ACLs:                []*acl.AccessLists_Acl{},
		Interfaces:          []*interfaces.Interfaces_Interface{},
		SingleHopBFDSession: []*bfd.SingleHopBFD_Session{},
		SingleHopBFDKey:     []*bfd.SingleHopBFD_Key{},
		SingleHopBFDEcho:    []*bfd.SingleHopBFD_EchoFunction{},
		BridgeDomains:       []*l2.BridgeDomains_BridgeDomain{},
		FibTableEntries:     []*l2.FibTableEntries_FibTableEntry{},
		XConnects:           []*l2.XConnectPairs_XConnectPair{},
		StaticRoutes:        []*l3.StaticRoutes_Route{},
		ArpEntries:          []*l3.ArpTable_ArpTableEntry{},
		L4Features:          &l4.L4Features{},
		AppNamespaces:       []*l4.AppNamespaces_AppNamespace{},
	}
}

// delegates full resync request
func (plugin *Plugin) resyncConfigPropageFullRequest(req *DataResyncReq) error {
	plugin.Log.Info("resync the VPP Configuration begin")
	startTime := time.Now()
	defer func() {
		vppResync := time.Since(startTime)
		plugin.Log.WithField("durationInNs", vppResync.Nanoseconds()).Infof("resync the VPP Configuration end in %v", vppResync)
	}()

	return plugin.resyncConfig(req)
}

// delegates optimize-cold-stasrt resync request
func (plugin *Plugin) resyncConfigPropageOptimizedRequest(req *DataResyncReq) error {
	plugin.Log.Info("resync the VPP Configuration begin")
	startTime := time.Now()
	defer func() {
		vppResync := time.Since(startTime)
		plugin.Log.WithField("durationInNs", vppResync.Nanoseconds()).Infof("resync the VPP Configuration end in %v", vppResync)
	}()

	// If the strategy is optimize-cold-start, run interface configurator resync which provides the information
	// whether resync should continue or be terminated
	stopResync := plugin.ifConfigurator.VerifyVPPConfigPresence(req.Interfaces)
	if stopResync {
		// terminate the resync operation
		return nil
	}
	// continue resync normally
	return plugin.resyncConfig(req)
}

// delegates resync request to ifplugin/l2plugin/l3plugin resync requests (in this particular order)
func (plugin *Plugin) resyncConfig(req *DataResyncReq) error {
	// store all resync errors
	var resyncErrs []error

	if err := plugin.ifConfigurator.Resync(req.Interfaces); err != nil {
		resyncErrs = append(resyncErrs, err)
	}
	if err := plugin.aclConfigurator.Resync(req.ACLs, plugin.Log); err != nil {
		resyncErrs = append(resyncErrs, err)
	}
	if err := plugin.bfdConfigurator.ResyncAuthKey(req.SingleHopBFDKey); err != nil {
		resyncErrs = append(resyncErrs, err)
	}
	if err := plugin.bfdConfigurator.ResyncSession(req.SingleHopBFDSession); err != nil {
		resyncErrs = append(resyncErrs, err)
	}
	if err := plugin.bfdConfigurator.ResyncEchoFunction(req.SingleHopBFDEcho); err != nil {
		resyncErrs = append(resyncErrs, err)
	}
	if err := plugin.bdConfigurator.Resync(req.BridgeDomains); err != nil {
		resyncErrs = append(resyncErrs, err)
	}
	if err := plugin.fibConfigurator.Resync(req.FibTableEntries); err != nil {
		resyncErrs = append(resyncErrs, err)
	}
	if err := plugin.xcConfigurator.Resync(req.XConnects); err != nil {
		resyncErrs = append(resyncErrs, err)
	}
	if err := plugin.routeConfigurator.Resync(req.StaticRoutes); err != nil {
		resyncErrs = append(resyncErrs, err)
	}
	if err := plugin.arpConfigurator.Resync(req.ArpEntries); err != nil {
		resyncErrs = append(resyncErrs, err)
	}
	if err := plugin.l4Configurator.ResyncFeatures(req.L4Features); err != nil {
		resyncErrs = append(resyncErrs, err)
	}
	if err := plugin.l4Configurator.ResyncAppNs(req.AppNamespaces); err != nil {
		resyncErrs = append(resyncErrs, err)
	}

	// log errors if any
	if len(resyncErrs) == 0 {
		return nil
	}
	for _, err := range resyncErrs {
		plugin.Log.Error(err)
	}
	return fmt.Errorf("%v errors occured during defaultplugins resync", len(resyncErrs))
}

func (plugin *Plugin) resyncParseEvent(resyncEv datasync.ResyncEvent) *DataResyncReq {
	req := NewDataResyncReq()
	for key := range resyncEv.GetValues() {
		plugin.Log.Debug("Received RESYNC key ", key)
	}
	for key, resyncData := range resyncEv.GetValues() {
		if strings.HasPrefix(key, acl.KeyPrefix()) {
			numAcls := appendACLInterface(resyncData, req)
			plugin.Log.Debug("Received RESYNC ACL values ", numAcls)
		} else if strings.HasPrefix(key, interfaces.InterfaceKeyPrefix()) {
			numInterfaces := appendResyncInterface(resyncData, req)
			plugin.Log.Debug("Received RESYNC interface values ", numInterfaces)
		} else if strings.HasPrefix(key, bfd.SessionKeyPrefix()) {
			numBfdSession := resyncAppendBfdSession(resyncData, req)
			plugin.Log.Debug("Received RESYNC BFD Session values ", numBfdSession)
		} else if strings.HasPrefix(key, bfd.AuthKeysKeyPrefix()) {
			numBfdAuthKeys := resyncAppendBfdAuthKeys(resyncData, req)
			plugin.Log.Debug("Received RESYNC BFD Auth Key values ", numBfdAuthKeys)
		} else if strings.HasPrefix(key, bfd.EchoFunctionKeyPrefix()) {
			numBfdEchos := resyncAppendBfdEcho(resyncData, req)
			plugin.Log.Debug("Received RESYNC BFD Echo values ", numBfdEchos)
		} else if strings.HasPrefix(key, l2.BridgeDomainKeyPrefix()) {
			numBDs, numL2FIBs := resyncAppendBDs(resyncData, req)
			plugin.Log.Debug("Received RESYNC BD values ", numBDs)
			plugin.Log.Debug("Received RESYNC L2 FIB values ", numL2FIBs)
		} else if strings.HasPrefix(key, l2.XConnectKeyPrefix()) {
			numXCons := resyncAppendXCons(resyncData, req)
			plugin.Log.Debug("Received RESYNC XConnects values ", numXCons)
		} else if strings.HasPrefix(key, l3.VrfKeyPrefix()) {
			numVRFs, numL3FIBs := resyncAppendVRFs(resyncData, req, plugin.Log)
			plugin.Log.Debug("Received RESYNC VRF values ", numVRFs)
			plugin.Log.Debug("Received RESYNC L3 FIB values ", numL3FIBs)
		} else if strings.HasPrefix(key, l3.ArpKeyPrefix()) {
			numARPs := resyncAppendARPs(resyncData, req, plugin.Log)
			plugin.Log.Debug("Received RESYNC ARP values ", numARPs)
		} else if strings.HasPrefix(key, l4.FeatureKeyPrefix()) {
			resyncFeatures(resyncData, req)
			plugin.Log.Debug("Received RESYNC AppNs feature flag")
		} else if strings.HasPrefix(key, l4.AppNamespacesKeyPrefix()) {
			numAppNs := resyncAppendAppNs(resyncData, req)
			plugin.Log.Debug("Received RESYNC AppNamespace values ", numAppNs)
		} else {
			plugin.Log.Warn("ignoring ", resyncEv, " by VPP standard plugins")
		}
	}
	return req
}

func resyncAppendARPs(resyncData datasync.KeyValIterator, req *DataResyncReq, log logging.Logger) int {
	num := 0
	for {
		if arpData, stop := resyncData.GetNext(); stop {
			break
		} else {
			entry := &l3.ArpTable_ArpTableEntry{}
			if err := arpData.GetValue(entry); err == nil {
				req.ArpEntries = append(req.ArpEntries, entry)
				num++
			}
		}
	}
	return num
}

func resyncAppendL3FIB(fibData datasync.KeyVal, vrfIndex string, req *DataResyncReq, log logging.Logger) error {
	route := &l3.StaticRoutes_Route{}
	err := fibData.GetValue(route)
	if err != nil {
		return err
	}
	// Ensure every route has the corresponding VRF index
	intVrfKeyIndex, err := strconv.Atoi(vrfIndex)
	if err != nil {
		return err
	}
	if vrfIndex != strconv.Itoa(int(route.VrfId)) {
		log.Warnf("Resync: VRF index from key (%v) and from config (%v) does not match, using value from the key",
			intVrfKeyIndex, route.VrfId)
		route.VrfId = uint32(intVrfKeyIndex)
	}

	req.StaticRoutes = append(req.StaticRoutes, route)
	return nil
}

func resyncAppendVRFs(resyncData datasync.KeyValIterator, req *DataResyncReq, log logging.Logger) (numVRFs, numL3FIBs int) {
	numVRFs = 0
	numL3FIBs = 0
	for {
		if vrfData, stop := resyncData.GetNext(); stop {
			break
		} else {
			key := vrfData.GetKey()
			fib, vrfIndex, _, _, _ := l3.ParseRouteKey(key)
			if fib {
				err := resyncAppendL3FIB(vrfData, vrfIndex, req, log)
				if err == nil {
					numL3FIBs++
				}
			} else {
				log.Warn("VRF RESYNC is not implemented")
			}
		}
	}
	return numVRFs, numL3FIBs
}

func resyncAppendXCons(resyncData datasync.KeyValIterator, req *DataResyncReq) int {
	num := 0
	for {
		if xConnectData, stop := resyncData.GetNext(); stop {
			break
		} else {
			value := &l2.XConnectPairs_XConnectPair{}
			err := xConnectData.GetValue(value)
			if err == nil {
				req.XConnects = append(req.XConnects, value)
				num++
			}
		}
	}
	return num
}
func resyncAppendL2FIB(fibData datasync.KeyVal, req *DataResyncReq) error {
	value := &l2.FibTableEntries_FibTableEntry{}
	err := fibData.GetValue(value)
	if err == nil {
		req.FibTableEntries = append(req.FibTableEntries, value)
	}
	return err
}

func resyncAppendBDs(resyncData datasync.KeyValIterator, req *DataResyncReq) (numBDs, numL2FIBs int) {
	numBDs = 0
	numL2FIBs = 0
	for {
		if bridgeDomainData, stop := resyncData.GetNext(); stop {
			break
		} else {
			key := bridgeDomainData.GetKey()
			fib, _, _ := l2.ParseFibKey(key)
			if fib {
				err := resyncAppendL2FIB(bridgeDomainData, req)
				if err == nil {
					numL2FIBs++
				}
			} else {
				value := &l2.BridgeDomains_BridgeDomain{}
				err := bridgeDomainData.GetValue(value)
				if err == nil {
					req.BridgeDomains = append(req.BridgeDomains, value)
					numBDs++
				}
			}
		}
	}
	return numBDs, numL2FIBs
}

func resyncAppendBfdEcho(resyncData datasync.KeyValIterator, req *DataResyncReq) int {
	value := &bfd.SingleHopBFD_EchoFunction{}
	num := 0
	for {
		if bfdData, stop := resyncData.GetNext(); stop {
			break
		} else {
			err := bfdData.GetValue(value)
			if err == nil {
				req.SingleHopBFDEcho = append(req.SingleHopBFDEcho, value)
				num++
			}
		}
	}
	return num
}

func resyncAppendBfdAuthKeys(resyncData datasync.KeyValIterator, req *DataResyncReq) int {
	value := &bfd.SingleHopBFD_Key{}
	num := 0
	for {
		if bfdData, stop := resyncData.GetNext(); stop {
			break
		} else {
			err := bfdData.GetValue(value)
			if err == nil {
				req.SingleHopBFDKey = append(req.SingleHopBFDKey, value)
				num++
			}
		}
	}
	return num
}

func resyncAppendBfdSession(resyncData datasync.KeyValIterator, req *DataResyncReq) int {
	value := &bfd.SingleHopBFD_Session{}
	num := 0
	for {
		if bfdData, stop := resyncData.GetNext(); stop {
			break
		} else {
			err := bfdData.GetValue(value)
			if err == nil {
				req.SingleHopBFDSession = append(req.SingleHopBFDSession, value)
				num++
			}
		}
	}
	return num
}

func appendACLInterface(resyncData datasync.KeyValIterator, req *DataResyncReq) int {
	num := 0
	for {
		if data, stop := resyncData.GetNext(); stop {
			break
		} else {
			value := &acl.AccessLists_Acl{}
			err := data.GetValue(value)
			if err == nil {
				req.ACLs = append(req.ACLs, value)
				num++
			}
		}
	}
	return num
}

func appendResyncInterface(resyncData datasync.KeyValIterator, req *DataResyncReq) int {
	num := 0
	for {
		if interfaceData, stop := resyncData.GetNext(); stop {
			break
		} else {
			value := &interfaces.Interfaces_Interface{}
			err := interfaceData.GetValue(value)
			if err == nil {
				req.Interfaces = append(req.Interfaces, value)
				num++
			}
		}
	}
	return num
}

func resyncFeatures(resyncData datasync.KeyValIterator, req *DataResyncReq) {
	for {
		appResyncData, stop := resyncData.GetNext()
		if stop {
			break
		}
		value := &l4.L4Features{}
		err := appResyncData.GetValue(value)
		if err == nil {
			req.L4Features = value
		}
	}
}

func resyncAppendAppNs(resyncData datasync.KeyValIterator, req *DataResyncReq) int {
	num := 0
	for {
		if appResyncData, stop := resyncData.GetNext(); stop {
			break
		} else {
			value := &l4.AppNamespaces_AppNamespace{}
			err := appResyncData.GetValue(value)
			if err == nil {
				req.AppNamespaces = append(req.AppNamespaces, value)
				num++
			}
		}
	}
	return num
}

// put here all registration for above channel select (it ensures proper order during initialization
func (plugin *Plugin) subscribeWatcher() (err error) {
	plugin.Log.Debug("subscribeWatcher begin")
	plugin.swIfIndexes.WatchNameToIdx(plugin.PluginName, plugin.ifIdxWatchCh)
	plugin.Log.Debug("swIfIndexes watch registration finished")
	plugin.bdIndexes.WatchNameToIdx(plugin.PluginName, plugin.bdIdxWatchCh)
	plugin.Log.Debug("bdIndexes watch registration finished")
	if plugin.linuxIfIndexes != nil {
		plugin.linuxIfIndexes.WatchNameToIdx(plugin.PluginName, plugin.linuxIfIdxWatchCh)
		plugin.Log.Debug("linuxIfIndexes watch registration finished")
	}

	plugin.watchConfigReg, err = plugin.Watch.
		Watch("Config VPP default plug:IF/L2/L3", plugin.changeChan, plugin.resyncConfigChan,
			acl.KeyPrefix(),
			interfaces.InterfaceKeyPrefix(),
			bfd.SessionKeyPrefix(),
			bfd.AuthKeysKeyPrefix(),
			bfd.EchoFunctionKeyPrefix(),
			l2.BridgeDomainKeyPrefix(),
			l2.XConnectKeyPrefix(),
			l3.VrfKeyPrefix(),
			l3.ArpKeyPrefix(),
			l4.FeatureKeyPrefix(),
			l4.AppNamespacesKeyPrefix(),
		)
	if err != nil {
		return err
	}

	plugin.watchStatusReg, err = plugin.Watch.
		Watch("Status VPP default plug:IF/L2/L3", nil, plugin.resyncStatusChan,
			interfaces.InterfaceStateKeyPrefix(), l2.BridgeDomainStateKeyPrefix())
	if err != nil {
		return err
	}

	plugin.Log.Debug("data Transport watch finished")

	return nil
}
