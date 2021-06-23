package vpp2106

import (
	"context"
	"fmt"
	"net"

	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/interface_types"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/ip_types"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106/teib"
	l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"
)

// VppAddTeibEntry adds a new TEIB entry.
func (h *TeibHandler) VppAddTeibEntry(ctx context.Context, entry *l3.TeibEntry) error {
	return h.vppAddDelTeibEntry(entry, false)
}

// VppDelTeibEntry removes an existing TEIB entry.
func (h *TeibHandler) VppDelTeibEntry(ctx context.Context, entry *l3.TeibEntry) error {
	return h.vppAddDelTeibEntry(entry, true)
}

func (h *TeibHandler) vppAddDelTeibEntry(entry *l3.TeibEntry, delete bool) error {
	peer, err := ipToAddress(entry.PeerAddr)
	if err != nil {
		return err
	}
	nh, err := ipToAddress(entry.NextHopAddr)
	if err != nil {
		return err
	}

	meta, found := h.ifIndexes.LookupByName(entry.Interface)
	if !found {
		return fmt.Errorf("interface %s not found", entry.Interface)
	}

	req := &teib.TeibEntryAddDel{
		Entry: teib.TeibEntry{
			SwIfIndex: interface_types.InterfaceIndex(meta.SwIfIndex),
			Peer:      peer,
			Nh:        nh,
			NhTableID: entry.VrfId,
		},
	}
	if !delete {
		req.IsAdd = 1
	}

	reply := &teib.TeibEntryAddDelReply{}
	return h.callsChannel.SendRequest(req).ReceiveReply(reply)
}

// DumpTeib dumps TEIB entries from VPP and fills them into the provided TEIB entry map.
func (h *TeibHandler) DumpTeib() (entries []*l3.TeibEntry, err error) {
	reqCtx := h.callsChannel.SendMultiRequest(&teib.TeibDump{})
	for {
		teibDetails := &teib.TeibDetails{}
		stop, err := reqCtx.ReceiveReply(teibDetails)
		if stop {
			break
		}
		if err != nil {
			return nil, err
		}
		entry := &l3.TeibEntry{
			PeerAddr:    ipAddrToStr(teibDetails.Entry.Peer),
			NextHopAddr: ipAddrToStr(teibDetails.Entry.Nh),
			VrfId:       teibDetails.Entry.NhTableID,
		}
		if ifName, _, exists := h.ifIndexes.LookupBySwIfIndex(uint32(teibDetails.Entry.SwIfIndex)); !exists {
			h.log.Warnf("TEIB dump: interface name for index %d not found", teibDetails.Entry.SwIfIndex)
		} else {
			entry.Interface = ifName
		}
		entries = append(entries, entry)
	}
	return
}

func ipAddrToStr(addr ip_types.Address) string {
	if addr.Af == ip_types.ADDRESS_IP6 {
		ip6Addr := addr.Un.GetIP6()
		return net.IP(ip6Addr[:]).To16().String()
	} else {
		ip4Addr := addr.Un.GetIP4()
		return net.IP(ip4Addr[:4]).To4().String()
	}
}
