# SR plugin

The `srplugin` is a Core Agent Plugin designed to configure Segment routing for IPv6 (SRv6) in the VPP.
Configuration managed by this plugin is modelled by [srv6 proto file](../../..api/models/vpp/srv6/srv6.proto).

All configuration must be stored in ETCD using the srv6 key prefix:
 
```
/vnf-agent/<agent-label>/config/vpp/srv6/v2/
```

## Configuring Local SIDs
The local SID can be configured using this key:
```
/vnf-agent/<agent-label>/config/vpp/srv6/v2/localsid/<SID>
```
where ```<SID>``` (Segment ID) is a unique ID of local sid and it must be an valid IPv6 address. The SID in NB key must
be the same as in the json configuration (value of NB key-value pair). 

## Configuring Policy
The segment routing policy can be configured using this key:
```
/vnf-agent/<agent-label>/config/vpp/srv6/v2/policy/<bsid>
```
where ```<bsid>``` is  unique binding SID of the policy. As any other SRv6 SID it must be an valid IPv6 address. Also 
the SID in NB key must be the same as in the json configuration (value of NB key-value pair).\
The policy can have defined inside value multiple segment lists. The VPP implementation doesn't allow to have policy 
without at least one segment list. Therefore inserting(updating with) policy that has not defined at least one segment 
list will fail (value can be written in ETCD, but its application to VPP will fail as validation error).

## Configuring Steering
The steering (the VPP's policy for steering traffic into SR policy) can be configured using this key:
```
/vnf-agent/<agent-label>/config/vpp/srv6/v2/steering/<name>
```
where ```<name>``` is a unique name of steering.

