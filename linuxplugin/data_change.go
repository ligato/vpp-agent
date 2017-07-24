package linuxplugin

import (
	"github.com/ligato/cn-infra/db"
	log "github.com/ligato/cn-infra/logging/logrus"
	intf "github.com/ligato/vpp-agent/linuxplugin/model/interfaces"
)

// DataChangeIface propagates data change to the ifConfigurator
func (plugin *Plugin) dataChangeIface(diff bool, value *intf.LinuxInterfaces_Interface, prevValue *intf.LinuxInterfaces_Interface,
	changeType db.PutDel) error {
	log.Debug("dataChangeIface ", diff, " ", changeType, " ", value, " ", prevValue)

	if db.Delete == changeType {
		return plugin.ifConfigurator.DeleteLinuxInterface(prevValue)
	} else if diff {
		return plugin.ifConfigurator.ModifyLinuxInterface(value, prevValue)
	}
	return plugin.ifConfigurator.ConfigureLinuxInterface(value)
}
