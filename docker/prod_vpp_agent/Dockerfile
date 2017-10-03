FROM ubuntu:16.04

MAINTAINER jozef.mak@pantheon.tech

RUN apt-get update && apt-get install -y \
        # general tools
        iproute2 iputils-ping inetutils-traceroute \
        # vpp requirements
        openssl python libapr1 libnuma1 \
        supervisor

# use /opt/vpp-agent/dev as the working directory
RUN mkdir -p /opt/vpp-agent/dev
WORKDIR /opt/vpp-agent/dev

# copy vpp deb packages and agent
# use ADD instead of COPY, because ADD will uncompress automatically
ADD vpp.tar.gz .
ADD agent.tar.gz .

# install vpp deb packages
RUN dpkg -i vpp/build-root/vpp-lib_*.deb vpp/build-root/vpp_*.deb vpp/build-root/vpp-plugins_*.deb

# install agent
RUN mv agent/vpp-agent agent/vpp-agent-ctl agent/agentctl /bin

# remove packages
RUN rm -rf vpp/ && rm -rf agent/

# add config files
COPY supervisord.conf /etc/supervisord/supervisord.conf
COPY etcd.conf .
COPY kafka.conf .
COPY govpp.conf .
COPY vpp.conf /etc/vpp/vpp.conf

WORKDIR /root/

# run supervisor as the default executable
CMD ["/usr/bin/supervisord", "-c", "/etc/supervisord/supervisord.conf"]



