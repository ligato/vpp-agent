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
	"fmt"
	"runtime/trace"
	"sync"
	"time"

	"github.com/ligato/cn-infra/logging"
	"golang.org/x/net/context"

	"github.com/gogo/protobuf/proto"
	kvs "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
)

// KeyVal associates value with its key.
type KeyVal struct {
	Key string
	Val proto.Message
}

// KVPairs represents key-value pairs.
type KVPairs map[string]proto.Message

type Dispatcher interface {
	ListData() KVPairs
	PushData(context.Context, []KeyVal) ([]kvs.KeyWithError, error)
}

type dispatcher struct {
	log   logging.Logger
	kvs   kvs.KVScheduler
	mu    sync.Mutex
	store *memStore
}

// ListData retrieves actual data.
func (p *dispatcher) ListData() KVPairs {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.store.ListAll()
}

// PushData updates actual data.
func (p *dispatcher) PushData(ctx context.Context, kvPairs []KeyVal) (kvErrs []kvs.KeyWithError, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	trace.Logf(ctx, "kvPairs", "%d", len(kvPairs))

	pr := trace.StartRegion(ctx, "prepare kv data")

	dataSrc, ok := DataSrcFromContext(ctx)
	if !ok {
		dataSrc = "global"
	}

	p.log.Debugf("Pushing data with %d KV pairs (source: %s)", len(kvPairs), dataSrc)

	txn := p.kvs.StartNBTransaction()

	if typ, _ := kvs.IsResync(ctx); typ == kvs.FullResync {
		trace.Log(ctx, "resyncType", typ.String())
		p.store.Reset(dataSrc)
		for _, kv := range kvPairs {
			if kv.Val == nil {
				p.log.Debugf(" - PUT: %q (skipped nil value)", kv.Key)
				continue
			}
			p.log.Debugf(" - PUT: %q ", kv.Key)
			p.store.Update(dataSrc, kv.Key, kv.Val)
		}
		allPairs := p.store.ListAll()
		p.log.Debugf("will resync %d pairs", len(allPairs))
		for k, v := range allPairs {
			txn.SetValue(k, v)
		}
	} else {
		for _, kv := range kvPairs {
			if kv.Val == nil {
				p.log.Debugf(" - DELETE: %q", kv.Key)
				txn.SetValue(kv.Key, nil)
				p.store.Delete(dataSrc, kv.Key)
			} else {
				p.log.Debugf(" - UPDATE: %q ", kv.Key)
				txn.SetValue(kv.Key, kv.Val)
				p.store.Update(dataSrc, kv.Key, kv.Val)
			}
		}
	}

	pr.End()

	t := time.Now()

	seqID, err := txn.Commit(ctx)
	if err != nil {
		if txErr, ok := err.(*kvs.TransactionError); ok && len(txErr.GetKVErrors()) > 0 {
			kvErrs = txErr.GetKVErrors()
			var errInfo = ""
			for i, kvErr := range kvErrs {
				errInfo += fmt.Sprintf(" - %3d. error (%s) %s - %v\n", i+1, kvErr.TxnOperation, kvErr.Key, kvErr.Error)
			}
			p.log.Errorf("Transaction #%d finished with %d errors", seqID, len(kvErrs))
			fmt.Println(errInfo)
		} else {
			p.log.Errorf("Transaction failed: %v", err)
		}
		return kvErrs, err
	}

	took := time.Since(t).Round(time.Microsecond * 100)
	p.log.Infof("Transaction #%d successful! (took %v)", seqID, took)

	return nil, nil
}

type dataSrcKeyT string

var dataSrcKey = dataSrcKeyT("dataSrc")

func DataSrcContext(ctx context.Context, dataSrc string) context.Context {
	return context.WithValue(ctx, dataSrcKey, dataSrc)
}

func DataSrcFromContext(ctx context.Context) (dataSrc string, ok bool) {
	dataSrc, ok = ctx.Value(dataSrcKey).(string)
	return
}
