*** Settings ***
Library      OperatingSystem
#Library      RequestsLibrary
#Library      SSHLibrary      timeout=60s
#Library      String

Resource     ../../variables/${VARIABLES}_variables.robot

Resource     ../../libraries/all_libs.robot

Force Tags        crud     IPv6    ExpectedFailure
Suite Setup       Testsuite Setup
Suite Teardown    Testsuite Teardown
Test Setup        TestSetup
Test Teardown     TestTeardown

*** Variables ***

${RESYNC_SLEEP}=       45s
${VARIABLES}=        common
${ENV}=              common
${NAME_TAP1}=        vpp1_tap1
${NAME_TAP2}=        vpp1_tap2
${MAC_TAP1}=         12:21:21:11:11:11
${MAC_TAP2}=         22:21:21:22:22:22
${IP_TAP1}=          fd33::1:b:0:0:1
${IP_STN_RULE}=      fd31::1:b:0:0:1
${IP_TAP2}=          fd33::1:b:0:0:2
${PREFIX}=           64
${MTU}=              4800
${UP_STATE}=         up
${RULE_NAME}         rule1
${WAIT_TIMEOUT}=     20s
${SYNC_SLEEP}=       3s

*** Test Cases ***
Configure Environment
    [Tags]    setup
    Configure Environment 1

Show Interfaces Before Setup
    Show Interfaces And Other Objects

Add TAP1 Interface
    vpp_term: Interface Not Exists  node=agent_vpp_1    mac=${MAC_TAP1}
    Put TAP Interface With IP    node=agent_vpp_1    name=${NAME_TAP1}    mac=${MAC_TAP1}    ip=${IP_TAP1}    prefix=${PREFIX}    host_if_name=linux_${NAME_TAP1}

Check TAP1 Interface Is Created
    ${interfaces}=       vat_term: Interfaces Dump    node=agent_vpp_1
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    vpp_term: Interface Is Created    node=agent_vpp_1    mac=${MAC_TAP1}
    ${actual_state}=    vpp_term: Check TAP IP6 interface State    agent_vpp_1    ${NAME_TAP1}    mac=${MAC_TAP1}    ipv6=${IP_TAP1}/${PREFIX}    state=${UP_STATE}

Add STN Rule
    Put STN Rule    node=agent_vpp_1    interface=${NAME_TAP1}    ip=${IP_STN_RULE}    rule_name=${RULE_NAME}

Check STN Rule Is Created
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    vpp_term: Check STN Rule State    node=agent_vpp_1    interface=${NAME_TAP1}    ip=${IP_STN_RULE}

Check TAP1 Interface Is Still Configured
    ${actual_state}=    vpp_term: Check TAP IP6 interface State    agent_vpp_1    ${NAME_TAP1}    mac=${MAC_TAP1}    ipv6=${IP_TAP1}/${PREFIX}    state=${UP_STATE}

Add TAP2 Interface
    vpp_term: Interface Not Exists  node=agent_vpp_1    mac=${MAC_TAP2}
    Put TAP Interface With IP    node=agent_vpp_1    name=${NAME_TAP2}    mac=${MAC_TAP2}    ip=${IP_TAP2}    prefix=${PREFIX}    host_if_name=linux_${NAME_TAP2}

Check TAP2 Interface Is Created
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    vpp_term: Interface Is Created    node=agent_vpp_1    mac=${MAC_TAP2}
    ${actual_state}=    vpp_term: Check TAP IP6 interface State    agent_vpp_1    ${NAME_TAP2}    mac=${MAC_TAP2}    ipv6=${IP_TAP2}/${PREFIX}    state=${UP_STATE}

Update STN Rule
    Put STN Rule    node=agent_vpp_1    interface=${NAME_TAP2}    ip=${IP_STN_RULE}    rule_name=${RULE_NAME}

Check STN Rule Is Updated
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    vpp_term: Check STN Rule State    node=agent_vpp_1    interface=${NAME_TAP2}    ip=${IP_STN_RULE}

Delete STN Rule
    Delete STN Rule    node=agent_vpp_1    rule_name=${RULE_NAME}

Check Deleted STN Rule
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    vpp_term: Check STN Rule Deleted    node=agent_vpp_1    interface=${NAME_TAP2}    ip=${IP_STN_RULE}

Add STN Rule Again
    Put STN Rule    node=agent_vpp_1    interface=${NAME_TAP1}    ip=${IP_STN_RULE}    rule_name=${RULE_NAME}

Check STN Rule Is Created Again
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    vpp_term: Check STN Rule State    node=agent_vpp_1    interface=${NAME_TAP1}    ip=${IP_STN_RULE}

Remove VPP Node
    Remove Node     agent_vpp_1
    Sleep    ${SYNC_SLEEP}

Start VPP Node
    Add Agent VPP Node    agent_vpp_1    vswitch=${TRUE}
    Sleep    ${RESYNC_SLEEP}
    Show Interfaces And Other Objects

Check STN Rule Is Created After Resync
    vpp_term: Check STN Rule State    node=agent_vpp_1    interface=${NAME_TAP1}    ip=${IP_STN_RULE}


*** Keywords ***
Show Interfaces And Other Objects
    vpp_term: Show Interfaces    agent_vpp_1
    Write To Machine    agent_vpp_1_term    show int
    Write To Machine    agent_vpp_1_term    show stn rules
    Write To Machine    agent_vpp_1_term    show int addr
    Write To Machine    agent_vpp_1_term    show h
    Write To Machine    agent_vpp_1_term    show brl
    Write To Machine    agent_vpp_1_term    show br 1 detail
    Write To Machine    agent_vpp_1_term    show vxlan tunnel
    Write To Machine    agent_vpp_1_term    show err
    vat_term: Interfaces Dump    agent_vpp_1
    Execute In Container    agent_vpp_1    ip a

TestSetup
    Make Datastore Snapshots    ${TEST_NAME}_test_setup

TestTeardown
    Make Datastore Snapshots    ${TEST_NAME}_test_teardown
