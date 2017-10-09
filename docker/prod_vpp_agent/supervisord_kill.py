#!/usr/bin/env python

import sys
import os
import signal


def write_stdout(msg):
    # only eventlistener protocol messages may be sent to stdout
    sys.stdout.write(msg)
    sys.stdout.flush()


def write_stderr(msg):
    sys.stderr.write("EVENT LISTENER: " + msg)
    sys.stderr.flush()


def main():
    while 1:
        # transition from ACKNOWLEDGED to READY
        write_stdout('READY\n')

        # read header line and print it to stderr
        line = sys.stdin.readline()
        write_stderr(line)

        # read event payload and print it to stderr
        headers = dict([x.split(':') for x in line.split()])
        data = sys.stdin.read(int(headers['len']))
        write_stderr(data)
        try:
            parsed_data = dict([x.split(':') for x in data.split()])
            # ignore non vpp events, skipping
            if parsed_data["processname"] != "vpp":
                msg = 'Ignoring event from ' + parsed_data["processname"]
                write_stderr(msg)
                write_stdout('RESULT 2\nOK')
                continue
            with open('/root/supervisord.pid', 'r') as pidfile:
                pid = int(pidfile.readline())
            write_stderr('Killing supervisors with pid: ' + str(pid))
            os.kill(pid, signal.SIGQUIT)
        except Exception as e:
            write_stderr('Could not kill supervisor: ' + str(e) + '\n')

        # transition from READY to ACKNOWLEDGED
        write_stdout('RESULT 2\nOK')


if __name__ == '__main__':
    main()
