FROM ubuntu:16.04

RUN apt-get update && apt-get install -y \
    sudo wget git build-essential gdb vim nano python \
    iproute2 iputils-ping inetutils-traceroute libapr1 supervisor \
    default-jre default-jdk telnet netcat software-properties-common

RUN add-apt-repository -y ppa:ubuntu-toolchain-r/test && \
    apt-get update && apt-get install -y gcc-7 && \
    cd /usr/bin/ && rm gcc && ln -s gcc-7 gcc

RUN mkdir -p /opt/vpp-agent/dev /opt/vpp-agent/plugin

WORKDIR /opt/vpp-agent/dev

ARG VPP_URL
ARG VPP_COMMIT

# build & install VPP
# memif source: git clone -b memif https://github.com/vpp-dev/vpp vpp
RUN /bin/bash -c "\
    if [ -n '${VPP_URL}' ]; then \
        wget ${VPP_URL} && \
        tar -xvf vpp.tar.gz && \
        cd vpp/build-root; \
    elif [ -n '${VPP_COMMIT}' ]; then \
        git clone https://gerrit.fd.io/r/vpp && \
        cd vpp && \
        git checkout ${VPP_COMMIT} && \
        yes | make install-dep && \
        make bootstrap && \
        make dpdk-install-dev && \
        make build && \
        make pkg-deb && \
        cd build-root && \
        rm -rf .ccache; \
    fi; \
    dpkg -i vpp_*.deb vpp-dev_*.deb vpp-lib_*.deb vpp-plugins_*.deb vpp-dbg_*.deb"

# install Go
ENV GOLANG_VERSION 1.9.3
RUN wget -O go.tgz "https://golang.org/dl/go$GOLANG_VERSION.linux-amd64.tar.gz" && \
    tar -C /usr/local -xzf go.tgz && \
    rm go.tgz

# build & install Protobuf & gogo protobuf compiler
RUN apt-get install -y \
    make autoconf automake libtool curl make g++ unzip \
    golang-gogoprotobuf-dev

RUN git clone https://github.com/google/protobuf.git && \
    cd protobuf && ./autogen.sh && ./configure && \
    make -j4 && \
    make install && \
    ldconfig && \
    cd .. && rm -rf protobuf

ARG AGENT_COMMIT

# copy and execute agent build script
COPY build-agent.sh .
RUN ./build-agent.sh ${AGENT_COMMIT}

# copy configs
COPY etcd.conf kafka.conf govpp.conf defaultplugins.conf linuxplugin.conf logs.conf ./
COPY vpp.conf /etc/vpp/vpp.conf

COPY supervisord.conf /etc/supervisord/supervisord.conf
COPY supervisord_kill.py /usr/bin/supervisord_kill.py

WORKDIR /root/

# run supervisor as the default executable
CMD rm -f /dev/shm/db /dev/shm/global_vm /dev/shm/vpe-api && \
    /usr/bin/supervisord -c /etc/supervisord/supervisord.conf
