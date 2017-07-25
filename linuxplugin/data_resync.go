package linuxplugin

import (
	"strings"

	"github.com/ligato/cn-infra/datasync"
	log "github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/vpp-agent/linuxplugin/model/interfaces"
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
	log.Info("resync the Linux Configuration")

	plugin.ifConfigurator.Resync(req.Interfaces)

	return nil
}

func resyncParseEvent(resyncEv datasync.ResyncEvent) *DataResyncReq {
	req := NewDataResyncReq()
	for key := range resyncEv.GetValues() {
		log.Debug("Received RESYNC key ", key)
	}
	for key, resyncData := range resyncEv.GetValues() {
		if strings.HasPrefix(key, interfaces.InterfaceKeyPrefix()) {
			numInterfaces := resyncAppendInterface(resyncData, req)
			log.Debug("Received RESYNC interface values ", numInterfaces)
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

func (plugin *Plugin) subscribeWatcher() (err error) {
	log.Debug("subscribeWatcher begin")

	plugin.watchDataReg, err = plugin.transport.
		WatchData("linuxplugin", plugin.changeChan, plugin.resyncChan, interfaces.InterfaceKeyPrefix())
	if err != nil {
		return err
	}

	log.Debug("data transport watch finished")

	return nil
}

func (plugin *Plugin) changePropagateRequest(dataChng datasync.ChangeEvent) error {
	var err error
	key := dataChng.GetKey()
	log.Debug("Start processing change for key: ", key)

	if strings.HasPrefix(key, interfaces.InterfaceKeyPrefix()) {
		var value, prevValue interfaces.LinuxInterfaces_Interface
		err = dataChng.GetValue(&value)
		if err != nil {
			return err
		}
		diff, err := dataChng.GetPrevValue(&prevValue)
		if err == nil {
			err = plugin.dataChangeIface(diff, &value, &prevValue, dataChng.GetChangeType())
		}
	} else {
		log.Warn("ignoring change ", dataChng) //NOT ERROR!
	}
	return err
}
