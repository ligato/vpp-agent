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

package orchestrator

import (
	"os"
	"strings"
	"sync"

	"github.com/go-errors/errors"
	"go.ligato.io/cn-infra/v2/datasync"
	"go.ligato.io/cn-infra/v2/datasync/resync"
	"go.ligato.io/cn-infra/v2/infra"
	"go.ligato.io/cn-infra/v2/logging"
	"go.ligato.io/cn-infra/v2/rpc/grpc"
	"golang.org/x/net/context"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/proto"

	"go.ligato.io/vpp-agent/v3/pkg/models"
	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/plugins/orchestrator/contextdecorator"
	"go.ligato.io/vpp-agent/v3/proto/ligato/generic"
	"go.ligato.io/vpp-agent/v3/proto/ligato/kvscheduler"
)

var (
	// EnableStatusPublishing enables status publishing.
	EnableStatusPublishing = os.Getenv("ENABLE_STATUS_PUBLISHING") != ""

	debugOrchestrator = os.Getenv("DEBUG_ORCHESTRATOR") != ""
)

// Plugin implements sync service for GRPC.
type Plugin struct {
	Deps

	*dispatcher
	manager *genericService

	reflection bool

	// datasync channels
	changeChan   chan datasync.ChangeEvent
	resyncChan   chan datasync.ResyncEvent
	watchDataReg datasync.WatchRegistration

	wg   sync.WaitGroup
	quit chan struct{}
}

// Deps represents dependencies for the plugin.
type Deps struct {
	infra.PluginDeps

	GRPC            grpc.Server
	KVScheduler     kvs.KVScheduler
	Watcher         datasync.KeyValProtoWatcher
	StatusPublisher datasync.KeyProtoValWriter
}

// Init registers the service to GRPC server.
func (p *Plugin) Init() (err error) {
	p.quit = make(chan struct{})

	p.dispatcher = &dispatcher{
		log: logging.DefaultRegistry.NewLogger("dispatcher"),
		db:  newMemStore(),
		kvs: p.KVScheduler,
	}

	// register grpc service
	p.manager = &genericService{
		log:      p.log,
		dispatch: p.dispatcher,
	}

	if grpcServer := p.GRPC.GetServer(); grpcServer != nil {
		p.Log.Debugf("registering generic manager and meta service")
		generic.RegisterManagerServiceServer(grpcServer, p.manager)
		generic.RegisterMetaServiceServer(grpcServer, p.manager)

		// register grpc services for reflection
		if p.reflection {
			p.Log.Debugf("registering grpc reflection service")
			reflection.Register(grpcServer)
		}
	} else {
		p.log.Infof("grpc server is not available")
	}

	p.Log.Infof("Found %d registered models", len(models.RegisteredModels()))
	for _, model := range models.RegisteredModels() {
		p.debugf("- model: %+v", *model.Spec())
	}

	var prefixes []string
	if nbPrefixes := p.kvs.GetRegisteredNBKeyPrefixes(); len(nbPrefixes) > 0 {
		p.log.Infof("Watching %d key prefixes from KVScheduler", len(nbPrefixes))
		for _, prefix := range nbPrefixes {
			p.debugf("- prefix: %s", prefix)
			prefixes = append(prefixes, prefix)
		}
	} else {
		p.log.Warnf("No key prefixes found in KVScheduler (ensure that all KVDescriptors are registered before this)")
	}

	// initialize datasync channels
	p.resyncChan = make(chan datasync.ResyncEvent)
	p.changeChan = make(chan datasync.ChangeEvent)

	p.watchDataReg, err = p.Watcher.Watch(p.PluginName.String(),
		p.changeChan, p.resyncChan, prefixes...)
	if err != nil {
		return err
	}

	return nil
}

// AfterInit subscribes to known NB prefixes.
func (p *Plugin) AfterInit() (err error) {
	// watch datasync events
	p.wg.Add(1)
	go p.watchEvents()

	statusChan := make(chan *kvscheduler.BaseValueStatus, 100)
	p.kvs.WatchValueStatus(statusChan, nil)

	// watch KVSchedular status changes
	p.wg.Add(1)
	go p.watchStatus(statusChan)

	return nil
}

func (p *Plugin) Close() (err error) {
	close(p.quit)
	p.wg.Wait()
	return nil
}

// InitialSync will start initial synchronization.
func (p *Plugin) InitialSync() error {
	// SB resync
	p.Log.Debugf("starting initial SB sync")
	txn := p.KVScheduler.StartNBTransaction()
	ctx := kvs.WithResync(context.Background(), kvs.DownstreamResync, true)
	if _, err := txn.Commit(ctx); err != nil {
		return errors.Errorf("initial SB sync failed: %v", err)
	}
	p.Log.Infof("initial SB sync complete")

	// NB resync
	p.Log.Debugf("starting initial NB sync")
	resync.DefaultPlugin.DoResync() // NB init file data is also resynced here
	p.Log.Infof("initial NB sync complete")

	return nil
}

