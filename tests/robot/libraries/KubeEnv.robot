*** Settings ***
Documentation     This is a library to handle actions related to kubernetes cluster,
...               such as kubernetes setup  etc.
...
...               The code is aimed at few selected deployments:
...               a: 1-node 10-pods, nospecific applications, test use ping and nc to check connectivity.
...
...               This Resource manages the following suite variables:
...               ${testbed_connection} SSH connection index towards host in 1-node k8s cluster.
...               #${client_pod_name} client pod name assigned by k8s in 1-node 2-pod scenario.
...               #${server_pod_name} server pod name assigned by k8s in 1-node 2-pod scenario.
Resource          ${CURDIR}/all_libs.robot

*** Variables ***
${ETCD_YAML_FILE_PATH}    ${CURDIR}/../resources/k8-yaml/etcd-k8.yaml
${KAFKA_YAML_FILE_PATH}    ${CURDIR}/../resources/k8-yaml/kafka-k8.yaml
${SFC_YAML_FILE_PATH}    ${CURDIR}/../resources/k8-yaml/sfc-k8.yaml
${VSWITCH_YAML_FILE_PATH}    ${CURDIR}/../resources/k8-yaml/vswitch-k8.yaml
${CN_INFRA_YAML_FILE_PATH}    ${CURDIR}/../resources/k8-yaml/dev-cn-infra-k8.yaml
${PULL_IMAGES_PATH}    ${CURDIR}/../resources/k8-scripts/pull-images.sh

${POD_DEPLOY_APPEARS_TIMEOUT}    30s
${POD_REMOVE_DEFAULT_TIMEOUT}    60s

*** Keywords ***
# TODO: Passing ${ssh_session} around is annoying. Make keywords assume the correct SSH session is already active.

# TODO: Exact steps should be investigated how to reinit Kubernetes properly
Reinit_1_Node_Cluster
    [Documentation]    Assuming active SSH connection, store its index, execute multiple commands to reinstall and restart 1node cluster, wait to see it running.
    ${normal_tag}    ${vpp_tag} =    Get_Docker_Tags
    BuiltIn.Set_Suite_Variable    ${testbed_connection}    ${VM_SSH_ALIAS_PREFIX}1
#    ${conn} =     SSHLibrary.Get_Connection    ${VM_SSH_ALIAS_PREFIX}1
#    Set_Suite_Variable    ${testbed_connection}    ${conn.index}
#    SSHLibrary.Set_Client_Configuration    timeout=${SSH_TIMEOUT}    prompt=$
    SshCommons.Switch_And_Execute_Command    ${testbed_connection}    sudo rm -rf $HOME/.kube
    KubeAdm.Reset    ${testbed_connection}
    Uninstall_Cri    ${normal_tag}
    Docker_Pull_Images    ${normal_tag}    ${vpp_tag}
    Install_Cri    ${normal_tag}
    ${stdout} =    KubeAdm.Init    ${testbed_connection}
    BuiltIn.Log    ${stdout}
    BuiltIn.Should_Contain    ${stdout}    Your Kubernetes master has initialized successfully
    SshCommons.Execute_Command_And_Log    mkdir -p $HOME/.kube
    SshCommons.Execute_Command_And_Log    sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
    SshCommons.Execute_Command_And_Log    sudo chown $(id -u):$(id -g) $HOME/.kube/config
    KubeCtl.Taint    ${testbed_connection}    nodes --all node-role.kubernetes.io/master-

Docker_Pull_Vpp_Agent
    [Arguments]    ${normal_tag}    ${vpp_tag}
    [Documentation]    Execute bash after applying edits to pull-images.sh.
    BuiltIn.Log_Many    ${normal_tag}    ${vpp_tag}
    ${file_path} =    BuiltIn.Set_Variable    ${RESULTS_FOLDER}/pull-images.sh
    # TODO: Add error checking for OperatingSystem calls.
    OperatingSystem.Run    cp -f ${PULL_IMAGES_PATH} ${file_path}
    OperatingSystem.Run    sed -i 's@vswitch:latest@vswitch:${vpp_tag}@g' ${file_path}
    OperatingSystem.Run    sed -i 's@:latest@:${normal_tag}@g' ${file_path}
    SshCommons.Execute_Command_With_Copied_File    ${file_path}    bash

