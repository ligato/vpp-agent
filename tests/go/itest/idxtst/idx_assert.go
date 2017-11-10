package idxtst

import (
	"time"

	idx "github.com/ligato/vpp-agent/idxvpp"
	. "github.com/onsi/gomega"
)

func lookupIdx(mapping idx.NameToIdx, lookupName string) func() uint32 {
	return func() uint32 {
		swIdx, _, _ := mapping.LookupIdx(lookupName)
		return swIdx
	}
}

// ContainsName verifies lookup name presence in provided mapping
func ContainsName(mapping idx.NameToIdx, lookupName string) uint32 {
	Eventually(lookupIdx(mapping, lookupName), 100*time.Millisecond, 10*time.Millisecond).ShouldNot(BeZero())
	// block until Eventually's timeout elapses
	time.Sleep(1 * time.Second)
	return lookupIdx(mapping, lookupName)()
}

// ContainsMeta verifies lookup meta presence in provided mapping
func ContainsMeta(mapping idx.NameToIdx, lookupName string) interface{} {
	var exists bool
	var meta interface{}
	for i := 0; i < 5; i++ {
		_, meta, exists = mapping.LookupIdx(lookupName)
		if !exists {
			time.Sleep(100 * time.Millisecond)
		} else {
			break
		}
	}
	if exists {
		return meta
	}
	return nil
}

// NotContainsNameAfter verifies that name is not present in mapping after specified time
func NotContainsNameAfter(mapping idx.NameToIdx, lookupName string) {
	time.Sleep(100 * time.Millisecond)
	swIdx := lookupIdx(mapping, lookupName)()
	Expect(swIdx).To(BeZero())
}
