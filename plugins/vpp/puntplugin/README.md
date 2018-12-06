# Punt plugin

The punt plugin is a core VPP-Agent Plugin that is designed to register punt configuration entries
allowing to punt packets to host via unix domain socket.
Configuration managed by this plugin is modelled by the [proto file](../model/punt/punt.proto). 
The configuration has to be stored in etcd using the following key:

```
/vnf-agent/<agent-label>/vpp/config/v1/punt/<name>

```

An example of configuration in json format can be found [here](../../../cmd/vpp-agent-ctl/json/punt-socket-register.json).

Note: the punt socket path needs to be defined in the VPP startup config. The VPP currently supports 
only one unix domain socket path. Example of startup config entry:

```
punt {
  socket /tmp/socket/punt
}
```

To insert config into etcd in json format [vpp-agent-ctl](../../../cmd/vpp-agent-ctl) 
can be used. We assume that we want to configure vpp with label `vpp1` and config is stored 
in the `punt-socket-register.json.json` file. 
```
vpp-agent-ctl -put "/vnf-agent/vpp1/vpp/config/v1/punt/punt1" json/punt-socket-register.json
```

The vpp-agent-ctl also contains a simple predefined punt config, suitable for testing purposes. 
To register punt socket, run:
```
vpp-agent-ctl -puntr
```

To deregister it:
```
vpp-agent-ctl -puntd
```

Note: registered entries currently cannot be dumped or shown via VPP CLI (Missing support in VPP).
