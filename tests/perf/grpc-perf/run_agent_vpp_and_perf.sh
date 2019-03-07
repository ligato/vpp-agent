#!/bin/bash

set -euo pipefail

tunnels="$1"

echo "--------------------------------------------------------------------------------"
echo " running test with ${tunnels} tunnels "
echo "--------------------------------------------------------------------------------"

function fail() {
    set +eu
    echo "${1:-Test failure!}" >&2
    exit "${2:-1}"
}

function start_vpp() {
	echo -n "-> starting VPP.. "
	rm -f /dev/shm/db /dev/shm/global_vm /dev/shm/vpe-api
	vpp -c /etc/vpp/vpp.conf > /tmp/perf_vpp.log 2>&1 &
	pid_vpp="$!"
	echo "OK! (PID:${pid_vpp})"
}

function stop_vpp() {
	set +ue
	echo "-> stopping VPP.."
	kill ${pid_vpp}
}

function check_vpp() {
	set +ue
	if ! ps -p $pid_vpp >/dev/null 2>&1; then
		wait $pid_vpp
		fail "VPP failure! (exit code: $?)"
	fi
}

function start_agent() {
	echo -n "-> starting agent.. "
	DEBUG_ENABLED=true DEBUG_CPUPROFILE=/tmp/perf_cpu.prof DEBUG_TRACEPROFILE=/tmp/perf_trace.out \
		vpp-agent -etcd-config=etcd.conf -grpc-config=grpc.conf > /tmp/perf_vpp-agent.log 2>&1 &
	pid_agent="$!"
	echo "OK! (PID:${pid_agent})"
}

function stop_agent() {
	set +ue
	echo "-> stopping agent.."
	kill -SIGINT ${pid_agent}
}

function check_agent() {
	set +ue
	if ! ps -p $pid_agent >/dev/null 2>&1; then
		wait $pid_agent
		fail "Agent failure! (exit code: $?)"
	fi
}

trap 'stop_agent >/dev/null 2>&1; stop_vpp >/dev/null 2>&1; exit' EXIT

# start vpp & agent
start_vpp
sleep 3
start_agent
sleep 3

# run test
echo "-> starting test.."
./grpc-perf -tunnels=$tunnels || echo "Test exit code: $?"

# check crashes
check_vpp
check_agent

echo "-> collecting data.."
curl -s -X GET -H "Content-Type: application/json" http://127.0.0.1:9191/scheduler/stats
#curl -s -X GET -H "Content-Type: application/json" http://127.0.0.1:1234/debug/vars
# TODO: collect more data

echo "-> test complete"


