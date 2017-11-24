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
${VARIABLES}=          common
${ENV}=                common
${SYNC_SLEEP}=         10s
*** Test Cases ***
Configure Environment
    [Tags]    setup
    Configure Environment 1

Show Interfaces Before Setup
    vpp_term: Show Interfaces    agent_vpp_1

Add Interfaces For ARP
    vpp_ctl: Put Memif Interface With IP    node=agent_vpp_1    name=vpp1_memif1    mac=62:61:61:61:61:61    master=true    id=1    ip=192.168.1.1
    vpp_ctl: Put Veth Interface With IP    node=agent_vpp_1    name=vpp1_veth1    mac=12:11:11:11:11:11    peer=vpp1_veth2    ip=10.10.1.1   prefix=24
    Sleep    1s
    vpp_ctl: Put Veth Interface    node=agent_vpp_1    name=vpp1_veth2    mac=12:12:12:12:12:12    peer=vpp1_veth1    enabled=true
    vpp_ctl: Put Afpacket Interface    node=agent_vpp_1    name=vpp1_afpacket1    mac=a2:a1:a1:a1:a1:a1    host_int=vpp1_veth2
    vpp_ctl: Put VXLan Interface    node=agent_vpp_1    name=vpp1_vxlan1    src=192.168.1.1    dst=192.168.1.2    vni=5
    vpp_ctl: Put Loopback Interface With IP    node=agent_vpp_1    name=vpp1_loop1    mac=12:21:21:11:11:11    ip=20.20.1.1   prefix=24   mtu=1400
    vpp_ctl: Put TAP Interface With IP    node=agent_vpp_1    name=vpp1_tap1    mac=32:21:21:11:11:11    ip=30.30.1.1   prefix=24      host_if_name=linux_vpp1_tap1
    Sleep    ${SYNC_SLEEP}

Check Memif Interface Created
    vpp_term: Interface Is Created    node=agent_vpp_1    mac=62:61:61:61:61:61
    vat_term: Check Memif Interface State     agent_vpp_1  vpp1_memif1  mac=62:61:61:61:61:61  role=master  id=1  ipv4=192.168.1.1/24  connected=0  enabled=1  socket=${AGENT_VPP_1_SOCKET_FOLDER}/default.sock

Check Veth Interface With IP Created
    linux: Interface Is Created    node=agent_vpp_1    mac=12:11:11:11:11:11
    linux: Check Veth Interface State     agent_vpp_1    vpp1_veth1    mac=12:11:11:11:11:11    ipv4=10.10.1.1/24    mtu=1500    state=up

Check Veth Interface Created
    linux: Interface Is Created    node=agent_vpp_1    mac=12:12:12:12:12:12
    linux: Check Veth Interface State     agent_vpp_1    vpp1_veth2    mac=12:12:12:12:12:12    state=up

Check AFpacket Interface Created
    vpp_term: Interface Is Created    node=agent_vpp_1    mac=a2:a1:a1:a1:a1:a1
    vat_term: Check Afpacket Interface State    agent_vpp_1    vpp1_afpacket1    enabled=1    mac=a2:a1:a1:a1:a1:a1

Check VXLan Interface Created
    vxlan: Tunnel Is Created    node=agent_vpp_1    src=192.168.1.1    dst=192.168.1.2    vni=5
    vat_term: Check VXLan Interface State    agent_vpp_1    vpp1_vxlan1    enabled=1    src=192.168.1.1    dst=192.168.1.2    vni=5

Check Loopback Interface Created
    vpp_term: Interface Is Created    node=agent_vpp_1    mac=12:21:21:11:11:11
    vat_term: Check Loopback Interface State    agent_vpp_1    vpp1_loop1    enabled=1     mac=12:21:21:11:11:11    mtu=1400  ipv4=20.20.1.1/24

Check TAP Interface Created
    vpp_term: Interface Is Created    node=agent_vpp_1    mac=32:21:21:11:11:11
    vpp_term: Check TAP interface State    agent_vpp_1    vpp1_tap1    mac=32:21:21:11:11:11    ipv4=30.30.1.1/24    state=up

Add ARP for memif
    vpp_ctl: Put ARP    agent_vpp_1    vpp1_memif1    155.155.155.155.    32:51:51:51:51:51    false

Check ARP is created
    vpp_term: Show ARP   agent_vpp_1

*** Keywords ***
TestSetup
    Make Datastore Snapshots    ${TEST_NAME}_test_setup

TestTeardown
    Make Datastore Snapshots    ${TEST_NAME}_test_teardown