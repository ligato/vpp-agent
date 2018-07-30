# REST API Plugin

The `restplugin` is a core Agent Plugin used to expose REST API for the following:
* Run VPP CLI commands
* Exposes existing Northbound objects
* Provides logging mechanism so that the VPPCLI command and response can be searched in elastic search

## VPP CLI commands
```
curl -H "Content-Type: application/json" -X POST -d '{"vppclicommand":"show interface"}' http://0.0.0.0:9191/
```

## Exposing existing Northbound objects

Here is the list of supported REST URLs. If configuration dump URL is used, the output is based on proto model
structure for given data type together with VPP-specific data which are not a part of the model (indexes for
interfaces or ACLs, internal names, etc.). Those data are in separate section labeled as `Meta`.

**Access lists**

URLs to obtain ACL IP/MACIP configuration are as follows.

```
curl http://0.0.0.0:9191/vpp/dump/v1/acl/ip
curl http://0.0.0.0:9191/vpp/dump/v1/acl/macip 
```

**Interfaces**

REST plugin exposes configured interfaces, which can be show all together, or only interfaces
of specific type.
 
```
curl http://0.0.0.0:9191/vpp/dump/v1/interfaces
curl http://0.0.0.0:9191/vpp/dump/v1/interfaces/loopback
curl http://0.0.0.0:9191/vpp/dump/v1/interfaces/ethernet
curl http://0.0.0.0:9191/vpp/dump/v1/interfaces/memif
curl http://0.0.0.0:9191/vpp/dump/v1/interfaces/tap
curl http://0.0.0.0:9191/vpp/dump/v1/interfaces/vxlan
curl http://0.0.0.0:9191/vpp/dump/v1/interfaces/afpacket
``` 
 
**BFD**

REST plugin allows to dump bidirectional forwarding detection sessions, authentication keys, 
or the whole configuration. 

```
curl http://0.0.0.0:9191/vpp/dump/v1/bfd
curl http://0.0.0.0:9191/vpp/dump/v1/bfd/sessions
curl http://0.0.0.0:9191/vpp/dump/v1/bfd/authkeys
``` 

**L2 plugin**

Support for bridge domains, FIBs and cross connects. It is also possible to get all 
the bridge domain IDs.

```
curl http://0.0.0.0:9191/vpp/dump/v1/bdid
curl http://0.0.0.0:9191/vpp/dump/v1/bd
curl http://0.0.0.0:9191/vpp/dump/v1/fibs
curl http://0.0.0.0:9191/vpp/dump/v1/xc
```

**L3 plugin**

ARPs and static routes exposed via REST:

```
curl http://0.0.0.0:9191/vpp/dump/v1/routes
curl http://0.0.0.0:9191/arps
```

Configure an IP ACL:
```
curl -H "Content-Type: application/json" -X POST -d '<acl-json>' http://0.0.0.0:9191/interface/acl/ip
```

## Logging mechanism
The REST API request is logged to stdout. The log contains VPPCLI command and VPPCLI response. It is searchable in elastic search using "VPPCLI".