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
	// TODO: implement this
	return nil
}
