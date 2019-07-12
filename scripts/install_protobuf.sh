#!/bin/bash

set -e

PROTOC_VERSION=3.7.1
PROTOC_OS_ARCH=linux-x86_64

download_url="https://github.com/google/protobuf/releases/download/v${PROTOC_VERSION}/protoc-${PROTOC_VERSION}-${PROTOC_OS_ARCH}.zip"

wget -O /tmp/protoc.zip ${download_url}
unzip -o /tmp/protoc.zip -d /tmp/protoc3
sudo mv /tmp/protoc3/bin/protoc /usr/local/bin
sudo mv /tmp/protoc3/include/google /usr/local/include
rm -rf /tmp/protoc.zip /tmp/protoc3
