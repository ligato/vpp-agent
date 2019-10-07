*** Settings ***
Library      OperatingSystem

Resource     ../../variables/${VARIABLES}_variables.robot

Resource     ../../libraries/vpp_api.robot
Resource     ../../libraries/vpp_term.robot
Resource     ../../libraries/docker.robot
Resource     ../../libraries/setup-teardown.robot
Resource     ../../libraries/configurations.robot
Resource     ../../libraries/etcdctl.robot
Resource     ../../libraries/linux.robot

Resource     ../../libraries/interface/vxlan.robot
Resource     ../../libraries/interface/loopback.robot
Resource     ../../libraries/interface/afpacket.robot
Resource     ../../libraries/interface/interface_generic.robot

Force Tags        crud     IPv6
Suite Setup       Testsuite Setup
Suite Teardown    Testsuite Teardown
Test Setup        TestSetup
Test Teardown     TestTeardown

*** Variables ***
${VARIABLES}=            common
${ENV}=                  common
${WAIT_TIMEOUT}=         20s
${SYNC_SLEEP}=           3s
${VETH1_MAC}=            1a:00:00:11:11:11
${VETH2_MAC}=            2a:00:00:22:22:22
${AFP1_MAC}=             a2:01:01:01:01:01
${VETH1_IP}=             fd30::1:e:0:0:1
${VETH2_IP}=             fd30::1:e:0:0:2
${PREFIX}=               64
${MEMIF_IP}=             fd31::1:1:0:0:1
${VXLAN_IP_SRC}=         fd31::1:1:0:0:1
${VXLAN_IP_DST}=         fd31::1:1:0:0:2
${LOOPBACK_IP}=          fd32::1:1:0:0:1
${TAP_IP}=               fd33::1:1:0:0:1
${ARP1_IP}=              ab:cd:12:34::
${ARP2_IP}=              ab:cd:12:35::
${ARP1_MAC}=             32:51:51:51:51:51
${ARP2_MAC}=             32:51:51:51:51:52
${ARP1_MAC_MODIFIED}=    32:51:51:51:51:53

*** Test Cases ***
Configure Environment
    [Tags]    setup
    Configure Environment 1

Show Interfaces Before Setup
    vpp_term: Show Interfaces    agent_vpp_1


Add Veth1 Interface Pair
    linux: Interface Not Exists    node=agent_vpp_1    mac=${VETH1_MAC}
    linux: Interface Not Exists    node=agent_vpp_1    mac=${VETH2_MAC}
    Put Veth Interface With IP    node=agent_vpp_1    name=vpp1_veth1
    ...    mac=${VETH1_MAC}    peer=vpp1_veth2    ip=${VETH1_IP}    prefix=${PREFIX}    mtu=1500
    Put Veth Interface With IP    node=agent_vpp_1    name=vpp1_veth2
    ...    mac=${VETH2_MAC}    peer=vpp1_veth1    ip=${VETH2_IP}    prefix=${PREFIX}    mtu=1500
    linux: Interface Is Created    node=agent_vpp_1    mac=${VETH1_MAC}
    linux: Interface Is Created    node=agent_vpp_1    mac=${VETH2_MAC}
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    linux: Check Veth Interface State     agent_vpp_1    vpp1_veth1
    ...    mac=${VETH1_MAC}    ipv6=${VETH1_IP}/${PREFIX}    mtu=1500    state=up
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    linux: Check Veth Interface State     agent_vpp_1    vpp1_veth2
    ...    mac=${VETH2_MAC}    ipv6=${VETH2_IP}/${PREFIX}    mtu=1500    state=up
    vpp_term: Show Interface Mode  agent_vpp_1
    vpp_term: Show Interface Mode  agent_vpp_2

