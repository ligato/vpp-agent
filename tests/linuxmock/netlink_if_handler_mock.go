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
	"net"
)

// NetlinkHandlerMock allows to mock netlink-related methods
type NetlinkHandlerMock struct {
	responses []*whenIfResp
	respCurr  int
	respMax   int
}

// NewNetlinkHandlerMock creates new instance of mock and initializes response list
func NewNetlinkHandlerMock() *NetlinkHandlerMock {
	return &NetlinkHandlerMock{
		responses: make([]*whenIfResp, 0),
	}
}

// Helper struct with single method call and desired response items
type whenIfResp struct {
	methodName string
	items      []interface{}
}

// When defines name of the related method. It creates a new instance of whenIfResp with provided method name and
// stores it to the mock.
func (mock *NetlinkHandlerMock) When(name string) *whenIfResp {
	resp := &whenIfResp{
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
func (when *whenIfResp) ThenReturn(item ...interface{}) {
	when.items = item
}

// Auxiliary method returns next return value for provided method as generic type
func (mock *NetlinkHandlerMock) getReturnValues(name string) (response []interface{}) {
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

/* Mocked netlink handler methods */ //todo define other

func (mock *NetlinkHandlerMock) AddVethInterfacePair(ifName, peerIfName string) (retVal error) {
	return retVal
}

func (mock *NetlinkHandlerMock) DelVethInterfacePair(ifName, peerIfName string) (retVal error) {
	return retVal
}

func (mock *NetlinkHandlerMock) InterfaceAdminUp(ifName string) (retVal error) {
	items := mock.getReturnValues("InterfaceAdminUp")
	if len(items) >= 1 {
		return items[0].(error)
	}
	return nil
}

func (mock *NetlinkHandlerMock) InterfaceAdminDown(ifName string) (retVal error) {
	return retVal
}

func (mock *NetlinkHandlerMock) AddInterfaceIP(ifName string, addr *net.IPNet) (retVal error) {
	return retVal
}

func (mock *NetlinkHandlerMock) DelInterfaceIP(ifName string, addr *net.IPNet) (retVal error) {
	return retVal
}

func (mock *NetlinkHandlerMock) SetInterfaceMac(ifName string, macAddress string) (retVal error) {
	return retVal
}

func (mock *NetlinkHandlerMock) SetInterfaceMTU(ifName string, mtu int) (retVal error) {
	return retVal
}

func (mock *NetlinkHandlerMock) RenameInterface(ifName string, newName string) (retVal error) {
	return retVal
}

func (mock *NetlinkHandlerMock) GetLinkFromInterface(ifName string) (retVal1 netlink.Link, retVal2 error) {
	return
}

func (mock *NetlinkHandlerMock) GetAddressList(ifName string) (retVal1 []netlink.Addr, retVal2 error) {
	return retVal1, retVal2
}

func (mock *NetlinkHandlerMock) InterfaceExists(ifName string) (retVal1 bool, retVal2 error) {
	items := mock.getReturnValues("InterfaceExists")
	if len(items) == 1 {
		switch typed := items[0].(type) {
		case bool:
			return typed, nil
		case error:
			return false, typed
		}
	} else if len(items) == 2 {
		return items[0].(bool), items[1].(error)
	}
	return false, nil
}

func (mock *NetlinkHandlerMock) GetInterfaceType(ifName string) (retVal1 string, retVal2 error) {
	return retVal1, retVal2
}

func (mock *NetlinkHandlerMock) GetVethPeerName(ifName string) (retVal1 string, retVal2 error) {
	return retVal1, retVal2
}

func (mock *NetlinkHandlerMock) GetInterfaceByName(ifName string) (retVal1 *net.Interface, retVal2 error) {
	items := mock.getReturnValues("GetInterfaceByName")
	if len(items) == 1 {
		switch typed := items[0].(type) {
		case *net.Interface:
			return typed, nil
		case error:
			return nil, typed
		}
	} else if len(items) == 2 {
		return items[0].(*net.Interface), items[1].(error)
	}
	return nil, nil
}

func (mock *NetlinkHandlerMock) SetStopwatch(stopwatch *measure.Stopwatch) {}
