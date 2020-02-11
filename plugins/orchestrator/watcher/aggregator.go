//  Copyright (c) 2019 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package watcher

import (
	"context"
	"fmt"
	"strings"

	"go.ligato.io/cn-infra/v2/datasync"
	"go.ligato.io/cn-infra/v2/datasync/kvdbsync/local"
	"go.ligato.io/cn-infra/v2/datasync/resync"
	"go.ligato.io/cn-infra/v2/datasync/syncbase"
	"go.ligato.io/cn-infra/v2/infra"
	"go.ligato.io/cn-infra/v2/logging"
	"go.ligato.io/cn-infra/v2/utils/safeclose"
)

// Option is a function that acts on a Plugin to inject Dependencies or configuration
type Option func(*Aggregator)

// UseWatchers returns option that sets watchers.
func UseWatchers(watchers ...datasync.KeyValProtoWatcher) Option {
	return func(p *Aggregator) {
		p.Watchers = watchers
	}
}

// Aggregator is an adapter that allows multiple
// watchers (KeyValProtoWatcher) to be aggregated in one.
// Watch request is delegated to all of them.
type Aggregator struct {
	infra.PluginDeps

	keyPrefixes []string
	localKVs    map[string]datasync.KeyVal

	Resync   *resync.Plugin
	Local    *syncbase.Registry
	Watchers []datasync.KeyValProtoWatcher
}

// NewPlugin creates a new Plugin with the provides Options
func NewPlugin(opts ...Option) *Aggregator {
	p := &Aggregator{}

	p.PluginName = "aggregator"
	p.Local = local.DefaultRegistry
	p.Resync = &resync.DefaultPlugin

	for _, o := range opts {
		o(p)
	}
	p.PluginDeps.SetupLog()

	return p
}

func (p *Aggregator) Init() error {
	p.localKVs = map[string]datasync.KeyVal{}
	return nil
}

// Watch subscribes to every transport available within transport aggregator
// and also subscribes to localclient (local.Registry).
// The function implements KeyValProtoWatcher.Watch().
func (p *Aggregator) Watch(
	resyncName string,
	changeChan chan datasync.ChangeEvent,
	resyncChan chan datasync.ResyncEvent,
	keyPrefixes ...string,
) (datasync.WatchRegistration, error) {

	p.keyPrefixes = keyPrefixes

	// prepare list of watchers
	var watchers []datasync.KeyValProtoWatcher
	for _, w := range p.Watchers {
		if l, ok := w.(*syncbase.Registry); ok && p.Local != nil && l == p.Local {
			p.Log.Warn("found local registry (localclient) in watchers, ignoring it..")
			continue
		}
		watchers = append(watchers, w)
	}
	p.Watchers = watchers

	// start watch for all watchers
	p.Log.Infof("Watch for %v with %d prefixes", resyncName, len(keyPrefixes))

	aggrResync := make(chan datasync.ResyncEvent, len(watchers))

	go p.watchAggrResync(aggrResync, resyncChan)

	var registrations []datasync.WatchRegistration
	for i, adapter := range watchers {
		partChange := make(chan datasync.ChangeEvent)
		partResync := make(chan datasync.ResyncEvent)

		name := fmt.Sprint(adapter) + "/" + resyncName
		watcherReg, err := adapter.Watch(name, changeChan, partResync, keyPrefixes...)
		if err != nil {
			return nil, err
		}

		go func(i int, chanChange chan datasync.ChangeEvent, chanResync chan datasync.ResyncEvent) {
			for {
				select {
				case e := <-chanChange:
					p.Log.Debugf("watcher %d got CHANGE PART, sending to aggregated", i)
					changeChan <- e

				case e := <-chanResync:
					p.Log.Debugf("watcher %d got RESYNC PART, sending to aggregated", i)
					aggrResync <- e
				}
			}
		}(i+1, partChange, partResync)

		if watcherReg != nil {
			registrations = append(registrations, watcherReg)
		}
	}

	// register and watch for localclient
	partResync := make(chan datasync.ResyncEvent)
	partChange := make(chan datasync.ChangeEvent)

	go p.watchLocalEvents(partChange, changeChan, partResync)

	name := "LOCAL" + "/" + resyncName
	localReg, err := p.Local.Watch(name, partChange, partResync, keyPrefixes...)
	if err != nil {
		return nil, err
	}

	p.Log.Debug("added localclient as aggregated watcher")

	registrations = append(registrations, localReg)

	return &WatchRegistration{
		Registrations: registrations,
	}, nil
}

