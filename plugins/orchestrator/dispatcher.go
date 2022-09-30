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
	"context"
	"fmt"
	"runtime/trace"
	"sync"
	"time"

	"github.com/pkg/errors"
	"go.ligato.io/cn-infra/v2/logging"
	"google.golang.org/protobuf/proto"

	"go.ligato.io/vpp-agent/v3/pkg/models"
	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/plugins/orchestrator/contextdecorator"
	"go.ligato.io/vpp-agent/v3/proto/ligato/kvscheduler"
)

// KeyVal associates value with its key.
type KeyVal struct {
	Key string
	Val proto.Message
}

// KVPairs represents key-value pairs.
type KVPairs map[string]proto.Message

// Label is string key-value pair associated with configuration item.
// Label key format guidelines: label key should be a lower-case alphanumeric string
// which may contain periods and hyphens (but it should not contain consecutive
// periods/hyphens and it should not start with period/hyphen). Labels for configuration
// items should be prefixed with the reverse DNS notation of a domain they originate from
// (with domain owner's permission) for example: com.example.foo-bar-label.
// The io.ligato.* and ligato.* prefixes are reserved by vpp-agent for internal use.
type Labels map[string]string

type Status = kvscheduler.ValueStatus

type Result struct {
	Key    string
	Status *Status
}

type Dispatcher interface {
	ListData() KVPairs
	PushData(context.Context, []KeyVal, map[string]Labels) ([]Result, error)
	GetStatus(key string) (*Status, error)
	ListState() (KVPairs, error)
	ListLabels(key string) Labels
}

type dispatcher struct {
	log logging.Logger
	kvs kvs.KVScheduler
	mu  sync.Mutex
	db  Store
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
func (p *dispatcher) PushData(ctx context.Context, kvPairs []KeyVal, keyLabels map[string]Labels) (results []Result, err error) {
	trace.Logf(ctx, "pushData", "%d KV pairs", len(kvPairs))

	// check key-value pairs for uniqness and validate key
	uniq := make(map[string]proto.Message)
	for _, kv := range kvPairs {
		if kv.Val != nil {
			// check if given key matches the key generated from value
			if k := models.Key(kv.Val); k != kv.Key {
				return nil, errors.Errorf("given key %q does not match with key generated from value: %q (value: %#v)", kv.Key, k, kv.Val)
			}
		}
		// check if key is unique
		if oldVal, ok := uniq[kv.Key]; ok {
			return nil, errors.Errorf("found multiple key-value pairs with same key: %q (value 1: %#v, value 2: %#v)", kv.Key, kv.Val, oldVal)
		}
		uniq[kv.Key] = kv.Val
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	pr := trace.StartRegion(ctx, "prepare kv data")

	dataSrc, ok := contextdecorator.DataSrcFromContext(ctx)
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
			p.db.ResetLabels(kv.Key)
			for lkey, lval := range keyLabels[kv.Key] {
				p.db.AddLabel(kv.Key, lkey, lval)
			}
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
				for lkey := range keyLabels[kv.Key] {
					p.db.DeleteLabel(kv.Key, lkey)
				}
			} else {
				p.log.Debugf(" - UPDATE: %q ", kv.Key)
				txn.SetValue(kv.Key, kv.Val)
				p.db.Update(dataSrc, kv.Key, kv.Val)
				p.db.ResetLabels(kv.Key)
				for lkey, lval := range keyLabels[kv.Key] {
					p.db.AddLabel(kv.Key, lkey, lval)
				}
			}
		}
	}

	pr.End()

	t := time.Now()

	seqID, err := txn.Commit(ctx)
	p.kvs.TransactionBarrier()
	results = append(results, Result{
		Key: "seqnum",
		Status: &Status{
			Details: []string{fmt.Sprint(seqID)},
		},
	})
	for key := range uniq {
		s := p.kvs.GetValueStatus(key)
		results = append(results, Result{
			Key:    key,
			Status: s.GetValue(),
		})
	}
	if err != nil {
		if txErr, ok := err.(*kvs.TransactionError); ok && len(txErr.GetKVErrors()) > 0 {
			kvErrs := txErr.GetKVErrors()
			var errInfo = ""
			for i, kvErr := range kvErrs {
				errInfo += fmt.Sprintf(" - %3d. error (%s) %s - %v\n", i+1, kvErr.TxnOperation, kvErr.Key, kvErr.Error)
			}
			p.log.Errorf("Transaction #%d finished with %d errors\n%s", seqID, len(kvErrs), errInfo)
		} else {
			p.log.Errorf("Transaction failed: %v", err)
			return nil, err
		}
		return results, err
	} else {
		took := time.Since(t)
		p.log.Infof("Transaction #%d successful! (took %v)", seqID, took.Round(time.Microsecond*100))
	}

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
			// status := p.kvs.GetValueStatus(d.Key)
			pairs[d.Key] = d.Value
		}
	}

	return pairs, nil
}

func (p *dispatcher) ListLabels(key string) Labels {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.db.ListLabels(key)
}
