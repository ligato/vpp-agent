#!/bin/bash

#set -x

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

cd ${SCRIPT_DIR}/grpc-perf

source "./test.sh"

run_test 500	
run_test 1000
run_test 2000
run_test 4000
run_test 8000

