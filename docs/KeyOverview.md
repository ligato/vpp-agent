GENERATED FILE, DO NOT EDIT BY HAND

This page is an overview of all keys supported for the VPP-Agent

# Key overview

- [VPP keys](#vpp)
- [Linux keys](#linux)

Parts of the key in `<>` must be set with the same value as in a model. The microservice label is set to `vpp1` in every mentioned key, but if different value is used, it needs to be replaced in the key as well.

Link in key title redirects to the associated proto definition.

### <a name="vpp">VPP keys</a>

**[ACL:](https://github.com/ligato/vpp-agent/blob/master/api/models/vpp/acls/acl.proto)**
```
/vnf-agent/<ms-label>/config/vpp/acls/v2/acl
```

**[ARPEntry:](https://github.com/ligato/vpp-agent/blob/master/api/models/vpp/arp/arp.proto)**
```
/vnf-agent/<ms-label>/config/vpp/v2/arp/<interface>/<ip-address>
```

**[BridgeDomain:](https://github.com/ligato/vpp-agent/blob/master/api/models/vpp/l2/bridge-domain.proto)**
```
/vnf-agent/<ms-label>/config/vpp/l2/v2/bridge-domain
```

**[DNat44:](https://github.com/ligato/vpp-agent/blob/master/api/models/vpp/nat/nat.proto)**
```
/vnf-agent/<ms-label>/config/vpp/nat/v2/dnat44/<label>
```

**[FIBEntry:](https://github.com/ligato/vpp-agent/blob/master/api/models/vpp/l2/fib.proto)**
```
/vnf-agent/<ms-label>/config/vpp/l2/v2/fib/<bridge-domain>/mac/<phys-address>
```

**[Interface:](https://github.com/ligato/vpp-agent/blob/master/api/models/vpp/interfaces/interface.proto)**
```
/vnf-agent/<ms-label>/config/vpp/v2/interfaces
```

**[IPRedirect:](https://github.com/ligato/vpp-agent/blob/master/api/models/vpp/ipredirect/punt.proto)**
```
/vnf-agent/<ms-label>/config/vpp/v2/ipredirect/l3/<l3-protocol>/tx/<tx-interface>
```

**[IPScanNeighbor:](https://github.com/ligato/vpp-agent/blob/master/api/models/vpp/ipscanneigh-global/l3.proto)**
```
/vnf-agent/<ms-label>/config/vpp/v2/ipscanneigh-global
```

**[Nat44Global:](https://github.com/ligato/vpp-agent/blob/master/api/models/vpp/nat/nat.proto)**
```
/vnf-agent/<ms-label>/config/vpp/nat/v2/nat44-global
```

**[ProxyARP:](https://github.com/ligato/vpp-agent/blob/master/api/models/vpp/proxyarp-global/l3.proto)**
```
/vnf-agent/<ms-label>/config/vpp/v2/proxyarp-global
```

**[Route:](https://github.com/ligato/vpp-agent/blob/master/api/models/vpp/route/route.proto)**
```
/vnf-agent/<ms-label>/config/vpp/v2/route/vrf/<vrf-id>/dst/<dst-network>/gw/<next-hop-addr>
```

**[Rule:](https://github.com/ligato/vpp-agent/blob/master/api/models/vpp/stn/stn.proto)**
```
/vnf-agent/<ms-label>/config/vpp/stn/v2/rule/<interface>/ip/<ip-address>
```

**[SecurityAssociation:](https://github.com/ligato/vpp-agent/blob/master/api/models/vpp/ipsec/ipsec.proto)**
```
/vnf-agent/<ms-label>/config/vpp/ipsec/v2/sa/<index>
```

**[SecurityPolicyDatabase:](https://github.com/ligato/vpp-agent/blob/master/api/models/vpp/ipsec/ipsec.proto)**
```
/vnf-agent/<ms-label>/config/vpp/ipsec/v2/spd/<index>
```

**[ToHost:](https://github.com/ligato/vpp-agent/blob/master/api/models/vpp/tohost/punt.proto)**
```
/vnf-agent/<ms-label>/config/vpp/v2/tohost/l3/<l3-protocol>/l4/<l4-protocol>/port/<port>
```

**[XConnectPair:](https://github.com/ligato/vpp-agent/blob/master/api/models/vpp/l2/xconnect.proto)**
```
/vnf-agent/<ms-label>/config/vpp/l2/v2/xconnect/<receive-interface>
```

### <a name="linux">Linux keys</a>

**[ARPEntry:](https://github.com/ligato/vpp-agent/blob/master/api/models/linux/l3/arp.proto)**
```
/vnf-agent/<ms-label>/config/linux/l3/v2/arp/<interface>/<ip-address>
```

**[Interface:](https://github.com/ligato/vpp-agent/blob/master/api/models/linux/interfaces/interface.proto)**
```
/vnf-agent/<ms-label>/config/linux/interfaces/v2/interface
```

**[Route:](https://github.com/ligato/vpp-agent/blob/master/api/models/linux/l3/route.proto)**
```
/vnf-agent/<ms-label>/config/linux/l3/v2/route/<dst-network>/<outgoing-interface>
```

