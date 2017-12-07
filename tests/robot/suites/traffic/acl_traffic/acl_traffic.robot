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
${RESYNC_SLEEP}=     20s

${AGENT1_VETH_MAC}=    02:00:00:00:00:01
${AGENT2_VETH_MAC}=    02:00:00:00:00:02
${REPLY_DATA_FOLDER}            replyACL
${VARIABLES}=       common
${ENV}=             common
${ACL1_NAME}=       acl1_tcp
${ACL2_NAME}=       acl2_tcp
${ACL3_NAME}=       acl3_UDP
${ACL4_NAME}=       acl4_UDP
${ACL5_NAME}=       acl5_ICMP
${ACL6_NAME}=       acl6_ICMP
${E_INTF1}=         IF_AFPIF_VSWITCH_agent_2_agent2_veth
${I_INTF1}=         IF_AFPIF_VSWITCH_agent_1_agent1_veth
${E_INTF2}=         IF_AFPIF_VSWITCH_agent_1_agent1_veth
${I_INTF2}=         IF_AFPIF_VSWITCH_agent_2_agent2_veth
${RULE_NM1_1}=         acl1_rule1
${RULE_NM2_1}=         acl2_rule1
${RULE_NM3_1}=         acl3_rule1
${RULE_NM4_1}=         acl3_rule1

${ACTION_DENY}=     1
${ACTION_PERMIT}=   2
${DEST_NTW}=        10.0.0.0/24
${SRC_NTW}=         10.0.0.0/24
${NO_PORT}=
${TCP_PORT}=     3000
${UDP_PORT}=     3001



*** Test Cases ***
Configure Environment
    [Tags]    setup
    Configure Environment 2      acl_basic.conf
    Sleep    ${SYNC_SLEEP}
    Show Interfaces And Other Objects

Check AfPackets On Vswitch
    vat_term: Check Afpacket Interface State    agent_vpp_1    IF_AFPIF_VSWITCH_agent_1_agent1_veth    enabled=1
    vat_term: Check Afpacket Interface State    agent_vpp_1    IF_AFPIF_VSWITCH_agent_2_agent2_veth    enabled=1

Check Veth Interface On Agent1
    linux: Interface With IP Is Created    node=agent_1    mac=${AGENT1_VETH_MAC}      ipv4=10.0.0.10/24
    # status check not implemented in linux plugin
    #linux: Check Veth Interface State     agent_vpp_1    agent1_veth     mac=${AGENT1_VETH_MAC}    ipv4=10.0.0.10/24    mtu=1500    state=up

Check Veth Interface On Agent2
    linux: Interface With IP Is Created    node=agent_2    mac=${AGENT2_VETH_MAC}      ipv4=10.0.0.11/24
    # status check not implemented in linux plugin
    #linux: Check Veth Interface State     agent_vpp_1    agent2_veth     mac=${AGENT2_VETH_MAC}    ipv4=10.0.0.11/24    mtu=1500    state=up

Check Bridge Domain Is Created
    vat_term: BD Is Created    agent_vpp_1    IF_AFPIF_VSWITCH_agent_1_agent1_veth    IF_AFPIF_VSWITCH_agent_2_agent2_veth


Show All Objects
    Show Interfaces And Other Objects

Start TCP And UDP Listeners
    Start UDP and TCP Ping Servers

Check Ping Agent1 -> Agent2
    linux: Check Ping    agent_1    10.0.0.11

Check Ping Agent2 -> Agent1
    linux: Check Ping    agent_2    10.0.0.10

Check UDP Ping Agent1 -> Agent2
    linux: UDPPing  agent_1     10.0.0.11   ${UDP_PORT}

Check TCP Ping Agent1 -> Agent2
    linux: TCPPing  agent_1     10.0.0.11   ${TCP_PORT}

Check UDP Ping Agent2 -> Agent1
    linux: UDPPing  agent_2     10.0.0.10   ${UDP_PORT}

Check TCP Ping Agent2 -> Agent1
    linux: TCPPing  agent_2     10.0.0.10   ${TCP_PORT}


Add ACL1_TCP Disable TCP Port
    vpp_ctl: Put ACL TCP   agent_vpp_1   ${ACL1_NAME}    ${E_INTF1}    ${I_INTF1}   ${RULE_NM1_1}    ${ACTION_DENY}     ${DEST_NTW}     ${SRC_NTW}   ${TCP_PORT}   ${TCP_PORT}    ${TCP_PORT}   ${TCP_PORT}
    Sleep    ${SYNC_SLEEP}

Check ACL1_TCP is created
    vpp_ctl: Check ACL Reply    agent_vpp_1    ${ACL1_NAME}    ${REPLY_DATA_FOLDER}/reply_acl1_tcp.txt    ${REPLY_DATA_FOLDER}/reply_acl1_tcp_term.txt

ADD ACL1_UDP Disable UDP Port
    vpp_ctl: Put ACL UDP    agent_vpp_1    ${ACL3_NAME}    ${E_INTF1}   ${I_INTF1}    ${E_INTF2}    ${I_INTF2}    ${RULE_NM3_1}    ${ACTION_DENY}    ${DEST_NTW}     ${SRC_NTW}   ${UDP_PORT}   ${UDP_PORT}    ${UDP_PORT}   ${UDP_PORT}
    Sleep    ${SYNC_SLEEP}

