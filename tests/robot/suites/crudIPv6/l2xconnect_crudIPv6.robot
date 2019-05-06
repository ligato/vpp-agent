*** Settings ***
Library      OperatingSystem
#Library      RequestsLibrary
#Library      SSHLibrary      timeout=60s
#Library      String

Resource     ../../variables/${VARIABLES}_variables.robot

Resource     ../../libraries/all_libs.robot

Force Tags        crud     IPv6
Suite Setup       Testsuite Setup
Suite Teardown    Testsuite Teardown
Test Setup        TestSetup
Test Teardown     TestTeardown

*** Variables ***
${VARIABLES}=          common
${ENV}=                common
${WAIT_TIMEOUT}=     20s
${SYNC_SLEEP}=       3s
${RESYNC_SLEEP}=     16s
${VETH1_MAC}=          1a:00:00:11:11:11
${VETH2_MAC}=          2a:00:00:22:22:22
${AFP1_MAC}=           a2:01:01:01:01:01
${VETH_IP}=            fd30::1:e:0:0:1
${PREFIX}=             64
${MEMIF_IP}=           fd31::1:1:0:0:1
${VXLAN_IP}=           fd31::1:1:0:0:2
${LOOPB_IP}=           fd32::1:1:0:0:1
${LOOPB2_IP}=          fd32:0:1:1:1::1
${TAP_IP}=             fd33::1:1:0:0:1
*** Test Cases ***
Configure Environment
    [Tags]    setup
    Configure Environment 5

Show Interfaces Before Setup
    vpp_term: Show Interfaces    agent_vpp_1

Add Veth1 Interface
    linux: Interface Not Exists    node=agent_vpp_1    mac=${VETH1_MAC}
    Put Veth Interface With IP    node=agent_vpp_1    name=vpp1_veth1    mac=${VETH1_MAC}    peer=vpp1_veth2    ip=${VETH_IP}    prefix=${PREFIX}    mtu=1500

Add Veth2 Interface
    linux: Interface Not Exists    node=agent_vpp_1    mac=${VETH2_MAC}
    Put Veth Interface    node=agent_vpp_1    name=vpp1_veth2    mac=${VETH2_MAC}    peer=vpp1_veth1

Check That Veth1 And Veth2 Interfaces Are Created
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    linux: Interface Is Created    node=agent_vpp_1    mac=${VETH1_MAC}
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    linux: Interface Is Created    node=agent_vpp_1    mac=${VETH2_MAC}
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    linux: Check Veth Interface State     agent_vpp_1    vpp1_veth1    mac=${VETH1_MAC}    ipv6=${VETH_IP}/${PREFIX}    mtu=1500    state=up
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    linux: Check Veth Interface State     agent_vpp_1    vpp1_veth2    mac=${VETH2_MAC}    state=up

Add Memif Interface
    Put Memif Interface With IP    node=agent_vpp_1    name=vpp1_memif1    mac=62:61:61:61:61:61    master=true    id=1    ip=${MEMIF_IP}    prefix=${PREFIX}   socket=default.sock

Check Memif Interface Created
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    vpp_term: Interface Is Created    node=agent_vpp_1    mac=62:61:61:61:61:61
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    vat_term: Check Memif Interface State     agent_vpp_1  vpp1_memif1  mac=62:61:61:61:61:61  role=master  id=1  ipv6=${MEMIF_IP}/${PREFIX}  connected=0  enabled=1  socket=${AGENT_VPP_1_MEMIF_SOCKET_FOLDER}/default.sock


Add VXLan Interface
    Put VXLan Interface    node=agent_vpp_1    name=vpp1_vxlan1    src=${MEMIF_IP}    dst=${VXLAN_IP}    vni=5

Check VXLan Interface Created
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    vxlan: Tunnel Is Created    node=agent_vpp_1    src=${MEMIF_IP}    dst=${VXLAN_IP}    vni=5
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    vat_term: Check VXLan Interface State    agent_vpp_1    vpp1_vxlan1    enabled=1    src=${MEMIF_IP}    dst=${VXLAN_IP}    vni=5

Add Loopback1 Interface
    Put Loopback Interface With IP    node=agent_vpp_1    name=vpp1_loop1    mac=12:21:21:11:11:11    ip=${LOOPB_IP}   prefix=${PREFIX}   mtu=1400

Check Loopback1 Interface Created
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    vpp_term: Interface Is Created    node=agent_vpp_1    mac=12:21:21:11:11:11
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    vat_term: Check Loopback Interface State    agent_vpp_1    vpp1_loop1    enabled=1     mac=12:21:21:11:11:11    mtu=1400  ipv6=${LOOPB_IP}/${PREFIX}

Add Loopback2 Interface
    Put Loopback Interface With IP    node=agent_vpp_1    name=vpp1_loop2    mac=22:21:21:11:11:11    ip=${LOOPB2_IP}   prefix=${PREFIX}   mtu=1400

