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

Add ARPs
    vpp_ctl: Put Linux ARP    agent_vpp_1    vpp1_veth1  veth1_arp  155.155.155.155    32:51:51:51:51:51    false
    vpp_ctl: Put Linux ARP    agent_vpp_1    vpp1_veth2  veth2_arp  155.155.155.156    32:51:51:51:51:52    false
    vpp_ctl: Put Linux ARP    agent_vpp_1    lo          loopback_arp  155.155.155.156    32:51:51:51:51:52    false
    vpp_ctl: Put Linux ARP    agent_vpp_1    eth0        eth_arp  155.155.155.156    32:51:51:51:51:52    false
    Sleep    ${SYNC_SLEEP}
    Show Info

ADD Afpacket Interface
    vpp_ctl: Put Afpacket Interface    node=agent_vpp_1    name=vpp1_afpacket1    mac=a2:a1:a1:a1:a1:a1    host_int=vpp1_veth2

Check AFpacket Interface Created
    vpp_term: Interface Is Created    node=agent_vpp_1    mac=a2:a1:a1:a1:a1:a1
    vat_term: Check Afpacket Interface State    agent_vpp_1    vpp1_afpacket1    enabled=1    mac=a2:a1:a1:a1:a1:a1
    Sleep    ${SYNC_SLEEP}
    Show Info

Delete Afpacket1 Interface
    vpp_ctl: Delete VPP Interface    node=agent_vpp_1    name=vpp1_afpacket1
    vpp_term: Interface Is Deleted    node=agent_vpp_1    mac=${AFP1_MAC}
    Sleep    ${SYNC_SLEEP}
    Show Info

#Check  ARP
#    vpp_term: Check ARP   agent_vpp_1     vpp1_memif1    155.155.155.155    32:51:51:51:51:51    True
#    vpp_term: Check ARP   agent_vpp_1     vpp1_memif1    155.155.155.156    32:51:51:51:51:52    True
#
#
#Delete ARPs
#    vpp_ctl: Delete ARP    agent_vpp_1    vpp1_memif1    155.155.155.156
#    vpp_ctl: Delete ARP    agent_vpp_1    vpp1_veth1    155.155.155.150
#    Sleep    ${SYNC_SLEEP}
#
#Check Memif ARP After Delete
#    vpp_term: Check ARP   agent_vpp_1     vpp1_memif1    155.155.155.155    32:51:51:51:51:51    True
#    vpp_term: Check ARP   agent_vpp_1     vpp1_memif1    155.155.155.156    32:51:51:51:51:52    False
#
#
#Modify ARPs
#    vpp_ctl: Put ARP    agent_vpp_1    vpp1_memif1    155.155.155.155    32:51:51:51:51:5    false
#    vpp_term:Show ARP   agent_vpp_1
#    Sleep    ${SYNC_SLEEP}
#
#Check Memif ARP After Modify
#    vpp_term: Check ARP   agent_vpp_1     vpp1_memif1    155.155.155.155    32:51:51:51:51:51    True
#


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
