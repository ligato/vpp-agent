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

	"github.com/golang/protobuf/proto"
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/infra"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/rpc/grpc"
	"golang.org/x/net/context"
	"google.golang.org/grpc/reflection"

	"go.ligato.io/vpp-agent/v3/pkg/models"
	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
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

	manager *genericService

	// datasync channels
	changeChan   chan datasync.ChangeEvent
	resyncChan   chan datasync.ResyncEvent
	watchDataReg datasync.WatchRegistration

	reflection bool

	*dispatcher
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
		generic.RegisterManagerServiceServer(grpcServer, p.manager)
		generic.RegisterMetaServiceServer(grpcServer, p.manager)
		// register grpc services for reflection
		if p.reflection {
			p.Log.Infof("registering grpc reflection service")
			reflection.Register(grpcServer)
		}
	} else {
		p.log.Infof("grpc server not available")
	}

	nbPrefixes := p.kvs.GetRegisteredNBKeyPrefixes()
	if len(nbPrefixes) > 0 {
		p.log.Infof("Watch starting for %d registered NB prefixes", len(nbPrefixes))
	} else {
		p.log.Warnf("No registered NB prefixes found in KVScheduler (ensure that all KVDescriptors are registered before this)")
	}

	var prefixes []string
	for _, prefix := range nbPrefixes {
		p.log.Debugf("- watching NB prefix: %s", prefix)
		prefixes = append(prefixes, prefix)
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
	go p.watchEvents()

	statusChan := make(chan *kvscheduler.BaseValueStatus, 100)
	p.kvs.WatchValueStatus(statusChan, nil)
	go p.watchStatus(statusChan)

	return nil
}

// InitialSync will start initial synchronization with downstream.
func (p *Plugin) InitialSync() {
	// FIXME: KVScheduler needs to have some type of sync that only refreshes state from SB
	p.Log.Debugf("starting initial sync")
	txn := p.KVScheduler.StartNBTransaction()
	ctx := kvs.WithResync(context.Background(), kvs.DownstreamResync, true)
	if _, err := txn.Commit(ctx); err != nil {
		p.Log.Warnf("initial sync failed: %v", err)
	} else {
		p.Log.Infof("initial sync complete")
	}
}

func (p *Plugin) watchEvents() {
	for {
		select {
		case e := <-p.changeChan:
			p.log.Debugf("=> received CHANGE event (%v changes)", len(e.GetChanges()))

			var err error
			var kvPairs []KeyVal

			for _, x := range e.GetChanges() {
				kv := KeyVal{
					Key: x.GetKey(),
				}
				if x.GetChangeType() != datasync.Delete {
					kv.Val, err = UnmarshalLazyValue(kv.Key, x)
					if err != nil {
						p.log.Errorf("decoding value for key %q failed: %v", kv.Key, err)
						continue
					}
				}
				kvPairs = append(kvPairs, kv)
			}

			if len(kvPairs) == 0 {
				p.log.Warn("no valid kv pairs received in change event")
				e.Done(nil)
				continue
			}

			p.log.Debugf("Change with %d items", len(kvPairs))

			ctx := e.GetContext()
			if ctx == nil {
				ctx = context.Background()
			}
			_, withDataSrc := DataSrcFromContext(ctx)
			if !withDataSrc {
				ctx = DataSrcContext(ctx, "datasync")
			}
			ctx = kvs.WithRetryDefault(ctx)

			_, err = p.PushData(ctx, kvPairs)
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
			_, withDataSrc := DataSrcFromContext(ctx)
			if !withDataSrc {
				ctx = DataSrcContext(ctx, "datasync")
			}
			ctx = kvs.WithResync(ctx, kvs.FullResync, true)
			ctx = kvs.WithRetryDefault(ctx)

			_, err := p.PushData(ctx, kvPairs)
			e.Done(err)

		}
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

func (p *Plugin) watchStatus(ch <-chan *kvscheduler.BaseValueStatus) {
	for {
		select {
		case s := <-ch:

			p.debugf("STATUS: %15s %v ===> %v (%v) %v",
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
