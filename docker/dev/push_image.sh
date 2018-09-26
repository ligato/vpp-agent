#!/bin/bash
# Usage: examples
#    ./push_image.sh
#    BRANCH_HEAD_TAG='git describe' ./push_image.sh
#    REPO_OWNER=stanislavchlebec BRANCH_HEAD_TAG=`git describe` ./push_image.sh
#
# Warning: use only IMMEDIATELY after docker/dev/build.sh to prevent INCONSISTENCIES such as 
#          a) after building image you switch to other branch which will result in mismatch of version of image and its tag
#          b) you do not build the new image but only simply run this script which will result in mismatch version of image and its tag because the image is older than repository 

set -e

# detect branch name
BRANCH_NAME="$(git symbolic-ref HEAD 2>/dev/null)" || BRANCH_NAME="(unnamed branch)"     # detached HEAD
BRANCH_NAME=${BRANCH_NAME##refs/heads/}
BRANCH_HEAD_TAG=${BRANCH_HEAD_TAG:-"`git name-rev --name-only --tags HEAD`"}
VERSION=$(git describe --always --tags --dirty)

REPO_OWNER=${REPO_OWNER:-'ligato'}
LOCAL_IMAGE='dev_vpp_agent'

#To prepare for future fat manifest image by multi-arch manifest,
#now build the docker image with its arch
#For fat manifest, please refer
#https://docs.docker.com/registry/spec/manifest-v2-2/#example-manifest-list

BUILDARCH=`uname -m`

case "$BUILDARCH" in
  "aarch64" )
    IMAGE_NAME='dev-vpp-agent-arm64'
    ;;

  "x86_64" )
    # for AMD64 platform is also used the default image (without suffix -amd64)
    DEFAULT_IMAGE_NAME='dev-vpp-agent'
    IMAGE_NAME='dev-vpp-agent-amd64'
    ;;
  * )
    echo "Architecture ${BUILDARCH} is not supported."
    exit
    ;;
esac


echo "=============================="
echo "Architecture: ${BUILDARCH}"
echo "=============================="

case "${BRANCH_NAME}" in
  "master" )
    if [ ${BRANCH_HEAD_TAG} != "undefined" ] ; then
      if [ ${BUILDARCH} = "x86_64" ] ; then
        # for AMD64 platform is used also the default image (without suffix -amd64)
        docker tag ${LOCAL_IMAGE}:latest ${REPO_OWNER}/${DEFAULT_IMAGE_NAME}:${BRANCH_HEAD_TAG}
        docker push ${REPO_OWNER}/${DEFAULT_IMAGE_NAME}:${BRANCH_HEAD_TAG}
        docker tag ${LOCAL_IMAGE}:latest ${REPO_OWNER}/${DEFAULT_IMAGE_NAME}:latest
        docker push ${REPO_OWNER}/${DEFAULT_IMAGE_NAME}:latest
      fi
      docker tag ${LOCAL_IMAGE}:latest ${REPO_OWNER}/${IMAGE_NAME}:${BRANCH_HEAD_TAG}
      docker push ${REPO_OWNER}/${IMAGE_NAME}:${BRANCH_HEAD_TAG}
      docker tag ${LOCAL_IMAGE}:latest ${REPO_OWNER}/${IMAGE_NAME}:latest
      docker push ${REPO_OWNER}/${IMAGE_NAME}:latest
    else
      echo "For branch ${BRANCH_NAME} is no setup for tagging and pushing docker images because HEAD has no tag."
    fi
    ;;
  "(unnamed branch)" )
    docker tag ${LOCAL_IMAGE}:latest ${REPO_OWNER}/${IMAGE_NAME}:${VERSION}
    echo "Repository is in detached state - please push manually:"
    echo docker push ${REPO_OWNER}/${IMAGE_NAME}:${VERSION}
    ;;
  * )
    docker tag ${LOCAL_IMAGE}:latest ${REPO_OWNER}/${IMAGE_NAME}:${BRANCH_NAME}
    docker push ${REPO_OWNER}/${IMAGE_NAME}:${BRANCH_NAME}
    ;;
esac
