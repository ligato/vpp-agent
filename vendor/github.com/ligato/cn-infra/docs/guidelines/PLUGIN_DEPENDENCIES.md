# Plugin Dependencies

1. Plugin structure dependencies are specified in the begining of structure definition
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
	
2. For plugins constructors are not needed. Because:
  * the dependencies are supposed to be exported fields (and injected).
  * Init() method is called on the plugin during agent startup [(see StartAgent in example main() function)(../../examples/simple-agent)

3. You can prefer [hand written code](../../examples/simple-agent/generic/generic.go) 
   that inject all dependencies between plugins or [automatic injection](https://godoc.org/github.com/facebookgo/inject)
   
4. Reusable combination of multiple plugins is called [Flavour](PLUGIN_FLAVOURS.md)