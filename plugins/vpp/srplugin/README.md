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
The policy can have multiple policy segment lists(each policy segment list defines one segment routing path and each 
policy segment list has its own weight). 
It can be configured using this key:
```
/vnf-agent/<agent-label>/config/vpp/srv6/v2/policysegmentlist/<name>/policy/<bsid> 
```
where ```name``` is a unique string name of the segment (unique within policy) and ```<bsid>``` is the binding SID of 
policy to which segment belongs. Binding SID from key must be the same as described in json configuration used as value for the key.

The VPP implementation doesn't allow to have empty segment routing policy (policy must have always at least one policy 
segment list). Therefore adding the policy configuration without at least one segment won't write into the VPP anything. 
The configuration of VPP is postponed until the first policy segment is configured.

WARNING! Removal of the LAST policy segment list will trigger also delete of the parent policy (and cascade delete of steering).
Prefer recreation of policy segments lists by adding policy segments first and then removing the old ones. This way you 
can't accidently delete the whole policy. This restriction is the result of a trade-off between API simplicity and possible 
implementations handling mentioned VPP restriction.
     
It is also possible to remove only the policy and the VPP will be configured to remove the policy with all its segments. 
Keep in mind that in this case the policy segment lists stay in key-value store (i.e. ETCD) and can be attached to later 
added policy with the same bsid. So correct way is to remove also policy segment lists.   


## Configuring Steering
The steering (the VPP's policy for steering traffic into SR policy) can be configured using this key:
```
/vnf-agent/<agent-label>/config/vpp/srv6/v2/steering/<name>
```
where ```<name>``` is a unique name of steering.

