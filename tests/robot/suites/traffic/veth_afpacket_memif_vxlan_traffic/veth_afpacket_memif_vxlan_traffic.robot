*** Settings ***
Library      OperatingSystem
#Library      RequestsLibrary
#Library      SSHLibrary      timeout=60s
#Library      String

Resource     ../../../variables/${VARIABLES}_variables.robot

Resource     ../../../libraries/all_libs.robot

Suite Setup       Testsuite Setup
Suite Teardown    Testsuite Teardown

*** Variables ***
${VARIABLES}=          common
${ENV}=                common
${CONFIG_SLEEP}=       1s
${RESYNC_SLEEP}=       1s
# wait for resync vpps after restart
${RESYNC_WAIT}=        30s

*** Test Cases ***
Configure Environment
    [Tags]    setup
    Add Agent VPP Node    agent_vpp_1
    Add Agent VPP Node    agent_vpp_2

Show Interfaces Before Setup
    vpp_term: Show Interfaces    agent_vpp_1
    vpp_term: Show Interfaces    agent_vpp_2
    Write To Machine    vpp_agent_ctl    vpp-agent-ctl ${AGENT_VPP_ETCD_CONF_PATH} -ps

Setup Interfaces
    vpp_ctl: Put Memif Interface With IP    node=agent_vpp_1    name=vpp1_memif1    mac=62:61:61:61:61:61    master=true    id=1    ip=192.168.1.1
    vpp_ctl: Put Veth Interface With IP    node=agent_vpp_1    name=vpp1_veth1    mac=12:11:11:11:11:11    peer=vpp1_veth2    ip=10.10.1.1
    vpp_ctl: Put Veth Interface    node=agent_vpp_1    name=vpp1_veth2    mac=12:12:12:12:12:12    peer=vpp1_veth1
    vpp_ctl: Put Afpacket Interface    node=agent_vpp_1    name=vpp1_afpacket1    mac=a2:a1:a1:a1:a1:a1    host_int=vpp1_veth2
    vpp_ctl: Put VXLan Interface    node=agent_vpp_1    name=vpp1_vxlan1    src=192.168.1.1    dst=192.168.1.2    vni=5
    @{ints}=    Create List    vpp1_vxlan1    vpp1_afpacket1
    vpp_ctl: Put Bridge Domain    node=agent_vpp_1    name=vpp1_bd1    ints=${ints}
    vpp_ctl: Put Loopback Interface With IP    node=agent_vpp_1    name=vpp1_loop1    mac=12:21:21:11:11:11    ip=20.20.1.1
    vpp_ctl: Put TAP Interface With IP    node=agent_vpp_1    name=vpp1_tap1    mac=32:21:21:11:11:11    ip=30.30.1.1    host_if_name=linux_vpp1_tap1

    vpp_ctl: Put Memif Interface With IP    node=agent_vpp_2    name=vpp2_memif1    mac=62:62:62:62:62:62    master=false    id=1    ip=192.168.1.2
    vpp_ctl: Put Veth Interface With IP    node=agent_vpp_2    name=vpp2_veth1    mac=22:21:21:21:21:21    peer=vpp2_veth2    ip=10.10.1.2
    vpp_ctl: Put Veth Interface    node=agent_vpp_2    name=vpp2_veth2    mac=22:22:22:22:22:22    peer=vpp2_veth1
    vpp_ctl: Put Afpacket Interface    node=agent_vpp_2    name=vpp2_afpacket1    mac=a2:a2:a2:a2:a2:a2    host_int=vpp2_veth2
    vpp_ctl: Put VXLan Interface    node=agent_vpp_2    name=vpp2_vxlan1    src=192.168.1.2    dst=192.168.1.1    vni=5
    @{ints}=    Create List    vpp2_vxlan1    vpp2_afpacket1
    vpp_ctl: Put Bridge Domain    node=agent_vpp_2    name=vpp2_bd1    ints=${ints}
    vpp_ctl: Put Loopback Interface With IP    node=agent_vpp_2    name=vpp2_loop1    mac=22:21:21:11:11:11    ip=20.20.1.2
    vpp_ctl: Put TAP Interface With IP    node=agent_vpp_2    name=vpp2_tap1    mac=32:22:22:11:11:11    ip=30.30.1.2    host_if_name=linux_vpp2_tap1
 
