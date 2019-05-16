ARG DEV_IMG=dev_vpp_agent
FROM ${DEV_IMG} as devimg

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
		python \
		supervisor \
 && rm -rf /var/lib/apt/lists/*

# install vpp
COPY --from=devimg \
    /vpp/libvppinfra_*.deb \
    /vpp/vpp-plugin-core_*.deb \
    /vpp/vpp-plugin-dpdk_*.deb \
    /vpp/vpp_*.deb \
 /opt/vpp/

RUN cd /opt/vpp/ \
 && dpkg -i *.deb \
 && rm *.deb

# Final image
FROM scratch
COPY --from=base / /

# Install agent
COPY --from=devimg \
    /go/bin/agentctl \
    /go/bin/vpp-agent \
 /bin/

# Copy configs
COPY etcd.conf /opt/vpp-agent/dev/
COPY vpp.conf /etc/vpp/vpp.conf
COPY supervisord.conf /etc/supervisord/supervisord.conf
COPY supervisord_kill.py /usr/bin/

WORKDIR /root/

CMD rm -f /dev/shm/db /dev/shm/global_vm /dev/shm/vpe-api && \
    exec /usr/bin/supervisord -c /etc/supervisord/supervisord.conf
