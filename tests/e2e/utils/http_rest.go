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

package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"go.ligato.io/cn-infra/v2/logging"
)

// HTTPClient provides client access to the HTTP server in agent.
type HTTPClient struct {
	addr string

	httpClient *http.Client

	Log logging.Logger
}

func NewHTTPClient(httpAddr string) *HTTPClient {
	httpClient := &http.Client{}
	return &HTTPClient{
		addr:       httpAddr,
		httpClient: httpClient,
	}
}

func (c *HTTPClient) debugf(f string, a ...interface{}) {
	if c.Log != nil {
		c.Log.Debugf(f, a...)
	}
}

func (c *HTTPClient) GET(path string) ([]byte, error) {
	return c.send(http.MethodGet, path, nil)
}

func (c *HTTPClient) PUT(path string, data interface{}) ([]byte, error) {
	return c.send(http.MethodPut, path, data)
}

func (c *HTTPClient) POST(path string, data interface{}) ([]byte, error) {
	return c.send(http.MethodPost, path, data)
}

func (c *HTTPClient) send(method, path string, data interface{}) ([]byte, error) {
	u, err := url.Parse("http://" + c.addr + path)
	if err != nil {
		return nil, err
	}

	var b []byte
	var r io.Reader

	if data != nil {
		b, err = json.MarshalIndent(data, "", "  ")
		if err != nil {
			return nil, err
		}
		r = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, u.String(), r)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	c.debugf("=> sending request: %s %s (%d bytes) request body:", req.Method, req.URL, req.ContentLength)
	c.debugf("request body:%s", b)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	c.debugf("<= received response: %v (%d)", resp.Status, resp.StatusCode)
	c.debugf("response body: %+v", resp)

	if resp.StatusCode > 400 {
		return nil, fmt.Errorf("response status: %s (%d)", resp.Status, resp.StatusCode)
	}

	msg, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return msg, nil
}
