#!/bin/sh
set -e

CACHE_DIR=$HOME/local/vpp
VPP_COMMIT="acd4c63e3c6e70ea3f58527d9bace7c0e38df719"

if [ ! -d "$CACHE_DIR" ]; then
    echo "Building VPP binaries."

    # build VPP
    git clone https://gerrit.fd.io/r/vpp 
    cd vpp 
    git checkout ${VPP_COMMIT} 
    yes | make install-dep     
    make bootstrap     
    make pkg-deb

    # copy deb packages to cache dir
    cp build-root/*.deb $CACHE_DIR
else
    echo "Using cached VPP binaries from $CACHE_DIR"
fi

# install VPP
cd $HOME/local/vpp
dpkg -i vpp_*.deb vpp-dev_*.deb vpp-lib_*.deb vpp-plugins_*.deb
