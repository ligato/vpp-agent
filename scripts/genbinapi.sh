#!/bin/bash
set -euo pipefail

export VPP_API_DIR=${VPP_API_DIR:-/usr/share/vpp/api}
export VPP_VERSION=

# Generate binapi
go generate -x ./"${VPP_BINAPI}"

# Apply patches
find "${VPP_BINAPI}" -maxdepth 2 -type f -name '*.patch' -exec \
	patch --no-backup-if-mismatch -p1 -i {} \;
