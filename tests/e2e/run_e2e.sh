#!/bin/bash
set -eu

args=($*)

# prepare vpp-agent executable
if [ -z "${COVER_DIR-}" ]; then
	go build -v -o ./tests/e2e/vpp-agent ./cmd/vpp-agent
else
	if [ ! -d ${COVER_DIR}/e2e-coverage ]; then
		mkdir ${COVER_DIR}/e2e-coverage
	elif [ "$(ls -A ${COVER_DIR}/e2e-coverage)" ]; then
		rm -f ${COVER_DIR}/e2e-coverage/*
	fi
	go test -covermode=count -coverpkg="github.com/ligato/vpp-agent/..." -c ./cmd/vpp-agent -o ./tests/e2e/vpp-agent
	DOCKER_ARGS="${DOCKER_ARGS-} -v ${COVER_DIR}/e2e-coverage:${COVER_DIR}/e2e-coverage"
	args+=("-cov=${COVER_DIR}/e2e-coverage")
fi

# compile test
go test -c ./tests/e2e -o ./tests/e2e/e2e.test

# start vpp image
cid=$(docker run -d -it \
	-v $(pwd)/tests/e2e/e2e.test:/e2e.test:ro \
	-v $(pwd)/tests/e2e/vpp-agent:/vpp-agent:ro \
	-v $(pwd)/tests/e2e/grpc.conf:/etc/grpc.conf:ro \
	-v /var/run/docker.sock:/var/run/docker.sock \
	--label e2e.test="$*" \
	--pid="host" \
	--privileged \
	--env KVSCHEDULER_GRAPHDUMP=true \
	--env VPP_IMG="$VPP_IMG" \
	--env GRPC_CONFIG=/etc/grpc.conf \
	${DOCKER_ARGS-} \
	"$VPP_IMG" bash)


on_exit() {
	docker stop -t 2 "$cid" >/dev/null
	docker rm "$cid" >/dev/null

	# merge coverage
	if [ ! -z "${COVER_DIR-}" ]; then
		go get github.com/wadey/gocovmerge
		find ${COVER_DIR}/e2e-coverage -type f | xargs gocovmerge > ${COVER_DIR}/e2e-cov.out
	fi
}

vppver=$(docker exec -i "$cid" dpkg-query -f '${Version}' -W vpp)

trap 'on_exit' EXIT

echo "============================================================="
echo -e " E2E Test - \e[1;33m${vppver}\e[0m"
echo "============================================================="

# run e2e test
if docker exec -i "$cid" /e2e.test -test.v ${args[@]}; then
	echo >&2 "-------------------------------------------------------------"
	echo >&2 -e " \e[32mPASSED\e[0m (took: ${SECONDS}s)"
	echo >&2 "-------------------------------------------------------------"
	exit 0
else
	res=$?
	echo >&2 "-------------------------------------------------------------"
	echo >&2 -e " \e[31mFAILED!\e[0m (exit code: $res)"
	echo >&2 "-------------------------------------------------------------"

	# dump container logs
	logs=$(docker logs --tail 10 "$cid")
	if [[ -n "$logs" ]]; then
		echo >&2 -e "\e[1;30m$logs\e[0m"
	fi

	exit $res
fi