Add Memif Interface
    Put Memif Interface With IP    node=agent_vpp_1    name=vpp1_memif1
    ...    mac=62:61:61:61:61:61    master=true    id=1    ip=${MEMIF_IP}
    ...    prefix=${PREFIX}    socket=default.sock
    vpp_term: Interface Is Created    node=agent_vpp_1    mac=62:61:61:61:61:61
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    vpp_term: Check Memif Interface State     agent_vpp_1  vpp1_memif1
    ...    mac=62:61:61:61:61:61  role=master  id=1  ipv6=${MEMIF_IP}/${PREFIX}
    ...    connected=0  enabled=1  socket=${AGENT_VPP_1_MEMIF_SOCKET_FOLDER}/default.sock

Add VXLan Interface
    Put VXLan Interface    node=agent_vpp_1    name=vpp1_vxlan1
    ...    src=${VXLAN_IP_SRC}    dst=${VXLAN_IP_DST}    vni=5
    VXLan Tunnel Is Created    node=agent_vpp_1    src=${VXLAN_IP_SRC}    dst=${VXLAN_IP_DST}    vni=5
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    vpp_api: Check VXLan Interface State    agent_vpp_1    vpp1_vxlan1
    ...    enabled=1    src=${VXLAN_IP_SRC}    dst=${VXLAN_IP_DST}    vni=5

Add Loopback Interface
    Put Loopback Interface With IP    node=agent_vpp_1    name=vpp1_loop1
    ...    mac=12:21:21:11:11:11    ip=${LOOPBACK_IP}   prefix=${PREFIX}   mtu=1400
    vpp_term: Interface Is Created    node=agent_vpp_1    mac=12:21:21:11:11:11
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    vpp_api: Check Loopback Interface State    agent_vpp_1    vpp1_loop1
    ...    enabled=1     mac=12:21:21:11:11:11    mtu=1400  ipv6=${LOOPBACK_IP}/${PREFIX}

Add Tap Interface
    Put TAPv2 Interface With IP    node=agent_vpp_1    name=vpp1_tap1
    ...    mac=32:21:21:11:11:11    ip=${TAP_IP}   prefix=${PREFIX}      host_if_name=linux_vpp1_tap1
    vpp_term: Interface Is Created    node=agent_vpp_1    mac=32:21:21:11:11:11
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    vpp_term: Check TAP IP6 interface State    agent_vpp_1    vpp1_tap1
    ...    mac=32:21:21:11:11:11    ipv6=${TAP_IP}/${PREFIX}    state=up

ADD Afpacket Interface
    Put Afpacket Interface    node=agent_vpp_1    name=vpp1_afpacket1
    ...    mac=a2:a1:a1:a1:a1:a1    host_int=vpp1_veth2
    vpp_term: Interface Is Created    node=agent_vpp_1    mac=a2:a1:a1:a1:a1:a1
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    vpp_api: Check Afpacket Interface State    agent_vpp_1    vpp1_afpacket1
    ...    enabled=1    mac=a2:a1:a1:a1:a1:a1

Check Stuff
    Show Interfaces And Other Objects

Add ARPs
    Put ARP          agent_vpp_1    vpp1_memif1       ${ARP1_IP}    ${ARP1_MAC}    false
    Put ARP          agent_vpp_1    vpp1_memif1       ${ARP2_IP}    ${ARP2_MAC}    false
    Put Linux ARP    agent_vpp_1    vpp1_veth1        ${ARP1_IP}    ${ARP1_MAC}
    Put Linux ARP    agent_vpp_1    vpp1_veth1        ${ARP2_IP}    ${ARP2_MAC}
    Put Linux ARP    agent_vpp_1    vpp1_veth2        ${ARP1_IP}    ${ARP1_MAC}
    Put Linux ARP    agent_vpp_1    vpp1_veth2        ${ARP2_IP}    ${ARP2_MAC}
    Put ARP          agent_vpp_1    vpp1_vxlan1       ${ARP1_IP}    ${ARP1_MAC}    false
    Put ARP          agent_vpp_1    vpp1_vxlan1       ${ARP2_IP}    ${ARP2_MAC}    false
    Put ARP          agent_vpp_1    vpp1_loop1        ${ARP1_IP}    ${ARP1_MAC}    false
    Put ARP          agent_vpp_1    vpp1_loop1        ${ARP2_IP}    ${ARP2_MAC}    false
    Put ARP          agent_vpp_1    vpp1_tap1         ${ARP1_IP}    ${ARP1_MAC}    false
    Put ARP          agent_vpp_1    vpp1_tap1         ${ARP2_IP}    ${ARP2_MAC}    false
    Put ARP          agent_vpp_1    vpp1_afpacket1    ${ARP1_IP}    ${ARP1_MAC}    false
    Put ARP          agent_vpp_1    vpp1_afpacket1    ${ARP2_IP}    ${ARP2_MAC}    false
    Sleep            ${SYNC_SLEEP}

