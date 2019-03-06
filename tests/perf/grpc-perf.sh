#!/bin/bash

# this script calls the run script in the grpc-perf folder varying the tunnel count
# not sure how you wanted the results collected ... they wil go to stdout

cd grpc-perf
./run_agent_vpp_and_perf.sh 100
./run_agent_vpp_and_perf.sh 1000
./run_agent_vpp_and_perf.sh 10000


