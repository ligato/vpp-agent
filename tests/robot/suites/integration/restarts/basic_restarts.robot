*** Settings ***
Resource    ../../../libraries/kubernetes/Restarts_Setup.robot
Resource    ../../../libraries/kubernetes/KubeTestOperations.robot

Resource     ../../../variables/${VARIABLES}_variables.robot

Library    SSHLibrary

Suite Setup       Basic Restarts Setup with ${1} VNFs and ${1} non-VPP containers
Suite Teardown    Basic Restarts Teardown

Documentation    Sanity test suite for Kubernetes pod restarts.

*** Variables ***
${VARIABLES}=       common
${ENV}=             common
${CLUSTER_ID}=      INTEGRATION1

*** Test Cases ***
Basic restart scenario - VNF
    [Documentation]    TODO
    Verify Pod Connectivity - Ping      ${novpp_connections[0]}    ${vnf_connections[0]}
    Verify Pod Connectivity - Ping      ${vnf_connections[0]}    ${novpp_connections[0]}

    Trigger Pod Restart                 ${vnf_connections[0]}    delete_pod
    Wait Until Restart                  ${vnf_pods[0]}
    Reconnect To Pod                    ${vnf_pods[0]}    ${vnf_connections[0]}
    Verify Pod Connectivity - Ping      ${novpp_connections[0]}    ${vnf_connections[0]}
    Verify Pod Connectivity - Ping      ${vnf_connections[0]}    ${novpp_connections[0]}

    Trigger Pod Restart                 ${vnf_connections[0]}    kill_process
    Wait Until Restart                  ${vnf_pods[0]}
    Reconnect To Pod                    ${vnf_pods[0]}    ${vnf_connections[0]}
    Verify Pod Connectivity - Ping      ${novpp_connections[0]}    ${vnf_connections[0]}
    Verify Pod Connectivity - Ping      ${vnf_connections[0]}    ${novpp_connections[0]}

Basic restart scenario - noVPP
    [Documentation]    TODO
    Trigger Pod Restart                 ${novpp_connections[0]}    delete_pod
    Wait Until Restart                  ${novpp_pods[0]}
    Reconnect To Pod                    ${novpp_pods[0]}    ${novpp_connections[0]}
    Verify Pod Connectivity - Ping      ${novpp_connections[0]}    ${vnf_connections[0]}
    Verify Pod Connectivity - Ping      ${vnf_connections[0]}    ${novpp_connections[0]}

    Trigger Pod Restart                 ${novpp_connections[0]}    kill_process
    Wait Until Restart                  ${novpp_pods[0]}
    Reconnect To Pod                    ${novpp_pods[0]}    ${novpp_connections[0]}
    Verify Pod Connectivity - Ping      ${novpp_connections[0]}    ${vnf_connections[0]}
    Verify Pod Connectivity - Ping      ${vnf_connections[0]}    ${novpp_connections[0]}

Basic restart scenario - VSwitch
    [Documentation]    TODO
    Trigger Pod Restart                 ${vswitch_connection}    delete_pod
    Wait Until Restart                  ${vswitch_pod}
    Reconnect To Pod                    ${vswitch_pod}    ${vswitch_connection}
    Verify Pod Connectivity - Ping      ${novpp_connections[0]}    ${vnf_connections[0]}
    Verify Pod Connectivity - Ping      ${vnf_connections[0]}    ${novpp_connections[0]}

    Trigger Pod Restart                 ${vswitch_connection}    kill_process
    Wait Until Restart                  ${vswitch_pod}
    Reconnect To Pod                    ${vswitch_pod}    ${vswitch_connection}
    Verify Pod Connectivity - Ping      ${novpp_connections[0]}    ${vnf_connections[0]}
    Verify Pod Connectivity - Ping      ${vnf_connections[0]}    ${novpp_connections[0]}

Basic Restart Scenario - VSwitch and VNF
    [Documentation]    TODO
    Trigger Pod Restart                 ${vnf_connections[0]}    delete_pod
    Trigger Pod Restart                 ${vswitch_connection}    delete_pod
    Wait Until Restart                  ${vnf_pods[0]}
    Wait Until Restart                  ${vswitch_pod}
    Reconnect To Pod                    ${vnf_pods[0]}    ${vnf_connections[0]}
    Reconnect To Pod                    ${vswitch_pod}    ${vswitch_connection}
    Verify Pod Connectivity - Ping      ${novpp_connections[0]}    ${vnf_connections[0]}
    Verify Pod Connectivity - Ping      ${vnf_connections[0]}    ${novpp_connections[0]}

    Trigger Pod Restart                 ${vnf_connections[0]}    kill_process
    Trigger Pod Restart                 ${vswitch_connection}    kill_process
    Wait Until Restart                  ${vnf_pods[0]}
    Wait Until Restart                  ${vswitch_pod}
    Reconnect To Pod                    ${vnf_pods[0]}    ${vnf_connections[0]}
    Reconnect To Pod                    ${vswitch_pod}    ${vswitch_connection}
    Verify Pod Connectivity - Ping      ${novpp_connections[0]}    ${vnf_connections[0]}
    Verify Pod Connectivity - Ping      ${vnf_connections[0]}    ${novpp_connections[0]}

#TODO:
Basic Restart Scenario - VSwitch and noVPP
Basic Restart Scenario - VSwitch, noVPP and VNF
Basic Restart Scenario - full topology (startup sequence: etcd-vswitch-pods-sfc)
Basic Restart Scenario - full topology (startup sequence: etcd-vswitch-sfc-pods)
Basic Restart Scenario - full topology (startup sequence: etcd-sfc-vswitch-pods)
Basic Restart Scenario - full topology (startup sequence: etcd-sfc-pods-vswitch)

#TODO: 4 memifs per VNF
#TODO: repeat test case execution X times
#TODO: verify connectivity with traffic (iperf,tcpkali,...) longer than memif ring size
#TODO: scale up to 16 VNFs and 50 non-VPP containers.
#TODO: measure pod restart time
