[Documentation]     Keywords for working with VPP terminal

*** Settings ***
Library      vpp_term.py

*** Variables ***
${interface_timeout}=     15s
${terminal_timeout}=      30s

*** Keywords ***

vpp_term: Check VPP Terminal
    [Arguments]        ${node}
    [Documentation]    Check terminal on node ${node}
    Log Many           ${node}    ${${node}_VPP_HOST_PORT}    ${${node}_VPP_TERM_PROMPT}
    ${command}=        Set Variable       telnet 0 ${${node}_VPP_HOST_PORT}
    ${out}=            Write To Machine   ${node}_term    ${command}
    Should Contain     ${out}             ${${node}_VPP_TERM_PROMPT}
    [Return]           ${out}

vpp_term: Open VPP Terminal
    [Arguments]    ${node}
    [Documentation]    Wait for VPP terminal on node ${node} or timeout
    wait until keyword succeeds  ${terminal_timeout}    5s   vpp_term: Check VPP Terminal    ${node}

vpp_term: Issue Command
    [Arguments]        ${node}     ${command}    ${delay}=${SSH_READ_DELAY}s
    Log Many           ${node}     ${command}    ${delay}    ${node}_term    ${${node}_VPP_TERM_PROMPT}
    ${out}=            Write To Machine Until String    ${node}_term    ${command}    ${${node}_VPP_TERM_PROMPT}    delay=${delay}
    Log                ${out}
#    Should Contain     ${out}             ${${node}_VPP_TERM_PROMPT}
    [Return]           ${out}

vpp_term: Exit VPP Terminal
    [Arguments]        ${node}
    Log Many           ${node}     ${node}_term
    ${ctrl_d}          Evaluate    chr(int(4))
    ${command}=        Set Variable       ${ctrl_d}
    ${out}=            Write To Machine   ${node}_term    ${command}
    [Return]           ${out}

vpp_term: Show Interfaces
    [Arguments]        ${node}    ${interface}=${EMPTY}
    [Documentation]    Show interfaces through vpp terminal
    Log Many           ${node}    ${interface}
    ${out}=            vpp_term: Issue Command  ${node}   sh int ${interface}
    [Return]           ${out}

vpp_term: Show Interfaces Address
    [Arguments]        ${node}    ${interface}=${EMPTY}
    [Documentation]    Show interfaces address through vpp terminal
    Log Many           ${node}    ${interface}
    ${out}=            vpp_term: Issue Command  ${node}   sh int addr ${interface}
    [Return]           ${out}

vpp_term: Show Hardware
    [Arguments]        ${node}    ${interface}=${EMPTY}
    [Documentation]    Show interfaces hardware through vpp terminal
    Log Many           ${node}    ${interface}
    ${out}=            vpp_term: Issue Command  ${node}   sh h ${interface}
    [Return]           ${out}

vpp_term: Show IP Fib
    [Arguments]        ${node}    ${ip}=${EMPTY}
    [Documentation]    Show IP fib output
    Log Many           ${node}    ${ip}
    ${out}=            vpp_term: Issue Command  ${node}    show ip fib ${ip}
    [Return]           ${out}

vpp_term: Show IP Fib Table
    [Arguments]        ${node}    ${id}
    [Documentation]    Show IP fib output for VRF table defined in input
    Log Many           ${node}    ${id}
    ${out}=            vpp_term: Issue Command  ${node}    show ip fib table ${id}
    [Return]           ${out}

vpp_term: Show L2fib
    [Arguments]        ${node}
    Log Many           ${node}
    [Documentation]    Show verbose l2fib output
    ${out}=            vpp_term: Issue Command  ${node}    show l2fib verbose
    [Return]           ${out}

vpp_term: Show Bridge-Domain Detail
    [Arguments]        ${node}    ${id}=1
    Log Many           ${node}
    [Documentation]    Show detail of bridge-domain
    ${out}=            vpp_term: Issue Command  ${node}    show bridge-domain ${id} detail
    [Return]           ${out}

vpp_term: Check Ping
    [Arguments]        ${node}    ${ip}
    Log Many           ${node}    ${ip}
    ${out}=            vpp_term: Issue Command    ${node}    ping ${ip}    delay=10s
    Should Contain     ${out}    from ${ip}
    Should Not Contain    ${out}    100% packet loss

