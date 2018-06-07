FROM ubuntu:16.04

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
ENV GOLANG_VERSION 1.10
RUN wget -O go.tgz "https://golang.org/dl/go${GOLANG_VERSION}.linux-amd64.tar.gz" \
 && tar -C /usr/local -xzf go.tgz \
 && rm go.tgz

ENV PATH /usr/local/go/bin:$PATH

# build & install Protobuf & gogo protobuf compiler
RUN git clone https://github.com/google/protobuf.git \
 && cd protobuf \
 && ./autogen.sh \
 && ./configure \
 && make -j4 \
 && make install \
 && ldconfig \
 && cd .. \
 && rm -rf protobuf

ARG AGENT_COMMIT

# copy and execute agent build script
COPY build-agent.sh .
RUN ./build-agent.sh ${AGENT_COMMIT}

# copy configs
COPY etcd.conf kafka.conf govpp.conf defaultplugins.conf linuxplugin.conf logs.conf ./
COPY vpp.conf /etc/vpp/vpp.conf
COPY supervisord.conf /etc/supervisord/supervisord.conf

# copy scripts
COPY runvpp.sh /usr/bin/runvpp
COPY agent_start.py /usr/bin/agent_start.py
COPY supervisord_kill.py /usr/bin/supervisord_kill.py

WORKDIR /root/

# start agent
ENV START_AGENT "True"
# kill supervisor if vpp or agent is closed
ENV RETAIN_SUPERVISOR "False"

# run supervisor as the default executable
CMD rm -f /dev/shm/db /dev/shm/global_vm /dev/shm/vpe-api && \
    /usr/bin/supervisord -c /etc/supervisord/supervisord.conf
