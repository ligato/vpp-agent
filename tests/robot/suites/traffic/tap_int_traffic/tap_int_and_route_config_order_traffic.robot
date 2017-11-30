*** Settings ***
Library      OperatingSystem
#Library      RequestsLibrary
#Library      SSHLibrary      timeout=60s
#Library      String

Resource     ../../../variables/${VARIABLES}_variables.robot
Resource     ../../../libraries/all_libs.robot
Resource    ../../../libraries/pretty_keywords.robot

Suite Setup       Testsuite Setup
Suite Teardown    Testsuite Teardown

*** Variables ***
${VARIABLES}=               common
${ENV}=                     common
${NAME_VPP1_TAP1}=          vpp1_tap1
${NAME_VPP2_TAP1}=          vpp2_tap1
${MAC_VPP1_TAP1}=           12:21:21:11:11:11
${MAC_VPP2_TAP1}=           22:21:21:22:22:22
${IP_VPP1_TAP1}=            10.10.1.1
${IP_VPP2_TAP1}=            20.20.1.1
${IP_LINUX_VPP1_TAP1}=      10.10.1.2
${IP_LINUX_VPP2_TAP1}=      20.20.1.2
${IP_VPP1_TAP1_NETWORK}=    10.10.1.0
${IP_VPP2_TAP1_NETWORK}=    20.20.1.0
${NAME_VPP1_MEMIF1}=        vpp1_memif1
${NAME_VPP2_MEMIF1}=        vpp2_memif1
${MAC_VPP1_MEMIF1}=         13:21:21:11:11:11
${MAC_VPP2_MEMIF1}=         23:21:21:22:22:22
${IP_VPP1_MEMIF1}=          192.168.1.1
${IP_VPP2_MEMIF1}=          192.168.1.2
${PREFIX}=                  24
${UP_STATE}=                up
${SYNC_SLEEP}=         10s
# wait for resync vpps after restart
${RESYNC_WAIT}=        50s

*** Test Cases ***
Configure Environment
    [Tags]    setup
    Configure Environment 1

Show Interfaces Before Setup
    vpp_term: Show Interfaces    agent_vpp_1
    vpp_term: Show Interfaces    agent_vpp_2
    Write To Machine    vpp_agent_ctl    vpp-agent-ctl ${AGENT_VPP_ETCD_CONF_PATH} -ps

Add VPP1_memif1 Interface
    vpp_term: Interface Not Exists    node=agent_vpp_1    mac=${MAC_VPP1_MEMIF1}
    vpp_ctl: Put Memif Interface With IP    node=agent_vpp_1    name=${NAME_VPP1_MEMIF1}    mac=${MAC_VPP1_MEMIF1}    master=true    id=1    ip=${IP_VPP1_MEMIF1}    prefix=24    socket=memif.sock
    vpp_term: Interface Is Created    node=agent_vpp_1    mac=${MAC_VPP1_MEMIF1}

Add VPP2_memif1 Interface
    vpp_term: Interface Not Exists    node=agent_vpp_2    mac=${MAC_VPP2_MEMIF1}
    vpp_ctl: Put Memif Interface With IP    node=agent_vpp_2    name=${NAME_VPP2_MEMIF1}    mac=${MAC_VPP2_MEMIF1}    master=false    id=1    ip=${IP_VPP2_MEMIF1}    prefix=24    socket=memif.sock
    vpp_term: Interface Is Created    node=agent_vpp_1    mac=${MAC_VPP1_MEMIF1}

Check Ping From VPP1 To VPP2_memif1
    vpp_term: Check Ping    node=agent_vpp_1    ip=${IP_VPP2_MEMIF1}

Check Ping From VPP2 To VPP1_memif1
    vpp_term: Check Ping    node=agent_vpp_2    ip=${IP_VPP1_MEMIF1}

Add Static Route From VPP1 Linux To VPP2
    linux: Add Route    node=agent_vpp_1    destination_ip=${IP_VPP2_TAP1_NETWORK}    prefix=${PREFIX}    next_hop_ip=${IP_VPP1_TAP1}

Add Static Route From VPP1 To VPP2
    Create Route On agent_vpp_1 With IP 20.20.1.0/24 With Next Hop 192.168.1.2 And Vrf Id 0

Add Static Route From VPP2 Linux To VPP1
    linux: Add Route    node=agent_vpp_2    destination_ip=${IP_VPP1_TAP1_NETWORK}    prefix=${PREFIX}    next_hop_ip=${IP_VPP2_TAP1}

Add Static Route From VPP2 To VPP1
    Create Route On agent_vpp_2 With IP 10.10.1.0/24 With Next Hop 192.168.1.1 And Vrf Id 0

Show Routes On VPP1 Linux
    ${out}=    Execute In Container     agent_vpp_1    ip addr show
    Log    ${out}

