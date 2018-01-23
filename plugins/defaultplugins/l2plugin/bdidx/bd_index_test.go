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

package bdidx_test

import (
	"testing"

	"github.com/ligato/vpp-agent/idxvpp"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/common/model/l2"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/bdidx"
	. "github.com/onsi/gomega"
)

const (
	//bridge domain name
	bdName0    = "bd0"
	bdName1    = "bd1"
	bdName2    = "bd2"
	ifaceAName = "interfaceA"
	ifaceBName = "interfaceB"

	idx0 uint32 = 0
	idx1 uint32 = 1
	idx2 uint32 = 2

	ifaceNameIndexKey = "ipAddrKey"
)

func testInitialization(t *testing.T, bdToIfaces map[string][]string) (idxvpp.NameToIdxRW, bdidx.BDIndexRW, []*l2.BridgeDomains_BridgeDomain) {
	//initialize index
	RegisterTestingT(t)
	nameToIdx := nametoidx.NewNameToIdx(nil, "testName", "bd_indexes_test", bdidx.IndexMetadata)
	bdIndex := bdidx.NewBDIndex(nameToIdx)
	names := nameToIdx.ListNames()
	Expect(names).To(BeEmpty())

	//data preparation
	var bridgeDomains []*l2.BridgeDomains_BridgeDomain
	for bdName, ifaces := range bdToIfaces {
		bridgeDomains = append(bridgeDomains, prepareBridgeDomainData(bdName, ifaces))
	}

	return nameToIdx, bdIndex, bridgeDomains
}

/**
TestIndexMetadatat tests whether func IndexMetadata return map filled with correct values
*/
func TestIndexMetadatat(t *testing.T) {
	RegisterTestingT(t)
	//data preparation
	bridgeDomain := prepareBridgeDomainData(bdName0, []string{ifaceAName, ifaceBName})

	//call tested func
	result := bdidx.IndexMetadata(bridgeDomain)

	//evaluate result
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
	nameToIdx, bdIndex, bridgeDomains := testInitialization(t, map[string][]string{bdName0: {ifaceAName, ifaceBName}})

	//call tested func
	bdIndex.RegisterName(bridgeDomains[0].Name, idx0, bridgeDomains[0])

	var names []string
	//evaluate result
	names = nameToIdx.ListNames()
	Expect(names).To(HaveLen(1))
	Expect(names).To(ContainElement(bridgeDomains[0].Name))

	//call tested func
	bdIndex.UnregisterName(bridgeDomains[0].Name)

	//evaluate result
	names = nameToIdx.ListNames()
	Expect(names).To(BeEmpty())
}

/**
TestLookupIndex tests method:
* LookupIndex
*/
func TestLookupIndex(t *testing.T) {
	RegisterTestingT(t)
	_, bdIndex, bridgeDomains := testInitialization(t, map[string][]string{bdName0: {ifaceAName, ifaceBName}})

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
	_, bdIndex, bridgeDomains := testInitialization(t, map[string][]string{bdName0: {ifaceAName, ifaceBName}})

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
	//defines 3 bridge domains
	_, bdIndex, bridgeDomains := testInitialization(t,
		map[string][]string{
			bdName0: {ifaceAName, ifaceBName},
			bdName1: {ifaceAName},
			bdName2: {ifaceBName}})

	bdIndex.RegisterName(bridgeDomains[0].Name, idx0, bridgeDomains[0])
	bdIndex.RegisterName(bridgeDomains[1].Name, idx1, bridgeDomains[1])
	bdIndex.RegisterName(bridgeDomains[2].Name, idx2, bridgeDomains[2])

	//return all bridge domains to which ifaceAName belongs
	foundBridgeDomains := bdIndex.LookupNameByIfaceName(ifaceAName)
	Expect(foundBridgeDomains).To(HaveLen(2))
	Expect(foundBridgeDomains).To(ContainElement(bdName0))
	Expect(foundBridgeDomains).To(ContainElement(bdName1))

	//return all bridge domains to which ifaceBName belongs
	foundBridgeDomains = bdIndex.LookupNameByIfaceName(ifaceBName)
	Expect(foundBridgeDomains).To(HaveLen(2))
	Expect(foundBridgeDomains).To(ContainElement(bdName0))
	Expect(foundBridgeDomains).To(ContainElement(bdName2))
}

func prepareBridgeDomainData(bdName string, ifaces []string) *l2.BridgeDomains_BridgeDomain {
	interfaces := []*l2.BridgeDomains_BridgeDomain_Interfaces{}
	for _, iface := range ifaces {
		interfaces = append(interfaces, &l2.BridgeDomains_BridgeDomain_Interfaces{Name: iface})
	}
	return &l2.BridgeDomains_BridgeDomain{Interfaces: interfaces, Name: bdName}
}
