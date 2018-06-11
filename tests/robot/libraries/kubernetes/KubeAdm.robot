*** Settings ***
Documentation     This is a library to handle kubeadm commands on the remote machine, towards which
...               ssh connection is opened.
Resource          ${CURDIR}/../all_libs.robot

*** Keywords ***
Reset
    [Arguments]    ${ssh_session}
    [Documentation]    Execute "sudo kubeadm reset" on \${ssh_session}.
    BuiltIn.Log_Many    ${ssh_session}
    BuiltIn.Run_Keyword_And_Return    SshCommons.Switch_And_Execute_Command    ${ssh_session}    sudo kubeadm reset