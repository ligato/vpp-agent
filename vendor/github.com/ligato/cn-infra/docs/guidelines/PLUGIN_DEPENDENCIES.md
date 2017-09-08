# Plugin Dependencies

1. Plugin structure dependencies are specified in the beginning of structure definition
```go
	package xy
	import (
	    "github.com/ligato/cn-infra/flavors/local"
	    "github.com/ligato/cn-infra/datasync"
	)
	
	type PluginXY struct {
	    // dependencies
	    Dep

		//other fields (usually private fields) ...
	}
	
	type Dep struct {
	    local.PluginLogDeps //Plugin Logger & Plugin Name
	    
	    //other dependencies:
	    
	    Watcher datasync.KeyValProtoWatcher
	}
	
    func (plugin *PluginXY) Init() error {
        //using the dependency (following line is shortcut for plugin.Dep.PluginLogDeps.Log)
        plugin.Log.Info("using injected logger in flavor")
        
        return nil
    }  

```
	
2. For plugins, constructors are not needed. The reasons:
  * The dependencies are supposed to be the exported fields (and injected).
  * The Init() method is called on the plugin during agent startup; 
    see []StartAgent in the example main() function the 
    [simple agent example](../../examples/simple-agent)

3. Prefer [hand written code](../../flavors/rpc/rpc_flavor.go) 
   that injects all dependencies between plugins
   
4. Reusable combination of multiple plugins is called a [Flavor](PLUGIN_FLAVORS.md).
