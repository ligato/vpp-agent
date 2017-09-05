#!/bin/bash

set +e
sudo docker rm -f extract 2>/dev/null
set -e

sudo docker run -itd --name extract dev_vpp_agent bash

rm -rf agent
mkdir -p agent
sudo docker cp extract:/root/go/bin/vpp-agent agent/
sudo docker cp extract:/root/go/bin/vpp-agent-ctl agent/
sudo docker cp extract:/root/go/bin/agentctl agent/

tar -zcvf agent.tar.gz agent

sudo docker rm -f extract
