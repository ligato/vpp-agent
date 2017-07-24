package linuxplugin

import (
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/vpp-agent/idxvpp"
)

// PluginID used in the Agent Core flavors
const PluginID core.PluginName = "linuxplugin"

// GetIfIndexes gives access to mapping of logical names (used in ETCD configuration) to corresponding Linux interface indexes.
// This mapping is especially helpful for plugins that need to watch for newly added or deleted Linux interfaces.
func GetIfIndexes() idxvpp.NameToIdx {
	if gPlugin == nil {
		return nil
	}
	return plugin().ifIndexes
}
