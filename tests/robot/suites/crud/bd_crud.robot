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

Resource     ../../libraries/bridge_domain/bridge_domain.robot
Resource     ../../libraries/interface/vxlan.robot
Resource     ../../libraries/interface/interface_generic.robot

Force Tags        crud     IPv4
Suite Setup       Testsuite Setup
Suite Teardown    Testsuite Teardown
Test Setup        TestSetup
Test Teardown     TestTeardown

*** Variables ***
${VARIABLES}=                    common
${ENV}=                          common
${WAIT_TIMEOUT}=                 20s
${SYNC_SLEEP}=                   3s
@{BD1_INTERFACES}=               vpp1_memif1    vpp1_vxlan1    vpp1_afpacket1
@{BD1_INTERFACES_UPDATED}=       vpp1_memif1    vpp1_vxlan1    bvi_vpp1_loop2
@{BD1_INTERFACES_VX_DELETED}=    vpp1_memif1    bvi_vpp1_loop2
@{BD2_INTERFACES}=               vpp1_memif2    vpp1_vxlan2    bvi_vpp1_loop3

*** Test Cases ***
Configure Environment
    [Tags]    setup
    Configure Environment 1

Show Interfaces Before Setup
    vpp_term: Show Interfaces    agent_vpp_1

Add Interfaces For BDs
    Put Memif Interface With IP    node=agent_vpp_1    name=vpp1_memif1    mac=62:61:61:61:61:61    master=true    id=1    ip=192.168.10.1
    Put Veth Interface With IP    node=agent_vpp_1    name=vpp1_veth1    mac=12:11:11:11:11:11    peer=vpp1_veth2    ip=10.10.1.1
    Put Veth Interface    node=agent_vpp_1    name=vpp1_veth2    mac=12:12:12:12:12:12    peer=vpp1_veth1
    Put Afpacket Interface    node=agent_vpp_1    name=vpp1_afpacket1    mac=a2:a1:a1:a1:a1:a1    host_int=vpp1_veth2
    Put VXLan Interface    node=agent_vpp_1    name=vpp1_vxlan1    src=192.168.1.1    dst=192.168.1.2    vni=5
    Put Loopback Interface With IP    node=agent_vpp_1    name=vpp1_loop1    mac=12:21:21:11:11:11    ip=20.20.1.1
    Put TAPv2 Interface With IP    node=agent_vpp_1    name=vpp1_tap1    mac=32:21:21:11:11:11    ip=30.30.1.1    host_if_name=linux_vpp1_tap1
    Put Memif Interface With IP    node=agent_vpp_1    name=vpp1_memif2    mac=62:61:61:61:61:62    master=true    id=2    ip=192.168.20.2
    Put VXLan Interface    node=agent_vpp_1    name=vpp1_vxlan2    src=192.168.2.1    dst=192.168.2.2    vni=15
    Put Loopback Interface With IP    node=agent_vpp_1    name=bvi_vpp1_loop2    mac=12:21:21:11:11:12    ip=20.20.2.1
    Put Loopback Interface With IP    node=agent_vpp_1    name=bvi_vpp1_loop3    mac=12:21:21:11:11:13    ip=20.20.3.1

Add BD1 Bridge Domain
    vpp_api: No Bridge Domains Exist    agent_vpp_1
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    Put Bridge Domain    node=agent_vpp_1    name=vpp1_bd1    ints=${BD1_INTERFACES}    flood=true    unicast=true    forward=true    learn=true    arp_term=true

Check BD1 Is Created
    vpp_api: BD Is Created    agent_vpp_1    vpp1_bd1
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    vpp_api: Check Bridge Domain State    agent_vpp_1  vpp1_bd1  flood=1  unicast=1  forward=1  learn=1  arp_term=1  interfaces=${BD1_INTERFACES}  bvi_int=none

Add BD2 Bridge Domain
    vpp_api: BD Not Exists    agent_vpp_1    vpp1_bd2
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    Put Bridge Domain    node=agent_vpp_1    name=vpp1_bd2    ints=${BD2_INTERFACES}    flood=true    unicast=true    forward=true    learn=true    arp_term=true

Check BD2 Is Created
    vpp_api: BD Is Created    agent_vpp_1    vpp1_bd2
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    vpp_api: Check Bridge Domain State    agent_vpp_1  vpp1_bd2  flood=1  unicast=1  forward=1  learn=1  arp_term=1  interfaces=${BD2_INTERFACES}  bvi_int=bvi_vpp1_loop3

