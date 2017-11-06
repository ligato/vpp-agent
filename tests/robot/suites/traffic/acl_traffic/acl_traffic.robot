*** Settings ***
Library      OperatingSystem
#Library      RequestsLibrary
#Library      SSHLibrary      timeout=60s
#Library      String

Resource     ../../../variables/${VARIABLES}_variables.robot

Resource     ../../../libraries/all_libs.robot

Suite Setup       Testsuite Setup
Suite Teardown    Suite Cleanup
Test Setup        TestSetup
Test Teardown     TestTeardown

*** Variables ***
${VARIABLES}=          common
${ENV}=                common
${FINAL_SLEEP}=        3s
${SYNC_SLEEP}=         10s
${VETH1_MAC}=          1a:00:00:11:11:11
${VETH2_MAC}=          2a:00:00:22:22:22
${AGENT1_VETH_MAC}=    1a:00:00:11:11:11
${AGENT2_VETH_MAC}=    2a:00:00:22:22:22


*** Test Cases ***
Configure Environment
    [Tags]    setup
    Configure Environment 2
    Sleep    ${SYNC_SLEEP}

Check AfPackets On Vswitch
    vat_term: Check Afpacket Interface State    agent_vpp_1    IF_AFPIF_VSWITCH_agent_1_agent1_afpacket1    enabled=1
    linux: Interface Is Created    node=agent_vpp_1    mac=${VETH1_MAC}
    linux: Check Veth Interface State     agent_vpp_1    vpp1_veth1   mac=${VETH1_MAC}    ipv4=10.10.1.1/24    mtu=1500    state=up
    linux: Interface Is Created    node=agent_vpp_1    mac=${VETH2_MAC}
    linux: Check Veth Interface State     agent_vpp_1    vpp1_veth2   mac=${VETH2_MAC}    ipv4=10.10.1.1/24    mtu=1500    state=up

Check Veth Interface On Agent1
    linux: Interface Is Created    node=agent_1    mac=${AGENT1_VETH_MAC}
    linux: Check Veth Interface State     agent_1    agent1_veth1    mac=${AGENT1_VETH1_MAC}    ipv4=10.10.1.1/24    mtu=1500    state=up

Check Veth Interface On Agent2
    linux: Interface Is Created    node=agent_2    mac=${AGENT2_VETH_MAC}
    linux: Check Veth Interface State     agent_2    agent2_veth1    mac=${AGENT2_VETH_MAC}    ipv4=10.10.1.1/24    mtu=1500    state=up

Show All Objects
    Show Interfaces And Other Objects

Check Ping Agent1 -> Agent2
    linux: Check Ping    agent_1    10.0.0.11

Check Ping Agent2 -> Agent1
    linux: Check Ping    agent_2    10.0.0.10

Remove Agent Nodes
    Remove All Nodes

Start Agent Nodes Again
    Add Agent VPP Node    agent_vpp_1    vswitch=${TRUE}
    Add Agent VPP Node    agent_1
    Add Agent VPP Node    agent_2
    Sleep    ${SYNC_SLEEP}

Check AfPackets On Vswitch After Resync
    vat_term: Check Afpacket Interface State    agent_vpp_1    IF_AFPIF_VSWITCH_agent_1_agent1_afpacket1    enabled=1
    linux: Interface Is Created    node=agent_vpp_1    mac=${VETH1_MAC}
    linux: Check Veth Interface State     agent_vpp_1    vpp1_veth1   mac=${VETH1_MAC}    ipv4=10.10.1.1/24    mtu=1500    state=up
    linux: Interface Is Created    node=agent_vpp_1    mac=${VETH2_MAC}
    linux: Check Veth Interface State     agent_vpp_1    vpp1_veth2   mac=${VETH2_MAC}    ipv4=10.10.1.1/24    mtu=1500    state=up

Check Veth Interface On Agent1 After Resync
    linux: Interface Is Created    node=agent_1    mac=${AGENT1_VETH_MAC}
    linux: Check Veth Interface State     agent_1    agent1_veth1    mac=${AGENT1_VETH1_MAC}    ipv4=10.10.1.1/24    mtu=1500    state=up

Check Veth Interface On Agent2 After Resync
    linux: Interface Is Created    node=agent_2    mac=${AGENT2_VETH_MAC}
    linux: Check Veth Interface State     agent_2    agent2_veth1    mac=${AGENT2_VETH_MAC}    ipv4=10.10.1.1/24    mtu=1500    state=up

Show All Objects After Resync
    Show Interfaces And Other Objects

Check Ping Agent1 -> Agent2 After Resync
    linux: Check Ping    agent_1    10.0.0.11

Check Ping Agent2 -> Agent1 After Resync
    linux: Check Ping    agent_2    10.0.0.10

Done
    [Tags]    debug
    No Operation

Final Sleep For Manual Checking
    [Tags]    debug
    Sleep   ${FINAL_SLEEP}

*** Keywords ***
Show Interfaces And Other Objects
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
    Execute In Container    agent_1    ip a
    Execute In Container    agent_2    ip a
    Make Datastore Snapshots    before_resync



TestSetup
    Make Datastore Snapshots    ${TEST_NAME}_test_setup

TestTeardown
    Make Datastore Snapshots    ${TEST_NAME}_test_teardown

Suite Cleanup
    Stop SFC Controller Container
    Testsuite Teardown