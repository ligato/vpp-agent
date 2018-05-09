#!/bin/bash

set -e

IMAGE_TAG=${IMAGE_TAG:-dev_vpp_agent}

BASE_IMG=${BASE_IMG:-ubuntu:16.04}
GOLANG_OS_ARCH=${GOLANG_OS_ARCH:-linux-amd64}

AGENT_COMMIT=`git rev-parse HEAD`

source ../../vpp.env
VPP_DEBUG_DEB=${VPP_DEBUG_DEB:-}

echo "base image:   ${BASE_IMG}"
echo "agent commit: ${AGENT_COMMIT}"
echo "vpp repo url: ${VPP_REPO_URL}"
echo "vpp commit:   ${VPP_COMMIT}"

while [ "$1" != "" ]; do
    case $1 in
        -a | --agent )          shift
                                AGENT_COMMIT=$1
                                echo "using agent commit: ${AGENT_COMMIT}"
                                ;;
        -v | --vpp )            shift
                                VPP_COMMIT=$1
                                echo "using vpp commit: ${VPP_COMMIT}"
                                ;;
        * )                     echo "invalid parameter $1"
                                exit 1
    esac
    shift
done

echo
echo "building docker image: ${IMAGE_TAG}"

sudo docker build --tag ${IMAGE_TAG} \
    --build-arg BASE_IMG=${BASE_IMG} \
    --build-arg VPP_COMMIT=${VPP_COMMIT} \
    --build-arg VPP_REPO_URL=${VPP_REPO_URL} \
    --build-arg VPP_DEBUG_DEB=${VPP_DEBUG_DEB} \
    --build-arg AGENT_COMMIT=${AGENT_COMMIT} \
    --build-arg GOLANG_OS_ARCH=${GOLANG_OS_ARCH} \
 .
