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

package client

import (
	"context"
	"encoding/json"
	"fmt"

	"go.fd.io/govpp/api"
)

func (c *Client) VppRunCli(ctx context.Context, cmd string) (reply string, err error) {
	data := map[string]interface{}{
		"vppclicommand": cmd,
	}
	resp, err := c.post(ctx, "/vpp/command", nil, data, nil)
	if err != nil {
		return "", fmt.Errorf("HTTP POST request failed: %v", err)
	}
	if err := json.NewDecoder(resp.body).Decode(&reply); err != nil {
		return "", fmt.Errorf("decoding reply failed: %v", err)
	}
	return reply, nil
}

func (c *Client) VppGetStats(ctx context.Context, typ string) error {
	// TODO: implement more generic stats provider that goes beyond GoVPP StatsProvider (git.fd.io/govpp/api/stats.go)
	//  and can dump any possible stats or all of them (just like in stats dump example in
	//  git.fd.io/govpp/examples/stats-client/stats_api.go)
	return nil
}

func (c *Client) VppGetInterfaceStats() (*api.InterfaceStats, error) {
	statsProvider, err := c.vppStatsProvider()
	if err != nil {
		return nil, fmt.Errorf("can't get vpp stats provider for interface stats retrieval due to: %v", err)
	}
	var stats api.InterfaceStats
	if err := statsProvider.GetInterfaceStats(&stats); err != nil {
		return nil, fmt.Errorf("getting interface stats failed due to: %v", err)
	}
	return &stats, nil
}

func (c *Client) VppGetErrorStats() (*api.ErrorStats, error) {
	statsProvider, err := c.vppStatsProvider()
	if err != nil {
		return nil, fmt.Errorf("can't get vpp stats provider for error stats retrieval due to: %v", err)
	}
	var stats api.ErrorStats
	if err := statsProvider.GetErrorStats(&stats); err != nil {
		return nil, fmt.Errorf("getting error stats failed due to: %v", err)
	}
	return &stats, nil
}

func (c *Client) VppGetSystemStats() (*api.SystemStats, error) {
	statsProvider, err := c.vppStatsProvider()
	if err != nil {
		return nil, fmt.Errorf("can't get vpp stats provider for system stats retrieval due to: %v", err)
	}
	var stats api.SystemStats
	if err := statsProvider.GetSystemStats(&stats); err != nil {
		return nil, fmt.Errorf("getting system stats failed due to: %v", err)
	}
	return &stats, nil
}

func (c *Client) VppGetNodeStats() (*api.NodeStats, error) {
	statsProvider, err := c.vppStatsProvider()
	if err != nil {
		return nil, fmt.Errorf("can't get vpp stats provider for node stats retrieval due to: %v", err)
	}
	var stats api.NodeStats
	if err := statsProvider.GetNodeStats(&stats); err != nil {
		return nil, fmt.Errorf("getting node stats failed due to: %v", err)
	}
	return &stats, nil
}

func (c *Client) VppGetBufferStats() (*api.BufferStats, error) {
	statsProvider, err := c.vppStatsProvider()
	if err != nil {
		return nil, fmt.Errorf("can't get vpp stats provider for buffer stats retrieval due to: %v", err)
	}
	var stats api.BufferStats
	if err := statsProvider.GetBufferStats(&stats); err != nil {
		return nil, fmt.Errorf("getting buffer stats failed due to: %v", err)
	}
	return &stats, nil
}

func (c *Client) vppStatsProvider() (api.StatsProvider, error) {
	proxyClient, err := c.GoVPPProxyClient()
	if err != nil {
		return nil, fmt.Errorf("can't get GoVPP proxy client due to: %v", err)
	}
	statsProvider, err := proxyClient.NewStatsClient()
	if err != nil {
		return nil, fmt.Errorf("can't get GoVPP's proxy stats client due to: %v", err)
	}
	return statsProvider, nil
}
