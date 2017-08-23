#!/bin/sh
set -e

VPP_CACHE_DIR=$HOME/build-cache/vpp
VPP_COMMIT="5f22f4ddded8ac41487dab3069ff8d77c3916205"
WORKDIR="$(pwd)"

if [ ! -d "$VPP_CACHE_DIR" ]; then
    echo "Building VPP binaries."

    date
    echo "!! start"
    # latest vpp requires newer NASM
    wget http://www.nasm.us/pub/nasm/releasebuilds/2.12.01/nasm-2.12.01.tar.bz2
    tar xfj nasm-2.12.01.tar.bz2

    date
    echo "!! nasm src cloned"
    cd nasm-2.12.01/
    ./autogen.sh
    ./configure --prefix=/usr/local/
    make
    make install
    cd ..
    date
    echo "!! nasm installed"

    # build VPP
    git clone https://gerrit.fd.io/r/vpp /tmp/vpp
    cd /tmp/vpp
    git checkout ${VPP_COMMIT}

    date
    echo "!! vpp checked out"
    yes | make install-dep
    make bootstrap
    make pkg-deb
    cd ${WORKDIR}

    # copy deb packages to cache dir
    mkdir $VPP_CACHE_DIR
    cp /tmp/vpp/build-root/*.deb $VPP_CACHE_DIR
else
    echo "Using cached VPP binaries from $VPP_CACHE_DIR"
fi

# install VPP
cd $VPP_CACHE_DIR
dpkg -i vpp_*.deb vpp-dev_*.deb vpp-lib_*.deb vpp-plugins_*.deb
