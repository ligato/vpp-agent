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

	"go.ligato.io/cn-infra/v2/health/probe"

	"go.ligato.io/vpp-agent/v3/cmd/agentctl/api/types"
)

// AgentVersion returns information about Agent.
func (c *Client) AgentVersion(ctx context.Context) (*types.Version, error) {
	resp, err := c.get(ctx, "/info/version", nil, nil)
	defer ensureReaderClosed(resp)
	if err != nil {
		return nil, err
	}
	var v types.Version
	v.APIVersion = resp.header.Get("API-Version")

	err = json.NewDecoder(resp.body).Decode(&v)
	return &v, err
}

// LoggerList returns list of all registered loggers in Agent.
func (c *Client) LoggerList(ctx context.Context) ([]types.Logger, error) {
	resp, err := c.get(ctx, "/log/list", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %v", err)
	}

	var loggers []types.Logger
	if err := json.NewDecoder(resp.body).Decode(&loggers); err != nil {
		return nil, fmt.Errorf("decoding reply failed: %v", err)
	}

	return loggers, nil
}

func (c *Client) LoggerSet(ctx context.Context, logger, level string) error {
	urlPath := "/log/" + logger + "/" + level

	resp, err := c.put(ctx, urlPath, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %v", err)
	}

	type Response struct {
		Logger string `json:"logger,omitempty"`
		Level  string `json:"level,omitempty"`
		Error  string `json:"Error,omitempty"`
	}

	var loggerSetResponse Response
	if err := json.NewDecoder(resp.body).Decode(&loggerSetResponse); err != nil {
		return fmt.Errorf("decoding reply failed: %v", err)
	}
	if loggerSetResponse.Error != "" {
		return fmt.Errorf("SERVER: %s", loggerSetResponse.Error)
	}

	return nil
}

func (c *Client) Status(ctx context.Context) (*probe.ExposedStatus, error) {
	resp, err := c.get(ctx, "/readiness", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %v", err)
	}

	var status probe.ExposedStatus
	if err := json.NewDecoder(resp.body).Decode(&status); err != nil {
		return nil, fmt.Errorf("decoding reply failed: %v", err)
	}

	return &status, nil
}

func (c *Client) GetMetricData(ctx context.Context, metricName string) (map[string]interface{}, error) {
	resp, err := c.get(ctx, "/metrics/"+metricName, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %v", err)
	}

	var metricData = make(map[string]interface{})
	if err := json.NewDecoder(resp.body).Decode(&metricData); err != nil {
		return nil, fmt.Errorf("decoding reply failed: %v", err)
	}

	return metricData, nil
}
