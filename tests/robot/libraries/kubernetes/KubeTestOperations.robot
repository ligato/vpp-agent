*** Settings ***
Resource    KubeEnv.robot
Resource    ../SshCommons.robot

*** Keywords ***
Verify Pod Connectivity - Unix Ping
    [Documentation]    Execute ping on the connection provided
    [Arguments]    ${source_pod_name}    ${destination_ip}     ${count}=5
    ${stdout} =    Run Command In Pod    ping -c ${count} -s 1400 ${destination_ip}    ${source_pod_name}
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
    ${stdout} =    Run Command In Pod    pkill --signal 11 -f /usr/bin/vpp    ${pod_name}
    log    ${stdout}

Trigger Pod Restart - Pod Deletion
    [Arguments]    ${ssh_session}    ${pod_name}    ${vswitch}=${FALSE}
    BuiltIn.Log Many    ${ssh_session}    ${pod_name}    ${vswitch}
    ${stdout} =    Switch_And_Execute_Command    ${ssh_session}    kubectl delete pod ${pod_name}
    log    ${stdout}
    Wait Until Keyword Succeeds    20sec    1sec    KubeEnv.Verify_Pod_Not_Terminating    ${ssh_session}    ${pod_name}
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
    Cleanup_Restarts_Deployment_On_Cluster    ${testbed_connection}
    :FOR    ${item}    IN    @{sequence}
    \    Run Keyword If    "${item}"=="etcd"       KubeEnv.Deploy_Etcd_And_Verify_Running    ${testbed_connection}
    \    Run Keyword If    "${item}"=="vswitch"    KubeEnv.Deploy_Vswitch_Pod_And_Verify_Running    ${testbed_connection}
    \    Run Keyword If    "${item}"=="sfc"        KubeEnv.Deploy_SFC_Pod_And_Verify_Running    ${testbed_connection}
    \    Run Keyword If    "${item}"=="vnf"        KubeEnv.Deploy_VNF_Pods    ${testbed_connection}    ${1}
    \    Run Keyword If    "${item}"=="novpp"      KubeEnv.Deploy_NoVPP_Pods    ${testbed_connection}    ${1}

Scale Ping Until Success - Unix Ping
    [Arguments]    ${timeout}=6h
    [Timeout]    ${timeout}
    BuiltIn.Log Many    ${topology}    ${timeout}
    :FOR    ${bridge_segment}    IN    @{topology}
    \    Iterate_Over_VNFs    ${bridge_segment}    ${timeout}

Iterate_Over_VNFs
    [Arguments]    ${bridge_segment}    ${timeout}    ${timeout}=1h
    :FOR    ${vnf_pod}    IN    @{bridge_segment["vnf"]}
    \    Iterate_Over_Novpps    ${bridge_segment}    ${vnf_pod}    ${timeout}

Iterate_Over_Novpps
    [Arguments]    ${bridge_segment}    ${vnf_pod}    ${timeout}=10s
    :FOR    ${novpp_pod}    IN    @{bridge_segment["novpp"]}
    \    Ping Until Success - Unix Ping    ${novpp_pod["name"]}    ${vnf_pod["ip"]}    ${timeout}

Wait For Reconnect - Unix Ping
    [Arguments]    ${source_pod_Name}     ${destination_ip}    ${timeout}    ${duration_list_name}
    BuiltIn.Log Many    ${source_pod_Name}     ${destination_ip}    ${timeout}
    ${start_time} =    DateTime.Get Current Date    result_format=epoch
    Ping Until Success - Unix Ping    ${source_pod_Name}     ${destination_ip}    ${timeout}
    ${end_time} =    DateTime.Get Current Date    result_format=epoch
    ${duration} =    Datetime.Subtract Date from Date    ${start_time}    ${end_time}   results_format=verbose
    Collections.Append To List    ${duration_list_name}    ${duration}

Wait For Reconnect - VPP Ping
    [Arguments]    ${source_pod_Name}     ${destination_ip}    ${timeout}    ${duration_list_name}
    BuiltIn.Log Many    ${source_pod_Name}     ${destination_ip}    ${timeout}
    ${start_time} =    DateTime.Get Current Date    result_format=epoch
    Ping Until Success - VPP Ping    ${source_pod_Name}     ${destination_ip}    ${timeout}
    ${end_time} =    DateTime.Get Current Date    result_format=epoch
    ${duration} =    Datetime.Subtract Date from Date    ${start_time}    ${end_time}   results_format=verbose
    Collections.Append To List    ${duration_list_name}    ${duration}
