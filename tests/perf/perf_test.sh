#!/bin/bash

set -euo pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
BASH_ENTRY_DIR="$(dirname $(readlink -e "${BASH_SOURCE[0]}"))"

# ------
# config
# ------

_test="${1}"
_requests="${2-1000}"

[ -z ${REPORT_DIR-} ] && REPORT_DIR="${SCRIPT_DIR}/reports/$_test"

export DEBUG_PROFILE_PATH="${REPORT_DIR}"

log_report="${REPORT_DIR}/report.log"

log_vpp="${REPORT_DIR}/vpp.log"
log_agent="${REPORT_DIR}/agent.log"

sys_info="${REPORT_DIR}/sys-info.txt"
vpp_info="${REPORT_DIR}/vpp-info.txt"
agent_info="${REPORT_DIR}/agent-info.txt"

cpuprof="${REPORT_DIR}/cpu.pprof"
memprof="${REPORT_DIR}/mem.pprof"

# -------
#  test
# -------

function run_test() {
	# create report directory
	rm -vrf ${REPORT_DIR}/*
	mkdir --mode=777 -p ${REPORT_DIR}

	perftest $* 2>&1 | tee $log_report
}

function perftest() {
	local perftest="$1"
	local requests="$2"
	
	echo "================================================================================"
	echo " PERF-TEST: $perftest - ${requests} requests"
	echo "  -> ${REPORT_DIR}"
	echo "================================================================================"
	
	prepare_test

	trap 'on_exit' EXIT

	start_vpp
	start_agent

	echo "-> running $perftest test.."
	echo "--------------------------------------------------------------"
	test_result=0
	$_test_client/$_test ${CLIENT_PARAMS-} --tunnels=$requests || test_result=$?
	echo "--------------------------------------------------------------"
	echo "-> $_test test finished (exit code: $test_result)"

	sleep 1

	check_vpp
	check_agent

	echo "-> collecting system info to: $sys_info"
	sysinfo "uname -a" > $sys_info
	sysinfo "env" >> $sys_info
	sysinfo "pwd" >> $sys_info
	sysinfo "lscpu" >> $sys_info
	sysinfo "ip addr" >> $sys_info
	sysinfo "free -h" >> $sys_info
	sysinfo "df -h" >> $sys_info
	sysinfo "ps faux" >> $sys_info

	echo "-> collecting agent info to: $agent_info"
	grep -B 6 "Starting agent version" $log_agent > $agent_info
	agentrest "scheduler/stats" >> $agent_info
	agentrest "govppmux/stats" >> $agent_info

	echo "-> collecting VPP info to: $vpp_info"
	echo -e "VPP info:\n\n" > $vpp_info
	vppcli "show version verbose" >> $vpp_info
	vppcli "show version cmdline" >> $vpp_info
	vppcli "show plugins" >> $vpp_info
	vppcli "show clock" >> $vpp_info
	vppcli "show threads" >> $vpp_info
	vppcli "show cpu" >> $vpp_info
	vppcli "show physmem" >> $vpp_info
	vppcli "show memory verbose" >> $vpp_info
	vppcli "show api clients" >> $vpp_info
	vppcli "show api histogram" >> $vpp_info
	vppcli "show api trace-status" >> $vpp_info
	vppcli "show api ring-stats" >> $vpp_info
	vppcli "api trace status" >> $vpp_info
	vppcli "show event-logger" >> $vpp_info
	vppcli "show unix errors" >> $vpp_info
	vppcli "show unix files" >> $vpp_info
	vppcli "show ip fib summary" >> $vpp_info

	if [[ "$test_result" == "0" ]]; then
		echo "--------------------------------------------------------------------------------"
		echo " ✓ Test run passed!"
		echo "--------------------------------------------------------------------------------"
	else
		fail "Test client failure (exit code: $test_result)"
	fi
	
	trap - EXIT
	stop_agent
	stop_vpp

	echo -n "-> processing profiles.. "
	[ -r "$cpuprof" ] && go tool pprof -dot "$cpuprof" | dot -Tsvg -o "$REPORT_DIR/cpu-profile.svg"
	[ -r "$memprof" ] && go tool pprof -alloc_space -dot "$memprof" | dot -Tsvg -o "$REPORT_DIR/mem-profile.svg"

	echo
}

function prepare_test() {
	#cd ${BASH_ENTRY_DIR}
	# build test client
	_test_client="$SCRIPT_DIR/$_test"

	#[[ -e "./$_test" ]] || {
		echo "-> compiling test client $_test.."
		go build -o "$_test_client/$_test" -v "$_test_client"
	#}
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

	echo "----------------------------------------------------"
	echo "$ $cmd"
	echo "----------------------------------------------------"
	bash -c "$cmd"
	echo
}

# ---------
#  VPP
# ---------

wait_vpp_boot=3

function start_vpp() {
	set +e

	if ps -C vpp_main >/dev/null 2>&1; then
		fail "VPP is already running"
	fi

	local _vpp="$(which vpp)"
	[[ -e "$_vpp" ]] || fail "VPP not found!"

	echo -n "-> starting VPP ($_vpp).. "
	rm -f /dev/shm/db /dev/shm/global_vm /dev/shm/vpe-api
	$_vpp -c /etc/vpp/vpp.conf > "$log_vpp" 2>&1 &
	pid_vpp="$!"
	timeout "${wait_vpp_boot}" grep -q "vlib_plugin_early_init" <(tail -qF $log_vpp)
	echo "ok! (PID:${pid_vpp})"
}

function stop_vpp() {
	set +ue

	if ps -p $pid_vpp >/dev/null 2>&1; then
		echo "-> stopping VPP (PID: $pid_vpp).."
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
	set +e

	local cmd=$1
	local _clisock="/run/vpp/cli.sock"

	echo "----------------------------------------------------"
	echo "vppctl $cmd"
	echo "----------------------------------------------------"
	if [ -S "$_clisock" ]; then
		vppctl "$cmd"
	else
		vppctl -s localhost:5002 "$cmd"
	fi
	echo
}

# -------------
#  vpp-agent
# -------------

wait_agent_boot=5

function start_agent() {
	local _agent="$(which vpp-agent)"
	[[ -e "$_agent" ]] || fail "vpp-agent not found!"

	export CONFIG_DIR="$_test_client"

	echo -n "-> starting agent.. "
	$_agent > "$log_agent" 2>&1 &
	pid_agent="$!"
	timeout "${wait_agent_boot}" grep -q "Agent started" <(tail -qF $log_agent) || {
		fail "timeout!"
	}
	echo "ok! (PID:${pid_agent})"
}

function stop_agent() {
	set +ue

	if ps -p $pid_agent >/dev/null 2>&1; then
		echo "-> stopping vpp-agent (PID: $pid_agent).."
		kill -SIGINT ${pid_agent}
		wait ${pid_agent}
	fi
}

function check_agent() {
	set +ue

	echo -n "-> checking vpp-agent.. "
	if ! ps -p $pid_agent >/dev/null 2>&1; then
		wait $pid_agent
		result=$?
		out=$(tail "$log_agent")
		fail "Agent down! (exit code: $result)\n\n$log_agent output:\n${out}\n"
	fi

	log_errors=$(grep 'level=error' "$log_agent")
	err_num=$(echo -n "$log_errors" | wc -l)

	if [[ "$err_num" -gt "0" ]]; then
		echo "found ${err_num} errors in log:"
		echo "-----"
		echo "$log_errors" | tail -n 10 | sed 's/.*/\t&/'
		echo "-----"
	else
		echo "ok!"
	fi
}

function agentrest() {
	local url="http://localhost:9191/$1"

	echo "----------------------------------------------------"
	echo "GET $url"
	echo "----------------------------------------------------"
	curl -sSfL -H "Content-Type: application/json" "$url"
	echo
}

# skip running test if no argument is given (source)
#[ -z "${1-}" ] || run_test $1

run_test "$_test" "$_requests"

