# KVScheduler Troubleshooting Guide

## Index

* [Common issues](#common-issues)
* [Commonly returned errors](#commonly-returned-errors)
* [Common programming mistakes / bad practises](#programming-mistakes)

## Common issues

### Value entered via NB was not configured in SB
[Read transaction logs](debugging.md#understanding-transaction-log), printed
by the agent into `stdout` and try to locate the transaction triggered
to configure the value:

1. **transaction is not triggered** (not found in the logs) or the **value is
   missing** in the transaction input
   - make sure the model is registered
     - note: [not using model is considered a programming error](#models-bad-usage)
     (**TODO**: link model documentation once it exists)
    - [check if the plugin implementing the value is loaded](debugging.md#how-to-debug-agent-plugin-lookup)
    - [check if the descriptor associated with the value is registered](debugging.md#how-to-list-descriptors)
    - [check if the key prefix is being watched](debugging.md#how-to-list-descriptors)
    - for NB KVDB make sure the key under which the value was put is correct
       - could be a bad key prefix (bad suffix would make the value `UNIMPLEMENTED`,
         but otherwise included in the transaction input)
    - [debug logs](debugging.md#how-to-enable-agent-debug-logs) of the Orchestrator
      (NB of KVScheduler), can also be used to learn the set of key-values received
      with each event from NB - following are examples for RESYNC and CHANGE events:
  ```
    DEBU[0005] => received RESYNC event (1 prefixes)         loc="orchestrator/orchestrator.go(150)" logger=orchestrator.dispatcher
    DEBU[0005]  -- key: config/mock/v1/interfaces/tap1       loc="orchestrator/orchestrator.go(168)" logger=orchestrator.dispatcher
    DEBU[0005]  -- key: config/mock/v1/interfaces/loopback1  loc="orchestrator/orchestrator.go(168)" logger=orchestrator.dispatcher
    DEBU[0005] - "config/mock/v1/interfaces/" (2 items)      loc="orchestrator/orchestrator.go(173)" logger=orchestrator.dispatcher
    DEBU[0005] 	 - "config/mock/v1/interfaces/tap1": (rev: 0)  loc="orchestrator/orchestrator.go(178)" logger=orchestrator.dispatcher
    DEBU[0005] 	 - "config/mock/v1/interfaces/loopback1": (rev: 0)  loc="orchestrator/orchestrator.go(178)" logger=orchestrator.dispatcher
    DEBU[0005] Resync with 2 items                           loc="orchestrator/orchestrator.go(181)" logger=orchestrator.dispatcher
    DEBU[0005] Pushing data with 2 KV pairs (source: watcher)  loc="orchestrator/dispatcher.go(67)" logger=orchestrator.dispatcher
    DEBU[0005]  - PUT: "config/mock/v1/interfaces/tap1"      loc="orchestrator/dispatcher.go(78)" logger=orchestrator.dispatcher
    DEBU[0005]  - PUT: "config/mock/v1/interfaces/loopback1"   loc="orchestrator/dispatcher.go(78)" logger=orchestrator.dispatcher
  ```

  ```
    DEBU[0012] => received CHANGE event (1 changes)          loc="orchestrator/orchestrator.go(121)" logger=orchestrator.dispatcher
    DEBU[0012] Pushing data with 1 KV pairs (source: watcher)  loc="orchestrator/dispatcher.go(67)" logger=orchestrator.dispatcher
    DEBU[0012]  - UPDATE: "config/mock/v1/interfaces/tap2"   loc="orchestrator/dispatcher.go(93)" logger=orchestrator.dispatcher
  ```

2. **transaction containing the value was triggered**, yet the value is not
   configured in SB - the issue could be one of the following:
    * the value is *pending*
      - [display graph](debugging.md#how-to-visualize-the-graph) and check
        the state of dependencies (follow black arrows coming out of the value)
      - dependency is either missing (state = `NONEXISTENT`) or in a failed
        state (state = `INVALID`/`FAILED`/`RETRYING`)
      - could be also that the plugin implementing the dependency
        [is not loaded](debugging.md#how-to-debug-agent-plugin-lookup)
        (state of a dependency = `UNIMPLEMENTED`)
      - perhaps an unintended dependency was added - double-check the
        implementation of the [Dependencies](kvdescriptor.md#dependencies) method
        of the value descriptor
    * the value is in the `UNIMPLEMENTED` state
      - with KVDB NB, could be that the suffix of the key under which the value
        was put is incorrect
         - i.e. the prefix is valid and watched by the agent, but the suffix,
           normally composed of value primary fields, is malformed and not matched
           by the descriptor's `KeySelector`
      - could be that `KeySelector` or `NBKeyPrefix` of the descriptor [do not
        use the model or use it incorrectly](#models-bad-usage)
         - e.g. `NBKeyPrefix` of this or another descriptor selects the value,
           but `KeySelector` does not
    * the value *failed* to get applied
      - [display the graph after txn](debugging.md#how-to-visualize-the-graph)
        and check the state of the value - as long as it is `FAILED`, `RETRYING`
        or `INVALID`, it cannot be assumed to be properly applied SB
      - [check for common error](#value-invalid-type) `value has invalid type for key`,
        usually caused by a mismatch between the descriptor and the model
      - could be that the set of value dependencies as listed by the descriptor
        is not actually complete (i.e. logical error) - check docs for SB
        (VPP/Linux/...) to learn if there are any additional dependencies needed
        for the value to be applied properly
    * derived value is treated as `PROPERTY` when it should have CRUD operations
      assigned
      - we do not yet provide tools to define models for derived values, therefore
        developers have to implement their own key building/parsing methods which,
        unless diligently covered by UTs, are easy to get wrong, especially
        in corner cases, and cause mismatch in the assignment of derived values
        to descriptors

### Resync triggers some operations even if the SB is in fact in-sync with NB
  - [consider running the KVScheduler in the verification mode](debugging.md#crud-verification-mode)
    to check for CRUD inconsistencies
  - descriptors of values unnecessarily updated by every resync often forget
    to consider equivalency between some attribute values inside `ValueComparator`
    \- for example, interface defined by NB with MTU 0 is supposed to be
    configured in SB with the default MTU, which, for most of the interface types,
    means that MTU 0 should be considered as equivalent to 1500 (i.e. not needing
    to trigger the `Update` operation to go from 0 to 1500 or vice-versa)
    - also, if `ValueComparator` is implemented as a separate method and not
      as a function literal inside the descriptor structure, make sure to not
      forget to plug it in via reference then

### Resync tries to create objects which already exist
  - most likely you forgot to implement the `Retrieve` method for the descriptor
    of duplicately created objects, or the [method has returned an error](#retrieve-failed):
  ```
     ERRO[0005] failed to retrieve values, refresh for the descriptor will be skipped  descriptor=mock-interface loc="kvscheduler/refresh.go(104)" logger=kvscheduler
  ```
  - make sure to avoid a common mistake of [implementing empty Retrieve method](#empty-retrieve)
    when the operation is not supported by SB

### Resync removes item not configured by NB
  - objects not configured by the agent, but instead created in SB automatically
    (e.g. default routes) should be Retrieved with [Origin](kvscheduler.md#value-origin)
    `FromSB`
  - `UnknownOrigin` can also be used and the scheduler will search through the
    history of transactions to see if the given value has been configured by the
    agent
    - defaults to `FromSB` when history is empty (first resync)

### Retrieve fails to find metadata for a dependency
  - for example, when VPP routes are being retrieved, the Route descriptor
    must also read metadata of interfaces to translate `sw_if_index` from
    the dump of routes into the logical interface names as used in NB models,
    meaning that interfaces must be dumped first to have their metadata
    up-to-date for the retrieval of routes
  - while the descriptor method `Dependencies` is used to restrict ordering of
    `Create`, `Update` and `Delete` operations between values,
    [RetrieveDependencies](kvdescriptor.md#retrievedependencies) is used to
    determine the ordering for the `Retrieve` operations between descriptors
  - if your implementation of the `Retrieve` method reads metadata of another
    descriptor, it must be mentioned inside `RetrieveDependencies`

### Value re-created when just Update should be called instead
  - check that the implementation of the `Update` method is indeed plugged into
    the descriptor structure (without `Update`, the re-creation becomes the only
    way to apply changes)
  - double-check your implementation of `UpdateWithRecreate` - perhaps you
    unintentionally requested the re-creation for the given update

### Metadata passed to `Update` or `Delete` are unexpectedly nil
  - could be that descriptor attribute `WithMetadata` is not set to `true`
    (it is not enough to just define the factory with `MetadataMapFactory`)
  - metadata for derived values are not supported (so don't expect to receive
    anything else than nil)
  - perhaps you forgot to return the new metadata in `Update` even if they have
    not changed

### Unexpected transaction plan (wrong ordering, missing operations)
  - [display the graph visualization](debugging.md#how-to-visualize-the-graph)
    and check:
    - if derived values and dependencies (relations, i.e. graph edges) are as
      expected
    - if the value states before and after the transaction are as expected
  - as a last resort, [follow the scheduler as it walks through the graph](debugging.md#graph-walk)
    during the transaction processing and try to localize the point where
    it diverges from the expected path - descriptor of the value where
    it happens is likely to have some bug(s)

## Commonly returned errors

 * <a name="value-invalid-type"></a>
   `value (...) has invalid type for key: ...` (transaction error)
    - mismatch between the proto message registered with the model and the
      value type name defined for the [descriptor adapter](kvdescriptor.md#descriptor-adapter)

 * <a name="retrieve-failed"></a>
   `failed to retrieve values, refresh for the descriptor will be skipped` (logs)
    - `Retrieve` of the given descriptor has failed and returned an error
    - the scheduler treats failed retrieval as non-fatal - the error is printed
      to the log as a warning and the graph refresh is skipped for the values
      of that particular descriptor
    - if this happens often for a given descriptor, double-check its implementation
      of the `Retrieve` operation and also make sure that `RetrieveDependencies`
      properly mentions all the dependencies

 * <a name="unimplemented-create"></a>
   `operation Create is not implemented` (transaction error)
    - descriptor of the value for which this error was returned is missing
      implementation of the `Create` method - perhaps the method is implemented,
      but it is not plugged into the descriptor structure?

 * <a name="unimplemented-delete"></a>
   `operation Delete is not implemented` (transaction error)
    - descriptor of the value for which this error was returned is missing
      implementation of the `Delete` method - perhaps the method is implemented,
      but it is not plugged into the descriptor structure?

 * <a name="descriptor-exists"></a>
  `descriptor already exist` (returned by `KVScheduler.RegisterKVDescriptor()`)
    - returned when the same descriptor is being registered more than once
    - make sure the `Name` attribute of the descriptor is unique across all
      descriptors of all initialized plugins

## <a name="programming-mistakes"></a> Common programming mistakes / bad practises

 * <a name="changing-value"></a>
   **changing the value content inside descriptor methods**
    - values should be manipulated as if they were mutable, otherwise it could
      confuse the scheduling algorithm and lead to incorrect transaction plans
    - for example, this is a bug:
``` golang
    func (d *InterfaceDescriptor) Create(key string, iface *interfaces.Interface) (metadata *ifaceidx.IfaceMetadata, err error) {
	    if iface.Mtu == 0 {
    		// MTU not set - work with the default MTU 1500
	    	iface.Mtu = 1500 // BUG - do not change content of "iface" (the value)
    	}
        //...
    }
```
   - the only exception when input argument can be edited in-place are metadata
     passed to the `Update` method, which are allowed to be re-used, e.g.:
```
    func (d *InterfaceDescriptor) Update(key string, oldIntf, newIntf *interfaces.Interface, oldMetadata *ifaceidx.IfaceMetadata) (newMetadata *ifaceidx.IfaceMetadata, err error) {
        // ...

        // update metadata
    	oldMetadata.IPAddresses = newIntf.IpAddresses
    	newMetadata = oldMetadata
    	return newMetadata, nil
    }
```

 * <a name="stateful-descriptor"></a>
   **implementing stateful descriptors**
    - descriptors are meant to be stateless and inside callbacks should operate
      only with method input arguments, as received from the scheduler
      (i.e. key, value, metadata)
    - it is still allowed (and very common) to implement CRUD operations
      as methods of a structure, but the structure should only act as a "static
      context" for the descriptor, storing for example references to the logger,
      SB handler(s), etc. - things that do not change once the descriptor is
      constructed (typically received as input arguments for the descriptor
      constructor)
    - to maintain an extra run-time data alongside values, use metadata and not
      context
    - all key-value pairs and the associated metadata are already stored
      inside the graph and fully exposed through transaction logs and REST APIs,
      therefore if descriptors do not hide any state internally, the system
      state will be fully visible from the outside and issues will be easier to
      reproduce
    - this would be considered bad practice:
``` golang
func (d *InterfaceDescriptor) Create(key string, intf *interfaces.Interface) (metadata *ifaceidx.IfaceMetadata, err error) {

	// BAD PRACTISE - do not use your own cache, instead use metadata to store
	// additional run-time data and metadata maps for lookups
	d.myCache[key] = intf
	anotherIface, anotherIfaceExists := d.myCache[<related-interace-key>]
	if anotherIfaceExists {
	    // ...
	}
    //...
}
```

 * <a name="metadata-with-derived"></a>
   **trying to use metadata with derived values**
    - it is not supported to associate metadata with a derived value - use
      the parent value instead
    - the limitation is due to the fact that derived values cannot be retrieved
      directly (and have metadata received from the `Receive` callback) - instead,
      they are derived from already retrieved parent values, which effectively
      means that they cannot carry additional state-data, beyond what is
      already included in the metadata of their parent values

 * <a name="derived-key-collision"></a>
   **deriving the same key from different values**
    - make sure to include the parent value ID inside a derived key, to ensure
      that derived values do not collide across all key-value pairs

 * <a name="models-bad-usage"></a>
   **not using models for non-derived values or mixing them with custom key
     building/parsing methods**
      - it is true that [models](kvscheduler.md#model) are
        work-in-progress and not yet supported with derived values
      - for non-derived values, however, the models are already mandatory and
        should be used to define these four descriptor fields: `NBKeyPrefix`,
        `ValueTypeName`, `KeySelector`, `KeySelector` - eventually these fields
        will all be replaced with a single `Model` reference, hence it is
        recommended to have the models prepared and already in-use for an easy
        transition
      - it is also a very bad practise to use a model only partially, e.g.:
```
    descriptor := &adapter.InterfaceDescriptor{
		Name:               InterfaceDescriptorName,
		NBKeyPrefix:        "config/vpp/v1/interfaces/", // BAD PRACTISE, use the model instead
		ValueTypeName:      interfaces.ModelInterface.ProtoName(),
		KeySelector:        interfaces.ModelInterface.IsKeyValid,
		KeyLabel:           interfaces.ModelInterface.StripKeyPrefix,
		// ...
	}
```

  * <a name="empty-descriptor-methods"></a>
    **leaving descriptor methods which are not needed defined**
      - relying too much on copy-pasting from [prepared skeletons](kvdescriptor.md#descriptor-skeletons)
        can lead to having unused callback skeleton leftovers
      - for example, if the `Update` method is not needed (update is always
        handled via full-recreation), then simply do not define the method
        instead of leaving it empty
      - descriptors are defined as structures and not as interfaces exactly
        for this reason - to allow the unused methods to remain undefined
        instead of being just empty, avoiding what would be otherwise nothing
        but a boiler-plate code

  * <a name="empty-retrieve"></a>
    **unsupported `Retrieve` defined to always return empty set of values**
      - if the `Retrieve` operation is not supported by SB for a particular
        value type, then simply leave the callback undefined in the descriptor
      - the scheduler skips refresh for values which cannot be retrieved
        (undefined `Retrieve`) and will assume that what has been set through
        previous transactions exactly corresponds with the current state of SB
      - if instead, `Retrieve` always return empty set of values, then the
        scheduler will re-Create every value defined by NB with each resync,
        thinking that they are all missing, which is likely to end with
        duplicate-value kind of errors

  * <a name="manipulating-with-derived"></a>
    **manipulating with value attributes which were derived out**
      - value attribute derived into a separate key-value pair and handled
        by CRUD operations of another descriptor, can be imagined as a slice
        of the original value that was split away - it still has an implicit
        dependency on its original value, but should no longer be considered
        as a part of it
      - for example, [if we would define a separate derived value for every
        IP address to be assigned to an interface](img/derived-interface-ip.svg),
        and there would be another descriptor which implements these assignments
        (i.e. `Create` = add IP, `Delete` = unassign IP, etc.), then the descriptor
        for interfaces should no longer:
         - consider IP addresses when comparing interfaces in `ValueComparator`
         - (un)assign IP addresses in `Create`/`Update`/`Delete`
         - consider IP addresses for interface dependencies
         - etc.

  * <a name="retrieve-derived"></a>
    **implementing Retrieve method for descriptor with only derived values
      in the scope**
      - derived values should never be Retrieved directly (returned by `Retrieve`),
        but always only returned by `DerivedValues()` of the descriptor
        with retrieved parent values

  * <a name="obtained-without-retrieve"></a>
    **not implementing Retrieve method for values announced to KVScheduler
      as `OBTAINED` via notifications**
      - it is a common mistake to forget that `OBTAINED` values also need to be
       refreshed, even though they are not touched by the resync
      - it is because NB-defined values may depend on `OBTAINED` values, and the
        scheduler therefore needs to know their state to determine if the
        dependencies are satisfied and plan the resync accordingly

  * <a name="blocking-crud"></a>
    **sleeping/blocking inside descriptor methods**
     - transaction processing is synchronous and sleeping inside a CRUD method
       would not only delay the remaining operations of the transactions,
       but other queued transactions as well
     - if a CRUD operation needs to wait for something, then express that
       "something" as a separate key-value pair and add it into the list
       of dependencies - then, when it becomes available, send notification
       using `KVScheduler.PushSBNotification()` and the scheduler will
       automatically apply pending operations which became ready for execution
     - in other words - do not hide any dependencies inside CRUD operations,
       instead use the framework to express them in a way that is visible
       to the scheduling algorithm

  * <a name="metadata-map-with-write"></a>
    **exposing metadata map with write access**
      - the KVScheduler is the owner of metadata maps, making sure they are
        always up-to-date
      - this is why custom metadata maps are not created by descriptors,
        but instead given to the scheduler in the form of factories
        (`KVDescriptor.MetadataMapFactory`)
      - maps retrieved from the scheduler using `KVScheduler.GetMetadataMap()`
        should remain read-only and exposed to other plugins as such

