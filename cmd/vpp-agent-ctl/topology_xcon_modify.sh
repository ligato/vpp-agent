#!/usr/bin/env bash

VSWITCH_NAME="vpp1"
RNG_NAME="rng-vpp"
USSCHED_NAME="ussched-vpp"
VNF_NAME="vnf-vpp"

# VSWITCH - change IP & MAC of the loopback interface
vpp-agent-ctl -put /vnf-agent/${VSWITCH_NAME}/vpp/config/v1/interface/loop1 - << EOF
{
  "name": "loop1",
  "enabled": true,
  "phys_address": "8a:f1:be:90:00:bb",
  "mtu": 1500,
  "ip_addresses": [
    "6.0.0.101/24"
  ]
}
EOF

# VSWITCH - delete memif master to RNG
vpp-agent-ctl -del /vnf-agent/${VSWITCH_NAME}/vpp/config/v1/interface/memif-to-rng

# RNG - delete memif slave to VSWITCH
vpp-agent-ctl -del /vnf-agent/${RNG_NAME}/vpp/config/v1/interface/memif-to-vswitch

# VSWITCH - add one more static route
vpp-agent-ctl -put /vnf-agent/${VSWITCH_NAME}/vpp/config/v1/vrf/0/fib - << EOF
{
  "ip": [
    {
      "description": "Static route",
      "destination_address": "6.0.0.0/24",
      "next_hops": [
        {
          "address": "8.42.0.1",
          "outgoing_interface": "GigabitEthernet0/8/0"
        }
      ]
    },
    {
      "description": "Static route 2",
      "destination_address": "20.5.0.0/24",
      "next_hops": [
        {
          "address": "8.42.0.1",
          "outgoing_interface": "GigabitEthernet0/8/0"
        }
      ]
    }
  ]
}
EOF

# VSWITCH - create cross-connection between interfaces memif-to-ussched and memif-to-vnf-1
vpp-agent-ctl -put /vnf-agent/${VSWITCH_NAME}/vpp/config/v1/xconnect/memif-to-ussched - << EOF
{
   "receive_interface": "memif-to-ussched",
   "transmit_interface": "memif-to-vnf-1"
}
EOF
