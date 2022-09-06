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

package govppmux

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"go.fd.io/govpp/core"

	"go.ligato.io/vpp-agent/v3/plugins/vpp/vppmock"
)

func TestRequestRetry(t *testing.T) {
	tests := []struct {
		name        string
		attempts    int
		timeout     time.Duration
		retErrs     []error
		expErr      error
		expAttempts int
	}{
		{name: "no retry fail",
			attempts:    0,
			timeout:     time.Millisecond,
			retErrs:     []error{core.ErrNotConnected},
			expErr:      core.ErrNotConnected,
			expAttempts: 0,
		},
		{name: "1 retry ok",
			attempts:    1,
			timeout:     time.Millisecond,
			retErrs:     []error{core.ErrNotConnected},
			expErr:      nil,
			expAttempts: 1,
		},
		{name: "3 retries fail",
			attempts:    3,
			timeout:     time.Millisecond,
			retErrs:     []error{core.ErrNotConnected, core.ErrNotConnected, core.ErrNotConnected, core.ErrNotConnected},
			expErr:      core.ErrNotConnected,
			expAttempts: 3,
		},
		{name: "no retry attempt",
			attempts:    3,
			timeout:     time.Millisecond,
			retErrs:     []error{core.ErrInvalidRequestCtx},
			expErr:      core.ErrInvalidRequestCtx,
			expAttempts: 0,
		},

		{name: "3 retries ok",
			attempts:    3,
			timeout:     time.Millisecond,
			retErrs:     []error{core.ErrNotConnected, core.ErrNotConnected, core.ErrNotConnected},
			expErr:      nil,
			expAttempts: 3,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := vppmock.SetupTestCtx(t)
			defer ctx.TeardownTestCtx()

			retryCfg := retryConfig{test.attempts, test.timeout}
			ch := newGovppChan(ctx.MockChannel, retryCfg)

			ctx.MockChannel.RetErrs = test.retErrs

			ctx.MockVpp.MockReply(&core.ControlPingReply{})
			for i := 0; i < test.attempts; i++ {
				ctx.MockVpp.MockReply(&core.ControlPingReply{})
			}

			req := &core.ControlPing{}
			reply := &core.ControlPingReply{}
			reqCtx := ch.SendRequest(req)
			err := reqCtx.ReceiveReply(reply)

			if test.expErr == nil {
				Expect(err).Should(Succeed())
			} else {
				Expect(err).Should(HaveOccurred())
			}
			Expect(ctx.MockChannel.Msgs).To(HaveLen(test.expAttempts + 1))
		})
	}
}
