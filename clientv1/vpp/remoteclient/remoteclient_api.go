// Copyright (c) 2017 Cisco and/or its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package remoteclient

import (
	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/vpp-agent/clientv1/vpp"
	"github.com/ligato/vpp-agent/clientv1/vpp/dbadapter"
	"github.com/ligato/vpp-agent/clientv1/vpp/grpcadapter"
	"github.com/ligato/vpp-agent/plugins/vpp/model/rpc"
)

// DataRequest is reusable helper object which allows to create data resync/change structures. Returned objects are
// of DSL (domain specific language) type, and can be chained with configuration-type calls (Interface, BD, etc.)
// and sent at once
type DataRequest interface {
	// Resync creates a data resync call. All the configuration put to the call will be send to the target broker/service
	Resync() vppclient.DataResyncDSL
	// Change creates a data change call. The call defines 'put' and 'delete' followed by respective configuration
	// type calls
	Change() vppclient.DataChangeDSL
}

// dataRequestDB holds proto broker for remote database requests
type dataRequestDB struct {
	broker keyval.ProtoBroker
}

// dataRequestGRPC holds resync/change client objects and proto brokers (if used). Services manage GRPC calls and
// brokers can be used to store GRPC configuration
type dataRequestGRPC struct {
	clientResync rpc.DataResyncServiceClient
	clientChange rpc.DataChangeServiceClient
	brokers      []keyval.ProtoBroker
}

// NewDataRequestDB returns new data request broker for database
func NewDataRequestDB(broker keyval.ProtoBroker) DataRequest {
	return &dataRequestDB{
		broker: broker,
	}
}

// Resync for DB request
func (dr *dataRequestDB) Resync() vppclient.DataResyncDSL {
	return dbadapter.NewDataResyncDSL(dr.broker.NewTxn(), dr.broker.ListKeys)
}

// Change for DB request
func (dr *dataRequestDB) Change() vppclient.DataChangeDSL {
	return dbadapter.NewDataChangeDSL(dr.broker.NewTxn())
}

// NewDataRequestGRPC returns new data request broker for GRPC
func NewDataRequestGRPC(rsClient rpc.DataResyncServiceClient, chClient rpc.DataChangeServiceClient, brokers ...keyval.ProtoBroker) DataRequest {
	return &dataRequestGRPC{
		clientResync: rsClient,
		clientChange: chClient,
		brokers:      brokers,
	}
}

// Resync for GRPC request
func (dr *dataRequestGRPC) Resync() vppclient.DataResyncDSL {
	return grpcadapter.NewDataResyncDSL(dr.clientResync, dr.brokers)
}

// Change for GRPC request
func (dr *dataRequestGRPC) Change() vppclient.DataChangeDSL {
	return grpcadapter.NewDataChangeDSL(dr.clientChange, dr.brokers)
}

// DataDumpRequestGRPC allows sending 'Dump' data requests conveniently (even without directly using Broker).
// User of the API does not need to be aware of keys.
func DataDumpRequestGRPC(client rpc.DataDumpServiceClient) vppclient.DataDumpDSL {
	return grpcadapter.NewDataDumpDSL(client)
}
