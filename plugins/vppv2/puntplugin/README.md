# Punt plugin

Punt plugin defines [model](../model/punt/punt.proto) and supports three scenarios:

1. **IP punt redirect:** packet matching one of the VPP IP addresses but otherwise would be dropped will be send to 
the defined next hop IP address via TX interface. The rule can be enhanced and input traffic filtered with RX 
interface - redirect applies only to traffic received via RX.

2. **Punt to host:** packet matching one of the VPP IP address and also defined L3 protocol, L4 protocol and also port
is punted to host

3. **Punt to host via socket** packet has to match all the criteria as in previous case, but packet is punted to 
unix domain socket

In order to use punt to socket, VPP has to be started with configuration:

```
punt {
  socket /path/to/socket/file
}
```

Otherwise VPP-agent returns error during setup.

### Limitations:
- only one socket path can be defined for the VPP
- the path has to match with the configuration in proto file
- punt to host entries cannot be removed with current VPP version 
- punt configuration cannot be dumped in current VPP version