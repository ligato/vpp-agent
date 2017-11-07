*** Settings ***
Library      OperatingSystem
#Library      RequestsLibrary
#Library      SSHLibrary      timeout=60s
#Library      String
Library      Collections

Resource     ../../variables/${VARIABLES}_variables.robot

Resource     ../../libraries/all_libs.robot

Suite Setup       Testsuite Setup
Suite Teardown    Suite Cleanup
Test Setup        TestSetup
Test Teardown     TestTeardown

*** Variables ***
${REPLY_DATA_FOLDER}            replyACL
${VARIABLES}=       common
${ENV}=             common
${ACL1_NAME}=       acl1_tcp
${ACL2_NAME}=       acl2_tcp
${ACL3_NAME}=       acl3_UDP
${ACL4_NAME}=       acl4_UDP
${ACL5_NAME}=       acl5_ICMP
${ACL6_NAME}=       acl6_ICMP
${E_INTF1}=
${I_INTF1}=
${E_INTF2}=
${I_INTF2}=
${RULE_NM1_1}=         acl1_rule1
${RULE_NM2_1}=         acl2_rule1
${RULE_NM3_1}=         acl3_rule1
${RULE_NM4_1}=         acl4_rule1
${RULE_NM5_1}=         acl5_rule1
${RULE_NM6_1}=         acl6_rule1
${ACTION_DENY}=     1
${ACTION_PERMIT}=   2
${DEST_NTW}=        10.0.0.0/32
${SRC_NTW}=         10.0.0.0/32
${1DEST_PORT_L}=     80
${1DEST_PORT_U}=     1000
${1SRC_PORT_L}=      10
${1SRC_PORT_U}=      2000
${2DEST_PORT_L}=     2000
${2DEST_PORT_U}=     2200
${2SRC_PORT_L}=      20010
${2SRC_PORT_U}=      20020
${SYNC_SLEEP}=      1s
${NO_ACL}=



*** Test Cases ***
Configure Environment
    [Tags]    setup
    Configure Environment 2

Show ACL Before Setup
    Check ACL Reply    agent_vpp_1    ${ACL1_NAME}    ${REPLY_DATA_FOLDER}/reply_acl_empty.txt     ${REPLY_DATA_FOLDER}/reply_acl_empty_term.txt

Add ACL1_TCP
    vpp_ctl: Put ACL TCP   agent_vpp_1   ${ACL1_NAME}    ${E_INTF1}    ${I_INTF1}   ${RULE_NM1_1}    ${ACTION_DENY}     ${DEST_NTW}     ${SRC_NTW}   ${1DEST_PORT_L}   ${1DEST_PORT_U}    ${1SRC_PORT_L}     ${1SRC_PORT_U}
    Sleep    ${SYNC_SLEEP}

Check ACL1 is created
    Check ACL Reply    agent_vpp_1    ${ACL1_NAME}    ${REPLY_DATA_FOLDER}/reply_acl1_tcp.txt    ${REPLY_DATA_FOLDER}/reply_acl1_tcp_term.txt


Add ACL2_TCP
    vpp_ctl: Put ACL TCP   agent_vpp_1   ${ACL2_NAME}    ${E_INTF1}    ${I_INTF1}   ${RULE_NM2_1}    ${ACTION_DENY}     ${DEST_NTW}     ${SRC_NTW}   ${2DEST_PORT_L}   ${2DEST_PORT_U}    ${2SRC_PORT_L}     ${2SRC_PORT_U}
    Sleep    ${SYNC_SLEEP}

Check ACL2 is created and ACL1 still Configured
    Check ACL Reply    agent_vpp_1    ${ACL2_NAME}   ${REPLY_DATA_FOLDER}/reply_acl2_tcp.txt    ${REPLY_DATA_FOLDER}/reply_acl2_tcp_term.txt



Update ACL1
    vpp_ctl: Put ACL TCP   agent_vpp_1   ${ACL1_NAME}    ${E_INTF1}     ${I_INTF1}   ${RULE_NM1_1}    ${ACTION_PERMIT}     ${DEST_NTW}    ${SRC_NTW}   ${1DEST_PORT_L}   ${1DEST_PORT_U}    ${1SRC_PORT_L}     ${1SRC_PORT_U}
    Sleep    ${SYNC_SLEEP}

Check ACL1 Is Changed and ACL2 not changed
    Check ACL Reply    agent_vpp_1    ${ACL1_NAME}    ${REPLY_DATA_FOLDER}/reply_acl1_update_tcp.txt    ${REPLY_DATA_FOLDER}/reply_acl1_update_tcp_term.txt

Delete ACL2
    vpp_ctl: Delete ACL     agent_vpp_1    ${ACL2_NAME}
    Sleep    ${SYNC_SLEEP}

Check ACL2 Is Deleted and ACL1 Is Not Changed
    Check ACL Reply    agent_vpp_1    ${ACL2_NAME}    ${REPLY_DATA_FOLDER}/reply_acl_empty.txt    ${REPLY_DATA_FOLDER}/reply_acl2_delete_tcp_term.txt

