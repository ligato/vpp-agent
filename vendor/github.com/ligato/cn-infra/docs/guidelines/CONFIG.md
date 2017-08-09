# Config

## Flags & Environment variables

1. Ligato source code uses the [flag package](https://github.com/namsral/flag)
   to define & parse command line flags and/or environment variables. 

2. The package level init() function defines one or more flags. If the 
   package is imported, then the flag is defined.

```go
    package xy

    import (
    "github.com/namsral/flag"
    )
    
    var defaultHTTPport string
    
    func init() {
        flag.StringVar(&defaultHTTPport, "httpPort", "9191", "Default port of the server")
    }  
```

## Config files

More complex configurations should be defined in one or more configuration 
files. Flags can be used to specify the name of the configuration file.

## Plugins

1. Plugin:
   1. loads its config in Init() method
   2. connects to a server or starts the server in AfterInit() method

```go
    package xy

    import (
    "github.com/namsral/flag"
    )
    
    type PluginXY struct {}
    
    func (plugin *PluginXY) Init() error {
        //load configuration
        return nil
    }  

    func (plugin *PluginXY) AfterInit() error {
        //use the configuration (connect somewhere etc.)
        return nil
    }  
```

2. Each plugin can have its own configuration (injected in [flavor](PLUGIN_FLAVORS.md))
   See following [Simple flag example](#Simple flag example) and [Complex configuration example](#Complex configuration example) 

### Simple flag example
```go
    package xy

    import (
    "github.com/namsral/flag"
    )
    
    var defaultHTTPport string
    
    type PluginXY struct {
        HTTPport string //can be injected
    }
    
    func (plugin *PluginXY) Init() error {
        //load configuration
        if plugin.HTTPport == "" {
           //apply global settings
           plugin.HTTPport = defaultHTTPport
        }
        
        return nil
    } 
```

### Complex configuration example
```go
    package xy

    import (
    "github.com/namsral/flag"
    )
    
    var defaultConfigName string
    
    type ConfigXY struct {
        HTTPport string
        //other fields...
    }
    
    type PluginXY struct {
        Config *ConfigXY //can be injected
        ConfigName string //can be injected
    }
    
    func (plugin *PluginXY) Init() error {
        //load configuration
        if plugin.Config == nil {
           //apply global settings
           if plugin.ConfigName == "" {
              plugin.ConfigName = defaultConfigName
           }
           //load config: ConfigBroker.GetValue(plugin.ConfigName, plugin.Config)
        }
        
        return nil
    } 
```
