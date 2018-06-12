*** Settings ***
Resource    KubeEnv.robot
*** Keywords ***
Verify Pod Connectivity - Ping
    [Arguments]    ${source_pod_connection}    ${destination_pod_connection}
    Fail    Not Implemented
#    ${stdout} =    Run Command In Pod    ping -c 5 ${cn_infra_ip}    ssh_session=${vswitch_connection}
#    BuiltIn.Should_Contain   ${stdout}    5 received, 0% packet loss
#    ${stdout} =    Run Command In Pod     ping -c 5 ${vswitch_ip}    ssh_session=${cn_infra_connection}
#    BuiltIn.Should_Contain   ${stdout}    5 received, 0% packet loss
