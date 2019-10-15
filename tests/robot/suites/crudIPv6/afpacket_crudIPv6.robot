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

Resource     ../../libraries/interface/afpacket.robot
Resource     ../../libraries/interface/interface_generic.robot

Force Tags        crud     IPv6
Suite Setup       Testsuite Setup
Suite Teardown    Testsuite Teardown
Test Setup        TestSetup
Test Teardown     TestTeardown

*** Variables ***
${VARIABLES}=          common
${ENV}=                common
${VETH1_MAC}=          1a:00:00:11:11:11
${VETH2_MAC}=          2a:00:00:22:22:22
${VETH3_MAC}=          3a:00:00:33:33:33
${VETH4_MAC}=          4a:00:00:44:44:44
${AFP1_MAC}=           a2:01:01:01:01:01
${AFP2_MAC}=           a2:02:02:02:02:02
${AFP2_SEC_MAC}=       a2:22:22:22:22:22
${IP_ADR}=             fd30:0:0:1:e::
${IP_ADR_MASK}=        fd30:0:0:1:e::/64
${IP_ADR2}=            fd30:0:0:2:f::
${IP_ADR_MASK2}=       fd30:0:0:2:f::/64
${WAIT_TIMEOUT}=       20s
${SYNC_SLEEP}=         3s

*** Test Cases ***
Configure Environment
    [Tags]    setup
    Configure Environment 1

Show Interfaces Before Setup
    vpp_term: Show Interfaces    agent_vpp_1

Add Veth1 And Veth2 Interfaces
    Put Veth Interface With IP    node=agent_vpp_1
    ...    name=vpp1_veth1    mac=${VETH1_MAC}    peer=vpp1_veth2
    ...    ip=${IP_ADR}    prefix=64    mtu=1500
    Put Veth Interface    node=agent_vpp_1
    ...    name=vpp1_veth2    mac=${VETH2_MAC}    peer=vpp1_veth1

Add Afpacket1 Interface
    vpp_term: Interface Not Exists    node=agent_vpp_1    mac=${AFP1_MAC}
    Put Afpacket Interface    node=agent_vpp_1
    ...    name=vpp1_afpacket1    mac=${AFP1_MAC}    host_int=vpp1_veth2

Check That Afpacket1 Interface Is Created
    vpp_term: Interface Is Created    node=agent_vpp_1    mac=${AFP1_MAC}
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    vpp_api: Check Afpacket Interface State    agent_vpp_1
    ...    vpp1_afpacket1    enabled=1    mac=${AFP1_MAC}

Check That Veth1 And Veth2 Interfaces Are Created And Not Affected By Afpacket1 Interface
    linux: Interface Is Created    node=agent_vpp_1    mac=${VETH1_MAC}
    linux: Interface Is Created    node=agent_vpp_1    mac=${VETH2_MAC}
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    linux: Check Veth Interface State     agent_vpp_1
    ...    vpp1_veth1    mac=${VETH1_MAC}
    ...    ipv6=${IP_ADR_MASK}    mtu=1500    state=up
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    linux: Check Veth Interface State     agent_vpp_1
    ...    vpp1_veth2    mac=${VETH2_MAC}    state=up

Add Afpacket2 Interface Before Veth3 And Veth4 Interfaces
    vpp_term: Interface Not Exists    node=agent_vpp_1    mac=${AFP2_MAC}
    Put Afpacket Interface    node=agent_vpp_1
    ...    name=vpp1_afpacket2    mac=${AFP2_MAC}    host_int=vpp1_veth3

Check That Afpacket2 Interface Is Not Created Without Veth3 And Veth4
    vpp_term: Interface Not Exists    node=agent_vpp_1    mac=${AFP2_MAC}

Add Veth3 Interface
    linux: Interface Not Exists    node=agent_vpp_1    mac=${VETH3_MAC}
    Put Veth Interface With IP    node=agent_vpp_1
    ...    name=vpp1_veth3    mac=${VETH3_MAC}    peer=vpp1_veth4
    ...    ip=${IP_ADR2}    prefix=64    mtu=1500
    linux: Interface Not Exists    node=agent_vpp_1    mac=${VETH3_MAC}

Check That Afpacket2 Is Not Created Without Veth4
    vpp_term: Interface Not Exists    node=agent_vpp_1    mac=${AFP2_MAC}

Add Veth4 Interface
    linux: Interface Not Exists    node=agent_vpp_1    mac=${VETH4_MAC}
    Put Veth Interface    node=agent_vpp_1
    ...    name=vpp1_veth4    mac=${VETH4_MAC}    peer=vpp1_veth3    enabled=true

