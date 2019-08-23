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
	allocated map[string]string // allocation name -> address
}

// Allocate simulates allocation of an address.
func (p *NetAlloc) Allocate(network, ifaceName, address string, addrType netalloc.AddressType) {
	addrAlloc := &netalloc.AddressAllocation{
		NetworkName:   network,
		InterfaceName: ifaceName,
		AddressType:   addrType,
	}
	allocName := models.Name(addrAlloc)
	p.allocated[allocName] = address
}

// Deallocate simulates de-allocation of an address.
func (p *NetAlloc) Deallocate(network, ifaceName string, addrType netalloc.AddressType) {
	addrAlloc := &netalloc.AddressAllocation{
		NetworkName:   network,
		InterfaceName: ifaceName,
		AddressType:   addrType,
	}
	allocName := models.Name(addrAlloc)
	delete(p.allocated, allocName)
}

// GetAddressAllocDep is not implemented here.
func (p *NetAlloc) GetAddressAllocDep(addrOrAllocRef, ifaceName, depLabelPrefix string) (
	dep kvs.Dependency, hasAllocDep bool) {
	return kvs.Dependency{}, false
}

// ValidateIPAddress checks validity of address reference or, if <addrOrAllocRef>
// already contains an actual IP address, it tries to parse it.
func (p *NetAlloc) ValidateIPAddress(addrOrAllocRef, ifaceName string) error {
	_, _, _, isRef, err := utils.ParseAddrAllocRef(addrOrAllocRef, ifaceName)
	if isRef {
		return err
	}
	_, err = utils.ParseIPAddr(addrOrAllocRef)
	return err
}

// GetOrParseIPAddress tries to get allocated IP address referenced by
// <addrOrAllocRef> in the requested form. But if the string contains
// an actual IP address instead of a reference, the address is parsed
// using methods from the net package and returned in the requested form.
// For ADDR_ONLY address form, the returned <addr> will have the mask unset
// and the IP address should be accessed as <addr>.IP
func (p *NetAlloc) GetOrParseIPAddress(addrOrAllocRef string, ifaceName string, addrForm netalloc.IPAddressForm) (
	addr *net.IPNet, err error) {

	network, iface, addrType, isRef, err := utils.ParseAddrAllocRef(addrOrAllocRef, ifaceName)
	if isRef && err != nil {
		return nil, err
	}

	if isRef {
		// de-reference
		allocName := models.Name(&netalloc.AddressAllocation{
			NetworkName:   network,
			InterfaceName: iface,
			AddressType:   addrType,
		})
		address, found := p.allocated[allocName]
		if !found {
			return nil, errors.New("address is not allocated")
		}
		addrOrAllocRef = address
	}

	// try to parse the address
	ipAddr, err := utils.ParseIPAddr(addrOrAllocRef)
	if err != nil {
		return nil, err
	}
	return utils.GetIPAddrInGivenForm(ipAddr, addrForm), nil
}

// CorrelateRetrievedIPs is not implemented here.
func (p *NetAlloc) CorrelateRetrievedIPs(expAddrsOrRefs []string, retrievedAddrs []string, ifaceName string,
	addrForm netalloc.IPAddressForm) []string {
	return retrievedAddrs
}
