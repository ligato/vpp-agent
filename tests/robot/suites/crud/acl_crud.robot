*** Settings ***
Library      OperatingSystem
Library      Collections

Resource     ../../variables/${VARIABLES}_variables.robot

Resource     ../../libraries/vpp_api.robot
Resource     ../../libraries/vpp_term.robot
Resource     ../../libraries/docker.robot
Resource     ../../libraries/setup-teardown.robot
Resource     ../../libraries/configurations.robot

Resource     ../../libraries/acl/acl_etcd.robot
Resource     ../../libraries/acl/acl_vpp.robot

Force Tags        crud     IPv4
Suite Setup       Testsuite Setup
Suite Teardown    Suite Cleanup
Test Setup        TestSetup
Test Teardown     TestTeardown

*** Variables ***
${VARIABLES}=        common
${ENV}=              common
${ACL1_NAME}=        acl1_tcp
${ACL2_NAME}=        acl2_tcp
${ACL3_NAME}=        acl3_UDP
${ACL4_NAME}=        acl4_UDP
${ACL5_NAME}=        acl5_ICMP
${ACL6_NAME}=        acl6_ICMP
${E_INTF1}=
${I_INTF1}=
${E_INTF2}=
${I_INTF2}=
${ACTION_DENY}=      1
${ACTION_PERMIT}=    2
${DEST_NTW}=         10.0.0.0/32
${SRC_NTW}=          10.0.0.0/32
${1DEST_PORT_L}=     80
${1DEST_PORT_U}=     1000
${1SRC_PORT_L}=      10
${1SRC_PORT_U}=      2000
${2DEST_PORT_L}=     2000
${2DEST_PORT_U}=     2200
${2SRC_PORT_L}=      20010
${2SRC_PORT_U}=      20020
${TCP_FLAGS_MASK}=   20
${TCP_FLAGS_VALUE}=  10
${ICMP_v4}=          false
${ICMP_CODE_L}=      2
${ICMP_CODE_U}=      4
${ICMP_TYPE_L}=      1
${ICMP_TYPE_U}=      3
${WAIT_TIMEOUT}=     20s
${SYNC_SLEEP}=       3s
${NO_ACL}=


*** Test Cases ***
Configure Environment
    [Tags]    setup
    ${DATA_FOLDER}=       Catenate     SEPARATOR=/       ${CURDIR}         ${TEST_DATA_FOLDER}
    Set Suite Variable          ${DATA_FOLDER}
    Configure Environment 2        acl_basic.conf

Add ACL1_TCP
    Put ACL TCP   agent_vpp_1    ${ACL1_NAME}
    ...    ${E_INTF1}    ${I_INTF1}    ${ACTION_DENY}
    ...    ${DEST_NTW}     ${SRC_NTW}
    ...    ${1DEST_PORT_L}   ${1DEST_PORT_U}
    ...    ${1SRC_PORT_L}     ${1SRC_PORT_U}
    ...    ${TCP_FLAGS_MASK}    ${TCP_FLAGS_VALUE}

Check ACL1 is created
    Check ACL TCP    agent_vpp_1    ${ACL1_NAME}
    ...    ${E_INTF1}    ${I_INTF1}    ${ACTION_DENY}
    ...    ${DEST_NTW}     ${SRC_NTW}
    ...    ${1DEST_PORT_L}   ${1DEST_PORT_U}
    ...    ${1SRC_PORT_L}     ${1SRC_PORT_U}
    ...    ${TCP_FLAGS_MASK}    ${TCP_FLAGS_VALUE}

Add ACL2_TCP
    Put ACL TCP   agent_vpp_1    ${ACL2_NAME}
    ...    ${E_INTF1}    ${I_INTF1}    ${ACTION_DENY}
    ...    ${DEST_NTW}     ${SRC_NTW}
    ...    ${2DEST_PORT_L}   ${2DEST_PORT_U}
    ...    ${2SRC_PORT_L}     ${2SRC_PORT_U}
    ...    ${TCP_FLAGS_MASK}    ${TCP_FLAGS_VALUE}

