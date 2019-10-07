[Documentation]     Keywords for working with afpacket interfaces using VPP API

*** Settings ***
Resource    ../vpp_api.robot
Resource    ./interface_generic.robot

*** Variables ***
${terminal_timeout}=      30s
${bd_timeout}=            15s

*** Keywords ***

vpp_api: Check Afpacket Interface State
    [Arguments]          ${node}    ${name}    @{desired_state}
    ${internal_name}=    Get Interface Internal Name    ${node}    ${name}
    ${int_state}=        vpp_api: Get Interface State By Name    ${node}    ${internal_name}
    ${ipv4_list}=        vpp_term: Get Interface IPs    ${node}    ${internal_name}
    ${ipv6_list}=        vpp_term: Get Interface IP6 IPs    ${node}    ${internal_name}
    ${config}=           Get VPP Interface Config As Json    ${node}    ${name}
    ${host_int}=         Set Variable    ${config["afpacket"]["host_if_name"]}
    ${enabled}=          Set Variable    ${int_state["admin_up_down"]}
    ${mtu}=              Set Variable    ${int_state["mtu"][0]}
    ${str_mac}=          Set Variable    ${int_state["l2_address"]}
    ${mac}=              Convert str To MAC Address    ${str_mac}
    ${actual_state}=     Create List    enabled=${enabled}    mtu=${mtu}    mac=${mac}
    :FOR    ${ip}    IN    @{ipv4_list}
    \    Append To List    ${actual_state}    ipv4=${ip}
    :FOR    ${ip}    IN    @{ipv6_list}
    \    Append To List    ${actual_state}    ipv6=${ip}
    List Should Contain Sub List    ${actual_state}    ${desired_state}
    [Return]             ${actual_state}
