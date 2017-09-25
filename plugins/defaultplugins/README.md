# Default plugins
 
 Default plugins manage basic configuration of VPP. The management of configuration is split among multiple
 packages. Detailed description can be found in particular READMEs:
 - [ifplugin](ifplugin)
 - [l2plugin](l2plugin)
 - [l3plugin](l3plugin)
 - [aclplugin](aclplugin)
 
# Config file 

 The default plugins can use configuration file `defaultplugins.conf` to set global maximum transmission unit value
 used in interface plugin. This mtu value is preferred before global setting which is set to 9000 bytes. Mtu is 
 written in config as follows:
 
 `mtu: <value>`
 
 To run the vpp-agent with defaultplugins.conf:
 
 `vpp-agent --defaultplugins-config=/opt/vpp-agent/dev/defaultplugins.conf`
 
 