Check ACL2 is created and ACL1 still Configured
    Check ACL TCP    agent_vpp_1    ${ACL1_NAME}
    ...    ${E_INTF1}    ${I_INTF1}    ${ACTION_DENY}
    ...    ${DEST_NTW}     ${SRC_NTW}
    ...    ${1DEST_PORT_L}   ${1DEST_PORT_U}
    ...    ${1SRC_PORT_L}     ${1SRC_PORT_U}
    ...    ${TCP_FLAGS_MASK}    ${TCP_FLAGS_VALUE}
    Check ACL TCP    agent_vpp_1    ${ACL2_NAME}
    ...    ${E_INTF1}    ${I_INTF1}    ${ACTION_DENY}
    ...    ${DEST_NTW}     ${SRC_NTW}
    ...    ${2DEST_PORT_L}   ${2DEST_PORT_U}
    ...    ${2SRC_PORT_L}     ${2SRC_PORT_U}
    ...    ${TCP_FLAGS_MASK}    ${TCP_FLAGS_VALUE}

Update ACL1
    Put ACL TCP   agent_vpp_1   ${ACL1_NAME}
    ...    ${E_INTF1}    ${I_INTF1}    ${ACTION_PERMIT}
    ...    ${DEST_NTW}    ${SRC_NTW}
    ...    ${1DEST_PORT_L}   ${1DEST_PORT_U}
    ...    ${1SRC_PORT_L}     ${1SRC_PORT_U}
    ...    ${TCP_FLAGS_MASK}    ${TCP_FLAGS_VALUE}

Check ACL1 Is Changed and ACL2 not changed
    Check ACL TCP    agent_vpp_1    ${ACL1_NAME}
    ...    ${E_INTF1}    ${I_INTF1}    ${ACTION_PERMIT}
    ...    ${DEST_NTW}     ${SRC_NTW}
    ...    ${1DEST_PORT_L}   ${1DEST_PORT_U}
    ...    ${1SRC_PORT_L}     ${1SRC_PORT_U}
    ...    ${TCP_FLAGS_MASK}    ${TCP_FLAGS_VALUE}
    Check ACL TCP    agent_vpp_1    ${ACL2_NAME}
    ...    ${E_INTF1}    ${I_INTF1}    ${ACTION_DENY}
    ...    ${DEST_NTW}     ${SRC_NTW}
    ...    ${2DEST_PORT_L}   ${2DEST_PORT_U}
    ...    ${2SRC_PORT_L}     ${2SRC_PORT_U}
    ...    ${TCP_FLAGS_MASK}    ${TCP_FLAGS_VALUE}

Delete ACL2
    Delete ACL     agent_vpp_1    ${ACL2_NAME}

Check ACL2 Is Deleted and ACL1 Is Not Changed
    Check ACL TCP    agent_vpp_1    ${ACL1_NAME}
    ...    ${E_INTF1}    ${I_INTF1}    ${ACTION_PERMIT}
    ...    ${DEST_NTW}     ${SRC_NTW}
    ...    ${1DEST_PORT_L}   ${1DEST_PORT_U}
    ...    ${1SRC_PORT_L}     ${1SRC_PORT_U}
    ...    ${TCP_FLAGS_MASK}    ${TCP_FLAGS_VALUE}
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    ACL in VPP should not exist    agent_vpp_1    ${ACL2_NAME}

Delete ACL1
    Delete ACL     agent_vpp_1    ${ACL1_NAME}

Check ACL1 Is Deleted
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    ACL in VPP should not exist    agent_vpp_1    ${ACL1_NAME}

ADD ACL3_UDP
    Put ACL UDP    agent_vpp_1    ${ACL3_NAME}
    ...    ${E_INTF1}   ${I_INTF1}    ${E_INTF2}    ${I_INTF2}      ${ACTION_DENY}
    ...    ${DEST_NTW}     ${SRC_NTW}
    ...    ${1DEST_PORT_L}   ${1DEST_PORT_U}
    ...    ${1SRC_PORT_L}     ${1SRC_PORT_U}

