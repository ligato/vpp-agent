package linuxplugin

import (
	log "github.com/ligato/cn-infra/logging/logrus"
	"golang.org/x/net/context"
)

// WatchEvents goroutine is used to watch for changes in the northbound configuration
func (plugin *Plugin) watchEvents(ctx context.Context) {
	plugin.wg.Add(1)
	defer plugin.wg.Done()

	for {
		select {
		case resyncEv := <-plugin.resyncChan:
			req := resyncParseEvent(resyncEv)
			err := plugin.resyncPropageRequest(req)

			resyncEv.Done(err)

		case dataChng := <-plugin.changeChan:
			err := plugin.changePropagateRequest(dataChng)

			dataChng.Done(err)

		case <-ctx.Done():
			log.Debug("Stop watching events")
			return
		}
	}
}
