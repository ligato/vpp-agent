#!/bin/bash

set +e
sudo docker rmi -f dev_vpp_agent_shrink 2>/dev/null
sudo docker rm -f shrink 2>/dev/null
set -e

sudo docker run -itd --name shrink dev_vpp_agent bash
sudo docker exec shrink /bin/bash -c 'mkdir -p /root/vpp/build-root && cp /opt/vpp-agent/dev/vpp/build-root/*.deb /root/vpp/build-root && rm -rf /opt/vpp-agent/dev/vpp && \
    mv /root/vpp /opt/vpp-agent/dev'
sudo docker export shrink >shrink.tar
sudo docker rm -f shrink
sudo docker import -c "WORKDIR /root/" -c 'CMD ["/usr/bin/supervisord", "-c", "/etc/supervisord.conf"]' shrink.tar dev_vpp_agent_shrink
rm shrink.tar
