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
${VETH1_MAC}=          1a:00:00:11:11:11
${VETH1_SEC_MAC}=      1a:00:00:11:11:12
${VETH2_MAC}=          2a:00:00:22:22:22
${VETH3_MAC}=          3a:00:00:33:33:33
${VETH4_MAC}=          4a:00:00:44:44:44

*** Test Cases ***
Configure Environment
    [Tags]    setup
    Configure Environment 1

Show Interfaces Before Setup
    vpp_term: Show Interfaces    agent_vpp_1

Add Veth1 Interface
    linux: Interface Not Exists    node=agent_vpp_1    mac=${VETH1_MAC}
    vpp_ctl: Put Veth Interface With IP    node=agent_vpp_1    name=vpp1_veth1    mac=${VETH1_MAC}    peer=vpp1_veth2    ip=10.10.1.1    prefix=24    mtu=1500
    linux: Interface Not Exists    node=agent_vpp_1    mac=${VETH1_MAC}

Add Veth2 Interface
    linux: Interface Not Exists    node=agent_vpp_1    mac=${VETH2_MAC}
    vpp_ctl: Put Veth Interface    node=agent_vpp_1    name=vpp1_veth2    mac=${VETH2_MAC}    peer=vpp1_veth1

Check That Veth1 And Veth2 Interfaces Are Created
    linux: Interface Is Created    node=agent_vpp_1    mac=${VETH1_MAC}
    linux: Interface Is Created    node=agent_vpp_1    mac=${VETH2_MAC}
    linux: Check Veth Interface State     agent_vpp_1    vpp1_veth1    mac=${VETH1_MAC}    ipv4=10.10.1.1/24    mtu=1500    state=up
    linux: Check Veth Interface State     agent_vpp_1    vpp1_veth2    mac=${VETH2_MAC}    state=up

Add Veth3 Interface
    linux: Interface Not Exists    node=agent_vpp_1    mac=${VETH3_MAC}
    vpp_ctl: Put Veth Interface With IP    node=agent_vpp_1    name=vpp1_veth3    mac=${VETH3_MAC}    peer=vpp1_veth4    ip=20.20.1.1    prefix=24    mtu=1500
    linux: Interface Not Exists    node=agent_vpp_1    mac=${VETH3_MAC}

Add Veth4 Interface
    linux: Interface Not Exists    node=agent_vpp_1    mac=${VETH4_MAC}
    vpp_ctl: Put Veth Interface    node=agent_vpp_1    name=vpp1_veth4    mac=${VETH4_MAC}    peer=vpp1_veth3    enabled=false

Check That Veth3 And Veth4 Interfaces Are Created
    linux: Interface Is Created    node=agent_vpp_1    mac=${VETH3_MAC}
    linux: Interface Is Created    node=agent_vpp_1    mac=${VETH4_MAC}
    linux: Check Veth Interface State     agent_vpp_1    vpp1_veth3    mac=${VETH3_MAC}    ipv4=20.20.1.1/24    mtu=1500    state=lowerlayerdown
    linux: Check Veth Interface State     agent_vpp_1    vpp1_veth4    mac=${VETH4_MAC}    state=down

Check That Veth1 And Veth2 Interfaces Are Still Configured
    linux: Check Veth Interface State     agent_vpp_1    vpp1_veth1    mac=${VETH1_MAC}    ipv4=10.10.1.1/24    mtu=1500    state=up
    linux: Check Veth Interface State     agent_vpp_1    vpp1_veth2    mac=${VETH2_MAC}    state=up

Update Veth1 Interface
    vpp_ctl: Put Veth Interface With IP    node=agent_vpp_1    name=vpp1_veth1    mac=${VETH1_SEC_MAC}    peer=vpp1_veth2    ip=11.11.1.1    prefix=28    mtu=1600
    linux: Interface Is Deleted    node=agent_vpp_1    mac=${VETH1_MAC}
    linux: Interface Is Created    node=agent_vpp_1    mac=${VETH1_SEC_MAC}
    linux: Check Veth Interface State     agent_vpp_1    vpp1_veth1    mac=${VETH1_SEC_MAC}    ipv4=11.11.1.1/28    mtu=1600    state=up

Check That Veth2 And Veth3 And Veth4 interfaces Are Still Configured
    linux: Check Veth Interface State     agent_vpp_1    vpp1_veth2    mac=${VETH2_MAC}    state=up
    linux: Check Veth Interface State     agent_vpp_1    vpp1_veth3    mac=${VETH3_MAC}    ipv4=20.20.1.1/24    mtu=1500    state=lowerlayerdown
    linux: Check Veth Interface State     agent_vpp_1    vpp1_veth4    mac=${VETH4_MAC}    state=down

Delete Veth2 Interface
    vpp_ctl: Delete Linux Interface    node=agent_vpp_1    name=vpp1_veth2
    linux: Interface Is Deleted    node=agent_vpp_1    mac=${VETH1_SEC_MAC}
    linux: Interface Is Deleted    node=agent_vpp_1    mac=${VETH2_MAC}

Check That Veth3 And Veth4 Are Still Configured
    linux: Check Veth Interface State     agent_vpp_1    vpp1_veth3    mac=${VETH3_MAC}    ipv4=20.20.1.1/24    mtu=1500    state=lowerlayerdown
    linux: Check Veth Interface State     agent_vpp_1    vpp1_veth4    mac=${VETH4_MAC}    state=down

Delete Veth3 Interface
    vpp_ctl: Delete Linux Interface    node=agent_vpp_1    name=vpp1_veth3
    linux: Interface Is Deleted    node=agent_vpp_1    mac=${VETH3_MAC}
    linux: Interface Is Deleted    node=agent_vpp_1    mac=${VETH4_MAC}

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

