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
	"path"

	"github.com/ligato/vpp-agent/api/types"
	"github.com/sirupsen/logrus"
)

// Ping pings the server and returns the value of the "API-Version" headers.
func (c *Client) Ping(ctx context.Context) (types.Ping, error) {
	var ping types.Ping
	logrus.Debugf("sending ping request")
	// Using cli.buildRequest() + cli.doRequest() instead of cli.sendRequest()
	// because ping requests are used during  API version negotiation, so we want
	// to hit the non-versioned /_ping endpoint, not /v1.xx/_ping
	req, err := c.buildRequest("GET", path.Join(c.basePath, "/ping"), nil, nil)
	if err != nil {
		return ping, err
	}
	serverResp, err := c.doRequest(ctx, req)
	defer ensureReaderClosed(serverResp)
	if err != nil {
		return ping, err
	}
	return parsePingResponse(c, serverResp)
}

func parsePingResponse(cli *Client, resp serverResponse) (types.Ping, error) {
	var ping types.Ping
	if resp.header == nil {
		err := cli.checkResponseErr(resp)
		return ping, err
	}
	ping.APIVersion = resp.header.Get("API-Version")
	err := cli.checkResponseErr(resp)
	return ping, err
}

// ServerVersion returns information of the client and agent host.
func (c *Client) ServerVersion(ctx context.Context) (types.Version, error) {
	resp, err := c.get(ctx, "/version", nil, nil)
	defer ensureReaderClosed(resp)
	if err != nil {
		return types.Version{}, err
	}

	var server types.Version
	err = json.NewDecoder(resp.body).Decode(&server)
	return server, err
}

func (c *Client) LoggerList(ctx context.Context, opts types.LoggerListOptions) ([]types.Logger, error) {
	resp, err := c.get(ctx, "/log/list", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("HTTP POST request failed: %v", err)
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
		return fmt.Errorf("HTTP POST request failed: %v", err)
	}

	type Response struct {
		Logger string `json:"logger,omitempty"`
		Level  string `json:"level,omitempty"`
		Error  string `json:"Error,omitempty"`
	}

	var loggerSetResponse Response
	if err := json.NewDecoder(resp.body).Decode(&resp); err != nil {
		return fmt.Errorf("decoding reply failed: %v", err)
	}
	if loggerSetResponse.Error != "" {
		return fmt.Errorf("SERVER: %s", loggerSetResponse.Error)
	}

	return nil
}
