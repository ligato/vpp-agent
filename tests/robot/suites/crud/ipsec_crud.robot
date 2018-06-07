*** Settings ***
Documentation    IPsec CRUD
Library     OperatingSystem
Library     String
#Library     RequestsLibrary

Resource     ../../variables/${VARIABLES}_variables.robot
Resource    ../../libraries/all_libs.robot
Resource    ../../libraries/pretty_keywords.robot

#Suite Setup       Run Keywords    Discard old results
Suite Setup       Testsuite Setup
Suite Teardown    Testsuite Teardown
#Test Setup        Test Setup
#Test Teardown     Test Teardown

*** Variables ***
${VARIABLES}=          common
${ENV}=                common

*** Test Cases ***
# CRUD tests for IPsec
Add Agent Vpp Node
    Add Agent VPP Node                 agent_vpp_1

Add SA1 Into VPP
    IP Sec On agent_vpp_1 Should Not Contain SA sa 1
    Create IPsec On agent_vpp_1 With SA sa10 And Json ipsec-sa10.json
    IP Sec On agent_vpp_1 Should Contain SA sa 1

Add SA2 Into VPP
    IP Sec On agent_vpp_1 Should Not Contain SA sa 2
    Create IPsec On agent_vpp_1 With SA sa20 And Json ipsec-sa20.json
    IP Sec On agent_vpp_1 Should Contain SA sa 2

Add SPD Into VPP
    IP Sec On agent_vpp_1 Should Not Contain SA spd 1
    Create IPsec On agent_vpp_1 With SPD spd1 And Json ipsec-spd.json
    IP Sec On agent_vpp_1 Should Contain SA spd 1

Check IPsec config On VPP
    Show IPsec On agent_vpp_1
    IP Sec Should Contain  agent_vpp_1  sa 1  sa 2  spd 1  IPSEC_ESP  outbound policies
#    IP Sec On agent_vpp_1 Should Contain SA sa 2
#    IP Sec On agent_vpp_1 Should Contain SA spd 1

Delete SAs And SPD For Default IPsec
    Delete IPsec On agent_vpp_1 And sa/sa10
    Delete IPsec On agent_vpp_1 And sa/sa20
    Delete IPsec On agent_vpp_1 And spd/spd1
    IP Sec On agent_vpp_1 Should Not Contain SA sa 1
    IP Sec On agent_vpp_1 Should Not Contain SA sa 2
    IP Sec On agent_vpp_1 Should Not Contain SA spd 1


*** Keywords ***
IP Sec On ${node} Should Not Contain SA ${sa}
    Log many    ${node}
    ${out}=    vpp_term: Show IPsec    ${node}
    log many    ${out}
    Should Not Contain  ${out}  ${sa}

IP Sec On ${node} Should Contain SA ${sa}
    Log many    ${node}
    ${out}=    vpp_term: Show IPsec    ${node}
    log many    ${out}
    Should Contain  ${out}  ${sa}

IP Sec Should Contain
    [Arguments]     ${node}  ${data1}  ${data2}  ${data3}  ${data4}  ${data5}
    Log many        ${node}
    ${out}=         vpp_term: Show IPsec    ${node}
    log many        ${out}
    Should Contain  ${out}  ${data1}
    Should Contain  ${out}  ${data2}
    Should Contain  ${out}  ${data3}
    Should Contain  ${out}  ${data4}
    Should Contain  ${out}  ${data5}
