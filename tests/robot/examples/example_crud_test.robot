*** Settings ***
Library      OperatingSystem
Library      Collections

Resource     ../variables/${VARIABLES}_variables.robot

Resource     ../libraries/all_libs.robot

Force Tags        crud     IPv4
Suite Setup       Testsuite Setup
Suite Teardown    Testsuite Teardown
Test Setup        TestSetup
Test Teardown     TestTeardown

*** Variables ***
${VARIABLES}=        common
${ENV}=              common
${WAIT_TIMEOUT}=     20s
${SYNC_SLEEP}=       3s

${NAME_TAP1}=        vpp1_tap1
${NAME_TAP2}=        vpp1_tap2
${MAC_TAP1}=         12:21:21:11:11:11
${MAC_TAP1_2}=       22:21:21:11:11:11
${MAC_TAP2}=         22:21:21:22:22:22
${IP_TAP1}=          20.20.1.1
${IP_TAP1_2}=        21.20.1.2
${IP_TAP2}=          20.20.2.1
${PREFIX}=           24
${MTU}=              4800
${UP_STATE}=         up


*** Test Cases ***
Configure Environment
    [Tags]    setup
    Configure Environment 1

Show Something Before Setup
    ${interfaces}=    vpp_term: Show Interfaces    agent_vpp_1

Add Something
    vpp_term: Interface Not Exists  node=agent_vpp_1    mac=${MAC_TAP1}
    Put TAP Interface With IP    node=agent_vpp_1    name=${NAME_TAP1}    mac=${MAC_TAP1}    ip=${IP_TAP1}    prefix=${PREFIX}    host_if_name=linux_${NAME_TAP1}

Check Something Is Created
    ${interfaces}=       vat_term: Interfaces Dump    node=agent_vpp_1
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    vpp_term: Interface Is Created    node=agent_vpp_1    mac=${MAC_TAP1}
    ${actual_state}=    vpp_term: Check TAP interface State    agent_vpp_1    ${NAME_TAP1}    mac=${MAC_TAP1}    ipv4=${IP_TAP1}/${PREFIX}    state=${UP_STATE}

Add Something_Other
    No Operation

Check Something_Other Is Created
    No Operation

Check Something Is Still Configured
    ${actual_state}=    vpp_term: Check TAP interface State    agent_vpp_1    ${NAME_TAP1}    mac=${MAC_TAP1}    ipv4=${IP_TAP1}/${PREFIX}    state=${UP_STATE}

Update Something
    Put TAP Interface With IP    node=agent_vpp_1    name=${NAME_TAP1}    mac=${MAC_TAP1_2}    ip=${IP_TAP1_2}    prefix=${PREFIX}    host_if_name=linux_${NAME_TAP1}

Check Something Is Created
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    vpp_term: Interface Is Created    node=agent_vpp_1    mac=${MAC_TAP1_2}
    ${actual_state}=    vpp_term: Check TAP interface State    agent_vpp_1    ${NAME_TAP1}    mac=${MAC_TAP1_2}    ipv4=${IP_TAP1_2}/${PREFIX}    state=${UP_STATE}

Check Something_Other Has Not Changed
    No Operation

Delete Something
    Delete VPP Interface    agent_vpp_1    ${NAME_TAP1}

Check Something Has Been Deleted
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    vpp_term: Interface Not Exists  node=agent_vpp_1    mac=${MAC_TAP1_2}

Check Something_Other Is Still Configured
    No Operation

Show Interfaces And Other Objects After Setup
    vpp_term: Show Interfaces    agent_vpp_1
    Write To Machine    agent_vpp_1_term    show int addr
    Write To Machine    agent_vpp_1_term    show h
    Write To Machine    agent_vpp_1_term    show br
    Write To Machine    agent_vpp_1_term    show br 1 detail
    Write To Machine    agent_vpp_1_term    show vxlan tunnel
    Write To Machine    agent_vpp_1_term    show err
    vat_term: Interfaces Dump    agent_vpp_1
    Execute In Container    agent_vpp_1    ip a

*** Keywords ***

TestSetup
    Make Datastore Snapshots    ${TEST_NAME}_test_setup

TestTeardown
    Make Datastore Snapshots    ${TEST_NAME}_test_teardown
