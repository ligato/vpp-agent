package config

import (
	"github.com/namsral/flag"
)

// PluginConfig is API for plugins to access configuration.
//
// Aim of this API is to let a particular plugin to bind it's configuration
// without knowing a particular key name. The key name is injected in flavor (Plugin Name).
type PluginConfig interface {
	// GetValue parse configuration for a plugin a store the results in data.
	// Argument data is a pointer to instance of a go structure.
	GetValue(data interface{}) (found bool, err error)

	// GetConfigName returns usually derived config name from plugin name
	// PluginName + "-config"
	GetConfigName() string
}

// ForPlugin returns API that is injectable to a particular Plugin
// to read it's configuration.
//
// It tries to lookup `plugin + "-config"` in flags.
func ForPlugin(pluginName string) PluginConfig {
	return &pluginConfig{pluginName: pluginName}
}

type pluginConfig struct {
	pluginName string
}

// GetValue binds the configuration to config method argument
func (p *pluginConfig) GetValue(config interface{}) (found bool, err error) {
	err = ParseConfigFromYamlFile(p.GetConfigName(), config) //TODO switch to Viper
	if err != nil {
		return false, err
	}

	return true, nil
}

// GetConfigName - see description in PluginConfig.GetConfigName
func (p *pluginConfig) GetConfigName() string {
	plugCfg := p.pluginName + "-config"
	flg := flag.CommandLine.Lookup(plugCfg)
	if flg != nil {
		val := flg.Value.String()

		if val != "" {
			plugCfg = val
		}
	}

	return plugCfg
}
