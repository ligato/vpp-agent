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

package resynceventimpl

import (
	"github.com/ligato/cn-infra/datasync/resync/resyncevent"
	"sync"
	"time"
)

// Registration for Resync
type Registration struct {
	rwlock    sync.RWMutex
	waitStamp time.Time // last timestamp of WaitUntilReconciliationFinishes

	statusChan chan resyncevent.StatusEvent

	//waingForReconciliationChan chan *PluginEvent
	reconciliationFinishedChan chan time.Time
}

// NewRegistration is a constructor
func NewRegistration(statusChan chan resyncevent.StatusEvent) *Registration {
	return &Registration{statusChan: statusChan}
}

// StatusChan is here for Plugins to get channel for notifications about Resync status
func (reg *Registration) StatusChan() chan resyncevent.StatusEvent {
	return reg.statusChan
}
