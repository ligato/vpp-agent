#!/bin/bash

set -e

IMAGE_TAG=${IMAGE_TAG:-prod_vpp_agent}

sudo docker build --tag ${IMAGE_TAG} .
