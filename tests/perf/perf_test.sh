#!/bin/bash

set -euo pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
BASH_ENTRY_DIR="$(dirname $(readlink -e "${BASH_SOURCE[0]}"))"

# ------
# config
# ------

_test="${1}"
_requests="${2-1000}"
_perreq="${3-5}"
_numclients="${4-1}"

_test_client="$SCRIPT_DIR/$_test"
_vpp_config="$_test_client/vpp.conf"

[ -z ${REPORT_DIR-} ] && REPORT_DIR="${SCRIPT_DIR}/reports/$_test"

export DEBUG_PROFILE_PATH="${REPORT_DIR}"

log_report="${REPORT_DIR}/report.log"
log_vpp="${REPORT_DIR}/vpp.log"
log_agent="${REPORT_DIR}/agent.log"

sys_info="${REPORT_DIR}/sys-info.txt"
vpp_info="${REPORT_DIR}/vpp-info.txt"
agent_info="${REPORT_DIR}/agent-info"

cpuprof="${REPORT_DIR}/cpu.pprof"
memprof="${REPORT_DIR}/mem.pprof"

rest_addr="${REST_ADDR:-http://127.0.0.1:9191}"
sleep_extra="${SLEEP_EXTRA:-5}"

# -------
#  test
# -------

function run_test() {
	echo "Preparing PERF testing.."

	# create report directory
	rm -vrf ${REPORT_DIR}/* 2>/dev/null
	mkdir --mode=777 -p ${REPORT_DIR}

	perftest $* 2>&1 | tee "$log_report"
}

function perftest() {
	local perftest="$1"
	local requests="$2"
	local tunnels="$3"
	local clients="$4"

	echo "================================================================================"
	echo " PERF TEST - ${perftest}"
	echo "================================================================================"
	echo "report dir: ${REPORT_DIR}"
	echo
	echo "settings:"
	echo " - requests per client: ${requests}"
	echo " - tunnels per request: ${_perreq}"
	echo " - clients: ${_numclients}"
	echo "--------------------------------------------------------------------------------"

	prepare_test

	trap 'on_exit' EXIT
	start_vpp
	start_agent

	echo "-> sleeping for $sleep_extra seconds before starting test"
	sleep "$sleep_extra"

	echo "-> starting $perftest test.."
	echo "--------------------------------------------------------------"
	test_result=0
	"$_test_client"/"$_test" --tunnels=$requests --numperreq=$tunnels --clients=$clients ${CLIENT_PARAMS:-}  || test_result=$?
	echo "--------------------------------------------------------------"
	echo "-> $_test test finished (exit code: $test_result)"

	sleep 1

	check_vpp
	check_agent

	set +e

	#curl -sSfL "http://127.0.0.1:9094/metrics" > "${REPORT_DIR}/metrics_client.txt" || true

	echo "-> collecting system info to: $sys_info"
	sysinfo "pwd" >> $sys_info
	sysinfo "env | sort" >> $sys_info
	sysinfo "uname -a" > $sys_info
	sysinfo "lscpu" >> $sys_info
	sysinfo "free -h" >> $sys_info
	sysinfo "df -h" >> $sys_info
	sysinfo "ip -br link" >> $sys_info
	sysinfo "ip -br addr" >> $sys_info
	sysinfo "ip -br route" >> $sys_info
	sysinfo "ps faux" >> $sys_info

	echo "-> collecting agent data.."
	curljson "$rest_addr/scheduler/stats" > "${REPORT_DIR}/agent-stats_scheduler.json"
	curljson "$rest_addr/govppmux/stats" > "${REPORT_DIR}/agent-stats_govppmux.json"
	curl -sSfL "$rest_addr/metrics" > "${REPORT_DIR}/metrics_agent.txt"

	echo "-> collecting VPP data to: $vpp_info"
	echo -e "VPP info:\n\n" > $vpp_info
	vppcli "show clock" >> $vpp_info
	vppcli "show version verbose" >> $vpp_info
	vppcli "show plugins" >> $vpp_info
	vppcli "show cpu" >> $vpp_info
	vppcli "show version cmdline" >> $vpp_info
	vppcli "show threads" >> $vpp_info
	vppcli "show physmem" >> $vpp_info
	vppcli "show memory main-heap verbose" >> $vpp_info
	vppcli "show memory api-segment verbose" >> $vpp_info
	vppcli "show memory stats-segment verbose" >> $vpp_info
	vppcli "show api histogram" >> $vpp_info
	vppcli "show api ring-stats" >> $vpp_info
	vppcli "show api trace-status" >> $vpp_info
	vppcli "api trace status" >> $vpp_info
	vppcli "show api clients" >> $vpp_info
	vppcli "show unix files" >> $vpp_info
	vppcli "show unix errors" >> $vpp_info
	vppcli "show event-logger" >> $vpp_info
	vppcli "show ip fib summary" >> $vpp_info

	if [[ "$test_result" == "0" ]]; then
		echo "--------------------------------------------------------------------------------"
		echo " ✓ Test run passed!"
		echo "--------------------------------------------------------------------------------"
	else
		fail "Test client failure (exit code: $test_result)"
	fi

	echo "-> sleeping for $sleep_extra seconds before stopping"
	sleep "$sleep_extra"
	
	trap - EXIT
	stop_agent
	stop_vpp

	echo -n "-> processing profiles.. "
	set +e
	if [ -r "$cpuprof" ]; then
		go tool pprof -dot "$cpuprof" | dot -Tsvg -o "$REPORT_DIR/cpu-profile.svg"
	fi
	if [ -r "$memprof" ]; then
		go tool pprof -alloc_space -dot "$memprof" | dot -Tsvg -o "$REPORT_DIR/mem-profile.svg"
	fi
	set -e

	echo
}

function prepare_test() {
	#cd ${BASH_ENTRY_DIR}
	# build test client

	#[[ -e "./$_test" ]] || {
		echo "-> compiling test client $_test.."
		go build -o "$_test_client/$_test" "$_test_client"
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
	rm -vf /dev/shm/db /dev/shm/global_vm /dev/shm/vpe-api
	$_vpp -c "${_vpp_config}" > "$log_vpp" 2>&1 &
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

wait_agent_boot=10

function start_agent() {
	local _agent="$(which vpp-agent)"
	[[ -e "$_agent" ]] || fail "vpp-agent not found!"

	export CONFIG_DIR="$_test_client"

	echo -n "-> starting agent.. "
	$_agent > "$log_agent" 2>&1 &
	pid_agent="$!"
	timeout "${wait_agent_boot}" grep -q "Agent started" <(tail -qF $log_agent) || {
		echo "AGENT LOG:"
		echo "---"
		tail "$log_agent"
		echo "---"
		fail "TIMEOUT!"
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

function curljson() {
	local url="$1"

	curl -sSfL -H "Content-Type: application/json" "$url"
}

# skip running test if no argument is given (source)
#[ -z "${1-}" ] || run_test $1

run_test "$_test" "$_requests" "$_perreq" "$_numclients"

