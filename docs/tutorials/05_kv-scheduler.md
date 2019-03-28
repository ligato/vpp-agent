# Tutorial: Using the KV Scheduler in your plugin

This tutorial shows how to use the ['KV Scheduler'][5] in our hello world plugin that was created in previous
tutorials. You will learn how to prepare a descriptor, generate the adapter and wire the plugin with the KV 
Scheduler. 

Requirements:
* Complete and understand the ['Hello World Agent'](https://ligato.io/cn-infra/tutorials/01_hello-world) tutorial
* Complete and understand the ['Plugin Dependencies'](https://ligato.io/cn-infra/tutorials/02_plugin-deps) tutorial

For simplicity, this tutorial does not use ETCD or any other northbound kv store. Instead, NB events are created 
programmatically in the example, using the KV Scheduler API.

The vpp-agent uses VPP binary API calls to configure the VPP. Each VPP binary API call is designed to create
a configuration item in the VPP or to add or modify one or more configuration parameters. In practice, these
actions can be dependent on each other. For example, an IP address can be assigned to an interface only if 
the interface is already present in the VPP. Another example is an L2 FIB entry, which can be added only if
the required interface and a bridge domain exist and the interface is also assigned to the bridge domain, 
creating a complex dependency tree. In general, it is true that:

1. Typically, more than one binary API call is required to configure a proto-modelled data item coming from the
   northbound 
2. To configure a configuration parameter, its parent must exist
2. Some configuration items are dependent on other configuration items, and they cannot be configured before their
   dependencies are met

This means that VPP binary API calls must be called in a certain order. The VPP agent uses the KV Scheduler to ensure
this order, managing configuration dependencies and caching configuration items until their dependencies are met.
Any plugin that configures something that is dependent on some other plugin's configutation items can be registered
with the KV scheduler and profit from this functionality. 
 
First, we define a simple northbound [proto model][1] that we will use in our example plugin. The model defines two
simple messages - an  `Interface` and a `Route` that depends on some interface. The model demonstrates a simple 
dependency between configuration  items (basically, we need an interface to configure a route).  
 
**Important note:** The vpp-agent uses the Orchestrator component, which is responsible for collecting northbound
data from multiple sources (mainly a KV Store and GRPC clients). To marshall/unmarshall proto messages defined in
northbound proto models, the Orchestrator needs message names to be present in the messages. To generate code where
message names are present in proto messages we must use the following special protobuf option (together with its
import):
```proto
import "github.com/gogo/protobuf/gogoproto/gogo.proto";
option (gogoproto.messagename_all) = true;
```

In order to register our Hello World plugin with the scheduler and to work with our new model, we need two new 
components - a **descriptor** and an **adapter** for every proto-defined type (proto message).
 
#### 1. Adapters
 
Let's start with adapters. The purpose of an adapter is to define conversion methods between uour proto-defined type and
a bare `proto.Message` that the KV Scheduler works with. Since this is boilerplate code, the is tooling to auto-generate
it. The code generator is called `descriptor-adapter` and it can be found [inside the KVScheduler plugin][2]. Build the 
binary file from the go files inside, and use it to generate the adapters for the `Interface` and `Route` proto messages:
 
```
descriptor-adapter --descriptor-name Interface --value-type *model.Interface --import "github.com/ligato/vpp-agent/examples/tutorials/05_kv-scheduler/model" --output-dir "descriptor"
descriptor-adapter --descriptor-name Route --value-type *model.Route --import "github.com/ligato/vpp-agent/examples/tutorials/05_kv-scheduler/model" --output-dir "descriptor"
```

It is good practice to add the above commands to the plugin's main .go file with the `//go:generate` directives. The 
`descriptor-adapter` generator will put the generated adapters into the `descriptor/adapter` directory within the
plugin folder.

#### 2. Descriptor without dependency

The next  step is to define descriptors. We start with the interface descriptor, which has no dependencies. 

A descriptor can be implemented in one of two ways:
1. Define the descriptor constructor which implements all required methods directly (good when descriptor methods
   are few and short in implementation)
2. Define a descriptor object, implement all methods on it and then put a method references in the descriptor
   constructor (the preferred way)

In the interface descriptor, we use the first approach. Let's create a new file - `descriptors.go` - so that the
descriptor code is outside of `main.go`. Next, add the following code:

```go
func NewIfDescriptor(logger logging.PluginLogger) *api.KVDescriptor {
	typedDescriptor := &adapter.InterfaceDescriptor{
		// descriptor implementation
	}
	return adapter.NewInterfaceDescriptor(typedDescriptor)
}
```

**Note:** descriptors in this example are all in a single file since they are short, but the preferred way is to put
each descriptor in its own `.go` file. 

`NewIfDescriptor` is a constructor function that returns a type-safe descriptor object. All potential descriptor 
dependencies (logger, various mappings, etc.) are provided via constructor parameters.  

If you have a look at `adapter.InterfaceDescriptor`, you will see that it defines several fields. The most important
fields are function-types with CRUD definitions and fields resolving dependencies. The full API list is documented in the [KvDescriptor structure][3]. Here, we implement the the APIs that we need for our simple example:

* Name of the descriptor, must be unique for all descriptors. 
```go
    Name: "if-descriptor",
```

* Northbound key prefix for configuration type handled by the descriptor
```go
NBKeyPrefix: "/interface/",
```

* String representation of the type
```go
ValueTypeName: proto.MessageName(&model.Interface{}),
```

* Configuration item identifier (label, name, index) is returned by this method. 
```go
KeyLabel: func(key string) string {
    return strings.TrimPrefix(key, "/interface/")
},
```

* Key selector returns true if the provided key is described by the given descriptor. A descriptor can support a 
  subset of keys, but it can only process one value type.
```go
KeySelector: func(key string) bool {
    if strings.HasPrefix(key, ifPrefix) {
        return true
    }
    return false
},
```

* This flag enables metadata for the given type
```go
WithMetadata: true,
```

* Create method configures a new configuration item (interface).
```go
Create: func(key string, value *model.Interface) (metadata interface{}, err error) {
    d.log.Infof("Interface %s created", value.Name)
    return value.Name, nil
},
```

This is how the completed interface descriptor looks like:
```go
func NewIfDescriptor(logger logging.PluginLogger) *api.KVDescriptor {
	typedDescriptor := &adapter.InterfaceDescriptor{
		Name: ifDescriptorName,
		NBKeyPrefix: ifPrefix,
		ValueTypeName: proto.MessageName(&model.Interface{}),
		KeyLabel: func(key string) string {
			return strings.TrimPrefix(key, ifPrefix)
		},
		KeySelector: func(key string) bool {
			if strings.HasPrefix(key, ifPrefix) {
				return true
			}
			return false
		},
		WithMetadata: true,
		Create: func(key string, value *model.Interface) (metadata interface{}, err error) {
			logger.Infof("Interface %s created", value.Name)
			return value.Name, nil
		},
	}
	return adapter.NewInterfaceDescriptor(typedDescriptor)
}
```

#### 3. Descriptor with dependency

Next, we continue with the route descriptor, which has a dependency on an interface. This descriptor defines additional
fields, since we need to define the dependency on the interface configuration item. Also here we specify descriptor
struct and implement methods outside of the descriptor constructor. Define the struct and constructor first:

```go
type RouteDescriptor struct {
	// dependencies
	log logging.PluginLogger
}

func NewRouteDescriptor(logger logging.PluginLogger) *api.KVDescriptor {
	typedDescriptor := &adapter.RouteDescriptor{
		// descriptor implementation
	}
	return adapter.NewRouteDescriptor(typedDescriptor)
}
```

The route descriptor fields, `NBKeyPrefix`, `KeyLabel`and `KeySelector` are implemented in the same manner as for the interface type, but outside of the constructor as methods with `RouteDescriptor` as pointer receiver (since they are `func` type):
```go
func (d *RouteDescriptor) KeyLabel(key string) string {
	return strings.TrimPrefix(key, routePrefix)
}

func (d *RouteDescriptor) KeySelector(key string) bool {
	if strings.HasPrefix(key, routePrefix) {
		return true
	}
	return false
}

func (d *RouteDescriptor) Dependencies(key string, value *model.Route) []api.Dependency {
	return []api.Dependency{
		{
			Label: routeInterfaceDepLabel,
			Key:   ifPrefix + value.InterfaceName,
		},
	}
}
```

The field `WithMetadata` is not needed here, so the `Create` method does not return any metadata value:

```go
func (d *RouteDescriptor) Create(key string, value *model.Route) (metadata interface{}, err error) {
	d.log.Infof("Created route %s dependent on interface %s", value.Name, value.InterfaceName)
	return nil, nil
}
``` 

In addition, there are two new fields:

* A list of dependencies - a key prefix and a unique label value are required for any given given configuration item.
The item will not be created while the dependent key does not exist. The label is informative and should be unique.
```go
func (d *RouteDescriptor) Dependencies(key string, value *model.Route) []api.Dependency {
	return []api.Dependency{
		{
			Label: routeInterfaceDepLabel,
			Key:   ifPrefix + value.InterfaceName,
		},
	}
}
```

* A list of descriptors where the dependent values are processed. In the example, we return the interface descriptor
  since that is the one handling interfaces.
```go
RetrieveDependencies: []string{ifDescriptorName},
```

Now define descriptor context of type `RouteDescriptor` within `NewRouteDescriptor`:
```go
func NewRouteDescriptor(logger logging.PluginLogger) *api.KVDescriptor {
	descriptorCtx := &RouteDescriptor{
		log: logger,
	}
	typedDescriptor := &adapter.RouteDescriptor{
		// descriptor implementation
	}
	return adapter.NewRouteDescriptor(typedDescriptor)
}
```

Set non-function fields:
```go
func NewRouteDescriptor(logger logging.PluginLogger) *api.KVDescriptor {
	descriptorCtx := &RouteDescriptor{
		log: logger,
	}
	typedDescriptor := &adapter.RouteDescriptor{
        Name: routeDescriptorName,
        NBKeyPrefix: routePrefix,
        ValueTypeName: proto.MessageName(&model.Route{}),      
        RetrieveDependencies: []string{ifDescriptorName},
	}
	return adapter.NewRouteDescriptor(typedDescriptor)
}
```

Set function fields as references to the `RouteDescriptor` methods. The complete descriptor:
```go
func NewRouteDescriptor(logger logging.PluginLogger) *api.KVDescriptor {
	descriptorCtx := &RouteDescriptor{
		log: logger,
	}
	typedDescriptor := &adapter.RouteDescriptor{
		Name: routeDescriptorName,
		NBKeyPrefix: routePrefix,
		ValueTypeName: proto.MessageName(&model.Route{}),
		KeyLabel: descriptorCtx.KeyLabel,
		KeySelector: descriptorCtx.KeySelector,
		Dependencies: descriptorCtx.Dependencies,
		Create: descriptorCtx.Create,
	}
	return adapter.NewRouteDescriptor(typedDescriptor)
}
```

The descriptor API provides more methods not used in the example in order to keep it simple (like Update(), Delete(), Retrieve(), Validate(), ...). The full list can be found in the [descriptor API documentation][3]

#### Wire our plugin with the KV scheduler

Now when the descriptors are completed, we will register them in `main.go`. First step is to add the `KVScheduler` to the
`HelloWorld` plugin as a plugin dependency:
```go
type HelloWorld struct {
	infra.PluginDeps
	KVScheduler api.KVScheduler
}
```

Now register descriptors to the KVScheduler in the hello world plugin `Init()`:
```go
func (p *HelloWorld) Init() error {
	p.Log.Println("Hello World!")

	err := p.KVScheduler.RegisterKVDescriptor(adapter.NewInterfaceDescriptor(NewIfDescriptor(p.Log).GetDescriptor()))
	if err != nil {
		// handle error
	}

	err = p.KVScheduler.RegisterKVDescriptor(adapter.NewRouteDescriptor(NewRouteDescriptor(p.Log).GetDescriptor()))
	if err != nil {
		// handle error
	}

	return nil
}
```

The last step is to replace the plugin initialization method with `AllPlugins()` in the `main()` to ensure that the 
KV Scheduler is loaded and initialized from the hello world plugin.
```go
a := agent.NewAgent(agent.AllPlugins(p))
```

Starting the agent now will load the KV Scheduler plugin together with the hello world plugin. The KV Scheduler will
receive all northbound data and pass them to the hello world descriptor in correct order. If dependencies for a 
configuration item are not met (i.e. if a route is programmed before its interface dependency is is met), the item 
will be cached.

#### Example and testing

The example code from this tutorial can be found [here][4]. It contains `main.go`, `descriptors.go` and two folders with
model and generated adapters. The tutorial example is extended for the `AfterInit()` method which starts a new go routine
with testing procedure. It performs three test-cases:
The example can be build and started without any config files. Northbound transactions are simulated with KV Scheduler method `StartNBTransaction()`.

* **1. Configure the interface and the route in a single transaction**

This is the part of the output labelled as `planned operations`:
```bash
1. CREATE:
  - key: /interface/if1
  - value: { name:"if1"  } 
2. CREATE:
  - key: /route/route1
  - value: { name:"route1" interface_name:"if1"  } 
``` 

As expected, the interface is created first and the route second, following the order values were set to the transaction.

* **2. Configure the route and the interface in a single transaction (reversed order)**

Output:
```bash
1. CREATE:
  - key: /interface/if2
  - value: { name:"if2"  } 
2. CREATE:
  - key: /route/route2
  - value: { name:"route2" interface_name:"if2"  } 
```

The order is exactly the same despite the fact values were added to transaction in the reversed order. As we can see, the scheduler ordered configuration items before creating the transaction.

* **3. Configure the route and the interface in separated transactions**

In this case, we have two outputs since there are two transactions:

```bash
1. CREATE [NOOP IS-PENDING]:
  - key: /route/route3
  - value: { name:"route3" interface_name:"if3"  } 
```

The route comes first, but it is postponed (cached) since the dependent interface does not exist and the scheduler does not know when it appears, marking the route as `[NOOP IS-PENDING]`.

```bash
1. CREATE:
  - key: /interface/if3
  - value: { name:"if3"  } 
2. CREATE [WAS-PENDING]:
  - key: /route/route3
  - value: { name:"route3" interface_name:"if3"  } 
```

The second transaction introduced the expected interface. The scheduler recognized it as a dependency for the cached route, sorted items to correct order and called the appropriate configuration method. The previously cached route is marked as `[WAS-PENDING]`, highlighting that this item was postponed.

 [1]: /examples/tutorials/05_kv-scheduler/model/model.proto
 [2]: https://github.com/ligato/vpp-agent/tree/master/plugins/kvscheduler/descriptor-adapter
 [3]: /plugins/kvscheduler/api/kv_descriptor_api.go
 [4]: /examples/tutorials/05_kv-scheduler
 [5]: https://github.com/ligato/vpp-agent/blob/master/docs/kvscheduler/README.md
 
 
