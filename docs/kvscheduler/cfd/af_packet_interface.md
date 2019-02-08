# Control Flow Diagrams

## Example: AF-Packet interface

`AF-Packet` is VPP interface type attached to a host OS interface, capturing
all its incoming traffic and also allowing to inject Tx packets through a special
type of socket.

The requirement is that the host OS interface exists already before the `AF-Packet`
interface gets created. The challenge is that the host interface may not be
from the scope of items configured by the agent. Instead, it could be
a pre-existing physical device or interface created by an external process or
an administrator during the agent run-time. In such cases, however, there would
be no key-value pair to reference from within `AF-Packet` dependencies. Therefore,
KVScheduler allows to notify about external objects through 
`PushSBNotification(key, value, metadata)` method. Values received through
notifications are denoted as `OBTAINED` and will not be removed by resync even
though they are not requested to be configured by NB. Obtained values are
allowed to have their own descriptors, but from the CRUD operations only
`Retrieve()` is ever called to refresh the graph. `Create`, `Delete` and `Update`
are never used, since obtained values are updated externally and the agent is
only notified about the changes *after* they has already happened.

Linux interface plugin ships with `InterfaceWatcher` descriptor, which retrieves
and notifies about Linux interface in the network namespace of the agent
(so-called default network namespace). Linux interfaces are assigned unique
keys using their host names: `linux/interface/host-name/eth1`
The `AF-Packet` interface then defines dependency referencing the key with the
host name of the interface it is supposed to attach to (cannot attach
to interfaces from other namespaces).

In this example, the host interface gets created after the request to configure
`AF-Packet` is received. Therefore, the scheduler keeps the `AF-Packet` in the
`PENDING` state until the notification is received. 


![CFD](https://raw.githubusercontent.com/milanlenco/vpp-agent/kvs-docs/docs/kvscheduler/cfd/uml/add_af_packet_interface.svg?sanitize=true)

 