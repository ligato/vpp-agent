FROM golang:1.8

# libseccomp in jessie is not _quite_ new enough -- need backports version
RUN echo 'deb http://httpredir.debian.org/debian jessie-backports main' > /etc/apt/sources.list.d/backports.list

RUN apt-get update && apt-get install -y \
    build-essential \
    curl \
    sudo \
    gawk \
    iptables \
    jq \
    pkg-config \
    libaio-dev \
    libcap-dev \
    libprotobuf-dev \
    libprotobuf-c0-dev \
    libnl-3-dev \
    libnet-dev \
    libseccomp2/jessie-backports \
    libseccomp-dev/jessie-backports \
    protobuf-c-compiler \
    protobuf-compiler \
    python-minimal \
    uidmap \
    --no-install-recommends \
    && apt-get clean

# Add a dummy user for the rootless integration tests. While runC does
# not require an entry in /etc/passwd to operate, one of the tests uses
# `git clone` -- and `git clone` does not allow you to clone a
# repository if the current uid does not have an entry in /etc/passwd.
RUN useradd -u1000 -m -d/home/rootless -s/bin/bash rootless

# install bats
RUN cd /tmp \
    && git clone https://github.com/sstephenson/bats.git \
    && cd bats \
    && git reset --hard 03608115df2071fff4eaaff1605768c275e5f81f \
    && ./install.sh /usr/local \
    && rm -rf /tmp/bats

# install criu
ENV CRIU_VERSION 3ca8e575b49763030d3ddfec4af190a4c9f9deef
RUN mkdir -p /usr/src/criu \
    && curl -sSL https://github.com/xemul/criu/archive/${CRIU_VERSION}.tar.gz | tar -v -C /usr/src/criu/ -xz --strip-components=1 \
    && cd /usr/src/criu \
    && make install-criu \
    && rm -rf /usr/src/criu

# setup a playground for us to spawn containers in
ENV ROOTFS /busybox
RUN mkdir -p ${ROOTFS}

COPY script/tmpmount /
WORKDIR /go/src/github.com/opencontainers/runc
ENTRYPOINT ["/tmpmount"]

ADD . /go/src/github.com/opencontainers/runc

RUN . tests/integration/multi-arch.bash \
    && curl -o- -sSL `get_busybox` | tar xfJC - ${ROOTFS}
