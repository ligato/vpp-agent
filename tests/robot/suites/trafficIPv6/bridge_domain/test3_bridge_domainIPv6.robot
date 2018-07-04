*** Settings ***

Library      OperatingSystem
Library      String

Resource     ../../../variables/${VARIABLES}_variables.robot

Resource     ../../../libraries/all_libs.robot
Resource     ../../../libraries/pretty_keywords.robot

Force Tags        trafficIPv6
Suite Setup       Testsuite Setup
Suite Teardown    Testsuite Teardown
Test Setup        TestSetup
Test Teardown     TestTeardown

*** Variables ***
${VARIABLES}=          common
${ENV}=                common

${MAC_LOOP1}=          8a:f1:be:90:00:00
${MAC_LOOP2}=          8a:f1:be:90:00:02
${MAC_LOOP3}=          8a:f1:be:90:00:03
${MAC_MEMIF1}=         02:f1:be:90:00:00
${MAC_MEMIF2}=         02:f1:be:90:00:02
${MAC_MEMIF3}=         02:f1:be:90:00:10
${MAC_MEMIF4}=         02:f1:be:90:00:03
${IP_1}=               fd30::1:b:0:0:1
${IP_2}=               fd30::1:b:0:0:2
${IP_3}=               fd30::1:b:0:0:3
${IP_4}=               fd31::1:b:0:0:1
${IP_5}=               fd31::1:b:0:0:2

${PREFIX}=             64
${SYNC_SLEEP}=         15s
*** Test Cases ***
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

TestSetup
    Make Datastore Snapshots    ${TEST_NAME}_test_setup

TestTeardown
    Make Datastore Snapshots    ${TEST_NAME}_test_teardown