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

package ifplugin

import (
	"os"
	"time"
)

var (
	// PeriodicPollingPeriod between statistics reads
	// TODO  should be configurable
	PeriodicPollingPeriod = time.Second * 5

	// StateUpdateDelay defines delay before dumping states
	StateUpdateDelay = time.Second * 3

	disableInterfaceStats   = os.Getenv("DISABLE_INTERFACE_STATS") != ""
	disableStatusPublishing = os.Getenv("DISABLE_STATUS_PUBLISHING") != ""
)

// Config defines configuration for VPP ifplugin.
type Config struct {
	MTU              uint32   `json:"mtu"`
	StatusPublishers []string `json:"status-publishers"`
}

// DefaultConfig returns Config with default values.
func DefaultConfig() Config {
	return Config{
		MTU: 0,
	}
}

func (p *IfPlugin) loadConfig() (*Config, error) {
	cfg := DefaultConfig()

	found, err := p.Cfg.LoadValue(&cfg)
	if err != nil {
		return nil, err
	} else if !found {
		p.Log.Debugf("config %s not found", p.Cfg.GetConfigName())
		return nil, nil
	}
	p.Log.Debugf("config %s found: %+v", p.Cfg.GetConfigName(), cfg)

	// vppStatusPublishers can override state publishers from the configuration file.
	if pubs := os.Getenv("VPP_STATUS_PUBLISHERS"); pubs != "" {
		p.Log.Debugf("status publishers from env: %v", pubs)
		cfg.StatusPublishers = append(cfg.StatusPublishers, pubs)
	}

	return &cfg, err
}
