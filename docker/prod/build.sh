#!/bin/bash

cd "$(dirname "$0")"

set -euo pipefail

buildArch=`uname -m`
case "${buildArch##*-}" in
	  aarch64) ;;
  	x86_64) ;;
  	*) echo "Current architecture (${buildArch}) is not supported."; exit 2; ;;
esac

echo "==============================================="
echo " Image: ${IMAGE_TAG:=prod_vpp_agent}"
echo "==============================================="
echo " - dev image: ${DEV_IMG:=dev_vpp_agent}"
echo " - VPP version: ${VPP_VERSION}"
echo "==============================================="

set -x

docker build -f Dockerfile \
    --build-arg DEV_IMG=${DEV_IMG} \
    --build-arg VPP_VERSION=${VPP_VERSION} \
	  --tag ${IMAGE_TAG} \
 ${DOCKER_BUILD_ARGS-} .

docker run --rm "${IMAGE_TAG}" vpp-agent -h || true
