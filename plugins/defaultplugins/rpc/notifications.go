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
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/interfaces"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/rpc"
	"sync"
)

// NotificationSvc forwards GRPC messages to external servers.
type NotificationSvc struct {
	mx sync.RWMutex

	// VPP notifications available for clients
	notifications []*rpc.NotificationsResponse
	idxSeq        uint32

	log logging.Logger
}

// Get returns available VPP notifications in the same order as they were received
func (svc *NotificationSvc) Get(fromIdx *rpc.FromIndex, server rpc.NotificationService_GetServer) error {
	svc.mx.RLock()
	defer svc.mx.RUnlock()

	for _, entry := range svc.notifications {
		// Skip notifications which are older than desired
		if fromIdx.Index >= entry.Index {
			continue
		}
		if err := server.Send(entry); err != nil {
			return err
		}
	}
	return nil
}

// Adds new notification to the pool. The order of notifications is preserved
func (svc *NotificationSvc) updateNotifications(ctx context.Context, notification *interfaces.InterfaceNotification) {
	svc.mx.Lock()
	defer svc.mx.Unlock()

	svc.idxSeq++
	svc.notifications = append(svc.notifications, &rpc.NotificationsResponse{
		Index:   svc.idxSeq,
		IfNotif: notification,
	})
}