Show Routes On VPP1
    ${out}=    vpp_term: Show Ip Fib    agent_vpp_1
    Log    ${out}

Show Routes On VPP2 Linux
    ${out}=    Execute In Container     agent_vpp_2    ip addr show
    Log    ${out}

Show Routes On VPP2
    ${out}=    vpp_term: Show Ip Fib    agent_vpp_2
    Log    ${out}

Add VPP1_TAP1 Interface
    vpp_term: Interface Not Exists  node=agent_vpp_1    mac=${MAC_VPP1_TAP1}
    vpp_ctl: Put TAP Interface With IP    node=agent_vpp_1    name=${NAME_VPP1_TAP1}    mac=${MAC_VPP1_TAP1}    ip=${IP_VPP1_TAP1}    prefix=${PREFIX}    host_if_name=linux_${NAME_VPP1_TAP1}
    linux: Set Host TAP Interface    node=agent_vpp_1    host_if_name=linux_${NAME_VPP1_TAP1}    ip=${IP_LINUX_VPP1_TAP1}    prefix=${PREFIX}

Check VPP1_TAP1 Interface Is Created
    ${interfaces}=       vat_term: Interfaces Dump    node=agent_vpp_1
    Log                  ${interfaces}
    vpp_term: Interface Is Created    node=agent_vpp_1    mac=${MAC_VPP1_TAP1}
    ${actual_state}=    vpp_term: Check TAP interface State    agent_vpp_1    ${NAME_VPP1_TAP1}    mac=${MAC_VPP1_TAP1}    ipv4=${IP_VPP1_TAP1}/${PREFIX}    state=${UP_STATE}

Check Ping Between VPP1 and linux_VPP1_TAP1 Interface
    linux: Check Ping    node=agent_vpp_1    ip=${IP_VPP1_TAP1}
    vpp_term: Check Ping    node=agent_vpp_1    ip=${IP_LINUX_VPP1_TAP1}

Add VPP2_TAP1 Interface
    vpp_term: Interface Not Exists  node=agent_vpp_2    mac=${MAC_VPP2_TAP1}
    vpp_ctl: Put TAP Interface With IP    node=agent_vpp_2    name=${NAME_VPP2_TAP1}    mac=${MAC_VPP2_TAP1}    ip=${IP_VPP2_TAP1}    prefix=${PREFIX}    host_if_name=linux_${NAME_VPP2_TAP1}
    linux: Set Host TAP Interface    node=agent_vpp_2    host_if_name=linux_${NAME_VPP2_TAP1}    ip=${IP_LINUX_VPP2_TAP1}    prefix=${PREFIX}

Check VPP2_TAP1 Interface Is Created
    ${interfaces}=       vat_term: Interfaces Dump    node=agent_vpp_1
    Log                  ${interfaces}
    vpp_term: Interface Is Created    node=agent_vpp_2    mac=${MAC_VPP2_TAP1}
    ${actual_state}=    vpp_term: Check TAP interface State    agent_vpp_2    ${NAME_VPP2_TAP1}    mac=${MAC_VPP2_TAP1}    ipv4=${IP_VPP2_TAP1}/${PREFIX}    state=${UP_STATE}

Check Ping Between VPP2 And linux_VPP2_TAP1 Interface
    linux: Check Ping    node=agent_vpp_2    ip=${IP_VPP2_TAP1}
    vpp_term: Check Ping    node=agent_vpp_2    ip=${IP_LINUX_VPP2_TAP1}

Check Ping From VPP1 Linux To VPP2_TAP1 And LINUX_VPP2_TAP1
    linux: Check Ping    node=agent_vpp_1    ip=${IP_VPP2_TAP1}
    linux: Check Ping    node=agent_vpp_1    ip=${IP_LINUX_VPP2_TAP1}

Check Ping From VPP2 Linux To VPP1_TAP1 And LINUX_VPP1_TAP1
    linux: Check Ping    node=agent_vpp_2    ip=${IP_VPP1_TAP1}
    linux: Check Ping    node=agent_vpp_2    ip=${IP_LINUX_VPP1_TAP1}


*** Keywords ***
Create Route
    [Arguments]    ${node}    ${routename}    ${ip}    ${next_hop}    ${prefix}=24    ${metric}=100    ${isdefault}=false
    Log Many    ${node}    ${namespace}    ${interface}    ${routename}    ${ip}    ${prefix}    ${next_hop}    ${metric}    ${isdefault}
    ${data}=              OperatingSystem.Get File      ${CURDIR}/../resources/linux_static_route.json
    ${uri}=               Set Variable                  /vnf-agent/${node}/linux/config/v1/route/${routename}
    Log Many              ${data}                       ${uri}
    ${data}=              Replace Variables             ${data}
    Log                   ${data}
    vpp_ctl: Put Json     ${uri}    ${data}


