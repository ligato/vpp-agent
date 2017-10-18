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

*** Test Cases ***
Configure Environment
    [Tags]    setup
    Configure Environment 1

Show Interfaces Before Setup
    vpp_term: Show Interfaces    agent_vpp_1

Add Interfaces For BDs
    vpp_ctl: Put Memif Interface With IP    node=agent_vpp_1    name=vpp1_memif1    mac=62:61:61:61:61:61    master=true    id=1    ip=192.168.1.1
    vpp_ctl: Put Veth Interface With IP    node=agent_vpp_1    name=vpp1_veth1    mac=12:11:11:11:11:11    peer=vpp1_veth2    ip=10.10.1.1
    vpp_ctl: Put Veth Interface    node=agent_vpp_1    name=vpp1_veth2    mac=12:12:12:12:12:12    peer=vpp1_veth1
    vpp_ctl: Put Afpacket Interface    node=agent_vpp_1    name=vpp1_afpacket1    mac=a2:a1:a1:a1:a1:a1    host_int=vpp1_veth2
    vpp_ctl: Put VXLan Interface    node=agent_vpp_1    name=vpp1_vxlan1    src=192.168.1.1    dst=192.168.1.2    vni=5
    vpp_ctl: Put Loopback Interface With IP    node=agent_vpp_1    name=vpp1_loop1    mac=12:21:21:11:11:11    ip=20.20.1.1
    vpp_ctl: Put TAP Interface With IP    node=agent_vpp_1    name=vpp1_tap1    mac=32:21:21:11:11:11    ip=30.30.1.1    host_if_name=linux_vpp1_tap1
    vpp_ctl: Put Memif Interface With IP    node=agent_vpp_1    name=vpp1_memif2    mac=62:61:61:61:61:62    master=true    id=2    ip=192.168.1.2
    vpp_ctl: Put VXLan Interface    node=agent_vpp_1    name=vpp1_vxlan2    src=192.168.2.1    dst=192.168.2.2    vni=15
    vpp_ctl: Put Loopback Interface With IP    node=agent_vpp_1    name=bvi_vpp1_loop2    mac=12:21:21:11:11:12    ip=20.20.2.1
    vpp_ctl: Put Loopback Interface With IP    node=agent_vpp_1    name=bvi_vpp1_loop3    mac=12:21:21:11:11:13    ip=20.20.3.1

Add BD1 Bridge Domain
    @{ints}=    Create List   vpp1_memif1  vpp1_vxlan1    vpp1_afpacket1
    vat_term: BD Not Exists    agent_vpp_1    @{ints}
    vpp_ctl: Put Bridge Domain    node=agent_vpp_1    name=vpp1_bd1    ints=${ints}    flood=true    unicast=true    forward=true    learn=true    arp_term=true

Check BD1 Is Created
    vat_term: BD Is Created    agent_vpp_1    vpp1_memif1    vpp1_afpacket1    vpp1_vxlan1
    vat_term: Check Bridge Domain State    agent_vpp_1  vpp1_bd1  flood=1  unicast=1  forward=1  learn=1  arp_term=1  interface=vpp1_memif1  interface=vpp1_afpacket1  interface=vpp1_vxlan1  bvi_int=none

Add BD2 Bridge Domain
    @{ints}=    Create List   vpp1_memif2  vpp1_vxlan2    bvi_vpp1_loop3
    vat_term: BD Not Exists    agent_vpp_1    @{ints}
    vpp_ctl: Put Bridge Domain    node=agent_vpp_1    name=vpp1_bd2    ints=${ints}    flood=true    unicast=true    forward=true    learn=true    arp_term=true

Check BD2 Is Created
    vat_term: BD Is Created    agent_vpp_1    vpp1_memif2    vpp1_vxlan2    bvi_vpp1_loop3
    vat_term: Check Bridge Domain State    agent_vpp_1  vpp1_bd2  flood=1  unicast=1  forward=1  learn=1  arp_term=1  interface=vpp1_memif2  interface=vpp1_vxlan2  interface=bvi_vpp1_loop3  bvi_int=bvi_vpp1_loop3

