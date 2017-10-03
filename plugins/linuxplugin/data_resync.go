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
	"strings"

	"github.com/ligato/cn-infra/datasync"
	log "github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/vpp-agent/plugins/linuxplugin/model/interfaces"
)

// DataResyncReq is used to transfer expected configuration of the Linux network stack to the plugins
type DataResyncReq struct {
	// Interfaces is a list af all interfaces that are expected to be in Linux after RESYNC
	Interfaces []*interfaces.LinuxInterfaces_Interface
}

// NewDataResyncReq is a constructor
func NewDataResyncReq() *DataResyncReq {
	return &DataResyncReq{
		// Interfaces is a list af all interfaces that are expected to be in Linux after RESYNC
		Interfaces: []*interfaces.LinuxInterfaces_Interface{},
	}
}

// DataResync delegates resync request only to interface configurator for now.
func (plugin *Plugin) resyncPropageRequest(req *DataResyncReq) error {
	log.DefaultLogger().Info("resync the Linux Configuration")

	plugin.ifConfigurator.Resync(req.Interfaces)

	return nil
}

func resyncParseEvent(resyncEv datasync.ResyncEvent) *DataResyncReq {
	req := NewDataResyncReq()
	for key := range resyncEv.GetValues() {
		log.DefaultLogger().Debug("Received RESYNC key ", key)
	}
	for key, resyncData := range resyncEv.GetValues() {
		if strings.HasPrefix(key, interfaces.InterfaceKeyPrefix()) {
			numInterfaces := resyncAppendInterface(resyncData, req)
			log.DefaultLogger().Debug("Received RESYNC interface values ", numInterfaces)
		} else {
			log.DefaultLogger().Warn("ignoring ", resyncEv)
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

func (plugin *Plugin) subscribeWatcher() (err error) {
	log.DefaultLogger().Debug("subscribeWatcher begin")

	plugin.watchDataReg, err = plugin.Watcher.
		Watch("linuxplugin", plugin.changeChan, plugin.resyncChan, interfaces.InterfaceKeyPrefix())
	if err != nil {
		return err
	}

	log.DefaultLogger().Debug("data watcher watch finished")

	return nil
}

func (plugin *Plugin) changePropagateRequest(dataChng datasync.ChangeEvent) error {
	var err error
	key := dataChng.GetKey()
	log.DefaultLogger().Debug("Start processing change for key: ", key)

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
	} else {
		log.DefaultLogger().Warn("ignoring change ", dataChng) //NOT ERROR!
	}
	return err
}
