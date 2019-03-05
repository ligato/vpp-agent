#!/bin/bash

# build perf test ...
echo "building main.go ..."
go build -v

# start up vpp, get its pid so we can kill it after the test, stdout has some crap from vpp
echo "starting vpp ..."
/usr/bin/exec_vpp.sh &
pid_vpp=$!
sleep 5

# start up the ageent, get its pid so we can kill it after the test ... send logs to a file
echo "starting vpp-agent"
vpp-agent -etcd-config=etcd.conf -grpc-config=grpc.conf > /tmp/vpp-agent.log 2>&1 &
pid_vpp_agent=$!
sleep 5

# now we can finally run the test with the tunnel count in $1
echo "running grpc perf test ..."
./grpc-perf -tunnels=$1

# test is complete, kill the vpp and vpp-agent
kill -9 $pid_vpp $pid_vpp_agent

