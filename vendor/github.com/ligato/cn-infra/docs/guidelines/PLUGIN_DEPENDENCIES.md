# Plugin Dependencies

1. Plugin structure dependencies are specified in the beginning of structure definition
```go
	package xy
	import (
	    "github.com/ligato/cn-infra/logging"
	)
	
	type Plugin struct {
		LogFactory     logging.LogFactory `inject:`
		//other dependencies ...
	}
```
	
2. For plugins, constructors are not needed. The reasons:
  * The dependencies are supposed to be the exported fields (and injected).
  * The Init() method is called on the plugin during agent startup; 
    see []StartAgent in the example main() function the 
    [simple agent example](../../examples/simple-agent)

3. You can prefer [hand written code](../../examples/simple-agent/generic/generic.go) 
   that injects all dependencies between plugins or [automatic injection](https://godoc.org/github.com/facebookgo/inject)
   
4. Reusable combination of multiple plugins is called a [Flavor](PLUGIN_FLAVORS.md).
