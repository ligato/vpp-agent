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

package vppcallmock

import (
	"testing"

	"git.fd.io/govpp.git/adapter/mock"
	govppapi "git.fd.io/govpp.git/api"
	"git.fd.io/govpp.git/core"
	. "github.com/onsi/gomega"
)

//TestCtx is helping structure for unit testing. It wraps VppAdapter which is used instead of real VPP
type TestCtx struct {
	MockVpp *mock.VppAdapter
	conn    *core.Connection
	Channel *govppapi.Channel
}

//SetupTestCtx sets up all fields of TestCtx structure at the begining of test
func SetupTestCtx(t *testing.T) *TestCtx {
	RegisterTestingT(t)

	ctx := &TestCtx{}
	ctx.MockVpp = &mock.VppAdapter{}

	var err error
	ctx.conn, err = core.Connect(ctx.MockVpp)
	Expect(err).ShouldNot(HaveOccurred())

	ctx.Channel, err = ctx.conn.NewAPIChannel()
	Expect(err).ShouldNot(HaveOccurred())

	return ctx
}

//TeardownTestCtx politely close all used resources
func (ctx *TestCtx) TeardownTestCtx() {
	ctx.Channel.Close()
	ctx.conn.Disconnect()
}