Delete ACL1
    vpp_ctl: Delete ACL     agent_vpp_1    ${ACL1_NAME}
    Sleep    ${SYNC_SLEEP}

Check ACL1 Is Deleted
    Check ACL Reply    agent_vpp_1    ${ACL1_NAME}    ${REPLY_DATA_FOLDER}/reply_acl_empty.txt   ${REPLY_DATA_FOLDER}/reply_acl_empty_term.txt


ADD ACL3_UDP
    vpp_ctl: Put ACL UDP    agent_vpp_1    ${ACL3_NAME}    ${E_INTF1}   ${I_INTF1}    ${E_INTF2}    ${I_INTF2}    ${RULE_NM3_1}    ${ACTION_DENY}    ${DEST_NTW}     ${SRC_NTW}   ${1DEST_PORT_L}   ${1DEST_PORT_U}    ${1SRC_PORT_L}     ${1SRC_PORT_U}
    Sleep    ${SYNC_SLEEP}

Check ACL3 Is Created
    Check ACL Reply    agent_vpp_1    ${ACL3_NAME}    ${REPLY_DATA_FOLDER}/reply_acl3_udp.txt    ${REPLY_DATA_FOLDER}/reply_acl3_udp_term.txt

ADD ACL4_UDP
    vpp_ctl: Put ACL UDP    agent_vpp_1    ${ACL4_NAME}    ${E_INTF1}    ${I_INTF1}    ${E_INTF2}    ${I_INTF2}    ${RULE_NM4_1}     ${ACTION_DENY}    ${DEST_NTW}     ${SRC_NTW}   ${1DEST_PORT_L}   ${1DEST_PORT_U}    ${1SRC_PORT_L}     ${1SRC_PORT_U}
    Sleep    ${SYNC_SLEEP}

Check ACL4 Is Created And ACL3 Still Configured
    Check ACL Reply    agent_vpp_1    ${ACL4_NAME}    ${REPLY_DATA_FOLDER}/reply_acl4_udp.txt     ${REPLY_DATA_FOLDER}/reply_acl4_udp_term.txt

Delete ACL4
    vpp_ctl: Delete ACL     agent_vpp_1    ${ACL4_NAME}
    Sleep    ${SYNC_SLEEP}

Check ACL4 Is Deleted and ACL3 Is Not Changed
    Check ACL Reply    agent_vpp_1    ${ACL4_NAME}   ${REPLY_DATA_FOLDER}/reply_acl_empty.txt     ${REPLY_DATA_FOLDER}/reply_acl3_udp_term.txt

Delete ACL3
    vpp_ctl: Delete ACL     agent_vpp_1    ${ACL3_NAME}
    Sleep    ${SYNC_SLEEP}

Check ACL3 Is Deleted
    Check ACL Reply    agent_vpp_1    ${ACL3_NAME}    ${REPLY_DATA_FOLDER}/reply_acl_empty.txt    ${REPLY_DATA_FOLDER}/reply_acl_empty_term.txt

ADD ACL5_ICMP
    vpp_ctl: Put ACL UDP    agent_vpp_1    ${ACL5_NAME}    ${E_INTF1}    ${I_INTF1}    ${E_INTF2}    ${I_INTF2}    ${RULE_NM5_1}    ${ACTION_DENY}    ${DEST_NTW}     ${SRC_NTW}   ${1DEST_PORT_L}   ${1DEST_PORT_U}    ${1SRC_PORT_L}     ${1SRC_PORT_U}
    Sleep    ${SYNC_SLEEP}

Check ACL5 Is Created
    Check ACL Reply    agent_vpp_1    ${ACL5_NAME}   ${REPLY_DATA_FOLDER}/reply_acl5_icmp.txt    ${REPLY_DATA_FOLDER}/reply_acl5_icmp_term.txt

ADD ACL6_ICMP
    vpp_ctl: Put ACL UDP    agent_vpp_1    ${ACL6_NAME}    ${E_INTF1}    ${I_INTF1}    ${E_INTF2}    ${I_INTF2}    ${RULE_NM6_1}    ${ACTION_DENY}  ${DEST_NTW}     ${SRC_NTW}   ${1DEST_PORT_L}   ${1DEST_PORT_U}    ${1SRC_PORT_L}     ${1SRC_PORT_U}
    Sleep    ${SYNC_SLEEP}

Check ACL6 Is Created And ACL5 Still Configured
    Check ACL Reply    agent_vpp_1    ${ACL6_NAME}    ${REPLY_DATA_FOLDER}/reply_acl6_icmp.txt    ${REPLY_DATA_FOLDER}/reply_acl6_icmp_term.txt

Delete ACL6
    vpp_ctl: Delete ACL     agent_vpp_1    ${ACL6_NAME}
    Sleep    ${SYNC_SLEEP}

Check ACL6 Is Deleted and ACL5 Is Not Changed
    Check ACL Reply    agent_vpp_1    ${ACL6_NAME}     ${REPLY_DATA_FOLDER}/reply_acl_empty.txt    ${REPLY_DATA_FOLDER}/reply_acl5_icmp_term.txt

Delete ACL5
    vpp_ctl: Delete ACL     agent_vpp_1    ${ACL5_NAME}
    Sleep    ${SYNC_SLEEP}

