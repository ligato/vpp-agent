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
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/ligato/cn-infra/exec/processmanager/status"
	"github.com/pkg/errors"
)

// Marked defines that the process should be always restarted
const infiniteRestarts = -1

// DefaultPDeathSignal is default signal used for parent death process attribute
var DefaultPDeathSignal = syscall.SIGKILL

func (p *Process) startProcess() (cmd *exec.Cmd, err error) {
	cmd, err = defaultProcessAttrs(p.cmd)
	if err != nil {
		return nil, err
	}

	// if options are set, adjust command attributes, otherwise set last required fields to prepare the command
	if p.options != nil {
		// args
		cmd.Args = append(cmd.Args, p.options.args...)
		// writer
		if p.options.outWriter != nil {
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				p.log.Errorf("failed to get stdout pipe: %v", err)
			}
			p.watchOutput(p.options.outWriter, stdout)
		}
		if p.options.errWriter != nil {
			errOut, err := cmd.StderrPipe()
			if err != nil {
				p.log.Errorf("failed to get stderr pipe: %v", err)
			}
			p.watchOutput(p.options.errWriter, errOut)
		}
		// detach (replace default)
		if p.options.detach {
			cmd.SysProcAttr = &syscall.SysProcAttr{
				Setpgid: true,
				Pgid:    0,
			}
		}
		// environment variables
		if p.options.environ != nil {
			cmd.Env = p.options.environ
		}
	}

	err = cmd.Start()
	if err != nil {
		return nil, errors.Errorf("failed to start new process (cmd: %s): %v", p.cmd, err)
	}
	p.startTime = time.Now()

	// now the process is running, start the status watcher
	if !p.isWatched {
		go p.watch()
	}

	if cmd.Process != nil {
		_, err = p.sh.ReadStatusFromPID(cmd.Process.Pid)
	}

	return cmd, err
}

func defaultProcessAttrs(cmd string) (*exec.Cmd, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, errors.Errorf("failed to get rooted path name for: %v", err)
	}

	return &exec.Cmd{
		Path: cmd,
		Args: append([]string{cmd}),
		Dir:  dir,
		SysProcAttr: &syscall.SysProcAttr{
			Pdeathsig: DefaultPDeathSignal,
		},
	}, nil
}

func (p *Process) stopProcess() (err error) {
	if p.command == nil || p.command.Process == nil {
		return errors.Errorf("asked to stop non-existing process instance")
	}

	if err = p.command.Process.Signal(syscall.SIGTERM); err != nil && !strings.Contains(err.Error(), alreadyFinished) {
		return errors.Errorf("process termination unsuccessful: %v", err)
	}

	p.startTime = time.Time{}
	return nil
}

func (p *Process) forceStopProcess() (err error) {
	if p.command == nil || p.command.Process == nil {
		return errors.Errorf("asked to force-stop non-existing process instance")
	}

	if err = p.command.Process.Signal(syscall.SIGKILL); err != nil && !strings.Contains(err.Error(), alreadyFinished) {
		return errors.Errorf("process forced termination unsuccessful: %v", err)
	}
	if err = p.command.Process.Release(); err != nil {
		return errors.Errorf("resource release failed: %v", err)
	}

	p.startTime = time.Time{}
	return nil
}

func (p *Process) isAlive() bool {
	if p.command == nil || p.command.Process == nil {
		return false
	}
	osProcess, err := os.FindProcess(p.command.Process.Pid)
	if err != nil {
		return false
	}
	err = osProcess.Signal(syscall.Signal(0))
	if err != nil && (strings.Contains(err.Error(), noSuchProcess) || strings.Contains(err.Error(), alreadyFinished)) {
		return false
	}
	// Error can be not nil and process may still exits (for example if process is alive but not owned by caller)
	return true
}

// WaitOnCommand waits until the command completes
func (p *Process) waitOnProcess() (*os.ProcessState, error) {
	if p.command == nil || p.command.Process == nil {
		return &os.ProcessState{}, nil
	}

	return p.command.Process.Wait()
}

// Delete stops the process and internal watcher
func (p *Process) deleteProcess() error {
	if p.command == nil || p.command.Process == nil {
		return nil
	}

	// Close the process watcher
	if p.cancelChan != nil {
		close(p.cancelChan)
	}

	p.log.Debugf("Process %s deleted", p.name)
	return nil
}

// WaitOnCommand waits until the command completes
func (p *Process) signalToProcess(signal os.Signal) error {
	if p.command == nil || p.command.Process == nil {
		p.log.Warn("Attempt to send signal to non-running process")
	}

	return p.command.Process.Signal(signal)
}

// Periodically tries to 'ping' process. If the process is unresponsive, marks it as terminated. Otherwise the process
// status is updated. If process status was changed, notification is sent. In addition, terminated processes are
// restarted if allowed by policy, and dead processes are cleaned up.
func (p *Process) watch() {
	if p.isWatched {
		p.log.Warnf("Process watcher already running")
		return
	}

	p.log.Debugf("Process %s watcher started", p.name)
	p.isWatched = true
	ticker := time.NewTicker(1 * time.Second)

	var last status.ProcessStatus
	var numRestarts int32
	var autoTerm bool
	if p.options != nil {
		numRestarts = p.options.restart
		autoTerm = p.options.autoTerm
	}

	for {
		select {
		case <-ticker.C:
			var current status.ProcessStatus
			// skip initial status since the process is not running yet
			if current == status.Initial {
				continue
			}
			if !p.isAlive() {
				current = status.Terminated
			} else {
				pStatus, err := p.GetStatus(p.GetPid())
				if err != nil {
					p.log.Warn(err)
				}
				if pStatus.State == "" {
					current = status.Unavailable
				} else {
					current = pStatus.State
				}
			}
			// identify status change
			if current != last {
				if p.GetNotificationChan() != nil {
					p.options.notifyChan <- current
				}
				// handle automatic process restarts
				if current == status.Terminated {
					if numRestarts > 0 || numRestarts == infiniteRestarts {
						go func() {
							var err error
							if p.command, err = p.startProcess(); err != nil {
								p.log.Error("attempt to restart process %s failed: %v", p.name, err)
							}
						}()
						numRestarts--
					} else {
						p.log.Debugf("no more attempts to restart process %s", p.name)
					}
				}
				// handle automatic zombie process cleanup
				if current == status.Zombie && autoTerm {
					p.log.Debugf("Terminating zombie process %d", p.GetPid())
					if _, err := p.Wait(); err != nil {
						p.log.Warnf("failed to terminate dead process: %s", p.GetPid(), err)
					}
				}
			}
			last = current
		case <-p.cancelChan:
			ticker.Stop()
			p.closeNotifyChan()
			return
		}
	}
}

// Watch output (either standard or custom). Terminates with process, since io.Copy reaches EOF.
func (p *Process) watchOutput(w io.Writer, r io.Reader) {
	go func() {
		if _, err := io.Copy(w, r); err != nil {
			p.log.Errorf("Output watcher error: %v", err)
		}
	}()
}

func (p *Process) closeNotifyChan() {
	// rescue wheel if somebody forgets to read the doc
	defer func() {
		if r := recover(); r != nil {
			p.log.Warnf("notify channel should not be closed by provider (recovered from panic: %v)", r)
		}
	}()
	if p.GetNotificationChan() != nil {
		close(p.options.notifyChan)
	}

	p.isWatched = false
	p.log.Debugf("Process %s watcher stopped", p.name)
}