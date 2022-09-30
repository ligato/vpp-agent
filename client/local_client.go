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

	"github.com/sirupsen/logrus"
	"go.ligato.io/cn-infra/v2/datasync/kvdbsync/local"
	"go.ligato.io/cn-infra/v2/datasync/syncbase"
	"go.ligato.io/cn-infra/v2/db/keyval"
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
var LocalClient = NewClient(&txnFactory{local.DefaultRegistry}, &orchestrator.DefaultPlugin)

type client struct {
	txnFactory ProtoTxnFactory
	dispatcher orchestrator.Dispatcher
}

// NewClient returns new instance that uses given registry for data propagation and dispatcher for data retrieval.
func NewClient(factory ProtoTxnFactory, dispatcher orchestrator.Dispatcher) ConfigClient {
	return &client{
		txnFactory: factory,
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

func (c *client) ResyncConfig(items ...proto.Message) error {
	txn := c.txnFactory.NewTxn(true)

	for _, item := range items {
		key, err := models.GetKey(item)
		if err != nil {
			return err
		}
		txn.Put(key, item)
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
		if !orchestrator.ContainsAllLabels(filter.Labels, labels) ||
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
	return c.GetFilteredConfig(Filter{}, dsts)
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
	txn := c.txnFactory.NewTxn(resync)
	for _, ui := range items {
		key, err := models.GetKey(ui.Message)
		if err != nil {
			return nil, err
		}
		txn.Put(key, ui.Message)
		_, withDataSrc := contextdecorator.DataSrcFromContext(ctx)
		if !withDataSrc {
			ctx = contextdecorator.DataSrcContext(ctx, "localclient")
		}
		ctx = contextdecorator.LabelsContext(ctx, ui.Labels)
	}
	if err := txn.Commit(ctx); err != nil {
		return nil, err
	}
	var updateResults []*UpdateResult
	r, _ := contextdecorator.PushDataResultFromContext(ctx)
	resWrapper, ok := r.(orchestrator.ResultWrapper)
	if !ok {
		return nil, fmt.Errorf("cannot retrieve update results!")
	}
	for _, res := range resWrapper.Results {
		var msg string
		if details := res.Status.GetDetails(); len(details) > 0 {
			msg = strings.Join(res.Status.GetDetails(), ", ")
		} else {
			msg = res.Status.GetError()
		}
		updateResults = append(updateResults, &UpdateResult{
			Key: res.Key,
			Status: &generic.ItemStatus{
				Status:  res.Status.State.String(),
				Message: msg,
			},
		})
	}
	return updateResults, nil
}

func (c *client) DeleteItems(ctx context.Context, items []UpdateItem) ([]*UpdateResult, error) {
	txn := c.txnFactory.NewTxn(false)
	for _, ui := range items {
		key, err := models.GetKey(ui.Message)
		if err != nil {
			return nil, err
		}
		txn.Delete(key)
		_, withDataSrc := contextdecorator.DataSrcFromContext(ctx)
		if !withDataSrc {
			ctx = contextdecorator.DataSrcContext(ctx, "localclient")
		}
	}
	if err := txn.Commit(ctx); err != nil {
		return nil, err
	}
	var updateResults []*UpdateResult
	r, _ := contextdecorator.PushDataResultFromContext(ctx)
	resWrapper, ok := r.(orchestrator.ResultWrapper)
	if !ok {
		return nil, fmt.Errorf("cannot retrieve update results!")
	}
	for _, res := range resWrapper.Results {
		var msg string
		if details := res.Status.GetDetails(); len(details) > 0 {
			msg = strings.Join(res.Status.GetDetails(), ", ")
		} else {
			msg = res.Status.GetError()
		}
		updateResults = append(updateResults, &UpdateResult{
			Key: res.Key,
			Status: &generic.ItemStatus{
				Status:  res.Status.State.String(),
				Message: msg,
			},
		})
	}
	return updateResults, nil
}

func (c *client) DumpState() ([]*generic.StateItem, error) {
	// TODO: use dispatcher to dump state
	return nil, nil
}

func (c *client) ChangeRequest() ChangeRequest {
	return &changeRequest{txn: c.txnFactory.NewTxn(false)}
}

type changeRequest struct {
	txn keyval.ProtoTxn
	err error
}

func (r *changeRequest) Update(items ...proto.Message) ChangeRequest {
	if r.err != nil {
		return r
	}
	for _, item := range items {
		key, err := models.GetKey(item)
		if err != nil {
			r.err = err
			return r
		}
		r.txn.Put(key, item)
	}
	return r
}

func (r *changeRequest) Delete(items ...proto.Message) ChangeRequest {
	if r.err != nil {
		return r
	}
	for _, item := range items {
		key, err := models.GetKey(item)
		if err != nil {
			r.err = err
			return r
		}
		r.txn.Delete(key)
	}
	return r
}

func (r *changeRequest) Send(ctx context.Context) error {
	if r.err != nil {
		return r.err
	}
	_, withDataSrc := contextdecorator.DataSrcFromContext(ctx)
	if !withDataSrc {
		ctx = contextdecorator.DataSrcContext(ctx, "localclient")
	}
	return r.txn.Commit(ctx)
}

// ProtoTxnFactory defines interface for keyval transaction provider.
type ProtoTxnFactory interface {
	NewTxn(resync bool) keyval.ProtoTxn
}

type txnFactory struct {
	registry *syncbase.Registry
}

func (p *txnFactory) NewTxn(resync bool) keyval.ProtoTxn {
	if resync {
		return local.NewProtoTxn(p.registry.PropagateResync)
	}
	return local.NewProtoTxn(p.registry.PropagateChanges)
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
