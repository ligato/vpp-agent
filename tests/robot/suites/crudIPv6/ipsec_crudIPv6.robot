*** Settings ***
Documentation    IPsec CRUD
Library     OperatingSystem
Library     String

Resource     ../../variables/${VARIABLES}_variables.robot
Resource    ../../libraries/all_libs.robot
Resource    ../../libraries/pretty_keywords.robot

Force Tags        crud     IPv6
Suite Setup       Testsuite Setup
Suite Teardown    Testsuite Teardown

*** Variables ***
${ENV}=                common
${REMOTE_ADDR_1}=      fd30::1:1:0:0:1
${LOCAL_ADDR_1}=       fd30::1:1:0:0:2
${REMOTE_ADDR_2}=      fd30::1:1:0:0:3
${LOCAL_ADDR_2}=       fd30::1:1:0:0:4
${WAIT_TIMEOUT}=     20s
${SYNC_SLEEP}=       3s

*** Test Cases ***
# CRUD tests for IPsec
Add Agent Vpp Node
    Add Agent VPP Node                 agent_vpp_1

Add SA1 Into VPP
    IP Sec On agent_vpp_1 Should Not Contain SA sa 1
    Create IPsec With SA And Json  agent_vpp_1   sa10   ipsec-sa.json  10  1001  4a506a794f574265564551694d653768  4339314b55523947594d6d3547666b45764e6a58
#    Create IPsec On agent_vpp_1 With SA sa10 And Json ipsec-sa.json
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    IP Sec On agent_vpp_1 Should Contain SA [0] sa 0xa spi 1001

Add SA2 Into VPP
    IP Sec On agent_vpp_1 Should Not Contain SA sa 2
    Create IPsec With SA And Json  agent_vpp_1   sa20   ipsec-sa.json  20  1000  4a506a794f574265564551694d653768  4339314b55523947594d6d3547666b45764e6a58
#    Create IPsec On agent_vpp_1 With SA sa20 And Json ipsec-sa20.json
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    IP Sec On agent_vpp_1 Should Contain SA [1] sa 0x14 spi 1000

Add SPD1 Into VPP
    IP Sec On agent_vpp_1 Should Not Contain SA spd 1
    Create IPsec With SPD And Json  agent_vpp_1    1    ipsec-spd.json    afp1    ${REMOTE_ADDR_1}    ${LOCAL_ADDR_1}    10  20
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    IP Sec On agent_vpp_1 Should Contain SA spd 1

Check IPsec config_1 On VPP
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    IP Sec Should Contain  agent_vpp_1  sa 1  sa 2  spd 1  ip4-outbound   ip4-inbound-protect

Add SA3 Into VPP
    IP Sec On agent_vpp_1 Should Not Contain SA sa 3
    Create IPsec With SA And Json  agent_vpp_1   sa30   ipsec-sa.json  30  1003  4a506a794f574265564551694d653770  4339314b55523947594d6d3547666b45764e6a60
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    IP Sec On agent_vpp_1 Should Contain SA [2] sa 0x1e spi 1003

Add SA4 Into VPP
    IP Sec On agent_vpp_1 Should Not Contain SA sa 4
    Create IPsec With SA And Json  agent_vpp_1   sa40   ipsec-sa.json  40  1002  4a506a794f574265564551694d653770  4339314b55523947594d6d3547666b45764e6a60
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    IP Sec On agent_vpp_1 Should Contain SA [3] sa 0x28 spi 1002

Add SPD2 Into VPP
    IP Sec On agent_vpp_1 Should Not Contain SA spd 2
    Create IPsec With SPD And Json  agent_vpp_1    2    ipsec-spd.json    afp2    ${REMOTE_ADDR_2}    ${LOCAL_ADDR_2}    30  40
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    IP Sec On agent_vpp_1 Should Contain SA spd 2

Check IPsec config_1 On VPP After Add SPD2
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    IP Sec Should Contain  agent_vpp_1  sa 1  sa 2  spd 1  ip4-outbound   ip4-inbound-protect

Check IPsec config_2 On VPP
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    IP Sec Should Contain  agent_vpp_1  sa 3  sa 4  spd 2  ip4-outbound   ip4-inbound-protect

Delete SAs And SPD1 For Default IPsec
    Delete IPsec On agent_vpp_1 With Prefix sa And Index 10
    Delete IPsec On agent_vpp_1 With Prefix sa And Index 20
    Delete IPsec On agent_vpp_1 With Prefix spd And Index 1
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    IP Sec On agent_vpp_1 Should Not Contain SA sa 10
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    IP Sec On agent_vpp_1 Should Not Contain SA sa 20
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    IP Sec On agent_vpp_1 Should Not Contain SA spd 1

Check IPsec config_2 On VPP After Delete SPD1
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    IP Sec Should Contain  agent_vpp_1  sa 3  sa 4  spd 2  ip4-outbound   ip4-inbound-protect

Delete SAs And SPD2 For Default IPsec
    Delete IPsec On agent_vpp_1 With Prefix sa And Index 30
    Delete IPsec On agent_vpp_1 With Prefix sa And Index 40
    Delete IPsec On agent_vpp_1 With Prefix spd And Index 2
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    IP Sec On agent_vpp_1 Should Not Contain SA sa 30
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    IP Sec On agent_vpp_1 Should Not Contain SA sa 40
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    IP Sec On agent_vpp_1 Should Not Contain SA spd 2


*** Keywords ***
IP Sec On ${node} Should Not Contain SA ${sa}
    ${out}=    vpp_term: Show IPsec    ${node}
    Should Not Contain  ${out}  ${sa}

IP Sec On ${node} Should Contain SA ${sa}
    ${out}=    vpp_term: Show IPsec    ${node}
    Should Contain  ${out}  ${sa}

IP Sec Should Contain
    [Arguments]     ${node}  ${sa_name_1}  ${sa_name_2}  ${spd_name_1}  ${inbound_policies}  ${outbound_policies}
    ${out}=         vpp_term: Show IPsec    ${node}
    Run Keyword Unless  "${sa_name_1}" == "${EMPTY}"   Should Contain  ${out}  ${sa_name_1}
    Run Keyword Unless  "${sa_name_2}" == "${EMPTY}"   Should Contain  ${out}  ${sa_name_2}
    Run Keyword Unless  "${spd_name_1}" == "${EMPTY}"   Should Contain  ${out}  ${spd_name_1}
    Run Keyword Unless  "${inbound_policies}" == "${EMPTY}"   Should Contain  ${out}  ${inbound_policies}
    Run Keyword Unless  "${outbound_policies}" == "${EMPTY}"   Should Contain  ${out}  ${outbound_policies}
