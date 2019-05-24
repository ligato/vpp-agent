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

package main

import (
	"fmt"
	"time"

	"github.com/ligato/cn-infra/agent"

	sv "github.com/ligato/cn-infra/exec/supervisor"
	"github.com/ligato/cn-infra/logging"
)

func main() {
	// The supervisor plugin defines a configuration file allowing to easily manage processes using process
	// manager plugin.
	//
	// The config file can be put to the supervisor either via flag "-supervisor-config="
	// or define its path in the environment variable "SUPERVISOR_CONFIG". Another option is
	// to define config directly with "UseConf()" - this option will be shown in the example. A sample of
	// YAML config file can be found in the processes/supervisor folder.

	// Use this commented function to start supervisor where the config is provided via flag or env var
	//
	//a := agent.NewAgent(agent.AllPlugins(sv.NewPlugin()))
	//if err := a.Run(); err != nil {
	//	panic(err)
	//}

	log := logging.DefaultLogger

	conf := sv.Config{
		Programs: []sv.Program{
			{
				Name:           "p1",
				LogfilePath:    "example.log",
				ExecutablePath: "../process-manager-plugin/test-process/test-process",
				ExecutableArgs: []string{"-max-uptime=60"},
			},
			{
				Name:           "p2",
				LogfilePath:    "example.log",
				ExecutablePath: "../process-manager-plugin/test-process/test-process",
			},
		},
	}

	// start plugin
	bsp := sv.NewPlugin(sv.UseConf(conf))

	a := agent.NewAgent(agent.AllPlugins(bsp))
	if err := a.Start(); err != nil {
		panic(err)
	}
	defer func() {
		if err := a.Stop(); err != nil {
			panic(err)
		}
	}()

	// give the agent time to start
	time.Sleep(3 * time.Second)

	// test if all processes are running
	checkLiveness("p1", true, bsp, log)
	checkLiveness("p2", true, bsp, log)

	// terminate p1
	stopProcess("p1", bsp)

	// test if all states are as required
	checkLiveness("p1", false, bsp, log)
	checkLiveness("p2", true, bsp, log)
}

func checkLiveness(name string, isAlive bool, bsp sv.Supervisor, log logging.Logger) {
	p1 := bsp.GetProgramByName(name)
	if p1.IsAlive() == isAlive {
		if isAlive {
			log.Infof("%s is running", name)
			return
		}
		log.Infof("%s terminated", name)
		return
	}
	panic(fmt.Sprintf("process %s is in wrong state", name))
}

func stopProcess(name string, bsp sv.Supervisor) {
	p1 := bsp.GetProgramByName(name)
	if p1 == nil {
		panic(fmt.Sprintf("expected running process %s", name))
	}
	if _, err := p1.StopAndWait(); err != nil {
		panic(fmt.Sprintf("failed to stop process %s: %v", name, err))
	}

	// give the process time to stop
	time.Sleep(2 * time.Second)
}
