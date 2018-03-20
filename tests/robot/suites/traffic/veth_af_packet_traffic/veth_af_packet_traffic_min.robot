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
${FINAL_SLEEP}=        5s
${SYNC_SLEEP}=         20s
${RESYNC_SLEEP}=       45s

${AGENT1_VETH_MAC}=    02:00:00:00:00:01
${AGENT2_VETH_MAC}=    02:00:00:00:00:02
${AGENT3_VETH_MAC}=    02:00:00:00:00:03

${VARIABLES}=       common
${ENV}=             common


*** Test Cases ***
Configure Environment
    [Tags]    setup
    Configure Environment 4     veth_basic.conf
    Sleep    ${SYNC_SLEEP}
    Show Interfaces And Other Objects

Check Stuff At Beginning
    Check Stuff

Check Ping At Beginning
    Check all Pings

Remove VPP And Two Nodes
    Remove Node     agent_vpp_1
    Remove Node     node_1
    Remove Node     node_2
    Remove Node     node_3
    Sleep    ${SYNC_SLEEP}

Start VPP And Two Nodes
    Add Agent VPP Node    agent_vpp_1    vswitch=${TRUE}
    Add Agent Node    node_1
    Add Agent Node    node_2
    Add Agent Node    node_3
    Sleep    55s

Check Stuff After Resync
    Check Stuff

Check Ping After Resync
    Check all Pings

Remove VPP
    Remove Node     agent_vpp_1
    Sleep    ${SYNC_SLEEP}

Start VPP
    Add Agent VPP Node    agent_vpp_1    vswitch=${TRUE}
    Sleep    ${RESYNC_SLEEP}

Check Stuff After VPP Restart
    Check Stuff

Check Ping After VPP Restart
    Check all Pings

Remove Node1
    Remove Node     node_1
    Sleep    ${SYNC_SLEEP}

Start Node1
    Add Agent Node    node_1
    Sleep    ${RESYNC_SLEEP}

Check Stuff After Node1 Restart
    Check Stuff

Check Ping After Node1 Restart
    Check all Pings


*** Keywords ***
Check all Pings
    linux: Check Ping    node_1    10.0.0.11
    linux: Check Ping    node_1    10.0.0.12
    linux: Check Ping    node_2    10.0.0.10
    linux: Check Ping    node_2    10.0.0.12
    linux: Check Ping    node_3    10.0.0.10
    linux: Check Ping    node_3    10.0.0.11

Show Interfaces And Other Objects
    vpp_term: Show Interfaces    agent_vpp_1
    Write To Machine    agent_vpp_1_term    show int addr
    Write To Machine    agent_vpp_1_term    show h
    Write To Machine    agent_vpp_1_term    show br
    Write To Machine    agent_vpp_1_term    show err
    vat_term: Interfaces Dump    agent_vpp_1
    Write To Machine    vpp_agent_ctl    vpp-agent-ctl ${AGENT_VPP_ETCD_CONF_PATH} -ps
    Execute In Container    agent_vpp_1    ip a
    Execute In Container    node_1    ip a
    Execute In Container    node_2    ip a
    Execute In Container    node_3    ip a
    Make Datastore Snapshots    before_check stuff

Check Stuff
    Show Interfaces And Other Objects
    vat_term: Check Afpacket Interface State    agent_vpp_1    IF_AFPIF_VSWITCH_node_1_nod1_veth    enabled=1
    vat_term: Check Afpacket Interface State    agent_vpp_1    IF_AFPIF_VSWITCH_node_2_nod2_veth    enabled=1
    vat_term: Check Afpacket Interface State    agent_vpp_1    IF_AFPIF_VSWITCH_node_3_nod3_veth    enabled=1
    linux: Interface With IP Is Created    node=node_1    mac=${AGENT1_VETH_MAC}      ipv4=10.0.0.10/24
    linux: Interface With IP Is Created    node=node_2    mac=${AGENT2_VETH_MAC}      ipv4=10.0.0.11/24
    linux: Interface With IP Is Created    node=node_3    mac=${AGENT3_VETH_MAC}      ipv4=10.0.0.12/24
    vat_term: BD Is Created    agent_vpp_1    IF_AFPIF_VSWITCH_node_1_nod1_veth    IF_AFPIF_VSWITCH_node_2_nod2_veth    IF_AFPIF_VSWITCH_node_3_nod3_veth



TestSetup
    Make Datastore Snapshots    ${TEST_NAME}_test_setup

TestTeardown
    Make Datastore Snapshots    ${TEST_NAME}_test_teardown

Suite Cleanup
    Stop SFC Controller Container
    Testsuite Teardown