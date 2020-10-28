#!/bin/bash
set -e

exec \
  gotestsum --raw-command --format testname -- \
  test2json -t -p "e2e" \
  /e2e.test -test.v "$@"
