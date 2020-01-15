//  Copyright (c) 2019 Cisco and/or its affiliates.
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

package vpp1904

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"go.ligato.io/vpp-agent/v3/plugins/govppmux/vppcalls"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1904/memclnt"
	"go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp1904/vpe"
)

// Ping sends VPP control ping.
func (h *VpeHandler) Ping(ctx context.Context) error {
	_, err := h.vpe.ControlPing(ctx, new(vpe.ControlPing))
	return err
}

// GetVersion retrieves version info from VPP.
func (h *VpeHandler) GetVersion(ctx context.Context) (*vppcalls.VersionInfo, error) {
	version, err := h.vpe.ShowVersion(ctx, new(vpe.ShowVersion))
	if err != nil {
		return nil, err
	}
	info := &vppcalls.VersionInfo{
		Program:        version.Program,
		Version:        version.Version,
		BuildDate:      version.BuildDate,
		BuildDirectory: version.BuildDirectory,
	}
	return info, nil
}

// GetSession retrieves session info from VPP.
func (h *VpeHandler) GetSession(ctx context.Context) (*vppcalls.SessionInfo, error) {
	pong, err := h.vpe.ControlPing(ctx, new(vpe.ControlPing))
	if err != nil {
		return nil, err
	}
	info := &vppcalls.SessionInfo{
		PID:       pong.VpePID,
		ClientIdx: pong.ClientIndex,
	}
	return info, nil
}

// GetModules retrieves module info from VPP.
func (h *VpeHandler) GetModules(ctx context.Context) ([]vppcalls.APIModule, error) {
	versions, err := h.memclnt.APIVersions(ctx, new(memclnt.APIVersions))
	if err != nil {
		return nil, err
	}
	var modules []vppcalls.APIModule
	for _, v := range versions.APIVersions {
		modules = append(modules, vppcalls.APIModule{
			Name:  strings.TrimSuffix(cleanString(v.Name), ".api"),
			Major: v.Major,
			Minor: v.Minor,
			Patch: v.Patch,
		})
	}
	return modules, nil
}

func (h *VpeHandler) GetPlugins(ctx context.Context) ([]vppcalls.PluginInfo, error) {
	out, err := h.RunCli(ctx, "show plugins")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(out, "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("empty output for 'show plugins'")
	}

	pluginPathLine := strings.TrimSpace(lines[0])
	const pluginPathPrefix = "Plugin path is:"
	if !strings.HasPrefix(pluginPathLine, pluginPathPrefix) {
		return nil, fmt.Errorf("unexpected output for 'show plugins'")
	}
	pluginPath := strings.TrimSpace(strings.TrimPrefix(pluginPathLine, pluginPathPrefix))
	if len(pluginPath) == 0 {
		return nil, fmt.Errorf("plugin path not found in output for 'show plugins'")
	}

	var plugins []vppcalls.PluginInfo
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		var i int
		if _, err := fmt.Sscanf(fields[0], "%d.", &i); err != nil {
			continue
		}
		if i <= 0 {
			continue
		}
		plugin := vppcalls.PluginInfo{
			Name:        strings.TrimSuffix(fields[1], "_plugin.so"),
			Path:        fields[1],
			Version:     fields[2],
			Description: strings.Join(fields[3:], " "),
		}
		plugins = append(plugins, plugin)
	}

	return plugins, nil
}

// RunCli sends CLI command to VPP and returns response.
func (h *VpeHandler) RunCli(ctx context.Context, cmd string) (string, error) {
	resp, err := h.vpe.CliInband(ctx, &vpe.CliInband{
		Cmd: cmd,
	})
	if err != nil {
		return "", errors.Wrapf(err, "VPP CLI command '%s' failed", cmd)
	}
	return resp.Reply, nil
}

func cleanString(b []byte) string {
	return string(bytes.SplitN(b, []byte{0x00}, 2)[0])
}
