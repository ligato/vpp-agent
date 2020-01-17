package mock

import (
	"errors"
	"net"

	"go.ligato.io/vpp-agent/v3/pkg/models"
	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	plugin "go.ligato.io/vpp-agent/v3/plugins/netalloc"
	"go.ligato.io/vpp-agent/v3/plugins/netalloc/utils"
	"go.ligato.io/vpp-agent/v3/proto/ligato/netalloc"
)

// NetAlloc is a mock version of the netplugin, suitable for unit testing.
type NetAlloc struct {
	realNetAlloc *plugin.Plugin
	allocated    map[string]*netalloc.IPAllocMetadata // allocation name -> parsed address
}

// NewMockNetAlloc is a constructor for mock netalloc plugin.
func NewMockNetAlloc() *NetAlloc {
	return &NetAlloc{
		realNetAlloc: &plugin.Plugin{},
		allocated:    make(map[string]*netalloc.IPAllocMetadata),
	}
}

// Allocate simulates allocation of an IP address.
func (p *NetAlloc) Allocate(network, ifaceName, address, gw string) {
	var (
		gwAddr *net.IPNet
		err    error
	)
	addrAlloc := &netalloc.IPAllocation{
		NetworkName:   network,
		InterfaceName: ifaceName,
	}
	allocName := models.Name(addrAlloc)

	ifaceAddr, _, err := utils.ParseIPAddr(address, nil)
	if err != nil {
		panic(err)
	}
	if gw != "" {
		gwAddr, _, err = utils.ParseIPAddr(gw, ifaceAddr)
		if err != nil {
			panic(err)
		}
	}
	p.allocated[allocName] = &netalloc.IPAllocMetadata{IfaceAddr: ifaceAddr, GwAddr: gwAddr}
}

// Deallocate simulates de-allocation of an IP address.
func (p *NetAlloc) Deallocate(network, ifaceName string) {
	addrAlloc := &netalloc.IPAllocation{
		NetworkName:   network,
		InterfaceName: ifaceName,
	}
	allocName := models.Name(addrAlloc)
	delete(p.allocated, allocName)
}

// CreateAddressAllocRef creates reference to an allocated IP address.
func (p *NetAlloc) CreateAddressAllocRef(network, iface string, getGW bool) string {
	return p.realNetAlloc.CreateAddressAllocRef(network, iface, getGW)
}

// ParseAddressAllocRef parses reference to an allocated IP address.
func (p *NetAlloc) ParseAddressAllocRef(addrAllocRef, expIface string) (
	network, iface string, isGW, isRef bool, err error) {
	return p.realNetAlloc.ParseAddressAllocRef(addrAllocRef, expIface)
}

// GetAddressAllocDep is not implemented here.
func (p *NetAlloc) GetAddressAllocDep(addrOrAllocRef, ifaceName, depLabelPrefix string) (
	dep kvs.Dependency, hasAllocDep bool) {
	return kvs.Dependency{}, false
}

// ValidateIPAddress checks validity of address reference or, if <addrOrAllocRef>
// already contains an actual IP address, it tries to parse it.
func (p *NetAlloc) ValidateIPAddress(addrOrAllocRef, ifaceName, fieldName string, gwCheck plugin.GwValidityCheck) error {
	return p.realNetAlloc.ValidateIPAddress(addrOrAllocRef, ifaceName, fieldName, gwCheck)
}

// GetOrParseIPAddress tries to get allocated interface (or GW) IP address
// referenced by <addrOrAllocRef> in the requested form. But if the string
// contains/ an actual IP address instead of a reference, the address is parsed
// using methods from the net package and returned in the requested form.
// For ADDR_ONLY address form, the returned <addr> will have the mask unset
// and the IP address should be accessed as <addr>.IP
func (p *NetAlloc) GetOrParseIPAddress(addrOrAllocRef string, ifaceName string,
	addrForm netalloc.IPAddressForm) (addr *net.IPNet, err error) {

	network, iface, getGW, isRef, err := utils.ParseAddrAllocRef(addrOrAllocRef, ifaceName)
	if isRef && err != nil {
		return nil, err
	}

	if isRef {
		// de-reference
		allocName := models.Name(&netalloc.IPAllocation{
			NetworkName:   network,
			InterfaceName: iface,
		})
		allocation, found := p.allocated[allocName]
		if !found {
			return nil, errors.New("address is not allocated")
		}
		if getGW {
			if allocation.GwAddr == nil {
				return nil, errors.New("gw address is not defined")
			}
			return utils.GetIPAddrInGivenForm(allocation.GwAddr, addrForm), nil
		}
		return utils.GetIPAddrInGivenForm(allocation.IfaceAddr, addrForm), nil
	}

	// try to parse the address
	ipAddr, _, err := utils.ParseIPAddr(addrOrAllocRef, nil)
	if err != nil {
		return nil, err
	}
	return utils.GetIPAddrInGivenForm(ipAddr, addrForm), nil
}

// CorrelateRetrievedIPs is not implemented here.
func (p *NetAlloc) CorrelateRetrievedIPs(expAddrsOrRefs []string, retrievedAddrs []string,
	ifaceName string, addrForm netalloc.IPAddressForm) (correlated []string) {
	return retrievedAddrs
}
