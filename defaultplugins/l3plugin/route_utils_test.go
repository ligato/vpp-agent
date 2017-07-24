package l3plugin

import (
	"github.com/onsi/gomega"
	"net"
	"testing"
)

var routeOne = &Route{
	0,
	net.IPNet{
		IP:   net.ParseIP("10.1.1.0").To4(),
		Mask: net.CIDRMask(24, 8*net.IPv4len),
	},
	NextHop{
		net.ParseIP("192.168.1.1").To4(),
		1,
		5,
	},
}

var routeTwo = &Route{
	0,
	net.IPNet{
		IP:   net.ParseIP("172.16.1.0").To4(),
		Mask: net.CIDRMask(24, 8*net.IPv4len),
	},
	NextHop{
		net.ParseIP("10.10.1.1").To4(),
		2,
		5,
	},
}

var routeThree = &Route{
	0,
	net.IPNet{
		IP:   net.ParseIP("172.16.1.0").To4(),
		Mask: net.CIDRMask(24, 8*net.IPv4len),
	},
	NextHop{
		net.ParseIP("10.10.1.1").To4(),
		2,
		5,
	},
}

var routeThreeW = &Route{
	0,
	net.IPNet{
		IP:   net.ParseIP("172.16.1.0").To4(),
		Mask: net.CIDRMask(24, 8*net.IPv4len),
	},
	NextHop{
		net.ParseIP("10.10.1.1").To4(),
		2,
		10,
	},
}

func TestDiffRoutesAddedOnly(t *testing.T) {
	gomega.RegisterTestingT(t)

	routesOld := []*Route{}

	routes := []*Route{
		routeOne,
		routeTwo,
	}

	cfg := RouteConfigurator{}
	del, add := cfg.diffRoutes(routes, routesOld)
	gomega.Expect(del).To(gomega.BeEmpty())
	gomega.Expect(add).NotTo(gomega.BeEmpty())
	gomega.Expect(add[0]).To(gomega.BeEquivalentTo(routeOne))
	gomega.Expect(add[1]).To(gomega.BeEquivalentTo(routeTwo))
}

func TestDiffRoutesDeleteOnly(t *testing.T) {
	gomega.RegisterTestingT(t)

	routesOld := []*Route{
		routeOne,
		routeTwo,
	}

	routes := []*Route{}

	cfg := RouteConfigurator{}
	del, add := cfg.diffRoutes(routes, routesOld)
	gomega.Expect(add).To(gomega.BeEmpty())
	gomega.Expect(del).NotTo(gomega.BeEmpty())
	gomega.Expect(del[0]).To(gomega.BeEquivalentTo(routeOne))
	gomega.Expect(del[1]).To(gomega.BeEquivalentTo(routeTwo))
}

func TestDiffRoutesOneAdded(t *testing.T) {
	gomega.RegisterTestingT(t)

	routesOld := []*Route{
		routeOne,
	}

	routes := []*Route{
		routeOne,
		routeTwo,
	}

	cfg := RouteConfigurator{}
	del, add := cfg.diffRoutes(routes, routesOld)
	gomega.Expect(del).To(gomega.BeEmpty())
	gomega.Expect(add).NotTo(gomega.BeEmpty())
	gomega.Expect(add[0]).To(gomega.BeEquivalentTo(routeTwo))
}

func TestDiffRoutesNoChange(t *testing.T) {
	gomega.RegisterTestingT(t)

	routesOld := []*Route{
		routeTwo,
		routeOne,
	}

	routes := []*Route{
		routeOne,
		routeTwo,
	}

	cfg := RouteConfigurator{}
	del, add := cfg.diffRoutes(routes, routesOld)
	gomega.Expect(del).To(gomega.BeEmpty())
	gomega.Expect(add).To(gomega.BeEmpty())
}

func TestDiffRoutesWeightChange(t *testing.T) {
	gomega.RegisterTestingT(t)

	routesOld := []*Route{
		routeThree,
	}

	routes := []*Route{
		routeThreeW,
	}

	cfg := RouteConfigurator{}
	del, add := cfg.diffRoutes(routes, routesOld)
	gomega.Expect(del).NotTo(gomega.BeEmpty())
	gomega.Expect(add).NotTo(gomega.BeEmpty())
	gomega.Expect(add[0]).To(gomega.BeEquivalentTo(routeThreeW))
	gomega.Expect(del[0]).To(gomega.BeEquivalentTo(routeThree))

}

func TestDiffRoutesMultipleChanges(t *testing.T) {
	gomega.RegisterTestingT(t)

	routesOld := []*Route{
		routeOne,
		routeTwo,
		routeThree,
	}

	routes := []*Route{
		routeThreeW,
		routeTwo,
	}

	cfg := RouteConfigurator{}
	del, add := cfg.diffRoutes(routes, routesOld)
	gomega.Expect(del).NotTo(gomega.BeEmpty())
	gomega.Expect(add).NotTo(gomega.BeEmpty())
	gomega.Expect(add[0]).To(gomega.BeEquivalentTo(routeThreeW))
	gomega.Expect(del[0]).To(gomega.BeEquivalentTo(routeOne))
	gomega.Expect(del[1]).To(gomega.BeEquivalentTo(routeThree))
}
