*** Settings ***

Library      OperatingSystem
#Library      String

Resource     ../../../variables/${VARIABLES}_variables.robot

Resource     ../../../libraries/all_libs.robot
Resource     ../../../libraries/pretty_keywords.robot

Suite Setup       Testsuite Setup
Suite Teardown    Testsuite Teardown
Test Setup        TestSetup
Test Teardown     TestTeardown

*** Variables ***
${VARIABLES}=          common
${ENV}=                common
${SYNC_SLEEP}=         15s

*** Test Cases ***
Configure Environment 1
    [Tags]    setup
    Add Agent VPP Node    agent_vpp_1
    Add Agent VPP Node    agent_vpp_3
    Sleep    ${SYNC_SLEEP}


Create Infs And BD On VPP1
    Create loopback interface bvi_loop0 on agent_vpp_1 with ip 10.1.1.1/24 and mac 8a:f1:be:90:00:00
    Create Master memif0 on agent_vpp_1 with MAC 02:f1:be:90:02:00, key 2 and m1.sock socket
    Create Bridge Domain bd With Autolearn On agent_vpp_1 with interfaces bvi_loop0, memif0
    Sleep    2s

Create Intfs And BD On VPP3
    Create loopback interface bvi_loop0 on agent_vpp_3 with ip 10.1.1.3/24 and mac 8a:f1:be:90:00:03
    Create Slave memif0 on agent_vpp_3 with MAC 02:f1:be:90:00:03, key 2 and m1.sock socket
    Create Bridge Domain bd With Autolearn On agent_vpp_3 with interfaces bvi_loop0, memif0
    Sleep    2s

#Ping VPP3 x VPP1
#    Ping from agent_vpp_1 to 10.1.1.3
#    Ping from agent_vpp_3 to 10.1.1.1

#Modify Loopback IP on VPP3
#    Create loopback interface bvi_loop0 on agent_vpp_3 with ip 20.1.1.3/24 and mac 8a:f1:be:90:00:03
#    Sleep    6s
#    vpp_term: Show Interfaces    agent_vpp_3
#
#Modify Loopback IP on VPP3 Back
#    Create loopback interface bvi_loop0 on agent_vpp_3 with ip 10.1.1.3/24 and mac 8a:f1:be:90:00:03
#    Sleep    6s
#    vpp_term: Show Interfaces    agent_vpp_3

Modify Loopback IP on VPP1
    Create loopback interface bvi_loop0 on agent_vpp_1 with ip 20.1.1.1/24 and mac 8a:f1:be:90:00:00
    Sleep    6s
    vpp_term: Show Interfaces    agent_vpp_1
    vpp_term: Show Interfaces    agent_vpp_3

Modify Loopback IP on VPP1 Back
    Create loopback interface bvi_loop0 on agent_vpp_1 with ip 10.1.1.1/24 and mac 8a:f1:be:90:00:00
    Sleep    6s
    vpp_term: Show Interfaces    agent_vpp_1
    vpp_term: Show Interfaces    agent_vpp_3

Ping VPP3 x VPP1
    Ping from agent_vpp_1 to 10.1.1.3
    Ping from agent_vpp_3 to 10.1.1.1

Modify Loopback IP on VPP3 Again
    Create loopback interface bvi_loop0 on agent_vpp_3 with ip 20.1.1.3/24 and mac 8a:f1:be:90:00:03
    Sleep    6s
    vpp_term: Show Interfaces    agent_vpp_3

*** Keywords ***

TestSetup
    Make Datastore Snapshots    ${TEST_NAME}_test_setup

TestTeardown
    Make Datastore Snapshots