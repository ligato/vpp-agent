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

package interfaces

// InterfaceStateNotificationType is a type of notification.
type InterfaceStateNotificationType int32

const (
	// UNKNOWN is default type.
	UNKNOWN InterfaceStateNotificationType = 0
	// UPDOWN represents Link UP/DOWN notification.
	UPDOWN InterfaceStateNotificationType = 1
	// COUNTERS represents interface state with updated counters.
	COUNTERS InterfaceStateNotificationType = 2
	// DELETED represents the event when the interface was deleted from the VPP.
	// Note that some north bound config updates require delete and create the network interface one more time.
	DELETED InterfaceStateNotificationType = 3
)

// InterfaceStateNotification aggregates status UP/DOWN/DELETED/UNKNOWN with
// the details (state) about the interfaces including counters.
type InterfaceStateNotification struct {
	// Type of the notification
	Type InterfaceStateNotificationType
	// State of the network interface
	State *InterfacesState_Interface
}
