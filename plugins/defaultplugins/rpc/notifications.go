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

//
const bufferSize = 100

// NotificationSvc forwards GRPC messages to external servers.
type NotificationSvc struct {
	mx sync.RWMutex

	// VPP notifications available for clients
	nBuffer [bufferSize]*rpc.NotificationsResponse
	nIdx    uint32

	log logging.Logger
}

// Get returns available VPP notifications in the same order as they were received
func (svc *NotificationSvc) Get(from *rpc.NotificationRequest, server rpc.NotificationService_GetServer) error {
	svc.mx.RLock()
	defer svc.mx.RUnlock()

	var msg string
	// If index was not provided, return all notifications in the buffer
	if from.Idx == 0 {
		from.Idx = 1 // reset index
	}

	// If desired index is older than buffer capacity, return all
	// available notifications and inform client about it
	if from.Idx > svc.nIdx%bufferSize {
		from.Idx = 1 // reset index
		msg = "Requested message exceeds notification buffer capacity"
	}

	var wasErr error
	for bufferIdx, entry := range svc.nBuffer {
		if entry == nil {
			continue
		}
		// Skip notifications which are older than desired
		if from.Idx >= entry.NextIdx {
			continue
		} else if  uint32(bufferIdx) <= svc.nIdx%bufferSize {
			entry.Message = msg
			if err := server.Send(entry); err != nil {
				svc.log.Error("Notification send error: %v", err)
				wasErr = err
			}
		}
	}
	return wasErr
}

// Adds new notification to the pool. The order of notifications is preserved
func (svc *NotificationSvc) updateNotifications(ctx context.Context, notification *interfaces.InterfaceNotification) {
	svc.mx.Lock()
	defer svc.mx.Unlock()

	svc.nBuffer[svc.nIdx%bufferSize] = &rpc.NotificationsResponse{
		NextIdx: svc.nIdx + 1,
		NIf:     notification,
	}
	svc.nIdx++
}
