*** Settings ***

Library      OperatingSystem
#Library      String

Resource     ../../../variables/${VARIABLES}_variables.robot

Resource     ../../../libraries/all_libs.robot
Resource     ../../../libraries/pretty_keywords.robot

Suite Setup       Testsuite Setup
Suite Teardown    Testsuite Teardown

*** Variables ***
${VARIABLES}=          common
${ENV}=                common
${SYNC_SLEEP}=         5s

*** Test Cases ***
Run Two Agents With L2 Memif's And Loopback Interfaces In bd1
    Add Agent VPP Node    agent_vpp_1
    Add Agent VPP Node    agent_vpp_2
    Create loopback interface bvi_loop0 on agent_vpp_1 with ip 10.1.1.1/24 and mac 8a:f1:be:90:00:00
    Create loopback interface bvi_loop0 on agent_vpp_2 with ip 10.1.1.2/24 and mac 8a:f1:be:90:00:02
    Create Master memif0 on agent_vpp_1 with MAC 02:f1:be:90:00:00, key 1 and m0.sock socket
    Create Slave memif0 on agent_vpp_2 with MAC 02:f1:be:90:00:02, key 1 and m0.sock socket
    Create Bridge Domain bd1 With Autolearn On agent_vpp_1 with interfaces bvi_loop0, memif0
    Create Bridge Domain bd1 With Autolearn On agent_vpp_2 with interfaces bvi_loop0, memif0
    Sleep    ${SYNC_SLEEP}
    Ping from agent_vpp_1 to 10.1.1.2
    Ping from agent_vpp_2 to 10.1.1.1

Update bd1 On vpp1
    Create Master memif1 on agent_vpp_1 with MAC 02:f1:be:90:02:00, key 2 and m1.sock socket
    Create Bridge Domain bd1 With Autolearn On agent_vpp_1 with interfaces bvi_loop0, memif0, memif1

Add vpp3 To bd1
    Add Agent VPP Node    agent_vpp_3
    Sleep    ${SYNC_SLEEP}
    Create loopback interface bvi_loop0 on agent_vpp_3 with ip 10.1.1.3/24 and mac 8a:f1:be:90:00:03
    vpp_term: Show Interfaces    agent_vpp_3
    Create Slave memif0 on agent_vpp_3 with MAC 02:f1:be:90:00:03, key 2 and m1.sock socket
    vpp_term: Show Interfaces    agent_vpp_3
    Create Bridge Domain bd1 With Autolearn On agent_vpp_3 with interfaces bvi_loop0, memif0
    vpp_term: Show Interfaces    agent_vpp_3
    Ping from agent_vpp_1 to 10.1.1.3
    Ping from agent_vpp_2 to 10.1.1.3

Add bd2 On vpp1 And vpp4
    Create loopback interface bvi_loop1 on agent_vpp_1 with ip 20.1.1.1/24 and mac 8a:f1:be:90:00:04
    Create Master memif2 on agent_vpp_1 with MAC 02:f1:be:90:01:00, key 3 and m2.sock socket
    vpp_term: Show Interfaces    agent_vpp_3
    Create Bridge Domain bd2 With Autolearn On agent_vpp_1 with interfaces bvi_loop1, memif2
    Add Agent VPP Node    agent_vpp_4
    vpp_term: Show Interfaces    agent_vpp_3
    Create loopback interface bvi_loop0 on agent_vpp_4 with ip 20.1.1.2/24 and mac 8a:f1:be:90:00:05
    Create Slave memif0 on agent_vpp_4 with MAC 02:f1:be:90:01:05, key 3 and m2.sock socket
    vpp_term: Show Interfaces    agent_vpp_3
    Create Bridge Domain bd2 With Autolearn On agent_vpp_4 with interfaces bvi_loop0, memif0
    vpp_term: Show Interfaces    agent_vpp_3
    Ping From agent_vpp_1 to 20.1.1.2
    Ping From agent_vpp_4 to 20.1.1.1

