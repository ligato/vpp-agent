# Control Flow Diagrams

This guide aims to explain the key-value scheduling algorithm using examples
with UML control-flow diagrams, each covering a specific scenario using real
configuration items from vpp and linux plugins of the agent. The control-flow
diagrams are structured to describe all the interactions between KVScheduler, NB
plane and KVDescriptor-s during transactions. To improve readability, the examples
use shortened keys (without prefixes) or even alternative and more descriptive
aliases as object identifiers. For example, `my-route` is used as a placeholder
for user-defined route, which otherwise would be identified by key composed
of destination network, outgoing interface and next hop address. Moreover, most
of the configuration items that are automatically pre-created in the SB plane
(i.e. `OBTAINED`), normally retrieved from VPP and Linux during the first resync
(e.g. default routes, physical interfaces, etc.), are omitted from the diagrams
since they do not play any role in the presented scenarios.

The UML diagrams are plotted as SVG images, also containing links to images
presenting the state of the graph with values at the end of every transaction.
But to be able to access these links, the UML diagrams have to be clicked at
to open them as standalone inside another web browser tab. From within github
web-UI the links are not directly accessible.

## Index

* [Create single interface (KVDB NB)][create-interface-kvdb]
  - the most basic (and also verbose) example, outlining all the interactions
    between KVScheduler, descriptors, models and NB (KVDB) needed to create
    a single value (VPP interface in this case) with a derived value but without
    any dependencies
* [Create single interface (GRPC NB)][create-interface-grpc]
  - variant of the example above, where NB is GRPC instead of KVDB
  - the transaction control-flow is collapsed, however, since there are actually
    no differences (for KVScheduler it is irrelevant how the desired
    configuration gets conveyed to the agent)
* [Interface with L3 route][interface-and-route]
  - this example shows how route, received from KVDB before the associated
    interface, gets delayed and stays in the `PENDING` state until the interface
    configuration is received and applied
* [Bridge domain][bridge-domain]
  - using bridge domain it is demonstrated how derived values can be used to
    "break" item into multiple parts with their own CRUD operations and
    dependencies
  - control flow diagram shows how and when the derived values get processed
* [Unnumbered interface][unnumbered-interface]
  - with unnumbered interface it is shown that derived values themselves can be
    a target of dependencies
* [Interface re-creation][recreate-interface]
  - the example outlines the control-flow of item re-creation (update which
    cannot be applied incrementally, but requires the target item to be removed
    first and then created with the new configuration)
* [Retry of failed operation][retry-failed-ops]
  - example of a failed (best-effort) transaction, fixed by a subsequent Retry
* [Revert of failed transaction][revert-failed-txn]
  - example of a failed transaction, with already executed changes fully reverted
    to leave no visible effects
* [AF-Packet interface][af-packet-interface]
  - this example describes control-flow of creating an AF-Packet interface
    depended on a host interface, presence of which is announced to KVScheduler
    via notifications


[create-interface-kvdb]: cfd/vpp_interface.md
[create-interface-grpc]: cfd/vpp_interface_via_grpc.md
[interface-and-route]: cfd/vpp_ip_route.md
[bridge-domain]: cfd/bridge_domain.md
[unnumbered-interface]: cfd/unnumbered_interface.md
[recreate-interface]: cfd/recreate_interface.md
[retry-failed-ops]: cfd/retry_failed_operation.md
[revert-failed-txn]: cfd/revert_failed_txn.md
[af-packet-interface]: cfd/af_packet_interface.md
