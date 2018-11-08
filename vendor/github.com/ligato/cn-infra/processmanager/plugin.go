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

package processmanager

import (
	"github.com/ligato/cn-infra/infra"
)

// API defines methods to create, delete or manage processes
type API interface {
	// TODO possibly add flag whether the process should be stopped on plugin close, or persisted
	// NewProcess creates new process instance with name, command to start and an optional list of arguments. New process is not
	// immediately started, process instance comprises from a set of methods to manage.
	NewProcess(name, cmd string, args ...string) ProcessAPI
	// GetProcess returns existing process instance and boolean flag whether it was found. Identifier can be process name or command
	GetProcess(identifier string) ProcessAPI
	// GetAll returns all known processes
	GetAll() []ProcessAPI
}

// Plugin implements API to create/delete/read processes
type Plugin struct {
	processes []*Process

	Deps
}

// Deps define process dependencies
type Deps struct {
	infra.PluginDeps
}

// Init does nothing for process manager plugin
func (p *Plugin) Init() error {
	p.Log.Debugf("Initializing process manager plugin")

	return nil
}

// Close does nothing for process manager plugin
func (p *Plugin) Close() error {
	return nil
}

// String return string representation of the plugin
func (p *Plugin) String() string {
	return p.PluginName.String()
}

// NewProcess creates new process and adds it to the process list
func (p *Plugin) NewProcess(name, cmd string, args ...string) ProcessAPI {
	if name == "" {
		name = cmd
	}
	newProcess := &Process{
		name: name,
		cmd:  cmd,
		args: args,
	}
	p.processes = append(p.processes, newProcess)
	return newProcess
}

// GetProcess uses identifier (name or command) to find a desired process. Also returns boolean flag whether
// the instance was found
func (p *Plugin) GetProcess(identifier string) ProcessAPI {
	for _, process := range p.processes {
		if process.name == identifier || process.cmd == identifier {
			return process
		}
	}
	return nil
}

// GetAll returns all processes created so far
func (p *Plugin) GetAll() []ProcessAPI {
	var processes []ProcessAPI
	for _, process := range p.processes {
		processes = append(processes, process)
	}
	return processes
}
