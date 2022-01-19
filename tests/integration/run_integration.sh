#!/usr/bin/env bash
set -Eeuo pipefail

echo "Preparing integration tests.."

args=($*)
VPP_IMG="${VPP_IMG:-ligato/vpp-base}"
testname="vpp-agent-integration-test"
imgname="vpp-agent-integration-tests"

# Compile testing suite
go test -c -o ./tests/integration/integration.test \
    -covermode atomic \
	-tags 'osusergo netgo integration' \
    -ldflags '-w -s -extldflags "-static"' \
    ./tests/integration/...

# Build testing image
docker build \
    -f ./tests/integration/Dockerfile.integration \
    --build-arg VPP_IMG \
    --tag "${imgname}" \
    ./tests/integration

vppver=$(docker run --rm -i "$VPP_IMG" dpkg-query -f '${Version}' -W vpp)

cleanup() {
	echo "stopping test container"
	docker stop -t 1 "${testname}" 2>/dev/null
	docker rm -v "${testname}" 2>/dev/null
}

trap 'cleanup' EXIT

echo "============================================================="
echo -e " VPP Integration Test - \e[1;33m${vppver}\e[0m"
echo "============================================================="

# Run integration tests
if docker run -i \
	--name "${testname}" \
	--privileged \
	--label io.ligato.vpp-agent.testsuite=integration \
	--label io.ligato.vpp-agent.testname="${testname}" \
	--env INITIAL_LOGLVL \
	--env VPPVER=${vppver:0:5} \
	--volume $(pwd)/report:/reports \
	${DOCKER_ARGS-} \
	"${imgname}" ${args[@]:-}
then
	echo >&2 "-------------------------------------------------------------"
	echo >&2 -e " \e[32mPASSED\e[0m (took: ${SECONDS}s)"
	echo >&2 "-------------------------------------------------------------"
	exit 0
else
	res=$?
	echo >&2 "-------------------------------------------------------------"
	echo >&2 -e " \e[31mFAILED!\e[0m (exit code: $res)"
	echo >&2 "-------------------------------------------------------------"
	exit $res
fi
