#!/bin/sh
set -e

CACHE_DIR=$HOME/build-cache
VPP_COMMIT="acd4c63e3c6e70ea3f58527d9bace7c0e38df719"

if [ ! -d "$CACHE_DIR" ]; then
    echo "Building VPP binaries."
    mkdir $CACHE_DIR
    cd $CACHE_DIR

    # build VPP
    git clone https://gerrit.fd.io/r/vpp 
    cd vpp 
    git checkout ${VPP_COMMIT} 
    yes | make install-dep     
    make bootstrap     
    make pkg-deb     
else
    echo "Using cached VPP binaries from $CACHE_DIR"
fi

# install VPP
cd $CACHE_DIR/vpp/build-root
dpkg -i vpp_*.deb vpp-dev_*.deb vpp-lib_*.deb vpp-plugins_*.deb

