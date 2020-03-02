#!/usr/bin/env bash
set -euo pipefail

# usage: ./scripts/run_perf_test.sh <num_req> <num_tunnels_per_req> <num_clients>

num_req=${1-10000}

image=${AGENT_IMG:-ligato/dev-vpp-agent}
reports=${REPORT_DIR:-report}
profiling_mode=${PROF_MODE-}

reports="$(cd $reports && pwd)"

runid=${RUN-"${num_req}-req"}
results="${reports}/perf-results-${runid}"

mkdir -p "$results"

echo "Starting perf test run: $runid"

cid=$(docker run -d -it --privileged \
	--label perf-run="$runid" \
	-v "$results":/report \
	-e REPORT_DIR=/report \
	-e ETCD_CONFIG=disabled \
	-e INITIAL_LOGLVL=info \
	-e DEBUG_ENABLED=y \
	-e DEBUG_PROFILE_MODE="$profiling_mode" \
	${DOCKER_EXTRA_ARGS:-} \
	-- \
	"$image" /bin/bash \
)

function on_exit() {
	docker stop -t 1 "$cid"
	exit
}
trap 'on_exit' EXIT

docker exec -it "$cid" bash ./tests/perf/perf_test.sh grpc-perf $*

echo "Test results stored in: $results"
