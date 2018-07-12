*** Settings ***
Resource    ../../../libraries/kubernetes/Restarts_Setup.robot
Resource    ../../../libraries/kubernetes/KubeTestOperations.robot

Resource     ../../../variables/${VARIABLES}_variables.robot

Library    SSHLibrary

Suite Setup       Run Keywords
...    KubeSetup.Kubernetes Suite Setup    ${CLUSTER_ID}
...    AND    Restarts Suite Setup with ${16} VNFs at ${6} memifs each and ${48} non-VPP containers
Suite Teardown    Restarts Suite Teardown
Test Teardown     Recreate Topology If Test Failed

Documentation    Test suite for Kubernetes pod restarts using a large number
...    of VNF and non-VPP pods.
...
...    Restart performed through kubernetes pod deletion and through
...    segmentation fault signal sent to VPP.
...
...    Connectivity verified using "ping" command from every non-VPP pod to each
...    visible VNF pod.

*** Variables ***
${VARIABLES}=       common
${ENV}=             common
${CLUSTER_ID}=      INTEGRATION1
${vnf0_ip}=         192.168.1.1
${novpp0_ip}=       192.168.1.2

${repeats}=         1

*** Test Cases ***
Scale restart scenario - VNF
    Repeat Keyword    ${repeats}    Scale restart scenario - VNF

Scale restart scenario - noVPP
    Repeat Keyword    ${repeats}    Scale restart scenario - noVPP

Scale restart scenario - VSwitch
    Repeat Keyword    ${repeats}    Scale restart scenario - VSwitch

Scale Restart Scenario - VSwitch and VNF
    Repeat Keyword    ${repeats}    Scale Restart Scenario - VSwitch and VNF

Scale Restart Scenario - VSwitch and noVPP
    Repeat Keyword    ${repeats}    Scale Restart Scenario - VSwitch and noVPP

Scale Restart Scenario - VSwitch, noVPP and VNF
    Repeat Keyword    ${repeats}    Scale Restart Scenario - VSwitch, noVPP and VNF

Scale Restart Scenario - full topology in sequence etcd-vswitch-pods-sfc
    Repeat Keyword    ${repeats}    Scale Restart Scenario - full topology in sequence etcd-vswitch-pods-sfc

Scale Restart Scenario - full topology in sequence etcd-vswitch-sfc-pods
    Repeat Keyword    ${repeats}    Scale Restart Scenario - full topology in sequence etcd-vswitch-sfc-pods

Scale Restart Scenario - full topology in sequence etcd-sfc-vswitch-pods
    Repeat Keyword    ${repeats}    Scale Restart Scenario - full topology in sequence etcd-sfc-vswitch-pods

Scale Restart Scenario - full topology in sequence etcd-sfc-pods-vswitch
    Repeat Keyword    ${repeats}    Scale Restart Scenario - full topology in sequence etcd-sfc-pods-vswitch


*** Keywords ***
Recreate Topology If Test Failed
    [Documentation]    After a failed test, delete the kubernetes topology
    ...    and create it again.
    BuiltIn.Run Keyword If Test Failed    Run Keywords
    ...    Log Pods For Debug    ${testbed_connection}
    ...    AND    Cleanup_Restarts_Deployment_On_Cluster    ${testbed_connection}
    ...    AND    Restarts Setup with ${1} VNFs at ${1} memifs each and ${1} non-VPP containers

Scale restart scenario - VNF
    [Documentation]    Restart all VNF nodes, ping their IP addresses from the non-VPP
    ...    nodes until each receives a reply, then verify connectivity both ways.

    Ping Until Success - Unix Ping    novpp-0    ${vnf0_ip}    timeout=120s
    Scale Ping Until Success - Unix Ping
    Trigger Pod Restart - Pod Deletion       ${testbed_connection}    vnf-vpp-0
    Ping Until Success - Unix Ping           novpp-0    ${vnf0_ip}    timeout=120s
    Scale Ping Until Success - Unix Ping

    Trigger Pod Restart - VPP SIGSEGV        ${vnf_pods[0]}
    Ping Until Success - Unix Ping           novpp-0    ${vnf0_ip}    timeout=120s
    Scale Ping Until Success - Unix Ping

