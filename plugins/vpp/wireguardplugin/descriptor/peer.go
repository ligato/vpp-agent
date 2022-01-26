//  Copyright (c) 2020 Doc.ai and/or its affiliates.
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

package descriptor

import (
	"fmt"
	"net"

	"github.com/pkg/errors"
	"go.ligato.io/cn-infra/v2/logging"
	"google.golang.org/protobuf/proto"

	"go.ligato.io/vpp-agent/v3/pkg/models"
	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2009/ip_types"
	vpp_ifdescriptor "go.ligato.io/vpp-agent/v3/plugins/vpp/ifplugin/descriptor"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/wireguardplugin/wgidx"

	"go.ligato.io/vpp-agent/v3/plugins/vpp/wireguardplugin/descriptor/adapter"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/wireguardplugin/vppcalls"
	wg "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/wireguard"
)

const (
	// PeerDescriptorName is the name of the descriptor for VPP wg peer.
	PeerDescriptorName = "vpp-wg-peer"

	// Length of wireguard public-key in base64. It should be equal 32 in binary
	PeerKeyLen = 44

	// MaxU16
	MaxU16 = 0xFFFF

	// dependency labels
	wgPeerVrfTableDep = "vrf-table-exists"
)

// A list of errors:
var (
	// ErrWgPeerKeyLen is returned when public-key length has wrong size.
	ErrWgPeerKeyLen = errors.New("Invalid wireguard peer public-key length")

	// ErrWgPeerWithoutInterface is returned when wireguard interface name is empty.
	ErrWgPeerWithoutInterface = errors.New("Wireguard interface is not defined")

	// ErrWgPeerPKeepalive is returned when persistent keepalive exceeds max value.
	ErrWgPeerPKeepalive = errors.New("Persistent keepalive exceeds the limits")

	// ErrWgPeerPort is returned when udp-port exceeds max value.
	ErrWgPeerPort = errors.New("Invalid wireguard peer port")

	// ErrWgPeerEndpointMissing is returned when endpoint address was not set or set to an empty string.
	ErrWgPeerEndpointMissing = errors.Errorf("Missing endpoint address for wireguard peer")

	// ErrWgSrcAddrBad is returned when endpoint address was not set to valid IP address.
	ErrWgPeerEndpointBad = errors.New("Invalid wireguard peer endpoint")

	// ErrWgPeerAllowedIPs is returned when one of allowedIp address was not set to valid IP address.
	ErrWgPeerAllowedIPs = errors.New("Invalid wireguard peer allowedIps")
)

// WgPeerDescriptor teaches KVScheduler how to configure VPP wg peer.
type WgPeerDescriptor struct {
	log       logging.Logger
	wgHandler vppcalls.WgVppAPI
}

// NewWgPeerDescriptor creates a new instance of the wireguard interface descriptor.
func NewWgPeerDescriptor(wgHandler vppcalls.WgVppAPI, log logging.PluginLogger) *WgPeerDescriptor {
	return &WgPeerDescriptor{
		wgHandler: wgHandler,
		log:       log.NewLogger("wg-peer-descriptor"),
	}
}

// GetDescriptor returns descriptor suitable for registration (via adapter) with
// the KVScheduler.
func (d *WgPeerDescriptor) GetDescriptor() *adapter.PeerDescriptor {
	return &adapter.PeerDescriptor{
		Name:                 PeerDescriptorName,
		NBKeyPrefix:          wg.ModelPeer.KeyPrefix(),
		ValueTypeName:        wg.ModelPeer.ProtoName(),
		KeySelector:          wg.ModelPeer.IsKeyValid,
		KeyLabel:             wg.ModelPeer.StripKeyPrefix,
		ValueComparator:      d.EquivalentWgPeers,
		Validate:             d.Validate,
		Create:               d.Create,
		Delete:               d.Delete,
		Retrieve:             d.Retrieve,
		RetrieveDependencies: []string{vpp_ifdescriptor.InterfaceDescriptorName},
		WithMetadata:         true,
	}
}

func (d *WgPeerDescriptor) EquivalentWgPeers(key string, oldPeer, newPeer *wg.Peer) bool {
	// compare base fields
	return proto.Equal(oldPeer, newPeer)
}

func (d *WgPeerDescriptor) Validate(key string, peer *wg.Peer) (err error) {
	if len(peer.PublicKey) != PeerKeyLen {
		return kvs.NewInvalidValueError(ErrWgPeerKeyLen, "public_key")
	}
	if peer.WgIfName == "" {
		return kvs.NewInvalidValueError(ErrWgPeerWithoutInterface, "wg_if_name")
	}
	if peer.PersistentKeepalive > MaxU16 {
		return kvs.NewInvalidValueError(ErrWgPeerPKeepalive, "persistent_keepalive")
	}
	if peer.Endpoint == "" {
		return kvs.NewInvalidValueError(ErrWgPeerEndpointMissing, "endpoint")
	}
	if net.ParseIP(peer.Endpoint).IsUnspecified() {
		return kvs.NewInvalidValueError(ErrWgPeerEndpointBad, "endpoint")
	}
	if peer.Port > MaxU16 {
		return kvs.NewInvalidValueError(ErrWgPeerPort, "port")
	}

	for _, allowedIp := range peer.AllowedIps {
		_, err := ip_types.ParsePrefix(allowedIp)
		if err != nil {
			return kvs.NewInvalidValueError(ErrWgPeerAllowedIPs, "allowed_ips")
		}
	}
	return nil
}

// Create adds a new wireguard peer.
func (d *WgPeerDescriptor) Create(key string, peer *wg.Peer) (metadata *wgidx.WgMetadata, err error) {
	var vppWgIndex uint32
	vppWgIndex, err = d.wgHandler.AddPeer(peer)
	if err != nil {
		d.log.Error(err)
		return nil, err
	}

	metadata = &wgidx.WgMetadata{
		Index: vppWgIndex,
	}
	return metadata, err
}

// Delete removes VPP wg peers.
func (d *WgPeerDescriptor) Delete(key string, peer *wg.Peer, metadata *wgidx.WgMetadata) error {
	if metadata == nil {
		return fmt.Errorf("failed to delete peer - metadata is nil")
	}
	err := d.wgHandler.RemovePeer(metadata.Index)
	if err != nil {
		d.log.Error(err)
	}
	return err
}

// Retrieve returns all wg peers.
func (d *WgPeerDescriptor) Retrieve(correlate []adapter.PeerKVWithMetadata) (dump []adapter.PeerKVWithMetadata, err error) {
	peers, err := d.wgHandler.DumpWgPeers()
	if err != nil {
		d.log.Error(err)
		return dump, err
	}
	for _, peer := range peers {
		dump = append(dump, adapter.PeerKVWithMetadata{
			Key:    models.Key(peer),
			Value:  peer,
			Origin: kvs.FromNB,
		})
	}

	return dump, nil
}