func (p *Plugin) watchEvents() {
	defer p.wg.Done()

	p.Log.Debugf("watching datasync events")
	defer p.Log.Debugf("done watching datasync events")

	for {
		select {
		case e := <-p.changeChan:
			p.log.Debugf("=> received CHANGE event (%v changes)", len(e.GetChanges()))

			var err error
			var kvPairs []KeyVal
			var keyLabels map[string]Labels

			ctx := e.GetContext()
			if ctx == nil {
				ctx = context.Background()
			}
			labels, ok := contextdecorator.LabelsFromContext(ctx)
			if !ok {
				labels = Labels{}
			}

			for _, x := range e.GetChanges() {
				key := x.GetKey()
				kv := KeyVal{
					Key: key,
				}
				if x.GetChangeType() != datasync.Delete {
					kv.Val, err = UnmarshalLazyValue(kv.Key, x)
					if err != nil {
						p.log.Errorf("decoding value for key %q failed: %v", kv.Key, err)
						continue
					}
				}
				kvPairs = append(kvPairs, kv)
				keyLabels[key] = labels
			}

			if len(kvPairs) == 0 {
				p.log.Warn("no valid kv pairs received in change event")
				e.Done(nil)
				continue
			}

			p.log.Debugf("Change with %d items", len(kvPairs))

			_, withDataSrc := contextdecorator.DataSrcFromContext(ctx)
			if !withDataSrc {
				ctx = contextdecorator.DataSrcContext(ctx, "datasync")
			}
			ctx = kvs.WithRetryDefault(ctx)
			res, err := p.PushData(ctx, kvPairs, keyLabels)
			if err == nil {
				ctx = contextdecorator.PushDataResultContext(ctx, ResultWrapper{Results: res})
			}
			e.Done(err)

		case e := <-p.resyncChan:
			p.log.Debugf("=> received RESYNC event (%v prefixes)", len(e.GetValues()))

			var kvPairs []KeyVal

			for prefix, iter := range e.GetValues() {
				var keyVals []datasync.KeyVal
				for x, done := iter.GetNext(); !done; x, done = iter.GetNext() {
					key := x.GetKey()
					val, err := UnmarshalLazyValue(key, x)
					if err != nil {
						p.log.Errorf("unmarshal value for key %q failed: %v", key, err)
						continue
					}
					kvPairs = append(kvPairs, KeyVal{
						Key: key,
						Val: val,
					})
					p.log.Debugf(" -- key: %s", x.GetKey())
					keyVals = append(keyVals, x)
				}
				if len(keyVals) > 0 {
					p.log.Debugf("- %q (%v items)", prefix, len(keyVals))
				} else {
					p.log.Debugf("- %q (no items)", prefix)
				}
				for _, x := range keyVals {
					p.log.Debugf("\t - %q: (rev: %v)", x.GetKey(), x.GetRevision())
				}
			}

			p.log.Debugf("Resync with %d items", len(kvPairs))

			ctx := e.GetContext()
			if ctx == nil {
				ctx = context.Background()
			}
			_, withDataSrc := contextdecorator.DataSrcFromContext(ctx)
			if !withDataSrc {
				ctx = contextdecorator.DataSrcContext(ctx, "datasync")
			}
			ctx = kvs.WithResync(ctx, kvs.FullResync, true)
			ctx = kvs.WithRetryDefault(ctx)

			res, err := p.PushData(ctx, kvPairs, nil)
			if err == nil {
				ctx = contextdecorator.PushDataResultContext(ctx, ResultWrapper{Results: res})
			}
			e.Done(err)

		case <-p.quit:
			return
		}
	}
}

func (p *Plugin) watchStatus(ch <-chan *kvscheduler.BaseValueStatus) {
	defer p.wg.Done()

	p.Log.Debugf("watching status changes")
	defer p.Log.Debugf("done watching status events")

	for {
		select {
		case s := <-ch:
			p.debugf("incoming status change: %15s %v ===> %v (%v) %v",
				s.Value.State, s.Value.Details, s.Value.Key, s.Value.LastOperation, s.Value.Error)
			for _, dv := range s.DerivedValues {
				p.debugf(" \t%15s %v ---> %v (%v) %v",
					dv.State, dv.Details, dv.Key, dv.LastOperation, dv.Error)
			}

			if EnableStatusPublishing {
				p.publishStatuses([]Result{
					{Key: s.Value.Key, Status: s.Value},
				})
			}

		case <-p.quit:
			return
		}
	}
}

func (p *Plugin) publishStatuses(results []Result) {
	if p.StatusPublisher == nil {
		return
	}

	p.debugf("publishing %d statuses", len(results))
	for _, res := range results {
		statusKey := strings.Replace(res.Key, "config/", "config-status/", 1)
		if statusKey == res.Key {
			p.debugf("replace for key %q failed", res.Key)
			continue
		}
		if err := p.StatusPublisher.Put(statusKey, res.Status, datasync.WithClientLifetimeTTL()); err != nil {
			p.debugf("publishing status for key %q failed: %v", statusKey, err)
		}
	}
}

func (p *Plugin) debugf(f string, a ...interface{}) {
	if debugOrchestrator {
		p.log.Debugf(f, a...)
	}
}

// UnmarshalLazyValue is helper function for unmarshalling from datasync.LazyValue.
func UnmarshalLazyValue(key string, lazy datasync.LazyValue) (proto.Message, error) {
	model, err := models.GetModelForKey(key)
	if err != nil {
		return nil, err
	}
	instance := model.NewInstance()
	// try to deserialize the value into instance
	if err := lazy.GetValue(instance); err != nil {
		return nil, err
	}
	return instance, nil
}

func ContainsAllLabels(want map[string]string, have Labels) bool {
	for wk, wv := range want {
		if hv, ok := have[wk]; !ok || wv != "" && wv != hv {
			return false
		}
	}
	return true
}

func ContainsItemID(want []*generic.Item_ID, have *generic.Item_ID) bool {
	if len(want) == 0 {
		return true
	}
	for _, w := range want {
		if w.Model == have.Model && w.Name == have.Name {
			return true
		}
	}
	return false
}

// TODO: This is hack to avoid import cycle between orchestrator and contextdecorator package.
// Figure out a way to pass result into local client without using wrapper type that implements
// a dummy interface defined inside contextdecorator package.
type ResultWrapper struct {
	Results []Result
}

// implement the dummy interface (see comment above ResultWrapper struct definition)
func (r ResultWrapper) IsPushDataResult() {}
