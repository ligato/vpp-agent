package defaultplugins

import (
	"strings"

	log "github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/vpp-agent/defaultplugins/ifplugin/model/interfaces"
	"golang.org/x/net/context"
	"github.com/ligato/vpp-agent/defaultplugins/l2plugin/model/l2"
)

// WatchEvents goroutine is used to watch for changes in the northbound configuration & NameToIdxMapping notifications
func (plugin *Plugin) watchEvents(ctx context.Context) {
	plugin.wg.Add(1)
	defer plugin.wg.Done()

	for {
		select {
		case resyncConfigEv := <-plugin.resyncConfigChan:
			req := resyncParseEvent(resyncConfigEv)
			err := plugin.resyncConfigPropageRequest(req)

			resyncConfigEv.Done(err)

		case resyncStatusEv := <-plugin.resyncStatusChan:
			var wasError error
			for key, vals := range resyncStatusEv.GetValues() {
				ifStatusPrefix := strings.HasPrefix(key, interfaces.IfStatePrefix)
				log.Debugf("trying to delete obsolete status for key %v begin ", key)
				if strings.HasPrefix(key, interfaces.IfStatePrefix) {
					keys := []string{}
					for {
						x, stop := vals.GetNext()
						if stop {
							break
						}
						keys = append(keys, x.GetKey())
					}
					if len(keys) > 0 {
						err := plugin.resyncIfStateEvents(keys)
						if err != nil {
							wasError = err
						}
					}
				} else if strings.HasPrefix(key, l2.BdStatePrefix) {
					keys := []string{}
					for {
						x, stop := vals.GetNext()
						if stop {
							break
						}
						keys = append(keys, x.GetKey())
					}
					if len(keys) > 0 {
						err := plugin.resyncBdStateEvents(keys)
						if err != nil {
							wasError = err
						}
					}
				}
			}
			resyncStatusEv.Done(wasError)

		case dataChng := <-plugin.changeChan:
			// For FIBs only: if changePropagateRequest ends up without errors, the dataChng.Done is called in l2fib_vppcalls,
			// otherwise the dataChng.Done is called here
			err := plugin.changePropagateRequest(dataChng, dataChng.Done)
			// When the request propagation is completed, send the error context (even if the error is nil)
			plugin.errorChannel <- ErrCtx{dataChng, err}
			if err != nil {
				dataChng.Done(err)
			}

		case ifIdxEv := <-plugin.ifIdxWatchCh:
			if !ifIdxEv.IsDelete() {
				// Keep order
				plugin.bdConfigurator.ResolveCreatedInterface(ifIdxEv.Name, ifIdxEv.Idx)
				plugin.fibConfigurator.ResolveCreatedInterface(ifIdxEv.Name, ifIdxEv.Idx, func(err error) {
					if err != nil {
						log.Error(err)
					}
				})
				// TODO propagate error
			} else {
				plugin.bdConfigurator.ResolveDeletedInterface(ifIdxEv.Name) //TODO ifIdxEv.Idx to not process data events
				plugin.fibConfigurator.ResolveDeletedInterface(ifIdxEv.Name, ifIdxEv.Idx, func(err error) {
					if err != nil {
						log.Error(err)
					}
				})
				// TODO propagate error
			}
			ifIdxEv.Done()
/*
		case linuxIfIdxEv := <-plugin.linuxIfIdxWatchCh:
			if !linuxIfIdxEv.IsDelete() {
				plugin.ifConfigurator.ResolveCreatedLinuxInterface(linuxIfIdxEv.Name, linuxIfIdxEv.Idx)
				// TODO propagate error
			} else {
				plugin.ifConfigurator.ResolveDeletedLinuxInterface(linuxIfIdxEv.Name)
				// TODO propagate error
			}
			linuxIfIdxEv.Done()
*/
		case bdIdxEv := <-plugin.bdIdxWatchCh:
			if !bdIdxEv.IsDelete() {
				plugin.fibConfigurator.ResolveCreatedBridgeDomain(bdIdxEv.Name, bdIdxEv.Idx, func(err error) {
					if err != nil {
						log.Error(err)
					}
				})
				// TODO propagate error
			} else {
				plugin.fibConfigurator.ResolveDeletedBridgeDomain(bdIdxEv.Name, bdIdxEv.Idx, func(err error) {
					if err != nil {
						log.Error(err)
					}
				})
				// TODO propagate error
			}
			bdIdxEv.Done()

		case <-ctx.Done():
			log.Debug("Stop watching events")
			return
		}
	}
}
