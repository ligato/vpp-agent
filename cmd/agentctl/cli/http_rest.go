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

package cli

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/ligato/cn-infra/logging"
)

func (c *AgentCli) HttpRestGET(path string) ([]byte, error) {
	a, err := url.Parse("http://" + c.HttpAddr + path)
	if err != nil {
		return nil, err
	}

	logging.Debugf("HttpRestGET: %q", a.String())

	resp, err := http.Get(a.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	msg, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func (c *AgentCli) HttpRestPUT(path string, data interface{}) ([]byte, error) {
	a, err := url.Parse("http://" + c.HttpAddr + path)
	if err != nil {
		return nil, err
	}

	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPut, a.String(), bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	msg, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func (c *AgentCli) HttpRestPOST(path string, data interface{}) ([]byte, error) {
	a, err := url.Parse("http://" + c.HttpAddr + path)
	if err != nil {
		return nil, err
	}

	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(a.String(), "application/json", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	msg, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return msg, nil
}
