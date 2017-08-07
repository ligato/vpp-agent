*** Settings ***

Library     OperatingSystem
Library     String
#Library     RequestsLibrary

Resource     ../../../variables/${VARIABLES}_variables.robot
Resource    ../../../libraries/all_libs.robot
Resource    ../../../libraries/pretty_keywords.robot

Suite Setup       Run Keywords    Discard old results

*** Variables ***
${VARIABLES}=          common
${ENV}=                common

*** Test Cases ***
Add Route, Then Delete Route And Again Add Route
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

Start Three Agents And Then Configure
    [Setup]         Test Setup
    [Teardown]      Test Teardown
    Add Agent VPP Node    agent_vpp_1
    Add Agent VPP Node    agent_vpp_2
    Add Agent VPP Node    agent_vpp_3
    #setup one side with agent2
    Create loopback interface bvi_loop0 on agent_vpp_1 with ip 10.1.1.1/24 and mac 8a:f1:be:90:00:00
    Create Master memif0 on agent_vpp_1 with MAC 02:f1:be:90:00:00, key 1 and m0.sock socket
    Create bridge domain bd1 With Autolearn on agent_vpp_1 with interfaces bvi_loop0, memif0
    #setup second side with agent3
    Create loopback interface bvi_loop1 on agent_vpp_1 with ip 20.1.1.1/24 and mac 8a:f1:be:90:02:00
    Create Master memif1 on agent_vpp_1 with MAC 02:f1:be:90:02:00, key 2 and m1.sock socket
    Create bridge domain bd2 With Autolearn on agent_vpp_1 with interfaces bvi_loop1, memif1
    # prepare second agent
    Create loopback interface bvi_loop0 on agent_vpp_2 with ip 10.1.1.2/24 and mac 8a:f1:be:90:00:02
    Create Slave memif0 on agent_vpp_2 with MAC 02:f1:be:90:00:02, key 1 and m0.sock socket
    Create bridge domain bd1 With Autolearn on agent_vpp_2 with interfaces bvi_loop0, memif0
    # prepare third agent
    Create loopback interface bvi_loop0 on agent_vpp_3 with ip 20.1.1.2/24 and mac 8a:f1:be:90:00:03
    Create Slave memif0 on agent_vpp_3 with MAC 02:f1:be:90:00:03, key 2 and m1.sock socket
    Create bridge domain bd1 With Autolearn on agent_vpp_3 with interfaces bvi_loop0, memif0
    # setup routes
    Create Route On agent_vpp_2 With IP 20.1.1.0/24 With Next Hop 10.1.1.1 And Vrf Id 0
    Create Route On agent_vpp_3 With IP 10.1.1.0/24 With Next Hop 20.1.1.1 And Vrf Id 0
    # try ping
    Ping From agent_vpp_1 To 10.1.1.2
    Ping From agent_vpp_1 To 20.1.1.2
    Ping From agent_vpp_2 To 20.1.1.2
    Ping From agent_vpp_3 To 10.1.1.2

First Configure Three Agents And Then Start Agents
    [Setup]         Test Setup
    [Teardown]      Test Teardown
    #prepare first agent
    Create loopback interface bvi_loop0 on agent_vpp_1 with ip 10.1.1.1/24 and mac 8a:f1:be:90:00:00
    Create Master memif0 on agent_vpp_1 with MAC 02:f1:be:90:00:00, key 1 and m0.sock socket
    Create loopback interface bvi_loop1 on agent_vpp_1 with ip 20.1.1.1/24 and mac 8a:f1:be:90:02:00
    Create Master memif1 on agent_vpp_1 with MAC 02:f1:be:90:02:00, key 2 and m1.sock socket
    Create bridge domain bd1 With Autolearn on agent_vpp_1 with interfaces bvi_loop0, memif0
    Create bridge domain bd2 With Autolearn on agent_vpp_1 with interfaces bvi_loop1, memif1
    #prepare second agent
    Create loopback interface bvi_loop0 on agent_vpp_2 with ip 10.1.1.2/24 and mac 8a:f1:be:90:00:02
    Create Slave memif0 on agent_vpp_2 with MAC 02:f1:be:90:00:02, key 1 and m0.sock socket
    Create bridge domain bd1 With Autolearn on agent_vpp_2 with interfaces bvi_loop0, memif0
    #prepare third agent
    Create loopback interface bvi_loop0 on agent_vpp_3 with ip 20.1.1.2/24 and mac 8a:f1:be:90:00:03
    Create Slave memif0 on agent_vpp_3 with MAC 02:f1:be:90:00:03, key 2 and m1.sock socket
    Create bridge domain bd1 With Autolearn on agent_vpp_3 with interfaces bvi_loop0, memif0
    #setup routes
    Create Route On agent_vpp_2 With IP 20.1.1.0/24 With Next Hop 10.1.1.1 And Vrf Id 0
    Create Route On agent_vpp_3 With IP 10.1.1.0/24 With Next Hop 20.1.1.1 And Vrf Id 0
    #start agents
    Add Agent VPP Node    agent_vpp_1
    Add Agent VPP Node    agent_vpp_2
    Add Agent VPP Node    agent_vpp_3
    #check ping
    Ping From agent_vpp_1 To 10.1.1.2
    Ping From agent_vpp_1 To 20.1.1.2
    Ping From agent_vpp_2 To 20.1.1.2
    Ping From agent_vpp_3 To 10.1.1.2

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
