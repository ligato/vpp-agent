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

package syslog

import (
	"go.ligato.io/cn-infra/v2/infra"
	"go.ligato.io/cn-infra/v2/logging"

	"go.ligato.io/vpp-agent/v3/examples/customize/custom_vpp_plugin/syslog/descriptor"
	"go.ligato.io/vpp-agent/v3/examples/customize/custom_vpp_plugin/syslog/descriptor/adapter"
	"go.ligato.io/vpp-agent/v3/examples/customize/custom_vpp_plugin/syslog/vppcalls"
	"go.ligato.io/vpp-agent/v3/plugins/govppmux"
	"go.ligato.io/vpp-agent/v3/plugins/kvscheduler"
	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"

	// This blank import registers vppcalls implementation for VPP 20.05
	_ "go.ligato.io/vpp-agent/v3/examples/customize/custom_vpp_plugin/syslog/vppcalls/vpp2005"
)

// This go generate directive generates adapter code for conversion between
// generic KV descriptor API and Syslog model API defined in proto.
//go:generate descriptor-adapter --output-dir descriptor --descriptor-name SyslogSender --value-type *vpp_syslog.Sender --import "go.ligato.io/vpp-agent/v3/examples/customize/custom_vpp_plugin/proto/custom/vpp/syslog"

// SyslogPlugin is a Ligato plugin that initializes vppcalls handler and
// registers KV descriptors providing API to VPP plugin syslog.
type SyslogPlugin struct {
	Deps

	handler vppcalls.SyslogVppAPI

	senderDescriptor *descriptor.SenderDescriptor
}

type Deps struct {
	infra.PluginDeps
	KVScheduler kvs.KVScheduler
	VPP         govppmux.API
}

func NewSyslogPlugin() *SyslogPlugin {
	p := &SyslogPlugin{}
	p.PluginName = "vpp-syslog-plugin"
	p.KVScheduler = &kvscheduler.DefaultPlugin
	p.VPP = &govppmux.DefaultPlugin
	p.Log = logging.ForPlugin(p.String())
	return p
}

func (p *SyslogPlugin) Init() (err error) {
	// Get compatible VPP calls handler
	p.handler = vppcalls.CompatibleVppHandler(p.VPP, p.Log)

	// Init and register KV descriptor
	p.senderDescriptor = descriptor.NewSenderDescriptor(p.handler, p.Log)
	senderDescriptor := adapter.NewSyslogSenderDescriptor(p.senderDescriptor.GetDescriptor())
	err = p.KVScheduler.RegisterKVDescriptor(senderDescriptor)
	if err != nil {
		return err
	}

	return nil
}
