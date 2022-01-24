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

func (c *client) GetConfig(dsts ...interface{}) error {
	protos := c.dispatcher.ListData()
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
