//  Copyright (c) 2018 Cisco and/or its affiliates.
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
	"time"

	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/datasync/kvdbsync/local"
	"github.com/ligato/cn-infra/infra"
	"github.com/ligato/cn-infra/rpc/grpc"
	"golang.org/x/net/context"

	"github.com/ligato/vpp-agent/api"
	kvs "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
)

// Registry is used for propagating transactions.
var Registry = local.DefaultRegistry

// Plugin implements sync service for GRPC.
type Plugin struct {
	Deps

	grpcSvc *grpcService

	// datasync channels
	changeChan   chan datasync.ChangeEvent
	resyncChan   chan datasync.ResyncEvent
	watchDataReg datasync.WatchRegistration
}

// Deps represents dependencies for the plugin.
type Deps struct {
	infra.PluginDeps

	GRPC        grpc.Server
	KVScheduler kvs.KVScheduler
	Watcher     datasync.KeyValProtoWatcher
}

// Init registers the service to GRPC server.
func (p *Plugin) Init() error {
	// initialize datasync channels
	p.resyncChan = make(chan datasync.ResyncEvent)
	p.changeChan = make(chan datasync.ChangeEvent)

	// register grpc service
	p.grpcSvc = &grpcService{p.Log}
	api.RegisterSyncServiceServer(p.GRPC.GetServer(), p.grpcSvc)

	return nil
}

// AfterInit subscribes to known NB prefixes.
func (p *Plugin) AfterInit() error {
	go p.watchEvents()

	var err error
	p.watchDataReg, err = p.Watcher.Watch("scheduler",
		p.changeChan, p.resyncChan, p.KVScheduler.GetRegisteredNBKeyPrefixes()...)
	if err != nil {
		return err
	}

	return nil
}

func (p *Plugin) watchEvents() {
	for {
		select {
		case e := <-p.changeChan:
			p.Log.Debugf("=> SCHEDULER received CHANGE EVENT: %v changes", len(e.GetChanges()))

			txn := p.KVScheduler.StartNBTransaction()
			for _, x := range e.GetChanges() {
				p.Log.Debugf("  - Change %v: %q (rev: %v)",
					x.GetChangeType(), x.GetKey(), x.GetRevision())
				if x.GetChangeType() == datasync.Delete {
					txn.SetValue(x.GetKey(), nil)
				} else {
					txn.SetValue(x.GetKey(), x)
				}
			}
			kvErrs, err := txn.Commit(kvs.WithRetry(context.Background(), time.Second, true))
			p.Log.Debugf("commit result: err=%v kvErrs=%+v", err, kvErrs)
			e.Done(err)

		case e := <-p.resyncChan:
			p.Log.Debugf("=> SCHEDULER received RESYNC EVENT: %v prefixes", len(e.GetValues()))

			txn := p.KVScheduler.StartNBTransaction()
			for prefix, iter := range e.GetValues() {
				var keyVals []datasync.KeyVal
				for x, done := iter.GetNext(); !done; x, done = iter.GetNext() {
					keyVals = append(keyVals, x)
					txn.SetValue(x.GetKey(), x)
				}
				p.Log.Debugf(" - Resync: %q (%v key-values)", prefix, len(keyVals))
				for _, x := range keyVals {
					p.Log.Debugf("\t%q: (rev: %v)", x.GetKey(), x.GetRevision())
				}
			}
			ctx := context.Background()
			ctx = kvs.WithRetry(ctx, time.Second, true)
			ctx = kvs.WithResync(ctx, kvs.FullResync, true)
			kvErrs, err := txn.Commit(ctx)
			p.Log.Debugf("commit result: err=%v kvErrs=%+v", err, kvErrs)
			e.Done(err)
		}
	}
}
