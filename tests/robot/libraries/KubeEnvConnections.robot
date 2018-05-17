*** Settings ***
Documentation     This is a library to handle connections to testing environment.
Resource          ${CURDIR}/all_libs.robot

*** Keywords ***
Open_CCMTS1_Connections
    [Arguments]    ${node_index}=1    ${cluster_id}=CCMTS1
    BuiltIn.Log    ${node_index}
    BuiltIn.Log    ${cluster_id}
    ${vswitch_connection}=    KubeEnv.Open_Connection_To_Node    vswitch    ${cluster_id}     ${node_index}
    BuiltIn.Set_Suite_Variable    ${vswitch_connection}
    KubeEnv.Get_Into_Container_Prompt_In_Pod    ${vswitch_connection}    ${vswitch_pod_name}    prompt=#
    ${vswitch_pod_details} =     KubeCtl.Describe_Pod    ${testbed_connection}    ${vswitch_pod_name}
    ${vswitch_ip} =     BuiltIn.Evaluate    &{vswitch_pod_details}[${vswitch_pod_name}]["IP"]
    BuiltIn.Set_Suite_Variable    ${vswitch_ip}

    ${cn_infra_connection}=    KubeEnv.Open_Connection_To_Node    cn-infra    ${cluster_id}     ${node_index}
    BuiltIn.Set_Suite_Variable    ${cn_infra_connection}
    KubeEnv.Get_Into_Container_Prompt_In_Pod    ${cn_infra_connection}    ${cn_infra_pod_name}    prompt=#
    ${cn_infra_pod_details} =     KubeCtl.Describe_Pod    ${testbed_connection}    ${cn_infra_pod_name}
    ${cn_infra_ip} =     BuiltIn.Evaluate    &{cn_infra_pod_details}[${cn_infra_pod_name}]["IP"]
    BuiltIn.Set_Suite_Variable    ${cn_infra_ip}

    ${sfc_connection}=    KubeEnv.Open_Connection_To_Node    sfc    ${cluster_id}     ${node_index}
    BuiltIn.Set_Suite_Variable    ${sfc_connection}
    KubeEnv.Get_Into_Container_Prompt_In_Pod    ${sfc_connection}    ${sfc_pod_name}    prompt=#
    ${sfc_pod_details} =     KubeCtl.Describe_Pod    ${testbed_connection}    ${sfc_pod_name}
    ${sfc_ip} =     BuiltIn.Evaluate    &{sfc_pod_details}[${sfc_pod_name}]["IP"]
    BuiltIn.Set_Suite_Variable    ${sfc_ip}

    ${etcd_connection}=    KubeEnv.Open_Connection_To_Node    etcd    ${cluster_id}     ${node_index}
    BuiltIn.Set_Suite_Variable    ${etcd_connection}
    KubeEnv.Get_Into_Container_Prompt_In_Pod    ${etcd_connection}    ${etcd_pod_name}    prompt=#
    ${etcd_pod_details} =     KubeCtl.Describe_Pod    ${testbed_connection}    ${etcd_pod_name}
    ${etcd_ip} =     BuiltIn.Evaluate    &{etcd_pod_details}[${etcd_pod_name}]["IP"]
    BuiltIn.Set_Suite_Variable    ${etcd_ip}

    ${kafka_connection}=    KubeEnv.Open_Connection_To_Node    kafka    ${cluster_id}     ${node_index}
    BuiltIn.Set_Suite_Variable    ${kafka_connection}
    KubeEnv.Get_Into_Container_Prompt_In_Pod    ${kafka_connection}    ${kafka_pod_name}    prompt=#
    ${kafka_pod_details} =     KubeCtl.Describe_Pod    ${testbed_connection}    ${kafka_pod_name}
    ${kafka_ip} =     BuiltIn.Evaluate    &{kafka_pod_details}[${kafka_pod_name}]["IP"]
    BuiltIn.Set_Suite_Variable    ${kafka_ip}

    ${vpp_connection}=    KubeEnv.Open_Connection_To_Node    vpp    ${cluster_id}    ${node_index}
    BuiltIn.Set_Suite_Variable    ${vpp_connection}
    KubeEnv.Get_Into_Container_Prompt_In_Pod    ${vpp_connection}    ${vswitch_pod_name}    prompt=#
    SshCommons.Switch_And_Write_Command    ${vpp_connection}    vppctl


Close_CCMTS1_Connections
    KubeEnv.Leave_Container_Prompt_In_Pod    ${vswitch_connection}
    KubeEnv.Leave_Container_Prompt_In_Pod    ${cn_infra_connection}
    KubeEnv.Leave_Container_Prompt_In_Pod    ${sfc_connection}
    KubeEnv.Leave_Container_Prompt_In_Pod    ${etcd_connection}
    KubeEnv.Leave_Container_Prompt_In_Pod    ${kafka_connection}
    SSHLibrary.Switch_Connection    ${vswitch_connection}
    SSHLibrary.Close_Connection
    SSHLibrary.Switch_Connection    ${cn_infra_connection}
    SSHLibrary.Close_Connection
    SSHLibrary.Switch_Connection    ${sfc_connection}
    SSHLibrary.Close_Connection
    SSHLibrary.Switch_Connection    ${etcd_connection}
    SSHLibrary.Close_Connection
    SSHLibrary.Switch_Connection    ${kafka_connection}
    SSHLibrary.Close_Connection
