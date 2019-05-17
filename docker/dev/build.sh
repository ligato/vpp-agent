#!/bin/bash

cd "$(dirname "$0")"

set -e

source ../../vpp.env

VERSION=$(git describe --always --tags --dirty)
COMMIT=$(git rev-parse HEAD)
DATE=$(git log -1 --format="%ct" | xargs -I{} date -d @{} +'%Y-%m-%dT%H:%M%:z')

echo "==============================================="
echo "Building dev image: ${IMAGE_TAG:=dev_vpp_agent}"
echo "==============================================="
echo " vpp image: ${VPP_IMG}"
echo "-----------------------------------------------"
echo "Agent"
echo "-----------------------------------------------"
echo " version: ${VERSION}"
echo " commit:  ${COMMIT}"
echo " date:    ${DATE}"
echo "==============================================="

docker build -f Dockerfile \
    --build-arg VPP_IMG=${VPP_IMG} \
    --build-arg VERSION=${VERSION} \
    --build-arg COMMIT=${COMMIT} \
    --build-arg DATE=${DATE} \
    --tag ${IMAGE_TAG} \
 ${DOCKER_BUILD_ARGS} ../..
