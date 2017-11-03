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
${MAC_TAP1_2}=       22:21:21:11:11:11
${MAC_TAP2}=         22:21:21:22:22:22
${IP_TAP1}=          20.20.1.1
${IP_TAP1_2}=        21.20.1.1
${IP_TAP2}=          20.20.1.2
${PREFIX}=           24
${MTU}=              4800

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
    ${actual_state}=    Check TAP interface State    agent_vpp_1    ${NAME_TAP1}    up    ${MAC_TAP1}    ${IP_TAP1}/${PREFIX}

Add TAP2 Interface
    vpp_term: Interface Not Exists  node=agent_vpp_1    mac=${MAC_TAP2}
    vpp_ctl: Put TAP Interface With IP    node=agent_vpp_1    name=${NAME_TAP2}    mac=${MAC_TAP2}    ip=${IP_TAP2}    prefix=${PREFIX}    host_if_name=linux_${NAME_TAP2}

Check TAP2 Interface Is Created
    vpp_term: Interface Is Created    node=agent_vpp_1    mac=${MAC_TAP2}
    ${actual_state}=    Check TAP interface State    agent_vpp_1    ${NAME_TAP2}    up     ${MAC_TAP2}    ${IP_TAP2}/${PREFIX}

Check TAP1 Interface Is Still Configured
    ${actual_state}=    Check TAP interface State    agent_vpp_1    ${NAME_TAP1}    up    ${MAC_TAP1}    ${IP_TAP1}/${PREFIX}

Update TAP1 Interface
    vpp_ctl: Put TAP Interface With IP    node=agent_vpp_1    name=${NAME_TAP1}    mac=${MAC_TAP1_2}    ip=${IP_TAP1_2}    prefix=${PREFIX}    host_if_name=linux_${NAME_TAP1}

Check TAP1_2 Interface Is Created
    vpp_term: Interface Is Created    node=agent_vpp_1    mac=${MAC_TAP1_2}
    ${actual_state}=    Check TAP interface State    agent_vpp_1    ${NAME_TAP1}    up    ${MAC_TAP1_2}    ${IP_TAP1_2}/${PREFIX}

Check TAP2 Interface Has Not Changed
    ${actual_state}=    Check TAP interface State    agent_vpp_1    ${NAME_TAP2}    up     ${MAC_TAP2}    ${IP_TAP2}/${PREFIX}

Delete TAP1_2 Interface
    vpp_ctl: Delete VPP Interface    agent_vpp_1    ${NAME_TAP1}

Check TAP1_2 Interface Has Been Deleted
    vpp_term: Interface Not Exists  node=agent_vpp_1    mac=${MAC_TAP1_2}

Check TAP2 Interface Is Still Configured
    ${actual_state}=    Check TAP interface State    agent_vpp_1    ${NAME_TAP2}    up     ${MAC_TAP2}    ${IP_TAP2}/${PREFIX}

Show Interfaces And Other Objects After Setup
    vpp_term: Show Interfaces    agent_vpp_1
    Write To Machine    agent_vpp_1_term    show int addr
    Write To Machine    agent_vpp_1_term    show h
    Write To Machine    agent_vpp_1_term    show br
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

Check TAP interface State
    [Arguments]          ${node}    ${name}    ${state}    ${mac}    ${ipv4}
    Log Many             ${node}    ${name}    ${state}    ${mac}    ${ipv4}
    @{desired_ipv4}=     Create List    ${ipv4}
    Log                  @{desired_ipv4}
    @{desired_state}=    Create List    mac=${mac}    ipv4=@{desired_ipv4}
    ${internal_name}=    vpp_ctl: Get Interface Internal Name    ${node}    ${name}
    Log                  ${internal_name}
    ${interface}=        vpp_term: Show Interfaces    ${node}    ${internal_name}
    Log                  ${interface}
    Should Contain       ${interface}    ${state}
    ${ipv4}=             vpp_term: Get Interface IPs    ${node}     ${internal_name}
    Log                  ${ipv4}
    ${mac}=              vpp_term: Get Interface MAC    ${node}    ${internal_name}
    Log                  ${mac}
    ${actual_state}=     Create List    mac=${mac}    ipv4=${ipv4}
    Log List             ${actual_state}
    List Should Contain Sub List    ${actual_state}    ${desired_state}
    [Return]             ${actual_state}
