#!/usr/bin/env bash

set -eo pipefail

fail() {
  echo "$@" >&2
  exit 1
}

if [ -z "${1}" ] || [ -z "${2}" ]; then
  fail "usage: ${0} <proto_path> <protoc_gen_go_out> [protoc_gen_go_parameter]"
fi

protoc --version

PROTO_PATH="${1}"
PROTOC_GEN_GO_OUT="${2}"
PROTOC_GEN_GO_PARAMETER="${3}"

PROTOC_GEN_GO_ARGS="${PROTOC_GEN_GO_OUT}"
if [ -n "${PROTOC_GEN_GO_PARAMETER}" ]; then
  PROTOC_GEN_GO_ARGS="${PROTOC_GEN_GO_PARAMETER}:${PROTOC_GEN_GO_ARGS}"
fi

mkdir -p "${PROTOC_GEN_GO_OUT}"

# all directories with proto files
protodirs=$(find "${PROTO_PATH}" -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
for dir in $protodirs; do
	protofiles=$(find "${dir}" -maxdepth 1 -name '*.proto')
	echo "$protofiles"

  	protoc --proto_path="${PROTO_PATH}" \
  		--go_out="${PROTOC_GEN_GO_ARGS}" $protofiles \
  		--go-grpc_out="${PROTOC_GEN_GO_ARGS}" $protofiles
done
