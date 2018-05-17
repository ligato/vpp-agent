[Documentation]     Reusable keywords for testsuite setup and teardown

*** Settings ***
Library       String
Library       RequestsLibrary
Library       SSHLibrary            timeout=60s
#Resource      ssh.robot
Resource      ${ENV}_setup-teardown.robot

*** Variables ***
${VM_SSH_ALIAS_PREFIX}     vm_
${snapshot_num}       0
@{NODES}              

*** Keywords ***
Open SSH Connection
    [Arguments]         ${name}    ${ip}    ${user}    ${pswd}
    Log Many            ${name}    ${ip}    ${user}    ${pswd}
    Open Connection     ${ip}      ${name}
    Run Keyword If      "${pswd}"!="rsa_id"   Login                              ${user}   ${pswd}
    Run Keyword If      "${pswd}"=="rsa_id"   SSHLibrary.Login_With_Public_Key   ${user}   %{HOME}/.ssh/id_rsa   any

Testsuite Setup
    [Documentation]  *Testsuite Setup*
    Discard old results
    Open Connection To Docker Host
    Create Connections For ETCD And Kafka
    Start Kafka Server
    Start ETCD Server
    Start VPP Ctl Container
    Make Datastore Snapshots    startup


Testsuite Teardown
    [Documentation]      *Testsuite Teardown*
    Make Datastore Snapshots    teardown
#    Log All SSH Outputs
    Remove All Nodes
    Stop ETCD Server
    Stop VPP Ctl Container
    Stop Kafka Server
    Get Connections
    Close All Connections
    Check Agent Logs For Errors

Test Setup
    Open Connection To Docker Host
    Create Connections For ETCD And Kafka
    Start Kafka Server
    Start ETCD Server
    Start VPP Ctl Container
    Make Datastore Snapshots    startup

Test Teardown
    Make Datastore Snapshots    teardown
    Stop VPP Ctl Container
    Stop Kafka Server
    Stop ETCD Server
#    Log All SSH Outputs
    Remove All Nodes
    Get Connections
    Close All Connections

Testsuite_K8Setup
    [Arguments]    ${cluster_id}
    [Documentation]    Perform actions common for setup of every suite.
    Discard_Old_Results
    Create_Connections_To_Kube_Cluster      ${cluster_id}

Testsuite_K8Teardown
    [Documentation]    Perform actions common for teardown of every suite.
    Log_All_K8_SSH_Outputs
    SSHLibrary.Get_Connections
    SSHLibrary.Close_All_Connections    
    
Discard old results
    [Documentation]    Remove and re-create ${RESULTS_FOLDER}.
    Remove Directory    ${RESULTS_FOLDER}                 recursive=True
    Create Directory    ${RESULTS_FOLDER}

Log_All_K8_Ssh_Outputs
    [Documentation]    Call Log_\${machine}_Output for every cluster node.
    [Timeout]    ${SSH_LOG_OUTPUTS_TIMEOUT}
    : FOR    ${index}    IN RANGE    1    ${K8_CLUSTER_${CLUSTER_ID}_NODES}+1
    \    Log_K8_${VM_SSH_ALIAS_PREFIX}${index}_Output    
    
Log All SSH Outputs
    [Documentation]           *Log All SSH Outputs*
    ...                       Logs all connections outputs
    [Timeout]                 120s
    :FOR    ${id}    IN    @{NODES}
    \    Log ${id} Output
    \    Run Keyword If    "vpp" in "${id}"    Log ${id}_term Output
    \    Run Keyword If    "vpp" in "${id}"    Log ${id}_vat Output          
    Log docker Output

Log_K8_${machine}_Output
    [Documentation]    Switch to \${machine} SSH connection, read with delay of ${SSH_READ_DELAY}, Log and append to log file.
    BuiltIn.Log_Many    ${machine}
    BuiltIn.Comment    TODO: Rewrite this keyword with ${machine} being explicit argument.
    SSHLibrary.Switch_Connection    ${machine}
    ${out} =    SSHLibrary.Read    delay=${SSH_READ_DELAY}s
    BuiltIn.Log    ${out}
    OperatingSystem.Append_To_File    ${RESULTS_FOLDER}/output_${machine}.log    ${out}
    
    
    
Log ${machine} Output
    [Documentation]         *Log ${machine} Output*
    ...                     Logs actual ${machine} output from begining
    Log                     ${machine}
    Switch Connection       ${machine}
    ${out}=                 Read                   delay=${SSH_READ_DELAY}s
    Log                     ${out}
    Append To File          ${RESULTS_FOLDER}/output_${machine}.log                ${out}

