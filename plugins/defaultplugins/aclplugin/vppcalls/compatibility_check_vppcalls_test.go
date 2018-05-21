package vppcalls

import (
	"testing"
	"github.com/ligato/vpp-agent/tests/vppcallmock"
	"github.com/ligato/cn-infra/logging/logrus"
	. "github.com/onsi/gomega"
)

func TestCheckMsgCompatibilityForACL(t *testing.T) {
	ctx := vppcallmock.SetupTestCtx(t)
	defer ctx.TeardownTestCtx()

	err := CheckMsgCompatibilityForACL(logrus.DefaultLogger(), ctx.MockChannel)
	Expect(err).To(BeNil())
}

