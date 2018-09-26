#!/bin/bash

set -e

REPO_OWNER='ligato'
IMAGE_TAG_LOCAL='prod_vpp_agent'

#To prepare for future fat manifest image by multi-arch manifest,
#now build the docker image with its arch
#For fat manifest, please refer
#https://docs.docker.com/registry/spec/manifest-v2-2/#example-manifest-list

BUILDARCH=`uname -m`

case "$BUILDARCH" in
  "aarch64" )
    IMAGE_TAG='vpp-agent-arm64'
    ;;

  "x86_64" )
    # for AMD64 platform is also used the default image (without suffix -amd64)
    DEFAULT_IMAGE_TAG='vpp-agent'
    IMAGE_TAG='vpp-agent-amd64'
    ;;
  * )
    echo "Architecture ${BUILDARCH} is not supported."
    exit
    ;;
esac

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
echo "image tag:  ${IMAGE_TAG}"
echo "=============================="

if [ ${BUILDARCH} = "x86_64" ] ; then
  # for AMD64 platform is used also the default image (without suffix -amd64)
  docker tag ${IMAGE_TAG_LOCAL}:latest ${REPO_OWNER}/${DEFAULT_IMAGE_TAG}:${VERSION}
  docker push ${REPO_OWNER}/${DEFAULT_IMAGE_TAG}:${VERSION}
  docker tag ${IMAGE_TAG_LOCAL}:latest ${REPO_OWNER}/${DEFAULT_IMAGE_TAG}:latest
  docker push ${REPO_OWNER}/${DEFAULT_IMAGE_TAG}:latest
fi

docker tag ${IMAGE_TAG_LOCAL}:latest ${REPO_OWNER}/${IMAGE_TAG}:${VERSION}
docker push ${REPO_OWNER}/${IMAGE_TAG}:${VERSION}
docker tag ${IMAGE_TAG_LOCAL}:latest ${REPO_OWNER}/${IMAGE_TAG}:latest
docker push ${REPO_OWNER}/${IMAGE_TAG}:latest
