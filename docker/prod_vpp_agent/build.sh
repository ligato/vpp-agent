#!/bin/bash

set +e
sudo docker rmi -f prod_vpp_agent 2>/dev/null
set -e

./extract_agent_files.sh
./extract_vpp_files.sh

sudo docker build -t prod_vpp_agent --no-cache .
