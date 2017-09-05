[Documentation]     Keywords for working with VAT terminal

*** Settings ***
Library      vat_term.py

*** Variables ***
${terminal_timeout}=      30s

*** Keywords ***

vat_term: Check VAT Terminal
    [Arguments]        ${node}
    [Documentation]    Check VAT terminal on node ${node}
    Log Many           ${node}     ${node}_vat    ${VAT_START_COMMAND}    ${${node}_VPP_VAT_PROMPT}
    ${out}=            Write To Machine    ${node}_vat    ${DOCKER_COMMAND} exec -it ${node} /bin/bash
    Log                ${out}
    ${command}=        Set Variable        ${VAT_START_COMMAND}
    ${out}=            Write To Machine    ${node}_vat    ${command}
    Log                ${out}
    Should Contain     ${out}              ${${node}_VPP_VAT_PROMPT}
    [Return]           ${out}

vat_term: Open VAT Terminal
    [Arguments]    ${node}
    [Documentation]    Wait for VAT terminal on node ${node} or timeout
    wait until keyword succeeds  ${terminal_timeout}    5s   vat_term: Check VAT Terminal    ${node}


vat_term: Exit VAT Terminal
    [Arguments]        ${node}
    Log Many           ${node}     ${node}_vat
    ${ctrl_c}          Evaluate    chr(int(3))
    ${command}=        Set Variable       ${ctrl_c}
    ${out}=            Write To Machine   ${node}_vat    ${command}
    Log                ${out}
    [Return]           ${out}

vat_term: Issue Command
    [Arguments]        ${node}     ${command}    ${delay}=${SSH_READ_DELAY}s
    Log Many           ${node}     ${command}    ${delay}    ${node}_vat    ${${node}_VPP_VAT_PROMPT}
    ${out}=            Write To Machine Until String    ${node}_vat    ${command}    ${${node}_VPP_VAT_PROMPT}    delay=${delay}
    Log                ${out}
#    Should Contain     ${out}             ${${node}_VPP_VAT_PROMPT}
    [Return]           ${out}

vat_term: Interfaces Dump
    [Arguments]        ${node}
    [Documentation]    Executing command sw_interface_dump
    Log Many           ${node}
    ${out}=            vat_term: Issue Command  ${node}  sw_interface_dump
    [Return]           ${out}

vat_term: IP FIB Dump
    [Arguments]        ${node}
    [Documentation]    Executing command ip_fib_dump
    Log Many           ${node}
    ${out}=            vat_term: Issue Command  ${node}  ip_fib_dump
    [Return]           ${out}

vat_term: VXLan Tunnel Dump
    [Arguments]        ${node}    ${args}=${EMPTY}
    [Documentation]    Executing command vxlan_tunnel_dump
    Log Many           ${node}    ${args}
    ${out}=            vat_term: Issue Command  ${node}  vxlan_tunnel_dump ${args}
    ${out}=            Evaluate    """${out}"""["""${out}""".find('['):"""${out}""".rfind(']')+1]
    Log                ${out}
    [Return]           ${out}

vat_term: Check VXLan Tunnel Presence
    [Arguments]        ${node}     ${src}    ${dst}    ${vni}    ${status}=${TRUE}
    [Documentation]    Checking if specified vxlan tunnel exists
    Log Many           ${node}     ${src}    ${dst}    ${vni}    ${status}
    ${out}=            vat_term: VXLan Tunnel Dump    ${node}
    ${result}  ${if_index}=    Check VXLan Tunnel Presence    ${out}    ${src}    ${dst}    ${vni}
    Should Be Equal    ${result}    ${status}
    [Return]           ${if_index}

vat_term: Get Interface Name
    [Arguments]        ${node}     ${index}
    [Documentation]    Return interface with specified index name
    Log Many           ${node}     ${index}
    ${out}=            vat_term: Interfaces Dump    ${node}
    ${name}=           Get Interface Name    ${out}    ${index}
    [Return]           ${name}

vat_term: Get Interface Index
    [Arguments]        ${node}     ${name}
    [Documentation]    Return interface index with specified name
    Log Many           ${node}     ${name}
    ${out}=            vat_term: Interfaces Dump    ${node}
    ${index}=          Get Interface Index    ${out}    ${name}
    [Return]           ${index}

vat_term: Check VXLan Interface State
    [Arguments]          ${node}    ${name}    @{desired_state}
    Log Many             ${node}    ${name}    ${desired_state}
    ${internal_name}=    vpp_ctl: Get Interface Internal Name    ${node}    ${name}
    Log                  ${internal_name}
    ${internal_index}=   vat_term: Get Interface Index    ${node}    ${internal_name}
    Log                  ${internal_index}
    ${vxlan_data}=       vat_term: VXLan Tunnel Dump    ${node}    sw_if_index ${internal_index}
    ${vxlan_data}=       Evaluate    json.loads('''${vxlan_data}''')    json
    Log                  ${vxlan_data}
    ${interfaces}=       vat_term: Interfaces Dump    ${node}
    Log                  ${interfaces}
    ${int_state}=        Get Interface State    ${interfaces}    ${internal_index}
    Log                  ${int_state}
    ${src}=              Set Variable    ${vxlan_data[0]["src_address"]}
    ${dst}=              Set Variable    ${vxlan_data[0]["dst_address"]}
    ${vni}=              Set Variable    ${vxlan_data[0]["vni"]}
    ${enabled}=          Set Variable    ${int_state["admin_up_down"]}
    ${actual_state}=     Create List    src=${src}    dst=${dst}    vni=${vni}    enabled=${enabled}
    Log List             ${actual_state}
    List Should Contain Sub List    ${actual_state}    ${desired_state}
    [Return]             ${actual_state}

