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

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
)

// EventType represents type of the given event
type EventType int

const (
	// ProcessStatus represents events about process status
	ProcessStatus EventType = 1

	// add more when needed
)

func (e EventType) String() string {
	switch e {
	case ProcessStatus:
		return "ProcessStatus"
	default:
		return fmt.Sprintf("EventType(%d)", e)
	}
}

// Config represents supervision setup where programs will be started at the beginning
// and hooks are special commands which are executed when certain event related to
// one of the processes occurs
type Config struct {
	// Bond supervisor process to given set of CPUs. Plugin uses taskset to assign process
	// to CPUs and uses the same hexadecimal format. Invalid value prints error but does
	// not terminate the process.
	// It is recommended to use this option only for testing, operating system CPU schedulers
	// are in general more superior in managing CPU cycles.
	SvCPUAffinityMask string `json:"sv-cpu-affinity-mask"`

	// A list of programs started by the supervisor.
	Programs []Program

	// A list of hooks managed by supervisor plugin. Hooks are additional commands or scripts
	// called after some specific process events.
	Hooks []Hook
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

	// Bond process to given set of CPUs. Plugin uses taskset to assign process to CPUs
	// and uses the same hexadecimal format. Invalid value prints error message but does
	// not terminate the process.
	// Note: use only when you know what you are doing, do not try to outsmart OS CPU
	// scheduling. If a program has its own config file to manage CPUs, prioritize it.
	// Keep in mind that incorrect use may slow down certain applications or that the
	// application may contain its own CPU manager which overrides this value.
	// Warning: Locking process to CPU does NOT keep other processes off that CPU.
	CPUAffinityMask string `json:"cpu-affinity-mask"`

	// This field can postpone CPU affinity setup for given time. Some processes may
	// manipulate CPU scheduling during startup, this option allows to "bypass" it,
	// waiting until the process is fully loaded and then lock it.
	CPUAffinitySetupDelay time.Duration `json:"cpu-affinity-setup-delay"`
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
