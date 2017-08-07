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
	"strings"

	intf "github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/model/l2"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l3plugin/model/l3"

	"github.com/ligato/cn-infra/datasync"
	log "github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/aclplugin/model/acl"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/model/bfd"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/model/interfaces"
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
	StaticRoutes *l3.StaticRoutes
}

// NewDataResyncReq is a constructor
func NewDataResyncReq() *DataResyncReq {
	return &DataResyncReq{
		// ACLs is a list af all access lists that are expected to be in VPP after RESYNC
		ACLs: []*acl.AccessLists_Acl{},
		// Interfaces is a list af all interfaces that are expected to be in VPP after RESYNC
		Interfaces: []*interfaces.Interfaces_Interface{},
		// SingleHopBFDSession is a list af all BFD sessions that are expected to be in VPP after RESYNC
		SingleHopBFDSession: []*bfd.SingleHopBFD_Session{},
		// SingleHopBFDKey is a list af all BFD authentication keys that are expected to be in VPP after RESYNC
		SingleHopBFDKey: []*bfd.SingleHopBFD_Key{},
		// SingleHopBFDEcho is a list af all BFD echo functions that are expected to be in VPP after RESYNC
		SingleHopBFDEcho: []*bfd.SingleHopBFD_EchoFunction{},
		// BridgeDomains is a list af all BDs that are expected to be in VPP after RESYNC
		BridgeDomains: []*l2.BridgeDomains_BridgeDomain{},
		// FibTableEntries is a list af all FIBs that are expected to be in VPP after RESYNC
		FibTableEntries: []*l2.FibTableEntries_FibTableEntry{},
		// XConnects is a list af all XCons that are expected to be in VPP after RESYNC
		XConnects: []*l2.XConnectPairs_XConnectPair{},
		// StaticRoutes is a list af all Static Routes that are expected to be in VPP after RESYNC
		StaticRoutes: nil}
}

// delegates resync request to ifplugin/l2plugin/l3plugin resync requests (in this particular order)
func (plugin *Plugin) resyncConfigPropageRequest(req *DataResyncReq) error {
	log.Info("resync the VPP Configuration begin")

	plugin.ifConfigurator.Resync(req.Interfaces)
	plugin.aclConfigurator.Resync(req.ACLs)
	plugin.bfdConfigurator.ResyncAuthKey(req.SingleHopBFDKey)
	plugin.bfdConfigurator.ResyncSession(req.SingleHopBFDSession)
	plugin.bfdConfigurator.ResyncEchoFunction(req.SingleHopBFDEcho)
	plugin.bdConfigurator.Resync(req.BridgeDomains)
	plugin.fibConfigurator.Resync(req.FibTableEntries)
	plugin.xcConfigurator.Resync(req.XConnects)
	plugin.routeConfigurator.Resync(req.StaticRoutes)

	log.Debug("resync the VPP Configuration end")

	return nil
}

