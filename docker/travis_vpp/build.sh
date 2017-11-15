#!/bin/bash

set -e

VPP_COMMIT=`git submodule status|grep vpp|cut -c2-41`
VPP_IMG_TAG=`echo ${VPP_COMMIT} | cut -c1-7`

VPPDEB_REPO="ligato/vppdeb"
VPPDEB_IMG="${VPPDEB_REPO}:${VPP_IMG_TAG}"

echo "Building travis image.. VPP_COMMIT=${VPP_COMMIT}"

docker build -t ligato/vppdeb --build-arg VPP_COMMIT=${VPP_COMMIT} .
docker tag ligato/vppdeb:latest ${VPPDEB_IMG}

echo "Pushing image ${VPPDEB_IMG} to Dockerhub.."

docker push ${VPPDEB_IMG}
