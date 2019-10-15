*** Settings ***
Library        linux.py

*** Variables ***
${interface_timeout}=     15s
${PINGSERVER_UDP}=        nc -uklp
${PINGSERVER_TCP}=        nc -klp
${UDPPING}=               nc -uzv
${TCPPING}=               nc -zv


*** Keywords ***
linux: Get Linux Interfaces
    [Arguments]        ${node}
    ${out}=    Execute In Container    ${node}    ip a
    ${ints}=    Parse Linux Interfaces    ${out}
    [Return]    ${ints}

linux: Check Veth Interface State
    [Arguments]          ${node}    ${name}    @{desired_state}
    ${veth_config}=      Get Linux Interface Config As Json    ${node}    ${name}
    ${peer}=             Set Variable    ${veth_config["veth"]["peer_if_name"]}
    ${ints}=             linux: Get Linux Interfaces    ${node}
    ${actual_state}=     Pick Linux Interface    ${ints}    ${name}\@${peer}
    List Should Contain Sub List    ${actual_state}    ${desired_state}
    [Return]             ${actual_state}

linux: Check Interface Is Present
    [Arguments]        ${node}     ${mac}
    [Documentation]    Checking if specified interface with mac exists in linux
    ${ints}=           linux: Get Linux Interfaces    ${node}
    ${result}=         Check Linux Interface Presence    ${ints}    ${mac}
    Should Be Equal    ${result}    ${TRUE}    values=False    msg=Interface with MAC ${mac} is not present in Linux.

linux: Check Interface Is Not Present
    [Arguments]        ${node}     ${mac}
    [Documentation]    Checking if specified interface with mac exists in linux
    ${ints}=           linux: Get Linux Interfaces    ${node}
    ${result}=         Check Linux Interface Presence    ${ints}    ${mac}
    Should Be Equal    ${result}    ${FALSE}    values=False    msg=Interface with MAC ${mac} is present in Linux but shouldn't.

linux: Check Interface With IP Presence
    [Arguments]        ${node}     ${mac}    ${ip}      ${status}=${TRUE}
    [Documentation]    Checking if specified interface with mac and ip exists in linux
    ${ints}=           linux: Get Linux Interfaces    ${node}
    ${result}=         Check Linux Interface IP Presence    ${ints}    ${mac}   ${ip}
    Should Be Equal    ${result}    ${status}

linux: Interface Is Created
    [Arguments]    ${node}    ${mac}                    
    Wait Until Keyword Succeeds    ${interface_timeout}   3s    linux: Check Interface Is Present    ${node}    ${mac}

linux: Interface With IP Is Created
    [Arguments]    ${node}    ${mac}    ${ipv4}
    Wait Until Keyword Succeeds    ${interface_timeout}   3s    linux: Check Interface With IP Presence    ${node}    ${mac}    ${ipv4}

linux: Interface Is Deleted
    [Arguments]    ${node}    ${mac}                    
    Wait Until Keyword Succeeds    ${interface_timeout}   3s    linux: Check Interface Is Not Present    ${node}    ${mac}

linux: Interface With IP Is Deleted
    [Arguments]    ${node}    ${mac}   ${ipv4}
    Wait Until Keyword Succeeds    ${interface_timeout}   3s    linux: Check Interface With IP Presence    ${node}    ${mac}    ${ipv4}   ${FALSE}

linux: Interface Exists
    [Arguments]    ${node}    ${mac}
    linux: Check Interface Is Present    ${node}    ${mac}

linux: Interface Not Exists
    [Arguments]    ${node}    ${mac}
    linux: Check Interface Is Not Present    ${node}    ${mac}

linux: Check Ping
    [Arguments]        ${node}    ${ip}    ${count}=5
    ${out}=            Execute In Container    ${node}    ping -c ${count} ${ip}
    Should Contain     ${out}    from ${ip}
    Should Not Contain    ${out}    100% packet loss

linux: Check Ping6
    [Arguments]        ${node}    ${ip}    ${count}=5
    ${out}=            Execute In Container    ${node}    ping6 -c ${count} ${ip}
    Should Contain     ${out}    from ${ip}    ignore_case=True
    Should Not Contain    ${out}    100% packet loss

linux: Install Executable Script
    [Arguments]        ${node}    ${scriptContent}    ${fileName}    ${fileDirectory}=/usr/bin
    ${fullFilePath}=    Catenate    SEPARATOR=    ${fileDirectory}    /    ${fileName}
    ${scriptContent}=    Replace String    ${scriptContent}    "    \\"           # preparing " character for echoing from sh script
    ${scriptContent}=    Replace String    ${scriptContent}    '    \'\"\'\"\'    # preparing ' character for echoing from sh script
    ${scriptContent}=    Replace String    ${scriptContent}    \n    \\n          # preparing linux EOL for echoing from sh script
    ${scriptContent}=    Replace String    ${scriptContent}    \r    \\r          # preparing windows EOL (\r\n) for echoing from sh script
    Execute In Container    ${node}    sh -c 'echo "${scriptContent}" > ${fullFilePath}'
    Execute In Container    ${node}    chmod a+x ${fullFilePath}

linux: Send Ethernet Frame
    [Arguments]        ${node}    ${out_interface}    ${source_address}    ${destination_address}    ${ethernet_type}    ${payload}    ${checksum}
    ${script}=    OperatingSystem.Get File    ${CURDIR}/../../robot/resources/sendEthernetFrame.py
    ${script}=    replace variables           ${script}
    linux: Install Executable Script    ${node}    ${script}    sendEthernetFrame.py
    Execute In Container    ${node}    sendEthernetFrame.py