Verify_All_Pods_Running
    [Arguments]    ${ssh_session}    ${excluded_pod_prefix}=invalid-pod-prefix-
    [Documentation]     Iterate over all pods of all namespaces (skipping \${excluded_pod_prefix} matches) and check running state.
    BuiltIn.Log_Many    ${ssh_session}    ${excluded_pod_prefix}
    ${all_pods_dict} =    KubeCtl.Get_Pods_All_Namespaces    ${ssh_session}
    ${pod_names} =    Collections.Get_Dictionary_Keys    ${all_pods_dict}
    : FOR    ${pod_name}   IN    @{pod_names}
    \     BuiltIn.Continue_For_Loop_If    """${excluded_pod_prefix}""" in """${pod_name}"""
    \     ${namesp} =    BuiltIn.Evaluate    &{all_pods_dict}[${pod_name}]['NAMESPACE']
    \     Verify_Pod_Running_And_Ready    ${ssh_session}    ${pod_name}    namespace=${namesp}

Verify_K8s_Running
    [Arguments]    ${ssh_session}
    [Documentation]     We check for a particular (hardcoded) number of pods after init. Might be later replaced with
    ...    more detailed asserts.
    BuiltIn.Log_Many    ${ssh_session}
    BuiltIn.Comment    TODO: Make the expected number of pods configurable.
    ${all_pods_dict} =    KubeCtl.Get_Pods_All_Namespaces    ${ssh_session}
    BuiltIn.Length_Should_Be   ${all_pods_dict}     9
    Verify_All_Pods_Running    ${ssh_session}

Get_Pod_Name_List_By_Prefix
    [Arguments]    ${ssh_session}    ${pod_prefix}
    [Documentation]    Get pods from all namespaces, parse with specified \${pod_prefix}, log and return the parsed result.
    BuiltIn.Log_Many    ${ssh_session}    ${pod_prefix}
    BuiltIn.Comment    TODO: Unify with Get_Pods or Get_Pods_All_Namespaces in KubeCtl.
    ${stdout} =    SshCommons.Switch_And_Execute_Command    ${ssh_session}    kubectl get pods
    ${output} =    kube_parser.parse_kubectl_get_pods_and_get_pod_name    ${stdout}    ${pod_prefix}
    Builtin.Log    ${output}
    [Return]    ${output}

Deploy_Etcd_Kafka_And_Verify_Running
    [Arguments]    ${ssh_session}    ${etcd_file}=${ETCD_YAML_FILE_PATH}    ${kafka_file}=${KAFKA_YAML_FILE_PATH}
    [Documentation]     Deploy and verify ETCD and KAFKA pods and store its name.
    BuiltIn.Log_Many    ${ssh_session}    ${etcd_file}    ${kafka_file}
    ${etcd_pod_name} =    Deploy_Pod_And_Verify_Running    ${ssh_session}    ${etcd_file}    ubuntu-client-etcd    timeout=${POD_DEPLOY_TIMEOUT}
    ${kafka_pod_name} =    Deploy_Pod_And_Verify_Running    ${ssh_session}    ${kafka_file}    kafka-    timeout=${POD_DEPLOY_TIMEOUT}
    BuiltIn.Set_Suite_Variable    ${etcd_pod_name}
    BuiltIn.Set_Suite_Variable    ${kafka_pod_name}

Deploy_Vswitch_Pod_And_Verify_Running
    [Arguments]    ${ssh_session}    ${vswitch_file}=${VSWITCH_YAML_FILE_PATH}
    [Documentation]     Deploy and verify switch pod and store its name.
    BuiltIn.Log_Many    ${ssh_session}    ${vswitch_file}
    ${vswitch_pod_name} =    Deploy_Pod_And_Verify_Running    ${ssh_session}    ${vswitch_file}    vswitch-    timeout=${POD_DEPLOY_TIMEOUT}
    BuiltIn.Set_Suite_Variable    ${vswitch_pod_name}

Deploy_SFC_Pod_And_Verify_Running
    [Arguments]    ${ssh_session}    ${sfc_file}=${SFC_YAML_FILE_PATH}
    [Documentation]     Deploy and verify switch pod and store its name.
    BuiltIn.Log_Many    ${ssh_session}    ${sfc_file}
    ${sfc_pod_name} =    Deploy_Pod_And_Verify_Running    ${ssh_session}    ${sfc_file}    sfc-    timeout=${POD_DEPLOY_TIMEOUT}
    BuiltIn.Set_Suite_Variable    ${sfc_pod_name}


