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

package client

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
	"go.ligato.io/cn-infra/v2/datasync"
	"go.ligato.io/cn-infra/v2/datasync/kvdbsync/local"
	"go.ligato.io/cn-infra/v2/datasync/syncbase"
	"google.golang.org/protobuf/proto"

	"go.ligato.io/vpp-agent/v3/pkg/models"
	"go.ligato.io/vpp-agent/v3/pkg/util"
	"go.ligato.io/vpp-agent/v3/plugins/orchestrator"
	"go.ligato.io/vpp-agent/v3/plugins/orchestrator/contextdecorator"
	"go.ligato.io/vpp-agent/v3/proto/ligato/generic"
)

// LocalClient is global client for direct local access.
// Updates and resyncs of this client use local.DefaultRegistry for propagating data to orchestrator.Dispatcher
// (going through watcher.Aggregator together with other data sources). However, data retrieval uses
// orchestrator.Dispatcher directly.
var LocalClient = NewClient(local.DefaultRegistry, &orchestrator.DefaultPlugin)

type client struct {
	registry   *syncbase.Registry
	dispatcher orchestrator.Dispatcher
}

// NewClient returns new instance that uses given registry for data propagation and dispatcher for data retrieval.
func NewClient(registry *syncbase.Registry, dispatcher orchestrator.Dispatcher) ConfigClient {
	return &client{
		registry:   registry,
		dispatcher: dispatcher,
	}
}

func (c *client) KnownModels(class string) ([]*ModelInfo, error) {
	var modules []*ModelInfo
	for _, model := range models.RegisteredModels() {
		if class == "" || model.Spec().Class == class {
			modules = append(modules, &models.ModelInfo{
				ModelDetail:       model.ModelDetail(),
				MessageDescriptor: model.NewInstance().ProtoReflect().Descriptor(),
			})
		}
	}
	return modules, nil
}

func (c *client) ResyncConfig(msgs ...proto.Message) error {
	txn := c.newLazyValTxn(true)

	uis := ProtosToUpdateItems(msgs)
	for _, ui := range uis {
		key, err := models.GetKey(ui.Message)
		if err != nil {
			return err
		}
		txn.Put(key, ui)
	}

	ctx := context.Background()
	ctx = contextdecorator.DataSrcContext(ctx, "localclient")
	return txn.Commit(ctx)
}

func (c *client) GetFilteredConfig(filter Filter, dsts ...interface{}) error {
	if filter.Ids != nil && filter.Labels != nil {
		return fmt.Errorf("both fields of the filter are not nil!")
	}
	protos := c.dispatcher.ListData()
	for key, data := range protos {
		item, err := models.MarshalItem(data)
		if err != nil {
			return err
		}
		labels := c.dispatcher.ListLabels(key)
		if !orchestrator.HasCorrectLabels(filter.Labels, labels) ||
			!orchestrator.ContainsItemID(filter.Ids, item.Id) {
			delete(protos, key)
		}
	}
	protoDsts := extractProtoMessages(dsts)
	if len(dsts) == len(protoDsts) { // all dsts are proto messages
		// TODO the clearIgnoreLayerCount function argument should be a option of generic.Client
		//  (the value 1 generates from dynamic config the same json/yaml output as the hardcoded
		//  configurator.Config and therefore serves for backward compatibility)
		util.PlaceProtosIntoProtos(protoMapToList(protos), 1, protoDsts...)
	} else {
		util.PlaceProtos(protos, dsts...)
	}
	return nil
}

func (c *client) GetConfig(dsts ...interface{}) error {
	return c.GetFilteredConfig(Filter{}, dsts...)
}

func (c *client) GetItems(ctx context.Context) ([]*ConfigItem, error) {
	var configItems []*ConfigItem
	for key, data := range c.dispatcher.ListData() {
		labels := c.dispatcher.ListLabels(key)
		item, err := models.MarshalItem(data)
		if err != nil {
			return nil, err
		}
		var itemStatus *generic.ItemStatus
		status, err := c.dispatcher.GetStatus(key)
		if err != nil {
			logrus.Warnf("GetStatus failed: %v", err)
		} else {
			var msg string
			if details := status.GetDetails(); len(details) > 0 {
				msg = strings.Join(status.GetDetails(), ", ")
			} else {
				msg = status.GetError()
			}
			itemStatus = &generic.ItemStatus{
				Status:  status.GetState().String(),
				Message: msg,
			}
		}
		configItems = append(configItems, &ConfigItem{
			Item:   item,
			Status: itemStatus,
			Labels: labels,
		})
	}
	return configItems, nil
}

func (c *client) UpdateItems(ctx context.Context, items []UpdateItem, resync bool) ([]*UpdateResult, error) {
	// TODO: use grpc client to update items with labels
	return nil, nil
}

