#!/bin/bash

set -e

# detect branch name
BRANCH_NAME="$(git symbolic-ref HEAD 2>/dev/null)" || BRANCH_NAME="(unnamed branch)"     # detached HEAD
BRANCH_NAME=${BRANCH_NAME##refs/heads/}
BRANCH_HEAD_TAG=`git name-rev --name-only --tags HEAD`

REPO_OWNER=${REPO_OWNER:-'ligato'}
IMAGE_TAG_LOCAL='dev_vpp_agent'

#To prepare for future fat manifest image by multi-arch manifest,
#now build the docker image with its arch
#For fat manifest, please refer
#https://docs.docker.com/registry/spec/manifest-v2-2/#example-manifest-list

BUILDARCH=`uname -m`

case "$BUILDARCH" in
  "aarch64" )
    IMAGE_TAG='dev-vpp-agent-arm64'
    ;;

  "x86_64" )
    # for AMD64 platform is also used the default image (without suffix -amd64)
    DEFAULT_IMAGE_TAG='dev-vpp-agent'
    IMAGE_TAG='dev-vpp-agent-amd64'
    ;;
  * )
    echo "Architecture ${BUILDARCH} is not supported."
    exit
    ;;
esac

VERSION=$(git describe --always --tags --dirty)

echo "=============================="
echo "Architecture: ${BUILDARCH}"
echo "=============================="

case "${BRANCH_NAME} " in
  "pantheon-dev" )
    docker tag ${IMAGE_TAG_LOCAL}:latest ${REPO_OWNER}/${DEFAULT_IMAGE_TAG}:pantheon-dev
    docker push ${REPO_OWNER}/${DEFAULT_IMAGE_TAG}:pantheon-dev
    ;;
  "master" )
    if [ ${BRANCH_HEAD_TAG} != "undefined" ]	  
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
    fi
    ;;
  * )
    echo "For branch ${BRANCH_NAME} is no setup for tagging and pushing docker images."  
esac
