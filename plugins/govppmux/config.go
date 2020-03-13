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

package govppmux

import "time"

// Config defines configurable parameters for govppmux plugin.
type Config struct {
	// ReconnectResync enables resync after reconnect to VPP.
	ReconnectResync bool `json:"resync-after-reconnect"`

	// ReplyTimeout defines timeout period for replies in channels from VPP.
	ReplyTimeout time.Duration `json:"reply-timeout"`

	// Connect to VPP for configuration requests via the shared memory instead of through the socket.
	ConnectViaShm bool `json:"connect-via-shm"`

	// ShmPrefix defines prefix prepended to the name used for shared memory (SHM) segments.
	// If not set, shared memory segments are created directly in the SHM directory /dev/shm.
	ShmPrefix string `json:"shm-prefix"`

	// BinAPISocketPath defines path to the binapi socket file.
	BinAPISocketPath string `json:"binapi-socket-path"`

	// StatsSocketPath defines path to the stats socket file.
	StatsSocketPath string `json:"stats-socket-path"`

	// How many times can be request resent in case vpp is suddenly disconnected.
	RetryRequestCount int `json:"retry-request-count"`

	// Time between request resend attempts. Default is 500ms.
	RetryRequestTimeout time.Duration `json:"retry-request-timeout"`

	// How many times can be connection request resent in case the vpp is not reachable.
	RetryConnectCount int `json:"retry-connect-count"`

	// Time between connection request resend attempts. Default is 1s.
	RetryConnectTimeout time.Duration `json:"retry-connect-timeout"`

	// Enable VPP proxy.
	ProxyEnabled bool `json:"proxy-enabled"`

	// Below are options used for VPP connection health checking.
	HealthCheckProbeInterval time.Duration `json:"health-check-probe-interval"`
	HealthCheckReplyTimeout  time.Duration `json:"health-check-reply-timeout"`
	HealthCheckThreshold     int           `json:"health-check-threshold"`

	// DEPRECATED: TraceEnabled is obsolete and used only in older versions.
	TraceEnabled bool `json:"trace-enabled"`
}

func DefaultConfig() *Config {
	return &Config{
		ReconnectResync:          true,
		HealthCheckProbeInterval: time.Second,
		HealthCheckReplyTimeout:  250 * time.Millisecond,
		HealthCheckThreshold:     1,
		ReplyTimeout:             time.Second,
		RetryRequestTimeout:      500 * time.Millisecond,
		RetryConnectTimeout:      time.Second,
		ProxyEnabled:             true,
	}
}

func (p *Plugin) loadConfig() (*Config, error) {
	cfg := DefaultConfig()
	found, err := p.Cfg.LoadValue(cfg)
	if err != nil {
		return nil, err
	} else if found {
		p.Log.Debugf("config loaded from file %q", p.Cfg.GetConfigName())
	} else {
		p.Log.Debugf("config file %q not found, using default config", p.Cfg.GetConfigName())
	}
	return cfg, nil
}
