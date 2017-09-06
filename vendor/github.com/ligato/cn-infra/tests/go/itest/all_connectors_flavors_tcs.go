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

package itest

import (
	"testing"

	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/flavors/connectors"
	"github.com/onsi/gomega"

	"github.com/ligato/cn-infra/db/keyval/etcdv3"
	etcdmock "github.com/ligato/cn-infra/db/keyval/etcdv3/mocks"
	"github.com/ligato/cn-infra/flavors/local"
	"github.com/ligato/cn-infra/messaging/kafka"
	kafkamux "github.com/ligato/cn-infra/messaging/kafka/mux"
)

type suiteFlavorAllConnectors struct {
	T *testing.T
	AgentT
	Given
	When
	Then
}

// AllConnectorsFlavorMocks
type AllConnectorsFlavorMocks struct {
	KafkaMock *kafkamux.KafkaMock
}

// Setup registers gomega and starts the agent with the flavor argument
func (t *suiteFlavorAllConnectors) Setup(flavor core.Flavor, golangT *testing.T) {
	t.AgentT.Setup(flavor, golangT)
}

// TC01 asserts that injection works fine and agent starts & stops.
// Not the connectors are not really connected (not event to the mock).
func (t *suiteFlavorAllConnectors) TC01StartStopWithoutConfig() {
	t.Setup(&connectors.AllConnectorsFlavor{}, t.T)
	defer t.Teardown()

	gomega.Expect(t.agent).ShouldNot(gomega.BeNil(), "agent is not initialized")
}

// MockAllConnectororsFlavor initializes embeded ETCD & Kafka MOCK
//
// Example:
//
//     kafkamock, _, _ := kafkamux.Mock(t)
//     MockAllConnectorsFlavor(t, localFlavor)
func MockAllConnectorsFlavor(t *testing.T, flavorLocal *local.FlavorLocal) (*connectors.AllConnectorsFlavor, *AllConnectorsFlavorMocks) {
	kafkaMock := kafkamux.Mock(t)

	embededEtcd := etcdmock.Embedded{}
	embededEtcd.Start(t)
	defer embededEtcd.Stop()

	etcdClientLogger := flavorLocal.LoggerFor("emedEtcdClient")
	etcdBytesCon, err := etcdv3.NewEtcdConnectionUsingClient(embededEtcd.Client(), etcdClientLogger)
	if err != nil {
		panic(err)
	}

	return &connectors.AllConnectorsFlavor{
		ETCD:  *etcdv3.FromExistingConnection(etcdBytesCon, &flavorLocal.ServiceLabel),
		Kafka: *kafka.FromExistingMux(kafkaMock.Mux),
	}, &AllConnectorsFlavorMocks{kafkaMock}
}

/* TODO
// TC02 asserts that injection works fine and agent starts & stops
func (t *suiteFlavorAllConnectors) TC02StartStopMocks() {
	flavor, _ := MockAllConnectorsFlavor(t.T, localFlavor)
	t.Setup(flavor, t.T)
	defer t.Teardown()

	gomega.Expect(t.agent).ShouldNot(gomega.BeNil(), "agent is not initialized")
}
*/