vpp_term: Check Interface Presence
    [Arguments]        ${node}     ${mac}    ${status}=${TRUE}
    [Documentation]    Checking if specified interface with mac exists in VPP
    Log Many           ${node}     ${mac}    ${status}
    ${ints}=           vpp_term: Show Hardware    ${node}
    ${result}=         Run Keyword And Return Status    Should Contain    ${ints}    ${mac}
    Should Be Equal    ${result}    ${status}
    
vpp_term: Interface Is Created
    [Arguments]    ${node}    ${mac}
    Log Many       ${node}    ${mac}
    Wait Until Keyword Succeeds    ${interface_timeout}   3s    vpp_term: Check Interface Presence    ${node}    ${mac}

vpp_term: Interface Is Deleted
    [Arguments]    ${node}    ${mac}
    Log Many       ${node}    ${mac}
    Wait Until Keyword Succeeds    ${interface_timeout}   3s    vpp_term: Check Interface Presence    ${node}    ${mac}    ${FALSE}

vpp_term: Interface Exists
    [Arguments]    ${node}    ${mac}
    Log Many       ${node}    ${mac}
    vpp_term: Check Interface Presence    ${node}    ${mac}
    
vpp_term: Interface Not Exists
    [Arguments]    ${node}    ${mac}
    Log Many       ${node}    ${mac}
    vpp_term: Check Interface Presence    ${node}    ${mac}    ${FALSE}

vpp_term: Check Interface UpDown Status
    [Arguments]          ${node}     ${interface}    ${status}=1
    [Documentation]      Checking up/down state of specified internal interface
    Log Many             ${node}     ${interface}    ${status}
    ${internal_index}=   vat_term: Get Interface Index    agent_vpp_1    ${interface}
    Log                  ${internal_index}
    ${interfaces}=       vat_term: Interfaces Dump    agent_vpp_1
    Log                  ${interfaces}
    ${int_state}=        Get Interface State    ${interfaces}    ${internal_index}
    Log                  ${int_state}
    ${enabled}=          Set Variable    ${int_state["admin_up_down"]}
    Should Be Equal As Integers    ${enabled}    ${status}

vpp_term: Get Interface IPs
    [Arguments]          ${node}     ${interface} 
    Log Many             ${node}     ${interface}
    ${int_addr}=         vpp_term: Show Interfaces Address    ${node}    ${interface}
    Log                  ${int_addr}
    @{ipv4_list}=        Find IPV4 In Text    ${int_addr}
    Log                  ${ipv4_list}
    [Return]             ${ipv4_list}

vpp_term: Get Interface MAC
    [Arguments]          ${node}     ${interface}
    Log Many             ${node}     ${interface}
    ${sh_h}=             vpp_term: Show Hardware    ${node}    ${interface}
    Log                  ${sh_h}
    ${mac}=              Find MAC In Text    ${sh_h}
    Log                  ${mac}
    [Return]             ${mac}

vpp_term: Interface Is Enabled
    [Arguments]          ${node}     ${interface}
    Log Many             ${node}     ${interface}
    Wait Until Keyword Succeeds    ${interface_timeout}   3s    vpp_term: Check Interface UpDown Status    ${node}     ${interface}

vpp_term: Interface Is Disabled
    [Arguments]          ${node}     ${interface}
    Log Many             ${node}     ${interface}
    Wait Until Keyword Succeeds    ${interface_timeout}   3s    vpp_term: Check Interface UpDown Status    ${node}     ${interface}    0

vpp_term: Interface Is Up
    [Arguments]          ${node}     ${interface}
    Log Many             ${node}     ${interface}
    vpp_term: Check Interface UpDown Status    ${node}     ${interface}

vpp_term: Interface Is Down
    [Arguments]          ${node}     ${interface}
    Log Many             ${node}     ${interface}
    vpp_term: Check Interface UpDown Status    ${node}     ${interface}    0

vpp_term: Show Memif
    [Arguments]        ${node}    ${interface}=${EMPTY}
    [Documentation]    Show memif interfaces through vpp terminal
    Log Many           ${node}    ${interface}
    ${out}=            vpp_term: Issue Command  ${node}   sh memif ${interface}
    [Return]           ${out}