Check ACL3 Is Created
    Check ACL UDP    agent_vpp_1    ${ACL3_NAME}
    ...    ${E_INTF1}   ${I_INTF1}    ${E_INTF2}    ${I_INTF2}      ${ACTION_DENY}
    ...    ${DEST_NTW}     ${SRC_NTW}
    ...    ${1DEST_PORT_L}   ${1DEST_PORT_U}
    ...    ${1SRC_PORT_L}     ${1SRC_PORT_U}

ADD ACL4_UDP
    Put ACL UDP    agent_vpp_1    ${ACL4_NAME}
    ...    ${E_INTF1}    ${I_INTF1}    ${E_INTF2}    ${I_INTF2}      ${ACTION_DENY}
    ...    ${DEST_NTW}     ${SRC_NTW}
    ...    ${1DEST_PORT_L}   ${1DEST_PORT_U}
    ...    ${1SRC_PORT_L}     ${1SRC_PORT_U}


Check ACL4 Is Created And ACL3 Still Configured
    Check ACL UDP    agent_vpp_1    ${ACL4_NAME}
    ...    ${E_INTF1}    ${I_INTF1}    ${E_INTF2}    ${I_INTF2}      ${ACTION_DENY}
    ...    ${DEST_NTW}     ${SRC_NTW}
    ...    ${1DEST_PORT_L}   ${1DEST_PORT_U}
    ...    ${1SRC_PORT_L}     ${1SRC_PORT_U}
    Check ACL UDP    agent_vpp_1    ${ACL3_NAME}
    ...    ${E_INTF1}    ${I_INTF1}    ${E_INTF2}    ${I_INTF2}      ${ACTION_DENY}
    ...    ${DEST_NTW}     ${SRC_NTW}
    ...    ${1DEST_PORT_L}   ${1DEST_PORT_U}
    ...    ${1SRC_PORT_L}     ${1SRC_PORT_U}

Delete ACL4
    Delete ACL     agent_vpp_1    ${ACL4_NAME}

Check ACL4 Is Deleted and ACL3 Is Not Changed
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    ACL in VPP should not exist    agent_vpp_1    ${ACL4_NAME}
    Check ACL UDP    agent_vpp_1    ${ACL3_NAME}
    ...    ${E_INTF1}    ${I_INTF1}    ${E_INTF2}    ${I_INTF2}      ${ACTION_DENY}
    ...    ${DEST_NTW}     ${SRC_NTW}
    ...    ${1DEST_PORT_L}   ${1DEST_PORT_U}
    ...    ${1SRC_PORT_L}     ${1SRC_PORT_U}

Delete ACL3
    Delete ACL     agent_vpp_1    ${ACL3_NAME}

Check ACL3 Is Deleted
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    ACL in VPP should not exist    agent_vpp_1    ${ACL3_NAME}

ADD ACL5_ICMP
    Put ACL ICMP    agent_vpp_1    ${ACL5_NAME}
    ...    ${E_INTF1}    ${I_INTF1}    ${ACTION_DENY}
    ...    ${DEST_NTW}     ${SRC_NTW}
    ...    ${ICMP_v4}
    ...    ${ICMP_CODE_L}   ${ICMP_CODE_U}
    ...    ${ICMP_TYPE_L}   ${ICMP_TYPE_U}

Check ACL5 Is Created
    Check ACL ICMP    agent_vpp_1    ${ACL5_NAME}
    ...    ${E_INTF1}    ${I_INTF1}    ${ACTION_DENY}
    ...    ${DEST_NTW}     ${SRC_NTW}
    ...    ${ICMP_v4}
    ...    ${ICMP_CODE_L}   ${ICMP_CODE_U}
    ...    ${ICMP_TYPE_L}   ${ICMP_TYPE_U}

