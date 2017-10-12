# Default plugins
 
 Default plugins manage basic configuration of VPP. The management of configuration is split among multiple
 packages. Detailed description can be found in particular READMEs:
 - [ifplugin](ifplugin)
 - [l2plugin](l2plugin)
 - [l3plugin](l3plugin)
 - [aclplugin](aclplugin)
 
# Config file 

 The default plugins can use configuration file `defaultplugins.conf` to:
  * set global maximum transmission unit 
  * enable/disable stopwatch
  * set VPP resync strategy
  
  To run the vpp-agent with defaultplugins.conf:
   
   `vpp-agent --defaultplugins-config=/opt/vpp-agent/dev/defaultplugins.conf`
  
 *MTU*
 
 MTU is used in interface plugin. The value is preferred before global setting which is set to 9000 bytes. Mtu is 
 written in config as follows:
 
 `mtu: <value>`
 
 *Stopwatch*
 
 Duration of VPP binary api call during resync can be measured using stopwatch. These data are then logged after 
 every partial resync (interfaces, bridge domains, fib entries etc.). Enable stopwatch in defaultplugins.conf: 
 
  `stopwatch: true` or  `stopwatch: false`
  
 Stopwatch is disabled by default (if there is no config available). 
 
 *Strategy*
 
 There is several strategies available for VPP resync:
 * **default** always performs the full resync of all VPP plugins. This is the default strategy if none is set. 
 * **interface-based** evaluates the existing configuration in the VPP at first. The state of interfaces is the 
 decision-maker: if there is any interface configured except local0, the resync is performed normally. Otherwise 
 it is skipped. Use it carefully because this strategy does not take into consideration the state of the etcd.
 * **skip** always omits the VPP resync 
 
 Strategy can be set in defaultplugins.conf:
 
 `strategy: default` or  `strategy: interface-based` or `strategy: skip`
