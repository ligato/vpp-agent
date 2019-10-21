#!/usr/bin/env bash
set -euo pipefail

# mode can be set to:
#  - 'check' -> fails on first difference
#
mode=${1-}

# Create tempdir
tmpdir=$(mktemp -d -t gen-proto.XXXXXX)
trap 'rm -rf $tmpdir' EXIT

# Install the working tree's protoc-gen-gen
mkdir -p $tmpdir/bin
PATH=$tmpdir/bin:$PATH
GOBIN=$tmpdir/bin go install ./vendor/github.com/golang/protobuf/protoc-gen-go

# search for proto files in directories:
PROTO_DIRS=(
  api
  plugins
)

different=0

echo "# Generating proto files.."
for dir in ${PROTO_DIRS[@]}; do
	echo "# $dir/"
	for proto in `find $dir -name "*.proto"`; do
		pb=$(echo $proto | sed -e 's/\.proto$/\.pb.go/')
		echo "# - $proto"
		mkdir -p $tmpdir/$dir

		protoc -I$dir -I../../ \
			--go_out=plugins=grpc,paths=source_relative,\
Mgoogle/protobuf/any.proto=github.com/golang/protobuf/ptypes/any,\
Mvppagent/api/vpp.proto=github.com/ligato/vpp-agent/api/models/vpp\
:$tmpdir/$dir $proto

		diff $tmpdir/$pb $pb || different=1
		if [ "$mode" != "check" ]; then
		 	cp $tmpdir/$pb $pb
		fi
  	done
done


if [ "$mode" == "check" ]; then
	[ $different -eq 1 ] \
		&& echo "Check failed! Some generated proto files are different." \
		|| echo "Check OK!"
	exit $different
else
	tree $tmpdir
fi
