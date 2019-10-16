#!/usr/bin/env bash

if [ -z $GOPATH ] ; then
    echo "GOPATH must be set"
    exit -1
fi

GODIR=$( echo $GOPATH | sed -e 's/:.*//' )
if [ $GODIR != $GOPATH ] ; then
        echo "Warning: Using $GODIR for GOPATH"
fi

set -euo pipefail

vpp_agent_dir=$PWD
api_dir=${vpp_agent_dir}/api

pb_pkg=github.com/golang/protobuf
pb_pkg_dir=$( go list -f {{.Dir}} -m ${pb_pkg} )

go_mod_dir="${GODIR}/pkg/mod"

# FIXME: And examples/ and pkg/ too?
protos=$(find api plugins -type f -name '*.proto')

for proto in $protos; do
	echo " - ${proto}";
	protoc \
		--proto_path=. \
		--proto_path=${go_mod_dir} \
		--proto_path=${api_dir} \
		--proto_path=${pb_pkg_dir} \
		--go_out=plugins=grpc,\
Mgoogle/protobuf/any.proto=${pb_pkg}/ptypes/any,paths=source_relative:. \
		"${proto}";
done
