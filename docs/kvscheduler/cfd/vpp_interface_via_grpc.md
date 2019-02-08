# Control Flow Diagrams

## Scenario: Create VPP interface via GRPC

Variant of [this][vpp-interface] example, with GRPC used as the NB interface
for the agent instead of KVDB. The transaction control-flow is collapsed, however,
since there are actually no differences between the two cases. For KVScheduler
it is completely irrelevant how the desired configuration gets conveyed
into the agent. The advantage of GRPC over KVDB is that the transaction error
value gets propagated back to the client, which is then able to react to it
accordingly. On the other hand, with KVDB the client does not have to maintain
a connection with the agent and the configuration can be submitted even when
the agent is restarting or not running at all.


[vpp-interface]: vpp_interface.md
![CFD](https://raw.githubusercontent.com/milanlenco/vpp-agent/kvs-docs/docs/kvscheduler/cfd/uml/add_interface_grpc_txn_collapsed.svg?sanitize=true)