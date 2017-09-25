# Etcd plugin

The Etcd plugin provides access to an etcd key-value data store.

**API**

Implements API described in the [skeleton](../plugin):
The plugin is documented in more detail in the [doc.go](doc.go) file.

**Configuration**

- Location of the Etcd configuration file can be defined either by the 
  command line flag `etcdv3-config` or set via the `ETCDV3_CONFIG` 
  environment variable.

**Status Check**

- If injected, Etcd plugin will use StatusCheck plugin to periodically
  issue a minimalistic GET request to check for the status
  of the connection.
  The etcd connection state affects the global status of the agent.
  If agent cannot establish connection with etcd, both the readiness
  and the liveness probe from the [probe plugin](../../../health/probe)
  will return a negative result (accessible only via REST API in such
  case).