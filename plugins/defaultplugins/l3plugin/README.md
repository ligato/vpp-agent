# L3 plugin

The l3plugin is a Core Agent Plugin that is designed to configure routes in the VPP. Configuration
managed by this plugin is modelled by the [proto file](model/l3/l3.proto). The configuration
must be stored in etcd using the following key:

```
/vnf-agent/<agent-label>/vpp/config/v1/vrf/0/fib/
```

An example of configuration in json format can be found [here](../../../cmd/vpp-agent-ctl/json/routes.json).

Note: Value `0` in vrfID field denotes default VRF in vpp. Since it is default value it is omitted in the config above.
 If you want to configure a route for a VRF other than default, make sure that the VRF has already been created.

To insert config into etcd in json format [vpp-agent-ctl](../../../cmd/vpp-agent-ctl/main.go) can be used.
We assume that we want to configure vpp with label `vpp1` and config is stored in the `routes.json` file
```
vpp-agent-ctl -put "/vnf-agent/vpp1/vpp/config/v1/vrf/0/fib" json/routes.json
```

The vpp-agent-ctl contains a simple predefined route config also. It can be used for testing purposes.
To setup the predefined route config run:
```
vpp-agent-ctl -cr
```
To remove it run:
```
vpp-agent-ctl -dr
```
