package iftst

import (
	intf "github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin/model/interfaces"
	"github.com/ligato/vpp-agent/tests/go/itest/idxtst"
	//"github.com/ligato/vpp-agent/plugins/defaultplugins/ifplugin"
	"time"

	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/logging/logroot"
	vppclient "github.com/ligato/vpp-agent/clientv1/defaultplugins"
	idx "github.com/ligato/vpp-agent/idxvpp"
	"github.com/ligato/vpp-agent/plugins/defaultplugins"
	. "github.com/onsi/gomega"
)

// ThenIface is a collection of test step methods (see Behavior Driven Development)
// (methods that will be called from test scenarios).
type ThenIface struct {
	NewChange func(name core.PluginName) vppclient.DataChangeDSL
	OperState ifstateGetter

	Log logging.Logger
	VPP defaultplugins.API
}

type ifstateGetter interface {
	// InterfaceState reads operational state of network interface
	// and fills it to ifState input parameter.
	InterfaceState(ifaceName string, ifState *intf.InterfacesState_Interface) (found bool, err error)
}

// SwIfIndexes is a constructor for interfaces.
func (step *ThenIface) SwIfIndexes() *SwIfIndexesAssertions {
	return &SwIfIndexesAssertions{VPP: step.VPP}
}

// BfdIndexes is a constructor for interfaces.
func (step *ThenIface) BfdIndexes() *BfdIndexesAssertions {
	return &BfdIndexesAssertions{}
}

// SwIfIndexesAssertions is a helper struct for fluent DSL in tests for interfaces.
type SwIfIndexesAssertions struct {
	VPP defaultplugins.API
}

// BfdIndexesAssertions is a helper struct for fluent DSL in tests for bfd.
type BfdIndexesAssertions struct {
}

// ContainsName checks several times if sw_if_index - ifName mapping exists.
func (a *SwIfIndexesAssertions) ContainsName(ifName string) {
	idxtst.ContainsName(a.VPP.GetSwIfIndexes().GetMapping(), ifName)
}

// ContainsName checks several times if there is an entry with the given name in bfd_index.
func (a *BfdIndexesAssertions) ContainsName(mapping idx.NameToIdx, bfdIface string) {
	idxtst.ContainsName(mapping, bfdIface)
}

// NotContainsName checks several times the sw_if_index - ifName mapping does not exist.
func (a *SwIfIndexesAssertions) NotContainsName(ifName string) {
	idxtst.NotContainsNameAfter(a.VPP.GetSwIfIndexes().GetMapping(), ifName)
}

// NotContainsName checks several times if there is no entry with the given name in bfd_index.
func (a *BfdIndexesAssertions) NotContainsName(mapping idx.NameToIdx, bfdInterface string) {
	idxtst.NotContainsNameAfter(mapping, bfdInterface)
}

// IfStateInDB asserts that there is InterfacesState_Interface_DOWN in ETCD for particular Interfaces_Interface.
func (step *ThenIface) IfStateInDB(status intf.InterfacesState_Interface_Status, data *intf.Interfaces_Interface) {
	logroot.StandardLogger().Debug("IfStateDownInDB begin")

	ifState := &intf.InterfacesState_Interface{}
	var found bool
	var err error
	for i := 0; i < 3; i++ {
		step.OperState.InterfaceState(data.Name, ifState)

		if err != nil {
			logroot.StandardLogger().Panic(err)
		}
		if found {
			break
		}
		time.Sleep(1 * time.Second)
	}
	Expect(found).Should(BeTrue(),
		"not found operational state "+data.Name)
	// Expect(ifState.OperStatus).Should(BeEquivalentTo(status),
	//	"Status needs to be down for " + data.Name) // TODO

	logroot.StandardLogger().Debug("IfStateDownInDB end")
}
