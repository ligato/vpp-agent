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

package statuscheck

import (
	"context"
	"sync"
	"time"

	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/health/statuscheck/model/status"
	"github.com/ligato/cn-infra/logging"
)

const (
	// Init state means that the initialization of the plugin is in progress.
	Init PluginState = "init"
	// OK state means that the plugin is healthy.
	OK PluginState = "ok"
	// Error state means that some error has occurred in the plugin.
	Error PluginState = "error"

	periodicWriteTimeout   time.Duration = time.Second * 10 // frequency of periodic writes of state data into ETCD
	periodicProbingTimeout time.Duration = time.Second * 5  // frequency of periodic plugin state probing
)

// Plugin struct holds all plugin-related data.
type Plugin struct {
	Deps

	access sync.Mutex // lock for the Plugin data

	agentStat   *status.AgentStatus             // overall agent status
	pluginStat  map[string]*status.PluginStatus // plugin's status
	pluginProbe map[string]PluginStateProbe     // registered status probes

	ctx    context.Context
	cancel context.CancelFunc // cancel can be used to cancel all goroutines and their jobs inside of the plugin
	wg     sync.WaitGroup     // wait group that allows to wait until all goroutines of the plugin have finished
}

// Deps is here to group injected dependencies of plugin
// to not mix with other plugin fields.
type Deps struct {
	Log        logging.PluginLogger       //inject
	PluginName core.PluginName            //inject
	Transport  datasync.KeyProtoValWriter //inject optional
}

// Init is the plugin entry point called by the Agent Core.
func (p *Plugin) Init() error {
	// write initial status data into ETCD
	p.agentStat = &status.AgentStatus{
		BuildVersion: core.BuildVersion,
		BuildDate:    core.BuildDate,
		State:        status.OperationalState_INIT,
		StartTime:    time.Now().Unix(),
		LastChange:   time.Now().Unix(),
	}

	// init pluginStat map
	p.pluginStat = make(map[string]*status.PluginStatus)

	// init map with plugin state probes
	p.pluginProbe = make(map[string]PluginStateProbe)

	// prepare context for all go routines
	p.ctx, p.cancel = context.WithCancel(context.Background())

	return nil
}

// AfterInit is called by the Agent Core after all plugins have been initialized.
func (p *Plugin) AfterInit() error {
	p.access.Lock()
	defer p.access.Unlock()

	// do periodic status probing for plugins that have provided the probe function
	go p.periodicProbing(p.ctx)

	// do periodic updates of the state data in ETCD
	go p.periodicUpdates(p.ctx)

	p.publishAgentData()

	// transition to OK state if there are no plugins
	if len(p.pluginStat) == 0 {
		p.agentStat.State = status.OperationalState_OK
		p.agentStat.LastChange = time.Now().Unix()
		p.publishAgentData()
	}

	return nil
}

// Close is called by the Agent Core when it's time to clean up the plugin.
func (p *Plugin) Close() error {
	p.cancel()
	p.wg.Wait()

	return nil
}

// Register a plugin for status change reporting.
func (p *Plugin) Register(pluginName core.PluginName, probe PluginStateProbe) {
	p.access.Lock()
	defer p.access.Unlock()

	stat := &status.PluginStatus{
		State:      status.OperationalState_INIT,
		LastChange: time.Now().Unix(),
	}
	p.pluginStat[string(pluginName)] = stat

	if probe != nil {
		p.pluginProbe[string(pluginName)] = probe
	}

	// write initial status data into ETCD
	p.publishPluginData(pluginName, stat)

	p.Log.Infof("Plugin %v: status check probe registered", pluginName)
}

