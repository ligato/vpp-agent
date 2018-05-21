#!/bin/bash

set -e

IMAGE_TAG=${IMAGE_TAG:-dev_vpp_agent}

BASE_IMG=${BASE_IMG:-ubuntu:16.04}
GOLANG_OS_ARCH=${GOLANG_OS_ARCH:-linux-amd64}

source ../../vpp.env
VPP_DEBUG_DEB=${VPP_DEBUG_DEB:-}

echo "base image:   ${BASE_IMG}"
echo "vpp repo url: ${VPP_REPO_URL}"
echo "vpp commit:   ${VPP_COMMIT}"
echo
echo "building docker image: ${IMAGE_TAG}"

sudo docker build --tag ${IMAGE_TAG} --file ./Dockerfile \
    --build-arg BASE_IMG=${BASE_IMG} \
    --build-arg VPP_COMMIT=${VPP_COMMIT} \
    --build-arg VPP_REPO_URL=${VPP_REPO_URL} \
    --build-arg VPP_DEBUG_DEB=${VPP_DEBUG_DEB} \
    --build-arg GOLANG_OS_ARCH=${GOLANG_OS_ARCH} \
 ../..
