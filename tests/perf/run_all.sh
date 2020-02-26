#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd $SCRIPT_DIR

function run() {
	test="${1}"
	typ="${2-basic}"
	requests="${3-${REQUESTS-1000}}"

	export DEBUG_ENABLED=y
	export REPORT_DIR="${reports}/${test}_${requests}_${typ}"
	./perf_test.sh "$test" "$requests"
}

export reports="${SCRIPT_DIR}/reports"

run "grpc-perf"

export CLIENT_PARAMS="--with-ips"
run "grpc-perf" "ips"
