package config

import (
	"sync"

	"os"

	"path"

	"strings"

	"github.com/ligato/cn-infra/logging/logroot"
	"github.com/namsral/flag"
)

// FlagSuffix is added to plugin name while loading plugins configuration
const FlagSuffix = "-config"

// DirFlag used as flag name (see implementation in declareFlags())
// It is used to define default directory where config files reside.
// This flag name is calculated from the name of the plugin.
const DirFlag = "config-dir"

// DirDefault - default value for flag "." represents current working directory
const DirDefault = "."

// DirUsage used as flag usage (see implementation in declareFlags())
const DirUsage = "Location of the configuration files; also set via 'CONFIG_DIR' env variable."

// PluginConfig is API for plugins to access configuration.
//
// Aim of this API is to let a particular plugin to bind it's configuration
// without knowing a particular key name. The key name is injected in flavor (Plugin Name).
type PluginConfig interface {
	// GetValue parse configuration for a plugin a store the results in data.
	// Argument data is a pointer to instance of a go structure.
	GetValue(data interface{}) (found bool, err error)

	// GetConfigName returns usually derived config name from plugin name:
	// flag = PluginName + FlagSuffix (evaluated most often a absolute path to a config file)
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
	access     sync.Mutex
	cfg        string
}

// GetValue binds the configuration to config method argument
func (p *pluginConfig) GetValue(config interface{}) (found bool, err error) {
	cfgName := p.GetConfigName()
	if cfgName == "" {
		return false, nil
	}
	err = ParseConfigFromYamlFile(cfgName, config) //TODO switch to Viper (possible to have one huge config file)
	if err != nil {
		return false, err
	}

	return true, nil
}

// GetConfigName lookups flag value and uses it to:
// 1. find config in flag value location
// 2. alternatively it tries to find it in config dir
// (see also Dir() comments)
func (p *pluginConfig) GetConfigName() string {
	p.access.Lock()
	defer p.access.Unlock()
	if p.cfg == "" {
		p.cfg = p.getConfigName()
	}

	return p.cfg
}

func (p *pluginConfig) getConfigName() string {
	flgName := p.pluginName + FlagSuffix
	flg := flag.CommandLine.Lookup(flgName)
	if flg != nil {
		flgVal := flg.Value.String()

		if flgVal != "" {
			// if exist value from flag
			if _, err := os.Stat(flgVal); !os.IsNotExist(err) {
				return flgVal
			}
			cfgDir, err := Dir()
			if err != nil {
				logroot.StandardLogger().Error(err)
				return ""
			}
			// if exist flag value in config dir
			flgValInConfigDir := path.Join(cfgDir, flgVal)
			if _, err := os.Stat(flgValInConfigDir); !os.IsNotExist(err) {
				return flgValInConfigDir
			}
		}
	}

	return ""
}

// Dir evaluates flag DirFlag. It interprets "." as current working directory.
func Dir() (string, error) {
	flg := flag.CommandLine.Lookup(DirFlag)
	if flg != nil {
		val := flg.Value.String()
		if strings.HasPrefix(val, ".") {
			cwd, err := os.Getwd()
			if err != nil {
				return cwd, err
			}

			if len(val) > 1 {
				return cwd + val[1:], nil
			}
			return cwd, nil
		}

		return val, nil
	}

	return "", nil
}
