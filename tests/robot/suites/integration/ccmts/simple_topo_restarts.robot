*** Settings ***
Documentation     Test suite to test basic ping, udp, tcp on basic ccmts topology inlcd. vswitch restart.
Library      OperatingSystem

Library      Collections

Resource     ../../variables/${VARIABLES}_variables.robot

Resource     ../../libraries/all_libs.robot

*** Variables ***
${REPLY_DATA_FOLDER}            replyACL
${VARIABLES}=       common
${ENV}=             common

Suite Setup       BasicCcmtsSetup
Suite Teardown    BasicCcmtsTeardown

*** Test Cases ***
Pod_To_Pod_Ping
    [Documentation]    Execute "ping -c 5" command between pods (both ways), require no packet loss.
    [Setup]    Setup_Hosts_Connections
    ${stdout} =    KubernetesEnv.Run_Finite_Command_In_Pod    ping -c 5 ${server_ip}    ssh_session=${client_connection}
    BuiltIn.Should_Contain   ${stdout}    5 received, 0% packet loss
    ${stdout} =    KubernetesEnv.Run_Finite_Command_In_Pod    ping -c 5 ${client_ip}    ssh_session=${server_connection}
    BuiltIn.Should_Contain   ${stdout}    5 received, 0% packet loss
    [Teardown]    Teardown_Hosts_Connections

Pod_To_Pod_Udp
    [Documentation]    Start UDP server and client, send message, stop both and check the message has been reseived.
    [Setup]    Setup_Hosts_Connections
    KubernetesEnv.Init_Infinite_Command_in_Pod    nc -ul -p 7000    ssh_session=${server_connection}
    KubernetesEnv.Init_Infinite_Command_in_Pod    nc -u ${server_ip} 7000    ssh_session=${client_connection}
    ${text} =    BuiltIn.Set_Variable    Text to be received
    SSHLibrary.Write    ${text}
    ${client_stdout} =    KubernetesEnv.Stop_Infinite_Command_In_Pod    ssh_session=${client_connection}
    ${server_stdout} =    KubernetesEnv.Stop_Infinite_Command_In_Pod    ssh_session=${server_connection}
    BuiltIn.Should_Contain   ${server_stdout}    ${text}
    [Teardown]    Teardown_Hosts_Connections

Pod_To_Pod_Tcp
    [Documentation]    Start TCP server, start client sending the message, stop server, check message has been received, stop client.
    [Setup]    Setup_Hosts_Connections
    ${text} =    BuiltIn.Set_Variable    Text to be received
    KubernetesEnv.Run_Finite_Command_In_Pod    cd; echo "${text}" > some.file    ssh_session=${client_connection}
    KubernetesEnv.Init_Infinite_Command_in_Pod    nc -l -p 4444    ssh_session=${server_connection}
    KubernetesEnv.Init_Infinite_Command_in_Pod    cd; nc ${server_ip} 4444 < some.file    ssh_session=${client_connection}
    ${server_stdout} =    KubernetesEnv.Stop_Infinite_Command_In_Pod    ssh_session=${server_connection}
    BuiltIn.Should_Contain   ${server_stdout}    ${text}
    ${client_stdout} =    KubernetesEnv.Stop_Infinite_Command_In_Pod    ssh_session=${client_connection}
    [Teardown]    Teardown_Hosts_Connections

Host_To_Pod_Ping
    [Documentation]    Execute "ping -c 5" command from host to both pods, require no packet loss.
    [Setup]    Setup_Hosts_Connections
    ${stdout} =    SshCommons.Switch_And_Execute_Command    ${testbed_connection}    ping -c 5 ${server_ip}
    BuiltIn.Should_Contain   ${stdout}    5 received, 0% packet loss
    ${stdout} =    SshCommons.Switch_And_Execute_Command    ${testbed_connection}    ping -c 5 ${client_ip}
    BuiltIn.Should_Contain   ${stdout}    5 received, 0% packet loss
    [Teardown]    Teardown_Hosts_Connections

Host_To_Pod_Udp
    [Documentation]    The same as Pod_To_Pod_Udp but client is on host instead of pod.
    [Setup]    Setup_Hosts_Connections
    KubernetesEnv.Init_Infinite_Command_in_Pod    nc -ul -p 7000    ssh_session=${server_connection}
    KubernetesEnv.Init_Infinite_Command_in_Pod    nc -u ${server_ip} 7000    ssh_session=${testbed_connection}
    ${text} =    BuiltIn.Set_Variable    Text to be received
    SSHLibrary.Write    ${text}
    ${client_stdout} =    KubernetesEnv.Stop_Infinite_Command_In_Pod    ssh_session=${testbed_connection}    prompt=$
    ${server_stdout} =    KubernetesEnv.Stop_Infinite_Command_In_Pod    ssh_session=${server_connection}
    BuiltIn.Should_Contain   ${server_stdout}    ${text}
    [Teardown]    Teardown_Hosts_Connections

