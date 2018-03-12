*** Settings ***
Library      OperatingSystem
#Library      RequestsLibrary
#Library      SSHLibrary      timeout=60s
#Library      String

Resource     ../../variables/${VARIABLES}_variables.robot

Resource     ../../libraries/all_libs.robot

Suite Setup       Testsuite Setup
Suite Teardown    Testsuite Teardown
Test Setup        TestSetup
Test Teardown     TestTeardown

*** Variables ***
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

*** Test Cases ***
Configure Environment
    [Tags]    setup
    Configure Environment 1

Show Interfaces Before Setup
    ${interfaces}=    vpp_term: Show Interfaces    agent_vpp_1

Add TAP1 Interface
    vpp_term: Interface Not Exists  node=agent_vpp_1    mac=${MAC_TAP1}
    vpp_ctl: Put TAP Interface With IP    node=agent_vpp_1    name=${NAME_TAP1}    mac=${MAC_TAP1}    ip=${IP_TAP1}    prefix=${PREFIX}    host_if_name=linux_${NAME_TAP1}

Check TAP1 Interface Is Created
    ${interfaces}=       vat_term: Interfaces Dump    node=agent_vpp_1
    Log                  ${interfaces}
    vpp_term: Interface Is Created    node=agent_vpp_1    mac=${MAC_TAP1}
    ${actual_state}=    vpp_term: Check TAP interface State    agent_vpp_1    ${NAME_TAP1}    mac=${MAC_TAP1}    ipv6=${IP_TAP1}/${PREFIX}    state=${UP_STATE}

Add STN Rule
    vpp_ctl: Put STN Rule    node=agent_vpp_1    interface=${NAME_TAP1}    ip=${IP_STN_RULE}    rule_name=${RULE_NAME}

Check STN Rule Is Created
    vpp_term: Check STN Rule State    node=agent_vpp_1    interface=${NAME_TAP1}    ipv6=${IP_STN_RULE}

Check TAP1 Interface Is Still Configured
    ${actual_state}=    vpp_term: Check TAP interface State    agent_vpp_1    ${NAME_TAP1}    mac=${MAC_TAP1}    ipv6=${IP_TAP1}/${PREFIX}    state=${UP_STATE}

Add TAP2 Interface
    vpp_term: Interface Not Exists  node=agent_vpp_1    mac=${MAC_TAP2}
    vpp_ctl: Put TAP Interface With IP    node=agent_vpp_1    name=${NAME_TAP2}    mac=${MAC_TAP2}    ip=${IP_TAP2}    prefix=${PREFIX}    host_if_name=linux_${NAME_TAP2}

Check TAP2 Interface Is Created
    vpp_term: Interface Is Created    node=agent_vpp_1    mac=${MAC_TAP2}
    ${actual_state}=    vpp_term: Check TAP interface State    agent_vpp_1    ${NAME_TAP2}    mac=${MAC_TAP2}    ipv6=${IP_TAP2}/${PREFIX}    state=${UP_STATE}

Update STN Rule
    vpp_ctl: Put STN Rule    node=agent_vpp_1    interface=${NAME_TAP2}    ip=${IP_STN_RULE}    rule_name=${RULE_NAME}

Check STN Rule Is Updated
    vpp_term: Check STN Rule State    node=agent_vpp_1    interface=${NAME_TAP2}    ipv6=${IP_STN_RULE}

Delete STN Rule
    vpp_ctl: Delete STN Rule    node=agent_vpp_1    rule_name=${RULE_NAME}

#TODO: Check Deleted STN Rule

Show Interfaces And Other Objects After Setup
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
    Write To Machine    vpp_agent_ctl    vpp-agent-ctl ${AGENT_VPP_ETCD_CONF_PATH} -ps
    Execute In Container    agent_vpp_1    ip a

*** Keywords ***
TestSetup
    Make Datastore Snapshots    ${TEST_NAME}_test_setup

TestTeardown
    Make Datastore Snapshots    ${TEST_NAME}_test_teardown
