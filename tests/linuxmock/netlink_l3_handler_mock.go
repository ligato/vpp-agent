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

package linuxmock

import (
	"github.com/ligato/cn-infra/logging/measure"
	"github.com/vishvananda/netlink"
)

// NetlinkHandlerMock allows to mock netlink-related methods
type L3NetlinkHandlerMock struct {
	responses []*whenL3Resp
	respCurr  int
	respMax   int
}

// NewNetlinkHandlerMock creates new instance of the mock and initializes response list
func NewL3NetlinkHandlerMock() *L3NetlinkHandlerMock {
	return &L3NetlinkHandlerMock{
		responses: make([]*whenL3Resp, 0),
	}
}

// Helper struct with single method call and desired response items
type whenL3Resp struct {
	methodName string
	items      []interface{}
}

// When defines name of the related method. It creates a new instance of whenL3Resp with provided method name and
// stores it to the mock.
func (mock *L3NetlinkHandlerMock) When(name string) *whenL3Resp {
	resp := &whenL3Resp{
		methodName: name,
	}
	mock.responses = append(mock.responses, resp)
	return resp
}

// ThenReturn receives array of items, which are desired to be returned in mocked method defined in "When". The full
// logic is:
// - When('someMethod').ThenReturn('values')
//
// Provided values should match return types of method. If method returns multiple values and only one is provided,
// mock tries to parse the value and returns it, while others will be nil or empty.
//
// If method is called several times, all cases must be defined separately, even if the return value is the same:
// - When('method1').ThenReturn('val1')
// - When('method1').ThenReturn('val1')
//
// All mocked methods are evaluated in same order they were assigned.
func (when *whenL3Resp) ThenReturn(item ...interface{}) {
	when.items = item
}

// Auxiliary method returns next return value for provided method as generic type
func (mock *L3NetlinkHandlerMock) getReturnValues(name string) (response []interface{}) {
	for i, resp := range mock.responses {
		if resp.methodName == name {
			// Remove used response but retain order
			mock.responses = append(mock.responses[:i], mock.responses[i+1:]...)
			return resp.items
		}
	}
	// Return empty response
	return
}

/* Mocked netlink handler methods */

func (mock *L3NetlinkHandlerMock) AddArpEntry(name string, arpEntry *netlink.Neigh) error {
	items := mock.getReturnValues("AddArpEntry")
	if len(items) >= 1 {
		return items[0].(error)
	}
	return nil
}

func (mock *L3NetlinkHandlerMock) SetArpEntry(name string, arpEntry *netlink.Neigh) error {
	items := mock.getReturnValues("SetArpEntry")
	if len(items) >= 1 {
		return items[0].(error)
	}
	return nil
}

func (mock *L3NetlinkHandlerMock) DelArpEntry(name string, arpEntry *netlink.Neigh) error {
	items := mock.getReturnValues("DelArpEntry")
	if len(items) >= 1 {
		return items[0].(error)
	}
	return nil
}

func (mock *L3NetlinkHandlerMock) GetArpEntries(interfaceIdx int, family int) ([]netlink.Neigh, error) {
	items := mock.getReturnValues("GetArpEntries")
	if len(items) == 1 {
		switch typed := items[0].(type) {
		case []netlink.Neigh:
			return typed, nil
		case error:
			return nil, typed
		}
	} else if len(items) == 2 {
		return items[0].([]netlink.Neigh), items[1].(error)
	}
	return nil, nil
}

func (mock *L3NetlinkHandlerMock) AddStaticRoute(name string, route *netlink.Route) error {
	items := mock.getReturnValues("AddStaticRoute")
	if len(items) >= 1 {
		return items[0].(error)
	}
	return nil
}

func (mock *L3NetlinkHandlerMock) ReplaceStaticRoute(name string, route *netlink.Route) error {
	items := mock.getReturnValues("ReplaceStaticRoute")
	if len(items) >= 1 {
		return items[0].(error)
	}
	return nil
}

func (mock *L3NetlinkHandlerMock) DelStaticRoute(name string, route *netlink.Route) error {
	items := mock.getReturnValues("DelStaticRoute")
	if len(items) >= 1 {
		return items[0].(error)
	}
	return nil
}

func (mock *L3NetlinkHandlerMock) SetStopwatch(stopwatch *measure.Stopwatch) {}
