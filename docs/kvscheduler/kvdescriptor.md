# Implementing your own KVDescriptor

KVDescriptor implements CRUD operations and defines derived values and
dependencies for a single value type. With these "descriptions",
the [KVScheduler](kvscheduler.md) is then able to manipulate with key-value
pairs generically, without having to understand what they actually represent.
The scheduler uses the learned dependencies, reads the SB state using provided
Dumps, and applies Add, Delete and Modify operations as needed to keep NB
in-sync with SB.

In VPP-Agent v2, all the VPP and Linux plugins were re-written (and decoupled
from each other), in a way that every supported configuration item is now
described by its own descriptor inside the corresponding plugin, i.e. there is
a descriptor for [Linux interfaces][linux-interface-descr],
[VPP interfaces][vpp-interface-descr], [VPP routes][vpp-route-descr], etc.
The full list of existing descriptors can be found [here][existing-descriptors].

This design pattern improves modularity, resulting in loosely coupled plugins,
allowing further extensibility beyond the already supported configuration items.
The KVScheduler is not even limited to VPP/Linux as the SB plane. Actually,
control plane for any system whose items can be represented as key-value pairs
and operated through CRUD operations qualifies for integration with the
framework. Here we provide a step-by-step guide to implementing and registering
your own KVDescriptor.

## Descriptor API

Let's start first by understanding the [descriptor API][descriptor-api].
First of all, descriptor is not an interface that needs to be implemented, but
rather a structure to be initialized with right attribute values and callbacks
to CRUD operations. This was chosen to reinforce the fact that descriptors are
meant to be `stateless` - the state of values is instead kept by the scheduler
and run-time information can be stored into the metadata optionally carried with
each value. The state of the graph with values and their metadata should
determine what exactly will be executed in the SB plane. The graph is already
exposed via formatted logs and REST API, therefore if descriptors do not hide
any state internally, the system state will be fully visible.

What follows is a list of all descriptor attributes, each with detailed
explanation. Optional fields can be left uninitialized (zero values).
* **Name** (string, mandatory)
    - put a name to your descriptor
    - it should be unique across all registered descriptors from all initialized
      plugins
* **KeySelector** (callback, mandatory)
    - provide a callback that will return true for keys identifying values
      described by your descriptor (i.e. check that the key matches the key
      template of your model)
