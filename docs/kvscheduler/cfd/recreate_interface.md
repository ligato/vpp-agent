# Control Flow Diagrams

## Example: Interface Re-creation

The example uses VPP interface with attached route to outline the control-flow
of item re-creation. A specific configuration update may not be supported by SB
to perform incrementally - instead the given item may need to be deleted and
re-created with the new configuration. Using `UpdateWithRecreate()` method,
a descriptor is able to tell the KVScheduler if the given item requires
full re-creation for the configuration update to be applied.

This example demonstrates re-creation using a VPP TAP interface and a NB request
to change the RX ring size, which is not supported for an already created
interface. Furthermore, the interface has an L3 route attached to it. The route
cannot exists without the interface, therefore it must be deleted and moved into
the `PENDING` state before interface re-creation, and configured back again once
the re-creation procedure has finalized.


![CFD](https://raw.githubusercontent.com/milanlenco/vpp-agent/kvs-docs/docs/kvscheduler/cfd/uml/recreate_interface_with_route.svg?sanitize=true)