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
	"strconv"
	"sync"
)

// NotificationSvc forwards GRPC messages to external servers.
type NotificationSvc struct {
	mx sync.Mutex

	// VPP notifications available for clients
	notifications []*rpc.NotificationsResponse
	idxSeq        int

	log logging.Logger
}

// Get returns available VPP notifications in the same order as they were received
func (svc *NotificationSvc) Get(fromIdx *rpc.FromIndex, server rpc.NotificationService_GetServer) error {
	svc.mx.Lock()
	defer svc.mx.Unlock()

	for _, entry := range svc.notifications {
		// Get index of current notification.
		index, err := strconv.Atoi(entry.Index)
		if err != nil {
			svc.log.Error("Incorrect notification index: %s", entry.Index)
			continue
		}
		// Skip notifications which are older than desired
		if fromIdx.Index >= uint32(index) {
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
		Index:   strconv.Itoa(svc.idxSeq),
		IfNotif: notification,
	})
}