Check Memif ARP
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    vpp_term: Check IPv6 Neighbor   agent_vpp_1     vpp1_memif1
    ...    ${ARP1_IP}    ${ARP1_MAC}    True
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    vpp_term: Check IPv6 Neighbor   agent_vpp_1     vpp1_memif1
    ...    ${ARP2_IP}    ${ARP2_MAC}    True

Check Veth1 ARP
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    linux: Check IPv6 Neighbor    agent_vpp_1    vpp1_veth1
    ...    ${ARP1_IP}    ${ARP1_MAC}    True
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    linux: Check IPv6 Neighbor    agent_vpp_1    vpp1_veth1
    ...    ${ARP2_IP}    ${ARP2_MAC}    True

Check Veth2 ARP
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    linux: Check IPv6 Neighbor    agent_vpp_1    vpp1_veth2
    ...    ${ARP1_IP}    ${ARP1_MAC}    True
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    linux: Check IPv6 Neighbor    agent_vpp_1    vpp1_veth2
    ...    ${ARP2_IP}    ${ARP2_MAC}    True

Check VXLan ARP
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    vpp_term: Check IPv6 Neighbor    agent_vpp_1    vpp1_vxlan1
    ...    ${ARP1_IP}    ${ARP1_MAC}    True
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    vpp_term: Check IPv6 Neighbor    agent_vpp_1    vpp1_vxlan1
    ...    ${ARP2_IP}    ${ARP2_MAC}    True

Check Loopback ARP
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    vpp_term: Check IPv6 Neighbor    agent_vpp_1    vpp1_loop1
    ...    ${ARP1_IP}   ${ARP1_MAC}    True
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    vpp_term: Check IPv6 Neighbor    agent_vpp_1    vpp1_loop1
    ...    ${ARP2_IP}   ${ARP2_MAC}    True

Check TAP ARP
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    vpp_term: Check IPv6 Neighbor    agent_vpp_1    vpp1_tap1
    ...    ${ARP1_IP}   ${ARP1_MAC}    True
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    vpp_term: Check IPv6 Neighbor    agent_vpp_1    vpp1_tap1
    ...    ${ARP2_IP}   ${ARP2_MAC}    True

Check Afpacket ARP
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    vpp_term: Check IPv6 Neighbor    agent_vpp_1    vpp1_afpacket1
    ...    ${ARP1_IP}   ${ARP1_MAC}    True
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    vpp_term: Check IPv6 Neighbor    agent_vpp_1    vpp1_afpacket1
    ...    ${ARP2_IP}   ${ARP2_MAC}    True

