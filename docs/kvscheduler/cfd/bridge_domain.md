# Control Flow Diagrams

## Example: Bridge Domain

Using bridge domain it is demonstrated how derived values can be used to
"break" item into multiple parts with their own CRUD operations and dependencies.

Bridge domain groups multiple interfaces to share the same flooding or broadcast
characteristics. Empty bridge domain has no dependencies and can be created
independently from interfaces. But to put an interface into a bridge domain,
both the interface and the domain must be created first. One solution for
the KVScheduler framework would be to handle bridge domain as a single key-value
pair depending on all the interfaces it is configured to contain. But this is
a rather coarse-grained approach that would prevent the existence of the bridge
domain even when only a single interface is missing. Moreover, with KVDB,
request to remove interface could overtake update of the bridge domain
configuration un-listing the interface, which would cause the bridge domain
to be temporarily removed and shortly afterwards fully re-created.

The concept of derived values allowed to specify binding between bridge
domain and every bridged interface as a separate derived value, handled
by its own `BDInterfaceDescriptor` descriptor, where `Create()` operation puts
interface into the bridge domain, `Delete()` breaks the binding, etc.
The bridge domain itself has no dependencies and will be configured as long as
it is demanded by NB.
The bindings, however, will each have a dependency on its associated interface
(and implicitly on the bridge domain it is derived from).
Even if one or more interface are missing or are being deleted, the remaining
of the bridge domain will remain unaffected and continuously functional.

The control-flow diagram shows that bridge domain is created even if the
interface that it is supposed to contain gets configured later. The binding
remains in the `PENDING` state until the interface is configured.
  

![CFD](https://raw.githubusercontent.com/milanlenco/vpp-agent/kvs-docs/docs/kvscheduler/cfd/uml/add_bd_before_interface.svg?sanitize=true)

 