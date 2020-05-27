// Copyright (c) 2019 Cisco and/or its affiliates.
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

//go:generate descriptor-adapter --descriptor-name RuleChain --value-type *linux_iptables.RuleChain --import "go.ligato.io/vpp-agent/v3/proto/ligato/linux/iptables" --output-dir "descriptor"

package iptablesplugin

import (
	"math"

	"go.ligato.io/cn-infra/v2/infra"

	kvs "go.ligato.io/vpp-agent/v3/plugins/kvscheduler/api"
	"go.ligato.io/vpp-agent/v3/plugins/linux/iptablesplugin/descriptor"
	"go.ligato.io/vpp-agent/v3/plugins/linux/iptablesplugin/linuxcalls"
	"go.ligato.io/vpp-agent/v3/plugins/linux/nsplugin"
)

const (
	// by default, at most 10 go routines will split the configured rule chains
	// to execute the Retrieve operation in parallel.
	defaultGoRoutinesCnt = 10

	// by default, no rules will be added by alternative performance strategy using
	// iptables-save/modify data/iptables-store technique
	// If this performance technique is needed, then the minimum rule limit should be lowered
	// by configuration to some lower value (0 means that the permance strategy is
	// always used)
	defaultMinRuleCountForPerfRuleAddition = math.MaxInt32
)

// Config holds the plugin configuration.
type Config struct {
	linuxcalls.HandlerConfig `json:"handler"`

	Disabled      bool `json:"disabled"`
	GoRoutinesCnt int  `json:"go-routines-count"`
}

func DefaultConfig() *Config {
	return &Config{
		// default configuration
		GoRoutinesCnt: defaultGoRoutinesCnt,
		HandlerConfig: linuxcalls.HandlerConfig{
			MinRuleCountForPerfRuleAddition: defaultMinRuleCountForPerfRuleAddition,
		},
	}
}

// IPTablesPlugin configures Linux iptables rules.
type IPTablesPlugin struct {
	Deps

	conf *Config

	// system handlers
	iptHandler linuxcalls.IPTablesAPI

	// descriptors
	ruleChainDescriptor *descriptor.RuleChainDescriptor
}

// Deps lists dependencies of the plugin.
type Deps struct {
	infra.PluginDeps
	KVScheduler kvs.KVScheduler
	NsPlugin    nsplugin.API
}

// Init initializes and registers descriptors and handlers for Linux iptables rules.
func (p *IPTablesPlugin) Init() (err error) {
	// parse configuration file
	p.conf, err = p.loadConfig()
	if err != nil {
		return err
	}
	p.Log.Debugf("Linux iptables config: %+v", p.conf)
	if p.conf.Disabled {
		p.Log.Infof("Disabling iptables plugin")
		return nil
	}

	// init iptables handler
	p.iptHandler = linuxcalls.NewIPTablesHandler()
	err = p.iptHandler.Init(&p.conf.HandlerConfig)
	if err != nil {
		// just warn here, iptables / ip6tables just may not be installed - will return
		// an error by attempt to configure it
		p.Log.Warnf("Error by initializing iptables handler: %v", err)
	}

	// init & register the descriptor
	ruleChainDescriptor := descriptor.NewRuleChainDescriptor(
		p.KVScheduler, p.iptHandler, p.NsPlugin, p.Log, p.conf.GoRoutinesCnt, p.conf.MinRuleCountForPerfRuleAddition)

	err = p.Deps.KVScheduler.RegisterKVDescriptor(ruleChainDescriptor)
	if err != nil {
		return err
	}

	return nil
}

// Close does nothing here.
func (p *IPTablesPlugin) Close() error {
	return nil
}

// loadConfig loads plugin configuration file.
func (p *IPTablesPlugin) loadConfig() (*Config, error) {
	config := DefaultConfig()
	if p.Cfg != nil {
		found, err := p.Cfg.LoadValue(config)
		if !found {
			p.Log.Debug("Linux IPTablesPlugin config not found")
			return config, nil
		}
		if err != nil {
			return nil, err
		}
	}
	return config, nil
}
