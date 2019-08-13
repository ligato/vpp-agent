# Combining etcd with localclient for NB config transport

The example shows how to use vpp-agent to create configuration items internally
via localclient, but at the same time also receive additional configuration items
from etcd.
The orchestrator plugin will ensure that the configuration from both of these
sources is merged during (second) resync.

How to run example:
1. Start etcd datastore:
```
./run_etcd.sh
```
2. Run VPP of suitable version with default configuration
3. Run the agent (localclient will resync):
```
./run_agent.sh
```
4. Send configuration requests via etcd:
```
./apply_config_via_etcd.sh
```
5. Restart agent to test that (eventually) the configuration from both sources
is maintained and kept in-sync on VPP
   - **note**: it is expected that localclient resync will first undo changes
     requested via etcd - the second resync, triggered by kvdbsync plugin,
     will restore this configuration