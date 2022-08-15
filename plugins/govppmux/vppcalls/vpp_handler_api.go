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

package vppcalls

import (
	"context"
	"fmt"

	govppapi "go.fd.io/govpp/api"
	"go.ligato.io/cn-infra/v2/logging"

	"go.ligato.io/vpp-agent/v3/plugins/vpp"
)

// VppCoreAPI provides methods for core VPP functionality.
type VppCoreAPI interface {
	// Ping sends control ping to VPP.
	Ping(context.Context) error
	// RunCli sends CLI command to VPP.
	RunCli(ctx context.Context, cmd string) (string, error)
	// GetVersion retrieves info about VPP version.
	GetVersion(context.Context) (*VersionInfo, error)
	// GetSession retrieves info about active session.
	GetSession(context.Context) (*SessionInfo, error)
	// GetModules retrieves info about VPP API modules.
	GetModules(context.Context) ([]APIModule, error)
	// GetPlugins retrieves info about loaded VPP plugins.
	GetPlugins(context.Context) ([]PluginInfo, error)
	// GetThreads retrieves info about VPP threads.
	GetThreads(ctx context.Context) ([]ThreadInfo, error)
}

// SessionInfo contains info about VPP session.
type SessionInfo struct {
	PID       uint32
	ClientIdx uint32
	Uptime    float64
}

// VersionInfo contains VPP version info.
type VersionInfo struct {
	Program        string
	Version        string
	BuildDate      string
	BuildDirectory string
}

// Release returns version in shortened format YY.MM that describes release.
func (v VersionInfo) Release() string {
	if len(v.Version) < 5 {
		return ""
	}
	return v.Version[:5]
}

// APIModule contains info about VPP API module.
type APIModule struct {
	Name  string
	Major uint32
	Minor uint32
	Patch uint32
}

func (m APIModule) String() string {
	return fmt.Sprintf("%s %d.%d.%d", m.Name, m.Major, m.Minor, m.Patch)
}

// PluginInfo contains info about loaded VPP plugin.
type PluginInfo struct {
	Name        string
	Path        string
	Version     string
	Description string
}

// ThreadInfo wraps all thread data counters.
type ThreadInfo struct {
	Name      string
	ID        uint32
	Type      string
	PID       uint32
	CPUID     uint32
	Core      uint32
	CPUSocket uint32
}

func (p PluginInfo) String() string {
	return fmt.Sprintf("%s - %s", p.Name, p.Description)
}

var Handler = vpp.RegisterHandler(vpp.HandlerDesc{
	Name:       "core",
	HandlerAPI: (*VppCoreAPI)(nil),
	NewFunc:    (*NewHandlerFunc)(nil),
})

type NewHandlerFunc func(vpp.Client) VppCoreAPI

// AddVersion registers vppcalls Handler for the given version.
func AddVersion(version vpp.Version, msgs []govppapi.Message, h NewHandlerFunc) {
	Handler.AddVersion(vpp.HandlerVersion{
		Version: version,
		Check: func(c vpp.Client) error {
			return c.CheckCompatiblity(msgs...)
		},
		NewHandler: func(c vpp.Client, a ...interface{}) vpp.HandlerAPI {
			return h(c)
		},
		New: h,
	})
}

func NewHandler(c vpp.Client) (VppCoreAPI, error) {
	v, err := Handler.GetCompatibleVersion(c)
	if err != nil {
		return nil, err
	}
	return v.New.(NewHandlerFunc)(c), nil
}

// CompatibleHandler is helper for returning compatible Handler.
func CompatibleHandler(c vpp.Client) VppCoreAPI {
	v, err := NewHandler(c)
	if err != nil {
		logging.Warn(err)
	}
	return v
}
