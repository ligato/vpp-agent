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
	"time"

	"github.com/ligato/cn-infra/datasync"
	log "github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/model/l2"
)

// ErrCtx is an error context struct which stores event change with object identifier (name, etc) and returned error (can be nil)
type ErrCtx struct {
	change  datasync.ChangeEvent
	errInfo error
}

// Maximum count of entries which can be stored in error log. If this number is exceeded, the oldest log entry will be
// removed
const maxErrCount = 10

// Wait on ErrCtx object which is then used to handle the error log. Two cases are treated:
// 1. If there is any error present, save it together with time stamp and change type under particular key
// 2. If error == nil and change type is 'Delete', it means some data were successfully removed and the whole error log
// related to the data can be removed as well
func (plugin *Plugin) changePropagateError() {
	for {
		select {
		case errorState := <-plugin.errorChannel:
			change := errorState.change
			errInfo := errorState.errInfo
			key := change.GetKey()
			changeType := change.GetChangeType()

			if errInfo == nil && change.GetChangeType() == datasync.Delete {
				// Data were successfully removed so delete all error entries related to the data (if exists)
				plugin.removeErrorLog(key)
			} else if errInfo != nil {
				// There is an error to store
				plugin.processError(errInfo, key, changeType, change)
			}
		}
	}
}

// Process provided error data and add a new entry
func (plugin *Plugin) processError(errInfo error, key string, changeType datasync.PutDel, change datasync.ChangeEvent) {
	// Interfaces
	if strings.HasPrefix(key, interfaces.InterfaceKeyPrefix()) {
		var err error
		var iface, prevIface interfaces.Interfaces_Interface
		if err := change.GetValue(&iface); err != nil {
			log.DefaultLogger().Errorf("Failed to propagate interface error, cause: %v", err)
			return
		}
		var prevValExists bool
		if prevValExists, err = change.GetPrevValue(&prevIface); err != nil {
			log.DefaultLogger().Errorf("Failed to propagate interface error, cause: %v", err)
			return
		}
		var ifaceName string
		if iface.Name == "" && prevValExists {
			ifaceName = prevIface.Name
		} else {
			ifaceName = iface.Name
		}
		ifaceErrKey := interfaces.InterfaceErrorKey(ifaceName)
		ifaceErrList := plugin.composeInterfaceErrors(ifaceName, changeType, errInfo)
		log.DefaultLogger().Infof("Logging error for interface %v", ifaceName)
		err = plugin.addErrorLogEntry(ifaceErrKey, ifaceErrList)
		if err != nil {
			log.DefaultLogger().Errorf("Failed to propagate interface error, cause: %v", err)
		}
		// Bridge domains
	} else if strings.HasPrefix(key, l2.BridgeDomainKeyPrefix()) {
		var err error
		var bd, prevBd l2.BridgeDomains_BridgeDomain
		if err := change.GetValue(&bd); err != nil {
			log.DefaultLogger().Errorf("Failed to propagate bridge domain error, cause: %v", err)
			return
		}
		var prevValExists bool
		if prevValExists, err = change.GetPrevValue(&prevBd); err != nil {
			log.DefaultLogger().Errorf("Failed to propagate bridgeDomain error, cause: %v", err)
			return
		}
		var bdName string
		if bd.Name == "" && prevValExists {
			bdName = prevBd.Name
		} else {
			bdName = bd.Name
		}
		bdErrKey := l2.BridgeDomainErrorKey(bdName)
		bdErrList := plugin.composeBridgeDomainErrors(bdName, changeType, errInfo)
		log.DefaultLogger().Infof("Logging error for bridge domain %v", bdName)
		err = plugin.addErrorLogEntry(bdErrKey, bdErrList)
		if err != nil {
			log.DefaultLogger().Errorf("Failed to propagate bridge domain error, cause: %v", err)
		}
	}
}

