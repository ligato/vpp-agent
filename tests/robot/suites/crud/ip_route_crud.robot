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
