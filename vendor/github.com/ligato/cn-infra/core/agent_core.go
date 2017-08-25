// Copyright (c) 2017 Cisco and/or its affiliates.
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

package core

import (
	"errors"
	"fmt"
	"time"

	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/namsral/flag"
)

// variables set by the Makefile using ldflags
var (
	BuildVersion string
	BuildDate    string
)

// Agent implements startup & shutdown procedure.
type Agent struct {
	// The startup/initialization must take no longer that maxStartup.
	MaxStartupTime time.Duration
	// plugin list
	plugins []*NamedPlugin
	logging.Logger
}

const (
	logErrorFmt       = "Plugin %s: init error '%s'"
	logSuccessFmt     = "Plugin %s: init success"
	logPostErrorFmt   = "Plugin %s: post-init error '%s'"
	logPostSuccessFmt = "Plugin %s: post-init success"
)

// NewAgent returns a new instance of the Agent with plugins.
func NewAgent(logger logging.Logger, maxStartup time.Duration, plugins ...*NamedPlugin) *Agent {
	a := Agent{
		maxStartup,
		plugins,
		logger,
	}
	return &a
}

// Start starts/initializes all plugins on the list.
// First it runs Init() method among all plugins in the list
// Then it tries to run AfterInit() method among all plugins t
// hat implements this optional method.
// It stops when first error occurs by calling Close() method
// for already initialized plugins in reverse order.
// The startup/initialization must take no longer that maxStartup.
// duration otherwise error occurs.
func (agent *Agent) Start() error {
	agent.WithFields(logging.Fields{"BuildVersion": BuildVersion, "BuildDate": BuildDate}).Info("Starting the agent...")

	doneChannel := make(chan struct{}, 0)
	errChannel := make(chan error, 0)

	if !flag.Parsed() {
		flag.Parse()
	}

	go func() {
		err := agent.initPlugins()
		if err != nil {
			errChannel <- err
			return
		}
		err = agent.handleAfterInit()
		if err != nil {
			errChannel <- err
			return
		}
		close(doneChannel)
	}()

	//block until all Plugins are initialized or timeout expires
	select {
	case err := <-errChannel:
		return err
	case <-doneChannel:
		agent.Info("All plugins initialized successfully")
		return nil
	case <-time.After(agent.MaxStartupTime):
		//TODO FIX - stop the initialization and close already initialized
		return fmt.Errorf("%s", "Some plugins not intialized before timeout")
	}
}

// Stop gracefully shuts down the Agent. It is called usually when the user
// interrupts the Agent from the EventLoopWithInterrupt().
//
// This implementation tries to call Close() method on every plugin on the list
// in revers order. It continues event if some error occurred.
func (agent *Agent) Stop() error {
	agent.Info("Stopping agent...")
	errMsg := ""
	for i := len(agent.plugins) - 1; i >= 0; i-- {
		agent.WithField("pluginName", agent.plugins[i].PluginName).Debug("Stopping plugin begin")
		err := safeclose.Close(agent.plugins[i].Plugin)
		if err != nil {
			if len(errMsg) > 0 {
				errMsg += "; "
			}
			errMsg += string(agent.plugins[i].PluginName)
			errMsg += ": " + err.Error()
		}
		agent.WithField("pluginName", agent.plugins[i].PluginName).Debug("Stopping plugin end ", err)
	}

	agent.Debug("Agent stopped")

	if len(errMsg) > 0 {
		return errors.New(errMsg)
	}
	return nil
}

// initPlugins calls Init() an all plugins on the list
func (agent *Agent) initPlugins() error {
	for i, plug := range agent.plugins {
		err := plug.Init()
		if err != nil {
			//Stop the plugins that are initialized
			for j := i; j >= 0; j-- {
				err := safeclose.Close(agent.plugins[j])
				if err != nil {
					agent.Warn("err closing ", agent.plugins[j].PluginName, " ", err)
				}
			}

			return fmt.Errorf(logErrorFmt, plug.PluginName, err)
		}
		agent.Info(fmt.Sprintf(logSuccessFmt, plug.PluginName))
	}
	return nil
}

// handleAfterInit calls the AfterInit handlers for plugins that can only
// finish their initialization after  all other plugins have been initialized.
func (agent *Agent) handleAfterInit() error {
	for _, plug := range agent.plugins {
		if plug2, ok := plug.Plugin.(PostInit); ok {
			agent.Debug("afterInit begin for ", plug.PluginName)
			err := plug2.AfterInit()
			if err != nil {
				agent.Stop()

				return fmt.Errorf(logPostErrorFmt, plug.PluginName, err)
			}
			agent.Info(fmt.Sprintf(logPostSuccessFmt, plug.PluginName))
		}
	}
	return nil
}