// Create a list of errors for the provided interface and register it. If the interface already has some errors logged,
// find it and add a new error log to the list
func (plugin *Plugin) composeInterfaceErrors(ifName string, change datasync.PutDel, errs ...error) *interfaces.InterfaceErrors_Interface {
	// Read registered data
	_, data, exists := plugin.errorIndexes.LookupIdx(ifName)

	// Compose new data
	var interfaceErrors []*interfaces.InterfaceErrors_Interface_ErrorData
	for _, err := range errs {
		if err == nil {
			continue
		}
		interfaceError := &interfaces.InterfaceErrors_Interface_ErrorData{
			ChangeType:   string(change),
			ErrorMessage: err.Error(),
			LastChange:   time.Now().Unix(),
		}
		interfaceErrors = append(interfaceErrors, interfaceError)
	}

	if exists {
		if loggedDataSet, ok := data.([]*interfaces.InterfaceErrors_Interface_ErrorData); ok {
			for _, loggedData := range loggedDataSet {
				interfaceErrors = append(interfaceErrors, loggedData)
			}
		}
	}

	// Register new data
	plugin.errorIndexes.RegisterName(ifName, plugin.errorIdxSeq, interfaceErrors)
	plugin.errorIdxSeq++

	return &interfaces.InterfaceErrors_Interface{
		InterfaceName: ifName,
		ErrorData:     interfaceErrors,
	}
}

// Create a list of errors for the provided bridge domain and register it. If the bridge domain already has some errors
// logged, find it and add a new error log to the list
func (plugin *Plugin) composeBridgeDomainErrors(bdName string, change datasync.PutDel, errs ...error) *l2.BridgeDomainErrors_BridgeDomain {
	// Read registered data
	_, data, exists := plugin.errorIndexes.LookupIdx(bdName)

	// Compose new data
	var bridgeDomainErrors []*l2.BridgeDomainErrors_BridgeDomain_ErrorData
	for _, err := range errs {
		if err == nil {
			continue
		}
		bridgeDomainError := &l2.BridgeDomainErrors_BridgeDomain_ErrorData{
			ChangeType:   string(change),
			ErrorMessage: err.Error(),
			LastChange:   time.Now().Unix(),
		}
		bridgeDomainErrors = append(bridgeDomainErrors, bridgeDomainError)
	}

	if exists {
		if loggedDataSet, ok := data.([]*l2.BridgeDomainErrors_BridgeDomain_ErrorData); ok {
			for _, loggedData := range loggedDataSet {
				bridgeDomainErrors = append(bridgeDomainErrors, loggedData)
			}
		}
	}

	// Register new data
	plugin.errorIndexes.RegisterName(bdName, plugin.errorIdxSeq, bridgeDomainErrors)
	plugin.errorIdxSeq++

	return &l2.BridgeDomainErrors_BridgeDomain{
		BdName:    bdName,
		ErrorData: bridgeDomainErrors,
	}
}

// Generic method which can be used to put error object under provided key to the ETCD. If there is more items stored
// than the defined maximal count, the first entry from the mapping is removed
func (plugin *Plugin) addErrorLogEntry(key string, errors interface{}) error {
	totalErrorCount, firstActiveIndex := plugin.calculateErrorMappingEntries()
	name, oldErrors, found := plugin.errorIndexes.LookupName(firstActiveIndex)
	if totalErrorCount > maxErrCount {
		// Remove oldest entry
		if !found {
			log.DefaultLogger().Infof("There is no error entry with index %v", firstActiveIndex)
		} else {
			var oldEntryKey string
			if _, ok := oldErrors.([]*interfaces.InterfaceErrors_Interface_ErrorData); ok {
				oldEntryKey = interfaces.InterfaceErrorKey(name)
			} else if _, ok := oldErrors.([]*l2.BridgeDomainErrors_BridgeDomain_ErrorData); ok {
				oldEntryKey = l2.BridgeDomainErrorKey(name)
			} else {
				log.DefaultLogger().Warnf("Unknown type od data: %v", errors)
			}
			log.DefaultLogger().Debugf("Removing error log entry from history: %v, %v", name, oldEntryKey)
			plugin.removeOldestErrorLogEntry(oldEntryKey)
		}
	}
	// Get errors type
	if data, ok := errors.(*interfaces.InterfaceErrors_Interface); ok {
		err := plugin.Publish.Put(key, data)
		if err != nil {
			return err
		}
	} else if data, ok := errors.(*l2.BridgeDomainErrors_BridgeDomain); ok {
		err := plugin.Publish.Put(key, data)
		if err != nil {
			return err
		}
	} else {
		log.DefaultLogger().Warnf("Unknown type od data: %v", errors)
	}
	return nil
}

