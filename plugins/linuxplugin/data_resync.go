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

package linuxplugin

import (
	"fmt"
	"strings"

	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/vpp-agent/plugins/linuxplugin/common/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/linuxplugin/common/model/l3"
)

// DataResyncReq is used to transfer expected configuration of the Linux network stack to the plugins.
type DataResyncReq struct {
	// Interfaces is a list af all interfaces that are expected to be in Linux after RESYNC.
	Interfaces []*interfaces.LinuxInterfaces_Interface
	// ARPs is a list af all arp entries that are expected to be in Linux after RESYNC.
	ARPs []*l3.LinuxStaticArpEntries_ArpEntry
	// Routes is a list af all routes that are expected to be in Linux after RESYNC.
	Routes []*l3.LinuxStaticRoutes_Route
}

// NewDataResyncReq is a constructor of object requirements which are expected to be re-synced.
func NewDataResyncReq() *DataResyncReq {
	return &DataResyncReq{
		// Interfaces is a list af all interfaces that are expected to be in Linux after RESYNC.
		Interfaces: []*interfaces.LinuxInterfaces_Interface{},
		// ARPs is a list af all arp entries that are expected to be in Linux after RESYNC.
		ARPs: []*l3.LinuxStaticArpEntries_ArpEntry{},
		// Routes is a list af all routes that are expected to be in Linux after RESYNC.
		Routes: []*l3.LinuxStaticRoutes_Route{},
	}
}

// DataResync delegates resync request linuxplugin configurators.
func (plugin *Plugin) resyncPropageRequest(req *DataResyncReq) error {
	plugin.Log.Info("resync the Linux Configuration")
	// store all resync errors
	var resyncErrs []error

	if errs := plugin.ifConfigurator.Resync(req.Interfaces); errs != nil {
		resyncErrs = append(resyncErrs, errs...)
	}

	if errs := plugin.arpConfigurator.Resync(req.ARPs); errs != nil {
		resyncErrs = append(resyncErrs, errs...)
	}

	if errs := plugin.routeConfigurator.Resync(req.Routes); errs != nil {
		resyncErrs = append(resyncErrs, errs...)
	}

	// log errors if any
	if len(resyncErrs) == 0 {
		return nil
	}
	for _, err := range resyncErrs {
		plugin.Log.Error(err)
	}

	return fmt.Errorf("%v errors occured during linuxplugin resync", len(resyncErrs))
}

func resyncParseEvent(resyncEv datasync.ResyncEvent, log logging.Logger) *DataResyncReq {
	req := NewDataResyncReq()
	for key := range resyncEv.GetValues() {
		log.Debug("Received RESYNC key ", key)
	}
	for key, resyncData := range resyncEv.GetValues() {
		if strings.HasPrefix(key, interfaces.InterfaceKeyPrefix()) {
			numInterfaces := resyncAppendInterface(resyncData, req)
			log.Debug("Received RESYNC interface values ", numInterfaces)
		} else if strings.HasPrefix(key, l3.StaticArpKeyPrefix()) {
			numARPs := resyncAppendARPs(resyncData, req)
			log.Debug("Received RESYNC ARP entry values ", numARPs)
		} else if strings.HasPrefix(key, l3.StaticRouteKeyPrefix()) {
			numRoutes := resyncAppendRoutes(resyncData, req)
			log.Debug("Received RESYNC route values ", numRoutes)
		} else {
			log.Warn("ignoring ", resyncEv)
		}
	}
	return req
}

func resyncAppendInterface(resyncData datasync.KeyValIterator, req *DataResyncReq) int {
	num := 0
	for {
		if interfaceData, stop := resyncData.GetNext(); stop {
			break
		} else {
			value := &interfaces.LinuxInterfaces_Interface{}
			err := interfaceData.GetValue(value)
			if err == nil {
				req.Interfaces = append(req.Interfaces, value)
				num++
			}
		}
	}
	return num
}

func resyncAppendARPs(resyncData datasync.KeyValIterator, req *DataResyncReq) int {
	num := 0
	for {
		if arpData, stop := resyncData.GetNext(); stop {
			break
		} else {
			value := &l3.LinuxStaticArpEntries_ArpEntry{}
			err := arpData.GetValue(value)
			if err == nil {
				req.ARPs = append(req.ARPs, value)
				num++
			}
		}
	}
	return num
}

func resyncAppendRoutes(resyncData datasync.KeyValIterator, req *DataResyncReq) int {
	num := 0
	for {
		if routeData, stop := resyncData.GetNext(); stop {
			break
		} else {
			value := &l3.LinuxStaticRoutes_Route{}
			err := routeData.GetValue(value)
			if err == nil {
				req.Routes = append(req.Routes, value)
				num++
			}
		}
	}
	return num
}

func (plugin *Plugin) subscribeWatcher() (err error) {
	plugin.Log.Debug("subscribeWatcher begin")
	plugin.ifIndexes.WatchNameToIdx(plugin.PluginName, plugin.ifIndexesWatchChan)
	plugin.watchDataReg, err = plugin.Watcher.
		Watch("linuxplugin", plugin.changeChan, plugin.resyncChan,
			interfaces.InterfaceKeyPrefix(),
			l3.StaticArpKeyPrefix(),
			l3.StaticRouteKeyPrefix())
	if err != nil {
		return err
	}

	plugin.Log.Debug("data watcher watch finished")

	return nil
}

func (plugin *Plugin) changePropagateRequest(dataChng datasync.ChangeEvent) error {
	var err error
	key := dataChng.GetKey()
	plugin.Log.Debug("Start processing change for key: ", key)

	if strings.HasPrefix(key, interfaces.InterfaceKeyPrefix()) {
		var value, prevValue interfaces.LinuxInterfaces_Interface
		err = dataChng.GetValue(&value)
		if err != nil {
			return err
		}
		var diff bool
		diff, err = dataChng.GetPrevValue(&prevValue)
		if err == nil {
			err = plugin.dataChangeIface(diff, &value, &prevValue, dataChng.GetChangeType())
		}
	} else if strings.HasPrefix(key, l3.StaticArpKeyPrefix()) {
		var value, prevValue l3.LinuxStaticArpEntries_ArpEntry
		err = dataChng.GetValue(&value)
		if err != nil {
			return err
		}
		var diff bool
		diff, err = dataChng.GetPrevValue(&prevValue)
		if err == nil {
			err = plugin.dataChangeArp(diff, &value, &prevValue, dataChng.GetChangeType())
		}
	} else if strings.HasPrefix(key, l3.StaticRouteKeyPrefix()) {
		var value, prevValue l3.LinuxStaticRoutes_Route
		err = dataChng.GetValue(&value)
		if err != nil {
			return err
		}
		var diff bool
		diff, err = dataChng.GetPrevValue(&prevValue)
		if err == nil {
			err = plugin.dataChangeRoute(diff, &value, &prevValue, dataChng.GetChangeType())
		}
	} else {
		plugin.Log.Warn("ignoring change ", dataChng) //NOT ERROR!
	}
	return err
}