* **ValueTypeName** (string, mandatory for [non-derived values](#derived-vals))
    - provide name of the protobuf message which defines your model
    - [here is an example][value-type-name] how the proto message name can be
      obtained from the generated type
* **KeyLabel** (callback, optional)
    - *optionally* provide callback that will "shorten the key" and return value
      identifier, that, unlike the original key, only needs to be unique in the
      key scope of the descriptor and not necessarily in the entire key space
      (e.g. interface name rather than the full key)
    - if defined, key label will be used as value identifier in the metadata map
      (i.e. it for example allows to ask for interface metadata simply by the
      interface name rather than using a full key)
* **ValueComparator** (callback, optional)
    - allows to optionally customize how two values are compared for equality 
    - normally, the scheduler compares two values for the same key using
      `proto.Equal` to determine if `Modify` operation is needed
    - sometimes, however, different values for the same field may be effectively
      equivalent - e.g. MTU 0 (default) might want to be treated as equivalent
      to MTU 1500 (i.e. change from 0 to 1500 or vice-versa should not trigger
      `Update`)
* **NBKeyPrefix** (string, optional)
    - put key prefix that the scheduler should watch in NB (e.g. `etcd`) to
      receive all values described by this descriptor
* **WithMetadata** (boolean, by default false)
    - enable if values should carry run-time metadata alongside the
      configuration
    - metadata allows to maintain extra state data that may change with CRUD
      operations or after agent restart and cannot be determined just from the
      value itself (e.g. sw_if_index for interface)
    - metadata are often used to correlate NB configuration with dumped SB data
* **MetadataMapFactory** (callback, optional)
    - can be used to provide a customized map implementation for value metadata,
      possibly extended with secondary lookups
    - if not defined, the scheduler will use the bare [NamedMapping][named-mapping]
      from the idxmap package.
* **Create** (callback, mandatory)
    - provide callback implementing operation to create a new value
      (C from CRUD)
* **Delete** (callback, mandatory)
    - provide callback implementing operation to delete an existing value
      (D from CRUD)  
* **Updatee** (callback, mandatory unless update is always performed with full
      re-creation)
    - provide callback implementing operation to update an existing value
      (U from CRUD)     
* **ModifyWithRecreate** (callback, optional - by default it is assumed that
  re-creation is not needed)
    - sometimes, for some or all kinds of updates, SB plane does not provide
      specific Update operation, instead the value has to be re-created
    - provide callback that will tell if the given value change requires full
      re-creation
* **Dump** (callback, optional)
  - provide callback implementing operation to read all values truly configured
    in the SB plane (R from CRUD)
  - it is optional in the sense that, if not provided, it is assumed that the
    Dump operation is not supported and therefore the state of SB for the given
    value type cannot be refreshed and will be assumed to be up-to-date
    (especially after an agent restart this might not be the case)        
* **IsRetriableFailure** (callback, optional)
    - optionally tell scheduler if the given error, returned by one of
      `Create`/`Delete`/`Update` handlers, will always be returned for the same
      value (non-retriable) or if the value can be theoretically fixed merely by
      repeating the operation
    - if the callback is not defined, every error will be considered retriable
* <a name="derived-vals">**DerivedValues**</a> (callback, optional, will be
  renamed to **Attributes** in the next release)
    - to break the value into multiple pieces managed separately by different
      descriptors, provide callback DerivedValues
    - derived value is typically a single field of the original value or its
      property, with possibly its own dependencies (dependency on the source
      value is implicit), custom implementations for CRUD operations and
      potentially used as a target for dependencies of other key-value pairs
    - for example, every [interface to be assigned to a bridge domain][bd-interface]
      is treated as a [separate key-value pair][bd-derived-vals], dependent on
      the [target interface to be created first][bd-iface-deps], but otherwise
      not blocking the rest of the bridge domain to be applied
* **Dependencies** (callback, optional)
    - for value that has one or more dependencies, provide callback that will
      tell which keys must already exist for the value to be considered ready
      for creation
    - dependency can be specified either exactly with a specific key, or using
      predicate `AnyOf` that must return true for at least one of the keys of
      already created values for the dependency to be considered satisfied
    - The callback is optional - if not defined, the kv-pairs of the descriptor
      are assumed to have no dependencies      
* **DumpDependencies** (slice of strings, optional)
  - if in order to dump values, some other descriptors have to be dumped first,
    here you can list them
  - [for example][dump-deps-example], in order to dump routes, interfaces have
    to be dumped first, to learn the mapping between interface names (NB ID) and
    their indexes (SB ID) from the metadata map of the interface plugin

## Descriptor Adapter

One inconvenience that you will quickly discover when using this generalized
approach of value description, is that the KVDescriptor API uses bare
`proto.Message` interface for values. It means that normally you cannot define
Add, Modify, Delete and other callbacks directly for your model, instead you
have to use `proto.Message` for input and output parameters and do all the
re-typing inside the callbacks.

To workaround this drawback, KVScheduler is shipped with a utility called
[descriptor-adapter][descriptor-adapter], generating an adapter for a given
value type that will prepare and hide all the type conversions.
The tool can be installed with:
```
make get-desc-adapter-generator
```

Then, to generate adapter for your descriptor, put `go:generate` command for
`descriptor-adapter` to (preferably) your plugin's main go file:
```
//go:generate descriptor-adapter --descriptor-name <your-descriptor-name>  --value-type <your-value-type-name> [--meta-type <your-metadata-type-name>] [--import <IMPORT-PATH>...] --output-dir "descriptor"
```
For example, `go:generate` for VPP interface can be found [here][vpp-iface-adapter].
The import paths have to include packages with your own data type definitions
for value (package with protobuf model) and, if used, also for metadata.
The import path can be relative to the file with the `go:generate` command
(hence the plugin's top-level directory is prefered).
Running `go generate <your-plugin-path>` will generate the adapter for your
descriptor into `adapter` sub-directory.

## Registering Descriptor

Once you have adapter generated and CRUD callbacks prepared, you can initialize
and register your descriptor.
First, import adapter into the go file with the descriptor
([assuming recommended directory layout](#plugin-directory-layout)):
```
import "github.com/<your-organization>/<your-agent>/plugins/<your-plugin>/descriptor/adapter"
```

Next, add constructor that will return your descriptor initialized and ready for
registration with the scheduler.
The adapter will present the KVDescriptor API with value type and metadata type
already casted to your own data types for every field:
```
func New<your-descriptor-name>Descriptor(<args>) *adapter.<your-descriptor-name>Descriptor {
	return &adapter.<your-descriptor-name>Descriptor{
		Name:        <your-descriptor-name>,
		KeySelector: <your-key-selector>,
                Add:         <your-Add-operation-implementation>,
                // etc., fill all the mandatory fields or whenever the default value is not suitable
	}
}
```

Next, inside the `Init` method of your plugin, import the package with all your
descriptors and register them using
[KVScheduler.RegisterKVDescriptor(<DESCRIPTOR>)][register-kvdescriptor] method:

```
import "github.com/<your-organization>/<your-agent>/plugins/<your-plugin>/descriptor"

func (p *YourPlugin) Init() error {
    yourDescriptor1 = descriptor.New<descriptor-name>Descriptor(<args>)
    p.Deps.KVScheduler.RegisterKVDescriptor(yourDescriptor1)
    ...
}
```

As you can see, the KVScheduler becomes plugin dependency, which needs to be
properly injected:
```
\\\\ plugin main go file:
import kvs "github.com/ligato/vpp-agent/plugins/kvscheduler/api"

type Deps struct {
    infra.PluginDeps
    KVScheduler kvs.KVScheduler
    ...
}

\\\\ options.go
import "github.com/ligato/vpp-agent/plugins/kvscheduler"

func NewPlugin(opts ...Option) *<your-plugin> {
    p := &<your-plugin>{}
    // ...
    p.KVScheduler = &kvscheduler.DefaultPlugin
    // ...
    return p
}
```

In order to obtain and expose the metadata map (if used), call
[KVScheduler.GetMetadataMap(<Descriptor-Name>)][get-metadata-map], after the
descriptor has been registered, which will give you a map reference that can be
then passed further via plugin's own API for other plugins to access read-only.
An example for VPP interface metadata map can be found
[here][get-metadata-map-example].

## <a name="plugin-directory-layout">Plugin Directory Layout</a>

While it is not mandatory, we recommend to follow the same directory layout used
for all VPP-agent plugins:
```
<your-plugin>/
├── model/  // + generated code
│   ├── model1.proto
│   ├── model2.proto
│   ├── ...
│   └── <modeln>.proto
├── descriptor/
│   ├── adapter/
│   │   ├── <generated-adapter-for-every-descriptor>...
│   ├── <descriptor-for-model1>.go
│   ├── <descriptor-for-model2>.go
│   ├── ...
│   └── <descriptor-for-modeln>.go
├── <metadata-map> // if custom secondary index over metadata is needed
│   └── <map-impl>.go
├── <your-plugin>.go
└── options.go
```

Directory `model` is where you would put all your proto models and the code
generated from it. `descriptor` directory is a place for all the descriptors
implemented by your plugin, optionally adapted for a specific protobuf type with
generated adapters nested further in the sub-directory `adapter` (adapters are
quite hidden since they should never need to be looked at and definitely not
edited manually). If you define custom metadata map, put the implementation into
a separate plugin's top-level directory, called for example `<model>idx`.
`<your-plugin.go>` is where you would implement the Plugin interface
(`Init`, `AfterInit`, `Close` methods) and register all the descriptors inside
the `Init` phase. It is a non-written rule to put plugin constructor and some
default options and default dependency injections into the file <options.go>
(example [option.go][options-example] for VPP ifplugin).

[existing-descriptors]: https://github.com/ligato/vpp-agent/wiki/KVDescriptors
[linux-interface-descr]: https://github.com/ligato/vpp-agent/blob/dev/plugins/linuxv2/ifplugin/descriptor/interface.go
[vpp-interface-descr]: https://github.com/ligato/vpp-agent/blob/dev/plugins/vppv2/ifplugin/descriptor/interface.go
[vpp-route-descr]: https://github.com/ligato/vpp-agent/blob/dev/plugins/vppv2/l3plugin/descriptor/static_route.go
[descriptor-api]: https://github.com/ligato/vpp-agent/blob/dev/plugins/kvscheduler/api/kv_descriptor_api.go#L99
[options-example]: https://github.com/ligato/vpp-agent/blob/dev/plugins/vppv2/ifplugin/options.go
[value-type-name]: https://github.com/ligato/vpp-agent/blob/dev/plugins/linuxv2/ifplugin/descriptor/interface.go#L144
[named-mapping]: https://github.com/ligato/cn-infra/blob/dev/idxmap/mem/inmemory_name_mapping.go
[bd-interface]: https://github.com/ligato/vpp-agent/blob/dev/plugins/vppv2/model/l2/bd.proto#L14
[bd-derived-vals]: https://github.com/ligato/vpp-agent/blob/dev/plugins/vppv2/l2plugin/descriptor/bridgedomain.go#L225
[bd-iface-deps]: https://github.com/ligato/vpp-agent/blob/dev/plugins/vppv2/l2plugin/descriptor/bd_interface.go#L128
[dump-deps-example]: https://github.com/ligato/vpp-agent/blob/dev/plugins/linuxv2/l3plugin/descriptor/route.go#L121
[descriptor-adapter]: https://github.com/ligato/vpp-agent/tree/dev/plugins/kvscheduler/descriptor-adapter
[vpp-iface-adapter]: https://github.com/ligato/vpp-agent/blob/dev/plugins/vppv2/ifplugin/ifplugin.go#L15
[register-kvdescriptor]: https://github.com/ligato/vpp-agent/blob/dev/plugins/kvscheduler/api/kv_scheduler_api.go#L206
[get-metadata-map]: https://github.com/ligato/vpp-agent/blob/dev/plugins/kvscheduler/api/kv_scheduler_api.go#L247
[get-metadata-map-example]: https://github.com/ligato/vpp-agent/blob/dev/plugins/vppv2/ifplugin/ifplugin.go#L167-L173
