[Documentation]     ACL-related keywords for working with VPP terminal

*** Settings ***
Library      Collections

Library      ../vpp_term.py
Library      acl_utils.py

*** Variables ***

*** Keywords ***

vpp_term: ACL Dump
    [Arguments]        ${node}
    [Documentation]    Executing command acl_dump
    ${out}=            vpp_term: Issue Command  ${node}  show acl-plugin acl
    ${out_data_vpp}=   Strip String     ${out}
    ${out_data}=       Remove String     ${out_data_vpp}    vpp#${SPACE}   vpp#
    [Return]           ${out_data}

Check ACL in VPP - TCP
    [Arguments]    ${node}    ${acl_name}    ${egr_intf1}   ${ingr_intf1}
    ...    ${acl_action}    ${dest_ntw}    ${src_ntw}    ${dest_port_low}
    ...    ${dest_port_up}    ${src_port_low}    ${src_port_up}
    ...    ${tcp_flags_mask}    ${tcp_flags_value}
    [Documentation]
    ...    Get ACL data from VPP terminal
    ...    and verify response against expected data.
    ${term_d}=          vpp_term: ACL Dump     ${node}
    ${term_d}=          Filter ACL By Name      ${term_d}    ${acl_name}
    ${data}=            OperatingSystem.Get File      ${CURDIR}/../../resources/acl/acl_TCP_response_term.txt
    ${data}=            Replace ACL Variables TCP    ${data}    ${acl_name}
    ...    ${egr_intf1}   ${ingr_intf1}
    ...    ${acl_action}
    ...    ${dest_ntw}    ${src_ntw}
    ...    ${dest_port_low}    ${dest_port_up}
    ...    ${src_port_low}    ${src_port_up}
    ...    ${tcp_flags_mask}    ${tcp_flags_value}
    Should Be Equal    ${term_d}    ${data}

Check ACL in VPP - UDP
    [Arguments]    ${node}    ${acl_name}    ${egr_intf1}   ${ingr_intf1}
    ...    ${egr_intf2}   ${ingr_intf2}
    ...    ${acl_action}    ${dest_ntw}    ${src_ntw}    ${dest_port_low}
    ...    ${dest_port_up}    ${src_port_low}    ${src_port_up}
    [Documentation]
    ...    Get ACL data from VPP terminal
    ...    and verify response against expected data.
    ${term_d}=          vpp_term: ACL Dump     ${node}
    ${term_d}=          Filter ACL By Name      ${term_d}    ${acl_name}
    ${data}=            OperatingSystem.Get File     ${CURDIR}/../../resources/acl/acl_UDP_response_term.txt
    ${data}=            Replace ACL Variables UDP    ${data}    ${acl_name}
    ...    ${egr_intf1}   ${ingr_intf1}
    ...    ${acl_action}
    ...    ${dest_ntw}    ${src_ntw}
    ...    ${dest_port_low}    ${dest_port_up}
    ...    ${src_port_low}    ${src_port_up}
    Should Be Equal    ${term_d}    ${data}

Check ACL in VPP - ICMP
    [Arguments]    ${node}    ${acl_name}    ${egr_intf1}   ${ingr_intf1}
    ...    ${acl_action}    ${dest_ntw}    ${src_ntw}
    ...    ${icmpv6}
    ...    ${icmp_range_low}    ${icmp_range_high}
    ...    ${icmpv6_range_low}    ${icmpv6_range_high}
    [Documentation]
    ...    Get ACL data from VPP terminal
    ...    and verify response against expected data.
    ${term_d}=          vpp_term: ACL Dump     ${node}
    ${term_d}=          Filter ACL By Name      ${term_d}    ${acl_name}
    ${data}=            OperatingSystem.Get File     ${CURDIR}/../../resources/acl/acl_ICMP_response_term.txt
    ${data}=            Replace ACL Variables ICMP    ${data}    ${acl_name}
    ...    ${egr_intf1}   ${ingr_intf1}
    ...    ${acl_action}
    ...    ${dest_ntw}    ${src_ntw}
    ...    ${icmpv6}
    ...    ${icmp_range_low}    ${icmp_range_high}
    ...    ${icmpv6_range_low}    ${icmpv6_range_high}
    Should Be Equal    ${term_d}    ${data}

vpp_term: Show ACL
    [Arguments]        ${node}
    [Documentation]    Show ACLs through vpp terminal
    ${out}=            vpp_term: Issue Command  ${node}   sh acl-plugin acl
    [Return]           ${out}

ACL in VPP should not exist
    [Arguments]    ${node}    ${acl_name}
    ${term_d}=          vpp_term: ACL Dump     ${node}
    Run Keyword And Expect Error
    ...    ACL name ${acl_name} not found in response data.
    ...    Filter ACL By Name    ${term_d}    ${acl_name}