Check Loopback2 Interface Created
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    vpp_term: Interface Is Created    node=agent_vpp_1    mac=22:21:21:11:11:11
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    vat_term: Check Loopback Interface State    agent_vpp_1    vpp1_loop2    enabled=1     mac=22:21:21:11:11:11    mtu=1400  ipv6=${LOOPB2_IP}/${PREFIX}

Add Tap Interface
    Put TAPv2 Interface With IP    node=agent_vpp_1    name=vpp1_tap1    mac=32:21:21:11:11:11    ip=${TAP_IP}   prefix=${PREFIX}      host_if_name=linux_vpp1_tap1

Check TAP Interface Created
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    vpp_term: Interface Is Created    node=agent_vpp_1    mac=32:21:21:11:11:11
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    vpp_term: Check TAP IP6 interface State    agent_vpp_1    vpp1_tap1    mac=32:21:21:11:11:11    ipv6=${TAP_IP}/${PREFIX}    state=up

Check Stuff
    Show Interfaces And Other Objects

Add L2XConnect1 for Memif and Loopback1
    Put L2XConnect  agent_vpp_1    vpp1_memif1    vpp1_loop1
    Put L2XConnect  agent_vpp_1    vpp1_loop1     vpp1_memif1

Check L2XConnect1 Memif and Loopback1 in XConnect mode
    ${out}=      vpp_term: Show Interface Mode    agent_vpp_1
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    Should Contain     ${out}      l2 xconnect memif1/1 loop0
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    Should Contain     ${out}      l2 xconnect loop0 memif1/1

Add L2XConnect2 for Tap and Loopback2
    Put L2XConnect  agent_vpp_1    vpp1_tap1    vpp1_loop2
    Put L2XConnect  agent_vpp_1    vpp1_loop2     vpp1_tap1

Check L2XConnect2 and L2XConnect1 still configured
    ${out}=      vpp_term: Show Interface Mode    agent_vpp_1
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    Should Contain     ${out}      l2 xconnect memif1/1 loop0
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    Should Contain     ${out}      l2 xconnect loop0 memif1/1
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    Should Contain     ${out}      l2 xconnect tap0 loop1
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    Should Contain     ${out}      l2 xconnect loop1 tap0

Modify L2XConnect1
    Delete L2XConnect      agent_vpp_1    vpp1_memif1
    Put L2XConnect  agent_vpp_1    vpp1_vxlan1    vpp1_loop1
    Put L2XConnect  agent_vpp_1    vpp1_loop1     vpp1_vxlan1

Check L2XConnect1 Modified and L2XConnect2 still configured
    ${out}=      vpp_term: Show Interface Mode    agent_vpp_1
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    Should Contain     ${out}      l2 xconnect vxlan_tunnel0 loop0
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    Should Contain     ${out}      l2 xconnect loop0 vxlan_tunnel0
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    Should Contain     ${out}      l2 xconnect tap0 loop1
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    Should Contain     ${out}      l2 xconnect loop1 tap0
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    Should Contain     ${out}      l3 memif1/1

Delete L2XConnect1
    Delete L2XConnect      agent_vpp_1    vpp1_vxlan1
    Delete L2XConnect      agent_vpp_1    vpp1_loop1

Check L2XConnect1 Deleted and L2XConnect2 still configured
    ${out}=      vpp_term: Show Interface Mode    agent_vpp_1
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    Should Contain     ${out}      l3 memif1/1
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    Should Contain     ${out}      l3 loop0
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    Should Contain     ${out}      l3 vxlan_tunnel0
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    Should Contain     ${out}      l2 xconnect tap0 loop1
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    Should Contain     ${out}      l2 xconnect loop1 tap0


Delete L2XConnect2
    Delete L2XConnect      agent_vpp_1    vpp1_tap1
    Delete L2XConnect      agent_vpp_1    vpp1_loop2

Check L2XConnect1 and L2XConnect2 Deleted
    ${out}=      vpp_term: Show Interface Mode    agent_vpp_1
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    Should Contain     ${out}      l3 memif1/1
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    Should Contain     ${out}      l3 loop0
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    Should Contain     ${out}      l3 vxlan_tunnel0
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    Should Contain     ${out}      l3 tap0
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    Should Contain     ${out}      l3 loop1

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
    Execute In Container    agent_vpp_1    ip a
    vpp_term: Show Interface Mode    agent_vpp_1
    Make Datastore Snapshots    before_resync

TestSetup
    Make Datastore Snapshots    ${TEST_NAME}_test_setup

TestTeardown
    Make Datastore Snapshots    ${TEST_NAME}_test_teardown

Suite Cleanup
    Stop SFC Controller Container
    Testsuite Teardown