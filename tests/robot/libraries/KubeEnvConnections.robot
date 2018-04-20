*** Settings ***
Documentation     This is a library to handle connections to testing environment.
Resource          ${CURDIR}/all_libs.robot

*** Keywords ***
Open_Client_Connection
    [Arguments]    ${node_index}=1
    BuiltIn.Log    ${node_index}
    ${client_connection}=    KubeEnv.Open_Connection_To_Node    client    ${node_index}
    BuiltIn.Set_Suite_Variable    ${client_connection}
    KubeEnv.Get_Into_Container_Prompt_In_Pod    ${client_connection}    ${client_pod_name}    prompt=#
    ${client_pod_details} =     KubeCtl.Describe_Pod    ${testbed_connection}    ${client_pod_name}
    ${client_ip} =     BuiltIn.Evaluate    &{client_pod_details}[${client_pod_name}]["IP"]
    BuiltIn.Set_Suite_Variable    ${client_ip}

Open_Server_Connection
    [Arguments]    ${node_index}=1
    BuiltIn.Log    ${node_index}
    ${server_connection}=    KubeEnv.Open_Connection_To_Node    server    ${node_index}
    BuiltIn.Set_Suite_Variable    ${server_connection}
    KubeEnv.Get_Into_Container_Prompt_In_Pod    ${server_connection}    ${server_pod_name}    prompt=#
    ${server_pod_details} =     KubeCtl.Describe_Pod    ${testbed_connection}    ${server_pod_name}
    ${server_ip} =     BuiltIn.Evaluate    &{server_pod_details}[${server_pod_name}]["IP"]
    BuiltIn.Set_Suite_Variable    ${server_ip}

Open_VPP_Connection
    [Arguments]    ${node_index}=1
    BuiltIn.Log    ${node_index}
    ${vpp_connection}=    KubeEnv.Open_Connection_To_Node    vpp    ${node_index}
    BuiltIn.Set_Suite_Variable    ${vpp_connection}
    SshCommons.Switch_And_Write_Command    ${vpp_connection}    vppctl

