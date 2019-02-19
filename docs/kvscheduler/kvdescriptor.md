# Implementing your own KVDescriptor

KVDescriptor implements CRUD operations and defines derived values and
dependencies for a single value type. With these "descriptions",
the [KVScheduler](kvscheduler.md) is then able to manipulate with key-value
pairs generically, without having to understand what they actually represent.
The scheduler uses the learned dependencies, reads the SB state using provided
`Retrieve` callbacks, and applies `Create`, `Delete` and `Update` operations
as needed to keep NB in-sync with SB.

In VPP-Agent v2, all the VPP and Linux plugins were re-written (and decoupled
from each other), in a way that every supported configuration item is now
described by its own descriptor inside the corresponding plugin, i.e. there is
a descriptor for [Linux interfaces][linux-interface-descr],
[VPP interfaces][vpp-interface-descr], [VPP routes][vpp-route-descr], etc.
A full list of existing descriptors can be found [here][existing-descriptors].

This design pattern improves modularity, resulting in loosely coupled plugins,
allowing further extensibility beyond the already supported configuration items.
The KVScheduler is not even limited to VPP/Linux as the SB plane. Actually,
control plane for any system whose items can be represented as key-value pairs
and operated through CRUD operations qualifies for integration with the
framework. Here we provide a step-by-step guide on how to implement and register
your own KVDescriptor.

## Descriptor API

Let's start first by understanding the [descriptor API][descriptor-api].
First of all, descriptor is not an interface that needs to be implemented, but
rather a structure to be initialized with the right attribute values and callbacks
to CRUD operations. This was chosen to reinforce the fact that descriptors are
meant to be **stateless** - the state of values is instead kept by the scheduler
and run-time information can be stored into the [metadata][kvscheduler-terminology],
optionally carried with each value. The state of the graph with values and their
metadata should determine what exactly will be executed next in the SB plane
for a given transaction.
The graph is already exposed via formatted logs and programming+REST APIs,
therefore if descriptors do not hide any state internally, the system state will
be fully visible from the outside.

What follows is a list of all descriptor attributes, split across sub-sections,
each with a detailed explanation and pointers to examples. Optional fields can
be left uninitialized (zero values).

