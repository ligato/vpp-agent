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

package grpcsync

import (
	"github.com/ligato/cn-infra/datasync"

	//TODO "github.com/gorilla/rpc/json"
	"fmt"
	"net"

	"github.com/ligato/cn-infra/datasync/syncbase"
	"github.com/ligato/cn-infra/datasync/syncbase/msg"
	"github.com/ligato/cn-infra/logging/logroot"
	"google.golang.org/grpc"
)

// NewAdapter creates a new instance of Adapter.
func NewAdapter() *Adapter {
	grpcServer := grpc.NewServer()
	//TODO grpcServer.RegisterCodec(json.NewCodec(), "application/json")
	adapter := &Adapter{syncbase.NewWatcher(), grpcServer}
	msg.RegisterDataMsgServiceServer(grpcServer, &DataMsgServiceServer{adapter})
	//registerHTTPHandler(grpcServer)

	go func() {
		//TODO Close the tcp listening
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 9192))
		if err != nil {
			logroot.StandardLogger().Error(err) //TODO
		}
		grpcServer.Serve(lis)
	}()
	return adapter
}

// Adapter is a GRPC transport adapter in front of Agent Plugins
type Adapter struct {
	base   *syncbase.Watcher
	server *grpc.Server
}

// Watch registers HTTP handlers - basically bridges them with local dbadapter
func (adapter *Adapter) Watch(resyncName string, changeChan chan datasync.ChangeEvent,
	resyncChan chan datasync.ResyncEvent, keyPrefixes ...string) (datasync.WatchRegistration, error) {

	logroot.StandardLogger().Debug("GRPC KeyValProtoWatcher WatchData ", resyncName, " ", keyPrefixes)

	return adapter.base.Watch(resyncName, changeChan, resyncChan, keyPrefixes...)
}

// Close closes the grpc server.
func (adapter *Adapter) Close() error {
	adapter.server.Stop()
	return nil
}
