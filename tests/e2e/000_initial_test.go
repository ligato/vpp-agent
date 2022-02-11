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
	ctx := Setup(t)
	defer ctx.Teardown()

	ctx.Expect(ctx.AgentInSync()).To(BeTrue())
}

func TestStartStopMicroservice(t *testing.T) {
	ctx := Setup(t)
	defer ctx.Teardown()

	const ms1 = "microservice1"
	key := ns.MicroserviceKey(MsNamePrefix + ms1)
	msState := func() kvscheduler.ValueState {
		return ctx.GetValueStateByKey(key)
	}

	ctx.StartMicroservice(ms1)
	ctx.Eventually(msState).Should(Equal(kvscheduler.ValueState_OBTAINED))

	ctx.StopMicroservice(ms1)
	ctx.Eventually(msState).Should(Equal(kvscheduler.ValueState_NONEXISTENT))
}

func TestStartStopAgent(t *testing.T) {
	ctx := Setup(t)
	defer ctx.Teardown()

	const agent1 = "agent1"
	key := ns.MicroserviceKey(agent1)
	msState := func() kvscheduler.ValueState {
		return ctx.GetValueStateByKey(key)
	}

	ctx.StartAgent(agent1)
	ctx.Eventually(msState).Should(Equal(kvscheduler.ValueState_OBTAINED))

	ctx.StopAgent(agent1)
	ctx.Eventually(msState).Should(Equal(kvscheduler.ValueState_NONEXISTENT))
}
