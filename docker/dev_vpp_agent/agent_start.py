#!/usr/bin/env python

import sys
import os

def write_stdout(msg):
    sys.stdout.write(msg)
    sys.stdout.flush()

def main():
    # start agent if set
    if 'START_AGENT' in os.environ:
        agent_start = os.environ['START_AGENT']
        if agent_start=='true' or agent_start == 'True':
            write_stdout('\nStarting vpp-agent...\n')
            os.system("/root/go/bin/vpp-agent --etcdv3-config=/opt/vpp-agent/dev/etcd.conf --kafka-config=/opt/vpp-agent/dev/kafka.conf --default-plugins-config=/opt/vpp-agent/dev/defaultplugins.conf --linuxplugin-config=/opt/vpp-agent/dev/linuxplugin.conf  --logs-config=/opt/vpp-agent/dev/logs.conf")
            return
    write_stdout('\nVpp-agent omitted\n')

if __name__ == '__main__':
    main()