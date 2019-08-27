#!/bin/bash
set -eu

# compile test
go test -c ./tests/e2e -o ./tests/e2e/e2e.test

# start vpp image
cid=$(docker run -d -it \
	-v $(pwd)/tests/e2e/e2e.test:/e2e.test:ro \
	-v /var/run/docker.sock:/var/run/docker.sock \
	--label e2e.test="$*" \
	--pid="host" \
	--privileged \
	${DOCKER_ARGS-} \
	"$VPP_IMG" bash)
#	--env KVSCHEDULER_GRAPHDUMP=true \

on_exit() {
	docker stop -t 2 "$cid" >/dev/null
	docker rm "$cid" >/dev/null
}

vppver=$(docker exec -i "$cid" dpkg-query -f '${Version}' -W vpp)

trap 'on_exit' EXIT

echo "============================================================="
echo -e " E2E Test - \e[1;33m${vppver}\e[0m"
echo "============================================================="

# run e2e test
if docker exec -i "$cid" /e2e.test -test.v $*; then
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
