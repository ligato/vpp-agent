# Tutorial: Add plugin to the KV scheduler

The tutorial shows how to extend our hello world plugin from previous tutorials to wire it with the KV scheduler. We will learn how to prepare a descriptor, generate the adapter and wire the plugin with the KV scheduler. 

Requirements:
* Complete and understand the ['Hello World Agent'](https://ligato.io/cn-infra/tutorials/01_hello-world) tutorial
* Complete and understand the ['Plugin Dependencies'](https://ligato.io/cn-infra/tutorials/02_plugin-deps) tutorial

For simplicity, this tutorial does not use ETCD or any other northbound database. Instead, the NB events are created manually in the example, using the KV Scheduler API.

The vpp-agent uses VPP API in form of binary API calls. Each VPP binary API call is designed to create a configuration item in the VPP or add or modify various parameters. In practice, these actions can be dependent on each other. The IP address can be assigned to an interface only if the interface is already present in the VPP. Another example is a VPP forwarding FIB entry, which can be added only if required interface and bridge domain exists and the interface is also assigned in the bridge domain, creating a complex dependency tree. In general, it is true that:

1. to configure all the proto-modelled data provided by the northbound, usually more than one binary API call is required
2. to configure parameters, a parent item is required to be present
2. some configuration items themselves are dependent on other and cannot be configured earlier

 It means that binary API calls have to be called in an exact order. The VPP agent uses a KV scheduler component to ensure it with system of caching and configuration dependency handling. Every plugin configuring something dependent on other plugin's configs in certain way can be added to the KV scheduler and profit from its advantages. 
 
 As a first step, we will use the special [proto model][1]. The model defines two simple messages - an `Interface` and a `Route` requiring some interface. The model demonstrates simple dependency between configuration items (since we need an interface to configure the route).  
 
 In order to register our hello world plugin to the scheduler and work with the new model, we need a two new components - a **descriptor** and an **adapter** for every proto-defined type.
 
 #### 1. Adapters
 
 Let's start with adapters. The purpose of the adapter is to define conversion methods between our proto-defined type and bare `proto.Message`. Since this code fulfils the definition of a boilerplate, we will generate it. The generator is called `descriptor-adapter` and can be found in the [inside the KVScheduler plugin][2]. Build the binary file from the go files inside, and use it to generate two adapters for the `Interface` and `Route` proto messages:
 
```
descriptor-adapter --descriptor-name Interface --value-type *model.Interface --import "github.com/ligato/vpp-agent/examples/tutorials/05_kv-scheduler/model" --output-dir "descriptor"
descriptor-adapter --descriptor-name Route --value-type *model.Route --import "github.com/ligato/vpp-agent/examples/tutorials/05_kv-scheduler/model" --output-dir "descriptor"
```

It is a good practice to add those commands to the plugin file with the `//go:generate` directives. Adapters were automatically created in the `adapter` directory within plugin folder.

#### 2. Descriptor without dependency

Another step is to define descriptors. We start with the interface descriptor, since there is no dependency on other messages. Let's create the `descriptors.go` so the code is outside of `main.go`, and define following code:

```go
type IfDescriptor struct {
	log logging.PluginLogger
}

func NewIfDescriptor(logger logging.PluginLogger) *IfDescriptor {
	return &IfDescriptor{
		log: logger,
	}
}
```

The `IfDescriptor` is a descriptor object which can define its own dependencies (used in descriptor methods). The `NewIfDescriptor` is a constructor (and also initializes descriptor dependencies). Next we define `GetDescriptor()` method which returns typed descriptor (from generated adapter code):

```go
func (d *IfDescriptor) GetDescriptor() *adapter.InterfaceDescriptor {
    return &adapter.InterfaceDescriptor{}
}
```

If you have a look at `adapter.InterfaceDescriptor`, you see that it defines several fields. Most important fields are function-types with CRUD definition and fields resolving dependencies. The full API list is documented in the [KvDescriptor structure][3]. Here, we implement the necessary ones:

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

* Key selector returns true if the provided key is valid for given descriptor. In the example, every descriptor uses own key, but there are cases where the descriptor processes other descriptor's keys or derived types.
```go
KeySelector: func(key string) bool {
    return true	
},
```

* This flag enables metadata for given type
```go
WithMetadata: true,
```

* Create method configures new configuration item.
```go
Create: func(key string, value *model.Car) (metadata interface{}, err error) {
    d.log.Infof("%s car created", value.Color)
    // Return color of the car so the scheduler remembers it
    return value.Color, nil
},
```

This is how the completed interface descriptor looks like:
```go
return &adapter.CarDescriptor{
    Name: "interface-descriptor",
    NBKeyPrefix: "/interface/",
    ValueTypeName: proto.MessageName(&model.Interface{}),
    KeyLabel: func(key string) string {
        return strings.TrimPrefix(key, "/interface/")
    },
    KeySelector: func(key string) bool {
        return true	
    },
    WithMetadata: true,
    Create: func(key string, value *model.Interface) (metadata interface{}, err error) {
        d.log.Infof("Interface %s created", value.Name)
        return value.Color, nil
    },
}
```

#### 3. Descriptor with dependency

This descriptor defines some additional fields, since we need to define dependency on the interface configuration item. Define the constructor and related methods first:
```go
type RouteDescriptor struct {
	// dependencies
	log logging.PluginLogger
}

func NewRouteDescriptor(logger logging.PluginLogger) *RouteDescriptor {
	return &RouteDescriptor{
		log: logger,
	}
}

func (d *RouteDescriptor) GetDescriptor() *adapter.RouteDescriptor {
	return &adapter.RouteDescriptor{}
	}
}
```

The route descriptor fields `Name`, `NBKeyPrefix`, `ValueTypeName`, `KeyLabel`and `KeySelector` are implemented in the same manner as for the interface type. The field `WithMetadata` is not needed here, so the `Create` method does not return any metadata value:
```go
Create: func(key string, value *model.Route) (metadata interface{}, err error) {
    d.log.Infof("Created route %s dependent on interface %s", value.Name, value.InterfaceName)
    return nil, nil
},
```

In addition, there are two new fields:

* A list of dependencies - complete keys (prefix+label) required for the given configuration item. The item will not be created while the dependent key does not exist. The label is informative and should be unique.
```go
Dependencies: func(key string, value *model.Route) []api.Dependency {
    return []api.Dependency{
        {
            Label: routeInterfaceDepLabel,
            Key:   ifPrefix + value.InterfaceName,
        },
    }
},
```

* A list of descriptors where the dependent values are processed. In the example, we return the interface descriptor since that is the one handling interfaces.
```go
RetrieveDependencies: []string{ifDescriptorName},
```

This is how the completed route descriptor looks like:
```go
func (d *RouteDescriptor) GetDescriptor() *adapter.RouteDescriptor {
	return &adapter.RouteDescriptor{
		Name: routeDescriptorName,
		NBKeyPrefix: routePrefix,
		ValueTypeName: proto.MessageName(&model.Route{}),
		KeyLabel: func(key string) string {
			return strings.TrimPrefix(key, routePrefix)
		},
		KeySelector: func(key string) bool {
			if strings.HasPrefix(key, routePrefix) {
				return true
			}
			return false
		},
		Dependencies: func(key string, value *model.Route) []api.Dependency {
			return []api.Dependency{
				{
					Label: routeInterfaceDepLabel,
					Key:   ifPrefix + value.InterfaceName,
				},
			}
		},
		RetrieveDependencies: []string{ifDescriptorName},
		Create: func(key string, value *model.Route) (metadata interface{}, err error) {
			d.log.Infof("Created route %s dependent on interface %s", value.Name, value.InterfaceName)
			return nil, nil
		},
	}
}
```

The descriptor API provides more methods not used in the example in order to keep it simple (like Update(), Delete(), Retrieve(), Validate(), ...). The full list can be found in the [descriptor API documentation][3]

#### Wire our plugin with the KV scheduler

Now when descriptors are completed, we will register them in `main.go`. First step is to add `KVScheduler` to the `HelloWorld` plugin as a plugin dependency:
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

And the last change is to replace plugin initialization method with `AllPlugins()` in the `main()` to ensure load of the KV scheduler from the hello world plugin.
```go
a := agent.NewAgent(agent.AllPlugins(p))
```

Starting the agent now will load KV scheduler plugin together with the hello world plugin. The KV scheduler will receive all the northbound data and passes them to the hello world descriptor in the correct order, or caches them if necessary.

#### Example and testing

The example code from this tutorial can be found [here][4]. It contains `main.go`, `descriptors.go` and two folders with model and generated adapters. The tutorial example is extended for the `AfterInit()` method which starts a new go routine with testing procedure. It performs three test-cases:
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

The order is exactly the same despite the fact values were added to transaction in the reversed order. As we cam see, the scheduler ordered configuration items before creating the transaction.

* **3. Configure the route and the interface in separated transactions**

In this case, we have two outputs since there are two transactions:

```bash
1. CREATE [NOOP IS-PENDING]:
  - key: /route/route3
  - value: { name:"route3" interface_name:"if3"  } 
```

The route comes first, but it is post-poned (cached) since the dependent interface does not exist and the scheduler does not know when it appears, marking the route as `[NOOP IS-PENDING]`.

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
 
 