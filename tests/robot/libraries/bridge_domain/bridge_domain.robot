[Documentation]     Keywords for working with bridge domains using VPP API

*** Settings ***

Library    bridge_utils.py
Library    Collections

*** Variables ***

*** Keywords ***


vpp_api: Bridge Domain Dump
    [Arguments]        ${node}
    [Documentation]    Executing command bridge_domain_dump
    ${out}=            Bridge Domain Dump    ${DOCKER_HOST_IP}    ${DOCKER_HOST_USER}    ${DOCKER_HOST_PSWD}    ${node}
    [Return]           ${out}
    
vpp_api: Check Bridge Domain State
    [Arguments]          ${node}    ${bd}    &{desired_state}
    ${bd_id}=            Get Bridge Domain ID    ${node}    ${bd}
    ${bd_dump}=          vpp_api: Bridge Domain Dump    ${node}
    ${data}=             Filter Bridge Domain Dump By ID   ${bd_dump}    ${bd_id}
    ${flood}=                              Set Variable    ${data["flood"]}
    ${forward}=                            Set Variable    ${data["forward"]}
    ${learn}=                              Set Variable    ${data["learn"]}
    ${arp_termination}=                    Set Variable    ${data["arp_term"]}
    ${unknown_unicast_flood}=              Set Variable    ${data["uu_flood"]}
    ${bridged_virtual_interface}=          Set Variable If    ${data["bvi_sw_if_index"]} == 4294967295
    ...    none    ${data["bvi_sw_if_index"]}
    ${actual_state}=     Create Dictionary
    ...    flood=${flood}    forward=${forward}    learn=${learn}    arp_term=${arp_termination}
    ...    unicast=${unknown_unicast_flood}    bvi_int=${bridged_virtual_interface}
    ${actual_state_interfaces}=    Create List
    :FOR    ${interface}    IN    @{data["sw_if_details"]}
    \    ${internal_name}=    vpp_api: Get Interface Name    ${node}    ${interface["sw_if_index"]}
    \    Append To List    ${actual_state_interfaces}    ${internal_name}
    Set To Dictionary    ${actual_state}    interfaces    ${actual_state_interfaces}
    ${internal_names}=    Create List
    :FOR    ${interface}    IN    @{desired_state["interfaces"]}
    \    ${internal_name}=    Get Interface Internal Name    ${node}    ${interface}
    \    Append To List    ${internal_names}    ${internal_name}
    Set To Dictionary    ${desired_state}    interfaces    ${internal_names}
    List Should Contain Sub List    ${actual_state}    ${desired_state}

vpp_api: Check Bridge Domain Presence
    [Arguments]          ${node}    ${bd_name}
    ${bd_id}=    Get Bridge Domain ID    ${node}    ${bd_name}
    ${output}=    vpp_api: Bridge Domain Dump    ${node}
    ${data}=     Filter Bridge Domain Dump By ID   ${output}    ${bd_id}

vpp_api: BD Is Created
    [Arguments]    ${node}    ${bd_name}
    Wait Until Keyword Succeeds    ${bd_timeout}   3s    vpp_api: Check Bridge Domain Presence    ${node}    ${bd_name}

vpp_api: BD Is Deleted
    [Arguments]    ${node}    ${bd_name}
    Run Keyword And Expect Error    *Bridge domain not found by id 0.*
    ...    vpp_api: BD Is Created    ${node}    ${bd_name}

vpp_api: BD Not Exists
    [Arguments]    ${node}    ${bd_name}
    ${bd_id}=    Get Bridge Domain ID    ${node}    ${bd_name}
    Should Be Equal    ${bd_id}    0

vpp_api: No Bridge Domains Exist
    [Arguments]    ${node}
    ${data}=    vpp_api: Bridge Domain Dump    ${node}
    Should Be Empty    ${data}

vpp_api: Check BD Presence
    [Arguments]        ${node}     ${interfaces}    ${status}=${TRUE}
    ${indexes}=    Create List
    :FOR    ${int}    IN    @{interfaces}
    \    ${sw_if_index}=    Get Interface Sw If Index    ${node}    ${int}
    \    Append To List    ${indexes}    ${sw_if_index}
    ${bd_dump}=        vpp_api: Bridge Domain Dump    ${node}
    ${result}=         Check BD Presence    ${bd_dump}    ${indexes}
    Should Be Equal    ${result}    ${status}