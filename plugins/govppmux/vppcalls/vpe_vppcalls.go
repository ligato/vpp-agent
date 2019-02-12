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

package vppcalls

import (
	"bytes"
	"fmt"
	"strings"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/memclnt"
	"github.com/ligato/vpp-agent/plugins/vpp/binapi/vpe"
)

func init() {
	Versions["vpp1901"] = HandlerVersion{
		Msgs: append(vpe.Messages, memclnt.Messages...),
		New: func(ch govppapi.Channel) VpeVppAPI {
			return &VpeHandler{ch}
		},
	}
}

type VpeHandler struct {
	ch govppapi.Channel
}

// GetVersionInfo retrieves version info
func (h *VpeHandler) GetVersionInfo() (*VersionInfo, error) {
	req := &vpe.ShowVersion{}
	reply := &vpe.ShowVersionReply{}

	if err := h.ch.SendRequest(req).ReceiveReply(reply); err != nil {
		return nil, err
	} else if reply.Retval != 0 {
		return nil, fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	info := &VersionInfo{
		Program:        reply.Program,
		Version:        reply.Version,
		BuildDate:      reply.BuildDate,
		BuildDirectory: reply.BuildDirectory,
	}

	return info, nil
}

// GetVpeInfo retrieves vpe information.
func (h *VpeHandler) GetVpeInfo() (*VpeInfo, error) {
	req := &vpe.ControlPing{}
	reply := &vpe.ControlPingReply{}

	if err := h.ch.SendRequest(req).ReceiveReply(reply); err != nil {
		return nil, err
	}

	info := &VpeInfo{
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
			info.ModuleVersions = append(info.ModuleVersions, ModuleVersion{
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
		return "", err
	} else if reply.Retval != 0 {
		return "", fmt.Errorf("%s returned %d", reply.GetMessageName(), reply.Retval)
	}

	return reply.Reply, nil
}

func cleanBytes(b []byte) []byte {
	return bytes.SplitN(b, []byte{0x00}, 2)[0]
}
