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
${SYNC_SLEEP}=         10s
${RESYNC_SLEEP}=       25s

${AGENT1_VETH_MAC}=    02:00:00:00:00:01
${AGENT2_VETH_MAC}=    02:00:00:00:00:02

${VARIABLES}=       common
${ENV}=             common


*** Test Cases ***
Configure Environment
    [Tags]    setup
    Configure Environment 2     veth_basic.conf
    Sleep    ${SYNC_SLEEP}
    Show Interfaces And Other Objects

Check Stuff At Beginning
    Check Stuff

Check Ping Node1 -> Node2
    linux: Check Ping    node_1    10.0.0.11

Check Ping Node2 -> Node1
    linux: Check Ping    node_2    10.0.0.10

Remove All Nodes
    Remove Node     agent_vpp_1
    Remove Node     node_1
    Remove Node     node_2
    Sleep    ${SYNC_SLEEP}

Start All Nodes
    Add Agent VPP Node    agent_vpp_1    vswitch=${TRUE}
    Add Agent Node    node_1
    Add Agent Node    node_2
    Sleep    ${RESYNC_SLEEP}

Check Stuff After Resync
    Check Stuff

Check Ping Node1 -> Node2 After Resync
    linux: Check Ping    node_1    10.0.0.11

Check Ping Node2 -> Node1 After Resync
    linux: Check Ping    node_2    10.0.0.10

Remove VPP
    Remove Node     agent_vpp_1
    Sleep    ${SYNC_SLEEP}

Start VPP
    Add Agent VPP Node    agent_vpp_1    vswitch=${TRUE}
    Sleep    ${RESYNC_SLEEP}

Check Stuff After VPP Restart
    Check Stuff

Check Ping Node1 -> Node2 After VPP Restart
    linux: Check Ping    node_1    10.0.0.11

Check Ping Node2 -> Node1 After VPP Restart
    linux: Check Ping    node_2    10.0.0.10

Remove Node1
    Remove Node     node_1
    Sleep    ${SYNC_SLEEP}

Start Node1
    Add Agent Node    node_1
    Sleep    ${RESYNC_SLEEP}

Check Stuff After Node1 Restart
    Check Stuff

Check Ping Node1 -> Node2 After Node1 Restart
    linux: Check Ping    node_1    10.0.0.11

Check Ping Node2 -> Node1 After Node1 Restart
    linux: Check Ping    node_2    10.0.0.10

Remove Node1 Again
    Remove Node     node_1
    Sleep    ${SYNC_SLEEP}

Start Node1 Again
    Add Agent Node    node_1
    Sleep    ${RESYNC_SLEEP}

Check Stuff After Node1 Restart Again
    Check Stuff

Check Ping Node1 -> Node2 After Node1 Restart Again
    linux: Check Ping    node_1    10.0.0.11

Check Ping Node2 -> Node1 After Node1 Restart Again
    linux: Check Ping    node_2    10.0.0.10

Remove Node2
    Remove Node     node_2
    Sleep    ${SYNC_SLEEP}

Start Node2
    Add Agent Node    node_2
    Sleep    ${RESYNC_SLEEP}

Check Stuff After Node2 Restart
    Check Stuff

Check Ping Node1 -> Node2 After Node2 Restart
    linux: Check Ping    node_1    10.0.0.11

Check Ping Node2 -> Node1 After Node2 Restart
    linux: Check Ping    node_2    10.0.0.10

Remove Node2 Again
    Remove Node     node_2
    Sleep    ${SYNC_SLEEP}

Start Node2 Again
    Add Agent Node    node_2
    Sleep    ${RESYNC_SLEEP}

Check Stuff After Node2 Restart Again
    Check Stuff

Check Ping Node1 -> Node2 After Node2 Restart Again
    linux: Check Ping    node_1    10.0.0.11

Check Ping Node2 -> Node1 After Node2 Restart Again
    linux: Check Ping    node_2    10.0.0.10

Remove Node 1 and Node2
    Remove Node     node_1
    Remove Node     node_2
    Sleep    ${SYNC_SLEEP}

Start Node 1 and Node2
    Add Agent Node    node_1
    Add Agent Node    node_2
    Sleep    ${RESYNC_SLEEP}

Check Stuff After Node1 and Node2 Restart
    Check Stuff

Check Ping Node1 -> Node2 After Node2 Restart
    linux: Check Ping    node_1    10.0.0.11

Check Ping Node2 -> Node1 After Node2 Restart
    linux: Check Ping    node_2    10.0.0.10

Remove Node 1 and Node2 Again
    Remove Node     node_1
    Remove Node     node_2
    Sleep    ${SYNC_SLEEP}

Start Node 1 and Node2 Again
    Add Agent Node    node_1
    Add Agent Node    node_2
    Sleep    ${RESYNC_SLEEP}

Check Stuff After Node1 and Node2 Restart Again
    Check Stuff

Check Ping Node1 -> Node2 After Node2 Restart Again
    linux: Check Ping    node_1    10.0.0.11

Check Ping Node2 -> Node1 After Node2 Restart Again
    linux: Check Ping    node_2    10.0.0.10


Done
    [Tags]    debug
    No Operation


Remove Agent Nodes Again
    Remove All Nodes

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
    Execute In Container    node_1    ip a
    Execute In Container    node_2    ip a
    linux: Check Processes on Node      node_1
    linux: Check Processes on Node      node_2
    Make Datastore Snapshots    before_resync

Check Stuff
    vat_term: Check Afpacket Interface State    agent_vpp_1    IF_AFPIF_VSWITCH_node_1_nod1_veth    enabled=1
    vat_term: Check Afpacket Interface State    agent_vpp_1    IF_AFPIF_VSWITCH_node_2_nod2_veth    enabled=1
    linux: Interface With IP Is Created    node=node_1    mac=${AGENT1_VETH_MAC}      ipv4=10.0.0.10/24
    linux: Interface With IP Is Created    node=node_2    mac=${AGENT2_VETH_MAC}      ipv4=10.0.0.11/24
    vat_term: BD Is Created    agent_vpp_1    IF_AFPIF_VSWITCH_node_1_nod1_veth    IF_AFPIF_VSWITCH_node_2_nod2_veth
    Show Interfaces And Other Objects


TestSetup
    Make Datastore Snapshots    ${TEST_NAME}_test_setup

TestTeardown
    Make Datastore Snapshots    ${TEST_NAME}_test_teardown

Suite Cleanup
    Stop SFC Controller Container
    Testsuite Teardown