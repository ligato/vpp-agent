FROM ubuntu:16.04

ARG VPP_URL
ARG VPP_COMMIT

RUN apt-get update && apt-get install -y \
    sudo wget git build-essential gdb vim nano python \
    iproute2 iputils-ping inetutils-traceroute libapr1 supervisor \
    default-jre default-jdk telnet netcat software-properties-common

RUN add-apt-repository -y ppa:ubuntu-toolchain-r/test && \
    apt-get update && apt-get install -y gcc-7 && \
    cd /usr/bin/ && rm gcc && ln -s gcc-7 gcc

RUN mkdir -p /opt/vpp-agent/dev /opt/vpp-agent/plugin

WORKDIR /opt/vpp-agent/dev

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
    dpkg -i vpp_*.deb vpp-dev_*.deb vpp-lib_*.deb vpp-plugins_*.deb"

# install Go & Glide
RUN wget https://storage.googleapis.com/golang/go1.9.2.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.9.2.linux-amd64.tar.gz && \
    rm -f go1.9.2.linux-amd64.tar.gz

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
COPY etcd.conf .
COPY kafka.conf .
COPY govpp.conf .
COPY defaultplugins.conf .
COPY linuxplugin.conf .
COPY logs.conf .

COPY vpp.conf /etc/vpp/vpp.conf
COPY supervisord.conf /etc/supervisord/supervisord.conf

COPY supervisord_kill.py /usr/bin/supervisord_kill.py

WORKDIR /root/

# run supervisor as the default executable
CMD rm -f /dev/shm/db /dev/shm/global_vm /dev/shm/vpe-api && \
    /usr/bin/supervisord -c /etc/supervisord/supervisord.conf
