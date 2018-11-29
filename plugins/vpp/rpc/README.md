# RPC plugin

The RPC plugin defines services required for resync, put or delete events via gRPC. Services are defined 
in [model](../model/rpc/rpc.proto)

## Data persistence

RPC plugin support data persistence - all resynced, created or deleted data can be mirrored to desired
database. 
Note: data are not read during resync.

To define persistence database, config file can be used and attached via flag:

```
-vpp-grpc-config=<path>
```

The config file contains one filed which defines persistence database:

```
persistence-db: bolt
```

Only one database can be defined at a time. Database plugin has to be loaded via appropriate 
config file as well.