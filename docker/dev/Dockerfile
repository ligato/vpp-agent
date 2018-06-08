ARG BASE_IMG=ubuntu:16.04

FROM golang as verify

RUN apt-get update \
 && apt-get install -y --no-install-recommends \
    python-pip patch \
 && rm -rf /var/lib/apt/lists/* \
 && pip install ply

ARG VPP_REPO_URL
ARG VPP_COMMIT

RUN /bin/bash -c "\
    git clone ${VPP_REPO_URL} /opt/vpp \
    && cd /opt/vpp \
    && git checkout ${VPP_COMMIT}"

COPY scripts/genjsonapi.sh /opt/genjsonapi.sh
RUN VPPDIR=/opt/vpp /opt/genjsonapi.sh

WORKDIR /go/src/github.com/ligato/vpp-agent

COPY plugins/vpp/binapi plugins/vpp/binapi
COPY vendor vendor
COPY Makefile vpp.env ./

RUN cp -r plugins/vpp/binapi /tmp/binapi \
 && make generate-binapi \
 && diff -r plugins/vpp/binapi /tmp/binapi

FROM ${BASE_IMG} as builder

RUN apt-get update \
 && apt-get install -y --no-install-recommends \
    sudo wget git build-essential gdb vim nano python \
    iproute2 iputils-ping inetutils-traceroute libapr1 supervisor \
    default-jre default-jdk telnet netcat software-properties-common \
    libmbedcrypto0 libmbedtls10 libmbedx509-0 ca-certificates \
    make autoconf automake libtool curl unzip ipsec-tools \
 && apt-get remove -y --purge gcc \
 && rm -rf /var/lib/apt/lists/*

RUN add-apt-repository -y ppa:ubuntu-toolchain-r/test \
 && apt-get update \
 && apt-get install -y gcc-7 g++-7 \
 && cd /usr/bin/ \
 && ln -s gcc-7 gcc \
 && ln -s g++-7 g++ \
 && rm -rf /var/lib/apt/lists/*

# build & install Protobuf
RUN git clone --depth=1 https://github.com/google/protobuf.git \
 && cd protobuf \
 && ./autogen.sh \
 && ./configure \
 && make -j4 \
 && make install \
 && ldconfig \
 && cd .. \
 && rm -rf protobuf

RUN mkdir -p /opt/vpp-agent/dev /opt/vpp-agent/plugin

WORKDIR /opt/vpp-agent/dev

ARG VPP_COMMIT
ARG VPP_REPO_URL

RUN /bin/bash -c "\
    git clone ${VPP_REPO_URL} \
    && cd vpp \
    && git checkout ${VPP_COMMIT} \
    && UNATTENDED=y make vpp_configure_args_vpp='--disable-japi --disable-vom' install-dep bootstrap dpdk-install-dev build build-release;"

ARG VPP_DEBUG_DEB

RUN /bin/bash -c "\
    cd vpp; \
    if [ -n '${VPP_DEBUG_DEB}' ]; then \
        cd build-root \
        && make PLATFORM=vpp TAG=vpp_debug install-deb; \
    else \
        make pkg-deb && \
        cd build-root; \
    fi; \
    dpkg -i vpp_*.deb vpp-dev_*.deb vpp-lib_*.deb vpp-plugins_*.deb vpp-dbg_*.deb;"

# install Go
ENV GOLANG_VERSION 1.10.2
ARG GOLANG_OS_ARCH=linux-amd64
RUN wget -O go.tgz "https://golang.org/dl/go${GOLANG_VERSION}.${GOLANG_OS_ARCH}.tar.gz" \
 && tar -C /usr/local -xzf go.tgz \
 && rm go.tgz

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" \
 && chmod -R 777 "$GOPATH"

# copy configs
COPY \
    docker/dev/etcd.conf \
    docker/dev/kafka.conf \
    docker/dev/govpp.conf \
    docker/dev/vpp-plugin.conf \
    docker/dev/linux-plugin.conf \
    docker/dev/logs.conf \
 ./

COPY docker/dev/vpp.conf /etc/vpp/vpp.conf
COPY docker/dev/supervisord.conf /etc/supervisord/supervisord.conf

# copy scripts
COPY \
    docker/dev/exec_vpp.sh \
    docker/dev/exec_agent.sh \
    docker/dev/supervisord_kill.py \
 /usr/bin/

ARG VERSION
ARG COMMIT

# copy & build agent
COPY . $GOPATH/src/github.com/ligato/vpp-agent

RUN cd $GOPATH/src/github.com/ligato/vpp-agent \
 && VERSION=$VERSION COMMIT=$COMMIT make install

WORKDIR /

# run supervisor as the default executable
CMD rm -f /dev/shm/db /dev/shm/global_vm /dev/shm/vpe-api && \
    /usr/bin/supervisord -c /etc/supervisord/supervisord.conf