ADD ACL6_ICMP
    Put ACL ICMP    agent_vpp_1    ${ACL6_NAME}
    ...    ${E_INTF1}    ${I_INTF1}    ${ACTION_DENY}
    ...    ${DEST_NTW}     ${SRC_NTW}
    ...    ${ICMP_v4}
    ...    ${ICMP_CODE_L}   ${ICMP_CODE_U}
    ...    ${ICMP_TYPE_L}   ${ICMP_TYPE_U}

Check ACL6 Is Created And ACL5 Still Configured
    Check ACL ICMP    agent_vpp_1    ${ACL5_NAME}
    ...    ${E_INTF1}    ${I_INTF1}    ${ACTION_DENY}
    ...    ${DEST_NTW}     ${SRC_NTW}
    ...    ${ICMP_v4}
    ...    ${ICMP_CODE_L}   ${ICMP_CODE_U}
    ...    ${ICMP_TYPE_L}   ${ICMP_TYPE_U}
    Check ACL ICMP    agent_vpp_1    ${ACL6_NAME}
    ...    ${E_INTF1}    ${I_INTF1}    ${ACTION_DENY}
    ...    ${DEST_NTW}     ${SRC_NTW}
    ...    ${ICMP_v4}
    ...    ${ICMP_CODE_L}   ${ICMP_CODE_U}
    ...    ${ICMP_TYPE_L}   ${ICMP_TYPE_U}

Delete ACL6
    Delete ACL     agent_vpp_1    ${ACL6_NAME}

Check ACL6 Is Deleted and ACL5 Is Not Changed
    Check ACL ICMP    agent_vpp_1    ${ACL5_NAME}
    ...    ${E_INTF1}    ${I_INTF1}    ${ACTION_DENY}
    ...    ${DEST_NTW}     ${SRC_NTW}
    ...    ${ICMP_v4}
    ...    ${ICMP_CODE_L}   ${ICMP_CODE_U}
    ...    ${ICMP_TYPE_L}   ${ICMP_TYPE_U}
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    ACL in VPP should not exist    agent_vpp_1    ${ACL6_NAME}

Delete ACL5
    Delete ACL     agent_vpp_1    ${ACL5_NAME}

Check ACL5 Is Deleted
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    ACL in VPP should not exist    agent_vpp_1    ${ACL5_NAME}

Add 6 ACLs
    Put ACL TCP   agent_vpp_1    ${ACL1_NAME}
    ...    ${E_INTF1}    ${I_INTF1}    ${ACTION_DENY}
    ...    ${DEST_NTW}     ${SRC_NTW}
    ...    ${1DEST_PORT_L}   ${1DEST_PORT_U}
    ...    ${1SRC_PORT_L}     ${1SRC_PORT_U}
    ...    ${TCP_FLAGS_MASK}    ${TCP_FLAGS_VALUE}
    Put ACL TCP   agent_vpp_1    ${ACL2_NAME}
    ...    ${E_INTF1}    ${I_INTF1}    ${ACTION_DENY}
    ...    ${DEST_NTW}     ${SRC_NTW}
    ...    ${2DEST_PORT_L}   ${2DEST_PORT_U}
    ...    ${2SRC_PORT_L}     ${2SRC_PORT_U}
    ...    ${TCP_FLAGS_MASK}    ${TCP_FLAGS_VALUE}
    Put ACL UDP    agent_vpp_1    ${ACL3_NAME}
    ...    ${E_INTF1}   ${I_INTF1}    ${E_INTF2}    ${I_INTF2}      ${ACTION_DENY}
    ...    ${DEST_NTW}     ${SRC_NTW}
    ...    ${1DEST_PORT_L}   ${1DEST_PORT_U}
    ...    ${1SRC_PORT_L}     ${1SRC_PORT_U}
    Put ACL UDP    agent_vpp_1    ${ACL4_NAME}
    ...    ${E_INTF1}    ${I_INTF1}    ${E_INTF2}    ${I_INTF2}      ${ACTION_DENY}
    ...    ${DEST_NTW}     ${SRC_NTW}
    ...    ${1DEST_PORT_L}   ${1DEST_PORT_U}
    ...    ${1SRC_PORT_L}     ${1SRC_PORT_U}
    Put ACL ICMP    agent_vpp_1    ${ACL5_NAME}
    ...    ${E_INTF1}    ${I_INTF1}    ${ACTION_DENY}
    ...    ${DEST_NTW}     ${SRC_NTW}
    ...    ${ICMP_v4}
    ...    ${ICMP_CODE_L}   ${ICMP_CODE_U}
    ...    ${ICMP_TYPE_L}   ${ICMP_TYPE_U}
    Put ACL ICMP    agent_vpp_1    ${ACL6_NAME}
    ...    ${E_INTF1}    ${I_INTF1}    ${ACTION_DENY}
    ...    ${DEST_NTW}     ${SRC_NTW}
    ...    ${ICMP_v4}
    ...    ${ICMP_CODE_L}   ${ICMP_CODE_U}
    ...    ${ICMP_TYPE_L}   ${ICMP_TYPE_U}

