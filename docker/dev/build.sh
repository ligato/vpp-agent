#!/bin/bash

cd "$(dirname "$0")"

set -e

IMAGE_TAG=${IMAGE_TAG:-'dev_vpp_agent'}
DOCKERFILE=${DOCKERFILE:-'Dockerfile'}

BASE_IMG=${BASE_IMG:-'ubuntu:16.04'}
GOLANG_OS_ARCH=${GOLANG_OS_ARCH:-'linux-amd64'}

source ../../vpp.env
VPP_DEBUG_DEB=${VPP_DEBUG_DEB:-}

VERSION=$(git describe --always --tags --dirty)
COMMIT=$(git rev-parse HEAD)

echo "=============================="
echo "VPP repo URL: ${VPP_REPO_URL}"
echo "VPP commit:   ${VPP_COMMIT}"
echo
echo "Agent version: ${VERSION}"
echo "Agent commit:  ${COMMIT}"
echo
echo "base image: ${BASE_IMG}"
echo "image tag:  ${IMAGE_TAG}"
echo "=============================="

docker build -f ${DOCKERFILE} \
    --tag ${IMAGE_TAG} \
    --build-arg BASE_IMG=${BASE_IMG} \
    --build-arg VPP_COMMIT=${VPP_COMMIT} \
    --build-arg VPP_REPO_URL=${VPP_REPO_URL} \
    --build-arg VPP_DEBUG_DEB=${VPP_DEBUG_DEB} \
    --build-arg GOLANG_OS_ARCH=${GOLANG_OS_ARCH} \
    --build-arg VERSION=${VERSION} \
    --build-arg COMMIT=${COMMIT} \
    ${DOCKER_BUILD_ARGS} ../..
