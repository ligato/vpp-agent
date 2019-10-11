[Documentation]     Keywords for working with vxlan interfaces using VPP API and cli terminal

*** Settings ***
Library     vxlan_utils.py

Resource    ../vpp_api.robot
Resource    interface_generic.robot

*** Variables ***
${terminal_timeout}=      30s
${bd_timeout}=            15s
${tunnel_timeout}=        15s

*** Keywords ***

vpp_api: Check VXLan Interface State
    [Arguments]          ${node}    ${name}    @{desired_state}
    ${internal_name}=    Get Interface Internal Name    ${node}    ${name}
    ${internal_index}=   vpp_api: Get Interface Index    ${node}    ${internal_name}
    ${vxlan_data_list}=       vpp_api: VXLan Tunnel Dump    ${node}
    ${vxlan_data}=         Filter VXLan Tunnel Dump By Index    ${vxlan_data_list}    ${internal_index}
    ${interfaces}=       vpp_api: Interfaces Dump    ${node}
    ${int_state}=        vpp_api: Get Interface State By Index    ${node}    ${internal_index}
    ${src}=              Set Variable    ${vxlan_data["src_address"]}
    ${dst}=              Set Variable    ${vxlan_data["dst_address"]}
    ${vni}=              Set Variable    ${vxlan_data["vni"]}
    ${enabled}=          Set Variable    ${int_state["admin_up_down"]}
    ${actual_state}=     Create List    src=${src}    dst=${dst}    vni=${vni}    enabled=${enabled}
    List Should Contain Sub List    ${actual_state}    ${desired_state}
    [Return]             ${actual_state}

vpp_api: VXLan Tunnel Dump
    [Arguments]        ${node}
    [Documentation]    Executing command vxlan_tunnel_dump
    ${out}=            VXLan Tunnel Dump    ${DOCKER_HOST_IP}    ${DOCKER_HOST_USER}    ${DOCKER_HOST_PSWD}    ${node}
    [Return]           ${out}

vpp_api: Check VXLan Tunnel Presence
    [Arguments]        ${node}     ${src}    ${dst}    ${vni}    ${status}=${TRUE}
    [Documentation]    Checking if specified vxlan tunnel exists
    ${out}=            vpp_api: VXLan Tunnel Dump    ${node}
    ${result}  ${if_index}=    Check VXLan Tunnel Presence From API    ${out}    ${src}    ${dst}    ${vni}
    Should Be Equal    ${result}    ${status}
    [Return]           ${if_index}

VXLan Tunnel Is Created
    [Arguments]    ${node}    ${src}    ${dst}    ${vni}
    ${int_index}=  Wait Until Keyword Succeeds    ${tunnel_timeout}   3s
    ...    vpp_api: Check VXLan Tunnel Presence    ${node}    ${src}    ${dst}    ${vni}
    [Return]       ${int_index}

VXLan Tunnel Is Deleted
    [Arguments]    ${node}    ${src}    ${dst}    ${vni}
    Wait Until Keyword Succeeds    ${tunnel_timeout}   3s    vpp_api: Check VXLan Tunnel Presence    ${node}    ${src}    ${dst}    ${vni}    ${NONE}

VXLan Tunnel Exists
    [Arguments]    ${node}    ${src}    ${dst}    ${vni}
    ${int_index}=  vat_term: Check VXLan Tunnel Presence    ${node}    ${src}    ${dst}    ${vni}
    [Return]       ${int_index}

VXLan Tunnel Not Exists
    [Arguments]    ${node}    ${src}    ${dst}    ${vni}
    vpp_api: Check VXLan Tunnel Presence    ${node}    ${src}    ${dst}    ${vni}    ${NONE}