Host_To_Pod_Tcp
    [Documentation]    The same as Pod_To_Pod_Tcp but client is on host instead of pod.
    [Setup]    Setup_Hosts_Connections
    ${text} =    BuiltIn.Set_Variable    Text to be received
    KubernetesEnv.Run_Finite_Command_In_Pod    cd; echo "${text}" > some.file    ssh_session=${testbed_connection}
    KubernetesEnv.Init_Infinite_Command_in_Pod    nc -l -p 4444    ssh_session=${server_connection}
    KubernetesEnv.Init_Infinite_Command_in_Pod    cd; nc ${server_ip} 4444 < some.file    ssh_session=${testbed_connection}
    ${server_stdout} =    KubernetesEnv.Stop_Infinite_Command_In_Pod    ssh_session=${server_connection}
    BuiltIn.Should_Contain   ${server_stdout}    ${text}
    ${client_stdout} =    KubernetesEnv.Stop_Infinite_Command_In_Pod    ssh_session=${testbed_connection}    prompt=$
    [Teardown]    Teardown_Hosts_Connections

*** Keywords ***
Cleanup_Basic_Ccmts_Deployment_On_Cluster
    [Documentation]    Assuming active SSH connection, store its index, execute multiple commands to cleanup 1node cluster, wait to see it running.
    BuiltIn.Set_Suite_Variable    ${testbed_connection}    ${VM_SSH_ALIAS_PREFIX}1
    SSHLibrary.Switch_Connection  ${testbed_connection}
    #KubeCtl.Taint    ${testbed_connection}    nodes --all node-role.kubernetes.io/master-
    SshCommons.Execute_Command_And_Log    kubectl delete --all services --namespace=default
	SshCommons.Execute_Command_And_Log    kubectl delete --all services --namespace=default
	SshCommons.Execute_Command_And_Log    kubectl delete --all deployments --namespace=default
	SshCommons.Execute_Command_And_Log    kubectl delete --all statefulSets --namespace=default
	SshCommons.Execute_Command_And_Log    kubectl delete --all daemonSets --namespace=default
	SshCommons.Execute_Command_And_Log    kubectl delete --all pods --namespace=default
	SshCommons.Execute_Command_And_Log    kubectl delete --all rs --namespace=default
	SshCommons.Execute_Command_And_Log    kubectl delete --all pv --namespace=default
	SshCommons.Execute_Command_And_Log    kubectl delete --all pvc --namespace=default
	SshCommons.Execute_Command_And_Log    kubectl delete --all ConfigMaps --namespace=default
	SshCommons.Execute_Command_And_Log    rm -rf /tmp/'+self.tName+'/default/*
	SshCommons.Execute_Command_And_Log    kubectl delete daemonSets snap --namespace=kube-system
	SshCommons.Execute_Command_And_Log    kubectl delete deployments heapster --namespace=kube-system
	SshCommons.Execute_Command_And_Log    kubectl delete serviceaccounts heapster --namespace=kube-system


BasicCcmtsSetup
    [Documentation]    Execute common setup, clean 1node cluster, deploy pods.
    setup-teardown-kube.Testsuite_Setup
    #KubernetesEnv.Reinit_1_Node_Cluster
    Cleanup_Basic_Ccmts_Deployment_On_Cluster
    KubernetesEnv.Deploy_Client_And_Server_Pod_And_Verify_Running    ${testbed_connection}

BasicCcmtsSetupTeardown
    [Documentation]    Log leftover output from pods, remove pods, execute common teardown.
    KubernetesEnv.Log_Pods_For_Debug    ${testbed_connection}    exp_nr_vswitch=1
    KubernetesEnv.Remove_Client_And_Server_Pod_And_Verify_Removed    ${testbed_connection}
    setup-teardown-kube.Testsuite Teardown

Setup_Hosts_Connections
    [Documentation]    Open and store two more SSH connections to master host, in them open
    ...    pod shells to client and server pod, parse their IP addresses and store them.
    EnvConnections.Open_Client_Connection
    EnvConnections.Open_Server_Connection

Teardown_Hosts_Connections
    [Documentation]    Exit pod shells, close corresponding SSH connections.
    KubernetesEnv.Leave_Container_Prompt_In_Pod    ${client_connection}
    KubernetesEnv.Leave_Container_Prompt_In_Pod    ${server_connection}
    SSHLibrary.Switch_Connection    ${client_connection}
    SSHLibrary.Close_Connection
    SSHLibrary.Switch_Connection    ${server_connection}
    SSHLibrary.Close_Connection





*** Test Cases ***
Configure Environment
    [Tags]    setup
    Configure Environment 2        acl_basic.conf

Show ACL Before Setup
    vpp_ctl: Check ACL Reply    agent_vpp_1    ${ACL1_NAME}    ${REPLY_DATA_FOLDER}/reply_acl_empty.txt     ${REPLY_DATA_FOLDER}/reply_acl_empty_term.txt

Add ACL1_TCP
    vpp_ctl: Put ACL TCP   agent_vpp_1   ${ACL1_NAME}    ${E_INTF1}    ${I_INTF1}   ${RULE_NM1_1}    ${ACTION_DENY}     ${DEST_NTW}     ${SRC_NTW}   ${1DEST_PORT_L}   ${1DEST_PORT_U}    ${1SRC_PORT_L}     ${1SRC_PORT_U}
    Sleep    ${SYNC_SLEEP}

Check ACL1 is created
    vpp_ctl: Check ACL Reply    agent_vpp_1    ${ACL1_NAME}    ${REPLY_DATA_FOLDER}/reply_acl1_tcp.txt    ${REPLY_DATA_FOLDER}/reply_acl1_tcp_term.txt


Add ACL2_TCP
    vpp_ctl: Put ACL TCP   agent_vpp_1   ${ACL2_NAME}    ${E_INTF1}    ${I_INTF1}   ${RULE_NM2_1}    ${ACTION_DENY}     ${DEST_NTW}     ${SRC_NTW}   ${2DEST_PORT_L}   ${2DEST_PORT_U}    ${2SRC_PORT_L}     ${2SRC_PORT_U}
    Sleep    ${SYNC_SLEEP}

Check ACL2 is created and ACL1 still Configured
    vpp_ctl: Check ACL Reply    agent_vpp_1    ${ACL2_NAME}   ${REPLY_DATA_FOLDER}/reply_acl2_tcp.txt    ${REPLY_DATA_FOLDER}/reply_acl2_tcp_term.txt



Update ACL1
    vpp_ctl: Put ACL TCP   agent_vpp_1   ${ACL1_NAME}    ${E_INTF1}     ${I_INTF1}   ${RULE_NM1_1}    ${ACTION_PERMIT}     ${DEST_NTW}    ${SRC_NTW}   ${1DEST_PORT_L}   ${1DEST_PORT_U}    ${1SRC_PORT_L}     ${1SRC_PORT_U}
    Sleep    ${SYNC_SLEEP}

Check ACL1 Is Changed and ACL2 not changed
    vpp_ctl: Check ACL Reply    agent_vpp_1    ${ACL1_NAME}    ${REPLY_DATA_FOLDER}/reply_acl1_update_tcp.txt    ${REPLY_DATA_FOLDER}/reply_acl1_update_tcp_term.txt

Delete ACL2
    vpp_ctl: Delete ACL     agent_vpp_1    ${ACL2_NAME}
    Sleep    ${SYNC_SLEEP}

Check ACL2 Is Deleted and ACL1 Is Not Changed
    vpp_ctl: Check ACL Reply    agent_vpp_1    ${ACL2_NAME}    ${REPLY_DATA_FOLDER}/reply_acl_empty.txt    ${REPLY_DATA_FOLDER}/reply_acl2_delete_tcp_term.txt

Delete ACL1
    vpp_ctl: Delete ACL     agent_vpp_1    ${ACL1_NAME}
    Sleep    ${SYNC_SLEEP}

Check ACL1 Is Deleted
    vpp_ctl: Check ACL Reply    agent_vpp_1    ${ACL1_NAME}    ${REPLY_DATA_FOLDER}/reply_acl_empty.txt   ${REPLY_DATA_FOLDER}/reply_acl_empty_term.txt


ADD ACL3_UDP
    vpp_ctl: Put ACL UDP    agent_vpp_1    ${ACL3_NAME}    ${E_INTF1}   ${I_INTF1}    ${E_INTF2}    ${I_INTF2}    ${RULE_NM3_1}    ${ACTION_DENY}    ${DEST_NTW}     ${SRC_NTW}   ${1DEST_PORT_L}   ${1DEST_PORT_U}    ${1SRC_PORT_L}     ${1SRC_PORT_U}
    Sleep    ${SYNC_SLEEP}

Check ACL3 Is Created
    vpp_ctl: Check ACL Reply    agent_vpp_1    ${ACL3_NAME}    ${REPLY_DATA_FOLDER}/reply_acl3_udp.txt    ${REPLY_DATA_FOLDER}/reply_acl3_udp_term.txt

ADD ACL4_UDP
    vpp_ctl: Put ACL UDP    agent_vpp_1    ${ACL4_NAME}    ${E_INTF1}    ${I_INTF1}    ${E_INTF2}    ${I_INTF2}    ${RULE_NM4_1}     ${ACTION_DENY}    ${DEST_NTW}     ${SRC_NTW}   ${1DEST_PORT_L}   ${1DEST_PORT_U}    ${1SRC_PORT_L}     ${1SRC_PORT_U}
    Sleep    ${SYNC_SLEEP}

Check ACL4 Is Created And ACL3 Still Configured
    vpp_ctl: Check ACL Reply    agent_vpp_1    ${ACL4_NAME}    ${REPLY_DATA_FOLDER}/reply_acl4_udp.txt     ${REPLY_DATA_FOLDER}/reply_acl4_udp_term.txt

Delete ACL4
    vpp_ctl: Delete ACL     agent_vpp_1    ${ACL4_NAME}
    Sleep    ${SYNC_SLEEP}

Check ACL4 Is Deleted and ACL3 Is Not Changed
    vpp_ctl: Check ACL Reply    agent_vpp_1    ${ACL4_NAME}   ${REPLY_DATA_FOLDER}/reply_acl_empty.txt     ${REPLY_DATA_FOLDER}/reply_acl3_udp_term.txt

Delete ACL3
    vpp_ctl: Delete ACL     agent_vpp_1    ${ACL3_NAME}
    Sleep    ${SYNC_SLEEP}

Check ACL3 Is Deleted
    vpp_ctl: Check ACL Reply    agent_vpp_1    ${ACL3_NAME}    ${REPLY_DATA_FOLDER}/reply_acl_empty.txt    ${REPLY_DATA_FOLDER}/reply_acl_empty_term.txt

ADD ACL5_ICMP
    vpp_ctl: Put ACL UDP    agent_vpp_1    ${ACL5_NAME}    ${E_INTF1}    ${I_INTF1}    ${E_INTF2}    ${I_INTF2}    ${RULE_NM5_1}    ${ACTION_DENY}    ${DEST_NTW}     ${SRC_NTW}   ${1DEST_PORT_L}   ${1DEST_PORT_U}    ${1SRC_PORT_L}     ${1SRC_PORT_U}
    Sleep    ${SYNC_SLEEP}

Check ACL5 Is Created
    vpp_ctl: Check ACL Reply    agent_vpp_1    ${ACL5_NAME}   ${REPLY_DATA_FOLDER}/reply_acl5_icmp.txt    ${REPLY_DATA_FOLDER}/reply_acl5_icmp_term.txt

ADD ACL6_ICMP
    vpp_ctl: Put ACL UDP    agent_vpp_1    ${ACL6_NAME}    ${E_INTF1}    ${I_INTF1}    ${E_INTF2}    ${I_INTF2}    ${RULE_NM6_1}    ${ACTION_DENY}  ${DEST_NTW}     ${SRC_NTW}   ${1DEST_PORT_L}   ${1DEST_PORT_U}    ${1SRC_PORT_L}     ${1SRC_PORT_U}
    Sleep    ${SYNC_SLEEP}

Check ACL6 Is Created And ACL5 Still Configured
    vpp_ctl: Check ACL Reply    agent_vpp_1    ${ACL6_NAME}    ${REPLY_DATA_FOLDER}/reply_acl6_icmp.txt    ${REPLY_DATA_FOLDER}/reply_acl6_icmp_term.txt

Delete ACL6
    vpp_ctl: Delete ACL     agent_vpp_1    ${ACL6_NAME}
    Sleep    ${SYNC_SLEEP}

Check ACL6 Is Deleted and ACL5 Is Not Changed
    vpp_ctl: Check ACL Reply    agent_vpp_1    ${ACL6_NAME}     ${REPLY_DATA_FOLDER}/reply_acl_empty.txt    ${REPLY_DATA_FOLDER}/reply_acl5_icmp_term.txt

Delete ACL5
    vpp_ctl: Delete ACL     agent_vpp_1    ${ACL5_NAME}
    Sleep    ${SYNC_SLEEP}

Check ACL5 Is Deleted
    vpp_ctl: Check ACL Reply    agent_vpp_1    ${ACL5_NAME}   ${REPLY_DATA_FOLDER}/reply_acl_empty.txt     ${REPLY_DATA_FOLDER}/reply_acl_empty_term.txt


Add 6 ACL
    vpp_ctl: Put ACL TCP   agent_vpp_1   ${ACL1_NAME}    ${E_INTF1}    ${I_INTF1}   ${RULE_NM1_1}    ${ACTION_DENY}     ${DEST_NTW}     ${SRC_NTW}   ${1DEST_PORT_L}   ${1DEST_PORT_U}    ${1SRC_PORT_L}     ${1SRC_PORT_U}
    vpp_ctl: Put ACL TCP   agent_vpp_1   ${ACL2_NAME}    ${E_INTF1}    ${I_INTF1}   ${RULE_NM2_1}    ${ACTION_DENY}     ${DEST_NTW}     ${SRC_NTW}   ${2DEST_PORT_L}   ${2DEST_PORT_U}    ${2SRC_PORT_L}     ${2SRC_PORT_U}
    vpp_ctl: Put ACL UDP   agent_vpp_1    ${ACL3_NAME}    ${E_INTF1}   ${I_INTF1}    ${E_INTF2}    ${I_INTF2}    ${RULE_NM3_1}    ${ACTION_DENY}    ${DEST_NTW}     ${SRC_NTW}   ${1DEST_PORT_L}   ${1DEST_PORT_U}    ${1SRC_PORT_L}     ${1SRC_PORT_U}
    vpp_ctl: Put ACL UDP   agent_vpp_1    ${ACL4_NAME}    ${E_INTF1}    ${I_INTF1}    ${E_INTF2}    ${I_INTF2}    ${RULE_NM4_1}     ${ACTION_DENY}    ${DEST_NTW}     ${SRC_NTW}   ${1DEST_PORT_L}   ${1DEST_PORT_U}    ${1SRC_PORT_L}     ${1SRC_PORT_U}
    vpp_ctl: Put ACL UDP   agent_vpp_1    ${ACL5_NAME}    ${E_INTF1}    ${I_INTF1}    ${E_INTF2}    ${I_INTF2}    ${RULE_NM5_1}    ${ACTION_DENY}    ${DEST_NTW}     ${SRC_NTW}   ${1DEST_PORT_L}   ${1DEST_PORT_U}    ${1SRC_PORT_L}     ${1SRC_PORT_U}
    vpp_ctl: Put ACL UDP   agent_vpp_1    ${ACL6_NAME}    ${E_INTF1}    ${I_INTF1}    ${E_INTF2}    ${I_INTF2}    ${RULE_NM6_1}    ${ACTION_DENY}  ${DEST_NTW}     ${SRC_NTW}   ${1DEST_PORT_L}   ${1DEST_PORT_U}    ${1SRC_PORT_L}     ${1SRC_PORT_U}

Check All 6 ACLs Added
    Check ACL All Reply    agent_vpp_1     ${REPLY_DATA_FOLDER}/reply_acl_all.txt        ${REPLY_DATA_FOLDER}/reply_acl_all_term.txt

*** Keywords ***

Check ACL All Reply
    [Arguments]         ${node}    ${reply_json}     ${reply_term}
    Log Many            ${node}    ${reply_json}     ${reply_term}
    ${acl_d}=           vpp_ctl: Get All ACL As Json    ${node}
    ${term_d}=          vat_term: Check All ACL     ${node}
    ${term_d_lines}=    Split To Lines    ${term_d}
    Log                 ${term_d_lines}
    ${data}=            OperatingSystem.Get File    ${reply_json}
    Should Be Equal     ${data}   ${acl_d}
    ${data}=            OperatingSystem.Get File    ${reply_term}
    ${t_data_lines}=    Split To Lines    ${data}
    Log                 ${t_data_lines}
    List Should Contain Sub List    ${term_d_lines}    ${t_data_lines}


TestSetup
    Make Datastore Snapshots    ${TEST_NAME}_test_setup

TestTeardown
    Make Datastore Snapshots    ${TEST_NAME}_test_teardown

Suite Cleanup
    Stop SFC Controller Container
    Testsuite Teardown