Check That BD1 Is Not Affected By Adding BD2
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    vpp_api: Check Bridge Domain State    agent_vpp_1  vpp1_bd1  flood=1  unicast=1  forward=1  learn=1  arp_term=1  interfaces=${BD1_INTERFACES}  bvi_int=none

Update BD1
    Put Bridge Domain    node=agent_vpp_1    name=vpp1_bd1    ints=${BD1_INTERFACES_UPDATED}    flood=false    unicast=false    forward=false    learn=false    arp_term=false
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    vpp_api: Check Bridge Domain State    agent_vpp_1  vpp1_bd1  flood=0  unicast=0  forward=0  learn=0  arp_term=0  interfaces=${BD1_INTERFACES_UPDATED}  bvi_int=bvi_vpp1_loop2

Check That BD2 Is Not Affected By Updating BD1
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    vpp_api: Check Bridge Domain State    agent_vpp_1  vpp1_bd2  flood=1  unicast=1  forward=1  learn=1  arp_term=1  interfaces=${BD2_INTERFACES}  bvi_int=bvi_vpp1_loop3

Delete VXLan1 Interface
    Delete VPP Interface    node=agent_vpp_1    name=vpp1_vxlan1
    VXLan Tunnel Is Deleted    node=agent_vpp_1    src=192.168.1.1    dst=192.168.1.2    vni=5

Check That VXLan1 Interface Is Deleted From BD1
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    vpp_api: Check Bridge Domain State    agent_vpp_1  vpp1_bd1  flood=0  unicast=0  forward=0  learn=0  arp_term=0  interfaces=${BD1_INTERFACES_VX_DELETED}  bvi_int=bvi_vpp1_loop2

Readd VXLan1 Interface
    Put VXLan Interface    node=agent_vpp_1    name=vpp1_vxlan1    src=192.168.1.1    dst=192.168.1.2    vni=5
    VXLan Tunnel Is Created    node=agent_vpp_1    src=192.168.1.1    dst=192.168.1.2    vni=5

Check That VXLan1 Interface Is Added To BD1
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    vpp_api: Check Bridge Domain State    agent_vpp_1  vpp1_bd1  flood=0  unicast=0  forward=0  learn=0  arp_term=0  interfaces=${BD1_INTERFACES_UPDATED}  bvi_int=bvi_vpp1_loop2

Delete BD1 Bridge Domain
    Delete Bridge Domain    agent_vpp_1    vpp1_bd1
    vpp_api: BD Is Deleted    agent_vpp_1    vpp1_bd1

Check That BD2 Is Not Affected By Deleting BD1
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    vpp_api: Check Bridge Domain State    agent_vpp_1  vpp1_bd2  flood=1  unicast=1  forward=1  learn=1  arp_term=1  interfaces=${BD2_INTERFACES}  bvi_int=bvi_vpp1_loop3

Show Interfaces And Other Objects After Test
    vpp_term: Show Interfaces    agent_vpp_1
    vpp_term: Show Interfaces    agent_vpp_2
    Write To Machine    agent_vpp_1_term    show int addr
    Write To Machine    agent_vpp_2_term    show int addr
    Write To Machine    agent_vpp_1_term    show h
    Write To Machine    agent_vpp_2_term    show h
    Write To Machine    agent_vpp_1_term    show br
    Write To Machine    agent_vpp_2_term    show br
    Write To Machine    agent_vpp_1_term    show br 1 detail
    Write To Machine    agent_vpp_2_term    show br 1 detail
    Write To Machine    agent_vpp_1_term    show vxlan tunnel
    Write To Machine    agent_vpp_2_term    show vxlan tunnel
    Write To Machine    agent_vpp_1_term    show err
    Write To Machine    agent_vpp_2_term    show err
    vpp_api: Interfaces Dump    agent_vpp_1
    vat_term: Interfaces Dump    agent_vpp_2
    Execute In Container    agent_vpp_1    ip a
    Execute In Container    agent_vpp_2    ip a

*** Keywords ***
TestSetup
    Make Datastore Snapshots    ${TEST_NAME}_test_setup

TestTeardown
    Make Datastore Snapshots    ${TEST_NAME}_test_teardown

