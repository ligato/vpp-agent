#!/bin/sh
set -e

VPP_CACHE_DIR=$HOME/build-cache/vpp
VPP_COMMIT="ce41a5c8032ff581a56b990361a2368c879a7adf"
VPP_IMG_TAG="ce41a5c"

if [ ! -d "$VPP_CACHE_DIR" ]; then
    echo "Building VPP binaries."

    IMG_NAME=ligato/vppdeb:${VPP_IMG_TAG}
    docker pull ${IMG_NAME}
    id=$(docker create ${IMG_NAME})
    docker cp $id:/vpp-deb/vpp.tar .
    docker rm -v $id

    # copy deb packages to cache dir
    mkdir $VPP_CACHE_DIR
    tar -xvf vpp.tar -C $VPP_CACHE_DIR

#    # latest vpp requires newer NASM
#    wget http://www.nasm.us/pub/nasm/releasebuilds/2.12.01/nasm-2.12.01.tar.bz2
#    tar xfj nasm-2.12.01.tar.bz2
#
#    cd nasm-2.12.01/
#    ./autogen.sh
#    ./configure --prefix=/usr/local/
#    make
#    make install
#    cd ..
#
#    # build VPP
#    git clone https://gerrit.fd.io/r/vpp /tmp/vpp
#    cd /tmp/vpp
#    git checkout ${VPP_COMMIT}
#
#    yes | make install-dep
#    make bootstrap
#    make pkg-deb
#    cd ${WORKDIR}
#
#    # copy deb packages to cache dir
#    mkdir $VPP_CACHE_DIR
#    cp /tmp/vpp/build-root/*.deb $VPP_CACHE_DIR
else
    echo "Using cached VPP binaries from $VPP_CACHE_DIR"
fi

# install VPP
cd $VPP_CACHE_DIR
dpkg -i vpp_*.deb vpp-dev_*.deb vpp-lib_*.deb vpp-plugins_*.deb
