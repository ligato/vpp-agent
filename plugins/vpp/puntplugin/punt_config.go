// Copyright (c) 2018 Cisco and/or its affiliates.
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

package puntplugin

import (
	govppapi "git.fd.io/govpp.git/api"
	"github.com/go-errors/errors"
	"github.com/ligato/cn-infra/logging"
	"github.com/ligato/cn-infra/utils/safeclose"
	"github.com/ligato/vpp-agent/idxvpp/nametoidx"
	"github.com/ligato/vpp-agent/plugins/govppmux"
	"github.com/ligato/vpp-agent/plugins/vpp/model/punt"
	"github.com/ligato/vpp-agent/plugins/vpp/puntplugin/puntidx"
	"github.com/ligato/vpp-agent/plugins/vpp/puntplugin/vppcalls"
)

// PuntConfigurator registers/de-registers punt to host via unix domain socket. Registered items are stored
// in the cache.
type PuntConfigurator struct {
	// logger
	log logging.Logger

	//  channel to communicate with VPP
	vppChan govppapi.Channel

	// vpp api handler
	puntHandler vppcalls.PuntVppAPI

	// cache
	mapping puntidx.PuntIndexRW
	idxSeq  uint32
}

// Init logger, VPP channel, VPP API handler and cache
func (c *PuntConfigurator) Init(logger logging.PluginLogger, goVppMux govppmux.API) (err error) {
	c.log = logger.NewLogger("punt-plugin")

	if c.vppChan, err = goVppMux.NewAPIChannel(); err != nil {
		return errors.Errorf("failed to create API channel: %v", err)
	}

	c.mapping = puntidx.NewPuntIndex(nametoidx.NewNameToIdx(c.log, "punt-indexes", nil))
	c.idxSeq = 1
	c.puntHandler = vppcalls.NewPuntVppHandler(c.vppChan, c.mapping, c.log)

	c.log.Info("Punt configurator initialized")

	return nil
}

// Close VPP channel
func (c *PuntConfigurator) Close() error {
	return safeclose.Close(c.vppChan)
}

// clearMapping prepares all in-memory-mappings and other cache fields. All previous cached entries are removed.
func (c *PuntConfigurator) clearMapping() {
	if c.mapping != nil {
		c.mapping.Clear()
	}
	c.log.Debugf("punt configurator mapping cleared")
}

// Add configures new punt to host via socket and stores it in the local mapping. Depending on L3 protocol
// setup, IPv4, IPv6 or both punts are registered
func (c *PuntConfigurator) Add(puntVal *punt.Punt) error {
	path, err := c.addPunt(puntVal)
	if err != nil {
		return err
	}

	c.mapping.RegisterName(puntVal.Name, c.idxSeq, &puntidx.PuntMetadata{Punt: puntVal, SocketPath: path})
	c.log.Debugf("Punt %s registered to local mapping", puntVal.Name)

	c.log.Infof("Punt %s configured", puntVal.Name)

	return nil
}

// Modify removes old entry, configures a new one and updates metadata
func (c *PuntConfigurator) Modify(oldPunt, newPunt *punt.Punt) error {
	// since punt cannot be modified via binary API, remove the odl value and add a new one
	if err := c.delPunt(oldPunt); err != nil {
		return errors.Errorf("punt modify (remove): %v", err)
	}
	path, err := c.addPunt(oldPunt)
	if err != nil {
		return errors.Errorf("punt modify (add): %v", err)
	}

	if !c.mapping.UpdateMetadata(newPunt.Name, &puntidx.PuntMetadata{Punt: newPunt, SocketPath: path}) {
		return errors.Errorf("failed to update metadata for %s", newPunt.Name)
	}
	c.log.Debugf("Punt %s metadata updated in local mapping", newPunt.Name)

	c.log.Infof("Punt %s modifier", newPunt.Name)

	return nil
}

// Delete the configuration of a punt
func (c *PuntConfigurator) Delete(puntVal *punt.Punt) error {
	if err := c.delPunt(puntVal); err != nil {
		return err
	}

	c.mapping.UnregisterName(puntVal.Name)
	c.log.Debugf("Punt %s unregistered from local mapping", puntVal.Name)

	c.log.Infof("Punt %s removed", puntVal.Name)

	return nil
}

func (c *PuntConfigurator) addPunt(puntVal *punt.Punt) (path []byte, err error) {
	if err = c.validate(puntVal); err != nil {
		return nil, err
	}

	// in L3Protocol_ALL case, both IPv4 and IPv6 punt is configured but the returned path is the same for both
	if puntVal.L3Protocol != punt.L3Protocol_IPv6 {
		path, err = c.puntHandler.RegisterPuntSocket(puntVal)
		if err != nil {
			return nil, errors.Errorf("failed to configure %s: %v", puntVal.Name, err)
		}
	}
	if puntVal.L3Protocol != punt.L3Protocol_IPv4 {
		path, err = c.puntHandler.RegisterPuntSocketIPv6(puntVal)
		if err != nil {
			return nil, errors.Errorf("failed to configure %s: %v", puntVal.Name, err)
		}
	}

	return path, nil
}

func (c *PuntConfigurator) delPunt(puntVal *punt.Punt) error {
	if err := c.validate(puntVal); err != nil {
		return err
	}

	if puntVal.L3Protocol != punt.L3Protocol_IPv6 {
		if err := c.puntHandler.DeregisterPuntSocket(puntVal); err != nil {
			return errors.Errorf("failed to remove %s: %v", puntVal.Name, err)
		}
	}
	if puntVal.L3Protocol != punt.L3Protocol_IPv4 {
		if err := c.puntHandler.DeregisterPuntSocketIPv6(puntVal); err != nil {
			return errors.Errorf("failed to remove %s: %v", puntVal.Name, err)
		}
	}

	return nil
}

func (c *PuntConfigurator) validate(puntVal *punt.Punt) error {
	if puntVal.SocketPath == "" {
		return errors.Errorf("failed to configure %s, socket path is not defined", puntVal.Name)
	}
	if puntVal.Port == 0 {
		return errors.Errorf("failed to configure %s, port is not defined", puntVal.Name)
	}
	return nil
}

// LogError prints error if not nil, including stack trace. The same value is also returned, so it can be easily propagated further
func (c *PuntConfigurator) LogError(err error) error {
	if err == nil {
		return nil
	}
	switch err.(type) {
	case *errors.Error:
		c.log.WithField("logger", c.log).Errorf(string(err.Error() + "\n" + string(err.(*errors.Error).Stack())))
	default:
		c.log.Error(err)
	}
	return err
}
