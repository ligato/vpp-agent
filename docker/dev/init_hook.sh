#!/usr/bin/env bash

terminate_process () {
    PID=$(pidof $1)
    if [[ ${PID} != "" ]]; then
        kill ${PID}
        echo "process $1 terminated"
    fi
}

if [[ "${SUPERVISOR_PROCESS_NAME}" = "agent" && "${SUPERVISOR_PROCESS_STATE}" = "terminated" ]]; then
    terminate_process vpp-agent-init
fi

if [[ "${SUPERVISOR_PROCESS_NAME}" = "vpp" && "${SUPERVISOR_PROCESS_STATE}" = "terminated" ]]; then
    terminate_process vpp-agent-init
fi