Check All 6 ACLs Added
    Check ACL TCP   agent_vpp_1    ${ACL1_NAME}
    ...    ${E_INTF1}    ${I_INTF1}    ${ACTION_DENY}
    ...    ${DEST_NTW}     ${SRC_NTW}
    ...    ${1DEST_PORT_L}   ${1DEST_PORT_U}
    ...    ${1SRC_PORT_L}     ${1SRC_PORT_U}
    ...    ${TCP_FLAGS_MASK}    ${TCP_FLAGS_VALUE}
    Check ACL TCP   agent_vpp_1    ${ACL2_NAME}
    ...    ${E_INTF1}    ${I_INTF1}    ${ACTION_DENY}
    ...    ${DEST_NTW}     ${SRC_NTW}
    ...    ${2DEST_PORT_L}   ${2DEST_PORT_U}
    ...    ${2SRC_PORT_L}     ${2SRC_PORT_U}
    ...    ${TCP_FLAGS_MASK}    ${TCP_FLAGS_VALUE}
    Check ACL UDP    agent_vpp_1    ${ACL3_NAME}
    ...    ${E_INTF1}   ${I_INTF1}    ${E_INTF2}    ${I_INTF2}      ${ACTION_DENY}
    ...    ${DEST_NTW}     ${SRC_NTW}
    ...    ${1DEST_PORT_L}   ${1DEST_PORT_U}
    ...    ${1SRC_PORT_L}     ${1SRC_PORT_U}
    Check ACL UDP    agent_vpp_1    ${ACL4_NAME}
    ...    ${E_INTF1}    ${I_INTF1}    ${E_INTF2}    ${I_INTF2}      ${ACTION_DENY}
    ...    ${DEST_NTW}     ${SRC_NTW}
    ...    ${1DEST_PORT_L}   ${1DEST_PORT_U}
    ...    ${1SRC_PORT_L}     ${1SRC_PORT_U}
    Check ACL ICMP    agent_vpp_1    ${ACL5_NAME}
    ...    ${E_INTF1}    ${I_INTF1}    ${ACTION_DENY}
    ...    ${DEST_NTW}     ${SRC_NTW}
    ...    ${ICMP_v4}
    ...    ${ICMP_CODE_L}   ${ICMP_CODE_U}
    ...    ${ICMP_TYPE_L}   ${ICMP_TYPE_U}
    Check ACL ICMP    agent_vpp_1    ${ACL6_NAME}
    ...    ${E_INTF1}    ${I_INTF1}    ${ACTION_DENY}
    ...    ${DEST_NTW}     ${SRC_NTW}
    ...    ${ICMP_v4}
    ...    ${ICMP_CODE_L}   ${ICMP_CODE_U}
    ...    ${ICMP_TYPE_L}   ${ICMP_TYPE_U}

