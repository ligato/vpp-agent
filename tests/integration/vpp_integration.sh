#!/bin/bash
set -eu

# compile test
go test -c ./tests/integration/vpp -o ./tests/integration/vpp/vpp-integration.test

# start vpp image
cid=$(docker run -d -it \
	-v $(pwd)/tests/integration/vpp/vpp-integration.test:/vpp-integration.test:ro \
	--label vpp.integration.test="$*" \
	${DOCKER_ARGS-} \
	"$VPP_IMG" bash)

on_exit() {
	docker stop -t 2 "$cid" >/dev/null
	docker rm "$cid" >/dev/null
}

vppver=$(docker exec -i "$cid" dpkg-query -f '${Version}' -W vpp)

trap 'on_exit' EXIT

echo "============================================================="
echo -e " VPP Integration Test - \e[1;33m${vppver}\e[0m"
echo "============================================================="

# run integration test
if docker exec -i "$cid" /vpp-integration.test $*; then
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
