# Etcd plugin

Etcd plugin provides access to etcd key-value data store.

**API**

Provides API described in the [skeleton](../plugin), the plugin is documented at the bottom of [doc.go](doc.go).

**Configuration**

- Location of the Etcd configuration file can be defined either by command line flag `etcdv3-config` or 
set via `ETCDV3_CONFIG` environment variable.

**Status Check**

- Etcd plugin has a mechanism to periodically check a status of the Etcd connection.  

**Dependencies**
- [Logging](../../../logging/plugin)
- [ServiceLabel](../../../servicelabel)