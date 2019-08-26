package mock

import (
	"errors"
	"net"

	"github.com/ligato/vpp-agent/api/models/netalloc"
	"github.com/ligato/vpp-agent/pkg/models"
	kvs "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/netalloc/utils"
)

// NetAlloc is a mock version of the netplugin, suitable for unit testing.
type NetAlloc struct {
	allocated map[string]*netalloc.IPAllocMetadata // allocation name -> parsed address
}

// Allocate simulates allocation of an IP address.
func (p *NetAlloc) Allocate(network, ifaceName, address, gw string) {
	addrAlloc := &netalloc.IPAllocation{
		NetworkName:   network,
		InterfaceName: ifaceName,
	}
	allocName := models.Name(addrAlloc)

	ifaceAddr, err := utils.ParseIPAddr(address, nil)
	if err != nil {
		panic(err)
	}
	gwAddr, err := utils.ParseIPAddr(gw, ifaceAddr)
	if err != nil {
		panic(err)
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

// ParseAddressAllocRef parses reference to an allocated IP address.
func (p *NetAlloc) ParseAddressAllocRef(addrAllocRef, expIface string) (
	network, iface string, isRef bool, err error) {
	return utils.ParseAddrAllocRef(addrAllocRef, expIface)
}

// GetAddressAllocDep is not implemented here.
func (p *NetAlloc) GetAddressAllocDep(addrOrAllocRef, ifaceName, depLabelPrefix string) (
	dep kvs.Dependency, hasAllocDep bool) {
	return kvs.Dependency{}, false
}

// ValidateIPAddress checks validity of address reference or, if <addrOrAllocRef>
// already contains an actual IP address, it tries to parse it.
func (p *NetAlloc) ValidateIPAddress(addrOrAllocRef, ifaceName, fieldName string) error {
	_, _, isRef, err := utils.ParseAddrAllocRef(addrOrAllocRef, ifaceName)
	if !isRef {
		_, err = utils.ParseIPAddr(addrOrAllocRef, nil)
	}
	if err != nil {
		if fieldName != "" {
			return kvs.NewInvalidValueError(err, fieldName)
		} else {
			return kvs.NewInvalidValueError(err)
		}
	}
	return nil
}

// GetOrParseIPAddress tries to get allocated interface (or GW) IP address
// referenced by <addrOrAllocRef> in the requested form. But if the string
// contains/ an actual IP address instead of a reference, the address is parsed
// using methods from the net package and returned in the requested form.
// For ADDR_ONLY address form, the returned <addr> will have the mask unset
// and the IP address should be accessed as <addr>.IP
func (p *NetAlloc) GetOrParseIPAddress(addrOrAllocRef string, ifaceName string,
	getGW bool, addrForm netalloc.IPAddressForm) (addr *net.IPNet, err error) {

	network, iface, isRef, err := utils.ParseAddrAllocRef(addrOrAllocRef, ifaceName)
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
			return utils.GetIPAddrInGivenForm(allocation.GwAddr, addrForm), nil
		}
		return utils.GetIPAddrInGivenForm(allocation.IfaceAddr, addrForm), nil
	}

	// try to parse the address
	ipAddr, err := utils.ParseIPAddr(addrOrAllocRef, nil)
	if err != nil {
		return nil, err
	}
	return utils.GetIPAddrInGivenForm(ipAddr, addrForm), nil
}

// CorrelateRetrievedIPs is not implemented here.
func (p *NetAlloc) CorrelateRetrievedIPs(expAddrsOrRefs []string, retrievedAddrs []string,
	ifaceName string, areGWs bool, addrForm netalloc.IPAddressForm) (correlated []string) {
	return retrievedAddrs
}
