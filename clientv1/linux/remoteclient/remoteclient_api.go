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
	"github.com/ligato/vpp-agent/clientv1/linux"
	"github.com/ligato/vpp-agent/clientv1/linux/dbadapter"
)

// DataResyncRequestDB allows to send RESYNC requests conveniently.
// User of the API does not need to be aware of keys.
// User of the API does not need do by himself the delete of obsolete objects/keys during RESYNC.
func DataResyncRequestDB(broker keyval.ProtoBroker) linux.DataResyncDSL {
	return dbadapter.NewDataResyncDSL(broker.NewTxn(), broker.ListKeys)
}

// DataChangeRequestDB allows to send Data Change requests conveniently (even without directly using Broker)
// User of the API does not need to be aware of keys.
func DataChangeRequestDB(broker keyval.ProtoBroker) linux.DataChangeDSL {
	return dbadapter.NewDataChangeDSL(broker.NewTxn())
}
