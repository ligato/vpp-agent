#!/usr/bin/env bash

set -euo pipefail

API_DIR=${1:-`pwd`/api}

protos=$(find "$API_DIR" -type f -name '*.proto')

for proto in $protos; do
	echo " - $proto";
	protoc \
		--proto_path=. \
		--proto_path=${API_DIR}/models \
		--proto_path=${API_DIR} \
		--proto_path=vendor/ \
		--proto_path=$GOPATH/src/github.com/gogo/protobuf \
		--proto_path=$GOPATH/src \
		--gogo_out=plugins=grpc,\
Mgoogle/protobuf/any.proto=github.com/gogo/protobuf/types:$GOPATH/src \
		"$proto";
done