Scale restart scenario - noVPP
    [Documentation]    Restart non-VPP node, ping it's IP address from the VNF
    ...    node until a reply is received, then verify connectivity both ways.

    Ping Until Success - Unix Ping    novpp-0    ${vnf0_ip}    timeout=120s
    Scale Ping Until Success - Unix Ping
    Trigger Pod Restart - Pod Deletion       ${testbed_connection}    novpp-0
    Ping Until Success - VPP Ping            ${vnf_pods[0]}           ${novpp0_ip}    timeout=120s
    Scale Ping Until Success - Unix Ping

Scale restart scenario - VSwitch
    [Documentation]    Restart the vswitch, ping the VNF's IP address from
    ...    the non-VPP node until a reply is received, then verify connectivity
    ...    both ways.

    Ping Until Success - Unix Ping    novpp-0    ${vnf0_ip}    timeout=120s
    Scale Ping Until Success - Unix Ping
    Trigger Pod Restart - Pod Deletion       ${testbed_connection}    ${vswitch_pod_name}    vswitch=${TRUE}
    Ping Until Success - Unix Ping           novpp-0    ${vnf0_ip}    timeout=120s
    Scale Ping Until Success - Unix Ping

    Trigger Pod Restart - VPP SIGSEGV        ${vswitch_pod_name}
    Ping Until Success - Unix Ping           novpp-0    ${vnf0_ip}    timeout=120s
    Scale Ping Until Success - Unix Ping

Scale Restart Scenario - VSwitch and VNF
    [Documentation]    Restart vswitch and VNF, ping the VNF's IP address from
    ...    the non-VPP node until a reply is received, then verify connectivity
    ...    both ways.

    Ping Until Success - Unix Ping    novpp-0    ${vnf0_ip}    timeout=120s
    Scale Ping Until Success - Unix Ping
    Trigger Pod Restart - Pod Deletion       ${testbed_connection}    vnf-vpp-0
    Trigger Pod Restart - Pod Deletion       ${testbed_connection}    ${vswitch_pod_name}    vswitch=${TRUE}
    Ping Until Success - Unix Ping           novpp-0    ${vnf0_ip}    timeout=120s
    Scale Ping Until Success - Unix Ping

    Trigger Pod Restart - VPP SIGSEGV        ${vnf_pods[0]}
    Trigger Pod Restart - VPP SIGSEGV        ${vswitch_pod_name}
    Ping Until Success - Unix Ping           novpp-0    ${vnf0_ip}    timeout=120s
    Scale Ping Until Success - Unix Ping

Scale Restart Scenario - VSwitch and noVPP
    [Documentation]    Restart vswitch and non-VPP pod, ping the non-VPP
    ...    pod's IP address from the VNF node until a reply is received, then
    ...    verify connectivity both ways.

    Ping Until Success - Unix Ping    novpp-0    ${vnf0_ip}    timeout=120s
    Scale Ping Until Success - Unix Ping
    Trigger Pod Restart - Pod Deletion       ${testbed_connection}    novpp-0
    Trigger Pod Restart - Pod Deletion       ${testbed_connection}    ${vswitch_pod_name}    vswitch=${TRUE}
    Ping Until Success - VPP Ping            ${vnf_pods[0]}           ${novpp0_ip}    timeout=120s
    Scale Ping Until Success - Unix Ping

    Trigger Pod Restart - Pod Deletion       ${testbed_connection}    novpp-0
    Trigger Pod Restart - VPP SIGSEGV        ${vswitch_pod_name}
    Ping Until Success - VPP Ping            ${vnf_pods[0]}           ${novpp0_ip}    timeout=120s
    Scale Ping Until Success - Unix Ping

