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
${VARIABLES}=       common
${ENV}=             common
${NAME_LOOP1}=      vpp1_loop1
${NAME_LOOP2}=      vpp1_loop2
${MAC_LOOP1}=       12:21:21:11:11:11
${MAC_LOOP1_2}=     22:21:21:11:11:11
${MAC_LOOP2}=       32:21:21:11:11:11
${IP_LOOP1}=        fd30:0:0:1:e::1
${IP_LOOP1_2}=      fd31:0:0:1:e::1
${IP_LOOP2}=        fd30:0:0:1:e::2
${PREFIX}=          64
${MTU}=             4800

*** Test Cases ***
Configure Environment
    [Tags]    setup
    Configure Environment 1

Show Interfaces Before Setup
    vpp_term: Show Interfaces    agent_vpp_1

Add Loopback1 Interface
    vpp_term: Interface Not Exists  node=agent_vpp_1    mac=${MAC_LOOP1}
    vpp_ctl: Put Loopback Interface With IP    node=agent_vpp_1    name=${NAME_LOOP1}    mac=${MAC_LOOP1}    ip=${IP_LOOP1}    prefix=${PREFIX}    mtu=${MTU}    enabled=true

Check Loopback1 Is Created
    vpp_term: Interface Is Created    node=agent_vpp_1    mac=${MAC_LOOP1}
    vat_term: Check Loopback Interface State    agent_vpp_1    ${NAME_LOOP1}    enabled=1     mac=${MAC_LOOP1}    mtu=${MTU}  ipv6=${IP_LOOP1}/${PREFIX}

Add Loopback2 Interface
    vpp_term: Interface Not Exists  node=agent_vpp_1    mac=${MAC_LOOP2}
    vpp_ctl: Put Loopback Interface With IP    node=agent_vpp_1     name=${NAME_LOOP2}    mac=${MAC_LOOP2}    ip=${IP_LOOP2}    prefix=${PREFIX}    mtu=${MTU}    enabled=true

Check Loopback2 Is Created
    vpp_term: Interface Is Created    node=agent_vpp_1    mac=${MAC_LOOP2}
    vat_term: Check Loopback Interface State    agent_vpp_1    ${NAME_LOOP2}    enabled=1     mac=${MAC_LOOP2}    mtu=${MTU}    ipv6=${IP_LOOP2}/${PREFIX}

Check Loopback1 Is Still Configured
    vat_term: Check Loopback Interface State    agent_vpp_1    ${NAME_LOOP1}    enabled=1     mac=${MAC_LOOP1}    mtu=${MTU}         ipv6=${IP_LOOP1}/${PREFIX}

Update Loopback1
    vpp_ctl: Put Loopback Interface With IP    node=agent_vpp_1     name=${NAME_LOOP1}    mac=${MAC_LOOP1_2}    ip=${IP_LOOP1_2}    prefix=${PREFIX}    mtu=${MTU}    enabled=true
    vpp_term: Interface Is Deleted    node=agent_vpp_1    mac=${MAC_LOOP1}
    vpp_term: Interface Is Created    node=agent_vpp_1    mac=${MAC_LOOP1_2}
    vat_term: Check Loopback Interface State    agent_vpp_1    ${NAME_LOOP1}    enabled=1     mac=${MAC_LOOP1_2}    mtu=${MTU}    ipv6=${IP_LOOP1_2}/${PREFIX}

Check Loopback2 Is Not Changed
    vat_term: Check Loopback Interface State    agent_vpp_1    ${NAME_LOOP2}    enabled=1     mac=${MAC_LOOP2}    mtu=${MTU}         ipv6=${IP_LOOP2}/${PREFIX}

Delete Loopback1_2 Interface
    vpp_ctl: Delete VPP Interface    node=agent_vpp_1    name=${NAME_LOOP1}
    vpp_term: Interface Is Deleted    node=agent_vpp_1    mac=${MAC_LOOP1_2}

Check Loopback2 Interface Is Still Configured
    vat_term: Check Loopback Interface State    agent_vpp_1    ${NAME_LOOP2}    enabled=1     mac=${MAC_LOOP2}    mtu=${MTU}         ipv6=${IP_LOOP2}/${PREFIX}

Show Interfaces And Other Objects After Setup
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

*** Keywords ***
TestSetup
    Make Datastore Snapshots    ${TEST_NAME}_test_setup

TestTeardown
    Make Datastore Snapshots    ${TEST_NAME}_test_teardown