Check ACL1_UDP Is Created
    vpp_ctl: Check ACL Reply    agent_vpp_1    ${ACL3_NAME}    ${REPLY_DATA_FOLDER}/reply_acl3_UDP.txt    ${REPLY_DATA_FOLDER}/reply_acl3_UDP_term.txt

Show ACLs on VPP
    vpp_term: Show ACL      agent_vpp_1

Check UDP Not Ping Agent2 -> Agent1 After Disabling
    linux: UDPPingNot  agent_2     10.0.0.10   ${UDP_PORT}

Check UDP Not Ping Agent1 -> Agent2 After Disabling
    linux: UDPPingNot  agent_1     10.0.0.11   ${UDP_PORT}

Check TCP Not Ping Agent1 -> Agent2
    linux: TCPPingNot  agent_1     10.0.0.11   ${TCP_PORT}

Check TCP Not Ping Agent2 -> Agent1
    linux: TCPPingNot  agent_2     10.0.0.10   ${TCP_PORT}

Remove Agent Nodes
    Remove All Nodes
    Sleep    ${RESYNC_SLEEP}

Start Agent Nodes Again
    Add Agent VPP Node    agent_vpp_1    vswitch=${TRUE}
    Add Agent Node    agent_1
    Add Agent Node    agent_2
    Sleep    ${SYNC_SLEEP}

Check AfPackets On Vswitch After Resync
    vat_term: Check Afpacket Interface State    agent_vpp_1    IF_AFPIF_VSWITCH_agent_1_agent1_veth    enabled=1
    vat_term: Check Afpacket Interface State    agent_vpp_1    IF_AFPIF_VSWITCH_agent_2_agent2_veth    enabled=1

Check Veth Interface On Agent1 After Resync
    linux: Interface With IP Is Created    node=agent_1    mac=${AGENT1_VETH_MAC}      ipv4=10.0.0.10/24
    # status check not implemented in linux plugin
    #linux: Check Veth Interface State     agent_vpp_1    agent1_veth     mac=${AGENT1_VETH_MAC}    ipv4=10.0.0.10/24    mtu=1500    state=up

Check Veth Interface On Agent2 After Resync
   linux: Interface With IP Is Created    node=agent_2    mac=${AGENT2_VETH_MAC}      ipv4=10.0.0.11/24
    # status check not implemented in linux plugin
    #linux: Check Veth Interface State     agent_vpp_1    agent2_veth     mac=${AGENT2_VETH_MAC}    ipv4=10.0.0.11/24    mtu=1500    state=up

Check Bridge Domain Is Created After Resync
    vat_term: BD Is Created    agent_vpp_1    IF_AFPIF_VSWITCH_agent_1_agent1_veth    IF_AFPIF_VSWITCH_agent_2_agent2_veth

Show All Objects After Resync
    Show Interfaces And Other Objects

Show ACLs on VPP After Resync
    vpp_term: Show ACL      agent_vpp_1

Start TCP And UDP Listeners After Resync
    Start UDP and TCP Ping Servers

Check Ping Agent1 -> Agent2 After Resync
    linux: Check Ping    agent_1    10.0.0.11

Check Ping Agent2 -> Agent1 After Resync
    linux: Check Ping    agent_2    10.0.0.10

Check UDP Not Ping Agent1 -> Agent2 After Resync
    linux: UDPPingNot  agent_1     10.0.0.11   ${UDP_PORT}

Check UDP Not Ping Agent2 -> Agent1 After Resync
    linux: UDPPingNot  agent_2     10.0.0.10   ${UDP_PORT}

Check TCP Not Ping Agent1 -> Agent2 After Resync
    linux: TCPPingNot  agent_1     10.0.0.11   ${TCP_PORT}

Check TCP Not Ping Agent2 -> Agent1 After Resync
    linux: TCPPingNot  agent_2     10.0.0.10   ${TCP_PORT}





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
    Execute In Container    agent_1    ip a
    Execute In Container    agent_2    ip a
    linux: Check Processes on Node      agent_1
    linux: Check Processes on Node      agent_2
    Make Datastore Snapshots    before_resync

Start UDP and TCP Ping Servers
    linux: Run TCP Ping Server On Node      agent_1     ${TCP_PORT}
    linux: Run UDP Ping Server On Node      agent_1     ${UDP_PORT}
    linux: Run TCP Ping Server On Node      agent_2     ${TCP_PORT}
    linux: Run UDP Ping Server On Node      agent_2     ${UDP_PORT}
    linux: Check Processes on Node      agent_1
    linux: Check Processes on Node      agent_2
    Sleep    ${SYNC_SLEEP}


TestSetup
    Make Datastore Snapshots    ${TEST_NAME}_test_setup

TestTeardown
    Make Datastore Snapshots    ${TEST_NAME}_test_teardown

Suite Cleanup
    Stop SFC Controller Container
    Testsuite Teardown