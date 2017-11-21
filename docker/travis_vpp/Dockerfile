FROM ubuntu:14.04 AS builder

RUN apt-get update && apt-get install -y \
    sudo git wget git build-essential vim python libapr1 automake autoconf

WORKDIR /root
RUN wget http://www.nasm.us/pub/nasm/releasebuilds/2.12.01/nasm-2.12.01.tar.bz2 &&\
    tar xfj nasm-2.12.01.tar.bz2 && cd nasm-2.12.01/ &&\
    ./autogen.sh && ./configure --prefix=/usr/local/ &&\
    make && make install &&\
    cd .. && rm -rf nasm-*

COPY travis.patch .

ARG VPP_COMMIT
RUN test -n "$VPP_COMMIT"

RUN git clone https://gerrit.fd.io/r/vpp

WORKDIR vpp
RUN git checkout $VPP_COMMIT
RUN git apply -C1 ../travis.patch

RUN (yes | make install-dep) && make bootstrap
RUN make pkg-deb

RUN cd build-root && tar -cvf \
    /root/vpp.tar vpp_*.deb vpp-dev_*.deb vpp-lib_*.deb vpp-plugins_*.deb

FROM alpine

WORKDIR vpp-deb
COPY --from=builder /root/vpp.tar .