Modify ARPs
    Put ARP          agent_vpp_1    vpp1_memif1        ${ARP1_IP}    ${ARP1_MAC_MODIFIED}    false
    Put Linux ARP    agent_vpp_1    vpp1_veth1         ${ARP1_IP}    ${ARP1_MAC_MODIFIED}
    Put Linux ARP    agent_vpp_1    vpp1_veth2         ${ARP1_IP}    ${ARP1_MAC_MODIFIED}
    Put ARP          agent_vpp_1    vpp1_vxlan1        ${ARP1_IP}    ${ARP1_MAC_MODIFIED}    false
    Put ARP          agent_vpp_1    vpp1_loop1         ${ARP1_IP}    ${ARP1_MAC_MODIFIED}    false
    Put ARP          agent_vpp_1    vpp1_tap1          ${ARP1_IP}    ${ARP1_MAC_MODIFIED}    false
    Put ARP          agent_vpp_1    host-vpp1_veth2    ${ARP1_IP}    ${ARP1_MAC_MODIFIED}    false
    Put ARP          agent_vpp_1    vpp1_afpacket1     ${ARP1_IP}    ${ARP1_MAC_MODIFIED}    false
    Sleep            ${SYNC_SLEEP}

Check Memif ARP After Modify
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    vpp_term: Check IPv6 Neighbor   agent_vpp_1     vpp1_memif1
    ...    ${ARP1_IP}    ${ARP1_MAC_MODIFIED}    True
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    vpp_term: Check IPv6 Neighbor   agent_vpp_1     vpp1_memif1
    ...    ${ARP1_IP}    ${ARP1_MAC}    False

Check Veth1 ARP After Modify
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    linux: Check IPv6 Neighbor    agent_vpp_1    vpp1_veth1
    ...    ${ARP1_IP}    ${ARP1_MAC_MODIFIED}    True
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    linux: Check IPv6 Neighbor    agent_vpp_1    vpp1_veth1
    ...    ${ARP1_IP}    ${ARP1_MAC}    False

Check Veth2 ARP After Modify
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    linux: Check IPv6 Neighbor    agent_vpp_1    vpp1_veth2
    ...    ${ARP1_IP}    ${ARP1_MAC_MODIFIED}    True
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    linux: Check IPv6 Neighbor    agent_vpp_1    vpp1_veth2
    ...    ${ARP1_IP}    ${ARP1_MAC}    False

Check VXLan ARP After Modify
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    vpp_term: Check IPv6 Neighbor    agent_vpp_1    vpp1_vxlan1
    ...    ${ARP1_IP}    ${ARP1_MAC_MODIFIED}    True
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    linux: Check IPv6 Neighbor    agent_vpp_1    vpp1_veth1
    ...    ${ARP1_IP}    ${ARP1_MAC}    False

Check Loopback ARP After Modify
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    vpp_term: Check IPv6 Neighbor    agent_vpp_1    vpp1_loop1
    ...    ${ARP1_IP}    ${ARP1_MAC_MODIFIED}    True
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    linux: Check IPv6 Neighbor    agent_vpp_1    vpp1_veth1
    ...    ${ARP1_IP}    ${ARP1_MAC}    False

Check TAP ARP After Modify
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    vpp_term: Check IPv6 Neighbor    agent_vpp_1    vpp1_tap1
    ...    ${ARP1_IP}    ${ARP1_MAC_MODIFIED}    True
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    linux: Check IPv6 Neighbor    agent_vpp_1    vpp1_veth1
    ...    ${ARP1_IP}    ${ARP1_MAC}    False

Check Afpacket ARP After Modify
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    vpp_term: Check IPv6 Neighbor    agent_vpp_1    vpp1_afpacket1
    ...    ${ARP1_IP}    ${ARP1_MAC_MODIFIED}    True
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    linux: Check IPv6 Neighbor    agent_vpp_1    vpp1_veth1
    ...    ${ARP1_IP}    ${ARP1_MAC}    False

Delete ARPs
    Delete ARP          agent_vpp_1    vpp1_memif1        ${ARP2_IP}
    Delete Linux ARP    agent_vpp_1    vpp1_veth1         ${ARP2_IP}
    Delete Linux ARP    agent_vpp_1    vpp1_veth2         ${ARP2_IP}
    Delete ARP          agent_vpp_1    vpp1_vxlan1        ${ARP2_IP}
    Delete ARP          agent_vpp_1    vpp1_loop1         ${ARP2_IP}
    Delete ARP          agent_vpp_1    vpp1_tap1          ${ARP2_IP}
    Delete ARP          agent_vpp_1    host-vpp1_veth2    ${ARP2_IP}
    Delete ARP          agent_vpp_1    vpp1_afpacket1     ${ARP2_IP}
    vpp_term:Show ARP   agent_vpp_1
    Execute In Container    agent_vpp_1    ip neigh
    Sleep    ${SYNC_SLEEP}

