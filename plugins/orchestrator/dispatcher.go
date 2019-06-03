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
	"github.com/pkg/errors"
	"golang.org/x/net/context"

	"github.com/gogo/protobuf/proto"
	"github.com/ligato/vpp-agent/pkg/models"
	kvs "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
)

// KeyVal associates value with its key.
type KeyVal struct {
	Key string
	Val proto.Message
}

// KVPairs represents key-value pairs.
type KVPairs map[string]proto.Message

type Status = kvs.ValueStatus

type Result struct {
	Key    string
	Status *Status
}

type Dispatcher interface {
	ListData() KVPairs
	PushData(context.Context, []KeyVal) ([]Result, error)
	GetStatus(key string) (*Status, error)
	ListState() (KVPairs, error)
}

type dispatcher struct {
	log logging.Logger
	kvs kvs.KVScheduler
	mu  sync.Mutex
	db  KVStore
}

// ListData retrieves actual data.
func (p *dispatcher) ListData() KVPairs {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.db.ListAll()
}

func (p *dispatcher) GetStatus(key string) (*Status, error) {
	s := p.kvs.GetValueStatus(key)
	status := s.GetValue()
	if status == nil {
		return nil, errors.Errorf("status for key %q not found", key)
	}
	return status, nil
}

// PushData updates actual data.
func (p *dispatcher) PushData(ctx context.Context, kvPairs []KeyVal) (results []Result, err error) {
	trace.Logf(ctx, "pushData", "%d KV pairs", len(kvPairs))

	// validate key-value pairs
	uniq := make(map[string]struct{})
	for _, kv := range kvPairs {
		if kv.Val != nil {
			// check if given key matches the key generated from value
			if k := models.Key(kv.Val); k != kv.Key {
				return nil, errors.Errorf("given key %q does not match with key generated from value: %q (value: %#v)", kv.Key, k, kv.Val)
			}
		}
		// check if key is unique
		if _, ok := uniq[kv.Key]; ok {
			return nil, errors.Errorf("found multiple key-value pairs with same key: ")
		}
		uniq[kv.Key] = struct{}{}
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	pr := trace.StartRegion(ctx, "prepare kv data")

	dataSrc, ok := DataSrcFromContext(ctx)
	if !ok {
		dataSrc = "global"
	}

	p.log.Debugf("Push data with %d KV pairs (source: %s)", len(kvPairs), dataSrc)

	txn := p.kvs.StartNBTransaction()

	if typ, _ := kvs.IsResync(ctx); typ == kvs.FullResync {
		trace.Log(ctx, "resyncType", typ.String())
		p.db.Reset(dataSrc)
		for _, kv := range kvPairs {
			if kv.Val == nil {
				p.log.Debugf(" - PUT: %q (skipped nil value for resync)", kv.Key)
				continue
			}
			p.log.Debugf(" - PUT: %q ", kv.Key)
			p.db.Update(dataSrc, kv.Key, kv.Val)
		}
		allPairs := p.db.ListAll()
		p.log.Debugf("will resync %d pairs", len(allPairs))
		for k, v := range allPairs {
			txn.SetValue(k, v)
		}
	} else {
		for _, kv := range kvPairs {
			if kv.Val == nil {
				p.log.Debugf(" - DELETE: %q", kv.Key)
				txn.SetValue(kv.Key, nil)
				p.db.Delete(dataSrc, kv.Key)
			} else {
				p.log.Debugf(" - UPDATE: %q ", kv.Key)
				txn.SetValue(kv.Key, kv.Val)
				p.db.Update(dataSrc, kv.Key, kv.Val)
			}
		}
	}

	pr.End()

	t := time.Now()

	seqID, err := txn.Commit(ctx)
	if err != nil {
		if txErr, ok := err.(*kvs.TransactionError); ok && len(txErr.GetKVErrors()) > 0 {
			kvErrs := txErr.GetKVErrors()
			var errInfo = ""
			for i, kvErr := range kvErrs {
				errInfo += fmt.Sprintf(" - %3d. error (%s) %s - %v\n", i+1, kvErr.TxnOperation, kvErr.Key, kvErr.Error)
			}
			p.log.Errorf("Transaction #%d finished with %d errors", seqID, len(kvErrs))
			fmt.Println(errInfo)
		} else {
			p.log.Errorf("Transaction failed: %v", err)
			return nil, err
		}
		return nil, err
	}

	p.kvs.TransactionBarrier()

	for key := range uniq {
		s := p.kvs.GetValueStatus(key)
		/*results = append(results, KeyVal{
			Key: key,
			Val: s.Value,
		})*/
		results = append(results, Result{
			Key:    key,
			Status: s.GetValue(),
		})
	}

	took := time.Since(t).Round(time.Microsecond * 100)
	p.log.Infof("Transaction #%d successful! (took %v)", seqID, took)

	return results, nil
}

// ListState retrieves running state.
func (p *dispatcher) ListState() (KVPairs, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	pairs := KVPairs{}
	for _, prefix := range p.kvs.GetRegisteredNBKeyPrefixes() {
		data, err := p.kvs.DumpValuesByKeyPrefix(prefix, kvs.CachedView)
		if err != nil {
			return nil, err
		}
		for _, d := range data {
			//status := p.kvs.GetValueStatus(d.Key)
			pairs[d.Key] = d.Value
		}
	}

	return pairs, nil
}