Move agent3 From bd1 To bd2
    Create loopback interface bvi_loop0 on agent_vpp_3 with ip 20.1.1.3/24 and mac 8a:f1:be:90:00:03
    vpp_term: Show Interfaces    agent_vpp_3
    Create Bridge Domain bd1 With Autolearn On agent_vpp_1 with interfaces bvi_loop0, memif0
    Create Bridge Domain bd2 With Autolearn On agent_vpp_1 with interfaces bvi_loop1, memif1, memif2
    vpp_term: Show Interfaces    agent_vpp_3
    Ping From agent_vpp_3 To 20.1.1.2
    Ping From agent_vpp_4 To 20.1.1.3

Move agent3 back from bd2 to bd1
    Create loopback interface bvi_loop0 on agent_vpp_3 with ip 10.1.1.3/24 and mac 8a:f1:be:90:00:03
    vpp_term: Show Interfaces    agent_vpp_3
    Create Bridge Domain bd2 With Autolearn On agent_vpp_1 with interfaces bvi_loop1, memif2
    Create Bridge Domain bd1 With Autolearn On agent_vpp_1 with interfaces bvi_loop0, memif0, memif1
    vpp_term: Show Interfaces    agent_vpp_3
    Ping from agent_vpp_1 to 10.1.1.3
    Ping from agent_vpp_2 to 10.1.1.3

Add Static Routes for both subnets
    Create Route On agent_vpp_2 With IP 20.1.1.0/24 With Next Hop 10.1.1.1 and vrf id 0
    Create Route On agent_vpp_3 With IP 20.1.1.0/24 With Next Hop 10.1.1.1 and vrf id 0
    Create Route On agent_vpp_4 With IP 10.1.1.0/24 With Next Hop 20.1.1.1 and vrf id 0
    vpp_term: Show Interfaces    agent_vpp_3
    Ping from agent_vpp_2 to 20.1.1.2
    Ping from agent_vpp_3 to 20.1.1.2
    Ping from agent_vpp_4 to 10.1.1.2
    Ping from agent_vpp_4 to 10.1.1.3

Create VXlan tunnel between agent2 and agent3
    Create Loopback Interface bvi_loop1 On agent_vpp_2 With Ip 30.1.1.2/24 And Mac 8a:f1:be:90:00:06
    Create Loopback Interface bvi_loop1 On agent_vpp_3 With Ip 30.1.1.3/24 And Mac 8a:f1:be:90:00:07
    Create VXLan vxlan1 from 10.1.1.2 to 10.1.1.3 with vni 13 on agent_vpp_2
    Create VXLan vxlan1 from 10.1.1.3 to 10.1.1.2 with vni 13 on agent_vpp_3
    Create Bridge Domain bd2 With Autolearn On agent_vpp_2 With Interfaces bvi_loop1, vxlan1
    Create Bridge Domain bd2 With Autolearn On agent_vpp_3 With Interfaces bvi_loop1, vxlan1
    Sleep    ${SYNC_SLEEP}
    Ping from agent_vpp_2 to 30.1.1.3
    Ping from agent_vpp_3 to 30.1.1.2

Remove vxlan and ping should not pass
    Remove Bridge Domain bd2 On agent_vpp_2
    Remove Bridge Domain bd2 On agent_vpp_3
    Remove Interface vxlan1 On agent_vpp_2
    Remove Interface vxlan1 On agent_vpp_3
    Remove Interface bvi_loop1 On agent_vpp_2
    Remove Interface bvi_loop1 On agent_vpp_3
    Sleep    ${SYNC_SLEEP}
    Command: Ping From agent_vpp_2 To 30.1.1.3 should fail

Remove agent3
    Create Bridge Domain bd1 With Autolearn On agent_vpp_1 With Interfaces bvi_loop0, memif0
    Remove Interface memif0 On agent_vpp_3
    Remove Interface memif1 On agent_vpp_1
    Remove Node        agent_vpp_3
    Sleep    ${SYNC_SLEEP}
    Command: Ping From agent_vpp_2 To 10.1.1.3 should fail

*** Keywords ***
TestSetup
    Make Datastore Snapshots    ${TEST_NAME}_test_setup

TestTeardown
    Make Datastore Snapshots    ${TEST_NAME}_test_teardown