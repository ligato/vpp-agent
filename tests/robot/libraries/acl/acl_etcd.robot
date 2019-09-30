[Documentation]     ACL-related keywords for working with ETCD

*** Settings ***
Library        String

Library        ../etcdctl.py
Library        ../vpp_term.py

Resource       ../etcdctl.robot

*** Variables ***

*** Keywords ***

Check ACL Reply
    [Arguments]         ${node}    ${acl_name}   ${reply_json}    ${reply_term}    ${api_h}=$(API_HANDLER}
    [Documentation]     Get ACL data from VAT terminal and verify response against expected data.
    ${acl_d}=           Get ACL As Json    ${node}    ${acl_name}
    ${term_d}=          vpp_term: Check ACL     ${node}    ${acl_name}
    ${data}=            OperatingSystem.Get File    ${reply_json}
    ${data}=            Replace Variables      ${data}
    Should Be Equal     ${data}   ${acl_d}
    ${data}=            OperatingSystem.Get File    ${reply_term}
    ${data}=            Replace Variables      ${data}
    Should be Equal     ${term_d}    ${data}

Check ACL in ETCD - TCP
    [Arguments]    ${node}    ${acl_name}    ${egr_intf1}   ${ingr_intf1}
    ...    ${acl_action}    ${dest_ntw}    ${src_ntw}    ${dest_port_low}
    ...    ${dest_port_up}    ${src_port_low}    ${src_port_up}
    ...    ${tcp_flags_mask}    ${tcp_flags_value}
    [Documentation]
    ...    Get ACL data from ETCD
    ...    and verify response against expected data.
    ${acl_d}=           Get ACL As Json    ${node}    ${acl_name}
    ${data}=            OperatingSystem.Get File      ${CURDIR}/../../resources/acl/acl_TCP.json
    ${data}=            Replace Variables             ${data}
    ${data}=            Strip String    ${data}
    ${acl_d}=           Strip String    ${data}
    Should Be Equal As Strings     ${data}   ${acl_d}

Check ACL in ETCD - UDP
    [Arguments]    ${node}    ${acl_name}
    ...    ${egr_intf1}   ${ingr_intf1}    ${egr_intf2}   ${ingr_intf2}
    ...    ${acl_action}    ${dest_ntw}    ${src_ntw}    ${dest_port_low}
    ...    ${dest_port_up}    ${src_port_low}    ${src_port_up}
    [Documentation]
    ...    Get ACL data from ETCD
    ...    and verify response against expected data.
    ${acl_d}=           Get ACL As Json    ${node}    ${acl_name}
    ${data}=            OperatingSystem.Get File      ${CURDIR}/../../resources/acl/acl_UDP.json
    ${data}=            Replace Variables             ${data}
    ${data}=            Strip String    ${data}
    ${acl_d}=           Strip String    ${data}
    Should Be Equal As Strings     ${data}   ${acl_d}

Check ACL in ETCD - ICMP
    [Arguments]    ${node}    ${acl_name}    ${egr_intf1}   ${ingr_intf1}
    ...    ${acl_action}    ${dest_ntw}    ${src_ntw}
    ...    ${icmpv6}   ${code_range_low}   ${code_range_up}    ${type_range_low}   ${type_range_up}
    [Documentation]
    ...    Get ACL data from ETCD
    ...    and verify response against expected data.
    ${acl_d}=           Get ACL As Json    ${node}    ${acl_name}
    ${data}=            OperatingSystem.Get File      ${CURDIR}/../../resources/acl/acl_ICMP.json
    ${data}=            Replace Variables             ${data}
    ${data}=            Strip String    ${data}
    ${acl_d}=           Strip String    ${data}
    Should Be Equal As Strings     ${data}   ${acl_d}