Please note that using [descriptor adapters](#descriptor-adapter), the signatures
of the callbacks will become adapted to use the real proto message type
(e.g. `*vpp_l3.Route`) as opposed to the bare `proto.Message` interface, avoiding
all the boiler-plate type casting.

**Note**: `KeySelector`, `ValueTypeName`, `KeyLabel` & `NBKeyPrefix`
will all be replaced in a future release with a single reference to the value
[model][kvscheduler-terminology] (**TODO: add link to the model documentation
once it exists**). Most descriptors already use the methods provided by models
to define these fields. But we do not yet have tools to build models for
[derived values](#derivedvalues) and without them we cannot fully switch
to models.

### Name

* `string` attribute, **mandatory**
* name (ID) of the descriptor
* it should be unique across all registered descriptors from all initialized
  plugins

### NBKeyPrefix

* `string` attribute, optional
* key prefix that the scheduler should watch in NB (e.g. `etcd`) to receive all
  values described by this descriptor
* descriptors for derived values do not need to define this field - the values
  they describe do not come from NB directly, instead get derived from other
  values which are in the scope of other descriptors
* [model][kvscheduler-terminology] can be used to obtain the key prefix
  using `KeyPrefix()` method - [here is an example][nb-key-prefix]

### KeySelector

* **mandatory** callback: `func(key string) bool`
* a predicate that should select (i.e. return true) for keys identifying values
  described by the given descriptor
* typically, selector uses `IsKeyValid` from the value [model][kvscheduler-terminology]
  to check if the key is valid for the model - [here is an example][key-selector]

### ValueTypeName

* `string` attribute, **mandatory for [non-derived values](#derivedvalues)**
* name of the protobuf message used to structure and serialize value data
* [model][kvscheduler-terminology] can be used to obtain the proto message name
  using `ProtoName()` method - [here is an example][value-type-name]

### KeyLabel

* optional callback: `func(key string) string`
* "a key shortener" - function that will receive a key and should return value
   identifier, that, unlike the original key, only needs to be unique in the
   key scope of the descriptor and not necessarily in the entire key space
   (e.g. interface name rather than the full key)
* [model][kvscheduler-terminology] provides key shortener off-the-shelf with
  the method `StripKeyPrefix()` - [here is an example][key-label]
* if defined, key label will be used as value identifier in the metadata map
  (it then for example allows to ask for interface metadata simply by the
  interface name rather than using a full key)

### ValueComparator

* optional callback: `func(key string, oldValue, newValue proto.Message) bool`
* allows to optionally customize how two values are compared for equality
* normally, the scheduler compares two values of the same key using `proto.Equal`
  to determine if `Update` operation is needed
* sometimes, however, different values for the same field may be effectively
  equivalent - [for example][compare-mtu], MTU 0 (default) might want to be
  treated as equivalent to MTU 1500 (i.e. change from 0 to 1500 or vice-versa
  should not trigger `Update`)

### WithMetadata

* `boolean` attribute, optional, by default `false`
* enable if values should carry run-time metadata alongside the configuration
* metadata allows to maintain extra state data that may change with CRUD
  operations or after agent restart and cannot be determined just from the
  value itself (e.g. sw_if_index for interface)
* metadata are often used in [Retrieve](#retrieve) to correlate NB configuration
  with retrieved SB data
* metadata are not supported with [derived values](#derivedvalues)
* **note**: in a future release the term "metadata" will be renamed to "statedata",
  which is more fitting

### MetadataMapFactory

* optional callback: `func() idxmap.NamedMappingRW`
* can be used to provide a customized map implementation for value metadata,
  possibly extended with additional secondary lookups
* if not defined, the scheduler will use the bare [NamedMapping][named-mapping]
  from the idxmap package
* for example, VPP [ifplugin][vpp-ifplugin] implements custom map called
  [ifaceidx][vpp-ifaceidx], which allows to [lookup interfaces by the assigned
  IP addresses][vpp-iface-by-ip] among other things

### Validate

* optional callback: `func(key string, value proto.Message) error`
* can be provided to implement validation of the value data received from NB
  (e.g. check for validity of interface configuration)
* `Validate` is called for every new value before it is Created or Updated into
* if the validations fails (returned error is non-nil), the scheduler will
  mark the value as invalid ([state][value-states] `INVALID`) and will not
  attempt to apply it
* the descriptor can further specify which field(s) are not valid by wrapping
  the validation error together with a slice of invalid fields using the error
  [InvalidValueError][invalid-val-error].
* for example, interface cannot be both unnumbered and at the same time have
  IP address assigned - this is validated [here][unnumbered-validation]

### Create

* callback: `func(key string, value proto.Message) (metadata interface{}, err error)`
* "C" from CRUD, implementing operation to create a new value
* **mandatory for descriptors that describe values received from NB**, but
  optional for descriptors with only `OBTAINED` values in their scope - i.e.
  values received from SB via notifications as already created
* for non-derived values, descriptor may return metadata to associate with
  the value
* KVScheduler ensures that all the dependencies are satisfied when the Create
  is being called
* for example, descriptor for VPP ARP entries simply [adds new ARP entry][vpp-arp-create]
  defined by the value with configuration - it knows that the associated
  interface it depends on is guaranteed to already exist by the scheduling
  algorithm and can therefore [read the interface index from its metadata][vpp-arp-get-iface-index],
  needed to build request for VPP

### Delete

* callback: `func(key string, value proto.Message, metadata Metadata) error`
* "D" from CRUD, implementing operation to delete an existing value
* **mandatory for descriptors that describe values received from NB**, but
  optional for descriptors with only `OBTAINED` values in their scope - i.e.
  values received from SB via notifications as already created
* KVScheduler ensures that all the items that depend on a value which is being
  removed are deleted first and put into the `PENDING` state
* for example, descriptor for VPP ARP entries simply [removes existing ARP entry][vpp-arp-delete],
  knowing that the scheduling algorithm guarantees that the associated interfaces,
  marked as a dependency of the ARP entry, will not be removed before the
  ARP and therefore the interface index, needed to build the delete request for
  VPP, can still be [read from the interfaces metadata][vpp-arp-get-iface-index]

### Update

* callback: `func(key string, oldValue, newValue proto.Message, oldMetadata interface{}) (newMetadata interface{}, err error)`
* "U" from CRUD, implementing operation to update an existing value
* the callback is optional - if undefined, updates will be always performed
  via re-creation, i.e. `Delete(key, oldValue, oldMetadata)` followed by
  `newMetadata, err = Create(key, newValue)`
* the current value metadata, passed to the callback as `oldMetadata`, can be
  edited in-place (i.e. without deep-copying) and returned as `newMetadata`
* not all the configuration updates are supported by SB to apply incrementally
  \- for example, changing interface type (e.g. going from VETH to TAP) cannot
  be done without fully re-creating the interface - on the other hand, Linux
  [interface host name can be changed][linux-rename-interface] via dedicated
  netlink call

### UpdateWithRecreate

* optional callback: `func(key string, oldValue, newValue proto.Message, oldMetadata interface{}) (newMetadata interface{}, err error)`
* sometimes, for some or all kinds of updates, SB plane does not provide
  specific Update operations, instead the value has to be re-created,
  i.e. calling `Delete(key, oldValue, oldMetadata)` followed by
  `newMetadata, err = Create(key, newValue)`
* through this callback the KVScheduler can be informed if the given value
  change requires full re-creation
* if not defined, KVScheduler will decide based on the (un)availability of the
  `Update` operation - if provided, it is assumed that any change can be applied
  incrementally, otherwise a full re-creation is the only way to go
* [for example][vpp-iface-recreate], changing VPP interface type (e.g. going
  from MEMIF to TAP) cannot be done without fully re-creating the interface

### Retrieve

* optional callback: `func(correlate []KVWithMetadata) ([]KVWithMetadata, error)`
  - where:
 ``` golang
 // KVWithMetadata encapsulates key-value pair with metadata and the origin mark.
 type KVWithMetadata struct {
     Key      string
     Value    proto.Message
     Metadata Metadata
     Origin   ValueOrigin
 }
 ```
* "R" from CRUD, implementing operation to read all the values in the scope of
  the descriptor, truly configured in the SB plane at that moment
* it is a key operation for state reconciliation (or as we call it - "resync"),
  as it gives KVScheduler the ability to refresh it's view of SB and determine
  the sequence of `Create`/`Update`/`Delete` operations needed to get the actual
  state (SB) in-sync with the desired state (NB)
* it is optional in the sense that, if not provided, it is assumed that the
  `Retrieve` operation is not supported and therefore the state of SB for the
  given value type cannot be refreshed and will be assumed to be up-to-date
  (but especially after an agent restart this might not be the case)
* input argument `correlate` represents the non-derived values currently created
  or getting applied as viewed from the northbound/scheduler point of view:
    - startup resync: `correlate` = values received from NB to be applied
	- run-time/downstream resync: `correlate` = cached values taken and applied
	  from the graph of values stored in-memory (scheduler's cached view of SB)

### IsRetriableFailure

* callback: `func(err error) bool`
* optionally tell scheduler if the given error, returned by one of the
  `Create`/`Delete`/`Update` callbacks, will always be returned for the same
  value (non-retriable) or if the value can be theoretically fixed merely by
  repeating the operation (retriable)
* if the callback is not defined, every C(R)UD error will be considered retriable
* validation errors (returned from [Validate](#validate)) are automatically
  considered non-retriable - no matter how many times an invalid configuration
  is re-applied, it is still invalid and the operation would fail
* if a C(R)UD operation fails with a retriable error and the associated
  (`best-effort`) transaction [allows Retry][retry-opt], the KVscheduler will
  schedule repeat for these failed operations to run in a separate transaction,
  triggered after a configurable time delay and with a limit to the maximum
  number of retries allowed - it can be requested to double the delay for every
  next attempt, feature known as exponential backoff

### DerivedValues

* optional callback: `func(key string, value proto.Message) []KeyValuePair`
* to break a complex value into multiple pieces managed separately by different
  descriptors, implement and provide callback `DerivedValues`
* derived value is typically a single field of the original value or its
  property, with possibly its own dependencies (dependency on the source value
  is implicit, i.e. source value is created before its derived values), custom
  implementations for CRUD operations and potentially used as a target for
  dependencies of other key-value pairs
* for example, every [interface to be assigned to a bridge domain][bd-interface]
  is treated as a [separate key-value pair][bd-derived-vals], dependent on
  the [target interface to be created first][bd-iface-deps], but otherwise
  not blocking the rest of the bridge domain to be applied

### Dependencies

* optional callback: `func(key string, value proto.Message) []Dependency`
  - where:
 ``` golang
 // Dependency references another kv pair that must exist before the associated
 // value can be created.
 type Dependency struct {
     // Label should be a short human-readable string labeling the dependency.
     // Must be unique in the list of dependencies for a value.
     Label string

     // Key of another kv pair that the associated value depends on.
     // If empty, AnyOf must be defined instead.
     Key string

     // AnyOf, if not nil, must return true for at least one of the already
     // created keys for the dependency to be considered satisfied.
     // Either Key or AnyOf should be defined, but not both at the same time.
     // Note: AnyOf comes with more overhead than a static key dependency,
     // so prefer to use the latter whenever possible.
     AnyOf KeySelector
 }

 // KeySelector is used to filter keys.
 type KeySelector func(key string) bool
 ```
* for value that has one or more dependencies, provide callback that will
  tell which keys must already exist for the value to be considered ready
  for creation
* dependency can either reference a specific key, or use the predicate `AnyOf`,
  which must return `true` for at least one of the keys of already created
  values for the dependency to be considered satisfied (i.e. matching keys are
  basically ORed)
* the callback is optional - if not defined, the kv-pairs of the descriptor
  are assumed to have no dependencies
* multiple listed dependencies must all be satisfied - i.e. they are ANDed
* [a basic example][vpp-arp-deps] is ARP entry which cannot be created until
  the associated interface exists
* [a more complex dependency][linux-route-gw-dep], which cannot be expressed
  using a static key but requires `AnyOf` selector, can be found in the
  linux/l3plugin: a Linux route cannot be created (request will fail) if the
  selected gateway (next hop) isn't already routable based on IP addresses
  assigned to interfaces from the same namespace or using link-local routes
  \- the implemented `AnyOf` selector returns `true` whenever it finds just such
  interface or a link-local route among already configured/obtained values

### RetrieveDependencies

* optional attribute, slice of strings
* if, in order to `Retrieve` values, some other descriptors have to be have
  their values refreshed first, here you can list them
* [for example][vpp-route-retrieve-deps], in order to retrieve routes and
  re-construct their configuration for NB models, interfaces have to be
  retrieved first, to learn the mapping between interface names (NB ID)
  and their indexes (SB ID) from the metadata map of the interface plugin
  - this is because the retrieved routes will reference outgoing interfaces
    through SB indexes, which need to be [translated into the logical names from
    NB][vpp-route-iface-name]

## Descriptor Adapter

One inconvenience that you will quickly discover when using this generalized
approach of unified value descriptions, is that the KVDescriptor API uses bare
`proto.Message` interface for all values. It means that normally you cannot
define Create, Update, Delete and other callbacks directly for your model,
instead you have to use `proto.Message` for input and output parameters and do
all the re-typing inside the callbacks, introducing lot's of boiler-plate code
into your descriptors.

To workaround this drawback, KVScheduler is shipped with a utility called
[descriptor-adapter][descriptor-adapter], generating an adapter for a given
value type that will prepare and hide all the type conversions.
The tool can be installed with:
```
make get-desc-adapter-generator
```

Then, to generate adapter for your descriptor, put `go:generate` command for
`descriptor-adapter` to (preferably) your plugin's main go file:
``` golang
//go:generate descriptor-adapter --descriptor-name <your-descriptor-name>  --value-type <your-value-type-name> [--meta-type <your-metadata-type-name>] [--import <IMPORT-PATH>...] --output-dir "descriptor"
```
Available arguments are:
* `output-dir`: output directory where "adapter" package will be generated
  - optional, current working directory is the default
  - it is recommended to generate adapter under the directory with the descriptor,
    therefore all the VPP and Linux plugins use `--output-dir "descriptor"`
* `descriptor-name`: name of the descriptor
  - mandatory argument
  - can be a short cut of the full descriptor name to avoid overly long generated
    type and variable names - the only requirement is that it is unique among
    the adapters of the same plugin
* `value-type`: type of the described values, e.g. `*vpp_l2.BridgeDomain`,
  `*vpp_l3.Route`, etc.
  - mandatory argument
* `meta-type`: type of the metadata carried along values, e.g. `*aclidx.ACLMetadata`
  (ACL metadata), `*idxvpp.OnlyIndex` (generic metadata storing only object
  index used in VPP), etc.
  - optional argument, by default the metadata are used with undefined type
    (`interface{}`)
* `import`: a list of packages to import into the generated adapter to get
  definition of the value type (i.e. package with the protobuf model) and, if
  used, also for metadata type
  - repeated and optional argument, but since `value-type` is mandatory, at least
    the package with the protobuf model of the value should be imported
  - import path can be relative to the file with the `go:generate` command
    (hence the plugin's top-level directory is preferred to avoid double dots)

For example, `go:generate` for VPP interface can be found [here][vpp-iface-adapter].
Running `go generate <your-plugin-path>` will generate the adapter for your
descriptor into `adapter` directory under `output-dir`.

## Registering Descriptor

Once you have adapter generated and CRUD callbacks prepared, you can initialize
and register your descriptor.
First, import adapter into the go file with the descriptor
([assuming recommended directory layout](#plugin-directory-layout)):
```golang
import "github.com/<your-organization>/<your-agent>/plugins/<your-plugin>/descriptor/adapter"
```

Next, add constructor that will return your descriptor initialized and ready for
registration with the scheduler.
The adapter will present the KVDescriptor API with value type and metadata type
already casted to your own data types for every field:
```golang
func New<your-descriptor-name>Descriptor(<args>) *KVDescriptor {
    typedDescriptor := &adapter.<your-descriptor-name>Descriptor{
        Name:        <your-descriptor-name>,
        KeySelector: <your-key-selector>,
        Create:      <your-Create-operation-implementation>,
        // etc., fill all the mandatory fields or whenever the default value is not suitable
    }
    return adapter.New<your-descriptor-name>Descriptor(typedDescriptor)
}
```
Note: even though descriptors are meant to be stateless, it is still common to
implement CRUD operations as methods of a structure. But the structure should
only serve as a "static context" for the descriptor, storing for example
references to the logger, SB handler(s), etc. - things that do not change once
the descriptor is constructed (typically received as input arguments for the
descriptor constructor) .

Next, inside the `Init` method of your plugin, import the package with all your
descriptors and register them using
[KVScheduler.RegisterKVDescriptor(<DESCRIPTOR>)][register-kvdescriptor] method:

```golang
import "github.com/<your-organization>/<your-agent>/plugins/<your-plugin>/descriptor"

func (p *YourPlugin) Init() error {
    yourDescriptor1 = descriptor.New<descriptor-name>Descriptor(<args>)
    p.Deps.KVScheduler.RegisterKVDescriptor(yourDescriptor1)
    ...
}
```

As you can see, the KVScheduler becomes plugin dependency, which needs to be
properly injected:
```golang
\\\\ plugin main go file:
import kvs "github.com/ligato/vpp-agent/plugins/kvscheduler/api"

type Deps struct {
    infra.PluginDeps
    KVScheduler kvs.KVScheduler
    ...
}

\\\\ options.go
import kvs "github.com/ligato/vpp-agent/plugins/kvscheduler"

func NewPlugin(opts ...Option) *<your-plugin> {
    p := &<your-plugin>{}
    // ...
    p.KVScheduler = &kvs.DefaultPlugin
    // ...
    return p
}
```


<a name="expose-metadata">In order to obtain and expose the metadata map</a>
(if used), call [KVScheduler.GetMetadataMap(<Descriptor-Name>)][get-metadata-map],
after the descriptor has been registered, which will give you a map reference
that can be then passed further via plugin's own API for other plugins to access
read-only. An example for VPP interface metadata map can be found
[here][get-metadata-map-example].

## Plugin Directory Layout

While it is not mandatory, we recommend to follow the same directory layout used
across all the VPP-Agent plugins:
```
<your-plugin>/
├── model/  // + protobuf-generated code
│   ├── model1.proto
│   ├── model2.proto
│   ├── ...
│   └── <modeln>.proto
├── descriptor/
│   ├── adapter/
│   │   ├── <generated-adapter-for-every-descriptor>
│   │   └──...
│   ├── <descriptor-for-model1>.go  // e.g. "acl.go"
│   ├── <descriptor-for-model2>.go
│   ├── ...
│   └── <descriptor-for-modeln>.go
├── <southband-name>calls/  // e.g. "vppcalls/"
├── <metadata-map>  // if custom secondary index for metadata is needed
│   └── <map-impl>.go
├── <your-plugin>.go
├── <your-plugin>_api.go
└── options.go
```

Directory `model` is where you would put all your proto models and the code
generated from it.

`descriptor` directory is a place for all the descriptors
implemented by your plugin, optionally adapted for a specific protobuf type with
generated adapters nested further in the sub-directory `adapter` (adapters are
quite hidden since they should never need to be looked at and definitely not
edited manually).

It is recommended to put implementation of every SB call needed for your
descriptor into a separate package `<southband-name>calls/`
(e.g. [linuxcalls][linuxcalls]) and expose them via interface. This will allow
to replace access to SB with mocks and make unit testing easier.

If you define custom metadata map, put the implementation into a separate
plugin's top-level directory, called for example `<model>idx`.

`<your-plugin.go>` is where you would implement the Plugin interface
(`Init`, `AfterInit`, `Close` methods) and register all the descriptors within
the `Init` phase.

`<your-plugin>_api.go` is a place to define API of your plugin - plugins most
commonly [expose read-only references to maps with metadata](#expose-metadata)
for configuration items they describe.

It is a non-written rule to put plugin constructor and some default options and
default dependency injections into the file `options.go` (example
[option.go][options-example] for VPP ifplugin).

## Descriptor examples

### Descriptor skeletons

TODO

### Mock SB

We have prepared an [interactive hands-on example][mock-plugins-example],
demonstrating the KVScheduler framework using replicated `vpp/ifplugin` and
`vpp/l2plugin`, where models are simplified and the VPP is replaced with a mock
southbound, printing the triggered CRUD operations into the stdout instead of
actually executing them. The example is fully focused on the scheduler and the
descriptors, and on that abstraction level the actual SB underneath is irrelevant. 

### Real-world examples

Since all the VPP and Linux plugins use the KVScheduler framework, there is
already plenty of descriptors available in the repository to take inspiration
from. Even though interfaces are the basis of network configuration, we recommend
to start studying descriptors for simpler objects, such as [VPP ARPs][vpp-arp-descriptor]
and [VPP routes][vpp-route-descriptor], which have simple CRUD methods
and a single dependency on the associated interface. Then, learn how to break
a more complex object into multiple values using [bridge domains][vpp-bd-descriptor]
and [BD-interface bindings][vpp-bd-iface-descriptor], derived one for every
interface to be assigned into the domain, as an example. Finally, check out the
[Linux interface watcher][linux-iface-watcher], which shows that values may enter
the graph even from below as SB notifications, and used as [targets for dependencies][afpacket-dep]
by other objects.
These descriptors cover most of the features and should help you to get started
implementing your own.


[existing-descriptors]: https://github.com/ligato/vpp-agent/wiki/KVDescriptors
[linux-interface-descr]: https://github.com/ligato/vpp-agent/blob/master/plugins/linux/ifplugin/descriptor/interface.go
[vpp-interface-descr]: https://github.com/ligato/vpp-agent/blob/master/plugins/vpp/ifplugin/descriptor/interface.go
[vpp-route-descr]: https://github.com/ligato/vpp-agent/blob/master/plugins/vpp/l3plugin/descriptor/route.go
[descriptor-api]: https://github.com/ligato/vpp-agent/blob/e8e54ef67b666e57ffef1bca555c8ce5585f215f/plugins/kvscheduler/api/kv_descriptor_api.go#L82-L248
[kvscheduler-terminology]: kvscheduler.md#terminology
[descriptor-adapter]: https://github.com/ligato/vpp-agent/tree/master/plugins/kvscheduler/descriptor-adapter
[vpp-iface-adapter]: https://github.com/ligato/vpp-agent/blob/e8e54ef67b666e57ffef1bca555c8ce5585f215f/plugins/vpp/ifplugin/ifplugin.go#L15
[register-kvdescriptor]: https://github.com/ligato/vpp-agent/blob/e8e54ef67b666e57ffef1bca555c8ce5585f215f/plugins/kvscheduler/api/kv_scheduler_api.go#L195-L199
[get-metadata-map]: https://github.com/ligato/vpp-agent/blob/e8e54ef67b666e57ffef1bca555c8ce5585f215f/plugins/kvscheduler/api/kv_scheduler_api.go#L228-L231
[get-metadata-map-example]: https://github.com/ligato/vpp-agent/blob/e8e54ef67b666e57ffef1bca555c8ce5585f215f/plugins/vpp/ifplugin/ifplugin.go#L177-L183
[options-example]: https://github.com/ligato/vpp-agent/blob/master/plugins/vpp/ifplugin/options.go
[vpp-arp-descriptor]: https://github.com/ligato/vpp-agent/blob/master/plugins/vpp/l3plugin/descriptor/arp_entry.go
[vpp-route-descriptor]: https://github.com/ligato/vpp-agent/blob/master/plugins/vpp/l3plugin/descriptor/route.go
[vpp-bd-descriptor]: https://github.com/ligato/vpp-agent/blob/master/plugins/vpp/l2plugin/descriptor/bridgedomain.go
[vpp-bd-iface-descriptor]: https://github.com/ligato/vpp-agent/blob/master/plugins/vpp/l2plugin/descriptor/bd_interface.go
[linux-iface-watcher]: https://github.com/ligato/vpp-agent/blob/master/plugins/linux/ifplugin/descriptor/watcher.go
[afpacket-dep]: https://github.com/ligato/vpp-agent/blob/e8e54ef67b666e57ffef1bca555c8ce5585f215f/plugins/vpp/ifplugin/descriptor/interface.go#L421-L426
[value-type-name]: https://github.com/ligato/vpp-agent/blob/e8e54ef67b666e57ffef1bca555c8ce5585f215f/plugins/linux/ifplugin/descriptor/interface.go#L145
[key-label]: https://github.com/ligato/vpp-agent/blob/e8e54ef67b666e57ffef1bca555c8ce5585f215f/plugins/linux/ifplugin/descriptor/interface.go#L147
[key-selector]: https://github.com/ligato/vpp-agent/blob/e8e54ef67b666e57ffef1bca555c8ce5585f215f/plugins/linux/ifplugin/descriptor/interface.go#L146
[nb-key-prefix]: https://github.com/ligato/vpp-agent/blob/e8e54ef67b666e57ffef1bca555c8ce5585f215f/plugins/linux/ifplugin/descriptor/interface.go#L144
[compare-mtu]: https://github.com/ligato/vpp-agent/blob/e8e54ef67b666e57ffef1bca555c8ce5585f215f/plugins/linux/ifplugin/descriptor/interface.go#L191-L194
[named-mapping]: https://github.com/ligato/cn-infra/blob/master/idxmap/mem/inmemory_name_mapping.go
[vpp-ifaceidx]: https://github.com/ligato/vpp-agent/tree/master/plugins/vpp/ifplugin/ifaceidx
[vpp-iface-by-ip]: https://github.com/ligato/vpp-agent/blob/master/plugins/vpp/ifplugin/ifaceidx/ifaceidx.go#L135-L139
[vpp-ifplugin]: https://github.com/ligato/vpp-agent/tree/master/plugins/vpp/ifplugin
[value-states]: https://github.com/ligato/vpp-agent/blob/master/plugins/kvscheduler/api/value_status.proto
[invalid-val-error]: https://github.com/ligato/vpp-agent/blob/e8e54ef67b666e57ffef1bca555c8ce5585f215f/plugins/kvscheduler/api/errors.go#L124-L160
[unnumbered-validation]: https://github.com/ligato/vpp-agent/blob/e8e54ef67b666e57ffef1bca555c8ce5585f215f/plugins/vpp/ifplugin/descriptor/interface.go#L380-L385
[vpp-arp-create]: https://github.com/ligato/vpp-agent/blob/e8e54ef67b666e57ffef1bca555c8ce5585f215f/plugins/vpp/l3plugin/descriptor/arp_entry.go#L79-L85
[vpp-arp-get-iface-index]: https://github.com/ligato/vpp-agent/blob/e8e54ef67b666e57ffef1bca555c8ce5585f215f/plugins/vpp/l3plugin/vppcalls/arp_vppcalls.go#L27-L37
[vpp-arp-delete]: https://github.com/ligato/vpp-agent/blob/e8e54ef67b666e57ffef1bca555c8ce5585f215f/plugins/vpp/l3plugin/descriptor/arp_entry.go#L88-L93
[linux-rename-interface]: https://github.com/ligato/vpp-agent/blob/e8e54ef67b666e57ffef1bca555c8ce5585f215f/plugins/linux/ifplugin/descriptor/interface.go#L413-L420
[linuxcalls]: https://github.com/ligato/vpp-agent/tree/master/plugins/linux/ifplugin/linuxcalls
[vpp-iface-recreate]: https://github.com/ligato/vpp-agent/blob/e8e54ef67b666e57ffef1bca555c8ce5585f215f/plugins/vpp/ifplugin/descriptor/interface.go#L393
[retry-opt]: https://github.com/ligato/vpp-agent/blob/e8e54ef67b666e57ffef1bca555c8ce5585f215f/plugins/kvscheduler/api/txn_options.go#L136-L184
[bd-interface]: https://github.com/ligato/vpp-agent/blob/e8e54ef67b666e57ffef1bca555c8ce5585f215f/api/models/vpp/l2/bridge-domain.proto#L19-L23
[bd-derived-vals]: https://github.com/ligato/vpp-agent/blob/e8e54ef67b666e57ffef1bca555c8ce5585f215f/plugins/vpp/l2plugin/descriptor/bridgedomain.go#L240-L251
[bd-iface-deps]: https://github.com/ligato/vpp-agent/blob/e8e54ef67b666e57ffef1bca555c8ce5585f215f/plugins/vpp/l2plugin/descriptor/bd_interface.go#L119-L127
[vpp-arp-deps]: https://github.com/ligato/vpp-agent/blob/e8e54ef67b666e57ffef1bca555c8ce5585f215f/plugins/vpp/l3plugin/descriptor/arp_entry.go#L116-L126
[linux-route-gw-dep]: https://github.com/ligato/vpp-agent/blob/e8e54ef67b666e57ffef1bca555c8ce5585f215f/plugins/linux/l3plugin/descriptor/route.go#L255-L273
[vpp-route-retrieve-deps]: https://github.com/ligato/vpp-agent/blob/e8e54ef67b666e57ffef1bca555c8ce5585f215f/plugins/vpp/l3plugin/descriptor/route.go#L74
[vpp-route-iface-name]: https://github.com/ligato/vpp-agent/blob/e8e54ef67b666e57ffef1bca555c8ce5585f215f/plugins/vpp/l3plugin/vppcalls/route_dump.go#L139-L150
[mock-plugins-example]: ../../examples/kvscheduler/mock_plugins/README.md