//  Copyright (c) 2019 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

/*
Package agent provides the life-cycle management agent for plugins. It is
intended to be used as a base point of an application used in main package.

Here is a common example usage:

	func main() {
		plugin := myplugin.NewPlugin()

		a := agent.NewAgent(
			agent.Plugins(plugin),
		)
		if err := a.Run(); err != nil {
			log.Fatal(err)
		}
	}

Options

There are various options available to customize agent:

	Version(ver, date, id)	- sets version of the program
	QuitOnClose(chan)   	- sets signal used to quit the running agent when closed
	QuitSignals(signals)	- sets signals used to quit the running agent (default: SIGINT, SIGTERM)
	StartTimeout(dur)   	- sets start timeout (default: 15s)
	StopTimeout(dur)    	- sets stop timeout (default: 5s)

There are two options for adding plugins to the agent:

	Plugins(...)	- adds just single plugins without lookup
	AllPlugins(...)	- adds plugin along with all of its plugin deps

*/
package agent
