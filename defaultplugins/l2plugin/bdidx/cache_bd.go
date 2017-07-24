package bdidx

import (
	"fmt"
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/logging/logroot"
	"github.com/ligato/vpp-agent/defaultplugins/l2plugin/model/l2"
	"github.com/ligato/vpp-agent/idxvpp/cacheutil"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
)

// Cache the network interfaces of a particular agent by watching (ETCD or different transport)
func Cache(watcher datasync.Watcher, caller core.PluginName) BDIndex {
	resyncName := fmt.Sprintf("bd-cache-%s", watcher)
	bdIdx := NewBDIndex(nametoidx.NewNameToIdx(logroot.Logger(), caller, resyncName, IndexMetadata))

	helper := cacheutil.CacheHelper{
		Prefix:        l2.BridgeDomainKeyPrefix(),
		IDX:           bdIdx.GetMapping(),
		DataPrototype: &l2.BridgeDomains_BridgeDomain{},
		ParseName:     l2.ParseBDNameFromKey}

	go helper.DoWatching(resyncName, watcher)

	return bdIdx
}
