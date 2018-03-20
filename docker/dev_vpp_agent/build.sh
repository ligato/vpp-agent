#!/bin/bash

set -e

VPP_DEBUG_DEB=${VPP_DEBUG_DEB:-}
IMAGE_TAG=${IMAGE_TAG:-dev_vpp_agent}

AGENT_COMMIT=`git rev-parse HEAD`
VPP_COMMIT=`cd ../.. && git submodule status|grep vpp|cut -c2-41`

echo "repo agent commit: ${AGENT_COMMIT}"
echo "repo vpp commit:   ${VPP_COMMIT}"

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

sudo docker build --tag ${IMAGE_TAG} --build-arg VPP_DEBUG_DEB=${VPP_DEBUG_DEB} --build-arg AGENT_COMMIT=${AGENT_COMMIT} --build-arg VPP_COMMIT=${VPP_COMMIT} .