func (p *Aggregator) watchAggrResync(aggrResync, resyncCh chan datasync.ResyncEvent) {
	aggregatedResync := func(allResyncs []datasync.ResyncEvent) {
		var prefixKeyVals = map[string]map[string]datasync.KeyVal{}

		kvToKeyVals := func(prefix string, kv datasync.KeyVal) {
			keyVals, ok := prefixKeyVals[prefix]
			if !ok {
				p.Log.Debugf(" - keyval prefix: %v", prefix)
				keyVals = map[string]datasync.KeyVal{}
				prefixKeyVals[prefix] = keyVals
			}
			key := kv.GetKey()
			if _, ok := keyVals[key]; ok {
				p.Log.Warnf("resync from watcher overwrites key: %v", key)
			}
			keyVals[key] = kv
		}

		// process resync events from all watchers
		p.Log.Debugf("preparing keyvals for aggregated resync from %d cached resyncs", len(allResyncs))
		for _, ev := range allResyncs {
			for prefix, iterator := range ev.GetValues() {
				for {
					kv, allReceived := iterator.GetNext()
					if allReceived {
						break
					}

					kvToKeyVals(prefix, kv)
				}
			}
		}

		// process keyvals from localclient
		p.Log.Debugf("preparing localclient keyvals for aggregated resync with %d keyvals", len(allResyncs))
		for key, kv := range p.localKVs {
			var kvprefix string
			for _, prefix := range p.keyPrefixes {
				if strings.HasPrefix(key, prefix) {
					kvprefix = prefix
					break
				}
			}
			if kvprefix == "" {
				p.Log.Warnf("not found registered prefix for keyval from localclient with key: %v", key)
			}
			kvToKeyVals(kvprefix, kv)
		}

		// prepare aggregated resync
		var vals = map[string]datasync.KeyValIterator{}
		for prefix, keyVals := range prefixKeyVals {
			var data []datasync.KeyVal
			for _, kv := range keyVals {
				data = append(data, kv)
			}
			vals[prefix] = syncbase.NewKVIterator(data)
		}
		resEv := syncbase.NewResyncEventDB(context.Background(), vals)

		p.Log.Debugf("sending aggregated resync event (%d prefixes) to original resync channel", len(vals))
		resyncCh <- resEv
		p.Log.Debugf("aggregated resync was accepted, waiting for done chan")
		resErr := <-resEv.DoneChan
		p.Log.Debugf("aggregated resync done (err=%v) watchers", resErr)

	}

	var cachedResyncs []datasync.ResyncEvent

	// process resync events from watchers
	for {
		select {
		case e, ok := <-aggrResync:
			if !ok {
				p.Log.Debugf("aggrResync channel was closed")
				return
			}

			cachedResyncs = append(cachedResyncs, e)
			p.Log.Debugf("watchers received resync event (%d/%d watchers done)", len(cachedResyncs), len(p.Watchers))

			e.Done(nil)
		}

		if len(cachedResyncs) == len(p.Watchers) {
			p.Log.Debug("resyncs from all watchers received, calling aggregated resync")
			aggregatedResync(cachedResyncs)
			// clear resyncs
			cachedResyncs = nil
		}
	}
}

func (p *Aggregator) watchLocalEvents(partChange, changeChan chan datasync.ChangeEvent, partResync chan datasync.ResyncEvent) {
	for {
		select {
		case e := <-partChange:
			p.Log.Debugf("LOCAL got CHANGE part, %d changes, sending to aggregated", len(e.GetChanges()))

			for _, change := range e.GetChanges() {
				key := change.GetKey()
				switch change.GetChangeType() {
				case datasync.Delete:
					p.Log.Debugf(" - DEL %s", key)
					delete(p.localKVs, key)
				case datasync.Put:
					p.Log.Debugf(" - PUT %s", key)
					p.localKVs[key] = change
				}
			}
			changeChan <- e

		case e := <-partResync:
			p.Log.Debugf("LOCAL watcher got RESYNC part, sending to aggregated")

			p.localKVs = map[string]datasync.KeyVal{}
			for _, iterator := range e.GetValues() {
				for {
					kv, allReceived := iterator.GetNext()
					if allReceived {
						break
					}

					key := kv.GetKey()
					p.localKVs[key] = kv
				}
			}
			p.Log.Debugf("LOCAL watcher resynced %d keyvals", len(p.localKVs))
			e.Done(nil)

			p.Log.Debug("LOCAL watcher calling RESYNC")
			p.Resync.DoResync()
		}
	}
}

// WatchRegistration is adapter that allows multiple
// registrations (WatchRegistration) to be aggregated in one.
// Close operation is applied collectively to all included registration.
type WatchRegistration struct {
	Registrations []datasync.WatchRegistration
}

// Register new key for all available aggregator objects. Call Register(keyPrefix) on specific registration
// to add the key from that registration only
func (wa *WatchRegistration) Register(resyncName, keyPrefix string) error {
	for _, registration := range wa.Registrations {
		if err := registration.Register(resyncName, keyPrefix); err != nil {
			logging.DefaultLogger.Warnf("aggregated register failed: %v", err)
		}
	}

	return nil
}

// Unregister closed registration of specific key under all available aggregator objects.
// Call Unregister(keyPrefix) on specific registration to remove the key from that registration only
func (wa *WatchRegistration) Unregister(keyPrefix string) error {
	for _, registration := range wa.Registrations {
		if err := registration.Unregister(keyPrefix); err != nil {
			logging.DefaultLogger.Warnf("aggregated unregister failed: %v", err)
		}
	}

	return nil
}

// Close every registration under the aggregator.
// This function implements WatchRegistration.Close().
func (wa *WatchRegistration) Close() error {
	return safeclose.Close(wa.Registrations)
}
