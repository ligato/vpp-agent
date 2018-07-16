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
curl http://0.0.0.0:9191/acl/interface/<if-sw-index>    // Get ACLs for interface <if-sw-index>
curl http://0.0.0.0:9191/acl/ip                         // Get all IP ACLs  
curl http://0.0.0.0:9191/acl/ip/example                 // Get an example ACL 
```

Configure an IP ACL:
```
curl -H "Content-Type: application/json" -X POST -d '<acl-json>' http://0.0.0.0:9191/interface/acl/ip
```

For example:
```
curl -H "Content-Type: application/json" -X POST -d '{
    "acl_name": "example"
    "rules": [
        {
            "rule_name": "acl1_rule1",                                                                                                        {                                                                                                                     "actions": {
            "acl_action": 1,
            "match": {
                "ip_rule": {
                    "ip": {
                        "destination_network": "1.2.3.4/24",
                        "source_network": "5.6.7.8/24"
                    },
                    "tcp": {
                        "destination_port_range": {
                            "lower_port": 80,
                            "upper_port": 8080
                        },
                        "source_port_range": {
                            "lower_port": 10,
                            "upper_port": 1010
                        },
                        "tcp_flags_mask": 255,
                        "tcp_flags_value": 9
                    }
                }
            }
        }
    ]
}' http://0.0.0.0:9191/interface/acl/ip
```
## Logging mechanism
The REST API request is logged to stdout. The log contains VPPCLI command and VPPCLI response. It is searchable in elastic search using "VPPCLI".