Deploy_Cn-Infra_Pod_And_Verify_Running
    [Arguments]    ${ssh_session}    ${cn-infra_file}=${CN_INFRA_YAML_FILE_PATH}
    [Documentation]     Deploy and verify cn-infra pod and store its name.
    BuiltIn.Log_Many    ${ssh_session}    ${cn-infra_file}
    ${cn_infra_pod_name} =    Deploy_Pod_And_Verify_Running    ${ssh_session}    ${cn-infra_file}    ubuntu-client-    timeout=${POD_DEPLOY_TIMEOUT}
    BuiltIn.Set_Suite_Variable    ${cn_infra_pod_name}


Remove_VSwitch_Pod_And_Verify_Removed
    [Arguments]    ${ssh_session}    ${vswitch_file}=${VSWITCH_YAML_FILE_PATH}
    [Documentation]    Execute delete commands, wait until  pod is removed.
    BuiltIn.Log_Many    ${ssh_session}    ${vswitch_file}
    KubeCtl.Delete_F    ${ssh_session}    ${vswitch_file}
    Wait_Until_Pod_Removed    ${ssh_session}    ${vswitch_pod_name}

Remove_SFC_Pod_And_Verify_Removed
    [Arguments]    ${ssh_session}    ${sfc_file}=${SFC_YAML_FILE_PATH}
    [Documentation]    Execute delete commands, wait until  pod is removed.
    BuiltIn.Log_Many    ${ssh_session}    ${sfc_file}
    KubeCtl.Delete_F    ${ssh_session}    ${sfc_file}
    Wait_Until_Pod_Removed    ${ssh_session}    ${sfc_pod_name}

Remove_Cn-Infra_Pod_And_Verify_Removed
    [Arguments]    ${ssh_session}    ${cn_infra_file}=${CN_INFRA_YAML_FILE_PATH}
    [Documentation]    Execute delete commands, wait until  pod is removed.
    BuiltIn.Log_Many    ${ssh_session}    ${cn_infra_file}
    KubeCtl.Delete_F    ${ssh_session}    ${cn_infra_file}
    Wait_Until_Pod_Removed    ${ssh_session}    ${cn_infra_pod_name}

Remove_ETCD And_KAFKA_Pod_And_Verify_Removed
    [Arguments]    ${ssh_session}    ${etcd_file}=${ETCD_YAML_FILE_PATH}    ${kafka_file}=${KAFKA_YAML_FILE_PATH}
    [Documentation]    Execute delete commands, wait until  pod is removed.
    BuiltIn.Log_Many    ${ssh_session}    ${etcd_file}    ${kafka_file}
    KubeCtl.Delete_F    ${ssh_session}    ${etcd_file}
    Wait_Until_Pod_Removed    ${ssh_session}    ${etcd_pod_name}
    KubeCtl.Delete_F    ${ssh_session}    ${kafka_file}
    Wait_Until_Pod_Removed    ${ssh_session}    ${kafka_pod_name}

Verify_Multireplica_Pods_Running
    [Arguments]    ${ssh_session}    ${pod_prefix}    ${nr_replicas}    ${namespace}
    [Documentation]     Check there is expected number of pods and they are running.
    BuiltIn.Log_Many    ${ssh_session}    ${pod_prefix}    ${nr_replicas}    ${namespace}
    BuiltIn.Comment    TODO: Join single- and multi- replica keywords.
    ${pods_list} =    Get_Pod_Name_List_By_Prefix    ${ssh_session}    ${pod_prefix}
    BuiltIn.Length_Should_Be   ${pods_list}     ${nr_replicas}
    : FOR    ${pod_name}    IN    @{pods_list}
    \    Verify_Pod_Running_And_Ready    ${ssh_session}    ${pod_name}    namespace= ${namespace}
    BuiltIn.Return_From_Keyword    ${pods_list}

Deploy_Multireplica_Pods_And_Verify_Running
    [Arguments]    ${ssh_session}    ${pod_file}    ${pod_prefix}    ${nr_replicas}    ${namespace}=default    ${setup_timeout}=${POD_DEPLOY_MULTIREPLICA_TIMEOUT}
    [Documentation]     Apply the provided yaml file with more replica specified, wait until pods are running, return pods details.
    BuiltIn.Log_Many    ${ssh_session}    ${pod_file}    ${pod_prefix}    ${nr_replicas}    ${namespace}    ${setup_timeout}
    BuiltIn.Comment    TODO: Join single- and multi- replica keywords.
    KubeCtl.Apply_F    ${ssh_session}    ${pod_file}
    ${pods_details} =    BuiltIn.Wait_Until_Keyword_Succeeds    ${setup_timeout}   4s    Verify_Multireplica_Pods_Running    ${ssh_session}    ${pod_prefix}    ${nr_replicas}    ${namespace}
    BuiltIn.Set_Suite_Variable    ${pods_details}