func (plugin *Plugin) removeErrorLog(key string) {
	dividedKey := strings.Split(key, "/")
	// Last part of the key is a name
	name := dividedKey[len(dividedKey)-1]
	// The rest is a prefix
	prefix := strings.Replace(key, name, "", 1)

	if prefix == interfaces.InterfacePrefix {
		key := interfaces.InterfaceErrorKey(name)
		plugin.Publish.Put(key, nil)
		log.DefaultLogger().Infof("Error status log for interface %v cleared", name)
	} else if prefix == l2.BdPrefix {
		key := l2.BridgeDomainErrorKey(name)
		plugin.Publish.Put(key, nil)
		log.DefaultLogger().Infof("Error status log for bridge domain %v cleared", name)
	} else {
		log.DefaultLogger().Infof("Error status log: unknown type of prefix: %v", prefix)
	}
}

// Generic method which can be used to remove oldest error data under provided key
func (plugin *Plugin) removeOldestErrorLogEntry(key string) {
	log.DefaultLogger().Warnf("Key: %v", key)
	var name string
	var metaData interface{}
	var exists bool
	if strings.HasPrefix(key, interfaces.IfErrorPrefix) {
		name = strings.Replace(key, interfaces.IfErrorPrefix, "", 1)
		_, metaData, exists = plugin.errorIndexes.LookupIdx(name)
	} else if strings.HasPrefix(key, l2.BdErrPrefix) {
		name = strings.Replace(key, l2.BdErrPrefix, "", 1)
		_, metaData, exists = plugin.errorIndexes.LookupIdx(name)
	}
	if !exists {
		log.DefaultLogger().Debugf("There is no error log related to the %v", name)
		return
	}
	if metaData == nil {
		log.DefaultLogger().Infof("Error-Idx-Map entry %v: missing metaData", name)
		return
	}
	log.DefaultLogger().Warnf("Name: %v", name)
	switch errData := metaData.(type) {
	// Interfaces
	case []*interfaces.InterfaceErrors_Interface_ErrorData:
		key := interfaces.InterfaceErrorKey(name)
		// If there are more than one error under the interface key, remove the oldest one
		if len(errData) > 1 {
			errData = append(errData[:0], errData[1:]...)
			log.DefaultLogger().Infof("Error log for interface %v: oldest entry removed", name)
			plugin.Publish.Put(key, &interfaces.InterfaceErrors_Interface{
				InterfaceName: name,
				ErrorData:     errData,
			})
			plugin.errorIndexes.RegisterName(name, plugin.errorIdxSeq, errData)
			plugin.errorIdxSeq++
		} else {
			log.DefaultLogger().Infof("Error log for interface %v cleared", name)
			plugin.Publish.Put(key, nil)
			plugin.errorIndexes.UnregisterName(name)
		}
		// Bridge domains
	case []*l2.BridgeDomainErrors_BridgeDomain_ErrorData:
		key := l2.BridgeDomainErrorKey(name)
		// If there are more than one error under the bridge domain key, remove the oldest one
		if len(errData) > 1 {
			errData = append(errData[:0], errData[1:]...)
			log.DefaultLogger().Infof("Error log for bridge domain %v: oldest entry removed", name)
			plugin.Publish.Put(key, &l2.BridgeDomainErrors_BridgeDomain{
				BdName:    name,
				ErrorData: errData,
			})
			plugin.errorIndexes.RegisterName(name, plugin.errorIdxSeq, errData)
			plugin.errorIdxSeq++
		} else {
			log.DefaultLogger().Infof("Error log for bridge domain %v cleared", name)
			plugin.Publish.Put(key, nil)
			plugin.errorIndexes.UnregisterName(name)
		}
	}
}

// Auxiliary method returns the count of all error entries under every interface/bridge domain in the error mapping and
// a index of the first element
func (plugin *Plugin) calculateErrorMappingEntries() (uint32, uint32) {
	var index uint32
	var count int
	var firstIndex uint32
	for index = 1; index <= plugin.errorIdxSeq; index++ {
		_, meta, exists := plugin.errorIndexes.LookupName(index)
		if exists {
			switch errDataList := meta.(type) {
			case []*interfaces.InterfaceErrors_Interface_ErrorData:
				count = count + len(errDataList)
			case []*l2.BridgeDomainErrors_BridgeDomain_ErrorData:
				count = count + len(errDataList)
			}
			if firstIndex == 0 {
				firstIndex = index
			}
		}
	}
	return uint32(count), firstIndex
}
