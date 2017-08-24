# Etcd plugin

The Etcd plugin provides access to an etcd key-value data store.

**API**

Provides API described in the [skeleton](../plugin), the plugin is 
documented at the bottom of the [doc.go](doc.go) file.

**Configuration**

- Location of the Etcd configuration file can be defined either by the 
  command line flag `etcdv3-config` or set via the `ETCDV3_CONFIG` 
  environment variable.

**Status Check**

- Etcd plugin has a mechanism to periodically check the status of an 
  Etcd connection.  
