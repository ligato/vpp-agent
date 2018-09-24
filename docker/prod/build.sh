#!/bin/bash
# Before run of this script you can set environmental variables
# IMAGE_TAG ... then  export them
# and to use defined values instead of default ones

cd "$(dirname "$0")"

set -e

#To prepare for future fat manifest image by multi-arch manifest,
#now build the docker image with its arch
#For fat manifest, please refer
#https://docs.docker.com/registry/spec/manifest-v2-2/#example-manifest-list

BUILDARCH=`uname -m`

case "$BUILDARCH" in
  "aarch64" )
    IMAGE_TAG=${IMAGE_TAG:-'prod_vpp_agent-arm64'}
    # Dockerfile  for prod_vpp_agent expects that dev-vpp-agent:latest image is built
    docker tag dev_vpp_agent-arm64:latest dev_vpp_agent:latest
    ;;

  "x86_64" )
    # for AMD64 platform is used the default image (without suffix -amd64)
    IMAGE_TAG=${IMAGE_TAG:-'prod_vpp_agent'}
    # Dockerfile expects that dev-vpp-agent:latest image is built
    # Here it is granted
    ;;
  * )
    echo "Architecture ${BUILDARCH} is not supported."
    exit
    ;;
esac

docker build  ${DOCKER_BUILD_ARGS} --tag ${IMAGE_TAG} .

echo "To push to repository please use command:"
case "$BUILDARCH" in
  "aarch64" )
    echo "docker tag ${IMAGE_TAG}:latest ligato/prod-vpp-agent-arm64:$(git describe --always --tags)"
    echo "docker tag ${IMAGE_TAG}:latest ligato/prod-vpp-agent-arm64:latest"

    # cleaning for arm64 platform
    docker rmi dev_vpp_agent:latest > /dev/null 2>&1 
    ;;

  "x86_64" )
    # create docker image tagged with -amd64 suffix for AMD64 platform
    echo "docker tag ${IMAGE_TAG}:latest ligato/prod-vpp-agent-amd64:$(git describe --always --tags)"
    echo "docker tag ${IMAGE_TAG}:latest ligato/prod-vpp-agent:$(git describe --always --tags)"
    echo "docker tag ${IMAGE_TAG}:latest ligato/prod-vpp-agent-amd64:latest"
    echo "docker tag ${IMAGE_TAG}:latest ligato/prod-vpp-agent:latest"
    ;;
  * )
    echo "Architecture ${BUILDARCH} is not supported."
    exit
    ;;
esac