vat_term: Check Afpacket Interface State
    [Arguments]          ${node}    ${name}    @{desired_state}
    Log Many             ${node}    ${name}    ${desired_state}
    ${internal_name}=    vpp_ctl: Get Interface Internal Name    ${node}    ${name}
    Log                  ${internal_name}
    ${internal_index}=   vat_term: Get Interface Index    ${node}    ${internal_name}
    Log                  ${internal_index}
    ${interfaces}=       vat_term: Interfaces Dump    ${node}
    Log                  ${interfaces}
    ${int_state}=        Get Interface State    ${interfaces}    ${internal_index}
    Log                  ${int_state}
    ${ipv4_list}=        vpp_term: Get Interface IPs    ${node}    ${internal_name}
    ${config}=           vpp_ctl: Get VPP Interface Config As Json    ${node}    ${name}
    ${host_int}=         Set Variable    ${config["afpacket"]["host_if_name"]}
    Should Contain       ${internal_name}    ${host_int}
    ${enabled}=          Set Variable    ${int_state["admin_up_down"]}
    ${mtu}=              Set Variable    ${int_state["mtu"]}
    ${dec_mac}=          Set Variable    ${int_state["l2_address"]}
    ${mac}=              Convert Dec MAC To Hex    ${dec_mac}
    ${actual_state}=     Create List    enabled=${enabled}    mtu=${mtu}    mac=${mac}
    :FOR    ${ip}    IN    @{ipv4_list}
    \    Append To List    ${actual_state}    ipv4=${ip}
    Log List             ${actual_state}
    List Should Contain Sub List    ${actual_state}    ${desired_state}
    [Return]             ${actual_state}

vat_term: Check Physical Interface State
    [Arguments]          ${node}    ${name}    @{desired_state}
    Log Many             ${node}    ${name}    ${desired_state}
    ${internal_name}=    vpp_ctl: Get Interface Internal Name    ${node}    ${name}
    Log                  ${internal_name}
    ${internal_index}=   vat_term: Get Interface Index    ${node}    ${internal_name}
    Log                  ${internal_index}
    ${interfaces}=       vat_term: Interfaces Dump    ${node}
    Log                  ${interfaces}
    ${int_state}=        Get Interface State    ${interfaces}    ${internal_index}
    Log                  ${int_state}
    ${ipv4_list}=        vpp_term: Get Interface IPs    ${node}    ${internal_name}
    ${enabled}=          Set Variable    ${int_state["admin_up_down"]}
    ${mtu}=              Set Variable    ${int_state["mtu"]}
    ${dec_mac}=          Set Variable    ${int_state["l2_address"]}
    ${mac}=              Convert Dec MAC To Hex    ${dec_mac}
    ${actual_state}=     Create List    enabled=${enabled}    mtu=${mtu}    mac=${mac}
    :FOR    ${ip}    IN    @{ipv4_list}
    \    Append To List    ${actual_state}    ipv4=${ip}
    Log List             ${actual_state}
    List Should Contain Sub List    ${actual_state}    ${desired_state}
    [Return]             ${actual_state}

vat_term: Check Loopback Interface State
    [Arguments]          ${node}    ${name}    @{desired_state}
    Log Many             ${node}    ${name}    ${desired_state}
    ${internal_name}=    vpp_ctl: Get Interface Internal Name    ${node}    ${name}
    Log                  ${internal_name}
    ${internal_index}=   vat_term: Get Interface Index    ${node}    ${internal_name}
    Log                  ${internal_index}
    ${interfaces}=       vat_term: Interfaces Dump    ${node}
    Log                  ${interfaces}
    ${int_state}=        Get Interface State    ${interfaces}    ${internal_index}
    Log                  ${int_state}
    ${ipv4_list}=        vpp_term: Get Interface IPs    ${node}    ${internal_name}
    ${enabled}=          Set Variable    ${int_state["admin_up_down"]}
    ${mtu}=              Set Variable    ${int_state["mtu"]}
    ${dec_mac}=          Set Variable    ${int_state["l2_address"]}
    ${mac}=              Convert Dec MAC To Hex    ${dec_mac}
    ${actual_state}=     Create List    enabled=${enabled}    mtu=${mtu}    mac=${mac}
    :FOR    ${ip}    IN    @{ipv4_list}
    \    Append To List    ${actual_state}    ipv4=${ip}
    Log List             ${actual_state}
    List Should Contain Sub List    ${actual_state}    ${desired_state}
    [Return]             ${actual_state}

vat_term: Check Memif Interface State
    [Arguments]          ${node}    ${name}    @{desired_state}
    Log Many             ${node}    ${name}    ${desired_state}
    ${internal_name}=    vpp_ctl: Get Interface Internal Name    ${node}    ${name}
    Log                  ${internal_name}
    ${memif_info}=       vpp_term: Show Memif    ${node}    ${internal_name}
    Log                  ${memif_info}
    ${memif_state}=      Parse Memif Info    ${memif_info}
    Log                  ${memif_state}    
    ${ipv4_list}=        vpp_term: Get Interface IPs    ${node}    ${internal_name}
    ${mac}=              vpp_term: Get Interface MAC    ${node}    ${internal_name}
    ${actual_state}=     Create List    mac=${mac}
    :FOR    ${ip}    IN    @{ipv4_list}
    \    Append To List    ${actual_state}    ipv4=${ip}
    Append To List       ${actual_state}    @{memif_state}
    Log List             ${actual_state}
    List Should Contain Sub List    ${actual_state}    ${desired_state}
    [Return]             ${actual_state}

