#!/bin/bash

cd "$(dirname "$0")"

set -euo pipefail

# This test script calls the run script in the ./grpc-perf subfolder 
# with varying the count of tunnels.

echo "================================================================================"
echo " gRPC Performance Test "
echo "================================================================================"

echo "-> preparing test.."

# install agent
make -C ../.. agent

# build test
cd grpc-perf
go build

./run_agent_vpp_and_perf.sh 10
./run_agent_vpp_and_perf.sh 100
./run_agent_vpp_and_perf.sh 1000
#./run_agent_vpp_and_perf.sh 10000

