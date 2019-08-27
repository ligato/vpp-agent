#!/bin/sh

docker exec etcd etcdctl put /vnf-agent/example-agent/config/vpp/v2/interfaces/tap0 \
        '{"name":"tap0","type":"TAP","enabled":true,"ip_addresses":["10.10.1.2/24"], "tap": {"version": 2, "host_if_name": "vpptap0"}}'

docker exec etcd etcdctl put /vnf-agent/example-agent/config/vpp/v2/interfaces/tap1 \
        '{"name":"tap1","type":"TAP","enabled":true,"ip_addresses":["10.20.1.2/24"], "tap": {"version": 2, "host_if_name": "vpptap1"}}'
