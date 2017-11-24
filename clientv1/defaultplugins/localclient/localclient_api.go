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

package localclient

import (
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/datasync/kvdbsync/local"
	"github.com/ligato/vpp-agent/clientv1/defaultplugins"
	"github.com/ligato/vpp-agent/clientv1/defaultplugins/dbadapter"
)

// PluginID defines the name of VPP (defaultplugins) localclient plugin.
const PluginID core.PluginName = "DefaultVppPlugins_LOCAL_CLIENT"

// DataResyncRequest allows creating a RESYNC request using convenient RESYNC
// DSL and sending it locally through go channels (i.e. without using Data Store).
func DataResyncRequest(caller core.PluginName) defaultplugins.DataResyncDSL {
	return dbadapter.NewDataResyncDSL(local.NewProtoTxn(local.Get().PropagateResync),
		nil /*no need to list anything*/)
}

// DataChangeRequest allows creating Data Change request(s) using convenient
// Data Change DSL and sending it locally through go channels (i.e. without using
// Data Store).
func DataChangeRequest(caller core.PluginName) defaultplugins.DataChangeDSL {
	return dbadapter.NewDataChangeDSL(local.NewProtoTxn(local.Get().PropagateChanges))
}
