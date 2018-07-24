package config

import (
	"fmt"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/namsral/flag"

	"github.com/ligato/cn-infra/logging/logrus"
)

const (
	// FlagSuffix is added to plugin name while loading plugins configuration.
	FlagSuffix = "-config"

	// EnvSuffix is added to plugin name while loading plugins configuration from ENV variable.
	EnvSuffix = "_CONFIG"

	// DirFlag as flag name (see implementation in declareFlags())
	// is used to define default directory where config files reside.
	// This flag name is derived from the name of the plugin.
	DirFlag = "config-dir"

	// DirDefault holds a default value "." for flag, which represents current working directory.
	DirDefault = "."

	// DirUsage used as a flag (see implementation in declareFlags()).
	DirUsage = "Location of the config files; can also be set via 'CONFIG_DIR' env variable."
)

// PluginConfig is API for plugins to access configuration.
//
// Aim of this API is to let a particular plugin to bind it's configuration
// without knowing a particular key name. The key name is injected in flavor (Plugin Name).
type PluginConfig interface {
	// GetValue parses configuration for a plugin and stores the results in data.
	// The argument data is a pointer to an instance of a go structure.
	GetValue(data interface{}) (found bool, err error)

	// GetConfigName returns config name derived from plugin name:
	// flag = PluginName + FlagSuffix (evaluated most often as absolute path to a config file)
	GetConfigName() string
}

// FlagSet is a type alias for flag.FlagSet.
type FlagSet = flag.FlagSet

// pluginFlags is used for storing flags for Plugins before agent starts.
var pluginFlags = make(map[string]*FlagSet)

// RegisterFlagsFor registers defined flags for plugin with given name.
func RegisterFlagsFor(name string) {
	if plugSet, ok := pluginFlags[name]; ok {
		plugSet.VisitAll(func(f *flag.Flag) {
			flag.Var(f.Value, f.Name, f.Usage)
		})
	}
}

// ForPlugin returns API that is injectable to a particular Plugin
// and is used to read it's configuration.
//
// It tries to lookup `plugin + "-config"` in flags and declare
// the flag if it still not exists. It uses the following
// opts (used to define flag (if it was not already defined)):
// - default value
// - usage
func ForPlugin(name string, moreFlags ...func(*FlagSet)) PluginConfig {
	flagSet := flag.NewFlagSet(name, flag.ExitOnError)

	for _, more := range moreFlags {
		more(flagSet)
	}

	cfgFlag := name + FlagSuffix
	if flagSet.Lookup(cfgFlag) == nil {
		cfgFlagDefault := name + ".conf"
		cfgFlagUsage := fmt.Sprintf(
			"Location of the %q plugin config file; can also be set via %q env variable.",
			cfgFlagDefault, strings.ToUpper(name)+EnvSuffix)
		flagSet.String(cfgFlag, cfgFlagDefault, cfgFlagUsage)
	}

	pluginFlags[name] = flagSet

	return &pluginConfig{
		configFlag: cfgFlag,
	}
}

type pluginConfig struct {
	configFlag string
	access     sync.Mutex
	cfg        string
}

// Dir evaluates the flag DirFlag. It interprets "." as current working directory.
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

// GetValue binds the configuration to config method argument.
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

// GetConfigName looks up flag value and uses it to:
// 1. Find config in flag value location.
// 2. Alternatively, it tries to find it in config dir
// (see also Dir() comments).
func (p *pluginConfig) GetConfigName() string {
	p.access.Lock()
	defer p.access.Unlock()
	if p.cfg == "" {
		p.cfg = p.getConfigName()
	}

	return p.cfg
}

func (p *pluginConfig) getConfigName() string {
	flg := flag.CommandLine.Lookup(p.configFlag)
	if flg != nil {
		if flgVal := flg.Value.String(); flgVal != "" {
			// if exist value from flag
			if _, err := os.Stat(flgVal); !os.IsNotExist(err) {
				return flgVal
			}
			cfgDir, err := Dir()
			if err != nil {
				logrus.DefaultLogger().Error(err)
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
