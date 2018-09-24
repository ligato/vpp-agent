#!/bin/bash

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

if [[ ${BUILDARCH} = "x86_64" && ${IMAGE_TAG} = "prod_vpp_agent" ]] ; then
  # tag also explicit docker image on AMD64 platform
  docker tag  ${IMAGE_TAG}:latest prod_vpp_agent-amd64:latest
fi

# cleaning for arm64 platform
if [ ${BUILDARCH} = "aarch64" ] ; then
  docker rmi dev_vpp_agent:latest
fi
