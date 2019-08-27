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
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/ligato/cn-infra/exec/processmanager/status"

	pm "github.com/ligato/cn-infra/exec/processmanager"
	"github.com/ligato/cn-infra/infra"
	"github.com/pkg/errors"
)

// Supervisor is a simple interface to gather information about running programs
type Supervisor interface {
	// GetProgramNames returns names of all running program instances
	GetProgramNames() []string
	// GetProgramByName returns the process instance of a given program
	GetProgramByName(name string) pm.ProcessInstance
}

// Plugin is a supervisor plugin representation
type Plugin struct {
	// a map of executed programs. Note that the map entries are never deleted,
	// the entry is still present with last returned state.
	mx       sync.Mutex
	programs map[string]*processWithStateChan

	// hook channel where all executed programs send their process state
	hookEventChan chan *processEvent
	hookDoneChan  chan struct{}

	// supervisor configuration
	config *Config

	wg sync.WaitGroup

	Deps
}

// Deps are supervisor plugin dependencies
type Deps struct {
	infra.PluginDeps
	PM pm.ProcessManager
}

// helper structure which holds process instance and channel to access its state.
// Together with name and logger, it is stored as internal plugin cache
type processWithStateChan struct {
	process   pm.ProcessInstance
	stateChan chan status.ProcessStatus
	doneChan  chan struct{}
	svLogger  *SvLogger
}

// helper structure with program name, status and event type. The object is passed to the hook
// resolver by every process watcher
type processEvent struct {
	name      string
	state     status.ProcessStatus
	eventType EventType
}

// Init supervisor config file, start event watcher and programs
func (p *Plugin) Init() error {
	// retrieve configuration file
	if err := p.getConfig(); err != nil {
		return errors.Errorf("failed to retrieve supervisor config file: %v", err)
	}
	if p.config == nil || len(p.config.Programs) == 0 {
		return errors.Errorf("supervisor config file not defined or does not contain any program")
	}
	p.programs = make(map[string]*processWithStateChan)
	p.hookEventChan = make(chan *processEvent)
	p.hookDoneChan = make(chan struct{})

	if p.config.SvCPUAffinityMask != "" {
		p.setSupervisorCPUAffinity(os.Getpid(), p.config.SvCPUAffinityMask)
	}

	go p.watchEvents()
	// start programs in another go routine (do not block init since it may take a while)
	go p.startPrograms()

	return nil
}

// Close local resources
func (p *Plugin) Close() error {
	p.Log.Info("stopping programs")
	for _, program := range p.programs {
		if program.process.IsAlive() {
			if _, err := program.process.StopAndWait(); err != nil {
				p.Log.Errorf("failed to stop program %s: %v", program.process.GetName(), err)
			}
		}
		if err := program.svLogger.Close(); err != nil {
			p.Log.Errorf("failed to close logger: %v", err)
		}
		// terminate watcher for given supervisor process
		close(program.doneChan)
	}

	// close hook watcher when all programs are terminated
	p.wg.Wait()
	close(p.hookEventChan)

	// wait for hook watcher to finish
	<-p.hookDoneChan

	return nil
}

// GetProgramNames returns names of all running process instances
func (p *Plugin) GetProgramNames() (names []string) {
	p.mx.Lock()
	defer p.mx.Unlock()

	for name := range p.programs {
		names = append(names, name)
	}
	return names
}

// GetProgramByName returns the process instance of a given program
func (p *Plugin) GetProgramByName(reqName string) pm.ProcessInstance {
	p.mx.Lock()
	defer p.mx.Unlock()

	for name, program := range p.programs {
		if name == reqName {
			return program.process
		}
	}
	return nil
}

func (p *Plugin) startPrograms() {
	for _, program := range p.config.Programs {
		if err := validate(&program); err != nil {
			p.Log.Errorf("cannot start program %s: %v", program.Name, err)
			continue
		}
		if err := p.execute(&program); err != nil {
			p.Log.Errorf("failed to start program %s: %v", program.Name, err)
			continue
		}
		p.Log.Debugf("program %s started", program.Name)
	}
}

func (p *Plugin) execute(program *Program) error {
	if _, ok := p.programs[program.Name]; ok {
		return errors.Errorf("process with name %s already exists", program.Name)
	}

	svLogger, err := NewSvLogger(program.LogfilePath)
	if err != nil {
		return errors.Errorf("error creating logger: %v", err)
	}

	stateChan := make(chan status.ProcessStatus)
	doneChan := make(chan struct{})

	p.wg.Add(1)
	go p.watch(stateChan, doneChan, program.Name)

	var process pm.ProcessInstance
	if program.Restarts > 0 {
		process = p.PM.NewProcess(program.Name, program.ExecutablePath,
			pm.Args(program.ExecutableArgs...),
			pm.Writer(svLogger, svLogger),
			pm.Notify(stateChan),
			pm.Restarts(int32(program.Restarts)),
			pm.AutoTerminate(),
			pm.CPUAffinityMask(program.CPUAffinityMask, program.CPUAffinitySetupDelay),
		)
	} else {
		process = p.PM.NewProcess(program.Name, program.ExecutablePath,
			pm.Args(program.ExecutableArgs...),
			pm.Writer(svLogger, svLogger),
			pm.Notify(stateChan),
			pm.AutoTerminate(),
			pm.CPUAffinityMask(program.CPUAffinityMask, program.CPUAffinitySetupDelay),
		)
	}
	if err := process.Start(); err != nil {
		return errors.Errorf("error starting process: %v", err)
	}

	p.programs[program.Name] = &processWithStateChan{
		process:   process,
		stateChan: stateChan,
		doneChan:  doneChan,
		svLogger:  svLogger,
	}

	return nil
}

func (p *Plugin) watch(stateChan chan status.ProcessStatus, doneChan chan struct{}, name string) {
	defer p.wg.Done()

	for {
		select {
		case state, ok := <-stateChan:
			if !ok {
				return
			}

			// forward info to the hook
			p.hookEventChan <- &processEvent{
				name:      name,
				state:     state,
				eventType: ProcessStatus,
			}
		case <-doneChan:
			return
		}
	}
}

func (p *Plugin) setSupervisorCPUAffinity(pid int, affinity string) {
	if _, err := strconv.ParseUint(affinity, 16, 64); err != nil {
		p.Log.Errorf("Provided CPU affinity value %s for supervisor is not valid (error: %v)", affinity, err)
		return
	}
	if _, err := exec.Command("taskset", "-p", affinity, strconv.Itoa(pid)).Output(); err != nil {
		p.Log.Errorf("Failed to assign CPU affinity %s to supervisor (PID: %d): %v", affinity, pid, err)
		return
	}
	p.Log.Debugf("CPU affinity of the supervisor changed to %s", affinity)
}

func validate(program *Program) error {
	if program.Name == "" && program.ExecutablePath == "" {
		return errors.Errorf("invalid program configuration: neither program name nor binary is defined")
	}
	if program.Name == "" {
		fileName := filepath.Base(program.ExecutablePath)
		program.Name = fileName[0 : len(fileName)-len(filepath.Ext(fileName))]
	}
	if program.ExecutablePath == "" {
		return errors.Errorf("invalid program configuration: executable is not defined")
	}
	return nil
}
