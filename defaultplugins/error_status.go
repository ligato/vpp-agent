package defaultplugins

import (
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/db"
	log "github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/vpp-agent/defaultplugins/ifplugin/model/interfaces"
	"github.com/ligato/vpp-agent/defaultplugins/l2plugin/model/l2"
	"strings"
	"time"
)

// ErrCtx is an error context struct which stores event change with object identifier (name, etc) and returned error (can be nil)
type ErrCtx struct {
	change  datasync.ChangeEvent
	errInfo error
}

// Maximum count of entries which can be stored in error log. If this number is exceeded, the oldest log entry will be
// removed
const maxErrCount = 50

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

			if errInfo == nil && change.GetChangeType() == db.Delete {
				// Data were successfully removed so delete all error entries related to the data (if exists)
				plugin.removeErrorLogEntry(key)
			} else if errInfo != nil {
				// There is an error to store
				plugin.processError(errInfo, key, changeType, change)
			}
		}
	}
}

// Process provided error data and add a new entry
func (plugin *Plugin) processError(errInfo error, key string, changeType db.PutDel, change datasync.ChangeEvent) {
	// Interfaces
	if strings.HasPrefix(key, interfaces.InterfaceKeyPrefix()) {
		var err error
		var iface, prevIface interfaces.Interfaces_Interface
		if err := change.GetValue(&iface); err != nil {
			log.Errorf("Failed to propagate interface error, cause: %v", err)
			return
		}
		var prevValExists bool
		if prevValExists, err = change.GetPrevValue(&prevIface); err != nil {
			log.Errorf("Failed to propagate interface error, cause: %v", err)
			return
		}
		var ifaceName string
		if iface.Name == "" && prevValExists {
			ifaceName = prevIface.Name
		} else {
			ifaceName = iface.Name
		}
		errorKey := interfaces.InterfaceErrorKey(ifaceName)
		errorMessages := plugin.composeInterfaceErrors(ifaceName, changeType, errInfo)
		log.Infof("Logging error for interface %v", ifaceName)
		err = plugin.addErrorLogEntry(errorKey, errorMessages)
		if err != nil {
			log.Errorf("Failed to propagate interface error, cause: %v", err)
		}
		// Bridge domains
	} else if strings.HasPrefix(key, l2.BridgeDomainKeyPrefix()) {
		var err error
		var bd, prevBd l2.BridgeDomains_BridgeDomain
		if err := change.GetValue(&bd); err != nil {
			log.Errorf("Failed to propagate bridge domain error, cause: %v", err)
			return
		}
		var prevValExists bool
		if prevValExists, err = change.GetPrevValue(&prevBd); err != nil {
			log.Errorf("Failed to propagate bridgeDomain error, cause: %v", err)
			return
		}
		var bdName string
		if bd.Name == "" && prevValExists {
			bdName = prevBd.Name
		} else {
			bdName = bd.Name
		}
		errorKey := l2.BridgeDomainErrorKey(bdName)
		errorMessages := plugin.composeBridgeDomainErrors(bdName, changeType, errInfo)
		log.Infof("Logging error for bridge domain %v", bdName)
		err = plugin.addErrorLogEntry(errorKey, errorMessages)
		if err != nil {
			log.Errorf("Failed to propagate bridge domain error, cause: %v", err)
		}
	}
}

// Create a list of errors for the provided interface and register it. If the interface already has some errors logged,
// find it and add a new error log to the list
func (plugin *Plugin) composeInterfaceErrors(ifName string, change db.PutDel, errs ...error) *interfaces.InterfaceErrors {
	// Read registered data
	_, data, exists := plugin.errorIndexes.LookupIdx(ifName)

	// Compose new data
	var interfaceErrors []*interfaces.InterfaceErrors_InterfaceError
	for _, err := range errs {
		if err == nil {
			continue
		}
		interfaceError := &interfaces.InterfaceErrors_InterfaceError{
			InterfaceName: ifName,
			ChangeType:    string(change),
			ErrorMessage:  err.Error(),
			LastChange:    time.Now().Unix(),
		}
		interfaceErrors = append(interfaceErrors, interfaceError)
	}

	if exists {
		if loggedDataSet, ok := data.([]*interfaces.InterfaceErrors_InterfaceError); ok {
			for _, loggedData := range loggedDataSet {
				interfaceErrors = append(interfaceErrors, loggedData)
			}
		}
	}

	// Register new data
	plugin.errorIndexes.RegisterName(ifName, plugin.errorIdxSeq, interfaceErrors)
	plugin.errorIdxSeq++

	return &interfaces.InterfaceErrors{
		Interface: interfaceErrors,
	}
}

