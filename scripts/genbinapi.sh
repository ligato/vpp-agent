#!/bin/bash

set -euo pipefail

go generate -x ./${VPP_BINAPI}

find ${VPP_BINAPI} -maxdepth 2 -type f -name '*.patch' -exec \
	patch --no-backup-if-mismatch -p1 -i {} \;
