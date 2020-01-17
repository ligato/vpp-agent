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

package main

import (
	"github.com/ligato/cn-infra/agent"
	"go.ligato.io/vpp-agent/v3/examples/tutorials/08_grpc/plugins/grpc"
)

func main() {
	// Prepare the App instance
	a := agent.NewAgent(
		agent.AllPlugins(New()),
	)

	// Start the App
	if err := a.Run(); err != nil {
		panic(err)
	}
}

// App represents application with the GRPC plugin
type App struct {
	GRPC *grpc.Client
}

// New starts new GRPC app
func New() *App {
	return &App{
		GRPC: &grpc.DefaultPlugin,
	}
}

// Init is empty, only to implement Plugin interface
func (App) Init() error {
	return nil
}

// Close is empty
func (App) Close() error {
	return nil
}

// String returns app plugin name
func (App) String() string {
	return "App-grpc-client"
}