Verify_Multireplica_Pods_Removed
    [Arguments]    ${ssh_session}    ${pod_prefix}
    [Documentation]     Check no pods are running with prefix: ${pod_prefix}
    BuiltIn.Log_Many    ${ssh_session}    ${pod_prefix}
    BuiltIn.Comment    TODO: Join single- and multi- replica keywords.
    ${pods_list} =    Get_Pod_Name_List_By_Prefix    ${ssh_session}    ${pod_prefix}
    BuiltIn.Length_Should_Be   ${pods_list}     0

Remove_Multireplica_Pods_And_Verify_Removed
    [Arguments]    ${ssh_session}    ${pod_file}    ${pod_prefix}
    [Documentation]     Remove pods and verify they are removed.
    BuiltIn.Log_Many    ${ssh_session}    ${pod_file}    ${pod_prefix}
    KubeCtl.Delete_F    ${ssh_session}    ${pod_file}
    BuiltIn.Wait_Until_Keyword_Succeeds    ${POD_REMOVE_MULTIREPLICA_TIMEOUT}    5s    Verify_Multireplica_Pods_Removed    ${ssh_session}    ${pod_prefix}


Remove_NonVPP_Pod_And_Verify_Removed
    [Arguments]    ${ssh_session}    ${nginx_file}=${NGINX_POD_FILE}
    [Documentation]    Remove pod and verify removal, nginx being the default file.
    BuiltIn.Log_Many    ${ssh_session}    ${nginx_file}
    KubeCtl.Delete_F    ${ssh_session}    ${nginx_file}
    Wait_Until_Pod_Removed    ${ssh_session}    ${nginx_pod_name}

Get_Deployed_Pod_Name
    [Arguments]    ${ssh_session}    ${pod_prefix}
    [Documentation]    Get list of pod names matching the prefix, check there is just one, return the name.
    BuiltIn.Log_Many    ${ssh_session}    ${pod_prefix}
    ${pod_name_list} =   Get_Pod_Name_List_By_Prefix    ${ssh_session}    ${pod_prefix}
    BuiltIn.Length_Should_Be    ${pod_name_list}    1
    ${pod_name} =    BuiltIn.Evaluate     ${pod_name_list}[0]
    [Return]    ${pod_name}

Deploy_Pod_And_Verify_Running
    [Arguments]    ${ssh_session}    ${pod_file}    ${pod_prefix}    ${timeout}=${POD_DEPLOY_DEFAULT_TIMEOUT}
    [Documentation]    Deploy pod defined by \${pod_file}, wait until a pod matching \${pod_prefix} appears, check it was only 1 such pod, extract its name, wait until it is running, log and return the name.
    Builtin.Log_Many    ${ssh_session}    ${pod_file}    ${pod_prefix}
    KubeCtl.Apply_F    ${ssh_session}    ${pod_file}
    ${pod_name} =    BuiltIn.Wait_Until_Keyword_Succeeds    ${POD_DEPLOY_APPEARS_TIMEOUT}    2s    Get_Deployed_Pod_Name    ${ssh_session}    ${pod_prefix}
    Wait_Until_Pod_Running    ${ssh_session}    ${pod_name}    timeout=${timeout}
    BuiltIn.Log    ${pod_name}
    [Return]    ${pod_name}

Remove_Pod_And_Verify_Removed
    [Arguments]    ${ssh_session}    ${pod_file}    ${pod_name}
    [Documentation]    Remove pod defined by \${pod_file}, wait for \${pod_name} to get removed.
    BuiltIn.Log_Many    ${ssh_session}    ${pod_file}    ${pod_name}
    KubeCtl.Delete_F    ${ssh_session}    ${pod_file}
    Wait_Until_Pod_Removed    ${ssh_session}    ${pod_name}

