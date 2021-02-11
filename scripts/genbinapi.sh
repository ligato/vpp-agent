#!/bin/bash
set -euo pipefail

export VPP_API_DIR=${VPP_API_DIR:-/usr/share/vpp/api}
export VPP_VERSION=

binapidir=$(basename "$VPP_BINAPI")
if [ "$binapidir" == "vpp2001" ]
then
    (
     cd $(mktemp -d)
     echo "module x" > go.mod
     GO111MODULE=on go get -v git.fd.io/govpp.git/cmd/binapi-generator@v0.3.5
    )
fi

# Generate binapi
go generate -x ./"${VPP_BINAPI}"

# Apply patches
find "${VPP_BINAPI}" -maxdepth 2 -type f -name '*.patch' -exec \
	patch --no-backup-if-mismatch -p1 -i {} \;
