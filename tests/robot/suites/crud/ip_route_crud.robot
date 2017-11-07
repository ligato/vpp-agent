*** Settings ***

Library     OperatingSystem
Library     String
#Library     RequestsLibrary

Resource     ../../variables/${VARIABLES}_variables.robot
Resource    ../../libraries/all_libs.robot
Resource    ../../libraries/pretty_keywords.robot

Suite Setup       Run Keywords    Discard old results

*** Variables ***
${VARIABLES}=          common
${ENV}=                common

*** Test Cases ***
# CRUD tests for routing
Add Route, Then Delete Route And Again Add Route For Default VRF
    [Setup]      Test Setup
    [Teardown]   Test Teardown

    Given Add Agent VPP Node                 agent_vpp_1
    Then IP Fib On agent_vpp_1 Should Not Contain Route With IP 10.1.1.0/24
    Then Create Route On agent_vpp_1 With IP 10.1.1.0/24 With Next Hop 192.168.1.1 And Vrf Id 0
    Then Show Interfaces On agent_vpp_1
    Then IP Fib On agent_vpp_1 Should Contain Route With IP 10.1.1.0/24
    Then Delete Routes On agent_vpp_1 And Vrf Id 0
    Then IP Fib On agent_vpp_1 Should Not Contain Route With IP 10.1.1.0/24
    Then Create Route On agent_vpp_1 With IP 10.1.1.0/24 With Next Hop 192.168.1.1 And Vrf Id 0

Add Route, Then Delete Route And Again Add Route For Non Default VRF
    [Setup]      Test Setup
    [Teardown]   Test Teardown

    Given Add Agent VPP Node                 agent_vpp_1
    Then IP Fib On agent_vpp_1 Should Not Contain Route With IP 10.1.1.0/24
    Then Create Route On agent_vpp_1 With IP 10.1.1.0/24 With Next Hop 192.168.1.1 And Vrf Id 2
    Then Show Interfaces On agent_vpp_1
    Then IP Fib On agent_vpp_1 Should Contain Route With IP 10.1.1.0/24
    Then IP Fib Table 0 On agent_vpp_1 Should Not Contain Route With IP 10.1.1.0/24
    Then IP Fib Table 2 On agent_vpp_1 Should Contain Route With IP 10.1.1.0/24
    Then Delete Routes On agent_vpp_1 And Vrf Id 2
    Then IP Fib On agent_vpp_1 Should Not Contain Route With IP 10.1.1.0/24
    Then IP Fib Table 2 On agent_vpp_1 Should Not Contain Route With IP 10.1.1.0/24
    Then Create Route On agent_vpp_1 With IP 10.1.1.0/24 With Next Hop 192.168.1.1 And Vrf Id 2
    Then IP Fib On agent_vpp_1 Should Contain Route With IP 10.1.1.0/24
    Then IP Fib Table 0 On agent_vpp_1 Should Not Contain Route With IP 10.1.1.0/24
    Then IP Fib Table 2 On agent_vpp_1 Should Contain Route With IP 10.1.1.0/24

# CRUD tests for VRF
Add VRF Table, Remove VRF table
    [Setup]      Test Setup
    [Teardown]   Test Teardown

    Given Add Agent VPP Node                 agent_vpp_1
    Then IP Fib Table 2 On agent_vpp_1 Should Be Empty
    # Only interface without IP address can be added to VRF table
    Then Create Master memif0 on agent_vpp_1 with MAC 02:f1:be:90:00:00, key 1 and m0.sock socket
    Then Show Interfaces On agent_vpp_1
    Then Create Vrf Table 2 On agent_vpp_1 With Interfaces memif0
    Then IP Fib Table 2 On agent_vpp_1 Should Be Empty
    ################################### Start - This is to replace unfinished functionality in vpp-agent
    # temporarily ip table add 2
    Then vpp_term: Issue Command    node=agent_vpp_1    command=ip table add 2
    # temporarily set int ip table memif0/1 2
    Then vpp_term: Issue Command    node=agent_vpp_1    command=set int ip table memif0/1 2
    ################################### Stop  - This is to replace unfinished functionality in vpp-agent
    # Now interface which was assigned to the VRF table is to be created with IP address
    Then Create Master memif0 on agent_vpp_1 with IP 192.168.1.1, MAC 02:f1:be:90:00:00, key 1 and m0.sock socket
    Then Show Interfaces On agent_vpp_1
    Then IP Fib Table 2 On agent_vpp_1 Should Contain Route With IP 192.168.1.1/32
    Then IP Fib Table 0 On agent_vpp_1 Should Not Contain Route With IP 192.168.1.1/32
    Then Remove Vrf Table 2 On agent_vpp_1
    ################################### Start - This is to replace unfinished functionality in vpp-agent
    # This step is to remove the interface from the VRF table
    Then Create Master memif0 on agent_vpp_1 with MAC 02:f1:be:90:00:00, key 1 and m0.sock socket
    # here we should wait for propagation of the change from etcd to VPP !!!
    Then Wait Until Keyword Succeeds    2 min    5 sec    IP Fib Table 2 On agent_vpp_1 Should Not Contain Route With IP 192.168.1.1/32
    # temporarily set int ip table memif0/1 0
    Then vpp_term: Issue Command    node=agent_vpp_1    command=set int ip table memif0/1 0
    # temporarily ip table del 2
    Then vpp_term: Issue Command    node=agent_vpp_1    command=ip table del 2
    # temporarily: Now interface is to be created with IP address - it will be added to the default VRF
    Then Create Master memif0 on agent_vpp_1 with IP 192.168.1.1, MAC 02:f1:be:90:00:00, key 1 and m0.sock socket
    ################################### Stop - This is to replace unfinished functionality in vpp-agent
    Then IP Fib Table 2 On agent_vpp_1 Should Be Empty
    Then IP Fib Table 0 On agent_vpp_1 Should Contain Route With IP 192.168.1.1/32

