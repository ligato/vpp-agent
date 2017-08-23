package config

import (
	"flag"
)

// PluginConfig is API for plugins to access configuration.
//
// Aim of this API is to let a particular plugin to bind it's configuration
// without knowing a particular key name. The key name is injected in flavor (Plugin Name).
type PluginConfig interface {
	// GetValue parse configuration for a plugin a store the results in data.
	// Argument data is a pointer to instance of a go structure.
	GetValue(data interface{}) (found bool, err error)
}

// ForPlugin returns API that is injectable to a particular Plugin
// to read it's configuration.
//
// It tries to lookup `plugin + "-config"` in flags.
func ForPlugin(pluginName string) PluginConfig {
	plugCfg := pluginName + "-config"
	flg := flag.CommandLine.Lookup(plugCfg)
	if flg != nil {
		val := flg.Value.String()
		if val != "" {
			plugCfg = val
		}
	}

	return &pluginConfig{configFile: plugCfg}
}

type pluginConfig struct {
	configFile string
}

// GetValue binds the configuration to config method argument
func (p *pluginConfig) GetValue(config interface{}) (found bool, err error) {
	err = ParseConfigFromYamlFile(p.configFile, config) //TODO switch to Viper
	if err != nil {
		return false, err
	}

	return true, nil
}
