// Copyright (c) 2017 Cisco and/or its affiliates.
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

package resync

import (
	"strings"
	"sync"
	"time"

	"github.com/ligato/cn-infra/infra"
	"github.com/ligato/cn-infra/logging"
)

var (
	// SingleResyncAcceptTimeout defines timeout for accepting resync start.
	SingleResyncAcceptTimeout = time.Second * 1
	// SingleResyncAckTimeout defines timeout for resync ack.
	SingleResyncAckTimeout = time.Second * 10
)

// Plugin implements Plugin interface, therefore it can be loaded with other plugins.
type Plugin struct {
	Deps

	mu            sync.Mutex
	regOrder      []string
	registrations map[string]*registration
}

// Deps groups dependencies injected into the plugin so that they are
// logically separated from other plugin fields.
type Deps struct {
	infra.PluginName
	Log logging.PluginLogger
}

// Init initializes variables.
func (p *Plugin) Init() error {
	p.registrations = make(map[string]*registration)
	return nil
}

// Close TODO set flag that ignore errors => not start Resync while agent is stopping
// TODO kill existing Resync timeout while agent is stopping
func (p *Plugin) Close() error {
	//TODO close error report channel

	p.mu.Lock()
	defer p.mu.Unlock()

	p.registrations = make(map[string]*registration)

	return nil
}

// Register function is supposed to be called in Init() by all VPP Agent plugins.
// The plugins are supposed to load current state of their objects when newResync() is called.
// The actual CreateNewObjects(), DeleteObsoleteObjects() and ModifyExistingObjects() will be orchestrated
// to ensure their proper order. If an error occurs during Resync, then new Resync is planned.
func (p *Plugin) Register(resyncName string) Registration {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, found := p.registrations[resyncName]; found {
		p.Log.WithField("resyncName", resyncName).
			Panic("You are trying to register same resync twice")
		return nil
	}
	// ensure that resync is triggered in the same order as the plugins were registered
	p.regOrder = append(p.regOrder, resyncName)

	reg := newRegistration(resyncName, make(chan StatusEvent))
	p.registrations[resyncName] = reg

	return reg
}

// DoResync can be used to start resync procedure outside of after init
func (p *Plugin) DoResync() {
	p.startResync()
}

// Call callback on plugins to create/delete/modify objects.
func (p *Plugin) startResync() {
	if len(p.regOrder) == 0 {
		p.Log.Warnf("No registrations, skipping resync")
		return
	}

	subs := strings.Join(p.regOrder, ", ")
	p.Log.Infof("Resync starting for %d registrations (%v)", len(p.regOrder), subs)

	resyncStart := time.Now()

	for _, regName := range p.regOrder {
		if reg, found := p.registrations[regName]; found {
			t := time.Now()
			p.startSingleResync(regName, reg)

			took := time.Since(t).Round(time.Millisecond)
			p.Log.Debugf("finished resync for %v took %v", regName, took)
		}
	}

	p.Log.Infof("Resync done (took: %v)", time.Since(resyncStart).Round(time.Millisecond))

	// TODO check if there ReportError (if not than report) if error occurred even during Resync
}
func (p *Plugin) startSingleResync(resyncName string, reg *registration) {
	started := newStatusEvent(Started)

	select {
	case reg.statusChan <- started:
		// accept
	case <-time.After(SingleResyncAcceptTimeout):
		p.Log.WithField("regName", resyncName).Warn("Timeout of resync start!")
		return
	}

	select {
	case <-started.ReceiveAck():
		// ack
	case <-time.After(SingleResyncAckTimeout):
		p.Log.WithField("regName", resyncName).Warn("Timeout of resync ACK!")
	}
}
