#!/bin/bash
set -e

exec \
  gotestsum --raw-command -- \
  test2json -t -p "e2e" \
  /e2e.test -test.v "$@"
