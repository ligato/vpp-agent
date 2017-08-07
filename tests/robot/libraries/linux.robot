*** Settings ***
Library        linux.py

*** Variables ***
${interface_timeout}=     15s

*** Keywords ***
linux: Get Linux Interfaces
    [Arguments]        ${node}
    Log                ${node}
    ${out}=    Execute In Container    ${node}    ip a
    Log    ${out}
    ${ints}=    Parse Linux Interfaces    ${out}
    Log    ${ints}
    [Return]    ${ints}

linux: Check Veth Interface State
    [Arguments]          ${node}    ${name}    @{desired_state}
    Log Many             ${node}    ${name}    ${desired_state}
    ${veth_config}=      vpp_ctl: Get Linux Interface Config As Json    ${node}    ${name}
    ${peer}=             Set Variable    ${veth_config["veth"]["peer_if_name"]}
    Log                  ${peer}
    ${ints}=             linux: Get Linux Interfaces    ${node}
    ${actual_state}=     Pick Linux Interface    ${ints}    ${name}\@${peer}
    Log List             ${actual_state}
    List Should Contain Sub List    ${actual_state}    ${desired_state}
    [Return]             ${actual_state}

linux: Check Interface Presence
    [Arguments]        ${node}     ${mac}    ${status}=${TRUE}
    [Documentation]    Checking if specified interface with mac exists in linux
    Log Many           ${node}     ${mac}    ${status}
    ${ints}=           linux: Get Linux Interfaces    ${node}
    ${result}=         Check Linux Interface Presence    ${ints}    ${mac}
    Should Be Equal    ${result}    ${status}

linux: Interface Is Created
    [Arguments]    ${node}    ${mac}                    
    Log Many       ${node}    ${mac} 
    Wait Until Keyword Succeeds    ${interface_timeout}   3s    linux: Check Interface Presence    ${node}    ${mac}

linux: Interface Is Deleted
    [Arguments]    ${node}    ${mac}                    
    Log Many       ${node}    ${mac} 
    Wait Until Keyword Succeeds    ${interface_timeout}   3s    linux: Check Interface Presence    ${node}    ${mac}    ${FALSE}

linux: Interface Exists
    [Arguments]    ${node}    ${mac}
    Log Many       ${node}    ${mac}
    linux: Check Interface Presence    ${node}    ${mac}

linux: Interface Not Exists
    [Arguments]    ${node}    ${mac}
    Log Many       ${node}    ${mac}
    linux: Check Interface Presence    ${node}    ${mac}    ${FALSE}

linux: Check Ping
    [Arguments]        ${node}    ${ip}
    Log Many           ${node}    ${ip}
    ${out}=            Execute In Container    ${node}    ping -c 5 ${ip}
    Should Contain     ${out}    from ${ip}
    Should Not Contain    ${out}    100% packet loss

