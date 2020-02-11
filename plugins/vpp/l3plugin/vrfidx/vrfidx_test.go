// Copyright (c) 2019 Cisco and/or its affiliates.
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

package vrfidx

import (
	"testing"

	"fmt"

	. "github.com/onsi/gomega"
	"go.ligato.io/cn-infra/v2/logging/logrus"
)

// Constants
const (
	vrfName0         = "vrf-0"
	vrfName1         = "vrf-1"
	vrfName2         = "vrf-2"
	idx0      uint32 = 0
	idx1      uint32 = 1
	idx2      uint32 = 2
	watchName        = "watchName"
)

func testInitialization(t *testing.T) VRFMetadataIndexRW {
	RegisterTestingT(t)
	return NewVRFIndex(logrus.DefaultLogger(), "vrf-meta-index")
}

// Tests registering and unregistering name to index
func TestRegisterAndUnregisterName(t *testing.T) {
	index := testInitialization(t)
	vrf := &VRFMetadata{Index: idx0}

	// Register vrf
	index.Put(vrfName0, vrf)
	names := index.ListAllVRFs()
	Expect(names).To(HaveLen(1))
	Expect(names).To(ContainElement(vrfName0))
	metadata := index.ListAllVrfMetadata()
	Expect(metadata).To(HaveLen(1))
	Expect(metadata[0].GetIndex()).To(Equal(idx0))

	// Unregister vrf
	index.Delete(vrfName0)
	names = index.ListAllVRFs()
	Expect(names).To(BeEmpty())
}

// Tests index mapping clear
func TestClearInterfaces(t *testing.T) {
	index := testInitialization(t)

	// Register entries
	index.Put(vrfName0, &VRFMetadata{Index: idx0})
	index.Put(vrfName1, &VRFMetadata{Index: idx1})
	index.Put(vrfName2, &VRFMetadata{Index: idx2})
	names := index.ListAllVRFs()
	Expect(names).To(HaveLen(3))

	// Clear
	index.Clear()
	names = index.ListAllVRFs()
	Expect(names).To(BeEmpty())
}

// Tests lookup by name
func TestLookupByName(t *testing.T) {
	index := testInitialization(t)
	vrf := &VRFMetadata{Index: idx0}

	index.Put(vrfName0, vrf)

	metadata, exist := index.LookupByName(vrfName0)
	Expect(exist).To(BeTrue())
	Expect(metadata.GetIndex()).To(Equal(idx0))
	Expect(metadata).To(Equal(vrf))
}

// Tests lookup by index
func TestLookupByIndex(t *testing.T) {
	index := testInitialization(t)
	vrf := &VRFMetadata{Index: idx0}

	index.Put(vrfName0, vrf)

	foundName, metadata, exist := index.LookupByVRFIndex(idx0)
	Expect(exist).To(BeTrue())
	Expect(foundName).To(Equal(vrfName0))
	Expect(metadata).To(Equal(vrf))
}

// Tests watch VRFs
func TestWatchNameToIdx(t *testing.T) {
	fmt.Println("TestWatchNameToIdx")
	index := testInitialization(t)
	vrf := &VRFMetadata{Index: idx0}

	c := make(chan VRFMetadataDto, 10)
	index.WatchVRFs(watchName, c)

	index.Put(vrfName0, vrf)

	var dto VRFMetadataDto
	Eventually(c).Should(Receive(&dto))

	Expect(dto.Name).To(Equal(vrfName0))
	Expect(dto.Metadata.GetIndex()).To(Equal(idx0))
}
