# REST API Plugin

The `restplugin` is a core Agent Plugin used of exposing REST API for the following:
* Run VPP CLI commands
* Exposes existing Northbound objects
* Provides logging mechanism so that the VPPCLI command and response can be searched in elastic search

## VPP CLI commands
```
curl -H "Content-Type: application/json" -X POST -d '{"vppclicommand":"show interface"}' http://0.0.0.0:9191/
```

## Exposing existing Northbound objects
```
curl http://0.0.0.0:9191/interfaces
curl http://0.0.0.0:9191/bridgedomainids
curl http://0.0.0.0:9191/bridgedomains
curl http://0.0.0.0:9191/fibs
curl http://0.0.0.0:9191/xconnectpairs
curl http://0.0.0.0:9191/staticroutes

curl -H "Content-Type: application/json" -X POST -d '{"swIndex":"0"}' http://0.0.0.0:9191/interface/acl
```
## Logging mechanism
The REST API request is logged to stdout. The log contains VPPCLI command and VPPCLI response. It is searchable in elastic search using "VPPCLI".