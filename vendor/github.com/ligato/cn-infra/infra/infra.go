package infra

import (
	"github.com/ligato/cn-infra/config"
	"github.com/ligato/cn-infra/logging"
)

// Plugin interface defines plugin's basic life-cycle methods.
type Plugin interface {
	// Init is called in the agent`s startup phase.
	Init() error
	// Close is called in the agent`s cleanup phase.
	Close() error
	// String returns unique name of the plugin.
	String() string
}

// PostInit interface defines an optional method for plugins with additional initialization.
type PostInit interface {
	// AfterInit is called once Init() of all plugins have returned without error.
	AfterInit() error
}

// PluginName is a part of the plugin's API.
// It's used by embedding it into Plugin to
// provide unique name of the plugin.
type PluginName string

// String returns the PluginName.
func (name PluginName) String() string {
	return string(name)
}

// SetName sets plugin name.
func (name *PluginName) SetName(n string) {
	*name = PluginName(n)
}

// Deps defines common dependencies for use in Plugins.
// It can easily be embedded in Deps for Plugin:
// type Deps struct {
//     infra.Deps
//     // other dependencies
// }
type Deps struct {
	PluginName
	Log logging.PluginLogger
	config.PluginConfig
}
