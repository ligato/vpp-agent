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

package vpp2106

import (
	"context"
	"fmt"
	"net"

	govppapi "go.fd.io/govpp/api"
	"go.ligato.io/cn-infra/v2/logging"

	"go.ligato.io/vpp-agent/v3/examples/customize/custom_vpp_plugin/binapi/syslog"
	"go.ligato.io/vpp-agent/v3/examples/customize/custom_vpp_plugin/syslog/vppcalls"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2106"
)

func init() {
	// Register this handler implementation for VPP 20.05 in the vppcalls package.
	vppcalls.AddHandlerVersion(vpp2106.Version, syslog.AllMessages(), NewHandler)
}

const DefaultMaxMsgSize = 480

type Handler struct {
	log logging.Logger
	rpc syslog.RPCService
}

func NewHandler(ch govppapi.Channel, log logging.Logger) vppcalls.SyslogVppAPI {
	return &Handler{
		log: log,
		rpc: syslog.NewServiceClient(ch),
	}
}

func (h *Handler) SetSender(sender vppcalls.SenderConfig) error {
	if sender.Source.To4() == nil {
		return fmt.Errorf("source (%v) must be IPv4 address", sender.Source)
	}
	if sender.Collector.To4() == nil {
		return fmt.Errorf("collector (%v) must be IPv4 address", sender.Source)
	}

	req := &syslog.SyslogSetSender{
		CollectorPort: uint16(sender.Port),
		VrfID:         ^uint32(0),
		MaxMsgSize:    DefaultMaxMsgSize,
	}
	copy(req.SrcAddress[:], sender.Source.To4())
	copy(req.CollectorAddress[:], sender.Collector.To4())

	h.log.Debugf("SetSender: %+v", req)

	_, err := h.rpc.SyslogSetSender(context.TODO(), req)
	if err != nil {
		return err
	}

	return nil
}

func (h *Handler) GetSender() (*vppcalls.SenderConfig, error) {
	cfg, err := h.rpc.SyslogGetSender(context.TODO(), &syslog.SyslogGetSender{})
	if err != nil {
		return nil, err
	}

	h.log.Debugf("GetSender: %+v", cfg)

	var srcIP, collectorIP net.IP
	srcIP = make(net.IP, 4)
	collectorIP = make(net.IP, 4)
	copy(srcIP, cfg.SrcAddress[:])
	copy(collectorIP, cfg.CollectorAddress[:])

	sender := &vppcalls.SenderConfig{}

	// This is a workaround for disabling syslog sender
	if cfg.CollectorPort != 1 {
		sender.Port = int(cfg.CollectorPort)
		if !srcIP.IsUnspecified() {
			sender.Source = srcIP
		}
		if !collectorIP.IsUnspecified() {
			sender.Collector = collectorIP
		}
	}

	return sender, nil
}

func (h *Handler) DisableSender() error {
	// VPP does not allow disabling syslog sender, so we set port to 1 as workaround
	req := &syslog.SyslogSetSender{
		SrcAddress:       syslog.IP4Address{0, 0, 0, 1},
		CollectorAddress: syslog.IP4Address{0, 0, 0, 1},
		CollectorPort:    1,
		VrfID:            ^uint32(0),
		MaxMsgSize:       DefaultMaxMsgSize,
	}

	h.log.Debugf("SetSender: %+v", req)

	_, err := h.rpc.SyslogSetSender(context.TODO(), req)
	if err != nil {
		return err
	}

	return nil
}
