// Copyright (c) 2017 Cisco and/or its affiliates.
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

package l4plugin

import "github.com/ligato/vpp-agent/plugins/defaultplugins/l4plugin/model/l4"

// Resync configures app namespaces to the empty VPP
func (plugin *L4Configurator) Resync(appNamespaces []*l4.AppNamespaces_AppNamespace) error {
	return nil
}