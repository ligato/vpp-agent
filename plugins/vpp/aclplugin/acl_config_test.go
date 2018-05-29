package aclplugin

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestDiffInterfaces(t *testing.T) {
	RegisterTestingT(t)

	for _, test := range []struct {
		oldIfaces, newIfaces       []uint32
		expectAdded, expectRemoved []uint32
	}{
		{
			oldIfaces:     []uint32{},
			newIfaces:     []uint32{},
			expectAdded:   nil,
			expectRemoved: nil,
		},
		{
			oldIfaces:     []uint32{1},
			newIfaces:     []uint32{1},
			expectAdded:   nil,
			expectRemoved: nil,
		},
		{
			oldIfaces:     []uint32{1},
			newIfaces:     []uint32{2},
			expectAdded:   []uint32{2},
			expectRemoved: []uint32{1},
		},
		{
			oldIfaces:     []uint32{1},
			newIfaces:     []uint32{},
			expectAdded:   nil,
			expectRemoved: []uint32{1},
		},
		{
			oldIfaces:     []uint32{},
			newIfaces:     []uint32{2},
			expectAdded:   []uint32{2},
			expectRemoved: nil,
		},
		{
			oldIfaces:     []uint32{1, 2, 3},
			newIfaces:     []uint32{2, 4},
			expectAdded:   []uint32{4},
			expectRemoved: []uint32{1, 3},
		},
	} {
		added, removed := diffInterfaces(test.oldIfaces, test.newIfaces)

		Expect(added).To(ConsistOf(test.expectAdded))
		Expect(removed).To(ConsistOf(test.expectRemoved))
	}
}
