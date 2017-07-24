package l3

// Prefixes
const (
	// RoutesPrefix is the relative key prefix for routes.
	RoutesPrefix = "vpp/config/v1/vrf/0/fib" //TODO <VRF>
)

// RouteKey returns the keys used in ETCD to store vpp routes for vpp instance.
func RouteKey() string {
	return RoutesPrefix
}
