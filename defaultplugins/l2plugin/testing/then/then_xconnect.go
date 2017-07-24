package then

import (
	//log "github.com/ligato/cn-infra/logging/logrus"
	//. "github.com/onsi/gomega"
	//"github.com/ligato/vpp-agent/defaultplugins/l2plugin"
	//"time"
	. "github.com/onsi/gomega"
)
import (
	"github.com/ligato/vpp-agent/defaultplugins"
	"github.com/ligato/vpp-agent/defaultplugins/l2plugin"
	"time"
)

// SwXConIndexes is a constructor
func SwXConIndexes() *XConIndexesAssertions {
	return &XConIndexesAssertions{}
}

// XConIndexesAssertions helper struct for fluent DSL in tests
type XConIndexesAssertions struct {
}

// ContainsName verifies xConnect pair presence in mapping
func (a *XConIndexesAssertions) ContainsName(rIface string) {
	mapping := defaultplugins.GetXConnectIndexes()
	var exists bool
	var index uint32
	for i := 0; i < 10; i++ {
		index, _, exists = mapping.LookupIdx(rIface)
		if !exists {
			time.Sleep(200 * time.Millisecond)
		}
	}
	Expect(index).ToNot(BeZero())
}

// DoesNotContainName that xConnect pair is not present in mapping
func (a *XConIndexesAssertions) DoesNotContainName(rIface string) {
	mapping := defaultplugins.GetXConnectIndexes()
	var exists bool
	var index uint32
	for i := 0; i < 10; i++ {
		index, _, exists = mapping.LookupIdx(rIface)
		if exists {
			time.Sleep(200 * time.Millisecond)
		}
	}
	Expect(index).To(BeZero())
}

// VerifyTransmitInterface that provided receive interface has correct transmit interface assigned in xConnect
func (a *XConIndexesAssertions) VerifyTransmitInterface(rIface string, expectedIface string) {
	mapping := defaultplugins.GetXConnectIndexes()
	var exists bool
	var index uint32
	var meta interface{}
	for i := 0; i < 10; i++ {
		index, meta, exists = mapping.LookupIdx(rIface)
		if !exists {
			time.Sleep(200 * time.Millisecond)
		}
	}
	Expect(index).ToNot(BeZero())
	Expect(meta).ToNot(BeNil())
	tIface := meta.(l2plugin.XConnectMeta).TransmitInterface
	Expect(expectedIface).To(BeEquivalentTo(tIface))
}
