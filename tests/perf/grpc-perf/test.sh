#!/bin/bash

set -euo pipefail

BASH_ENTRY_DIR="$(dirname $(readlink -e "${BASH_SOURCE[0]}"))"

# ------
# config
# ------

[ -z ${REPORT_DIR-} ] && REPORT_DIR=/tmp/perf-report

log_report="${REPORT_DIR}/report.log"

log_vpp="${REPORT_DIR}/vpp.log"
log_agent="${REPORT_DIR}/agent.log"

sys_info="${REPORT_DIR}/sys-info.txt"
vpp_info="${REPORT_DIR}/vpp-info.txt"
agent_info="${REPORT_DIR}/agent-info.txt"

# -------
#  test
# -------

function run_test() {
	# create report directory
	mkdir -p ${REPORT_DIR}
	perftest "$1" 2>&1 | tee $log_report
}

function perftest() {
	local requests="${1:-1000}"
	
	echo "================================================================================"
	echo " grpc-perf test - ${requests} requests"
	echo "================================================================================"
	
	prepare_test

	trap 'on_exit' EXIT

	start_vpp
	start_agent

	echo "-> running grpc-perf.."
	echo "--------------------------------------------------------------"
	test_result=0
	./grpc-perf -tunnels=$requests || test_result=$?
	echo "--------------------------------------------------------------"
	echo "-> grpc-perf finished (exit code: $test_result)"

	sleep 1
	check_vpp
	check_agent

	echo "-> collecting system info to: $sys_info"
	sysinfo "uname -a" > $sys_info
	sysinfo "lscpu" >> $sys_info
	sysinfo "ip addr" >> $sys_info
	sysinfo "free -m" >> $sys_info
	sysinfo "df -h" >> $sys_info
	sysinfo "env" >> $sys_info
	
	echo "-> collecting agent info to: $agent_info"
	grep -B 6 "Starting agent version" $log_agent > $agent_info
	agentrest "scheduler/stats" >> $agent_info
	agentrest "vpp/binapitrace" >> $agent_info

	echo "-> collecting VPP info to: $vpp_info"
	vppcli "show version" > $vpp_info
	vppcli "show memory" >> $vpp_info
	vppcli "show api ring-stats" >> $vpp_info
	vppcli "show api histogram" >> $vpp_info
	
	if [[ "$test_result" == "0" ]]; then
		echo "--------------------------------------------------------------------------------"
		echo " ✓ Test run passed!"
		echo "--------------------------------------------------------------------------------"
	else
		fail "Test client grpc-perf failure (exit code: $test_result)"
	fi
	
	trap - EXIT
	stop_agent
	stop_vpp
	
	echo
}

function prepare_test() {
	cd ${BASH_ENTRY_DIR}
	# build test client
	[[ -e "./grpc-perf" ]] || { 
		echo "-> compiling grpc-perf.."
		go build -v
	}
}

function fail() {
    set +eu
    local msg=${1:-"Unknown cause."}
    echo -e "$msg" >&2
	echo "--------------------------------------------------------------------------------"
    echo " ✗ Test run failed!"
	echo "--------------------------------------------------------------------------------"
    exit "${2:-1}"
}

function on_exit() {
	echo "-> cleaning up"
	stop_agent
	stop_vpp
	exit
}

function sysinfo() {
	local cmd=$1
	echo "$ $cmd:"
	echo "----------------------------------------------------"
	sh -c "$cmd"
	echo "----------------------------------------------------"
	echo
}

# -----
#  VPP
# -----

wait_vpp_boot=3

function start_vpp() {
	echo -n "-> starting VPP.. "
	rm -f /dev/shm/db /dev/shm/global_vm /dev/shm/vpe-api
	vpp -c /etc/vpp/vpp.conf > "$log_vpp" 2>&1 &
	pid_vpp="$!"
	timeout "${wait_vpp_boot}" grep -q "vlib_plugin_early_init" <(tail -qF $log_vpp)
	echo "ok! (PID:${pid_vpp})"
}

function stop_vpp() {
	set +ue
	if ps -p $pid_vpp >/dev/null 2>&1; then	
		echo "-> stopping VPP.."
		kill -SIGTERM $pid_vpp
		wait $pid_vpp
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

function vppcli() {
	local cmd=$1
	echo "vpp# $cmd"
	echo "----------------------------------------------------"
	vppctl -s localhost:5002 "$cmd"
	echo "----------------------------------------------------"
	echo
}

# -------
#  agent
# -------

wait_agent_boot=5

function start_agent() {
	echo -n "-> starting agent.. "
	vpp-agent -etcd-config=etcd.conf -grpc-config=grpc.conf > "$log_agent" 2>&1 &
	pid_agent="$!"
	timeout "${wait_agent_boot}" grep -q "Agent started" <(tail -qF $log_agent) || {
		fail "timeout!"
	}
	echo "ok! (PID:${pid_agent})"
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

function agentrest() {
	local url=$1
	echo "curl $url"
	echo "----------------------------------------------------"
	curl -s -X GET -H "Content-Type: application/json" http://127.0.0.1:9191/$url
	echo "----------------------------------------------------"
	echo
}

# skip running test if no argument is given (source)
[ -z "${1-}" ] || run_test $1

