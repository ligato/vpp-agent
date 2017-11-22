package bdidx

import (
	"github.com/stretchr/testify/assert"
	"github.com/ligato/vpp-agent/plugins/defaultplugins/l2plugin/model/l2"
	"testing"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
	"github.com/ligato/vpp-agent/idxvpp"
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
)

func testInitialization(t *testing.T, bdToIfaces map[string][]string) (idxvpp.NameToIdxRW, BDIndexRW, []*l2.BridgeDomains_BridgeDomain) {
	//initialize index
	nameToIdx := nametoidx.NewNameToIdx(nil, "testName", "bd_indexes_test", IndexMetadata)
	bdIndex := NewBDIndex(nameToIdx)
	names := nameToIdx.ListNames()
	assert.Empty(t, names)
	//data preparation

	var bridgeDomains []*l2.BridgeDomains_BridgeDomain
	for bdName, ifaces := range bdToIfaces {
		bridgeDomains = append(bridgeDomains,prepareBridgeDomainData(bdName, ifaces))
	}

	return nameToIdx, bdIndex, bridgeDomains
}

/**
	TestIndexMetadatat tests whether func IndexMetadata return map filled with correct values
 */
func TestIndexMetadatat(t *testing.T) {
	//data preparation
	bridgeDomain := prepareBridgeDomainData(bdName0, []string{ifaceAName, ifaceBName})

	//call tested func
	result := IndexMetadata(bridgeDomain)

	//evaluate result
	assert.Len(t, result, 1)

	ifaceNames := result[ifaceNameIndexKey]
	assert.Len(t, ifaceNames, 2)

	assert.Contains(t, ifaceNames, ifaceAName)
	assert.Contains(t, ifaceNames, ifaceBName)
}

/**
	TestRegisterAndUnregisterName tests methods:
	* RegisterName()
	* UnregisterName()
 */
func TestRegisterAndUnregisterName(t *testing.T) {
	nameToIdx, bdIndex, bridgeDomains := testInitialization(t,map[string][]string{bdName0:{ifaceAName, ifaceBName}})

	//call tested func
	bdIndex.RegisterName(bridgeDomains[0].Name, idx0, bridgeDomains[0])

	var names []string
	//evaluate result
	names = nameToIdx.ListNames()
	assert.Len(t, names, 1)
	assert.Contains(t, names, bridgeDomains[0].Name)

	//call tested func
	bdIndex.UnregisterName(bridgeDomains[0].Name)

	//evaluate result
	names = nameToIdx.ListNames()
	assert.Empty(t, names)
}

/**
	TestLookupIndex tests method:
	* LookupIndex
 */
func TestLookupIndex(t *testing.T) {
	_, bdIndex, bridgeDomains := testInitialization(t,map[string][]string{bdName0:{ifaceAName, ifaceBName}})

	bdIndex.RegisterName(bridgeDomains[0].Name, idx0, bridgeDomains[0])

	foundIdx, metadata, exist := bdIndex.LookupIdx(bdName0)
	assert.True(t, exist)
	assert.Equal(t, idx0, foundIdx)
	assert.Equal(t, bridgeDomains[0], metadata)
}

/**
	TestLookupIndex tests method:
	* LookupIndex
 */
func TestLookupName(t *testing.T) {
	_, bdIndex, bridgeDomains := testInitialization(t, map[string][]string{bdName0:{ifaceAName, ifaceBName}})

	bdIndex.RegisterName(bridgeDomains[0].Name, idx0, bridgeDomains[0])

	foundName, metadata, exist := bdIndex.LookupName(idx0)
	assert.True(t, exist)
	assert.Equal(t, bridgeDomains[0].Name, foundName)
	assert.Equal(t, bridgeDomains[0], metadata)
}

/**
	TestLookupNameByIfaceName tests method:
	* LookupNameByIfaceName
 */
func TestLookupByIfaceName(t *testing.T) {
	//defines 3 bridge domains
	_, bdIndex, bridgeDomains := testInitialization(t,
		map[string][]string{
			bdName0:{ifaceAName, ifaceBName},
			bdName1:{ifaceAName},
			bdName2:{ifaceBName}})

	bdIndex.RegisterName(bridgeDomains[0].Name, idx0, bridgeDomains[0])
	bdIndex.RegisterName(bridgeDomains[1].Name, idx1, bridgeDomains[1])
	bdIndex.RegisterName(bridgeDomains[2].Name, idx2, bridgeDomains[2])

	//return all bridge domains to which ifaceAName belongs
	foundBridgeDomains := bdIndex.LookupNameByIfaceName(ifaceAName)
	assert.Len(t, foundBridgeDomains, 2)
	assert.Contains(t, foundBridgeDomains, bdName0)
	assert.Contains(t, foundBridgeDomains, bdName1)

	//return all bridge domains to which ifaceBName belongs
	foundBridgeDomains = bdIndex.LookupNameByIfaceName(ifaceBName)
	assert.Len(t, foundBridgeDomains, 2)
	assert.Contains(t, foundBridgeDomains, bdName0)
	assert.Contains(t, foundBridgeDomains, bdName2)
}

func prepareBridgeDomainData(bdName string, ifaces []string) *l2.BridgeDomains_BridgeDomain{
	interfaces := []*l2.BridgeDomains_BridgeDomain_Interfaces{}
	for _,iface := range ifaces {
		interfaces = append(interfaces, &l2.BridgeDomains_BridgeDomain_Interfaces{Name: iface})
	}
	return &l2.BridgeDomains_BridgeDomain{Interfaces:interfaces, Name:bdName}
}