Check Linux Interfaces On VPP1
    ${out}=    Execute In Container    agent_vpp_1    ip a
    Log    ${out}
    Should Contain    ${out}    vpp1_veth2@vpp1_veth1
    Should Contain    ${out}    vpp1_veth1@vpp1_veth2
    Should Contain    ${out}    linux_vpp1_tap1

Check Interfaces On VPP1
    ${out}=    vpp_term: Show Interfaces    agent_vpp_1
    Log    ${out}
    ${int}=    vpp_ctl: Get Interface Internal Name    agent_vpp_1    vpp1_memif1
    Should Contain    ${out}    ${int}
    ${int}=    vpp_ctl: Get Interface Internal Name    agent_vpp_1    vpp1_afpacket1
    Should Contain    ${out}    ${int}
    ${int}=    vpp_ctl: Get Interface Internal Name    agent_vpp_1    vpp1_vxlan1
    Should Contain    ${out}    ${int}
    ${int}=    vpp_ctl: Get Interface Internal Name    agent_vpp_1    vpp1_loop1
    Should Contain    ${out}    ${int}
    ${int}=    vpp_ctl: Get Interface Internal Name    agent_vpp_1    vpp1_tap1
    Should Contain    ${out}    ${int}

Check Linux Interfaces On VPP2
    ${out}=    Execute In Container    agent_vpp_2    ip a
    Log    ${out}
    Should Contain    ${out}    vpp2_veth2@vpp2_veth1
    Should Contain    ${out}    vpp2_veth1@vpp2_veth2
    Should Contain    ${out}    linux_vpp2_tap1            

Check Interfaces On VPP2
    ${out}=    vpp_term: Show Interfaces    agent_vpp_2
    Log    ${out}
    ${int}=    vpp_ctl: Get Interface Internal Name    agent_vpp_2    vpp2_memif1
    Should Contain    ${out}    ${int}
    ${int}=    vpp_ctl: Get Interface Internal Name    agent_vpp_2    vpp2_afpacket1
    Should Contain    ${out}    ${int}
    ${int}=    vpp_ctl: Get Interface Internal Name    agent_vpp_2    vpp2_vxlan1
    Should Contain    ${out}    ${int}
    ${int}=    vpp_ctl: Get Interface Internal Name    agent_vpp_2    vpp2_loop1
    Should Contain    ${out}    ${int}
    ${int}=    vpp_ctl: Get Interface Internal Name    agent_vpp_2    vpp2_tap1
    Should Contain    ${out}    ${int}

Show Interfaces And Other Objects After Config
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

Check Ping From VPP1 to VPP2
    linux: Check Ping    agent_vpp_1    10.10.1.2

Check Ping From VPP2 to VPP1
    linux: Check Ping    agent_vpp_2    10.10.1.1

Config Done
    No Operation

Final Sleep After Config For Manual Checking
    Sleep   ${CONFIG_SLEEP}

Remove VPP Nodes
    Remove All Nodes

Start VPP1 And VPP2 Again
    Add Agent VPP Node    agent_vpp_1
    Add Agent VPP Node    agent_vpp_2
    Sleep    ${RESYNC_WAIT}

Show Interfaces And Other Objects After Resync
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

Check Ping After Resync From VPP1 to VPP2
    linux: Check Ping    agent_vpp_1    10.10.1.2

Check Ping After Resync From VPP2 to VPP1
    linux: Check Ping    agent_vpp_2    10.10.1.1

Resync Done
    No Operation

Final Sleep After Resync For Manual Checking
    Sleep   ${RESYNC_SLEEP}


*** Keywords ***
