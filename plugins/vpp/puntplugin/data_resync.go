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
	"github.com/ligato/vpp-agent/plugins/vpp/model/punt"
)

// Resync configures punt entries.
func (c *PuntConfigurator) Resync(punts []*punt.Punt) error {
	// TODO since the dump API is not available, all punts are just configured. Is should cause no harm to the VPP
	// even if those entries already exists

	var lastErr error
	for _, puntVal := range punts {
		if err := c.Add(puntVal); err != nil {
			c.log.Errorf("RESYNC Punt %s error: %v", puntVal.Name, err)
			lastErr = err
		}
	}

	c.log.Debugf("RESYNC punt completed, configured %s items", len(punts))

	return lastErr
}
