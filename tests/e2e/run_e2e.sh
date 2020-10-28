#!/usr/bin/env bash
set -Eeuo pipefail

echo "Preparing end-to-end tests.."

args=($*)
VPP_IMG="${VPP_IMG:-ligato/vpp-base}"
testname="vpp-agent-e2e-test"
imgname="vpp-agent-e2e-tests"

# Compile vpp-agent for testing
if [ -z "${COVER_DIR-}" ]; then
	go build -o ./tests/e2e/vpp-agent.test \
	  -tags 'osusergo netgo' \
    -ldflags '-w -s -extldflags "-static" -X go.ligato.io/vpp-agent/v3/pkg/version.app=vpp-agent.test-e2e' \
    ./cmd/vpp-agent
else
	if [ ! -d ${COVER_DIR}/e2e-coverage ]; then
		mkdir ${COVER_DIR}/e2e-coverage
	elif [ "$(ls -A ${COVER_DIR}/e2e-coverage)" ]; then
		rm -f ${COVER_DIR}/e2e-coverage/*
	fi
	go test -c -o ./tests/e2e/vpp-agent.test \
	  -tags teste2e \
	  -covermode=count \
	  -coverpkg="go.ligato.io/vpp-agent/v3/..." \
	  -ldflags '-w -s -extldflags "-static" -X go.ligato.io/vpp-agent/v3/pkg/version.app=vpp-agent.test-e2e' \
	  ./cmd/vpp-agent
	DOCKER_ARGS="${DOCKER_ARGS-} -v ${COVER_DIR}/e2e-coverage:${COVER_DIR}/e2e-coverage"
	args+=("-cov=${COVER_DIR}/e2e-coverage")
fi

# Compile agentctl for testing
go build -o ./tests/e2e/agentctl.test \
	  -tags 'osusergo netgo' \
    -ldflags '-w -s -extldflags "-static"' \
    ./cmd/agentctl

# Compile testing suite
go test -c -o ./tests/e2e/e2e.test \
	  -tags 'osusergo netgo' \
    -ldflags '-w -s -extldflags "-static"' \
    ./tests/e2e

# Build testing image
docker build \
    -f ./tests/e2e/Dockerfile.e2e \
    --build-arg VPP_IMG \
    --tag "${imgname}" \
    ./tests/e2e

vppver=$(docker run --rm -i "$VPP_IMG" dpkg-query -f '${Version}' -W vpp)

cleanup() {
	echo "stopping test container"
	set -x
	docker stop -t 1 "${testname}" 2>/dev/null
	docker rm -v "${testname}" 2>/dev/null

	# merge coverage
	if [ ! -z "${COVER_DIR-}" ]; then
		go get github.com/wadey/gocovmerge
		find ${COVER_DIR}/e2e-coverage -type f | xargs gocovmerge > ${COVER_DIR}/e2e-cov.out
	fi
}

trap 'cleanup' EXIT

echo "============================================================="
echo -e " E2E TEST - VPP \e[1;33m${vppver}\e[0m"
echo "============================================================="

# Run e2e tests
if docker run -i \
	--name "${testname}" \
	--pid=host \
	--privileged \
	--label io.ligato.vpp-agent.testsuite=e2e \
	--label io.ligato.vpp-agent.testname="${testname}" \
	--volume $(pwd)/tests/e2e/resources/certs:/etc/certs:ro \
	--volume /var/run/docker.sock:/var/run/docker.sock \
	--env CERTS_PATH="$PWD/tests/e2e/resources/certs" \
	--env INITIAL_LOGLVL \
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