Verify_Pod_Running_And_Ready
    [Arguments]    ${ssh_session}    ${pod_name}    ${namespace}=default
    [Documentation]    Get pods of \${namespace}, parse status of \${pod_name}, check it is Running, parse for ready containes of \${pod_name}, check it is all of them.
    BuiltIn.Log_Many    ${ssh_session}    ${pod_name}    ${namespace}
    &{pods} =     KubeCtl.Get_Pods    ${ssh_session}    namespace=${namespace}
    ${status} =    BuiltIn.Evaluate    &{pods}[${pod_name}]['STATUS']
    BuiltIn.Should_Be_Equal_As_Strings    ${status}    Running
    ${ready} =    BuiltIn.Evaluate    &{pods}[${pod_name}]['READY']
    ${ready_containers}    ${out_of_containers} =    String.Split_String    ${ready}    separator=${/}    max_split=1
    BuiltIn.Should_Be_Equal_As_Strings    ${ready_containers}    ${out_of_containers}

Wait_Until_Pod_Running
    [Arguments]    ${ssh_session}    ${pod_name}    ${timeout}=${POD_RUNNING_DEFAULT_TIMEOUT}    ${check_period}=5s    ${namespace}=default
    [Documentation]    WUKS around Verify_Pod_Running_And_Ready.
    BuiltIn.Log_Many    ${ssh_session}    ${pod_name}    ${timeout}    ${check_period}    ${namespace}
    BuiltIn.Wait_Until_Keyword_Succeeds    ${timeout}    ${check_period}    Verify_Pod_Running_And_Ready    ${ssh_session}    ${pod_name}    namespace=${namespace}

Verify_Pod_Not_Present
    [Arguments]    ${ssh_session}    ${pod_name}    ${namespace}=default
    [Documentation]    Get pods for \${namespace}, check \${pod_name} is not one of them.
    BuiltIn.Log_Many    ${ssh_session}    ${pod_name}    ${namespace}
    ${pods} =     KubeCtl.Get_Pods    ${ssh_session}    namespace=${namespace}
    Collections.Dictionary_Should_Not_Contain_Key     ${pods}    ${pod_name}

Wait_Until_Pod_Removed
    [Arguments]    ${ssh_session}    ${pod_name}    ${timeout}=${POD_REMOVE_DEFAULT_TIMEOUT}    ${check_period}=5s    ${namespace}=default
    [Documentation]    WUKS around Verify_Pod_Not_Present.
    BuiltIn.Log_Many    ${ssh_session}    ${pod_name}    ${timeout}    ${check_period}    ${namespace}
    BuiltIn.Wait_Until_Keyword_Succeeds    ${timeout}    ${check_period}    Verify_Pod_Not_Present    ${ssh_session}    ${pod_name}    namespace=${namespace}

Run_Finite_Command_In_Pod
    [Arguments]    ${command}    ${ssh_session}=${EMPTY}    ${prompt}=${EMPTY}
    [Documentation]    Switch if \${ssh_session}, configure if \${prompt}, write \${command}, read until prompt, log and return text output.
    BuiltIn.Log_Many    ${command}     ${ssh_session}     ${prompt}
    BuiltIn.Comment    TODO: Do not mention pods and move to SshCommons.robot or similar.
    BuiltIn.Run_Keyword_If    """${ssh_session}""" != """${EMPTY}"""     SSHLibrary.Switch_Connection    ${ssh_session}
    BuiltIn.Run_Keyword_If    """${prompt}""" != """${EMPTY}"""    SSHLibrary.Set_Client_Configuration    prompt=${prompt}
    SSHLibrary.Write    ${command}
    ${output} =     SSHLibrary.Read_Until_Prompt
    SshCommons.Append_Command_Log    ${command}    ${output}
    [Return]    ${output}

Init_Infinite_Command_In_Pod
    [Arguments]    ${command}    ${ssh_session}=${EMPTY}    ${prompt}=${EMPTY}
    [Documentation]    Switch if \${ssh_session}, configure if \${prompt}, write \${command}.
    BuiltIn.Log_Many    ${command}    ${ssh_session}    ${prompt}
    BuiltIn.Comment    TODO: Do not mention pods and move to SshCommons.robot or similar.
    BuiltIn.Run_Keyword_If    """${ssh_session}""" != """${EMPTY}"""     SSHLibrary.Switch_Connection    ${ssh_session}
    BuiltIn.Run_Keyword_If    """${prompt}""" != """${EMPTY}"""    SSHLibrary.Set_Client_Configuration    prompt=${prompt}
    SSHLibrary.Write    ${command}
    SshCommons.Append_Command_Log    ${command}

