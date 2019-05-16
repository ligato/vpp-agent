#!/bin/bash

cd "$(dirname "$0")"

set -e

buildArch=`uname -m`
case "${buildArch##*-}" in
	aarch64) ;;
  	x86_64) ;;
  	*) echo "Current architecture (${buildArch}) is not supported."; exit 2; ;;
esac

echo "==============================================="
echo "Building prod image: ${IMAGE_TAG:=prod_vpp_agent}"
echo "==============================================="
echo " dev image: ${DEV_IMG:=dev_vpp_agent}"
echo "-----------------------------------------------"

docker build -f Dockerfile \
	--tag ${IMAGE_TAG} \
	${DOCKER_BUILD_ARGS} .