func (c *client) DeleteItems(ctx context.Context, items []UpdateItem) ([]*UpdateResult, error) {
	// TODO: use grpc client to delete items with labels
	return nil, nil
}

func (c *client) DumpState() ([]*generic.StateItem, error) {
	// TODO: use dispatcher to dump state
	return nil, nil
}

func (c *client) ChangeRequest() ChangeRequest {
	return &changeRequest{itemChange: itemChange{txn: c.newLazyValTxn(false)}}
}

func (c *client) NewItemChange() ItemChange {
	return &itemChange{txn: c.newLazyValTxn(false)}
}

func (c *client) newLazyValTxn(resync bool) *LazyValTxn {
	if resync {
		return NewLazyValTxn(c.registry.PropagateResync)
	}
	return NewLazyValTxn(c.registry.PropagateChanges)
}

type itemChange struct {
	txn *LazyValTxn
	err error
}

func (r *itemChange) update(delete bool, items ...UpdateItem) *itemChange {
	if r.err != nil {
		return r
	}
	for _, item := range items {
		key, err := models.GetKey(item.Message)
		if err != nil {
			r.err = err
			return r
		}
		if delete {
			r.txn.Delete(key)
		} else {
			r.txn.Put(key, item)
		}
	}
	return r
}

func (r *itemChange) Update(items ...UpdateItem) ItemChange {
	return r.update(false, items...)
}

func (r *itemChange) Delete(items ...UpdateItem) ItemChange {
	return r.update(true, items...)
}

func (r *itemChange) Send(ctx context.Context) error {
	if r.err != nil {
		return r.err
	}
	_, withDataSrc := contextdecorator.DataSrcFromContext(ctx)
	if !withDataSrc {
		ctx = contextdecorator.DataSrcContext(ctx, "localclient")
	}
	return r.txn.Commit(ctx)
}

type changeRequest struct {
	itemChange
}

func (r *changeRequest) Update(msgs ...proto.Message) ChangeRequest {
	uis := ProtosToUpdateItems(msgs)
	r.itemChange = *r.update(false, uis...)
	return r
}

func (r *changeRequest) Delete(msgs ...proto.Message) ChangeRequest {
	uis := ProtosToUpdateItems(msgs)
	r.itemChange = *r.update(true, uis...)
	return r
}

func (r *changeRequest) Send(ctx context.Context) error {
	return r.itemChange.Send(ctx)
}

type LazyValTxn struct {
	mu      sync.Mutex
	changes map[string]datasync.ChangeValue
	commit  func(context.Context, map[string]datasync.ChangeValue) error
}

func NewLazyValTxn(commit func(context.Context, map[string]datasync.ChangeValue) error) *LazyValTxn {
	return &LazyValTxn{
		changes: make(map[string]datasync.ChangeValue),
		commit:  commit,
	}
}

func (txn *LazyValTxn) Put(key string, value datasync.LazyValue) *LazyValTxn {
	txn.mu.Lock()
	defer txn.mu.Unlock()

	txn.changes[key] = NewChangeLazy(key, value, 0, datasync.Put)
	return txn
}

func (txn *LazyValTxn) Delete(key string) *LazyValTxn {
	txn.mu.Lock()
	defer txn.mu.Unlock()

	txn.changes[key] = NewChangeLazy(key, nil, 0, datasync.Delete)
	return txn
}

func (txn *LazyValTxn) Commit(ctx context.Context) error {
	txn.mu.Lock()
	defer txn.mu.Unlock()

	return txn.commit(ctx, txn.changes)
}

func NewChangeLazy(key string, value datasync.LazyValue, rev int64, changeType datasync.Op) *syncbase.Change {
	// syncbase.Change does not export its changeType field so we set it first with syncbase.NewChange
	change := syncbase.NewChange("", nil, 0, changeType)
	change.KeyVal = syncbase.NewKeyVal(key, value, rev)
	return change
}

func ProtosToUpdateItems(msgs []proto.Message) []UpdateItem {
	var uis []UpdateItem
	for _, msg := range msgs {
		// resulting UpdateItem will have no labels
		uis = append(uis, UpdateItem{Message: msg})
	}
	return uis
}

func extractProtoMessages(dsts []interface{}) []proto.Message {
	msgs := make([]proto.Message, 0)
	for _, dst := range dsts {
		msg, ok := dst.(proto.Message)
		if ok {
			msgs = append(msgs, msg)
		} else {
			logrus.Debugf("at least one of the %d items is not proto message, but: %#v", len(dsts), dst)
			break
		}
	}
	return msgs
}

func protoMapToList(protoMap map[string]proto.Message) []proto.Message {
	result := make([]proto.Message, 0, len(protoMap))
	for _, msg := range protoMap {
		result = append(result, msg)
	}
	return result
}