Stop_Infinite_Command_In_Pod
    [Arguments]    ${ssh_session}=${EMPTY}     ${prompt}=${EMPTY}
    [Documentation]    Switch if \${ssh_session}, configure if \${prompt}, write ctrl+c, read until prompt, log and return output.
    BuiltIn.Log_Many    ${ssh_session}    ${prompt}
    BuiltIn.Comment    TODO: Do not mention pods and move to SshCommons.robot or similar.
    BuiltIn.Run_Keyword_If    """${ssh_session}""" != """${EMPTY}"""     SSHLibrary.Switch_Connection    ${ssh_session}
    BuiltIn.Run_Keyword_If    """${prompt}""" != """${EMPTY}"""    SSHLibrary.Set_Client_Configuration    prompt=${prompt}
    Write_Bare_Ctrl_C
    ${output1} =     SSHLibrary.Read_Until    ^C
    ${output2} =     SSHLibrary.Read_Until_Prompt
    BuiltIn.Log_Many     ${output1}    ${output2}
    ${output} =    Builtin.Set_Variable    ${output1}${output2}
    SshCommons.Append_Command_Log    ^C    ${output}
    [Return]    ${output}

Write_Bare_Ctrl_C
    [Documentation]    Construct ctrl+c character and SSH-write it (without endline) to the current SSH connection.
    ...    Do not read anything yet.
    BuiltIn.Comment    TODO: Move to SshCommons.robot or similar.
    ${ctrl_c} =    BuiltIn.Evaluate    chr(int(3))
    SSHLibrary.Write_Bare    ${ctrl_c}

Get_Into_Container_Prompt_In_Pod
    [Arguments]    ${ssh_session}    ${pod_name}    ${prompt}=${EMPTY}
    [Documentation]    Configure if prompt, execute interactive bash in ${pod_name}, read until prompt, log and return output.
    BuiltIn.Log_Many    ${ssh_session}    ${pod_name}    ${prompt}
    # TODO: PodBash.robot?
    ${docker} =    BuiltIn.Set_Variable    ${K8_CLUSTER_${CLUSTER_ID}_DOCKER_COMMAND}
    ${container_id} =    KubeCtl.Get_Container_Id    ${ssh_session}    ${pod_name}
    # That already switched the ssh session.
    BuiltIn.Run_Keyword_If    """${prompt}""" != """${EMPTY}"""    SSHLibrary.Set_Client_Configuration    prompt=${prompt}
    ${command} =    BuiltIn.Set_Variable    ${docker} exec -i -t --privileged=true ${container_id} /bin/sh
    SSHLibrary.Write    ${command}
    ${output} =     SSHLibrary.Read_Until_Prompt
    SshCommons.Append_Command_Log    ${command}    ${output}
    [Return]    ${output}

Leave_Container_Prompt_In_Pod
    [Arguments]     ${ssh_session}    ${prompt}=$
    [Documentation]    Configure prompt, send ctrl+c, write "exit", read until prompt, log and return output.
    BuiltIn.Log_Many    ${ssh_session}    ${prompt}
    # TODO: PodBash.robot?
    SSHLibrary.Switch_Connection    ${ssh_session}
    SSHLibrary.Set_Client_Configuration    prompt=${prompt}
    Write_Bare_Ctrl_C
    SSHLibrary.Write    exit
    ${output} =     SSHLibrary.Read_Until_Prompt
    SshCommons.Append_Command_Log    ^Cexit    ${output}
    [Return]    ${output}

Verify_Cluster_Node_Ready
    [Arguments]    ${ssh_session}    ${node_name}
    [Documentation]    Get nodes, parse status of \${node_name}, check it is Ready, return nodes.
    BuiltIn.Log_Many    ${ssh_session}    ${node_name}
    BuiltIn.Comment    FIXME: Avoid repeated get_nodes when called from Verify_Cluster_Ready.
    ${nodes} =    KubeCtl.Get_Nodes    ${ssh_session}
    BuiltIn.Log    ${nodes}
    ${status} =    BuiltIn.Evaluate    &{nodes}[${node_name}]['STATUS']
    BuiltIn.Should_Be_Equal    ${status}    Ready
    [Return]    ${nodes}

