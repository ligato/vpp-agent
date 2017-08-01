package then

import (
	"github.com/ligato/vpp-agent/plugins/defaultplugins"
	"github.com/ligato/vpp-agent/idxvpp/idxtst"
	. "github.com/onsi/gomega"
)

// FIBIndexes is a constructor
func FIBIndexes() *FIBIndexesAssertions {
	return &FIBIndexesAssertions{}
}

// FIBIndexesAssertions helper struct for fluent DSL in tests
type FIBIndexesAssertions struct {
}

// ContainsName verifies that particular FIB entry exists in mapping
func (then *FIBIndexesAssertions) ContainsName(fibMac string) {
	idxtst.ContainsName(defaultplugins.GetFIBIndexes(), fibMac)
}

// NotContainsNameAfter verifies that FIB mac is not present in mapping
func (then *FIBIndexesAssertions) NotContainsNameAfter(mac string) {
	_, _, exists := defaultplugins.GetFIBIndexes().LookupIdx(mac)
	Expect(exists).Should(BeFalse())
}

// IsCached verifies that particular FIB entry exists in FIB cache mapping
func (then *FIBIndexesAssertions) IsCached(fibMac string) {
	idxtst.ContainsName(defaultplugins.GetFIBDesIndexes(), fibMac)
}

// IsNotCached verifies that particular FIB entry does not exists in FIB cache mapping
func (then *FIBIndexesAssertions) IsNotCached(fibMac string) {
	_, _, exists := defaultplugins.GetFIBDesIndexes().LookupIdx(fibMac)
	Expect(exists).Should(BeFalse())
}
