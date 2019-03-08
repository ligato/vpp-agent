#!/bin/bash

set -euo pipefail

tunnels="${1:-1}"

log_vpp=/tmp/grpc-perf_vpp.log
log_agent=/tmp/grpc-perf_agent.log

echo "--------------------------------------------------------------------------------"
echo " Test run: ${tunnels} tunnels "
echo "--------------------------------------------------------------------------------"

function fail() {
    set +eu
    echo -e "${1:-Test failure!}" >&2
    exit "${2:-1}"
}

function start_vpp() {
	echo -n "-> starting VPP.. "
	rm -f /dev/shm/db /dev/shm/global_vm /dev/shm/vpe-api
	vpp -c /etc/vpp/vpp.conf > "$log_vpp" 2>&1 &
	pid_vpp="$!"
	echo "ok! (PID:${pid_vpp})"
	sleep 3
}

function stop_vpp() {
	set +ue
	if ps -p $pid_vpp >/dev/null 2>&1; then	
		echo "-> stopping VPP.."
		kill -9 $pid_vpp
	fi
}

function check_vpp() {
	set +ue
	echo -n "-> checking VPP.. "
	if ! ps -p $pid_vpp >/dev/null 2>&1; then
		wait $pid_vpp
		vpp_result=$?
		out=$(tail "$log_vpp")
		fail "VPP down! (exit code: $vpp_result)\n\n$log_vpp output:\n${out}\n"
	fi
	log_errors=$(grep 'WARNING' "$log_vpp" | wc -l)
	if [[ "$log_errors" -gt "0" ]]; then
		echo "found ${log_errors} warnings!"
	else
		echo "ok!"
	fi
}

function start_agent() {
	echo -n "-> starting agent.. "
	DEBUG_ENABLED=true DEBUG_CPUPROFILE=/tmp/perf_cpu.prof DEBUG_TRACEPROFILE=/tmp/perf_trace.out \
		vpp-agent -etcd-config=etcd.conf -grpc-config=grpc.conf > "$log_agent" 2>&1 &
	pid_agent="$!"
	echo "ok! (PID:${pid_agent})"
	sleep 3
}

function stop_agent() {
	set +ue
	if ps -p $pid_agent >/dev/null 2>&1; then
		echo "-> stopping agent.."
		kill -SIGINT ${pid_agent}
		wait ${pid_agent}
	fi
}

function check_agent() {
	set +ue
	echo -n "-> checking agent.. "
	if ! ps -p $pid_agent >/dev/null 2>&1; then
		wait $pid_agent
		result=$?
		out=$(tail "$log_agent")
		fail "Agent down! (exit code: $result)\n\n$log_agent output:\n${out}\n"		
	fi	
	log_errors=$(grep 'level=error' "$log_agent" | wc -l)
	if [[ "$log_errors" -gt "0" ]]; then
		echo "found ${log_errors} errors!"
	else
		echo "ok!"
	fi
}

trap 'stop_agent; stop_vpp; exit' EXIT

start_vpp
start_agent

echo "-> running test.."
echo "--------------------------------------------------------------------------------"
test_result=0
./grpc-perf -tunnels=$tunnels || test_result=$?
echo "--------------------------------------------------------------------------------"
if [[ "$test_result" == "0" ]]; then
	echo " ✓ Test passed!"
else
	echo " ✗ Test failed! (exit code: $test_result)"
fi
echo "--------------------------------------------------------------------------------"

sleep 1
check_vpp
check_agent
sleep 1

if [[ "$test_result" != "0" ]]; then
	fail "Test failure!"
fi

echo "-> collecting data.."
curl -s -X GET -H "Content-Type: application/json" http://127.0.0.1:9191/scheduler/stats
curl -s -X GET -H "Content-Type: application/json" http://127.0.0.1:9191/vpp/binapitrace
#curl -s -X GET -H "Content-Type: application/json" http://127.0.0.1:1234/debug/vars
# TODO: collect more data

echo "-> test finished."
sleep 1


