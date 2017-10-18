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
	"github.com/ligato/vpp-agent/clientv1/defaultplugins"
	"github.com/ligato/vpp-agent/clientv1/defaultplugins/dbadapter"
	"github.com/ligato/vpp-agent/flavors/rpc/model/vppsvc"
	"github.com/ligato/vpp-agent/clientv1/defaultplugins/grpcadapter"
)

// DataResyncRequestDB allows to create a RESYNC request using convenient RESYNC
// DSL and send it through the provided <broker>.
// User of the API does not need to be aware of keys.
// User of the API does not need to delete the obsolete objects/keys
// prior to RESYNC - it is handled by DataResyncDSL.
func DataResyncRequestDB(broker keyval.ProtoBroker) defaultplugins.DataResyncDSL {
	return dbadapter.NewDataResyncDSL(broker.NewTxn(), broker.ListKeys)
}

// DataChangeRequestDB allows to create Data Change requests using convenient
// Data Change DSL and send it through the provided <broker>.
// User of the API does not need to be aware of keys.
func DataChangeRequestDB(broker keyval.ProtoBroker) defaultplugins.DataChangeDSL {
	return dbadapter.NewDataChangeDSL(broker.NewTxn())
}

// DataResyncRequestGRPC allows to send RESYNC requests conveniently.
// User of the API does not need to be aware of keys.
// User of the API does not need do by himself the delete of obsolete objects/keys during RESYNC.
func DataResyncRequestGRPC(client vppsvc.ResyncConfigServiceClient) defaultplugins.DataResyncDSL {
	return grpcadapter.NewDataResyncDSL(client)
}

// DataChangeRequestGRPC allows to send Data Change requests conveniently (even without directly using Broker)
// User of the API does not need to be aware of keys.
func DataChangeRequestGRPC(client vppsvc.ChangeConfigServiceClient) defaultplugins.DataChangeDSL {
	return grpcadapter.NewDataChangeDSL(client)
}
