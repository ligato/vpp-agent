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

package datasync

import (
	"github.com/golang/protobuf/proto"
	"sync"

	"github.com/ligato/cn-infra/utils/safeclose"
)

var (
	defaultTransport = &compositeTransport{[]TransportAdapter{}}
	access           sync.Mutex

	factoryOfDifferentAgent func(microserviceLabel string) TransportAdapter
)

type nilTransportAdapter struct{}

// WatchData using ETCD or any other data transport
func (*nilTransportAdapter) WatchData(resyncName string, changeChan chan ChangeEvent, resyncChan chan ResyncEvent,
	keyPrefixes ...string) (WatchDataRegistration, error) {
	return &nilReg{}, nil
}

type nilReg struct{}

func (*nilReg) Close() error {
	return nil
}

// PublishData to ETCD or any other data transport (from other Agent Plugins)
func (*nilTransportAdapter) PublishData(key string, data proto.Message) error {
	return nil
}

// RegisterTransport adds transport to the slice of registered default transports
func RegisterTransport(transport TransportAdapter) error {
	access.Lock()
	defer access.Unlock()

	defaultTransport.transports = append(defaultTransport.transports, transport)

	return nil
}

// UnregisterTransport clears the registered transport. Used in tests where the agent is started multiple
// times in one process and new transport needs to be registered.
func UnregisterTransport() {
	access.Lock()
	defer access.Unlock()

	defaultTransport = &compositeTransport{[]TransportAdapter{}}
}

// GetTransport returns previously registered transport by func RegisterTransport
func GetTransport() TransportAdapter {
	access.Lock()
	defer access.Unlock()

	return defaultTransport
}

// OfDifferentAgent allows access DB of different agent
func OfDifferentAgent(microserviceLabel string) TransportAdapter {
	access.Lock()
	defer access.Unlock()

	return factoryOfDifferentAgent(microserviceLabel)
}

// RegisterTransportOfDifferentAgent is similar to RegisterTransport
func RegisterTransportOfDifferentAgent(factory func(microserviceLabel string) TransportAdapter) error {
	access.Lock()
	defer access.Unlock()

	factoryOfDifferentAgent = factory

	return nil
}

type compositeTransport struct {
	transports []TransportAdapter
}

func (x *compositeTransport) WatchData(resyncName string, changeChan chan ChangeEvent, resyncChan chan ResyncEvent,
	keyPrefixes ...string) (WatchDataRegistration, error) {
	access.Lock()
	defer access.Unlock()

	partialRegs := []WatchDataRegistration{}
	for _, transport := range x.transports {
		reg, err := transport.WatchData(resyncName, changeChan, resyncChan, keyPrefixes...)
		if err != nil {
			return nil, err
		}
		partialRegs = append(partialRegs, reg)
	}

	return &compositeWatchDataRegistration{partialRegs}, nil
}

// PublishData to all registered transports
func (x *compositeTransport) PublishData(key string, data proto.Message) error {
	for _, transport := range x.transports {
		err := transport.PublishData(key, data)
		if err != nil {
			return err
		}
	}
	return nil
}

type compositeWatchDataRegistration struct {
	partialRegs []WatchDataRegistration
}

func (r *compositeWatchDataRegistration) Close() error {
	var wasError error
	for _, reg := range r.partialRegs {
		err := safeclose.Close(reg)
		if err != nil {
			wasError = err
		}
	}

	return wasError
}
