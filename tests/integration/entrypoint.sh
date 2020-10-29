#!/bin/bash
set -e

exec \
  gotestsum --raw-command --format testname -- \
  test2json -t -p "integration" \
  /integration.test -test.v "$@"