Check That BD1 Is Not Affected By Adding BD2
    vat_term: Check Bridge Domain State    agent_vpp_1  vpp1_bd1  flood=1  unicast=1  forward=1  learn=1  arp_term=1  interface=vpp1_memif1  interface=vpp1_afpacket1  interface=vpp1_vxlan1  bvi_int=none

Update BD1
    @{ints}=    Create List   vpp1_memif1  vpp1_vxlan1    bvi_vpp1_loop2
    vat_term: BD Not Exists    agent_vpp_1    @{ints}
    vpp_ctl: Put Bridge Domain    node=agent_vpp_1    name=vpp1_bd1    ints=${ints}    flood=false    unicast=false    forward=false    learn=false    arp_term=false
    vat_term: BD Is Deleted    agent_vpp_1    vpp1_memif1    vpp1_afpacket1    vpp1_vxlan1
    vat_term: BD Is Created    agent_vpp_1    vpp1_memif1    vpp1_vxlan1    bvi_vpp1_loop2
    vat_term: Check Bridge Domain State    agent_vpp_1  vpp1_bd1  flood=0  unicast=0  forward=0  learn=0  arp_term=0  interface=vpp1_memif1  interface=vpp1_vxlan1  interface=bvi_vpp1_loop2  bvi_int=bvi_vpp1_loop2

Check That BD2 Is Not Affected By Updating BD1
    vat_term: Check Bridge Domain State    agent_vpp_1  vpp1_bd2  flood=1  unicast=1  forward=1  learn=1  arp_term=1  interface=vpp1_memif2  interface=vpp1_vxlan2  interface=bvi_vpp1_loop3  bvi_int=bvi_vpp1_loop3

Delete VXLan1 Interface
    vpp_ctl: Delete VPP Interface    node=agent_vpp_1    name=vpp1_vxlan1
    vxlan: Tunnel Is Deleted    node=agent_vpp_1    src=192.168.1.1    dst=192.168.1.2    vni=5

Check That VXLan1 Interface Is Deleted From BD1
    vat_term: BD Is Deleted    agent_vpp_1    vpp1_memif1    vpp1_vxlan1    bvi_vpp1_loop2
    vat_term: Check Bridge Domain State    agent_vpp_1  vpp1_bd1  flood=0  unicast=0  forward=0  learn=0  arp_term=0  interface=vpp1_memif1  interface=bvi_vpp1_loop2  bvi_int=bvi_vpp1_loop2

Readd VXLan1 Interface
    vpp_ctl: Put VXLan Interface    node=agent_vpp_1    name=vpp1_vxlan1    src=192.168.1.1    dst=192.168.1.2    vni=5
    vxlan: Tunnel Is Created    node=agent_vpp_1    src=192.168.1.1    dst=192.168.1.2    vni=5

Check That VXLan1 Interface Is Added To BD1
    vat_term: Check Bridge Domain State    agent_vpp_1  vpp1_bd1  flood=0  unicast=0  forward=0  learn=0  arp_term=0  interface=vpp1_memif1  interface=vpp1_vxlan1  interface=bvi_vpp1_loop2  bvi_int=bvi_vpp1_loop2

Delete BD1 Bridge Domain
    vpp_ctl: Delete Bridge Domain    agent_vpp_1    vpp1_bd1
    vat_term: BD Is Deleted    agent_vpp_1    vpp1_memif1    vpp1_vxlan1    bvi_vpp1_loop2

Check That BD2 Is Not Affected By Deleting BD1
    vat_term: Check Bridge Domain State    agent_vpp_1  vpp1_bd2  flood=1  unicast=1  forward=1  learn=1  arp_term=1  interface=vpp1_memif2  interface=vpp1_vxlan2  interface=bvi_vpp1_loop3  bvi_int=bvi_vpp1_loop3

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
    vat_term: Interfaces Dump    agent_vpp_1
    vat_term: Interfaces Dump    agent_vpp_2
    Write To Machine    vpp_agent_ctl    vpp-agent-ctl ${AGENT_VPP_ETCD_CONF_PATH} -ps
    Execute In Container    agent_vpp_1    ip a
    Execute In Container    agent_vpp_2    ip a

*** Keywords ***
TestSetup
    Make Datastore Snapshots    ${TEST_NAME}_test_setup

TestTeardown
    Make Datastore Snapshots    ${TEST_NAME}_test_teardown