Check ACL5 Is Deleted
    Check ACL Reply    agent_vpp_1    ${ACL5_NAME}   ${REPLY_DATA_FOLDER}/reply_acl_empty.txt     ${REPLY_DATA_FOLDER}/reply_acl_empty_term.txt


Add 6 ACL
    vpp_ctl: Put ACL TCP   agent_vpp_1   ${ACL1_NAME}    ${E_INTF1}    ${I_INTF1}   ${RULE_NM1_1}    ${ACTION_DENY}     ${DEST_NTW}     ${SRC_NTW}   ${1DEST_PORT_L}   ${1DEST_PORT_U}    ${1SRC_PORT_L}     ${1SRC_PORT_U}
    vpp_ctl: Put ACL TCP   agent_vpp_1   ${ACL2_NAME}    ${E_INTF1}    ${I_INTF1}   ${RULE_NM2_1}    ${ACTION_DENY}     ${DEST_NTW}     ${SRC_NTW}   ${2DEST_PORT_L}   ${2DEST_PORT_U}    ${2SRC_PORT_L}     ${2SRC_PORT_U}
    vpp_ctl: Put ACL UDP   agent_vpp_1    ${ACL3_NAME}    ${E_INTF1}   ${I_INTF1}    ${E_INTF2}    ${I_INTF2}    ${RULE_NM3_1}    ${ACTION_DENY}    ${DEST_NTW}     ${SRC_NTW}   ${1DEST_PORT_L}   ${1DEST_PORT_U}    ${1SRC_PORT_L}     ${1SRC_PORT_U}
    vpp_ctl: Put ACL UDP   agent_vpp_1    ${ACL4_NAME}    ${E_INTF1}    ${I_INTF1}    ${E_INTF2}    ${I_INTF2}    ${RULE_NM4_1}     ${ACTION_DENY}    ${DEST_NTW}     ${SRC_NTW}   ${1DEST_PORT_L}   ${1DEST_PORT_U}    ${1SRC_PORT_L}     ${1SRC_PORT_U}
    vpp_ctl: Put ACL UDP   agent_vpp_1    ${ACL5_NAME}    ${E_INTF1}    ${I_INTF1}    ${E_INTF2}    ${I_INTF2}    ${RULE_NM5_1}    ${ACTION_DENY}    ${DEST_NTW}     ${SRC_NTW}   ${1DEST_PORT_L}   ${1DEST_PORT_U}    ${1SRC_PORT_L}     ${1SRC_PORT_U}
    vpp_ctl: Put ACL UDP   agent_vpp_1    ${ACL6_NAME}    ${E_INTF1}    ${I_INTF1}    ${E_INTF2}    ${I_INTF2}    ${RULE_NM6_1}    ${ACTION_DENY}  ${DEST_NTW}     ${SRC_NTW}   ${1DEST_PORT_L}   ${1DEST_PORT_U}    ${1SRC_PORT_L}     ${1SRC_PORT_U}

Check All 6 ACLs Added
    Check ACL All Reply    agent_vpp_1     ${REPLY_DATA_FOLDER}/reply_acl_all.txt        ${REPLY_DATA_FOLDER}/reply_acl_all_term.txt

*** Keywords ***

Check ACL Reply
    [Arguments]         ${node}    ${acl_name}   ${reply_json}    ${reply_term}
    Log Many            ${node}    ${acl_name}   ${reply_json}    ${reply_term}
    ${acl_d}=           vpp_ctl: Get ACL As Json    ${node}    ${acl_name}
    ${term_d}=          vat_term: Check ACL     ${node}    ${acl_name}
    ${term_d_lines}=    Split To Lines    ${term_d}
    Log                 ${term_d_lines}
    ${data}=            OperatingSystem.Get File    ${reply_json}
    Should Be Equal     ${data}   ${acl_d}
    ${data}=            OperatingSystem.Get File    ${reply_term}
    ${t_data_lines}=    Split To Lines    ${data}
    Log                 ${t_data_lines}
    List Should Contain Sub List    ${term_d_lines}    ${t_data_lines}

Check ACL All Reply
    [Arguments]         ${node}    ${reply_json}     ${reply_term}
    Log Many            ${node}    ${reply_json}     ${reply_term}
    ${acl_d}=           vpp_ctl: Get All ACL As Json    ${node}
    ${term_d}=          vat_term: Check All ACL     ${node}
    ${term_d_lines}=    Split To Lines    ${term_d}
    Log                 ${term_d_lines}
    ${data}=            OperatingSystem.Get File    ${reply_json}
    Should Be Equal     ${data}   ${acl_d}
    ${data}=            OperatingSystem.Get File    ${reply_term}
    ${t_data_lines}=    Split To Lines    ${data}
    Log                 ${t_data_lines}
    List Should Contain Sub List    ${term_d_lines}    ${t_data_lines}


TestSetup
    Make Datastore Snapshots    ${TEST_NAME}_test_setup

TestTeardown
    Make Datastore Snapshots    ${TEST_NAME}_test_teardown

Suite Cleanup
    Stop SFC Controller Container
    Testsuite Teardown