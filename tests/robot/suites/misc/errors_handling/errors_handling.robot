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
${AFP1_MAC_GOOD}=           a2:01:01:01:01:01
${AFP1_MAC_BAD}=           a2:01:01:01:01:01:xy

*** Test Cases ***
Configure Environment
    [Tags]    setup
    Add Agent VPP Node    agent_vpp_1

Show Interfaces Before Setup
    vpp_term: Show Interfaces    agent_vpp_1

Interface Should Not Be Present
    vpp_term: Interface Not Exists    node=agent_vpp_1    mac=${AFP1_MAC}
    ${int_key}=    Set Variable    /vnf-agent/${node}/vpp/status/v1/interface/vpp1_afpacket1
    ${int_error_key}=    Set Variable    /vnf-agent/${node}/vpp/status/v1/interface/error/vpp1_afpacket1
    Log Many    ${int_key}    ${int_error_key}
    ${out}=    vpp_ctl: Read Key    ${int_key}
    Shoud Be Empty    ${out}
    ${out}=    vpp_ctl: Read Key    ${int_error_key}
    Shoud Be Empty    ${out}


    vpp_ctl: Put Afpacket Interface    node=agent_vpp_1    name=vpp1_afpacket1    mac=${AFP1_MAC_BAD}    host_int=vpp1_veth2

test_end
   sleep   5

Show Interfaces And Other Objects After Setup
    vpp_term: Show Interfaces    agent_vpp_1
    Write To Machine    agent_vpp_1_term    show int addr
    Write To Machine    agent_vpp_1_term    show h
    Write To Machine    agent_vpp_1_term    show memif
    Write To Machine    agent_vpp_1_term    show err
    vat_term: Interfaces Dump    agent_vpp_1
    Write To Machine    vpp_agent_ctl    vpp-agent-ctl ${AGENT_VPP_ETCD_CONF_PATH} -ps
    Execute In Container    agent_vpp_1    ip a

*** Keywords ***
TestSetup
    Make Datastore Snapshots    ${TEST_NAME}_test_setup

TestTeardown
    Make Datastore Snapshots    ${TEST_NAME}_test_teardown