// ReportStateChange can be used to report a change in the status of a previously registered plugin.
func (p *Plugin) ReportStateChange(pluginName core.PluginName, state PluginState, lastError error) {
	p.access.Lock()
	defer p.access.Unlock()

	stat, ok := p.pluginStat[string(pluginName)]
	if !ok {
		p.Log.Errorf("Unregistered plugin %s is reporting the state, ignoring.", pluginName)
		return
	}

	// update the state only if it has really changed
	changed := true
	if stateToProto(state) == stat.State {
		if lastError == nil && stat.Error == "" {
			changed = false
		}
		if lastError != nil && lastError.Error() == stat.Error {
			changed = false
		}
	}
	if !changed {
		return
	}

	p.Log.WithFields(map[string]interface{}{"plugin": pluginName, "state": state, "lastErr": lastError}).Info(
		"Agent plugin state update.")

	// update plugin state
	stat.State = stateToProto(state)
	stat.LastChange = time.Now().Unix()
	if lastError != nil {
		stat.Error = lastError.Error()
	} else {
		stat.Error = ""
	}
	p.publishPluginData(pluginName, stat)

	// update global state if needed
	changeGlobalState := true
	if state == OK {
		// by transition to OK state, check if all plugins are OK
		for _, s := range p.pluginStat {
			if s.State != status.OperationalState_OK {
				changeGlobalState = false
				break
			}
		}
	}
	if changeGlobalState {
		p.agentStat.State = stateToProto(state)
		p.agentStat.LastChange = time.Now().Unix()
		p.publishAgentData()
	}
}

// publishAgentData writes the current global agent state into ETCD.
func (p *Plugin) publishAgentData() error {
	p.agentStat.LastUpdate = time.Now().Unix()
	if p.Transport != nil {
		return p.Transport.Put(status.AgentStatusKey(), p.agentStat)
	}
	return nil
}

// publishPluginData writes the current plugin state into ETCD.
func (p *Plugin) publishPluginData(pluginName core.PluginName, pluginStat *status.PluginStatus) error {
	pluginStat.LastUpdate = time.Now().Unix()
	if p.Transport != nil {
		return p.Transport.Put(status.PluginStatusKey(string(pluginName)), pluginStat)
	}
	return nil
}

// publishAllData publishes global agent + all plugins state data into ETCD.
func (p *Plugin) publishAllData() {
	p.access.Lock()
	defer p.access.Unlock()

	p.publishAgentData()
	for name, s := range p.pluginStat {
		p.publishPluginData(core.PluginName(name), s)
	}
}

// periodicProbing does periodic status probing for all plugins that have registered probe functions.
func (p *Plugin) periodicProbing(ctx context.Context) {
	p.wg.Add(1)
	defer p.wg.Done()

	for {
		select {
		case <-time.After(periodicProbingTimeout):
			for pluginName, probe := range p.pluginProbe {
				state, lastErr := probe()
				p.ReportStateChange(core.PluginName(pluginName), state, lastErr)
				// just check in-between probes if the plugin is closing
				select {
				case <-ctx.Done():
					return
				default:
					continue
				}
			}

		case <-ctx.Done():
			return
		}
	}
}

// periodicUpdates does periodic writes of state data into ETCD.
func (p *Plugin) periodicUpdates(ctx context.Context) {
	p.wg.Add(1)
	defer p.wg.Done()

	for {
		select {
		case <-time.After(periodicWriteTimeout):
			p.publishAllData()

		case <-ctx.Done():
			return
		}
	}
}

// getAgentState return current global operational state of the agent.
func (p *Plugin) getAgentState() status.OperationalState {
	p.access.Lock()
	defer p.access.Unlock()
	return p.agentStat.State
}

// GetAgentStatus return current global operational state of the agent.
func (p *Plugin) GetAgentStatus() status.AgentStatus {
	p.access.Lock()
	defer p.access.Unlock()
	return *(p.agentStat)
}

// stateToProto converts agent state type into protobuf agent state type.
func stateToProto(state PluginState) status.OperationalState {
	switch state {
	case Init:
		return status.OperationalState_INIT
	case OK:
		return status.OperationalState_OK
	default:
		return status.OperationalState_ERROR
	}
}
