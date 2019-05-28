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

	"github.com/ligato/cn-infra/exec/processmanager/status"
)

// POptions is common object which holds all selected options
type POptions struct {
	// args
	args []string

	// restarts
	restart int32

	// writer
	outWriter io.Writer
	errWriter io.Writer

	// detach
	detach bool

	// environment variables
	environ []string

	// template
	template     bool
	runOnStartup bool

	// notify
	notifyChan chan status.ProcessStatus

	// auto-terminate
	autoTerm bool
}

// POption is helper function to set process options
type POption func(*POptions)

// Args if process should start with arguments
func Args(args ...string) POption {
	return func(p *POptions) {
		p.args = args
	}
}

// Restarts defines number of automatic restarts of given process
func Restarts(restart int32) POption {
	return func(p *POptions) {
		p.restart = restart
	}
}

// Writer allows to use custom writer instance. Can be defined with nil parameters, in such a case
// standard output will be used
func Writer(outW, errW io.Writer) POption {
	return func(p *POptions) {
		p.outWriter, p.errWriter = outW, errW
	}
}

// Detach process from parent after start, so it can survive after parent process is terminated
func Detach() POption {
	return func(p *POptions) {
		p.detach = true
	}
}

// EnvVar allows to set custom environment variables. If not set, os.Environ is used instead
func EnvVar(env []string) POption {
	return func(p *POptions) {
		p.environ = env
	}
}

// Template will be created for given process. Process template also requires a flag whether the process
// should be started automatically with plugin
func Template(runOnStartup bool) POption {
	return func(p *POptions) {
		p.template = true
		p.runOnStartup = runOnStartup
	}
}

// Notify will send process status change notifications to the provided channel
// Note: caller should not close the channel, since plugin is a sender, it handles the close
func Notify(notifyChan chan status.ProcessStatus) POption {
	return func(p *POptions) {
		p.notifyChan = notifyChan
	}
}

// AutoTerminate causes that zombie processes are automatically terminated
func AutoTerminate() POption {
	return func(p *POptions) {
		p.autoTerm = true
	}
}
