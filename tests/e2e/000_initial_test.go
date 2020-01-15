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

package e2e

import (
	"testing"

	. "github.com/onsi/gomega"

	"go.ligato.io/vpp-agent/v3/proto/ligato/kvscheduler"
	ns "go.ligato.io/vpp-agent/v3/proto/ligato/linux/namespace"
)

func TestAgentInSync(t *testing.T) {
	ctx := setupE2E(t)
	defer ctx.teardownE2E()
	Expect(ctx.agentInSync()).To(BeTrue())
}

func TestStartStopMicroservice(t *testing.T) {
	ctx := setupE2E(t)
	defer ctx.teardownE2E()

	const msName = "microservice1"
	key := ns.MicroserviceKey(msNamePrefix + msName)
	msState := func() kvscheduler.ValueState {
		return ctx.getValueStateByKey(key)
	}

	ctx.startMicroservice(msName)
	Eventually(msState).Should(Equal(kvscheduler.ValueState_OBTAINED))
	ctx.stopMicroservice(msName)
	Eventually(msState).Should(Equal(kvscheduler.ValueState_NONEXISTENT))
}
