*** Settings ***
Library      OperatingSystem
#Library      RequestsLibrary
#Library      SSHLibrary      timeout=60s
#Library      String

Resource     ../../../variables/${VARIABLES}_variables.robot

Resource     ../../../libraries/all_libs.robot

Force Tags        sfcIPv4
Suite Setup       Testsuite Setup
Suite Teardown    Suite Cleanup
Test Setup        TestSetup
Test Teardown     TestTeardown

*** Variables ***
${VARIABLES}=          common
${ENV}=                common

*** Test Cases ***
Configure Environment
    [Tags]    setup
    Add Agent VPP Node    agent_vpp_1
    Add Agent VPP Node    agent_vpp_2
    ${DATA_FOLDER}=       Catenate     SEPARATOR=/       ${CURDIR}         ${TEST_DATA_FOLDER}
    Set Suite Variable          ${DATA_FOLDER}
    Start SFC Controller Container With Own Config    simple.conf
    Sleep    15s

Check Memif Interface On VPP1
    ${out}=    vpp_term: Show Interfaces    agent_vpp_1
    Log    ${out}
    ${int}=    vpp_ctl: Get Interface Internal Name    agent_vpp_1    vpp1_memif1
    Should Contain    ${out}    ${int}
    ${out}=    Write To Machine    agent_vpp_1_term    show h
    Should Contain    ${out}    02:02:02:02:02:02

Check Memif Interface On VPP2
    ${out}=    vpp_term: Show Interfaces    agent_vpp_2
    Log    ${out}
    ${int}=    vpp_ctl: Get Interface Internal Name    agent_vpp_2    vpp2_memif1
    Should Contain    ${out}    ${int}
    ${out}=    Write To Machine    agent_vpp_2_term    show int addr
    Should Contain    ${out}    10.0.0.10

Show Interfaces And Other Objects For Debug
    [Tags]    debug
    vpp_term: Show Interfaces    agent_vpp_1
    vpp_term: Show Interfaces    agent_vpp_2            
    Write To Machine    agent_vpp_1_term    show int addr
    Write To Machine    agent_vpp_2_term    show int addr
    Write To Machine    agent_vpp_1_term    show h
    Write To Machine    agent_vpp_2_term    show h
    Write To Machine    agent_vpp_1_term    show err     
    Write To Machine    agent_vpp_2_term    show err     

Done
    [Tags]    debug
    No Operation


*** Keywords ***
Suite Cleanup
    Stop SFC Controller Container
    Testsuite Teardown

*** Keywords ***
TestSetup
    Make Datastore Snapshots    ${TEST_NAME}_test_setup

TestTeardown
    Make Datastore Snapshots    ${TEST_NAME}_test_teardown