Scale Restart Scenario - VSwitch, noVPP and VNF
    [Documentation]    Restart vswitch, VNF and non-VPP pod, ping the non-VPP
    ...    pod's IP address from the VNF node until a reply is received, then
    ...    verify connectivity both ways.

    Ping Until Success - Unix Ping    novpp-0    ${vnf0_ip}    timeout=120s
    Scale Ping Until Success - Unix Ping
    Trigger Pod Restart - Pod Deletion       ${testbed_connection}    vnf-vpp-0
    Trigger Pod Restart - Pod Deletion       ${testbed_connection}    novpp-0
    Trigger Pod Restart - Pod Deletion       ${testbed_connection}    ${vswitch_pod_name}    vswitch=${TRUE}
    Ping Until Success - VPP Ping            ${vnf_pods[0]}           ${novpp0_ip}    timeout=120s
    Scale Ping Until Success - Unix Ping

    Trigger Pod Restart - VPP SIGSEGV        ${vnf_pods[0]}
    Trigger Pod Restart - Pod Deletion       ${testbed_connection}    novpp-0
    Trigger Pod Restart - VPP SIGSEGV        ${vswitch_pod_name}
    Ping Until Success - VPP Ping            ${vnf_pods[0]}           ${novpp0_ip}    timeout=120s
    Scale Ping Until Success - Unix Ping

Scale Restart Scenario - full topology in sequence etcd-vswitch-pods-sfc
    [Documentation]    Restart the full topology, then bring it back up in the
    ...    specified sequence and verify connectivity between VNF and non-VPP
    ...    pods.

    Ping Until Success - Unix Ping    novpp-0    ${vnf0_ip}    timeout=120s
    Scale Ping Until Success - Unix Ping
    Restart Topology With Startup Sequence    etcd    vswitch    vnf    novpp    sfc
    Ping Until Success - Unix Ping           novpp-0    ${vnf0_ip}    timeout=120s
    Scale Ping Until Success - Unix Ping

Scale Restart Scenario - full topology in sequence etcd-vswitch-sfc-pods
    [Documentation]    Restart the full topology, then bring it back up in the
    ...    specified sequence and verify connectivity between VNF and non-VPP
    ...    pods.

    Ping Until Success - Unix Ping    novpp-0    ${vnf0_ip}    timeout=120s
    Scale Ping Until Success - Unix Ping
    Restart Topology With Startup Sequence    etcd    vswitch    sfc    vnf    novpp
    Ping Until Success - Unix Ping           novpp-0    ${vnf0_ip}    timeout=120s
    Scale Ping Until Success - Unix Ping

Scale Restart Scenario - full topology in sequence etcd-sfc-vswitch-pods
    [Documentation]    Restart the full topology, then bring it back up in the
    ...    specified sequence and verify connectivity between VNF and non-VPP
    ...    pods.

    Ping Until Success - Unix Ping    novpp-0    ${vnf0_ip}    timeout=120s
    Scale Ping Until Success - Unix Ping
    Restart Topology With Startup Sequence    etcd    sfc    vswitch    vnf    novpp
    Ping Until Success - Unix Ping           novpp-0    ${vnf0_ip}    timeout=120s
    Scale Ping Until Success - Unix Ping

Scale Restart Scenario - full topology in sequence etcd-sfc-pods-vswitch
    [Documentation]    Restart the full topology, then bring it back up in the
    ...    specified sequence and verify connectivity between VNF and non-VPP
    ...    pods.

    Ping Until Success - Unix Ping    novpp-0    ${vnf0_ip}    timeout=120s
    Scale Ping Until Success - Unix Ping
    Restart Topology With Startup Sequence    etcd    sfc    vnf    novpp    vswitch
    Ping Until Success - Unix Ping           novpp-0    ${vnf0_ip}    timeout=120s
    Scale Ping Until Success - Unix Ping
