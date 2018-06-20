*** Settings ***
Resource    KubeEnv.robot
Resource    ../SshCommons.robot

*** Keywords ***
Verify Pod Connectivity - Unix Ping
    [Documentation]    Execute ping on the connection provided
    [Arguments]    ${source_pod_name}    ${destination_ip}     ${count}=5
    ${stdout} =    Run Command In Pod    ping -c ${count} ${destination_ip}    ${source_pod_name}
    BuiltIn.Log Many    ${source_pod_name}    ${destination_ip}     ${count}
    BuiltIn.Should Contain    ${stdout}    ${count} received, 0% packet loss

Verify Pod Connectivity - VPP Ping
    [Arguments]    ${source_pod_name}    ${destination_ip}     ${count}=5
    BuiltIn.Log Many    ${source_pod_name}    ${destination_ip}     ${count}
    ${stdout} =    Run Command In Pod    vppctl ping ${destination_ip} repeat ${count}    ${source_pod_name}
    BuiltIn.Should Contain    ${stdout}    ${count} received, 0% packet loss

Trigger Pod Restart - VPP SIGSEGV
    [Arguments]    ${pod_name}
    BuiltIn.Log    ${pod_name}
    ${stdout} =    Run Command In Pod    ${pod_name}    pkill --signal 11 -f vpp
    log    ${stdout}

Trigger Pod Restart - Pod Deletion
    [Arguments]    ${ssh_session}    ${pod_name}    ${vswitch}=${FALSE}
    BuiltIn.Log Many    ${ssh_session}    ${pod_name}    ${vswitch}
    ${stdout} =    Switch_And_Execute_Command    ${ssh_session}    kubectl delete pod ${pod_name}
    log    ${stdout}
    Wait Until Keyword Succeeds    90sec    5sec    KubeEnv.Verify_Pod_Not_Terminating    ${ssh_session}    ${pod_name}
    Run Keyword If    ${vswitch}    Get Vswitch Pod Name    ${ssh_session}

Reconnect To Pod
    [Arguments]    ${pod_name}    ${pod_index}    ${pod_connection}    ${cluster_node_index}=1
    BuiltIn.Log Many    ${pod_name}    ${pod_index}    ${pod_connection}    ${cluster_node_index}
    ${vnf_connection}=    KubeEnv.Open_Connection_To_Node    ${pod_name}    ${cluster_id}    ${cluster_node_index}
    Set List Value    ${vnf_connections}    ${pod_index}    ${vnf_connection}
    KubeEnv.Get_Into_Container_Prompt_In_Pod    ${vnf_connection}    ${pod_name}    prompt=#
    BuiltIn.Set_Suite_Variable    ${vnf_connections}

Ping Until Success - Unix Ping
    [Arguments]    ${source_pod_name}    ${destination_ip}    ${timeout}
    [Timeout]    ${timeout}
    BuiltIn.Log Many    ${source_pod_name}    ${destination_ip}    ${timeout}
    ${stdout} =    Run Command In Pod    /bin/bash -c "until ping -c1 -w1 ${destination_ip} &>/dev/null; do :; done"    ${source_pod_name}
    log    ${stdout}

Ping Until Success - VPP Ping
    [Arguments]    ${source_pod_name}    ${destination_ip}    ${timeout}
    BuiltIn.Log Many    ${source_pod_name}    ${destination_ip}    ${timeout}
    BuiltIn.Wait Until Keyword Succeeds    ${timeout}    5s    Verify Pod Connectivity - VPP Ping    ${source_pod_name}    ${destination_ip}    count=1

Get Vswitch Pod Name
    [Arguments]    ${ssh_session}
    BuiltIn.Log Many    ${ssh_session}
    ${vswitch_pod_name} =    Get_Deployed_Pod_Name    ${ssh_session}    vswitch-deployment-
    Set Global Variable    ${vswitch_pod_name}

Restart Topology With Startup Sequence
    [Arguments]    @{sequence}
    BuiltIn.Log Many    @{sequence}
    Cleanup_Basic_Restarts_Deployment_On_Cluster    ${testbed_connection}
    :FOR    ${item}    IN    @{sequence}
    \    Run Keyword If    "${item}"=="etcd"       KubeEnv.Deploy_Etcd_And_Verify_Running    ${testbed_connection}
    \    Run Keyword If    "${item}"=="vswitch"    KubeEnv.Deploy_Vswitch_Pod_And_Verify_Running    ${testbed_connection}
    \    Run Keyword If    "${item}"=="sfc"        KubeEnv.Deploy_SFC_Pod_And_Verify_Running    ${testbed_connection}
    \    Run Keyword If    "${item}"=="vnf"        KubeEnv.Deploy_VNF_Pods    ${testbed_connection}    ${1}
    \    Run Keyword If    "${item}"=="novpp"      KubeEnv.Deploy_NoVPP_Pods    ${testbed_connection}    ${1}
