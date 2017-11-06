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

package core

import (
	"testing"

	"github.com/ligato/cn-infra/logging"
	"github.com/onsi/gomega"
)

// Test01NoPluginsInFlavor checks that the flavor containing no fields
// that would implement Plugin interface (here we miss Close())
// returns empty slice.
func TestListPlugins01NoPluginsInFlavor(t *testing.T) {
	gomega.RegisterTestingT(t)

	flavor := FlavorNoPlugin{}
	plugs := flavor.Plugins()
	t.Log("plugs ", plugs)
	gomega.Expect(plugs).To(gomega.BeNil())
}

// Test02OnePluginInFlavor checks that the flavor containing multiple fields
// but only one of them implementing Plugin interface (other fields miss Close())
// returns slice with one particular plugin.
func TestListPlugins02OnePluginInFlavor(t *testing.T) {
	gomega.RegisterTestingT(t)

	flavor := &FlavorOnePlugin{}
	plugs := flavor.Plugins()
	t.Log("plugs ", plugs)
	gomega.Expect(plugs).To(gomega.Equal([]*NamedPlugin{&NamedPlugin{
		PluginName("Plugin2"), &flavor.Plugin2}}))
}

// FlavorNoPlugin contains no plugins.
type FlavorNoPlugin struct {
	Plugin1 MissignCloseMethod
	Plugin2 struct {
		Dep1B string
	}
}

// FlavorOnePlugin contains one plugin (another is missing Close method).
type FlavorOnePlugin struct {
	Plugin1 MissignCloseMethod
	Plugin2 DummyPlugin
}

// MissignCloseMethod implements only Init() but not Close() method.
type MissignCloseMethod struct {
}

// Init does nothing.
func (*MissignCloseMethod) Init() error {
	return nil
}

// DummyPlugin only defines Init() and Close() with empty method bodies.
type DummyPlugin struct {
	internalFlag bool
}

// Init does nothing.
func (*DummyPlugin) Init() error {
	return nil
}

// Close does nothing.
func (*DummyPlugin) Close() error {
	return nil
}

// DummyPlugin only defines Init() and Close() with empty method bodies.
type DummyPluginDep2 struct {
	internalFlag bool
}

// Init does nothing.
func (*DummyPluginDep2) Init() error {
	return nil
}

// Close does nothing.
func (*DummyPluginDep2) Close() error {
	return nil
}

// DummyPluginDep1 only defines Init() and Close() with empty method bodies.
type DummyPluginDep1 struct {
	internalFlag bool
}

// Init does nothing.
func (*DummyPluginDep1) Init() error {
	return nil
}

// Close does nothing.
func (*DummyPluginDep1) Close() error {
	return nil
}

// Inject does nothing.
func (f *FlavorNoPlugin) Inject() bool {
	return false
}

// LogRegistry is mock implementation - returns nil.
func (f *FlavorNoPlugin) LogRegistry() logging.Registry {
	return nil
}

// Plugins lists plugins in this flavor.
func (f *FlavorNoPlugin) Plugins() []*NamedPlugin {
	return ListPluginsInFlavor(f)
}

// Plugins lists plugins in this flavor.
func (f *FlavorOnePlugin) Plugins() []*NamedPlugin {
	return ListPluginsInFlavor(f)
}

// Inject does nothing.
func (f *FlavorOnePlugin) Inject() bool {
	return false
}

// LogRegistry is mock implementation - returns nil.
func (f *FlavorOnePlugin) LogRegistry() logging.Registry {
	return nil
}
