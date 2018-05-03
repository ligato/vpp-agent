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

package l2idx_test

import (
	"testing"

	"github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/vpp-agent/idxvpp"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/l2"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/bdidx"
	. "github.com/onsi/gomega"
)

const (
	// bridge domain name
	bdName0    = "bd0"
	bdName1    = "bd1"
	bdName2    = "bd2"
	ifaceAName = "interfaceA"
	ifaceBName = "interfaceB"
	ifaceCName = "interfaceC"
	ifaceDName = "interfaceD"

	idx0 uint32 = 0
	idx1 uint32 = 1
	idx2 uint32 = 2

	ifaceNameIndexKey = "ipAddrKey"
)

func testInitialization(t *testing.T, bdToIfaces map[string][]string) (idxvpp.NameToIdxRW, l2idx.BDIndexRW, []*l2.BridgeDomains_BridgeDomain) {
	RegisterTestingT(t)

	// initialize index
	nameToIdx := nametoidx.NewNameToIdx(logrus.DefaultLogger(), "testName", "bd_indexes_test", l2idx.IndexMetadata)
	bdIndex := l2idx.NewBDIndex(nameToIdx)
	names := nameToIdx.ListNames()
	Expect(names).To(BeEmpty())

	// data preparation
	var bridgeDomains []*l2.BridgeDomains_BridgeDomain
	for bdName, ifaces := range bdToIfaces {
		bridgeDomains = append(bridgeDomains, prepareBridgeDomainData(bdName, ifaces))
	}

	return bdIndex.GetMapping(), bdIndex, bridgeDomains
}

func prepareBridgeDomainData(bdName string, ifaces []string) *l2.BridgeDomains_BridgeDomain {
	var interfaces []*l2.BridgeDomains_BridgeDomain_Interfaces
	for _, iface := range ifaces {
		interfaces = append(interfaces, &l2.BridgeDomains_BridgeDomain_Interfaces{Name: iface})
	}
	return &l2.BridgeDomains_BridgeDomain{Interfaces: interfaces, Name: bdName}
}

/**
TestIndexMetadatat tests whether func IndexMetadata return map filled with correct values
*/
func TestIndexMetadatat(t *testing.T) {
	RegisterTestingT(t)

	bridgeDomain := prepareBridgeDomainData(bdName0, []string{ifaceAName, ifaceBName})

	result := l2idx.IndexMetadata(nil)
	Expect(result).To(HaveLen(0))

	result = l2idx.IndexMetadata(bridgeDomain)
	Expect(result).To(HaveLen(1))

	ifaceNames := result[ifaceNameIndexKey]
	Expect(ifaceNames).To(HaveLen(2))
	Expect(ifaceNames).To(ContainElement(ifaceAName))
	Expect(ifaceNames).To(ContainElement(ifaceBName))
}

/**
TestRegisterAndUnregisterName tests methods:
* RegisterName()
* UnregisterName()
*/
func TestRegisterAndUnregisterName(t *testing.T) {
	RegisterTestingT(t)

	nameToIdx, bdIndex, bridgeDomains := testInitialization(t, map[string][]string{
		bdName0: {ifaceAName, ifaceBName},
	})

	bdIndex.RegisterName(bridgeDomains[0].Name, idx0, bridgeDomains[0])
	names := nameToIdx.ListNames()
	Expect(names).To(HaveLen(1))
	Expect(names).To(ContainElement(bridgeDomains[0].Name))

	bdIndex.UnregisterName(bridgeDomains[0].Name)
	names = nameToIdx.ListNames()
	Expect(names).To(BeEmpty())
}

/**
TestUpdateMetadata tests methods:
* UpdateMetadata()
*/
func TestUpdateMetadata(t *testing.T) {
	RegisterTestingT(t)

	nameToIdx, bdIndex, _ := testInitialization(t, nil)
	bd := prepareBridgeDomainData(bdName0, []string{ifaceAName, ifaceBName})
	bdUpdt1 := prepareBridgeDomainData(bdName0, []string{ifaceCName})
	bdUpdt2 := prepareBridgeDomainData(bdName0, []string{ifaceDName})

	// Update before registration (no entry created)
	success := bdIndex.UpdateMetadata(bd.Name, bd)
	Expect(success).To(BeFalse())
	_, metadata, found := nameToIdx.LookupIdx(bd.Name)
	Expect(found).To(BeFalse())
	Expect(metadata).To(BeNil())

	// Register bridge domain
	bdIndex.RegisterName(bd.Name, idx0, bd)
	var names []string
	names = nameToIdx.ListNames()
	Expect(names).To(HaveLen(1))
	Expect(names).To(ContainElement(bd.Name))

	// Evaluate entry metadata
	_, metadata, found = nameToIdx.LookupIdx(bd.Name)
	Expect(found).To(BeTrue())
	Expect(metadata).ToNot(BeNil())

	bdData, ok := metadata.(*l2.BridgeDomains_BridgeDomain)
	Expect(ok).To(BeTrue())
	Expect(bdData.Interfaces).To(HaveLen(2))

	var ifNames []string
	for _, ifData := range bdData.Interfaces {
		ifNames = append(ifNames, ifData.Name)
	}
	Expect(ifNames).To(ContainElement(ifaceAName))
	Expect(ifNames).To(ContainElement(ifaceBName))

	// Update metadata (same name, different data)
	success = bdIndex.UpdateMetadata(bdUpdt1.Name, bdUpdt1)
	Expect(success).To(BeTrue())

	// Evaluate updated metadata
	_, metadata, found = nameToIdx.LookupIdx(bd.Name)
	Expect(found).To(BeTrue())
	Expect(metadata).ToNot(BeNil())

	bdData, ok = metadata.(*l2.BridgeDomains_BridgeDomain)
	Expect(ok).To(BeTrue())
	Expect(bdData.Interfaces).To(HaveLen(1))

	ifNames = []string{}
	for _, ifData := range bdData.Interfaces {
		ifNames = append(ifNames, ifData.Name)
	}
	Expect(ifNames).To(ContainElement(ifaceCName))

	// Update metadata again
	success = bdIndex.UpdateMetadata(bdUpdt2.Name, bdUpdt2)
	Expect(success).To(BeTrue())

	// Evaluate updated metadata
	_, metadata, found = nameToIdx.LookupIdx(bd.Name)
	Expect(found).To(BeTrue())
	Expect(metadata).ToNot(BeNil())

	bdData, ok = metadata.(*l2.BridgeDomains_BridgeDomain)
	Expect(ok).To(BeTrue())
	Expect(bdData.Interfaces).To(HaveLen(1))

	ifNames = []string{}
	for _, ifData := range bdData.Interfaces {
		ifNames = append(ifNames, ifData.Name)
	}
	Expect(ifNames).To(ContainElement(ifaceDName))

	// Unregister
	bdIndex.UnregisterName(bd.Name)

	// Evaluate unregistration
	names = nameToIdx.ListNames()
	Expect(names).To(BeEmpty())
}

