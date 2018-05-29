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

package l3plugin

import (
	"github.com/ligato/vpp-agent/plugins/vpp/l3plugin/vppcalls"
	"github.com/onsi/gomega"
	"net"
	"testing"
)

var routeOne = &vppcalls.Route{
	VrfID: 0,
	DstAddr: net.IPNet{
		IP:   net.ParseIP("10.1.1.0").To4(),
		Mask: net.CIDRMask(24, 8*net.IPv4len),
	},
	NextHopAddr: net.ParseIP("192.168.1.1").To4(),
	OutIface:    1,
	Weight:      5,
}

var routeTwo = &vppcalls.Route{
	VrfID: 0,
	DstAddr: net.IPNet{
		IP:   net.ParseIP("172.16.1.0").To4(),
		Mask: net.CIDRMask(24, 8*net.IPv4len),
	},
	NextHopAddr: net.ParseIP("10.10.1.1").To4(),
	OutIface:    2,
	Weight:      5,
}

var routeThree = &vppcalls.Route{
	VrfID: 0,
	DstAddr: net.IPNet{
		IP:   net.ParseIP("172.16.1.0").To4(),
		Mask: net.CIDRMask(24, 8*net.IPv4len),
	},
	NextHopAddr: net.ParseIP("10.10.1.1").To4(),
	OutIface:    2,
	Weight:      5,
}

var routeThreeW = &vppcalls.Route{
	VrfID: 0,
	DstAddr: net.IPNet{
		IP:   net.ParseIP("172.16.1.0").To4(),
		Mask: net.CIDRMask(24, 8*net.IPv4len),
	},
	NextHopAddr: net.ParseIP("10.10.1.1").To4(),
	OutIface:    2,
	Weight:      10,
}

func TestDiffRoutesAddedOnly(t *testing.T) {
	gomega.RegisterTestingT(t)

	routesOld := []*vppcalls.Route{}

	routes := []*vppcalls.Route{
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

	routesOld := []*vppcalls.Route{
		routeOne,
		routeTwo,
	}

	routes := []*vppcalls.Route{}

	cfg := RouteConfigurator{}
	del, add := cfg.diffRoutes(routes, routesOld)
	gomega.Expect(add).To(gomega.BeEmpty())
	gomega.Expect(del).NotTo(gomega.BeEmpty())
	gomega.Expect(del[0]).To(gomega.BeEquivalentTo(routeOne))
	gomega.Expect(del[1]).To(gomega.BeEquivalentTo(routeTwo))
}

func TestDiffRoutesOneAdded(t *testing.T) {
	gomega.RegisterTestingT(t)

	routesOld := []*vppcalls.Route{
		routeOne,
	}

	routes := []*vppcalls.Route{
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

	routesOld := []*vppcalls.Route{
		routeTwo,
		routeOne,
	}

	routes := []*vppcalls.Route{
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

	routesOld := []*vppcalls.Route{
		routeThree,
	}

	routes := []*vppcalls.Route{
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

	routesOld := []*vppcalls.Route{
		routeOne,
		routeTwo,
		routeThree,
	}

	routes := []*vppcalls.Route{
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