func resyncParseEvent(resyncEv datasync.ResyncEvent) *DataResyncReq {
	req := NewDataResyncReq()
	for key := range resyncEv.GetValues() {
		log.Debug("Received RESYNC key ", key)
	}
	for key, resyncData := range resyncEv.GetValues() {
		if strings.HasPrefix(key, acl.KeyPrefix()) {
			numAcls := appendACLInterface(resyncData, req)
			log.Debug("Received RESYNC ACL values ", numAcls)
		} else if strings.HasPrefix(key, intf.InterfaceKeyPrefix()) {
			numInterfaces := appendResyncInterface(resyncData, req)
			log.Debug("Received RESYNC interface values ", numInterfaces)
		} else if strings.HasPrefix(key, bfd.SessionKeyPrefix()) {
			numBfdSession := resyncAppendBfdSession(resyncData, req)
			log.Debug("Received RESYNC BFD Session values ", numBfdSession)
		} else if strings.HasPrefix(key, bfd.AuthKeysKeyPrefix()) {
			numBfdAuthKeys := resyncAppendBfdAuthKeys(resyncData, req)
			log.Debug("Received RESYNC BFD Auth Key values ", numBfdAuthKeys)
		} else if strings.HasPrefix(key, bfd.EchoFunctionKeyPrefix()) {
			numBfdEchos := resyncAppendBfdEcho(resyncData, req)
			log.Debug("Received RESYNC BFD Echo values ", numBfdEchos)
		} else if strings.HasPrefix(key, l2.BridgeDomainKeyPrefix()) {
			numBDs := resyncAppendBDs(resyncData, req)
			log.Debug("Received RESYNC BD values ", numBDs)
		} else if strings.HasPrefix(key, l2.XConnectKeyPrefix()) {
			numXCons := resyncAppendXCons(resyncData, req)
			log.Debug("Received RESYNC XConnects values ", numXCons)
		} else if strings.HasPrefix(key, l3.RouteKey()) {
			numL3FIBs := resyncAppendRoutes(resyncData, req)
			log.Debug("Received RESYNC L3 FIB values ", numL3FIBs)
		} else {
			log.Warn("ignoring ", resyncEv, " by VPP standard plugins")
		}
	}
	return req
}
func resyncAppendRoutes(resyncData datasync.KeyValIterator, req *DataResyncReq) int {
	num := 0
	if staticRouteData, stop := resyncData.GetNext(); !stop {
		value := &l3.StaticRoutes{}
		err := staticRouteData.GetValue(value)
		if err == nil {
			req.StaticRoutes = value
			num++
		}
	}
	return num
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
func resyncAppendFIB(fibData datasync.KeyVal, req *DataResyncReq) error {
	value := &l2.FibTableEntries_FibTableEntry{}
	err := fibData.GetValue(value)
	if err == nil {
		req.FibTableEntries = append(req.FibTableEntries, value)
	}
	return err
}

func resyncAppendBDs(resyncData datasync.KeyValIterator, req *DataResyncReq) int {
	num := 0
	for {
		if bridgeDomainData, stop := resyncData.GetNext(); stop {
			break
		} else {
			key := bridgeDomainData.GetKey()
			fib, _, fibMac := l2.ParseFibKey(key)
			if fib {
				log.Debugf("Received RESYNC L2 FIB entry (%s)", fibMac)
				err := resyncAppendFIB(bridgeDomainData, req)
				if err == nil {
					num++
				}
			} else {
				value := &l2.BridgeDomains_BridgeDomain{}
				err := bridgeDomainData.GetValue(value)
				if err == nil {
					req.BridgeDomains = append(req.BridgeDomains, value)
					num++
				}
			}
		}
	}
	return num
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
			value := &intf.Interfaces_Interface{}
			err := interfaceData.GetValue(value)
			if err == nil {
				req.Interfaces = append(req.Interfaces, value)
				num++
			}
		}
	}
	return num
}

// put here all registration for above channel select (it ensures proper order during initialization
func (plugin *Plugin) subscribeWatcher() (err error) {
	log.Debug("subscribeWatcher begin")
	plugin.swIfIndexes.WatchNameToIdx(PluginID, plugin.ifIdxWatchCh)
	log.Debug("swIfIndexes watch registration finished")
	plugin.bdIndexes.WatchNameToIdx(PluginID, plugin.bdIdxWatchCh)
	log.Debug("bdIndexes watch registration finished")
	if plugin.linuxIfIndexes != nil {
		plugin.linuxIfIndexes.Watch(PluginID, nametoidx.ToChan(plugin.linuxIfIdxWatchCh))
		log.Debug("linuxIfIndexes watch registration finished")
	}

	plugin.watchConfigReg, err = plugin.Transport.
		WatchData("Config VPP default plug:IF/L2/L3", plugin.changeChan, plugin.resyncConfigChan,
			acl.KeyPrefix(),
			intf.InterfaceKeyPrefix(),
			bfd.SessionKeyPrefix(),
			bfd.AuthKeysKeyPrefix(),
			bfd.EchoFunctionKeyPrefix(),
			l2.BridgeDomainKeyPrefix(),
			l2.XConnectKeyPrefix(),
			l3.RouteKey())
	if err != nil {
		return err
	}

	plugin.watchStatusReg, err = plugin.Transport.
		WatchData("Status VPP default plug:IF/L2/L3", nil, plugin.resyncStatusChan,
			intf.InterfaceStateKeyPrefix(), l2.BridgeDomainStateKeyPrefix())
	if err != nil {
		return err
	}

	log.Debug("data Transport watch finished")

	return nil
}