Check That Afpacket2 Interface Is Created
    vpp_term: Interface Is Created    node=agent_vpp_1    mac=${AFP2_MAC}
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    vpp_api: Check Afpacket Interface State    agent_vpp_1
    ...    vpp1_afpacket2    enabled=1    mac=${AFP2_MAC}

Check That Veth3 And Veth4 Interfaces Are Created And Not Affected By Afpacket2 Interface
    linux: Interface Is Created    node=agent_vpp_1    mac=${VETH3_MAC}
    linux: Interface Is Created    node=agent_vpp_1    mac=${VETH4_MAC}
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    linux: Check Veth Interface State     agent_vpp_1
    ...    vpp1_veth3    mac=${VETH3_MAC}
    ...    ipv6=${IP_ADR_MASK2}    mtu=1500    state=up
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    linux: Check Veth Interface State     agent_vpp_1
    ...    vpp1_veth4    mac=${VETH4_MAC}    state=up

Check That Afpacket1 Interface Is Still Configured
    vpp_api: Check Afpacket Interface State    agent_vpp_1
    ...    vpp1_afpacket1    enabled=1    mac=${AFP1_MAC}

Update Afpacket2 Interface
    Put Afpacket Interface    node=agent_vpp_1
    ...    name=vpp1_afpacket2    mac=${AFP2_SEC_MAC}    host_int=vpp1_veth4
    vpp_term: Interface Is Deleted    node=agent_vpp_1    mac=${AFP2_MAC}
    vpp_term: Interface Is Created    node=agent_vpp_1    mac=${AFP2_SEC_MAC}
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    vpp_api: Check Afpacket Interface State    agent_vpp_1
    ...    vpp1_afpacket2    enabled=1    mac=${AFP2_SEC_MAC}

Check That Afpacket1 Interface Is Still Configured After Update
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    vpp_api: Check Afpacket Interface State    agent_vpp_1
    ...    vpp1_afpacket1    enabled=1    mac=${AFP1_MAC}

Check That Veth3 And Veth4 Interfaces Are Not Affected By Change Of Afpacket2 Interface
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    linux: Check Veth Interface State     agent_vpp_1
    ...    vpp1_veth3    mac=${VETH3_MAC}
    ...    ipv6=${IP_ADR_MASK2}    mtu=1500    state=up
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    linux: Check Veth Interface State     agent_vpp_1
    ...    vpp1_veth4    mac=${VETH4_MAC}    state=up

Delete Afpacket1 Interface
    Delete VPP Interface    node=agent_vpp_1    name=vpp1_afpacket1
    vpp_term: Interface Is Deleted    node=agent_vpp_1    mac=${AFP1_MAC}

Check That Afpacket2 Interface Is Still Configured
    vpp_api: Check Afpacket Interface State    agent_vpp_1
    ...    vpp1_afpacket2    enabled=1    mac=${AFP2_SEC_MAC}

Check That Veth1 And Veth2 Interfaces Are Not Affected By Delete Of Afpacket1 Interface
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    linux: Check Veth Interface State     agent_vpp_1
    ...    vpp1_veth1    mac=${VETH1_MAC}
    ...    ipv6=${IP_ADR_MASK}    mtu=1500    state=up
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    linux: Check Veth Interface State     agent_vpp_1
    ...    vpp1_veth2    mac=${VETH2_MAC}    state=up

Delete Veth3 Interface
    Delete Linux Interface    node=agent_vpp_1    name=vpp1_veth3
    linux: Interface Is Deleted    node=agent_vpp_1    mac=${VETH3_MAC}
    linux: Interface Is Deleted    node=agent_vpp_1    mac=${VETH4_MAC}

Check That Afpacket2 Interface Is Deleted After Deleting Veth3 And Veth4
    vpp_term: Interface Is Deleted    node=agent_vpp_1    mac=${AFP2_SEC_MAC}

Show Interfaces And Other Objects After Setup
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
    Write To Machine    agent_vpp_1_term    show err
    Write To Machine    agent_vpp_2_term    show err
    vpp_api: Interfaces Dump    agent_vpp_1
    vpp_api: Interfaces Dump    agent_vpp_2
    Execute In Container    agent_vpp_1    ip a
    Execute In Container    agent_vpp_2    ip a

*** Keywords ***
TestSetup
    Make Datastore Snapshots    ${TEST_NAME}_test_setup

TestTeardown
    Make Datastore Snapshots    ${TEST_NAME}_test_teardown