Put ACL TCP
    [Arguments]    ${node}    ${acl_name}    ${egr_intf1}   ${ingr_intf1}
    ...    ${acl_action}    ${dest_ntw}    ${src_ntw}    ${dest_port_low}
    ...    ${dest_port_up}    ${src_port_low}    ${src_port_up}
    ...    ${tcp_flags_mask}    ${tcp_flags_value}
    ${data}=              OperatingSystem.Get File      ${CURDIR}/../../resources/acl/acl_TCP.json
    ${uri}=               Set Variable          /vnf-agent/${node}/config/vpp/acls/${AGENT_VER}/acl/${acl_name}
    ${data}=              Replace Variables             ${data}
    #OperatingSystem.Create File   ${REPLY_DATA_FOLDER}/reply.json     ${data}
    Put Json     ${uri}    ${data}

Put ACL UDP
    [Arguments]    ${node}    ${acl_name}    ${egr_intf1}    ${ingr_intf1}     ${egr_intf2}    ${ingr_intf2}     ${acl_action}    ${dest_ntw}   ${src_ntw}    ${dest_port_low}   ${dest_port_up}    ${src_port_low}    ${src_port_up}
    ${data}=              OperatingSystem.Get File      ${CURDIR}/../../resources/acl/acl_UDP.json
    ${uri}=               Set Variable          /vnf-agent/${node}/config/vpp/acls/${AGENT_VER}/acl/${acl_name}
    ${data}=              Replace Variables             ${data}
    #OperatingSystem.Create File   ${REPLY_DATA_FOLDER}/reply.json     ${data}
    Put Json     ${uri}    ${data}

Put ACL MACIP
    [Arguments]    ${node}    ${acl_name}    ${egr_intf1}    ${ingr_intf1}    ${acl_action}    ${src_addr}    ${src_addr_prefix}    ${src_mac_addr}   ${src_mac_addr_mask}
    ${data}=              OperatingSystem.Get File      ${CURDIR}/../../resources/acl/acl_MACIP.json
    ${uri}=               Set Variable          /vnf-agent/${node}/config/vpp/acls/${AGENT_VER}/acl/${acl_name}
    ${data}=              Replace Variables             ${data}
    #OperatingSystem.Create File   ${REPLY_DATA_FOLDER}/reply.json     ${data}
    Put Json     ${uri}    ${data}

Put ACL ICMP
    [Arguments]    ${node}    ${acl_name}    ${egr_intf1}   ${ingr_intf1}    ${acl_action}   ${dest_ntw}    ${src_ntw}    ${icmpv6}   ${code_range_low}   ${code_range_up}    ${type_range_low}   ${type_range_up}
    ${data}=              OperatingSystem.Get File      ${CURDIR}/../../resources/acl/acl_ICMP.json
    ${uri}=               Set Variable          /vnf-agent/${node}/config/vpp/acls/${AGENT_VER}/acl/${acl_name}
    ${data}=              Replace Variables             ${data}
    #OperatingSystem.Create File   ${REPLY_DATA_FOLDER}/reply.json     ${data}
    Put Json     ${uri}    ${data}

Get ACL As Json
    [Arguments]           ${node}  ${acl_name}
    ${key}=               Set Variable          /vnf-agent/${node}/config/vpp/acls/${AGENT_VER}/acl/${acl_name}
    ${data}=              Read Key    ${key}
    ${data}=              Set Variable If      '''${data}'''=="" or '''${data}'''=='None'    {}    ${data}
    [Return]              ${data}

Get All ACL As Json
    [Arguments]           ${node}
    ${key}=               Set Variable          /vnf-agent/${node}/config/vpp/acls/${AGENT_VER}/acl
    #${data}=              etcd: Get ETCD Tree    ${key}
    ${data}=              Read Key    ${key}    true
    ${data}=              Set Variable If      '''${data}'''=="" or '''${data}'''=='None'    {}    ${data}
    [Return]              ${data}

Delete ACL
    [Arguments]    ${node}    ${name}
    ${uri}=      Set Variable    /vnf-agent/${node}/config/vpp/acls/${AGENT_VER}/acl/${name}
    ${out}=      Delete key    ${uri}
    [Return]    ${out}