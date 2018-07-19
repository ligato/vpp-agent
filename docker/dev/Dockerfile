ARG BASE_IMG=ubuntu:16.04

FROM golang as verify-stage

RUN apt-get update \
 && apt-get install -y --no-install-recommends \
    patch python-pip \
 && rm -rf /var/lib/apt/lists/* \
 && pip install ply

ARG VPP_REPO_URL
ARG VPP_COMMIT

RUN git clone ${VPP_REPO_URL} /opt/vpp \
 && cd /opt/vpp \
 && git checkout ${VPP_COMMIT}

COPY scripts/genjsonapi.sh /opt/genjsonapi.sh
RUN /opt/genjsonapi.sh

WORKDIR /go/src/github.com/ligato/vpp-agent

COPY plugins/vpp/binapi plugins/vpp/binapi
COPY vendor vendor
COPY Makefile vpp.env ./

RUN cp -r plugins/vpp/binapi /tmp/orig_binapi \
 && make generate-binapi \
 && diff -r plugins/vpp/binapi /tmp/orig_binapi

FROM ${BASE_IMG} as dev-stage

RUN apt-get update \
 && apt-get install -y --no-install-recommends \
    autoconf automake build-essential ca-certificates curl gdb git \
    inetutils-traceroute iproute2 ipsec-tools iputils-ping \
    libapr1 libmbedcrypto0 libmbedtls10 libmbedx509-0 libtool \
    make mc nano netcat python software-properties-common sudo supervisor \
    telnet unzip wget \
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

ARG VPP_REPO_URL=https://gerrit.fd.io/r/vpp.git
ARG VPP_COMMIT

ARG VPP_DEBUG_DEB

RUN set -eux; \
    git clone "${VPP_REPO_URL}" vpp; \
    cd vpp; \
    git checkout "${VPP_COMMIT}"; \
    export UNATTENDED=y; \
    make install-dep dpdk-install-dev; \
    if [ -n "${VPP_DEBUG_DEB}" ]; then \
        make vpp_configure_args_vpp='--disable-japi' build; \
        make -C build-root PLATFORM=vpp TAG=vpp_debug install-deb; \
    else \
        make vpp_configure_args_vpp='--disable-japi' build-release pkg-deb; \
    fi; \
    cd build-root; \
    dpkg -i vpp_*.deb vpp-dev_*.deb vpp-lib_*.deb vpp-plugins_*.deb vpp-dbg_*.deb; \
    rm -rf .ccache /var/lib/apt/lists/*; \
    find . -type f -name '*.o' -exec rm -rf '{}' \;

# install Go
ENV GOLANG_VERSION 1.10.3
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
