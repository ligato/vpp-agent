# State of implementation of rx-mode for various interface types

| interface type | rx-modes | implemented | how to check on VPP | example of creation of interface |
| ---- | ---- | ---- | ---- | ---- |
| tap interface | PIA | yes  | ? | _#tap connect tap1_|
| memory interface |  PIA | yes | both sides of memif (slave and master) has to be configured = 2 VPPs.</br>_#sh memif_ | _#create memif master_ |
| vxlan tunnel | PIA | yes | ? | #_create vxlan tunnel src 192.168.168.168 dst 192.168.168.170 vni 40_
| software loopback | PIA | yes | ? | _#create loopback interface_
| ethernet csmad | P | yes | _#show interface rx-placement_ | vpp will adopt interfaces on start up
| af packet | PIA | yes | _#show interface rx-placement_ | _#create host-interface name <ifname>_

Legend:

- P - polling
- I - interrupt
- A - adaptive
