*** Settings ***
Documentation     Test suite to test basic ping, udp, tcp on basic ccmts topology inlcd. vswitch restart.
Library      OperatingSystem

Library      Collections

Resource     ../../../variables/${VARIABLES}_variables.robot

Resource     ../../../libraries/all_libs.robot

*** Variables ***
${REPLY_DATA_FOLDER}            replyACL
${VARIABLES}=       common
${ENV}=             common
${CLUSTRER_ID}=     CCMTS1

Suite Setup       BasicCcmtsSetup
Suite Teardown    BasicCcmtsTeardown
Test Setup        TestSetup
Test Teardown     TestTeardown

*** Test Cases ***
Pod_To_Pod_Ping
    [Documentation]    Execute "ping -c 5" command between pods (both ways), require no packet loss.
    [Setup]    Setup_Hosts_Connections
    ${stdout} =    KubeEnv.Run_Finite_Command_In_Pod    ping -c 5 ${server_ip}    ssh_session=${client_connection}
    BuiltIn.Should_Contain   ${stdout}    5 received, 0% packet loss
    ${stdout} =    KubeEnv.Run_Finite_Command_In_Pod    ping -c 5 ${client_ip}    ssh_session=${server_connection}
    BuiltIn.Should_Contain   ${stdout}    5 received, 0% packet loss
    [Teardown]    Teardown_Hosts_Connections

Pod_To_Pod_Udp
    [Documentation]    Start UDP server and client, send message, stop both and check the message has been reseived.
    [Setup]    Setup_Hosts_Connections
    KubernetesEnv.Init_Infinite_Command_in_Pod    nc -ul -p 7000    ssh_session=${server_connection}
    KubernetesEnv.Init_Infinite_Command_in_Pod    nc -u ${server_ip} 7000    ssh_session=${client_connection}
    ${text} =    BuiltIn.Set_Variable    Text to be received
    SSHLibrary.Write    ${text}
    ${client_stdout} =    KubeEnv.Stop_Infinite_Command_In_Pod    ssh_session=${client_connection}
    ${server_stdout} =    KubeEnv.Stop_Infinite_Command_In_Pod    ssh_session=${server_connection}
    BuiltIn.Should_Contain   ${server_stdout}    ${text}
    [Teardown]    Teardown_Hosts_Connections

Pod_To_Pod_Tcp
    [Documentation]    Start TCP server, start client sending the message, stop server, check message has been received, stop client.
    [Setup]    Setup_Hosts_Connections
    ${text} =    BuiltIn.Set_Variable    Text to be received
    KubeEnv.Run_Finite_Command_In_Pod    cd; echo "${text}" > some.file    ssh_session=${client_connection}
    KubeEnv.Init_Infinite_Command_in_Pod    nc -l -p 4444    ssh_session=${server_connection}
    KubeEnv.Init_Infinite_Command_in_Pod    cd; nc ${server_ip} 4444 < some.file    ssh_session=${client_connection}
    ${server_stdout} =    KubeEnv.Stop_Infinite_Command_In_Pod    ssh_session=${server_connection}
    BuiltIn.Should_Contain   ${server_stdout}    ${text}
    ${client_stdout} =    KubeEnv.Stop_Infinite_Command_In_Pod    ssh_session=${client_connection}
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
    KubeEnv.Init_Infinite_Command_in_Pod    nc -ul -p 7000    ssh_session=${server_connection}
    KubeEnv.Init_Infinite_Command_in_Pod    nc -u ${server_ip} 7000    ssh_session=${testbed_connection}
    ${text} =    BuiltIn.Set_Variable    Text to be received
    SSHLibrary.Write    ${text}
    ${client_stdout} =    KubeEnv.Stop_Infinite_Command_In_Pod    ssh_session=${testbed_connection}    prompt=$
    ${server_stdout} =    KubeEnv.Stop_Infinite_Command_In_Pod    ssh_session=${server_connection}
    BuiltIn.Should_Contain   ${server_stdout}    ${text}
    [Teardown]    Teardown_Hosts_Connections

Host_To_Pod_Tcp
    [Documentation]    The same as Pod_To_Pod_Tcp but client is on host instead of pod.
    [Setup]    Setup_Hosts_Connections
    ${text} =    BuiltIn.Set_Variable    Text to be received
    KubeEnv.Run_Finite_Command_In_Pod    cd; echo "${text}" > some.file    ssh_session=${testbed_connection}
    KubeEnv.Init_Infinite_Command_in_Pod    nc -l -p 4444    ssh_session=${server_connection}
    KubeEnv.Init_Infinite_Command_in_Pod    cd; nc ${server_ip} 4444 < some.file    ssh_session=${testbed_connection}
    ${server_stdout} =    KubeEnv.Stop_Infinite_Command_In_Pod    ssh_session=${server_connection}
    BuiltIn.Should_Contain   ${server_stdout}    ${text}
    ${client_stdout} =    KubeEnv.Stop_Infinite_Command_In_Pod    ssh_session=${testbed_connection}    prompt=$
    [Teardown]    Teardown_Hosts_Connections

*** Keywords ***
Cleanup_Basic_Ccmts_Deployment_On_Cluster
    [Documentation]    Assuming active SSH connection, store its index, execute multiple commands to cleanup 1node cluster, wait for running.
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
    setup-teardown.Testsuite_K8Setup    ${CLUSTRER_ID}
    #KubeEnv.Reinit_1_Node_Cluster
    Cleanup_Basic_Ccmts_Deployment_On_Cluster
    KubeEnv.Deploy_Etcd_Kafka_And_Verify_Running    ${testbed_connection}
    KubeEnv.Deploy_Vswitch_Pod_And_Verify_Running    ${testbed_connection}
    KubeEnv.Deploy_SFC_Pod_And_Verify_Running    ${testbed_connection}
    KubeEnv.Deploy_Cn-Infra_Pod_And_Verify_Running    ${testbed_connection}

BasicCcmtsTeardown
    [Documentation]    Log leftover output from pods, remove pods, execute common teardown.
    KubeEnv.Log_Pods_For_Debug    ${testbed_connection}    exp_nr_vswitch=1
    KubeEnv.Remove_Etcd_Kafka_And_Verify_Removed    ${testbed_connection}
    KubeEnv.Remove_VSwitch_Pod_And_Verify_Removed   ${testbed_connection}
    KubeEnv.Remove_SFC_Pod_And_Verify_Removed   ${testbed_connection}
    KubeEnv.Remove_Cn-Infra_Pod_And_Verify_Removed   ${testbed_connection}
    setup-teardown.Testsuite_K8Teardown

Setup_Hosts_Connections
    [Documentation]    Open and store two more SSH connections to master host, in them open
    ...    pod shells to client and server pod, parse their IP addresses and store them.
    KubeEnvConnections.Open_CCMTS1_Connection

Teardown_Hosts_Connections
    [Documentation]    Exit pod shells, close corresponding SSH connections.
    KubeEnvConnections.Close_CCMTS1_Connection

TestSetup
    Make Datastore Snapshots    ${TEST_NAME}_test_setup

TestTeardown
    Make Datastore Snapshots    ${TEST_NAME}_test_teardown

