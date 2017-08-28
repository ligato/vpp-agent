# Plugin Flavors 

Plugin Flavors:
1. A reusable combination of multiple plugins is called a 'ReusedFlavor'. 
   Consider the following code snippet. The ReusedFlavor structure is 
   basically a combination of Logrus, HTTP, LogManager, ServiceLabel, 
   StatusCheck, ETCD and Kafka plugins. All these plugins are implicitly
   instantiated. They do intentionally not contain pointers:
    1. to minimize the number of lines in the Flavor
    2. those plugins are not optional (if some of them would be, it would
       be a pointer)
    3. garbage collector ignores those field objects (since they are not 
       pointers - small optimization) 
2. The Inject() method contains hand written code (that is normally checked
   by the compiler rather than instantiated automatically by using reflection.
3. The Plugin() method returns a sorted list (slice) of plugins for agent 
   startup.
4. The CompositeFlavor example below demonstrates how to reuse some flavor
   (in this example RPCFlavor).

```go
package flavorexample

import (
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/flavors/rpc"
	"github.com/ligato/cn-infra/rpc/rest"
)

type CompositeFlavor struct {
	*rpc.FlavorRPC     // Reused Flavor
	PluginXY PluginXY // Added custom plugin to flavor
}

func (flavor *CompositeFlavor) Inject() bool {
	if !flavor.FlavorRPC.Inject() {
	    return false
	}

    flavor.PluginXY.HTTP = &flavor.FlavorRPC.HTTP
	// inject all other dependencies...
	
	return nil
}

func (flavor *CompositeFlavor) Plugins() []*core.NamedPlugin {
	flavor.Inject()
	
	if flavor.FlavorRPC == nil {
	    flavor.FlavorRPC = &rpc.FlavorRPC{}
	}
	
	return core.ListPluginsInFlavor(flavor)
}


type PluginXY struct {
    Dep // plugin dependencies
}

type Dep struct {
    HTTP rest.HTTPHandlers // injected, this plugin just depends on the API interface
}

func (plugin* PluginXY) Init() error {
    // use injected dependency
    plugin.HTTP.RegisterHTTPHandler(...)
    
    return nil
}

func (plugin* PluginXY) Close() error {
    // do something
    return nil
}
```
