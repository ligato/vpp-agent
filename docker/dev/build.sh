#!/bin/bash
# Before run of this script you can set environmental variables
# IMAGE_TAG, DOCKERFILE, BASE_IMG, GOLANG_OS_ARCH, .. then  export them
# and to use defined values instead of default ones

cd "$(dirname "$0")"

set -e

#IMAGE_TAG=${IMAGE_TAG:-'dev_vpp_agent'}
DOCKERFILE=${DOCKERFILE:-'Dockerfile'}

BASE_IMG=${BASE_IMG:-'ubuntu:18.04'}
#GOLANG_OS_ARCH=${GOLANG_OS_ARCH:-'linux-amd64'}

#To prepare for future fat manifest image by multi-arch manifest,
#now build the docker image with its arch
#For fat manifest, please refer
#https://docs.docker.com/registry/spec/manifest-v2-2/#example-manifest-list

BUILDARCH=`uname -m`

case "$BUILDARCH" in
  "aarch64" )
    IMAGE_TAG=${IMAGE_TAG:-'dev_vpp_agent-arm64'}
    GOLANG_OS_ARCH=${GOLANG_OS_ARCH:-'linux-arm64'}
    ;;

  "x86_64" )
    # for AMD64 platform is used the default image (without suffix -amd64)
    IMAGE_TAG=${IMAGE_TAG:-'dev_vpp_agent'}
    GOLANG_OS_ARCH=${GOLANG_OS_ARCH:-'linux-amd64'}
    ;;
  * )
    echo "Architecture ${BUILDARCH} is not supported."
    exit
    ;;
esac


source ../../vpp.env
VPP_DEBUG_DEB=${VPP_DEBUG_DEB:-}

VERSION=$(git describe --always --tags --dirty)
COMMIT=$(git rev-parse HEAD)

echo "=============================="
echo "Architecture: ${BUILDARCH}"
echo
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

if [[ ${BUILDARCH} = "x86_64" && ${IMAGE_TAG} = "dev_vpp_agent" ]] ; then
  # create docker image tagged with -amd64 suffix for AMD64 platform
  docker tag  ${IMAGE_TAG}:latest dev_vpp_agent-amd64:latest
fi
