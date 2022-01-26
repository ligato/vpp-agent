//  Copyright (c) 2019 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package configurator

import (
	"log"
	"sync"
	"sync/atomic"

	"github.com/google/go-cmp/cmp"
	"go.ligato.io/cn-infra/v2/logging"
	"google.golang.org/protobuf/proto"
	proto3 "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"

	pb "go.ligato.io/vpp-agent/v3/proto/ligato/configurator"
)

// Maximum number of messages stored in the buffer. Buffer is always filled from left
// to right (it means that if the buffer is full, a new entry is written to the index 0)
const bufferSize = 10000

// notifyService forwards GRPC messages to external servers.
type notifyService struct {
	log logging.Logger

	// VPP notifications available for clients
	mx     sync.RWMutex
	buffer [bufferSize]*pb.NotifyResponse
	curIdx uint32
	watchs chan *watcher
	notifs chan *pb.NotifyResponse
}

// Notify returns all required VPP notifications (or those available in the buffer) in the same order as they were received
func (svc *notifyService) Notify(from *pb.NotifyRequest, server pb.ConfiguratorService_NotifyServer) error {
	svc.mx.RLock()

	// Copy requested index locally
	fromIdx := from.Idx

	// Check if requested index overflows buffer length
	if svc.curIdx-from.Idx > bufferSize {
		fromIdx = svc.curIdx - bufferSize
	}

	// Start from requested index until the most recent entry
	for i := fromIdx; i < svc.curIdx; i++ {
		entry := svc.buffer[i%bufferSize]
		if !isFilter(entry.GetNotification(), from.Filters) {
			continue
		}
		if err := server.Send(entry); err != nil {
			svc.mx.RUnlock()
			svc.log.Warnf("Notify send error: %v", err)
			return err
		}
	}

	svc.mx.RUnlock()

	w := &watcher{
		notifs:  make(chan *pb.NotifyResponse),
		done:    make(chan struct{}),
		filters: from.Filters,
	}

	select {
	case svc.watchs <- w:
	case <-server.Context().Done():
		return server.Context().Err()
	}

	defer func() {
		close(w.done)
	}()

	for {
		select {
		case n := <-w.notifs:
			if err := server.Send(n); err != nil {
				svc.log.Warnf("Notify send error: %v", err)
				return err
			}
		case <-server.Context().Done():
			return server.Context().Err()
		}
	}
}

type watcher struct {
	notifs  chan *pb.NotifyResponse
	done    chan struct{}
	filters []*pb.Notification
}

// Pushes new notification to the buffer. The order of notifications is preserved.
func (svc *notifyService) pushNotification(notification *pb.Notification) {
	// notification is cloned to ensure it does not get changed after storing
	notifCopy := proto.Clone(notification).(*pb.Notification)

	// Notification index starts with 1
	idx := atomic.LoadUint32(&svc.curIdx)
	notif := &pb.NotifyResponse{
		NextIdx:      atomic.AddUint32(&svc.curIdx, 1),
		Notification: notifCopy,
	}
	svc.notifs <- notif

	svc.mx.Lock()
	defer svc.mx.Unlock()

	svc.buffer[idx%bufferSize] = notif
}

func (svc *notifyService) init() {
	svc.notifs = make(chan *pb.NotifyResponse)
	svc.watchs = make(chan *watcher)

	var watchers []*watcher

	go func() {
		for {
			select {
			case n := <-svc.notifs:
				for i, w := range watchers {
					select {
					case <-w.done:
						svc.log.Infof("removing watcher")
						copy(watchers[i:], watchers[i+1:])    // Shift a[i+1:] left one index.
						watchers[len(watchers)-1] = nil       // Erase last element (write zero value).
						watchers = watchers[:len(watchers)-1] // Truncate slice.
						continue
					default:
					}
					if !isFilter(n.GetNotification(), w.filters) {
						continue
					}
					select {
					case w.notifs <- n:
					default:
						svc.log.Infof("watcher full")
					}
				}

			case w := <-svc.watchs:
				watchers = append(watchers, w)
			}
		}
	}()
}

func isFilter(n *pb.Notification, filters []*pb.Notification) bool {
	if len(filters) == 0 {
		return true
	}

	for _, f := range filters {
		transforms := []cmp.Option{
			protocmp.IgnoreDefaultScalars(),
			protocmp.IgnoreEmptyMessages(),
			protocmp.Transform(),
		}
		dst := proto3.Clone(n)
		proto3.Merge(dst, f)
		if diff := cmp.Diff(dst, n, transforms...); diff != "" {
			log.Printf("diff mismatch (-want +got):\n%s", diff)
		} else {
			return true
		}
	}

	return false
}
