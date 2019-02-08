# Control Flow Diagrams

## Example: Transaction revert

An update transaction (i.e. not resync) can be configured to run in either
[best-effort mode][retry_failed_operation.md], allowing partial completion,
or to terminate upon first failure and revert already applied changes so that no
visible effects are left in the system (i.e. the true transaction definition).
This behaviour is only supported with GRPC or localclient as the agent NB
interface. With KVDB, the agent will run in the best-effort mode to get as close
to the desired configuration as it is possible.

In this example, a transaction is planned and executed to create a VPP interface
`my-tap` with an attached route `my-route`. While the interface is successfully
created, the route, on the other hand, fails to get configured. The scheduler
then triggers the *revert procedure*. First the current value of `my-route` is
retrieved, the state of which cannot be assumed since the creation failed
somewhere in-progress. The route is indeed not found to be configured, therefore
only the interface must be deleted to undo already executed changes. Once
the interface is removed, the state of the system is back to where it was before
the transaction started. Finally, the transaction error is returned back to the
northbound plane.


![CFD](https://raw.githubusercontent.com/milanlenco/vpp-agent/kvs-docs/docs/kvscheduler/cfd/uml/interface_and_route_with_revert.svg?sanitize=true)