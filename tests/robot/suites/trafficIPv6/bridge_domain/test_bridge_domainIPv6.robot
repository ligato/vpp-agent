*** Settings ***

Library      OperatingSystem
Library      String

Resource     ../../../variables/${VARIABLES}_variables.robot

Resource     ../../../libraries/all_libs.robot
Resource     ../../../libraries/pretty_keywords.robot

Force Tags        trafficIPv6
Suite Setup       Run Keywords    Discard old results

*** Variables ***
${VARIABLES}=          common
${ENV}=                common

${MAC_LOOP1}=          8a:f1:be:90:00:00
${MAC_LOOP2}=          8a:f1:be:90:00:02
${MAC_MEMIF1}=         02:f1:be:90:00:00
${MAC_MEMIF2}=         02:f1:be:90:00:02
${IP_1}=               fd30::1:b:0:0:1
${IP_2}=               fd30::1:b:0:0:2
${IP_3}=               fd30::1:b:0:0:3
${IP_4}=               fd31::1:b:0:0:1
${IP_5}=               fd31::1:b:0:0:2

${PREFIX}=             64
${SYNC_SLEEP}=         10s
*** Test Cases ***
Create 2 agents in Bridge Domain with Memif interfaces
    [Setup]     Test Setup
    [Teardown]  Test Teardown

    Add Agent VPP Node                 agent_vpp_1
    Add Agent VPP Node                 agent_vpp_2
    Sleep    ${SYNC_SLEEP}
    Create Loopback Interface bvi_loop0 On agent_vpp_1 With Ip ${IP_1}/${PREFIX} And Mac ${MAC_LOOP1}
    Create Master memif0 On agent_vpp_1 With MAC ${MAC_MEMIF1}, Key 1 And m1.sock Socket
    Create Loopback Interface bvi_loop0 On agent_vpp_2 With Ip ${IP_2}/${PREFIX} And Mac ${MAC_LOOP2}
    Create Slave memif0 On agent_vpp_2 With MAC ${MAC_MEMIF2}, Key 1 And m1.sock Socket
    Create Bridge Domain bd1 With Autolearn On agent_vpp_1 With Interfaces bvi_loop0, memif0
    Create Bridge Domain bd1 With Autolearn On agent_vpp_2 With Interfaces bvi_loop0, memif0

    #Check created interfaces
    Sleep    ${SYNC_SLEEP}
    vat_term: Check Loopback Interface State    agent_vpp_1    bvi_loop0    enabled=1     mac=${MAC_LOOP1}   ipv6=${IP_1}/${PREFIX}
    vat_term: Check Loopback Interface State    agent_vpp_2    bvi_loop0    enabled=1     mac=${MAC_LOOP2}   ipv6=${IP_2}/${PREFIX}
    vat_term: Check Memif Interface State     agent_vpp_1  memif0  mac=${MAC_MEMIF1}  role=master  id=1   connected=1  enabled=1  socket=m1.sock
    vat_term: Check Memif Interface State     agent_vpp_2  memif0  mac=${MAC_MEMIF2}  role=slave  id=1   connected=1  enabled=1  socket=m1.sock
    #Check Traffic  VPP1-VPP2
    Ping6 From agent_vpp_1 To ${IP_2}
    Ping6 From agent_vpp_2 To ${IP_1}

    #Create 3. agent and Interfaces
    Add Agent VPP Node                 agent_vpp_3
    Create Master memif1 On agent_vpp_1 With MAC 02:f1:be:90:00:10, Key 2 And m2.sock Socket
    Create Loopback Interface bvi_loop0 On agent_vpp_3 With Ip ${IP_3}/64 And Mac 8a:f1:be:90:00:03
    Create Slave memif0 On agent_vpp_3 With MAC 02:f1:be:90:00:03, Key 2 And m2.sock Socket
    Create Bridge Domain bd1 With Autolearn On agent_vpp_1 With Interfaces bvi_loop0, memif0, memif1
    Create Bridge Domain bd1 With Autolearn On agent_vpp_3 With Interfaces bvi_loop0, memif0

    #Check Traffic  VPP2-VPP3
    Ping6 From agent_vpp_2 To ${IP_3}
    Ping6 From agent_vpp_3 To ${IP_2}

First configure Bridge Domain with Memif interfaces and VXLan then add two agents and try traffic
    [Setup]     Test Setup
    [Teardown]  Test Teardown

    Create Master memif0 On agent_vpp_1 With IP ${IP_1}, MAC ${MAC_MEMIF1}, Key 1 And m0.sock Socket
    Create Slave memif0 On agent_vpp_2 With IP ${IP_2}, MAC ${MAC_MEMIF2}, Key 1 And m0.sock Socket
    Create Loopback Interface bvi_loop0 On agent_vpp_1 With Ip ${IP_4}/64 And Mac ${MAC_LOOP1}
    Create Loopback Interface bvi_loop0 On agent_vpp_2 With Ip ${IP_5}/64 And Mac ${MAC_LOOP2}
    Create VXLan vxlan1 From ${IP_1} To ${IP_2} With Vni 13 On agent_vpp_1
    Create VXLan vxlan1 From ${IP_2} To ${IP_1} With Vni 13 On agent_vpp_2
    Create Bridge Domain bd1 With Autolearn On agent_vpp_1 With Interfaces bvi_loop0, vxlan1
    Create Bridge Domain bd1 With Autolearn On agent_vpp_2 With Interfaces bvi_loop0, vxlan1

    Add Agent VPP Node                 agent_vpp_1
    Add Agent VPP Node                 agent_vpp_2
    Ping6 From agent_vpp_1 To ${IP_2}
    Ping6 From agent_vpp_2 To ${IP_1}
    Ping6 From agent_vpp_1 To ${IP_5}
    Ping6 From agent_vpp_2 To ${IP_4}


Create Bridge Domain without autolearn
    [Setup]     Test Setup
    [Teardown]  Test Teardown

    Add Agent VPP Node                 agent_vpp_1
    Add Agent VPP Node                 agent_vpp_2
    # setup first agent
    Create Loopback Interface bvi_loop0 On agent_vpp_1 With Ip ${IP_1}/64 And Mac ${MAC_LOOP1}
    Create Master memif0 On agent_vpp_1 With MAC ${MAC_MEMIF1}, Key 1 And m1.sock Socket
    Create Bridge Domain bd1 Without Autolearn On agent_vpp_1 With Interfaces bvi_loop0, memif0
    # setup second agent
    Create Loopback Interface bvi_loop0 On agent_vpp_2 With Ip ${IP_2}/64 And Mac ${MAC_LOOP2}
    Create Slave memif0 On agent_vpp_2 With MAC ${MAC_MEMIF2}, Key 1 And m1.sock Socket
    Create Bridge Domain bd1 Without Autolearn On agent_vpp_2 With Interfaces bvi_loop0, memif0
    # without static fib entries ping should fail
    Command: Ping From agent_vpp_1 To ${IP_2} should fail
    Command: Ping From agent_vpp_2 To ${IP_1} should fail
    Add fib entry for ${MAC_LOOP2} in bd1 over memif0 on agent_vpp_1
    Add fib entry for ${MAC_MEMIF2} in bd1 over memif0 on agent_vpp_1
    Add fib entry for ${MAC_LOOP1} in bd1 over memif0 on agent_vpp_2
    Add fib entry for ${MAC_MEMIF1} in bd1 over memif0 on agent_vpp_2
    # and now ping must pass
    Ping6 From agent_vpp_1 To ${IP_2}
    Ping6 From agent_vpp_2 To ${IP_1}

*** Keywords ***