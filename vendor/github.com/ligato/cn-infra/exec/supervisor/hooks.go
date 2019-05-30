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
	"os/exec"
)

// Environment variables set for executed hook command
const (
	svProcessName  = "SUPERVISOR_PROCESS_NAME="
	svProcessState = "SUPERVISOR_PROCESS_STATE="
	seEventType    = "SUPERVISOR_EVENT_TYPE="
)

func (p *Plugin) watchEvents() {
	for {
		processInfo, ok := <-p.hookEventChan
		if !ok {
			p.Log.Debugf("Supervisor hook watcher ended")
			close(p.hookDoneChan)
			return
		}

		// execute all hooks with env vars set
		for _, hook := range p.config.Hooks {
			cmd := exec.Command(hook.Cmd, hook.CmdArgs...)

			cmd.Env = append(cmd.Env, svProcessName+processInfo.name)
			cmd.Env = append(cmd.Env, svProcessState+string(processInfo.state))
			cmd.Env = append(cmd.Env, seEventType+string(processInfo.eventType))

			out, err := cmd.CombinedOutput()
			if err != nil {
				p.Log.Errorf("%v", err)
			}
			if len(out) > 0 {
				p.Log.Debugf("%s", out)
			}
		}
	}
}
