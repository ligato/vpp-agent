*** Settings ***
Library     SSHLibrary
Library     ${CURDIR}/kube_config_gen.py
Resource    ${CURDIR}/KubeEnv.robot
Resource    ${CURDIR}/KubeSetup.robot
Resource    ${CURDIR}/KubeCtl.robot
Resource    ${CURDIR}/../SshCommons.robot

*** Keywords ***
Basic Restarts Setup with ${vnf_count} VNFs and ${novpp_count} non-VPP containers
    [Documentation]    Execute common setup, clean 1node cluster, deploy pods.
    KubeSetup.Kubernetes Suite Setup   ${CLUSTER_ID}
    Cleanup_Basic_Restarts_Deployment_On_Cluster    ${testbed_connection}
    Generate YAML Config Files    ${vnf_count}    ${novpp_count}
    KubeEnv.Deploy_Etcd_And_Verify_Running    ${testbed_connection}
    KubeEnv.Deploy_Vswitch_Pod_And_Verify_Running    ${testbed_connection}
    KubeEnv.Deploy_VNF_Pods    ${testbed_connection}    ${vnf_count}
    KubeEnv.Deploy_NoVPP_Pods    ${testbed_connection}    ${novpp_count}
    KubeEnv.Deploy_SFC_Pod_And_Verify_Running    ${testbed_connection}
    Open_Restarts_Connections    ${vnf_count}    ${novpp_count}    node_index=1    cluster_id=${CLUSTER_ID}

Basic Restarts Teardown
    [Documentation]    Log leftover output from pods, remove pods, execute common teardown.
    KubeEnv.Log_Pods_For_Debug    ${testbed_connection}    exp_nr_vswitch=1
    Cleanup_Basic_Restarts_Deployment_On_Cluster    ${testbed_connection}
    KubeSetup.Kubernetes Suite Teardown    ${CLUSTER_ID}

Cleanup_Basic_Restarts_Deployment_On_Cluster
    [Arguments]    ${testbed_connection}
    [Documentation]    Assuming active SSH connection, delete all Kubernetes elements and wait for completion.
    SSHLibrary.Switch_Connection  ${testbed_connection}
    SshCommons.Execute_Command_And_Log    kubectl delete all --all --namespace=default

Setup_Hosts_Connections
    [Documentation]    Open and store two more SSH connections to master host, in them open
    ...    pod shells to client and server pod, parse their IP addresses and store them.
    Open_Restarts_Connections

Teardown_Hosts_Connections
    [Documentation]    Exit pod shells, close corresponding SSH connections.
    Close_Restarts_Connections

Open_Restarts_Connections
    [Arguments]    ${vnf_count}    ${novpp_count}    ${node_index}=1    ${cluster_id}=INTEGRATION1
    BuiltIn.Log Many    ${vnf_count}    ${novpp_count}    ${node_index}    ${cluster_id}

    ${etcd_connection}=    KubeEnv.Open_Connection_To_Node    etcd    ${cluster_id}     ${node_index}
    BuiltIn.Set_Suite_Variable    ${etcd_connection}
    KubeEnv.Get_Into_Container_Prompt_In_Pod    ${etcd_connection}    ${etcd_pod_name}    prompt=#

    ${vswitch_connection}=    KubeEnv.Open_Connection_To_Node    vswitch    ${cluster_id}     ${node_index}
    BuiltIn.Set_Suite_Variable    ${vswitch_connection}
    KubeEnv.Get_Into_Container_Prompt_In_Pod    ${vswitch_connection}    ${vswitch_pod_name}    prompt=#

    ${sfc_connection}=    KubeEnv.Open_Connection_To_Node    sfc    ${cluster_id}     ${node_index}
    BuiltIn.Set_Suite_Variable    ${sfc_connection}
    KubeEnv.Get_Into_Container_Prompt_In_Pod    ${sfc_connection}    ${sfc_pod_name}    prompt=#

    ${vnf_connections}=    Create List
    :FOR    ${vnf}    IN RANGE    ${vnf_count}
    \    ${vnf_connection}=    KubeEnv.Open_Connection_To_Node    vnf-vpp-${vnf}    ${cluster_id}     ${node_index}
    \    Append To List    ${vnf_connections}    ${vnf_connection}
    \    KubeEnv.Get_Into_Container_Prompt_In_Pod    ${vnf_connection}    vnf-vpp-${vnf}    prompt=#
    BuiltIn.Set_Suite_Variable    ${vnf_connections}

    ${novpp_connections}=    Create List
    :FOR    ${novpp}    IN RANGE    ${novpp_count}
    \    ${novpp_connection}=    KubeEnv.Open_Connection_To_Node    novpp-${novpp}    ${cluster_id}     ${node_index}
    \    Append To List    ${novpp_connections}    ${novpp_connection}
    \    KubeEnv.Get_Into_Container_Prompt_In_Pod    ${novpp_connection}    novpp-${novpp}    prompt=#
    BuiltIn.Set_Suite_Variable    ${novpp_connections}

Close_Restarts_Connections
    BuiltIn.Log Many    @{vnf_connections}
    KubeEnv.Leave_Container_Prompt_In_Pod    ${vswitch_connection}
    KubeEnv.Leave_Container_Prompt_In_Pod    ${sfc_connection}
    KubeEnv.Leave_Container_Prompt_In_Pod    ${etcd_connection}
    :FOR    ${vnf_connection}    IN     @{vnf_connections}
    \   KubeEnv.Leave_Container_Prompt_In_Pod    ${vnf_connection}
    :FOR    ${novpp_connection}    IN     @{novpp_connections}
    \   KubeEnv.Leave_Container_Prompt_In_Pod    ${novpp_connection}

    SSHLibrary.Switch_Connection    ${vswitch_connection}
    SSHLibrary.Close_Connection
    SSHLibrary.Switch_Connection    ${sfc_connection}
    SSHLibrary.Close_Connection
    SSHLibrary.Switch_Connection    ${etcd_connection}
    SSHLibrary.Close_Connection
    :FOR    ${vnf_connection}    IN     @{vnf_connections}
    \   SSHLibrary.Switch_Connection    ${vnf_connection}
    \   SSHLibrary.Close_Connection
    :FOR    ${novpp_connection}    IN     @{novpp_connections}
    \   SSHLibrary.Switch_Connection    ${novpp_connection}
    \   SSHLibrary.Close_Connection

Generate YAML Config Files
    [Arguments]    ${vnf_count}    ${novpp_count}
    BuiltIn.Log Many    ${vnf_count}    ${novpp_count}
    kube_config_gen.generate_config    ${vnf_count}    ${novpp_count}    ${CURDIR}/../../resources/k8-yaml    ${K8_GENERATED_CONFIG_FOLDER}