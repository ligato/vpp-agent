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

package govppmux

import (
	"context"
	"errors"
	"sync"

	"git.fd.io/govpp.git/adapter"
	"git.fd.io/govpp.git/adapter/vppapiclient"
	"git.fd.io/govpp.git/api"
	govpp "git.fd.io/govpp.git/core"
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/logging"
	log "github.com/ligato/cn-infra/logging/logrus"
	"github.com/ligato/cn-infra/health/statuscheck"
)

// PluginID used in the Agent Core flavors
const PluginID core.PluginName = "GOVPP"

// GOVPPPlugin implements the govppmux plugin interface.
type GOVPPPlugin struct {
	LogFactory  logging.LogFactory
	StatusCheck *statuscheck.Plugin

	logging.Logger

	vppConn    *govpp.Connection
	vppAdapter adapter.VppAdapter
	vppConChan chan govpp.ConnectionEvent

	cancel context.CancelFunc // cancel can be used to cancel all goroutines and their jobs inside of the plugin
	wg     sync.WaitGroup     // wait group that allows to wait until all goroutines of the plugin have finished
}

// FromExistingAdapter is used mainly for testing purposes.
func FromExistingAdapter(vppAdapter adapter.VppAdapter) *GOVPPPlugin {
	ret := &GOVPPPlugin{
		vppConn:    nil,
		vppAdapter: vppAdapter,
	}

	return ret
}

// Init is the entry point called by Agent Core. A single binary-API connection to VPP is established.
func (plugin *GOVPPPlugin) Init() error {
	var err error

	// register for providing status reports (push mode)
	plugin.StatusCheck.Register(PluginID, nil)

	govppLogger := plugin.LogFactory.NewLogger("GoVpp")

	govppLogger.SetLevel(logging.InfoLevel)
	if logger, ok := govppLogger.(*log.Logger); ok {
		govpp.SetLogger(logger.StandardLogger())
	}

	if plugin.vppAdapter == nil {
		plugin.vppAdapter = vppapiclient.NewVppAdapter()
	} else {
		govppLogger.Info("Reusing existing vppAdapter") //this is used for testing
	}

	plugin.vppConn, plugin.vppConChan, err = govpp.AsyncConnect(plugin.vppAdapter)
	if err != nil {
		return err
	}

	// TODO: async connect & automatic reconnect support is not yet implemented in the agent,
	// so synchronously wait until connected to VPP
	status := <-plugin.vppConChan
	if status.State != govpp.Connected {
		return errors.New("unable to connect to VPP")
	}

	plugin.StatusCheck.ReportStateChange(PluginID, statuscheck.OK, nil)
	govppLogger.Debug("govpp connect success ", plugin.vppConn)

	var ctx context.Context
	ctx, plugin.cancel = context.WithCancel(context.Background())
	go plugin.handleVPPConnectionEvents(ctx)

	return nil
}

// Close cleans up the resources allocated by the govppmux plugin.
func (plugin *GOVPPPlugin) Close() error {
	plugin.cancel()
	plugin.wg.Wait()

	defer func() {
		if plugin.vppConn != nil {
			plugin.vppConn.Disconnect()
		}
	}()

	return nil
}

// NewAPIChannel returns a new API channel for communication with VPP via govpp core.
// It uses default buffer sizes for the request and reply Go channels.
//
// Example of binary API call from some plugin using GOVPP:
//      ch, _ := govpp_mux.NewAPIChannel()
//      ch.SendRequest(req).ReceiveReply
func (plugin *GOVPPPlugin) NewAPIChannel() (*api.Channel, error) {
	return plugin.vppConn.NewAPIChannel()
}

// NewAPIChannelBuffered returns a new API channel for communication with VPP via govpp core.
// It allows to specify custom buffer sizes for the request and reply Go channels.
//
// Example of binary API call from some plugin using GOVPP:
//      ch, _ := govpp_mux.NewAPIChannelBuffered(100, 100)
//      ch.SendRequest(req).ReceiveReply
func (plugin *GOVPPPlugin) NewAPIChannelBuffered(reqChanBufSize, replyChanBufSize int) (*api.Channel, error) {
	return plugin.vppConn.NewAPIChannelBuffered(reqChanBufSize, replyChanBufSize)
}

// handleVPPConnectionEvents handles VPP connection events
func (plugin *GOVPPPlugin) handleVPPConnectionEvents(ctx context.Context) {
	plugin.wg.Add(1)
	defer plugin.wg.Done()

	// TODO: support for VPP reconnect

	for {
		select {
		case status := <-plugin.vppConChan:
			if status.State == govpp.Connected {
				plugin.StatusCheck.ReportStateChange(PluginID, statuscheck.OK, nil)
			} else {
				plugin.StatusCheck.ReportStateChange(PluginID, statuscheck.Error, errors.New("VPP disconnected"))
			}

		case <-ctx.Done():
			return
		}
	}
}
