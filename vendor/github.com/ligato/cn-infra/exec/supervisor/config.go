// Copyright (c) 2019 Cisco and/or its affiliates.
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

package supervisor

import "github.com/pkg/errors"

// Event types
const (
	ProcessStatus EventType = 1

	// add more when needed
)

// EventType represents type of the given event
type EventType int

// Config represents supervision setup where programs will be started at the beginning
// and hooks are special commands which are executed when certain event related to
// one of the processes occurs
type Config struct {
	Programs []Program
	Hooks    []Hook
}

// Program is a single program representation
type Program struct {
	// Name is an optional parameter, if not set it will be derived from the executable
	Name string `json:"name"`

	// ExecutablePath is a path to the binary file, required field
	ExecutablePath string `json:"executable-path"`

	// ExecutableArgs is a list of arguments passed to the binary
	ExecutableArgs []string `json:"executable-args"`

	// LogfilePath is defined when the log output should be written to the file. No file
	// is written if not set
	LogfilePath string `json:"logfile-path"`

	// Number of automatic restarts. Note that any termination hooks are executed also
	// when the program is restarted since the operating system sends events in order
	// termination -> starting -> sleeping/idle
	Restarts int `json:"restarts"`
}

// Hook is a procedure called when a program gets into certain state.
type Hook struct {
	// Command which will be executed
	Cmd string `json:"cmd"`

	// Command arguments
	CmdArgs []string `json:"cmd-args"`
}

// NewEmptyConfig prepares empty configuration ready to populate from the file
func NewEmptyConfig() *Config {
	return &Config{
		Programs: []Program{},
		Hooks:    []Hook{},
	}
}

func (p *Plugin) getConfig() error {
	// skip if the config was already defined
	if p.config != nil {
		return nil
	}
	p.config = NewEmptyConfig()
	found, err := p.Cfg.LoadValue(p.config)
	if err != nil {
		return errors.Errorf("failed to load supervisor config file: %v", err)
	}
	if !found {
		return errors.Errorf("failed to load supervisor config file: not found")
	}

	return nil
}
