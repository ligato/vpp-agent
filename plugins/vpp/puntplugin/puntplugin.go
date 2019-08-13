// Copyright (c) 2018 Cisco and/or its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:generate descriptor-adapter --descriptor-name IPPuntRedirect --value-type *vpp_punt.IPRedirect --import "github.com/ligato/vpp-agent/api/models/vpp/punt" --output-dir "descriptor"
//go:generate descriptor-adapter --descriptor-name PuntToHost --value-type *vpp_punt.ToHost --import "github.com/ligato/vpp-agent/api/models/vpp/punt" --output-dir "descriptor"
//go:generate descriptor-adapter --descriptor-name PuntException --value-type *vpp_punt.Exception --import "github.com/ligato/vpp-agent/api/models/vpp/punt" --output-dir "descriptor"

package puntplugin

import (
	"strings"

	govppapi "git.fd.io/govpp.git/api"
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/health/statuscheck"
	"github.com/ligato/cn-infra/infra"
	"github.com/pkg/errors"

	vpp_punt "github.com/ligato/vpp-agent/api/models/vpp/punt"
	"github.com/ligato/vpp-agent/pkg/models"
	"github.com/ligato/vpp-agent/plugins/govppmux"
	kvs "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/plugins/vpp/ifplugin"
	"github.com/ligato/vpp-agent/plugins/vpp/puntplugin/descriptor"
	"github.com/ligato/vpp-agent/plugins/vpp/puntplugin/descriptor/adapter"
	"github.com/ligato/vpp-agent/plugins/vpp/puntplugin/vppcalls"

	_ "github.com/ligato/vpp-agent/plugins/vpp/puntplugin/vppcalls/vpp1901"
	_ "github.com/ligato/vpp-agent/plugins/vpp/puntplugin/vppcalls/vpp1904"
	_ "github.com/ligato/vpp-agent/plugins/vpp/puntplugin/vppcalls/vpp1908"
)

// PuntPlugin configures VPP punt to host or unix domain socket entries and IP redirect entries using GoVPP.
type PuntPlugin struct {
	Deps

	// GoVPP
	vppCh govppapi.Channel

	// handler
	puntHandler vppcalls.PuntVppAPI

	// descriptors
	ipRedirectDescriptor    *descriptor.IPRedirectDescriptor
	toHostDescriptor        *descriptor.PuntToHostDescriptor
	puntExceptionDescriptor *descriptor.PuntExceptionDescriptor
}

// Deps lists dependencies of the punt plugin.
type Deps struct {
	infra.PluginDeps
	KVScheduler  kvs.KVScheduler
	GoVppmux     govppmux.API
	IfPlugin     ifplugin.API
	PublishState datasync.KeyProtoValWriter     // optional
	StatusCheck  statuscheck.PluginStatusWriter // optional
}

// Init registers STN-related descriptors.
func (p *PuntPlugin) Init() (err error) {
	// GoVPP channels
	if p.vppCh, err = p.GoVppmux.NewAPIChannel(); err != nil {
		return errors.Errorf("failed to create GoVPP API channel: %v", err)
	}

	// init punt handler
	p.puntHandler = vppcalls.CompatiblePuntVppHandler(p.vppCh, p.IfPlugin.GetInterfaceIndex(), p.Log)

	// init and register IP punt redirect
	p.ipRedirectDescriptor = descriptor.NewIPRedirectDescriptor(p.puntHandler, p.Log)
	ipRedirectDescriptor := adapter.NewIPPuntRedirectDescriptor(p.ipRedirectDescriptor.GetDescriptor())
	err = p.KVScheduler.RegisterKVDescriptor(ipRedirectDescriptor)
	if err != nil {
		return err
	}

	// init and register punt descriptor
	p.toHostDescriptor = descriptor.NewPuntToHostDescriptor(p.puntHandler, p.Log)
	toHostDescriptor := adapter.NewPuntToHostDescriptor(p.toHostDescriptor.GetDescriptor())
	err = p.KVScheduler.RegisterKVDescriptor(toHostDescriptor)
	if err != nil {
		return err
	}

	// init and register punt exception descriptor
	p.puntExceptionDescriptor = descriptor.NewPuntExceptionDescriptor(p.puntHandler, p.Log)
	exceptionDescriptor := adapter.NewPuntExceptionDescriptor(p.puntExceptionDescriptor.GetDescriptor())
	err = p.KVScheduler.RegisterKVDescriptor(exceptionDescriptor)
	if err != nil {
		return err
	}

	// FIXME: temporary workaround for publishing registered sockets
	p.toHostDescriptor.RegisterSocketFn = func(register bool, toHost *vpp_punt.ToHost, socketPath string) {
		if p.PublishState == nil {
			return
		}
		key := strings.Replace(models.Key(toHost), "config/", "status/", -1)
		if register {
			puntToHost := *toHost
			puntToHost.SocketPath = socketPath
			if err := p.PublishState.Put(key, &puntToHost, datasync.WithClientLifetimeTTL()); err != nil {
				p.Log.Errorf("publishing registered punt socket failed: %v", err)
			}
		} else {
			if err := p.PublishState.Put(key, nil); err != nil {
				p.Log.Errorf("publishing unregistered punt socket failed: %v", err)
			}
		}
	}
	p.puntExceptionDescriptor.RegisterSocketFn = func(register bool, puntExc *vpp_punt.Exception, socketPath string) {
		if p.PublishState == nil {
			return
		}
		key := strings.Replace(models.Key(puntExc), "config/", "status/", -1)
		if register {
			punt := *puntExc
			punt.SocketPath = socketPath
			if err := p.PublishState.Put(key, &punt, datasync.WithClientLifetimeTTL()); err != nil {
				p.Log.Errorf("publishing registered punt exception socket failed: %v", err)
			}
		} else {
			if err := p.PublishState.Put(key, nil); err != nil {
				p.Log.Errorf("publishing unregistered punt exception socket failed: %v", err)
			}
		}
	}

	return nil
}

// AfterInit registers plugin with StatusCheck.
func (p *PuntPlugin) AfterInit() error {
	if p.StatusCheck != nil {
		p.StatusCheck.Register(p.PluginName, nil)
	}
	return nil
}
