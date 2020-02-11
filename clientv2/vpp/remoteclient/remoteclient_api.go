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
	"go.ligato.io/cn-infra/v2/db/keyval"

	"go.ligato.io/vpp-agent/v3/clientv2/vpp"
	"go.ligato.io/vpp-agent/v3/clientv2/vpp/dbadapter"
	//"github.com/ligato/vpp-agent/clientv2/vpp/grpcadapter"
	//"github.com/ligato/vpp-agent/plugins/vpp/model/rpc"
)

// DataResyncRequestDB allows creating a RESYNC request using convenient RESYNC
// DSL and sending it through the provided <broker>.
// User of the API does not need to be aware of keys.
// User of the API does not need to delete the obsolete objects/keys
// prior to RESYNC - it is handled by DataResyncDSL.
func DataResyncRequestDB(broker keyval.ProtoBroker) vppclient.DataResyncDSL {
	return dbadapter.NewDataResyncDSL(broker.NewTxn(), broker.ListKeys)
}

// DataChangeRequestDB allows createing Data Change requests using convenient
// Data Change DSL and sending it through the provided <broker>.
// User of the API does not need to be aware of keys.
func DataChangeRequestDB(broker keyval.ProtoBroker) vppclient.DataChangeDSL {
	return dbadapter.NewDataChangeDSL(broker.NewTxn())
}

// TODO: GRPC TBD
/*
// DataResyncRequestGRPC allows sending RESYNC requests conveniently.
// User of the API does not need to be aware of keys.
// User of the API does not need to delete the obsolete objects/keys during RESYNC.
func DataResyncRequestGRPC(client rpc.DataResyncServiceClient) vppclient.DataResyncDSL {
	return grpcadapter.NewDataResyncDSL(client)
}

// DataChangeRequestGRPC allows sending Data Change requests conveniently (even without directly using Broker).
// User of the API does not need to be aware of keys.
func DataChangeRequestGRPC(client rpc.DataChangeServiceClient) vppclient.DataChangeDSL {
	return grpcadapter.NewDataChangeDSL(client)
}
*/
