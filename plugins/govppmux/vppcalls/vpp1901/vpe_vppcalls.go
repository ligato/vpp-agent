//  Copyright (c) 2018 Cisco and/or its affiliates.
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

package vpp1901

import (
	"bytes"
	"strings"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/pkg/errors"

	"github.com/ligato/vpp-agent/plugins/govppmux/vppcalls"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1901/memclnt"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpp1901/vpe"
)

func init() {
	var msgs []govppapi.Message
	msgs = append(msgs, vpe.Messages...)
	msgs = append(msgs, memclnt.Messages...)

	vppcalls.Versions["vpp1901"] = vppcalls.HandlerVersion{
		Msgs: msgs,
		New: func(ch govppapi.Channel) vppcalls.VpeVppAPI {
			return NewVpeHandler(ch)
		},
	}
}

type VpeHandler struct {
	ch govppapi.Channel
}

func NewVpeHandler(ch govppapi.Channel) *VpeHandler {
	return &VpeHandler{ch}
}

// Ping pings the VPP.
func (h *VpeHandler) Ping() error {
	req := &vpe.ControlPing{}
	reply := &vpe.ControlPingReply{}

	return h.ch.SendRequest(req).ReceiveReply(reply)
}

// GetVersionInfo retrieves version info
func (h *VpeHandler) GetVersionInfo() (*vppcalls.VersionInfo, error) {
	req := &vpe.ShowVersion{}
	reply := &vpe.ShowVersionReply{}

	if err := h.ch.SendRequest(req).ReceiveReply(reply); err != nil {
		return nil, err
	}

	info := &vppcalls.VersionInfo{
		Program:        reply.Program,
		Version:        reply.Version,
		BuildDate:      reply.BuildDate,
		BuildDirectory: reply.BuildDirectory,
	}

	return info, nil
}

// GetVpeInfo retrieves vpe information.
func (h *VpeHandler) GetVpeInfo() (*vppcalls.VpeInfo, error) {
	req := &vpe.ControlPing{}
	reply := &vpe.ControlPingReply{}

	if err := h.ch.SendRequest(req).ReceiveReply(reply); err != nil {
		return nil, err
	}

	info := &vppcalls.VpeInfo{
		PID:       reply.VpePID,
		ClientIdx: reply.ClientIndex,
	}

	{
		req := &memclnt.APIVersions{}
		reply := &memclnt.APIVersionsReply{}

		if err := h.ch.SendRequest(req).ReceiveReply(reply); err != nil {
			return nil, err
		}

		for _, v := range reply.APIVersions {
			name := string(cleanBytes(v.Name))
			name = strings.TrimSuffix(name, ".api")
			info.ModuleVersions = append(info.ModuleVersions, vppcalls.ModuleVersion{
				Name:  name,
				Major: v.Major,
				Minor: v.Minor,
				Patch: v.Patch,
			})
		}
	}

	return info, nil
}

// RunCli executes CLI command and returns output
func (h *VpeHandler) RunCli(cmd string) (string, error) {
	req := &vpe.CliInband{
		Cmd: cmd,
	}
	reply := &vpe.CliInbandReply{}

	if err := h.ch.SendRequest(req).ReceiveReply(reply); err != nil {
		return "", errors.Wrapf(err, "running VPP CLI command '%s' failed", cmd)
	}

	return reply.Reply, nil
}

func cleanBytes(b []byte) []byte {
	return bytes.SplitN(b, []byte{0x00}, 2)[0]
}