func (plugin *Plugin) changePropagateRequest(dataChng datasync.ChangeEvent, callback func(error)) error {
	key := dataChng.GetKey()

	// Skip potential changes on error keys
	if strings.HasPrefix(key, interfaces.InterfaceErrorPrefix()) || strings.HasPrefix(key, l2.BridgeDomainErrorPrefix()) {
		return nil
	}

	log.Debug("Start processing change for key: ", key)
	if strings.HasPrefix(key, acl.KeyPrefix()) {
		var value, prevValue acl.AccessLists_Acl
		if err := dataChng.GetValue(&value); err != nil {
			return err
		}
		if diff, err := dataChng.GetPrevValue(&prevValue); err == nil {
			if err := plugin.dataChangeACL(diff, &value, &prevValue, dataChng.GetChangeType()); err != nil {
				return err
			}
		} else {
			return err
		}
	} else if strings.HasPrefix(key, intf.InterfaceKeyPrefix()) {
		var value, prevValue intf.Interfaces_Interface
		if err := dataChng.GetValue(&value); err != nil {
			return err
		}
		if diff, err := dataChng.GetPrevValue(&prevValue); err == nil {
			if err := plugin.dataChangeIface(diff, &value, &prevValue, dataChng.GetChangeType()); err != nil {
				return err
			}
		} else {
			return err
		}
	} else if strings.HasPrefix(key, bfd.SessionKeyPrefix()) {
		var value, prevValue bfd.SingleHopBFD_Session
		if err := dataChng.GetValue(&value); err != nil {
			return err
		}
		if diff, err := dataChng.GetPrevValue(&prevValue); err == nil {
			if err := plugin.dataChangeBfdSession(diff, &value, &prevValue, dataChng.GetChangeType()); err != nil {
				return err
			}
		} else {
			return err
		}
	} else if strings.HasPrefix(key, bfd.AuthKeysKeyPrefix()) {
		var value, prevValue bfd.SingleHopBFD_Key
		if err := dataChng.GetValue(&value); err != nil {
			return err
		}
		if diff, err := dataChng.GetPrevValue(&prevValue); err == nil {
			if err := plugin.dataChangeBfdKey(diff, &value, &prevValue, dataChng.GetChangeType()); err != nil {
				return err
			}
		} else {
			return err
		}
	} else if strings.HasPrefix(key, bfd.EchoFunctionKeyPrefix()) {
		var value, prevValue bfd.SingleHopBFD_EchoFunction
		if err := dataChng.GetValue(&value); err != nil {
			return err
		}
		if diff, err := dataChng.GetPrevValue(&prevValue); err == nil {
			if err := plugin.dataChangeBfdEchoFunction(diff, &value, &prevValue, dataChng.GetChangeType()); err != nil {
				return err
			}
		} else {
			return err
		}
	} else if strings.HasPrefix(key, l2.BridgeDomainKeyPrefix()) {
		fib, _, _ := l2.ParseFibKey(key)
		if fib {
			// L2 FIB entry
			var value, prevValue l2.FibTableEntries_FibTableEntry
			if err := dataChng.GetValue(&value); err != nil {
				return err
			}
			if diff, err := dataChng.GetPrevValue(&prevValue); err == nil {
				if err := plugin.dataChangeFIB(diff, &value, &prevValue, dataChng.GetChangeType(), callback); err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			// Bridge domain
			var value, prevValue l2.BridgeDomains_BridgeDomain
			if err := dataChng.GetValue(&value); err != nil {
				return err
			}
			if diff, err := dataChng.GetPrevValue(&prevValue); err == nil {
				if err := plugin.dataChangeBD(diff, &value, &prevValue, dataChng.GetChangeType()); err != nil {
					return err
				}
			} else {
				return err
			}
		}
	} else if strings.HasPrefix(key, l2.XConnectKeyPrefix()) {
		var value, prevValue l2.XConnectPairs_XConnectPair
		if err := dataChng.GetValue(&value); err != nil {
			return err
		}
		if diff, err := dataChng.GetPrevValue(&prevValue); err == nil {
			if err := plugin.dataChangeXCon(diff, &value, &prevValue, dataChng.GetChangeType()); err != nil {
				return err
			}
		} else {
			return err
		}
	} else if strings.HasPrefix(key, l3.RouteKey()) {
		var value, prevValue l3.StaticRoutes
		if err := dataChng.GetValue(&value); err != nil {
			return err
		}
		if diff, err := dataChng.GetPrevValue(&prevValue); err == nil {
			if err := plugin.dataChangeStaticRoute(diff, &value, &prevValue, dataChng.GetChangeType()); err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		log.Warn("ignoring change ", dataChng, " by VPP standard plugins") //NOT ERROR!
	}
	return nil
}
