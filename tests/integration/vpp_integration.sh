#!/usr/bin/env bash

set -euo pipefail


function on_exit() {
	echo "-> cleaning up"
	docker stop -t 3 vpp-integration
}

# compile vpp integration test
go test -v -c ./tests/integration/vpp

# start vpp image
docker run --rm --name vpp-integration -d -i -v $(pwd):/data:ro "${VPP_IMG}" bash -i
trap 'on_exit' EXIT

# run integration test
docker exec -it vpp-integration /data/vpp.test -test.v -vpp-config=/etc/vpp/startup.conf || {
	res=$?
	echo >&2 "VPP integration tests FAILED!"
	docker logs vpp-integration
	exit $res
}
