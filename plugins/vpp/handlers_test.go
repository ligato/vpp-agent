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

package vpp_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"go.ligato.io/vpp-agent/v3/plugins/vpp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/vppmock"
)

type testHandlerAPI interface {
	A() error
}

type testHandler struct{}

func (t *testHandler) A() error {
	return nil
}

func TestRegisterHandler(t *testing.T) {
	c := vppmock.SetupTestCtx(t)
	vpp.ClearRegisteredHandlers()

	handler := vpp.RegisterHandler(vpp.HandlerDesc{
		Name:       "handlerA",
		HandlerAPI: (*testHandlerAPI)(nil),
	})

	Expect(handler).ToNot(BeNil())
	Expect(handler.Versions()).To(BeEmpty())
	Expect(handler.FindCompatibleVersion(c.MockVPPClient)).To(BeNil())
}

func TestRegisterHandlerVersions(t *testing.T) {
	c := vppmock.SetupTestCtx(t)
	vpp.ClearRegisteredHandlers()

	const (
		version vpp.Version = "19.08-test"
	)

	handler := vpp.RegisterHandler(vpp.HandlerDesc{
		Name:       "handlerA",
		HandlerAPI: (*testHandlerAPI)(nil),
	})
	handler.AddVersion(vpp.HandlerVersion{
		Version: version,
		Check: func(client vpp.Client) error {
			return nil
		},
		NewHandler: func(client vpp.Client, i ...interface{}) vpp.HandlerAPI {
			return &testHandler{}
		},
	})

	Expect(handler.Versions()).ToNot(BeEmpty())

	ver := handler.FindCompatibleVersion(c.MockVPPClient)
	Expect(ver).ToNot(BeNil())
	Expect(ver.Version).To(Equal(version))
	Expect(ver.NewHandler(c.MockVPPClient)).To(BeAssignableToTypeOf(&testHandler{}))
}
