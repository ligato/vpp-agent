#!/bin/bash

cd "$(dirname "$0")"

set -euo pipefail

[ -n "${VPP_IMG-}" ] || {
  echo "VPP_IMG not set, use 'make images' to build docker images"
  exit 1
}

echo "==============================================="
echo " Image: ${IMAGE_TAG:=dev_vpp_agent}"
echo "==============================================="
echo " VPP"
echo "-----------------------------------------------"
echo " - base image: ${VPP_IMG}"
echo " - binapi dir: ${VPP_BINAPI}"
echo "-----------------------------------------------"
echo " Agent"
echo "-----------------------------------------------"
echo " - version: ${VERSION}"
echo " - commit:  ${COMMIT}"
echo " - branch:  ${BRANCH}"
echo " - date:    ${BUILD_DATE}"
echo "==============================================="

set -x

docker build -f Dockerfile \
    --build-arg VPP_IMG=${VPP_IMG} \
    --build-arg VPP_BINAPI=${VPP_BINAPI} \
    --build-arg VERSION=${VERSION} \
    --build-arg COMMIT=${COMMIT} \
    --build-arg BRANCH=${BRANCH} \
    --build-arg BUILD_DATE=${BUILD_DATE} \
    --build-arg SKIP_CHECK=${SKIP_CHECK:-} \
    --tag ${IMAGE_TAG} \
 ${DOCKER_BUILD_ARGS-} ../..

docker run --rm "${IMAGE_TAG}" vpp-agent -h || true