// Create a list of errors for the provided bridge domain and register it. If the bridge domain already has some errors
// logged, find it and add a new error log to the list
func (plugin *Plugin) composeBridgeDomainErrors(bdName string, change db.PutDel, errs ...error) *l2.BridgeDomainErrors {
	// Read registered data
	_, data, exists := plugin.errorIndexes.LookupIdx(bdName)

	// Compose new data
	var bridgeDomainErrors []*l2.BridgeDomainErrors_BridgeDomainError
	for _, err := range errs {
		if err == nil {
			continue
		}
		bridgeDomainError := &l2.BridgeDomainErrors_BridgeDomainError{
			BdName:       bdName,
			ChangeType:   string(change),
			ErrorMessage: err.Error(),
			LastChange:   time.Now().Unix(),
		}
		bridgeDomainErrors = append(bridgeDomainErrors, bridgeDomainError)
	}

	if exists {
		if loggedDataSet, ok := data.([]*l2.BridgeDomainErrors_BridgeDomainError); ok {
			for _, loggedData := range loggedDataSet {
				bridgeDomainErrors = append(bridgeDomainErrors, loggedData)
			}
		}
	}

	// Register new data
	plugin.errorIndexes.RegisterName(bdName, plugin.errorIdxSeq, bridgeDomainErrors)
	plugin.errorIdxSeq++

	return &l2.BridgeDomainErrors{
		BridgeDomain: bridgeDomainErrors,
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
			log.Infof("There is no error entry with index %v", firstActiveIndex)
		} else {
			var oldEntryKey string
			if _, ok := oldErrors.([]*interfaces.InterfaceErrors_InterfaceError); ok {
				oldEntryKey = interfaces.InterfaceErrorKey(name)
			} else if _, ok := oldErrors.([]*l2.BridgeDomainErrors_BridgeDomainError); ok {
				oldEntryKey = l2.BridgeDomainErrorKey(name)
			} else {
				log.Warnf("Unknown type od data: %v", errors)
			}
			log.Debugf("Removing error log entry from history: %v, %v", name, oldEntryKey)
			plugin.removeErrorLogEntry(oldEntryKey)
		}
	}
	// Get errors type
	if data, ok := errors.(*interfaces.InterfaceErrors); ok {
		err := plugin.Transport.PublishData(key, data)
		if err != nil {
			return err
		}
	} else if data, ok := errors.(*l2.BridgeDomainErrors); ok {
		err := plugin.Transport.PublishData(key, data)
		if err != nil {
			return err
		}
	} else {
		log.Warnf("Unknown type od data: %v", errors)
	}
	return nil
}

// Generic method which can be used to remove error data under provided key
func (plugin *Plugin) removeErrorLogEntry(key string) {
	dividedKey := strings.Split(key, "/")
	// Last part of the key is a name
	name := dividedKey[len(dividedKey)-1]
	// The rest is a prefix
	prefix := strings.Replace(key, name, "", 1)

	_, data, exists := plugin.errorIndexes.LookupIdx(name)
	if !exists {
		log.Debugf("There is no error log related to the %v", name)
		return
	}
	if data == nil {
		log.Infof("Error-Idx-Map entry %v: missing data", name)
		return
	}
	// Interfaces
	if prefix == interfaces.InterfaceErrorPrefix() {
		key := interfaces.InterfaceErrorKey(name)
		log.Infof("Error log for interface %v cleared", name)
		plugin.Transport.PublishData(key, nil)
		// Bridge domains
	} else if prefix == l2.BridgeDomainErrorPrefix() {
		key := l2.BridgeDomainErrorKey(name)
		log.Infof("Error log for bridge domain %v cleared", name)
		plugin.Transport.PublishData(key, nil)
	} else {
		log.Infof("Unknown key prefix %v", prefix)
	}
	plugin.errorIndexes.UnregisterName(name)
}

// Auxiliary method returns the count of all entries in the error mapping and an index of the first element
func (plugin *Plugin) calculateErrorMappingEntries() (uint32, uint32) {
	var index uint32
	var count uint32
	var firstIndex uint32
	for index = 1; index <= plugin.errorIdxSeq; index++ {
		_, _, exists := plugin.errorIndexes.LookupName(index)
		if exists {
			count++
			if firstIndex == 0 {
				firstIndex = index
			}
		}
	}
	return count, firstIndex
}
