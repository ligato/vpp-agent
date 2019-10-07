[Documentation]     ACL-related keywords for working with VPP terminal

*** Settings ***
Library      Collections

Library      ../vpp_term.py
Library      acl_utils.py

*** Variables ***

*** Keywords ***

vpp_term: ACL Dump
    [Arguments]        ${node}
    [Documentation]    Execute command acl_dump
    ${out}=            vpp_term: Issue Command  ${node}  show acl-plugin acl
    ${out_data_vpp}=   Strip String     ${out}
    ${out_data}=       Remove String     ${out_data_vpp}    vpp#${SPACE}   vpp#
    [Return]           ${out_data}

Check ACL in VPP - TCP
    [Arguments]    ${node}    ${acl_name}    ${egr_intf1}   ${ingr_intf1}
    ...    ${acl_action}    ${dest_ntw}    ${src_ntw}    ${dest_port_low}
    ...    ${dest_port_high}    ${src_port_low}    ${src_port_high}
    ...    ${tcp_flags_mask}    ${tcp_flags_value}
    [Documentation]
    ...    Get ACL data from VPP API
    ...    and verify response against expected data.
    ${api_dump_list}=    vpp_api: ACL Dump                     ${node}
    ${api_dump}=         Filter ACL Dump By Name               ${api_dump_list}    ${acl_name}
    Should Be Equal                ${api_dump["acl_name"]}                ${acl_name}
    Should Be Equal As Integers    ${api_dump["acl_action"]}              ${acl_action}
    Should Be Equal As Integers    ${api_dump["protocol"]}                6
    Should Be Equal                ${api_dump["destination_network"]}     ${dest_ntw}
    Should Be Equal                ${api_dump["source_network"]}          ${src_ntw}
    Should Be Equal As Integers    ${api_dump["destination_port_low"]}    ${dest_port_low}
    Should Be Equal As Integers    ${api_dump["destination_port_high"]}   ${dest_port_high}
    Should Be Equal As Integers    ${api_dump["source_port_low"]}         ${src_port_low}
    Should Be Equal As Integers    ${api_dump["source_port_high"]}        ${src_port_high}
    Should Be Equal As Integers    ${api_dump["tcp_flags_mask"]}          ${tcp_flags_mask}
    Should Be Equal As Integers    ${api_dump["tcp_flags_value"]}         ${tcp_flags_value}

Check ACL in VPP - UDP
    [Arguments]    ${node}    ${acl_name}    ${egr_intf1}   ${ingr_intf1}
    ...    ${egr_intf2}   ${ingr_intf2}
    ...    ${acl_action}    ${dest_ntw}    ${src_ntw}    ${dest_port_low}
    ...    ${dest_port_high}    ${src_port_low}    ${src_port_high}
    [Documentation]
    ...    Get ACL data from VPP API
    ...    and verify response against expected data.
    ${api_dump_list}=    vpp_api: ACL Dump     ${node}
    ${api_dump}=         Filter ACL Dump By Name    ${api_dump_list}      ${acl_name}
    Should Be Equal                ${api_dump["acl_name"]}                ${acl_name}
    Should Be Equal As Integers    ${api_dump["acl_action"]}              ${acl_action}
    Should Be Equal As Integers    ${api_dump["protocol"]}                17
    Should Be Equal                ${api_dump["destination_network"]}     ${dest_ntw}
    Should Be Equal                ${api_dump["source_network"]}          ${src_ntw}
    Should Be Equal As Integers    ${api_dump["destination_port_low"]}    ${dest_port_low}
    Should Be Equal As Integers    ${api_dump["destination_port_high"]}   ${dest_port_high}
    Should Be Equal As Integers    ${api_dump["source_port_low"]}         ${src_port_low}
    Should Be Equal As Integers    ${api_dump["source_port_high"]}        ${src_port_high}

Check ACL in VPP - ICMP
    [Arguments]    ${node}    ${acl_name}    ${egr_intf1}   ${ingr_intf1}
    ...    ${acl_action}    ${dest_ntw}    ${src_ntw}
    ...    ${icmpv6}
    ...    ${icmp_code_low}    ${icmp_code_high}
    ...    ${icmp_type_low}    ${icmp_type_high}
    [Documentation]
    ...    Get ACL data from VPP API
    ...    and verify response against expected data.
    ${protocol}=    Set Variable If    "${icmpv6}" == "true"    58    1
    ${api_dump_list}=    vpp_api: ACL Dump     ${node}
    ${api_dump}=         Filter ACL Dump By Name    ${api_dump_list}    ${acl_name}
    Should Be Equal                ${api_dump["acl_name"]}                ${acl_name}
    Should Be Equal As Integers    ${api_dump["acl_action"]}              ${acl_action}
    Should Be Equal As Integers    ${api_dump["protocol"]}                ${protocol}
    Should Be Equal                ${api_dump["destination_network"]}     ${dest_ntw}
    Should Be Equal                ${api_dump["source_network"]}          ${src_ntw}
    Should Be Equal As Integers    ${api_dump["icmp_code_low"]}           ${icmp_code_low}
    Should Be Equal As Integers    ${api_dump["icmp_code_high"]}          ${icmp_code_high}
    Should Be Equal As Integers    ${api_dump["icmp_type_low"]}           ${icmp_type_low}
    Should Be Equal As Integers    ${api_dump["icmp_type_high"]}          ${icmp_type_high}

ACL in VPP should not exist
    [Arguments]    ${node}    ${acl_name}
    ${term_d}=          vpp_api: ACL Dump     ${node}
    Run Keyword And Expect Error
    ...    ACL not found by name ${acl_name}.
    ...    Filter ACL Dump By Name    ${term_d}    ${acl_name}

vpp_api: ACL Dump
    [Arguments]        ${node}
    [Documentation]    Executing command acl_dump
    ${out}=            ACL Dump    ${DOCKER_HOST_IP}    ${DOCKER_HOST_USER}    ${DOCKER_HOST_PSWD}    ${node}
    [Return]           ${out}