Verify_Cluster_Ready
    [Arguments]     ${ssh_session}    ${nr_nodes}
    [Documentation]    Get nodes, check there are \${nr_nodes}, for each node Verify_Cluster_Node_Ready.
    BuiltIn.Log_Many     ${ssh_session}    ${nr_nodes}
    ${nodes} =    KubeCtl.Get_Nodes    ${ssh_session}
    BuiltIn.Log    ${nodes}
    BuiltIn.Length_Should_Be    ${nodes}    ${nr_nodes}
    ${names} =     Collections.Get_Dictionary_Keys     ${nodes}
    : FOR    ${name}    IN    @{names}
    \    Verify_Cluster_Node_Ready    ${ssh_session}    ${name}

Wait_Until_Cluster_Ready
    [Arguments]    ${ssh_session}    ${nr_nodes}    ${timeout}=${CLUSTER_READY_TIMEOUT}    ${check_period}=5s
    [Documentation]    WUKS around Verify_Cluster_Ready.
    BuiltIn.Log_Many    ${ssh_session}    ${nr_nodes}    ${timeout}    ${check_period}
    BuiltIn.Wait_Until_Keyword_Succeeds    ${timeout}    ${check_period}    Verify_Cluster_Ready    ${ssh_session}    ${nr_nodes}

Log_Etcd
    [Arguments]    ${ssh_session}
    [Documentation]    Check there is exactly one etcd pod, get its logs
    ...    (and do nothing with them, except the implicit Log).
    Builtin.Log_Many    ${ssh_session}
    ${pod_list} =    Get_Pod_Name_List_By_Prefix    ${ssh_session}    etcd
    BuiltIn.Log    ${pod_list}
    BuiltIn.Length_Should_Be    ${pod_list}    1
    KubeCtl.Logs    ${ssh_session}    @{pod_list}[0]    namespace=default

Log_Vswitch
    [Arguments]    ${ssh_session}    ${exp_nr_vswitch}=${KUBE_CLUSTER_${CLUSTER_ID}_NODES}
    [Documentation]    Check there is expected number of vswitch pods, get logs from them an cn-infra containers
    ...    (and do nothing except the implicit Log).
    Builtin.Log_Many    ${ssh_session}    ${exp_nr_vswitch}
    ${pod_list} =    Get_Pod_Name_List_By_Prefix    ${ssh_session}    vswitch
    BuiltIn.Log    ${pod_list}
    BuiltIn.Length_Should_Be    ${pod_list}    ${exp_nr_vswitch}
    : FOR    ${vswitch_pod}    IN    @{pod_list}
    # \    KubeCtl.Logs    ${ssh_session}    ${vswitch_pod}    namespace=default    container=cn-infra
    \    KubeCtl.Logs    ${ssh_session}    ${vswitch_pod}    namespace=default    container=vswitch

Log_Kube_Dns
    [Arguments]    ${ssh_session}
    [Documentation]    Check there is exactly one dns pod, get logs from kubedns, dnsmasq and sidecar containers
    ...    (and do nothing with them, except the implicit Log).
    Builtin.Log_Many    ${ssh_session}
    ${pod_list} =    Get_Pod_Name_List_By_Prefix    ${ssh_session}    kube-dns-
    BuiltIn.Log    ${pod_list}
    BuiltIn.Length_Should_Be    ${pod_list}    1
    KubeCtl.Logs    ${ssh_session}    @{pod_list}[0]    namespace=default    container=kubedns

Log_Pods_For_Debug
    [Arguments]    ${ssh_session}    ${exp_nr_vswitch}=${KUBE_CLUSTER_${CLUSTER_ID}_NODES}
    [Documentation]    Call multiple keywords to get various logs
    ...    (and do nothing with them, except the implicit Log).
    Builtin.Log_Many    ${ssh_session}    ${exp_nr_vswitch}
    Log_Etcd    ${ssh_session}
    Log_Vswitch    ${ssh_session}    ${exp_nr_vswitch}
    Log_Kube_Dns    ${ssh_session}

Open_Connection_To_Node
    [Arguments]    ${name}    ${cluster_id}    ${node_index}
    BuiltIn.Log_Many    ${name}    ${node_index}
    ${connection}=    SshCommons.Open_Ssh_Connection_Kube    ${name}    ${K8_CLUSTER_${cluster_id}_VM_${node_index}_PUBLIC_IP}    ${K8_CLUSTER_${cluster_id}_VM_${node_index}_USER}    ${K8_CLUSTER_${cluster_id}_VM_${node_index}_PSWD}
    [Return]    ${connection}
