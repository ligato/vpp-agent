#!/bin/bash
set -e

exec \
  gotestsum --raw-command -- \
  test2json -t -p "integration" \
  /integration.test -test.v "$@"
