#!/bin/sh
set -e

VPP_CACHE_DIR=$HOME/build-cache/vpp

VPP_COMMIT="d95c39e87bf9d21b2a9d4c49fdf7ebca2a5eab3d"
VPP_IMG_TAG=`echo ${VPP_COMMIT} | cut -c1-7`

# check if cache folder contains same version
if [ -d "$VPP_CACHE_DIR" ] && [ ! -f ${VPP_CACHE_DIR}/vpp_*${VPP_IMG_TAG}*.deb ]; then
    echo "Removing cached VPP folder with different version"
    rm -rf "$VPP_CACHE_DIR"
fi

if [ ! -d "$VPP_CACHE_DIR" ]; then
    echo "Pulling VPP binaries (commit ${VPP_IMG_TAG})"

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
    echo "Using cached VPP binaries (commit ${VPP_IMG_TAG})"
fi

# install VPP
cd $VPP_CACHE_DIR
dpkg -i vpp_*.deb vpp-dev_*.deb vpp-lib_*.deb vpp-plugins_*.deb
