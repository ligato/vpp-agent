#!/usr/bin/env bash
set -euo pipefail

# mode can be set to:
#  - 'check' -> fails on first difference
#
mode=${1-}

curdir=$(pwd)

# Create tempdir
tmpdir=$(mktemp -d -t gen-proto.XXXXXX)
trap 'rm -rf $tmpdir' EXIT

# Install the protoc-gen-go
mkdir -p $tmpdir/bin
PATH=$tmpdir/bin:$PATH
GOBIN=$tmpdir/bin go install github.com/golang/protobuf/protoc-gen-go

pb_pkg_dir=$( go list -f '{{.Dir}}' -m github.com/golang/protobuf )

#rm -rf $tmpdir/src
#mkdir -p $tmpdir/src/ligato
#ln -s $curdir $tmpdir/src/ligato/vpp-agent

#cd $tmpdir/src
cd proto

different=0

echo "# Generating proto files.."
for proto in `find . -name "*.proto"`; do
	echo "# - $proto"

	protoc -I. \
		--go_out=plugins=grpc,paths=source_relative,\
Mgoogle/protobuf/any.proto=github.com/golang/protobuf/ptypes/any,\
:$tmpdir $proto

	pb=$(echo $proto | sed -e 's/\.proto$/\.pb.go/')
	diff $tmpdir/$pb $pb || different=$(( different+1 ))

	if [ "$mode" != "check" ]; then
		cp $tmpdir/$pb $pb
	fi
done

if [ "$mode" == "check" ]; then
	if [ $different -gt 1 ]; then
		echo "Check failed! $different generated proto files are not up to date."
		exit 1
	else
		echo "Check OK!"
	fi
fi