# TODO: add tests for MACIP case
# TODO: test ingress/egress interfaces

*** Keywords ***

TestSetup
    Make Datastore Snapshots    ${TEST_NAME}_test_setup

TestTeardown
    Make Datastore Snapshots    ${TEST_NAME}_test_teardown

Suite Cleanup
    Stop SFC Controller Container
    Testsuite Teardown

Check ACL TCP
    [Arguments]    ${agent}    ${acl_name}
    ...    ${egress_interface}    ${ingress_interface}    ${acl_action}
    ...    ${destination_network}     ${source_network}
    ...    ${destination_port_min}   ${destination_port_max}
    ...    ${source_port_min}     ${source_port_max}
    ...    ${tcp_flags_mask}    ${tcp_flags_value}
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    Check ACL in ETCD - TCP    ${agent}    ${acl_name}
    ...    ${egress_interface}    ${ingress_interface}    ${acl_action}
    ...    ${destination_network}     ${source_network}
    ...    ${destination_port_min}   ${destination_port_max}
    ...    ${source_port_min}     ${source_port_max}
    ...    ${tcp_flags_mask}    ${tcp_flags_value}
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    Check ACL in VPP - TCP    ${agent}    ${acl_name}
    ...    ${egress_interface}    ${ingress_interface}    ${acl_action}
    ...    ${destination_network}     ${source_network}
    ...    ${destination_port_min}   ${destination_port_max}
    ...    ${source_port_min}     ${source_port_max}
    ...    ${tcp_flags_mask}    ${tcp_flags_value}

Check ACL UDP
    [Arguments]    ${agent}    ${acl_name}
    ...    ${egress_interface1}    ${ingress_interface1}
    ...    ${egress_interface2}    ${ingress_interface2}
    ...    ${acl_action}
    ...    ${destination_network}     ${source_network}
    ...    ${destination_port_min}   ${destination_port_max}
    ...    ${source_port_min}     ${source_port_max}
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    Check ACL in ETCD - UDP    ${agent}    ${acl_name}
    ...    ${egress_interface1}    ${ingress_interface1}
    ...    ${egress_interface2}    ${ingress_interface2}
    ...    ${acl_action}
    ...    ${destination_network}     ${source_network}
    ...    ${destination_port_min}   ${destination_port_max}
    ...    ${source_port_min}     ${source_port_max}
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    Check ACL in VPP - UDP    ${agent}    ${acl_name}
    ...    ${egress_interface1}    ${ingress_interface1}
    ...    ${egress_interface2}    ${ingress_interface2}
    ...    ${acl_action}
    ...    ${destination_network}     ${source_network}
    ...    ${destination_port_min}   ${destination_port_max}
    ...    ${source_port_min}     ${source_port_max}

Check ACL ICMP
    [Arguments]    ${agent}    ${acl_name}
    ...    ${egress_interface}    ${ingress_interface}    ${acl_action}
    ...    ${destination_network}     ${source_network}
    ...    ${icmpv6}
    ...    ${icmp_code_min}   ${icmp_code_max}
    ...    ${icmp_type_min}   ${icmp_type_max}
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    Check ACL in ETCD - ICMP    ${agent}    ${acl_name}
    ...    ${egress_interface}    ${ingress_interface}    ${acl_action}
    ...    ${destination_network}     ${source_network}
    ...    ${icmpv6}
    ...    ${icmp_code_min}   ${icmp_code_max}
    ...    ${icmp_type_min}   ${icmp_type_max}
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    Check ACL in VPP - ICMP    ${agent}    ${acl_name}
    ...    ${egress_interface}    ${ingress_interface}    ${acl_action}
    ...    ${destination_network}     ${source_network}
    ...    ${icmpv6}
    ...    ${icmp_code_min}   ${icmp_code_max}
    ...    ${icmp_type_min}   ${icmp_type_max}