/**
TestLookupIndex tests method:
* LookupIndex
*/
func TestLookupIndex(t *testing.T) {
	RegisterTestingT(t)

	_, bdIndex, bridgeDomains := testInitialization(t, map[string][]string{
		bdName0: {ifaceAName, ifaceBName},
	})

	bdIndex.RegisterName(bridgeDomains[0].Name, idx0, bridgeDomains[0])

	foundIdx, metadata, exist := bdIndex.LookupIdx(bdName0)
	Expect(exist).To(BeTrue())
	Expect(foundIdx).To(Equal(idx0))
	Expect(metadata).To(Equal(bridgeDomains[0]))
}

/**
TestLookupIndex tests method:
* LookupIndex
*/
func TestLookupName(t *testing.T) {
	RegisterTestingT(t)

	_, bdIndex, bridgeDomains := testInitialization(t, map[string][]string{
		bdName0: {ifaceAName, ifaceBName},
	})

	bdIndex.RegisterName(bridgeDomains[0].Name, idx0, bridgeDomains[0])

	foundName, metadata, exist := bdIndex.LookupName(idx0)
	Expect(exist).To(BeTrue())
	Expect(foundName).To(Equal(bridgeDomains[0].Name))
	Expect(metadata).To(Equal(bridgeDomains[0]))
}

/**
TestLookupNameByIfaceName tests method:
* LookupNameByIfaceName
*/
func TestLookupByIfaceName(t *testing.T) {
	RegisterTestingT(t)

	// defines 3 bridge domains
	_, bdIndex, bridgeDomains := testInitialization(t, map[string][]string{
		bdName0: {ifaceAName, ifaceBName},
		bdName1: {ifaceCName},
		bdName2: {ifaceDName},
	})

	// Assign correct index to every bridge domain
	for _, bridgeDomain := range bridgeDomains {
		if bridgeDomain.Name == bdName0 {
			bdIndex.RegisterName(bridgeDomain.Name, idx0, bridgeDomain)
		} else if bridgeDomain.Name == bdName1 {
			bdIndex.RegisterName(bridgeDomain.Name, idx1, bridgeDomain)
		} else {
			bdIndex.RegisterName(bridgeDomain.Name, idx2, bridgeDomain)
		}
	}

	// return all bridge domains to which ifaceAName belongs
	bdIdx, _, _, exists := bdIndex.LookupBdForInterface(ifaceAName)
	Expect(exists).To(BeTrue())
	Expect(bdIdx).To(BeEquivalentTo(0))

	bdIdx, _, _, exists = bdIndex.LookupBdForInterface(ifaceBName)
	Expect(exists).To(BeTrue())
	Expect(bdIdx).To(BeEquivalentTo(0))

	bdIdx, _, _, exists = bdIndex.LookupBdForInterface(ifaceCName)
	Expect(exists).To(BeTrue())
	Expect(bdIdx).To(BeEquivalentTo(1))

	bdIdx, _, _, exists = bdIndex.LookupBdForInterface(ifaceDName)
	Expect(exists).To(BeTrue())
	Expect(bdIdx).To(BeEquivalentTo(2))

	_, _, _, exists = bdIndex.LookupBdForInterface("")
	Expect(exists).To(BeFalse())
}

func TestWatchNameToIdx(t *testing.T) {
	RegisterTestingT(t)

	_, bdIndex, bridgeDomains := testInitialization(t, map[string][]string{
		bdName0: {ifaceAName, ifaceBName},
	})

	c := make(chan l2idx.BdChangeDto)
	bdIndex.WatchNameToIdx("testName", c)

	bdIndex.RegisterName(bridgeDomains[0].Name, idx0, bridgeDomains[0])

	var dto l2idx.BdChangeDto
	Eventually(c).Should(Receive(&dto))

	Expect(dto.Name).To(Equal(bridgeDomains[0].Name))
	Expect(dto.Metadata).To(Equal(bridgeDomains[0]))
}
