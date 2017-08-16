*** Settings ***
Library      OperatingSystem
#Library      RequestsLibrary
#Library      SSHLibrary      timeout=60s
#Library      String

Resource     ../../../../variables/${VARIABLES}_variables.robot

Resource     ../../../../libraries/all_libs.robot

Suite Setup       Testsuite Setup
Suite Teardown    Suite Cleanup

*** Variables ***
${VARIABLES}=          common
${ENV}=                common
${FINAL_SLEEP}=        3s

*** Test Cases ***
Configure Environment
    [Tags]    setup
    Start SFC Controller Container With Own Config    basic.conf
    Add Agent VPP Node    agent_vpp_1    vswitch=${TRUE}
    Add Agent VPP Node    agent_vpp_2
    Add Agent Node    agent_1

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

Show Interfaces And Other Objects After Config
    vpp_term: Show Interfaces    agent_vpp_1
    vpp_term: Show Interfaces    agent_vpp_2
    vpp_term: Show Interfaces    agent_vpp_3
    Write To Machine    agent_vpp_1_term    show int addr
    Write To Machine    agent_vpp_2_term    show int addr
    Write To Machine    agent_vpp_3_term    show int addr
    Write To Machine    agent_vpp_1_term    show h
    Write To Machine    agent_vpp_2_term    show h
    Write To Machine    agent_vpp_3_term    show h
    Write To Machine    agent_vpp_1_term    show br
    Write To Machine    agent_vpp_2_term    show br
    Write To Machine    agent_vpp_3_term    show br
    Write To Machine    agent_vpp_1_term    show br 1 detail
    Write To Machine    agent_vpp_2_term    show br 1 detail
    Write To Machine    agent_vpp_3_term    show br 1 detail
    Write To Machine    agent_vpp_1_term    show vxlan tunnel
    Write To Machine    agent_vpp_2_term    show vxlan tunnel
    Write To Machine    agent_vpp_3_term    show vxlan tunnel
    Write To Machine    agent_vpp_1_term    show err
    Write To Machine    agent_vpp_2_term    show err
    Write To Machine    agent_vpp_3_term    show err
    vat_term: Interfaces Dump    agent_vpp_1
    vat_term: Interfaces Dump    agent_vpp_2
    vat_term: Interfaces Dump    agent_vpp_3
    Write To Machine    vpp_agent_ctl    vpp-agent-ctl ${AGENT_VPP_ETCD_CONF_PATH} -ps
    Execute In Container    agent_vpp_1    ip a
    Execute In Container    agent_vpp_2    ip a
    Execute In Container    agent_vpp_3    ip a

Done
    [Tags]    debug
    No Operation

Final Sleep For Manual Checking
    [Tags]    debug
    Sleep   ${FINAL_SLEEP}

*** Keywords ***
Suite Cleanup
    Stop SFC Controller Container
    Testsuite Teardown
