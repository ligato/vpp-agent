// Copyright (c) 2018 Cisco and/or its affiliates.
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

package rpc

import (
	"context"
	"fmt"

	"github.com/ligato/cn-infra/logging"
	rpcServer "github.com/ligato/cn-infra/rpc/grpc"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/rpc"
)

// NotificationSvc forwards GRPC messages to external servers.
type NotificationSvc struct {
	GRPCClient rpcServer.Client

	clients []rpc.NotificationServiceClient
	log     logging.Logger
}

// ConnectEndpoints tries to make connection to every external GRPC server defined by address list
func (plugin *NotificationSvc) connectEndpoints(addresses []string) {
	if plugin.GRPCClient == nil || plugin.GRPCClient.IsDisabled() {
		plugin.log.Warn("failed to connect grpc endpoints, GRPC plugin not available")
		return
	}

	for _, address := range addresses {
		conn, err := plugin.GRPCClient.Connect(address)
		if err != nil {
			err := fmt.Errorf("failed to establish connection to %s: %v", address, err)
			plugin.log.Error(err)
			continue
		}

		client := rpc.NewNotificationServiceClient(conn)
		plugin.clients = append(plugin.clients, client)
		plugin.log.Debugf("Address %s registered for GRPC notifications", address)
	}
}

// Calls notification service to send data to every client
func (plugin *NotificationSvc) sendNotification(ctx context.Context, notification *interfaces.InterfaceNotification) {
	for _, client := range plugin.clients {
		client.Send(ctx, &rpc.Notifications{
			IfNotif: notification,
		})
	}
}
