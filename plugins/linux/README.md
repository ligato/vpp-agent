# Linux Plugin

The `linuxplugin` is a core Agent Plugin for the management of a subset of the Linux
network configuration. Configuration of VETH (virtual ethernet pair) interfaces, linux routes and ARP entries
is currently supported. Detailed description can be found in particular READMEs:
 - [ifplugin](ifplugin)
 - [l3plugin](l3plugin)
 - [nsplugin](nsplugin)
 
In general, the northbound configuration is translated to a sequence of Netlink API
calls (using `github.com/vishvananda/netlink` and `github.com/vishvananda/netns` libraries).

## Config file

*Stopwatch*

Duration of the linux netlink procedure can be measured using stopwatch feature. These data are logged after 
every event(any resync, interfaces, routes, etc.). Enable stopwatch in linux.conf: 

`stopwatch: true` or  `stopwatch: false`

Stopwatch is disabled by default (if there is no config available). 