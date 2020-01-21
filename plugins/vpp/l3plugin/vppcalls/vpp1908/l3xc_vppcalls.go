//  Copyright (c) 2020 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package vpp1908

import (
	"context"
	"io"
	"net"

	"github.com/pkg/errors"

	"go.ligato.io/vpp-agent/v3/plugins/vpp"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1908/l3xc"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/l3plugin/vppcalls"
)

func (h *L3XCHandler) DumpAllL3XC(ctx context.Context) ([]vppcalls.L3XC, error) {
	return h.DumpL3XC(ctx, ^uint32(0))
}

func (h *L3XCHandler) DumpL3XC(ctx context.Context, index uint32) ([]vppcalls.L3XC, error) {
	if h.l3xc == nil {
		// no-op when disabled
		return nil, nil
	}

	dump, err := h.l3xc.DumpL3xc(ctx, &l3xc.L3xcDump{
		SwIfIndex: index,
	})
	if err != nil {
		return nil, err
	}
	l3xcs := make([]vppcalls.L3XC, 0)
	for {
		recv, err := dump.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		paths := make([]vppcalls.Path, len(recv.L3xc.Paths))
		for i, p := range recv.L3xc.Paths {
			var nextHop net.IP
			if p.Proto == l3xc.FIB_API_PATH_NH_PROTO_IP6 {
				ip6Addr := p.Nh.Address.GetIP6()
				nextHop = net.IP(ip6Addr[:]).To16()
			} else {
				ip4Addr := p.Nh.Address.GetIP4()
				nextHop = net.IP(ip4Addr[:4]).To4()
			}
			paths[i] = vppcalls.Path{
				SwIfIndex:  p.SwIfIndex,
				Weight:     p.Weight,
				Preference: p.Preference,
				NextHop:    nextHop,
			}
		}
		l3xcs = append(l3xcs, vppcalls.L3XC{
			SwIfIndex: recv.L3xc.SwIfIndex,
			IsIPv6:    recv.L3xc.IsIP6 == 1,
			Paths:     paths,
		})
	}
	return l3xcs, nil
}

func (h *L3XCHandler) UpdateL3XC(ctx context.Context, xc *vppcalls.L3XC) error {
	if h.l3xc == nil {
		return errors.WithMessage(vpp.ErrPluginDisabled, "l3xc")
	}

	paths := make([]l3xc.FibPath, len(xc.Paths))
	for i, p := range xc.Paths {
		fibPath := l3xc.FibPath{
			SwIfIndex:  p.SwIfIndex,
			Weight:     p.Weight,
			Preference: p.Preference,
			Type:       l3xc.FIB_API_PATH_TYPE_NORMAL,
		}
		fibPath.Nh, fibPath.Proto = getL3XCFibPathNhAndProto(p.NextHop)
		paths[i] = fibPath
	}
	_, err := h.l3xc.L3xcUpdate(ctx, &l3xc.L3xcUpdate{
		L3xc: l3xc.L3xc{
			SwIfIndex: xc.SwIfIndex,
			IsIP6:     boolToUint(xc.IsIPv6),
			Paths:     paths,
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (h *L3XCHandler) DeleteL3XC(ctx context.Context, index uint32, ipv6 bool) error {
	if h.l3xc == nil {
		return errors.Wrap(vpp.ErrPluginDisabled, "l3xc")
	}

	_, err := h.l3xc.L3xcDel(ctx, &l3xc.L3xcDel{
		SwIfIndex: index,
		IsIP6:     boolToUint(ipv6),
	})
	if err != nil {
		return err
	}
	return nil
}

func getL3XCFibPathNhAndProto(netIP net.IP) (nh l3xc.FibPathNh, proto l3xc.FibPathNhProto) {
	var addrUnion l3xc.AddressUnion
	if netIP.To4() == nil {
		proto = l3xc.FIB_API_PATH_NH_PROTO_IP6
		var ip6addr l3xc.IP6Address
		copy(ip6addr[:], netIP.To16())
		addrUnion.SetIP6(ip6addr)
	} else {
		proto = l3xc.FIB_API_PATH_NH_PROTO_IP4
		var ip4addr l3xc.IP4Address
		copy(ip4addr[:], netIP.To4())
		addrUnion.SetIP4(ip4addr)
	}
	return l3xc.FibPathNh{
		Address:            addrUnion,
		ViaLabel:           NextHopViaLabelUnset,
		ClassifyTableIndex: ClassifyTableIndexUnset,
	}, proto
}
