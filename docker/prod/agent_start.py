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
            os.system("/bin/vpp-agent --config-dir=/opt/vpp-agent/dev")
            return
    write_stdout('\nVpp-agent omitted\n')

if __name__ == '__main__':
    main()