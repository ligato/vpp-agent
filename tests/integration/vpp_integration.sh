#!/usr/bin/env bash

set -exuo pipefail

_image="ligato/dev-vpp-agent:dev"

docker run -i -v $(pwd):/go/src/github.com/ligato/vpp-agent ${_image} go test -v ./tests/integration/vpp
