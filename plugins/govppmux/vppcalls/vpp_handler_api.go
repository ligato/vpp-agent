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

	govppapi "git.fd.io/govpp.git/api"

	"go.ligato.io/vpp-agent/v2/plugins/vpp"
)

var Handler = vpp.RegisterHandler(vpp.HandlerDesc{
	HandlerName: "vppcore",
	HandlerType: (*VppHandlerAPI)(nil),
})

// VppHandlerAPI provides methods for core VPP functionality.
type VppHandlerAPI interface {
	// Ping sends control ping to VPP.
	Ping(context.Context) error
	// GetSession retrieves info about active session.
	GetSession(context.Context) (*SessionInfo, error)
	// GetVersion retrieves info about VPP version.
	GetVersion(context.Context) (*VersionInfo, error)
	// GetModules retrieves info about VPP API modules.
	GetModules(context.Context) ([]ModuleInfo, error)
	// GetPlugins retrieves info about loaded VPP plugins.
	GetPlugins(context.Context) ([]PluginInfo, error)
	// RunCli sends CLI commmand to VPP.
	RunCli(ctx context.Context, cmd string) (string, error)
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

// ModuleInfo contains info about VPP API module.
type ModuleInfo struct {
	Name  string
	Major uint32
	Minor uint32
	Patch uint32
}

func (m ModuleInfo) String() string {
	return fmt.Sprintf("%s %d.%d.%d", m.Name, m.Major, m.Minor, m.Patch)
}

// PluginInfo contains info about loaded VPP plugin.
type PluginInfo struct {
	Name        string
	Path        string
	Version     string
	Description string
}

func (p PluginInfo) String() string {
	return fmt.Sprintf("%s - %s", p.Name, p.Description)
}

type NewHandlerFunc func(govppapi.Channel) VppHandlerAPI

// AddVersion registers vppcalls Handler for the given version.
func AddVersion(version string, msgs []govppapi.Message, h NewHandlerFunc) {
	Handler.AddVersion(vpp.HandlerVersion{
		Version: version,
		Check: func(c vpp.Client) error {
			return c.CheckCompatiblity(msgs...)
		},
		NewHandler: func(c vpp.Client, a ...interface{}) vpp.HandlerAPI {
			ch, err := c.NewAPIChannel()
			if err != nil {
				return err
			}
			return h(ch)
		},
	})
}

// CompatibleHandler is helper for returning comptabile Handler.
func CompatibleHandler(c vpp.Client) VppHandlerAPI {
	if v := Handler.FindCompatibleVersion(c); v != nil {
		return v.NewHandler(c).(VppHandlerAPI)
	}
	return nil
}
