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
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/pkg/errors"
)

// Plugin-defined process states
const (
	running = "running"
	stopped = "stopped"
	failed  = "failed"
)

// ProcessStatus is string representation of process status
type ProcessStatus string

// ProcessAPI defines methods to manage given process
type ProcessAPI interface {
	// Start starts the process. Depending on the procedure result, the status is set to 'running' or 'failed'. Start
	// also stores *os.Process in the instance for future use.
	Start() error
	// IsAlive returns true if process is alive, or false if not or if the inner instance does not exist.
	IsAlive() bool
	// Restart briefly stops and starts the process. If the process is not running, it is started.
	Restart() error
	// Stop sends the termination signal to the process. The status is set to 'stopped' (or 'failed' if not successful).
	// Attempt to stop a non-existing process instance results in error
	Stop() error
	// Kill immediately terminates the process and releases all resources associated with it. Attempt to kill
	// a non-existing process instance results in error
	Kill() error
	// Wait for process to exit and return its process state describing its status and error (if any)
	Wait() (*os.ProcessState, error)
	// GetName returns process name
	GetName() string
	// GetCommand returns process command
	GetCommand() string
	// GetArguments returns process arguments if set
	GetArguments() []string
	// GetPid returns process ID
	GetPid() (int, error)
	// GetStartTime returns time when the process was started
	GetStartTime() int64

	// TODO another ideas: uptime, send user defined signal, ...
}

// Process is wrapper around the os.Process
type Process struct {
	sync.Mutex

	name string
	cmd  string
	args []string

	process *os.Process

	status    ProcessStatus
	startTime int64
}

// Start is an implementation of the process API
func (p *Process) Start() (err error) {
	p.Lock()
	defer p.Unlock()

	if p.process, err = p.startProcess(); err != nil {
		return err
	}

	return nil
}

// IsAlive is an implementation of the process API
func (p *Process) IsAlive() bool {
	return p.isAlive()
}

// Restart is an implementation of the process API
func (p *Process) Restart() (err error) {
	p.Lock()
	defer p.Unlock()

	if p.process == nil {
		p.process, err = p.startProcess()
		return err
	}
	if p.isAlive() {
		if err = p.stopProcess(); err != nil {
			return err
		}
	}
	p.process, err = p.startProcess()
	return err
}

// Stop is an implementation of the process API
func (p *Process) Stop() error {
	p.Lock()
	defer p.Unlock()

	return p.stopProcess()
}

// Kill is an implementation of the process API
func (p *Process) Kill() error {
	return p.forceStopProcess()
}

// Wait is an implementation of the process API
func (p *Process) Wait() (*os.ProcessState, error) {
	p.Lock()
	defer p.Unlock()

	if p.process == nil {
		return nil, errors.Errorf("process %s does not exist", p.name)
	}
	return p.process.Wait()
}

// GetName is an implementation of the process API
func (p *Process) GetName() string {
	return p.name
}

// GetCommand is an implementation of the process API
func (p *Process) GetCommand() string {
	return p.cmd
}

// GetArguments is an implementation of the process API
func (p *Process) GetArguments() []string {
	return p.args
}

// GetPid is an implementation of the process API
func (p *Process) GetPid() (int, error) {
	if p.process == nil {
		return 0, errors.Errorf("process %s does not exist", p.name)
	}
	return p.process.Pid, nil
}

// GetStartTime is an implementation of the process API
func (p *Process) GetStartTime() int64 {
	return p.startTime
}

func (p *Process) startProcess() (*os.Process, error) {
	// Prepare the new process attributes
	wd, err := os.Getwd()
	if err != nil {
		p.setStatus(failed)
		return nil, errors.Errorf("failed to get rooted path name for: %v", err)
	}
	var attr = os.ProcAttr{
		Dir:   wd,
		Env:   os.Environ(),
		Files: []*os.File{os.Stdin, nil, nil},
	}
	pArgs := append([]string{p.cmd}, p.args...)
	// Start process with given arguments and attributes
	process, err := os.StartProcess(p.cmd, pArgs, &attr)
	if err != nil {
		p.setStatus(failed)
		return nil, errors.Errorf("failed to start new process %s (cmd: %s): %v", p.name, p.cmd, err)
	}
	p.setStatus(running)
	p.startTime = time.Now().Unix()

	if err != nil {
		return nil, errors.Errorf("failed to write pid file for process %s (cmd: %s): %v", p.name, p.cmd, err)
	}

	return process, nil
}

func (p *Process) stopProcess() (err error) {
	if p.process == nil {
		return errors.Errorf("process %s does not exist", p.name)
	}
	if err = p.process.Signal(syscall.SIGTERM); err != nil {
		p.setStatus(failed)
		return err
	}
	p.setStatus(stopped)
	return nil
}

func (p *Process) forceStopProcess() (err error) {
	p.Lock()
	defer p.Unlock()

	if p.process != nil {
		if err := p.process.Signal(syscall.SIGKILL); err != nil {
			return err
		}
		if err := p.process.Release(); err != nil {
			return err
		}
		return nil
	}
	return errors.Errorf("process %s does not exist", p.name)
}

func (p *Process) isAlive() bool {
	if p.process == nil {
		return false
	}
	osProcess, err := os.FindProcess(p.process.Pid)
	if err != nil {
		return false
	}
	return osProcess.Signal(syscall.Signal(0)) == nil
}

func (p *Process) setStatus(status ProcessStatus) {
	p.status = status
}
