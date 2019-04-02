#!/bin/bash

#####################################################
# cmpbench - script for comparing benchmark results
#
# args:
# - $1:	path to package that will be benchmarked,
#		optional, defaults to: plugins/kvscheduler
#####################################################

set -euo pipefail

SCRIPT_DIR="$(dirname $(readlink -e "${BASH_SOURCE[0]}"))"

BENCH_PKG_DIR=${1:-${SCRIPT_DIR}/../plugins/kvscheduler}

[ -z ${OUTPUT_DIR-} ] && OUTPUT_DIR=/tmp/bench
mkdir -p "$OUTPUT_DIR"

OLD_BENCH="${OUTPUT_DIR}/old.txt"
NEW_BENCH="${OUTPUT_DIR}/new.txt"
CMP_BENCH="${OUTPUT_DIR}/cmp.txt"
CMP_BENCH_IMG="${OUTPUT_DIR}/cmp.svg"

function benchtest() {
	cd "$BENCH_PKG_DIR"
	go test -run=NONE -bench=. -benchmem -benchtime=3s
}

BENCH_PKG=$(go list $BENCH_PKG_DIR)

echo "-> running cmpbench for package: $BENCH_PKG"

[ -f "${OLD_BENCH}" ] && grep -q "pkg: $BENCH_PKG" $OLD_BENCH || {
	rm -f $OLD_BENCH $NEW_BENCH $CMP_BENCH $CMP_BENCH_IMG
	echo "-> creating old bench.."
	echo "--------------------------------------------------------------"
	benchtest | tee $OLD_BENCH
	echo "--------------------------------------------------------------"
	echo "old bench at: $OLD_BENCH"
	echo "run this script again for comparing with new benchmark"
	exit 0
}

echo "-> found old bench at: $OLD_BENCH"
echo "--------------------------------------------------------------"
cat $OLD_BENCH
echo "--------------------------------------------------------------"

echo "-> creating new bench.."
echo "--------------------------------------------------------------"
benchtest | tee $NEW_BENCH
echo "--------------------------------------------------------------"
echo "-> new bench at: $OLD_BENCH"

echo "-> comparing benchmarks.."

if ! which benchcmp >/dev/null; then
	echo "-> downloading benchcmp.."
	go get golang.org/x/tools/cmd/benchcmp
fi

if ! which benchviz >/dev/null; then
	echo "-> downloading benchviz.."
	go get github.com/ajstarks/svgo/benchviz
fi

echo "--------------------------------------------------------------"
benchcmp $OLD_BENCH $NEW_BENCH | tee $CMP_BENCH
echo "--------------------------------------------------------------"

cat $CMP_BENCH | benchviz -title="$BENCH_PKG" > $CMP_BENCH_IMG
echo "-> bench comparison image at: $CMP_BENCH_IMG"

if [ -t 1 ]; then exo-open $CMP_BENCH_IMG; fi

