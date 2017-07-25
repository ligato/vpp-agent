package ifaceidx

import (
	"fmt"
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/logging/logroot"
	"github.com/ligato/vpp-agent/defaultplugins/ifplugin/model/interfaces"
	"github.com/ligato/vpp-agent/idxvpp/cacheutil"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
)

// Cache the network interfaces of a particular agent by watching (ETCD or different transport)
// Beware: the indexes in cache do not correspond to the real indexes.
func Cache(watcher datasync.Watcher, caller core.PluginName) SwIfIndex {
	resyncName := fmt.Sprintf("iface-cache-%s-%s", caller, watcher)
	swIdx := NewSwIfIndex(nametoidx.NewNameToIdx(logroot.Logger(), caller, resyncName, IndexMetadata))

	helper := cacheutil.CacheHelper{
		Prefix:        interfaces.InterfaceKeyPrefix(),
		IDX:           swIdx.GetMapping(),
		DataPrototype: &interfaces.Interfaces_Interface{Name: "aaa"},
		ParseName:     interfaces.ParseNameFromKey}

	go helper.DoWatching(resyncName, watcher)

	return swIdx
}