Check Memif ARP After Delete
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    vpp_term: Check IPv6 Neighbor   agent_vpp_1     vpp1_memif1
    ...    ${ARP1_IP}    ${ARP1_MAC_MODIFIED}    True
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    vpp_term: Check IPv6 Neighbor   agent_vpp_1     vpp1_memif1
    ...    ${ARP2_IP}    ${ARP2_MAC}    False

Check Veth1 ARP After Delete
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    linux: Check IPv6 Neighbor    agent_vpp_1    vpp1_veth1
    ...    ${ARP1_IP}    ${ARP1_MAC_MODIFIED}    True
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    linux: Check IPv6 Neighbor    agent_vpp_1    vpp1_veth1
    ...    ${ARP1_IP}    ${ARP2_MAC}    False

Check Veth2 ARP After Delete
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    linux: Check IPv6 Neighbor    agent_vpp_1    vpp1_veth2
    ...    ${ARP1_IP}    ${ARP1_MAC_MODIFIED}    True
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    linux: Check IPv6 Neighbor    agent_vpp_1    vpp1_veth2
    ...    ${ARP1_IP}    ${ARP2_MAC}    False

Check VXLan ARP After Delete
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    vpp_term: Check IPv6 Neighbor    agent_vpp_1    vpp1_vxlan1
    ...    ${ARP1_IP}    ${ARP1_MAC_MODIFIED}    True
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    vpp_term: Check IPv6 Neighbor    agent_vpp_1    vpp1_vxlan1
    ...    ${ARP2_IP}    ${ARP2_MAC}    False

Check Loopback ARP After Delete
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    vpp_term: Check IPv6 Neighbor    agent_vpp_1    vpp1_loop1
    ...    ${ARP1_IP}   ${ARP1_MAC_MODIFIED}    True
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    vpp_term: Check IPv6 Neighbor    agent_vpp_1    vpp1_loop1
    ...    ${ARP2_IP}   ${ARP2_MAC}    False

Check TAP ARP After Delete
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    vpp_term: Check IPv6 Neighbor    agent_vpp_1    vpp1_tap1
    ...    ${ARP1_IP}   ${ARP1_MAC_MODIFIED}    True
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    vpp_term: Check IPv6 Neighbor    agent_vpp_1    vpp1_tap1
    ...    ${ARP2_IP}   ${ARP2_MAC}    False

Check Afpacket ARP After Delete
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    vpp_term: Check IPv6 Neighbor    agent_vpp_1    vpp1_afpacket1
    ...    ${ARP1_IP}   ${ARP1_MAC_MODIFIED}    True
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    vpp_term: Check IPv6 Neighbor    agent_vpp_1    vpp1_afpacket1
    ...    ${ARP2_IP}   ${ARP2_MAC}    False


*** Keywords ***
Show Interfaces And Other Objects
    vpp_term: Show Interfaces    agent_vpp_1
    Write To Machine    agent_vpp_1_term    show int addr
    Write To Machine    agent_vpp_1_term    show h
    Write To Machine    agent_vpp_1_term    show br
    Write To Machine    agent_vpp_1_term    show err
    vpp_api: Interfaces Dump    agent_vpp_1
    Execute In Container    agent_vpp_1    ip a
    Make Datastore Snapshots    before_check stuff


TestSetup
    Make Datastore Snapshots    ${TEST_NAME}_test_setup

TestTeardown
    Make Datastore Snapshots    ${TEST_NAME}_test_teardown