linux: Run TCP Ping Server On Node
    [Arguments]    ${node}   ${port}
    [Documentation]    Run TCP PingServer as listener on node ${node}
    ${out}=            Execute In Container Background    ${node}    ${PINGSERVER_TCP} ${port}

linux: Run UDP Ping Server On Node
    [Arguments]    ${node}   ${port}
    [Documentation]    Run UDP PingServer as listener on node ${node}
    ${out}=            Execute In Container Background    ${node}    ${PINGSERVER_UDP} ${port}

linux: TCPPing
    [Arguments]        ${node}    ${ip}     ${port}
    #${out}=            Execute In Container    ${node}    ${TCPPING} ${ip} ${port}
    #${out}=            Write To Container Until Prompt   ${node}     ${TCPPING} ${ip} ${port}
    ${out}=            Write Command to Container   ${node}     ${TCPPING} ${ip} ${port}
    Should Contain     ${out}    Connection to ${ip} ${port} port [tcp/*] succeeded!
    Should Not Contain    ${out}    Connection refused

linux: TCPPingNot
    [Arguments]        ${node}    ${ip}     ${port}
    #${out}=            Execute In Container    ${node}    ${TCPPING} ${ip} ${port}
    #${out}=            Write To Container Until Prompt   ${node}     ${TCPPING} ${ip} ${port}
    ${out}=            Write Command to Container   ${node}     ${TCPPING} ${ip} ${port}
    Should Not Contain     ${out}    Connection to ${ip} ${port} port [tcp/*] succeeded!
    Should Contain    ${out}    Connection refused

linux: UDPPing
    [Arguments]        ${node}    ${ip}     ${port}
    #${out}=            Execute In Container    ${node}    ${UDPPING} ${ip} ${port}
    #${out}=            Write To Container Until Prompt    ${node}    ${UDPPING} ${ip} ${port}
    ${out}=            Write Command to Container    ${node}    ${UDPPING} ${ip} ${port}
    Should Contain     ${out}    Connection to ${ip} ${port} port [udp/*] succeeded!
    Should Not Contain    ${out}    Connection refused

linux: UDPPingNot
    [Arguments]        ${node}    ${ip}     ${port}
    #${out}=            Execute In Container    ${node}    ${UDPPING} ${ip} ${port}
    #${out}=            Write To Container Until Prompt    ${node}    ${UDPPING} ${ip} ${port}
    ${out}=            Write Command to Container    ${node}    ${UDPPING} ${ip} ${port}
    Should Not Contain     ${out}    Connection to ${ip} ${port} port [udp/*] succeeded!
    Should Contain    ${out}    Connection refused

linux: Check Processes on Node
    [Arguments]        ${node}
    ${out}=            Execute In Container    ${node}    ps aux

linux: Set Host TAP Interface
    [Arguments]    ${node}    ${host_if_name}    ${ip}    ${prefix}    ${mac}=    ${second_ip}=    ${second_prefix}=
    ${out}=    Execute In Container    ${node}    ip link set dev ${host_if_name} up
    ${out}=    Execute In Container    ${node}    ip addr add ${ip}/${prefix} dev ${host_if_name}
    Run Keyword If    "${second_ip}" != ""    Execute In Container    ${node}    ip addr add ${second_ip}/${second_prefix} dev ${host_if_name}
    Run Keyword If    "${mac}" != ""          Execute In Container    ${node}    ip link set ${host_if_name} address ${mac}

linux: Add Route
    [Arguments]    ${node}    ${destination_ip}    ${prefix}    ${next_hop_ip}
    Execute In Container    ${node}    ip route add ${destination_ip}/${prefix} via ${next_hop_ip}

linux: Delete Route
    [Arguments]    ${node}    ${destination_ip}    ${prefix}    ${next_hop_ip}
    Execute In Container    ${node}    ip route del ${destination_ip}/${prefix} via ${next_hop_ip}

linux: Check ARP
    [Arguments]        ${node}      ${interface}    ${ipv4}     ${MAC}    ${presence}
    [Documentation]    Check ARP presence in linux
    ${out}=            Execute In Container    ${node}    cat /proc/net/arp
    ${arps}=           Parse Linux ARP Entries    ${out}
    ${wanted}=         Create Dictionary    interface=${interface}    ip_addr=${ipv4}    mac_addr=${MAC}
    Run Keyword If     "${presence}" == "True"
    ...    Should Contain     ${arps}    ${wanted}
    ...    ELSE    Should Not Contain    ${arps}    ${wanted}

linux: Check IPv6 Neighbor
    [Arguments]        ${node}      ${interface}    ${ip_address}    ${mac_address}    ${presence}
    [Documentation]    Check IPv6 Neighbor presence in linux
    ${out}=            Execute In Container    ${node}    ip -6 neighbour
    ${arps}=           Parse Linux IPv6 Neighbor Entries    ${out}
    ${wanted}=         Create Dictionary    interface=${interface}    ip_addr=${ip_address}    mac_addr=${mac_address}
    Run Keyword If     "${presence}" == "True"
    ...    Should Contain     ${arps}    ${wanted}
    ...    ELSE    Should Not Contain    ${arps}    ${wanted}
