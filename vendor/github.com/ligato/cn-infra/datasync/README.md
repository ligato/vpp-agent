# Fault tolerance vs. data synchronization
In a fault-tolerant solution, there is a need to recover from faults. This plugin helps to solve the
data resynchronization (data RESYNC) aspect of fault recovery.
When the Agent looses connectivity to a northbound client (ETCD, Kafka...) or VPP, it needs to recover from that fault:
1. When connectivity is reestablished after a failure, the agent needs to resynchronize configuration from a northbound 
   client with configuration in the VPP.
2. Sometimes it is easier to use "brute force" and restart the container (both VPP and the agent) 
   and skip the resynchronization.

# Responsibility of plugins
Each plugin is responsible for its own part of configuration data received from northbound clients. Each plugin needs 
to be decoupled from a particular datasync transport (ETCD, GRPC, REST...)
The data of one plugin can have references to data of another plugin. Therefore, we need 
to orchestrate the right order of data resynchronization between plugins.

The datasync plugin helps other plugins to:
 1. Determine the right time to resynchronize their data (event will be send on a dedicated resync channel)
 2. Report a fault/error occurred and that the resync plugin should start data resynchronization. 
 3. be decoupled from a particular datasync transport (ETCD, GRPC, REST ...). 
    Every other plugin receives configuration data only through GO interfaces defined in datasync_api.go

# Example Workflow (TODO move to VPP agent repo...)

Note, about other events: a/ data change events, b/ IDX changes. Data change events are supposed to be buffered
during data resync (by the plugin transport). Buffered changes need to be filtered if they were not already processed
during revision (see the dbmux plugin for more details). The IDX changes will be delivered during overall resync
process. Let's explain it on this example:
1. given network interfaces and BD's configured in the VPP
2. resync is started
3. first plugin finds in the resync process that particular network interface needs to be deleted
4. this first plugin notifies that it is going to delete network interface (using IDX API)
5. second plugin handles the notification and adapts the BD configuration in the VPP
6. then second plugin callbacks that it is Done()
7. the first plugin finally deletes the network interface