Get_K8_Machine_Status
    [Arguments]    ${machine}
    [Documentation]    Execute df, free, ifconfig -a, ps -aux... on \${machine}, assuming ssh connection there is active.
    BuiltIn.Log_Many    ${machine}
    SshCommons.Execute_Command_And_Log    whoami
    SshCommons.Execute_Command_And_Log    pwd
    SshCommons.Execute_Command_And_Log    df
    SshCommons.Execute_Command_And_Log    free
    SshCommons.Execute_Command_And_Log    ip address
    SshCommons.Execute_Command_And_Log    ps aux
    SshCommons.Execute_Command_And_Log    export
    SshCommons.Execute_Command_And_Log    docker images
    SshCommons.Execute_Command_And_Log    docker ps -as
    BuiltIn.Return_From_Keyword_If    """${machine}""" != """${VM_SSH_ALIAS_PREFIX}1"""
    SshCommons.Execute_Command_And_Log    kubectl get nodes    ignore_stderr=True    ignore_rc=True
    SshCommons.Execute_Command_And_Log    kubectl get pods    ignore_stderr=True    ignore_rc=True
    
Create_Connections_To_Kube_Cluster
    [Arguments]    ${cluster_id}
    [Documentation]    Create connection and log machine status for each node.
    : FOR    ${index}    IN RANGE    1    ${K8_CLUSTER_${cluster_id}_NODES}+1
    \    SshCommons.Open_Ssh_Connection_Kube    ${VM_SSH_ALIAS_PREFIX}${index}    ${K8_CLUSTER_${cluster_id}_VM_${index}_PUBLIC_IP}    ${K8_CLUSTER_${cluster_id}_VM_${index}_USER}    ${K8_CLUSTER_${cluster_id}_VM_${index}_PSWD}
    \    SSHLibrary.Set_Client_Configuration    prompt=${K8_CLUSTER_${cluster_id}_VM_${index}_PROMPT}
    \    Get_K8_Machine_Status    ${VM_SSH_ALIAS_PREFIX}${index}
    
Get Machine Status
    [Arguments]              ${machine}
    [Documentation]          *Get Machine Status ${machine}*
    ...                      Executing df, free, ifconfig -a, ps -aux... on ${machine}
    Log                      ${machine}
    Execute On Machine       ${machine}                df
    Execute On Machine       ${machine}                free
    Execute On Machine       ${machine}                ifconfig -a
    Execute On Machine       ${machine}                ps aux
    Execute On Machine       ${machine}                echo $PATH

Open Connection To Docker Host
    Open SSH Connection    docker    ${DOCKER_HOST_IP}    ${DOCKER_HOST_USER}    ${DOCKER_HOST_PSWD}
    Get Machine Status     docker
    Execute On Machine     docker    ${DOCKER_COMMAND} images
    Execute On Machine     docker    ${DOCKER_COMMAND} ps -as

Create Connections For ETCD And Kafka
    Open SSH Connection    etcd    ${DOCKER_HOST_IP}    ${DOCKER_HOST_USER}    ${DOCKER_HOST_PSWD}
    Open SSH Connection    kafka    ${DOCKER_HOST_IP}    ${DOCKER_HOST_USER}    ${DOCKER_HOST_PSWD}

Make_K8_Datastore_Snapshots
    [Arguments]    ${tag}=notag
    [Documentation]    Log ${tag}, compute next prefix (and do nothing with it).
    BuiltIn.Log_Many    ${tag}
    ${prefix} =    Create_K8_Next_Snapshot_Prefix

Create_K8_Next_Snapshot_Prefix
    [Documentation]    Contruct new prefix, store next snapshot num. Return the prefix.
    ${prefix} =    BuiltIn.Evaluate    str(${snapshot_num}).zfill(3)
    ${snapshot_num} =    BuiltIn.Evaluate    ${snapshot_num}+1
    BuiltIn.Set_Global_Variable    ${snapshot_num}
    [Return]    ${prefix}
    
    
Make Datastore Snapshots
    [Arguments]            ${tag}=notag
    Log                    ${tag}
    ${prefix}=             Create Next Snapshot Prefix
    Take ETCD Snapshots    ${prefix}_${tag}

Get ETCD Dump
    ${command}=         Set Variable    ${DOCKER_COMMAND} exec etcd etcdctl get --prefix="true" ""
    ${out}=             Execute On Machine    docker    ${command}    log=false
    [Return]            ${out}

Take ETCD Snapshots
    [Arguments]         ${tag}
    Log                 ${tag}
    ${dump}=            Get ETCD Dump
    Append To File      ${RESULTS_FOLDER}/etcd_dump-${tag}.txt    ${dump}
    ${errors}=          Get Lines Containing String    ${dump}    /error/
    ${status}=          Run Keyword And Return Status    Should Be Empty    ${errors}
    Run Keyword If      ${status}==False         Log     Errors detected in keys: ${errors}    level=WARN
    
Create Next Snapshot Prefix
    ${prefix}=          Evaluate    str(${snapshot_num}).zfill(3)
    ${snapshot_num}=    Evaluate    ${snapshot_num}+1
    Set Global Variable  ${snapshot_num}
    [Return]            ${prefix}

Check Agent Logs For Errors
    @{logs}=    OperatingSystem.List Files In Directory    ${RESULTS_FOLDER}/    *_container_agent.log
    Log List    ${logs}
    :FOR    ${log}    IN    @{logs}
    \    ${data}=    OperatingSystem.Get File    ${RESULTS_FOLDER}/${log}
    \    Should Not Contain    ${data}    exited: agent (exit status
    \    Should Not Contain    ${data}    exited: vpp (exit status
