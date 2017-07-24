package l3plugin

import (
	"bytes"
	log "github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/cn-infra/utils/addrs"
	"github.com/ligato/vpp-agent/defaultplugins/l3plugin/model/l3"
	"net"
	"sort"
)

// Route represents a forward IP route entry.
type Route struct {
	vrfID    uint32
	destAddr net.IPNet
	nexthop  NextHop
}

// NextHop defines the parameters of gateway to which packets should be forwarded
// when a given routing table entry is applied.
type NextHop struct {
	addr   net.IP
	intf   uint32
	weight uint32
}

// SortedRoutes type is used to implement sort interface for slice of Route
type SortedRoutes []*Route

// Returns length of slice
// Implements sort.Interface
func (arr SortedRoutes) Len() int {
	return len(arr)
}

// Swap swaps two items in slice identified by indexes
// Implements sort.Interface
func (arr SortedRoutes) Swap(i, j int) {
	arr[i], arr[j] = arr[j], arr[i]
}

// Less returns true if the item in slice at index i in slice
// should be sorted before the element with index j
// Implements sort.Interface
func (arr SortedRoutes) Less(i, j int) bool {
	return lessRoute(arr[i], arr[j])
}

func eqRoutes(a *Route, b *Route) bool {
	return a.vrfID == b.vrfID &&
		bytes.Equal(a.destAddr.IP, b.destAddr.IP) &&
		bytes.Equal(a.destAddr.Mask, b.destAddr.Mask) &&
		bytes.Equal(a.nexthop.addr, b.nexthop.addr) &&
		a.nexthop.intf == b.nexthop.intf &&
		a.nexthop.weight == b.nexthop.weight
}

func lessRoute(a *Route, b *Route) bool {
	if a.vrfID != b.vrfID {
		return a.vrfID < b.vrfID
	}
	if !bytes.Equal(a.destAddr.IP, b.destAddr.IP) {
		return bytes.Compare(a.destAddr.IP, b.destAddr.IP) < 0
	}
	if !bytes.Equal(a.destAddr.Mask, b.destAddr.Mask) {
		return bytes.Compare(a.destAddr.Mask, b.destAddr.Mask) < 0
	}
	if !bytes.Equal(a.nexthop.addr, b.nexthop.addr) {
		return bytes.Compare(a.nexthop.addr, b.nexthop.addr) < 0
	}
	if a.nexthop.intf != b.nexthop.intf {
		return a.nexthop.intf < b.nexthop.intf
	}
	return a.nexthop.weight < b.nexthop.weight

}

func (plugin *RouteConfigurator) protoRoutesToStruct(r *l3.StaticRoutes) []*Route {
	var result []*Route
	if r.Ip != nil {
		for _, ipAddress := range r.Ip {
			var (
				ifindex uint32
				exists  bool
			)
			if ipAddress == nil || ipAddress.DestinationAddress == "" {
				continue
			}
			parsedDestIP, isIpv6, err := addrs.ParseIPWithPrefix(ipAddress.DestinationAddress)
			if err != nil {
				log.Error(err)
				continue
			}
			vrfID := ipAddress.VrfId
			for _, nextHop := range ipAddress.NextHops {
				name := nextHop.OutgoingInterface
				ifindex, _, exists = plugin.SwIfIndexes.LookupIdx(name)
				if name != "" && !exists {
					log.WithField("Interface", name).Warn("Interface not found next hop skipped")
					continue
				}
				if !exists {
					ifindex = nextHopOutgoingIfUnset
				}
				nextHopIP := net.ParseIP(nextHop.Address)
				if isIpv6 {
					nextHopIP = nextHopIP.To16()
				} else {
					nextHopIP = nextHopIP.To4()
				}
				route := &Route{
					vrfID,
					*parsedDestIP,
					NextHop{nextHopIP, ifindex, nextHop.Weight},
				}
				result = append(result, route)
			}
		}
	}
	return result
}

func (plugin *RouteConfigurator) diffRoutes(new []*Route, old []*Route) (toBeDeleted []*Route, toBeAdded []*Route) {
	newSorted := SortedRoutes(new)
	oldSorted := SortedRoutes(old)
	sort.Sort(newSorted)
	sort.Sort(oldSorted)

	//compare
	i := 0
	j := 0
	for i < len(newSorted) && j < len(oldSorted) {
		if eqRoutes(newSorted[i], oldSorted[j]) {
			i++
			j++
		} else {
			if lessRoute(newSorted[i], oldSorted[j]) {
				toBeAdded = append(toBeAdded, newSorted[i])
				i++
			} else {
				toBeDeleted = append(toBeDeleted, oldSorted[j])
				j++
			}
		}
	}

	for ; i < len(newSorted); i++ {
		toBeAdded = append(toBeAdded, newSorted[i])
	}

	for ; j < len(oldSorted); j++ {
		toBeDeleted = append(toBeDeleted, oldSorted[j])
	}
	return
}