*** Keywords ***
Show IP Fib On ${node}
    Log Many    ${node}
    ${out}=     vpp_term: Show IP Fib    ${node}
    Log Many    ${out}

Show Interfaces On ${node}
    ${out}=   vpp_term: Show Interfaces    ${node}
    Log Many  ${out}

IP Fib On ${node} Should Not Contain Route With IP ${ip}/${prefix}
    Log many    ${node}
    ${out}=    vpp_term: Show IP Fib    ${node}
    log many    ${out}
    Should Not Match Regexp    ${out}  ${ip}\\/${prefix}\\s*unicast\\-ip4-chain\\s*\\[\\@0\\]:\\ dpo-load-balance:\\ \\[proto:ip4\\ index:\\d+\\ buckets:\\d+\\ uRPF:\\d+\\ to:\\[0:0\\]\\]

IP Fib On ${node} Should Contain Route With IP ${ip}/${prefix}
    Log many    ${node}
    ${out}=    vpp_term: Show IP Fib    ${node}
    log many    ${out}
    Should Match Regexp        ${out}  ${ip}\\/${prefix}\\s*unicast\\-ip4-chain\\s*\\[\\@0\\]:\\ dpo-load-balance:\\ \\[proto:ip4\\ index:\\d+\\ buckets:\\d+\\ uRPF:\\d+\\ to:\\[0:0\\]\\]

IP Fib Table ${id} On ${node} Should Contain Route With IP ${ip}/${prefix}
    Log many    ${node} ${id}
    ${out}=    vpp_term: Show IP Fib Table    ${node}   ${id}
    log many    ${out}
    Should Match Regexp        ${out}  ${ip}\\/${prefix}\\s*unicast\\-ip4-chain\\s*\\[\\@0\\]:\\ dpo-load-balance:\\ \\[proto:ip4\\ index:\\d+\\ buckets:\\d+\\ uRPF:\\d+\\ to:\\[0:0\\]\\]

IP Fib Table ${id} On ${node} Should Not Contain Route With IP ${ip}/${prefix}
    Log many    ${node} ${id}
    ${out}=    vpp_term: Show IP Fib Table    ${node}   ${id}
    log many    ${out}
    Should Not Match Regexp        ${out}  ${ip}\\/${prefix}\\s*unicast\\-ip4-chain\\s*\\[\\@0\\]:\\ dpo-load-balance:\\ \\[proto:ip4\\ index:\\d+\\ buckets:\\d+\\ uRPF:\\d+\\ to:\\[0:0\\]\\]

IP Fib Table ${id} On ${node} Should Be Empty
    Log many    ${node} ${id}
    ${out}=    vpp_term: Show IP Fib Table    ${node}   ${id}
    log many    ${out}
    Should Be Equal    ${out}   vpp#${SPACE}

IP Fib Table ${id} On ${node} Should Not Be Empty
    Log many    ${node} ${id}
    ${out}=    vpp_term: Show IP Fib Table    ${node}   ${id}
    log many    ${out}
    Should Not Be Equal    ${out}   vpp#${SPACE}
