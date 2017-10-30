# ACL plugin

The `aclplugin` is a Core Agent Plugin designed to configure ACL in the VPP.
Configuration managed by this plugin is modelled by [acl proto file](model/acl/acl.proto).

Model allows to define configuration for the agent:
 - ACLs
 - Interfaces referencing those ACLs

The configuration must be stored in ETCD using following keys:

```
/vnf-agent/<agent-label>/vpp/config/v1/acl/<acl-name>
```
