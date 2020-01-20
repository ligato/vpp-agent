package netalloc

import (
	"net"

	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/proto/ligato/netalloc"
)

// GwValidityCheck is used in ValidateIPAddress to tell if a GW reference is (un)expected/required.
type GwValidityCheck int

const (
	// GWRefAllowed is used when it doesn't matter if reference points to interface
	// address or GW address.
	GWRefAllowed GwValidityCheck = iota
	// GWRefRequired is used when an IP address reference should point to GW address.
	GWRefRequired
	// GwRefUnexpected is used when an IP address reference should not point to GW address.
	GwRefUnexpected
)

// AddressAllocator provides methods for descriptors of other plugins to reference
// and obtain allocated addresses.
//
// For example, if a model of some configuration item contains field IpAddress
// (of type string) which could reference allocated IP address (assigned to non
// pre-determined interface) and should be applied without mask, the descriptor
// for that item would implement some of the methods as follows:
//
//     func (d *Descriptor) Validate(key string, intf *mymodel.MyModel) error {
//         err := d.netallocPlugin.ValidateIPAddress(item.IpAddress, "", "IpAddress")
//         if err != nil {
//             return err
//         }
//     }
//
//     func (d *Descriptor) Dependencies(key string, item *mymodel.MyModel) (dependencies []kvs.Dependency) {
//         // Note: it is actually preferred to derive the IP address into a separate key-value
//         //       pair and assign the allocation dependency to it rather than to the entire configuration item
//         dep, hasAllocDep := d.netallocPlugin.GetAddressAllocDep(item.IpAddress, "", "")
//         if hasAllocDep {
//             dependencies = append(dependencies, dep)
//         }
//         // ...
//     }
//
//     func (d *Descriptor) Create(key string, item *mymodel.MyModel) (metadata interface{}, err error) {
//         addr, err := d.netallocPlugin.GetOrParseIPAddress(item.IpAddress, "", netalloc.ADDR_ONLY)
//         if err != nil {
//             d.log.Error(err)
//             return nil, err
//         }
//         fmt.Printf("Assign IP address: %v", addr.IP)
//         ...
//     }
//
//     func (d *Descriptor) Delete(key string, item *mymodel.MyModel) (err error) {
//         addr, err := d.netallocPlugin.GetOrParseIPAddress(item.IpAddress, "", netalloc.ADDR_ONLY)
//         if err != nil {
//             d.log.Error(err)
//             return nil, err
//         }
//         fmt.Printf("Un-assign IP address: %v", addr.IP)
//         ...
//     }
//
//     func (d *Descriptor) Update(key string, oldItem, newItem *mymodel.MyModel, oldMetadata interface{}) (newMetadata interface{}, err error) { {
//         prevAddr, err := d.netallocPlugin.GetOrParseIPAddress(oldItem.IpAddress, "", netalloc.ADDR_ONLY)
//         if err != nil {
//             d.log.Error(err)
//             return nil, err
//         }
//         newAddr, err := d.netallocPlugin.GetOrParseIPAddress(newItem.IpAddress, "", netalloc.ADDR_ONLY)
//         if err != nil {
//             d.log.Error(err)
//             return nil, err
//         }
//         fmt.Printf("Changing assigned IP address from %v to %v", prevAddr.IP, newAddr.IP)
//         ...
//     }
//
//     func (d *Descriptor) Retrieve(correlate []adapter.MyModelKVWithMetadata) (retrieved []adapter.MyModelKVWithMetadata, err error) {
//         // Retrieve instances of mymodel.MyModel ...
//         // Use CorrelateRetrievedIPs to replace actual IP address with reference if it was used.
//         for _, item := range retrieved {
//             // get expected item configuration ... (store to expCfg)
//             item.IpAddress = d.netallocPlugin.CorrelateRetrievedIPs(
//                 []string{expCfg.IpAddress}, []string{item.IpAddress}, "", netalloc.ADDR_ONLY)[0]
//         }
//     }
//
// Also don't forget to include netalloc descriptors in the list of "RetrieveDependencies"
// (for IP allocations, the descriptor name is stored in the constant IPAllocDescriptorName
// defined in plugins/netalloc/descriptor)
type AddressAllocator interface {
	// CreateAddressAllocRef creates reference to an allocated IP address.
	CreateAddressAllocRef(network, iface string, getGW bool) string

	// ParseAddressAllocRef parses reference to an allocated IP address.
	ParseAddressAllocRef(addrAllocRef, expIface string) (
		network, iface string, isGW, isRef bool, err error)

	// GetAddressAllocDep reads what can be potentially a reference to an allocated
	// IP address. If <allocRef> is indeed a reference, the function returns
	// the corresponding dependency to be passed further into KVScheduler
	// from the descriptor. Otherwise <hasAllocDep> is returned as false, and
	// <allocRef> should be an actual address and not a reference.
	GetAddressAllocDep(addrOrAllocRef, expIface, depLabelPrefix string) (
		dep kvs.Dependency, hasAllocDep bool)

	// ValidateIPAddress checks validity of address reference or, if <addrOrAllocRef>
	// already contains an actual IP address, it tries to parse it.
	ValidateIPAddress(addrOrAllocRef, expIface, fieldName string, gwCheck GwValidityCheck) error

	// GetOrParseIPAddress tries to get allocated interface (or GW) IP address
	// referenced by <addrOrAllocRef> in the requested form. But if the string
	// contains/ an actual IP address instead of a reference, the address is parsed
	// using methods from the net package and returned in the requested form.
	// For ADDR_ONLY address form, the returned <addr> will have the mask unset
	// and the IP address should be accessed as <addr>.IP
	GetOrParseIPAddress(addrOrAllocRef string, expIface string, addrForm netalloc.IPAddressForm) (
		addr *net.IPNet, err error)

	// CorrelateRetrievedIPs should be used in Retrieve to correlate one or group
	// of (model-wise indistinguishable) retrieved interface or GW IP addresses
	// with the expected configuration. The method will replace retrieved addresses
	// with the corresponding allocation references from the expected configuration
	// if there are any.
	// The method returns one IP address or address-allocation reference for every
	// address from <retrievedAddrs>.
	CorrelateRetrievedIPs(expAddrsOrRefs []string, retrievedAddrs []string, expIface string,
		addrForm netalloc.IPAddressForm) []string
}
