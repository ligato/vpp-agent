FROM ubuntu:16.04

ARG VPP_URL="xxx"
ARG AGENT_COMMIT="xxx"
ARG VPP_COMMIT="xxx"

RUN apt-get update && apt-get install -y \
    sudo wget git build-essential gdb vim nano python \
    iproute2 iputils-ping inetutils-traceroute libapr1 supervisor \
    default-jre default-jdk telnet netcat \
    golang-gogoprotobuf-dev software-properties-common net-tools

RUN add-apt-repository ppa:ubuntu-toolchain-r/test && \
    apt-get update && apt-get install -y gcc-7

RUN mkdir -p /opt/vpp-agent/dev /opt/vpp-agent/plugin /etc/supervisord

RUN cd /usr/bin/ && rm gcc && ln -s gcc-7 gcc 

WORKDIR /opt/vpp-agent/dev

# build & install VPP
# memif source: git clone -b memif https://github.com/vpp-dev/vpp vpp
RUN /bin/bash -c "if [ '${VPP_URL}' == 'xxx' ]; then \
        git clone https://gerrit.fd.io/r/vpp && \
        cd vpp && \
        git checkout ${VPP_COMMIT} && \
        yes | make install-dep && \
        make bootstrap && \
        make dpdk-install-dev && \
        make build && \
        make pkg-deb && \
        cd build-root && \
        dpkg -i vpp_*.deb vpp-dev_*.deb vpp-lib_*.deb vpp-plugins_*.deb ; \
      else wget $VPP_URL && \
        tar -xvf vpp.tar.gz && \
        dpkg -i vpp/build-root/vpp_*.deb vpp/build-root/vpp-dev_*.deb vpp/build-root/vpp-lib_*.deb vpp/build-root/vpp-plugins_*.deb ; \
    fi"

# install Go & Glide
RUN wget https://storage.googleapis.com/golang/go1.9.linux-amd64.tar.gz && \
    tar -xvf go1.9.linux-amd64.tar.gz && \
    mv go /usr/local && \
    rm -f go1.9.linux-amd64.tar.gz

# build & install Protobuf & gogo protobuf compiler
RUN apt-get install -y make && \
    apt-get install -y autoconf automake libtool curl make g++ unzip

RUN git clone https://github.com/google/protobuf.git && \
    cd protobuf && ./autogen.sh && ./configure && \
    make -j4 && \
    make install && \
    ldconfig && \
    cd .. && rm -rf protobuf

# copy and execute agent build script and agent conf files
COPY build-agent.sh etcd.conf kafka.conf govpp.conf .
RUN ./build-agent.sh ${AGENT_COMMIT}

COPY vpp.conf /etc/vpp/vpp.conf
COPY supervisord_kill.py /usr/bin/supervisord_kill.py

WORKDIR /root/

# add supervisor conf file
COPY supervisord.conf /etc/supervisord/supervisord.conf

# and default plugins conf
COPY defaultplugins.conf linuxplugin.conf logs.conf /opt/vpp-agent/dev/

# run supervisor as the default executable
#CMD ["/usr/bin/supervisord", "-c", "/etc/supervisord/supervisord.conf"]
CMD rm -f /dev/shm/db /dev/shm/global_vm /dev/shm/vpe-api && /usr/bin/supervisord -c /etc/supervisord/supervisord.conf
