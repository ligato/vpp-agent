#!/bin/bash

set +e
sudo docker rm -f extract 2>/dev/null
set -e

sudo docker run -itd --name extract dev_vpp_agent bash
sudo docker exec extract /bin/bash -c 'mkdir -p /root/vpp/build-root && cp /opt/vpp-agent/dev/vpp/build-root/*.deb /root/vpp/build-root/ && cd /root && tar -zcvf /root/vpp.tar.gz vpp/*'
sudo docker cp extract:/root/vpp.tar.gz .
sudo docker rm -f extract
