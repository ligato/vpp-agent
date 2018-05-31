#!/usr/bin/env bash

set -e

if [ -n "$OMIT_AGENT" ]; then
    echo "Start of vpp-agent is omitted (unset OMIT_AGENT to disable it)"
else
    echo "Starting vpp-agent.."
    exec vpp-agent --config-dir=/opt/vpp-agent/dev
fi
