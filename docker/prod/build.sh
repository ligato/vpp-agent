#!/bin/bash

cd "$(dirname "$0")"

set -e

IMAGE_TAG=${IMAGE_TAG:-prod_vpp_agent}

sudo docker build --tag ${IMAGE_TAG} .
