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
${SYNC_SLEEP}=         6s
${VETH1_MAC}=          1a:00:00:11:11:11
${VETH2_MAC}=          2a:00:00:22:22:22
${AFP1_MAC}=           a2:01:01:01:01:01
${NAMESPACE}=
${NSTYPE}=            3

*** Test Cases ***
Configure Environment
    [Tags]    setup
    Configure Environment 1

Show Interfaces Before Setup
    Show Info

Add Veth1 Interface
    linux: Interface Not Exists    node=agent_vpp_1    mac=${VETH1_MAC}
    vpp_ctl: Put Veth Interface With IP    node=agent_vpp_1    name=vpp1_veth1    mac=${VETH1_MAC}    peer=vpp1_veth2    ip=10.10.1.1    prefix=24    mtu=1500
    linux: Interface Not Exists    node=agent_vpp_1    mac=${VETH1_MAC}

Add Veth2 Interface
    linux: Interface Not Exists    node=agent_vpp_1    mac=${VETH2_MAC}
    vpp_ctl: Put Veth Interface    node=agent_vpp_1    name=vpp1_veth2    mac=${VETH2_MAC}    peer=vpp1_veth1
    Show Info

Check That Veth1 And Veth2 Interfaces Are Created
    linux: Interface Is Created    node=agent_vpp_1    mac=${VETH1_MAC}
    linux: Interface Is Created    node=agent_vpp_1    mac=${VETH2_MAC}
    linux: Check Veth Interface State     agent_vpp_1    vpp1_veth1    mac=${VETH1_MAC}    ipv4=10.10.1.1/24    mtu=1500    state=up
    linux: Check Veth Interface State     agent_vpp_1    vpp1_veth2    mac=${VETH2_MAC}    state=up


ADD Afpacket Interface
    vpp_ctl: Put Afpacket Interface    node=agent_vpp_1    name=vpp1_afpacket1    mac=a2:a1:a1:a1:a1:a1    host_int=vpp1_veth2

Check AFpacket Interface Created
    vpp_term: Interface Is Created    node=agent_vpp_1    mac=a2:a1:a1:a1:a1:a1
    vat_term: Check Afpacket Interface State    agent_vpp_1    vpp1_afpacket1    enabled=1    mac=a2:a1:a1:a1:a1:a1
    Sleep    ${SYNC_SLEEP}

Add ARPs
    vpp_ctl: Put Linux ARP    agent_vpp_1    vpp1_veth1  veth1_arp  155.155.155.155    32:51:51:51:51:51
    vpp_ctl: Put Linux ARP    agent_vpp_1    vpp1_veth2  veth2_arp  155.155.155.156    32:51:51:51:51:52
    vpp_ctl: Put Linux ARP    agent_vpp_1    lo          loopback_arp  155.155.155.156    32:51:51:51:51:52
    vpp_ctl: Put Linux ARP    agent_vpp_1    eth0        eth_arp  155.155.155.156    32:51:51:51:51:52
    Sleep    ${SYNC_SLEEP}

Check ARPSs
    ${out}=       Execute In Container    agent_vpp_1    ip neigh
    Log           ${out}
    Should Contain     ${out}    155.155.155.156 dev vpp1_veth2 lladdr 32:51:51:51:51:52 PERMANENT
    Should Contain     ${out}    155.155.155.155 dev vpp1_veth1 lladdr 32:51:51:51:51:51 PERMANENT
    Should Contain     ${out}    155.155.155.156 dev eth0 lladdr 32:51:51:51:51:52 PERMANENT
    Should Contain     ${out}    155.155.155.156 dev lo lladdr 32:51:51:51:51:52 PERMANENT

Change ARPs
    vpp_ctl: Put Linux ARP    agent_vpp_1    vpp1_veth1  veth1_arp  155.255.155.155    32:61:51:51:51:51
    vpp_ctl: Put Linux ARP    agent_vpp_1    vpp1_veth2  veth2_arp  155.255.155.156    32:61:51:51:51:52
    vpp_ctl: Put Linux ARP    agent_vpp_1    lo          loopback_arp  155.255.155.156    32:61:51:51:51:52
    vpp_ctl: Put Linux ARP    agent_vpp_1    eth0        eth_arp  155.255.155.156    32:61:51:51:51:52
    Sleep    ${SYNC_SLEEP}

Check ARPSs Again
    ${out}=       Execute In Container    agent_vpp_1    ip neigh
    Log           ${out}
    Should Contain     ${out}    155.255.155.156 dev vpp1_veth2 lladdr 32:61:51:51:51:52 PERMANENT
    Should Contain     ${out}    155.255.155.155 dev vpp1_veth1 lladdr 32:61:51:51:51:51 PERMANENT
    Should Contain     ${out}    155.255.155.156 dev eth0 lladdr 32:61:51:51:51:52 PERMANENT
    Should Contain     ${out}    155.255.155.156 dev lo lladdr 32:61:51:51:51:52 PERMANENT

Delete ARPs
    vpp_ctl: Delete Linux ARP    agent_vpp_1    veth1_arp
    vpp_ctl: Delete Linux ARP    agent_vpp_1    veth2_arp
    vpp_ctl: Delete Linux ARP    agent_vpp_1    loopback_arp
    vpp_ctl: Delete Linux ARP    agent_vpp_1    eth_arp

Check ARPSs After Delete
    ${out}=       Execute In Container    agent_vpp_1    ip neigh
    Log           ${out}
    Should Not Contain     ${out}    155.255.155.156 dev vpp1_veth2 lladdr 32:61:51:51:51:52 PERMANENT
    Should Not Contain     ${out}    155.255.155.155 dev vpp1_veth1 lladdr 32:61:51:51:51:51 PERMANENT
    Should Not Contain     ${out}    155.255.155.156 dev eth0 lladdr 32:61:51:51:51:52 PERMANENT
    Should Not Contain     ${out}    155.255.155.156 dev lo lladdr 32:61:51:51:51:52 PERMANENT


*** Keywords ***
TestSetup
    Make Datastore Snapshots    ${TEST_NAME}_test_setup

TestTeardown
    Make Datastore Snapshots    ${TEST_NAME}_test_teardown

Show Info
    Execute In Container    agent_vpp_1    ip a
    Execute In Container    agent_vpp_1    ip neigh
    vpp_term:Show ARP   agent_vpp_1
    vpp_term: Show Interfaces    agent_vpp_1
