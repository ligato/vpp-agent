package itest

import (
	"testing"

	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/flavors/allcon"
	"github.com/onsi/gomega"
)

type suiteFlavorAllConnectors struct {
	T *testing.T
	AgentT
	Given
	When
	Then
}

// Setup registers gomega and starts the agent with the flavor argument
func (t *suiteFlavorAllConnectors) Setup(flavor core.Flavor, golangT *testing.T) {
	t.AgentT.Setup(flavor, golangT)
}

// TC01 asserts that injection works fine and agent starts & stops
func (t *suiteFlavorAllConnectors) TC01StartStop() {
	t.Setup(&allcon.AllConnectorsFlavor{}, t.T)
	defer t.Teardown()

	gomega.Expect(t.agent).ShouldNot(gomega.BeNil(), "agent is not initialized")
}
