FROM dev_vpp_agent as devimg

FROM ubuntu:18.04 as base

RUN apt-get update \
 && apt-get install -y --no-install-recommends \
     # general tools
     inetutils-traceroute \
     iproute2 \
     iputils-ping \
     # vpp requirements
     ca-certificates \
     libapr1 \
     libc6 \
     libmbedcrypto1 \
     libmbedtls10 \
     libmbedx509-0 \
     libnuma1 \
     openssl \
     # other
     ipsec-tools \
     python \
     supervisor \
     netcat \
 && rm -rf /var/lib/apt/lists/*

# install vpp
COPY --from=devimg \
    /opt/vpp-agent/dev/vpp/build-root/libvppinfra_*.deb \
    /opt/vpp-agent/dev/vpp/build-root/vpp-plugin-core_*.deb \
    /opt/vpp-agent/dev/vpp/build-root/vpp-plugin-dpdk_*.deb \
    /opt/vpp-agent/dev/vpp/build-root/vpp_*.deb \
 /opt/vpp/

RUN cd /opt/vpp/ \
 && dpkg -i *.deb \
 && rm *.deb

FROM scratch
COPY --from=base / /

# install agent
COPY --from=devimg \
    /go/bin/vpp-agent \
    /go/bin/vpp-agent-ctl \
 /bin/

# copy configs
COPY \
    etcd.conf \
    vpp-ifplugin.conf \
    linux-ifplugin.conf \
 /opt/vpp-agent/dev/

COPY vpp.conf /etc/vpp/vpp.conf
COPY supervisord.conf /etc/supervisord/supervisord.conf

# copy scripts
COPY \
    exec_agent.sh \
    supervisord_kill.py \
 /usr/bin/

WORKDIR /root/

# run supervisor as the default executable
CMD rm -f /dev/shm/db /dev/shm/global_vm /dev/shm/vpe-api && \
    exec /usr/bin/supervisord -c /etc/supervisord/supervisord.conf
