[Documentation]     Keywords for working with VPP Ctl container

*** Settings ***
Library        vpp_ctl.py
Library        String

*** Variables ***

*** Keywords ***

vpp_ctl: Put Json
    [Arguments]        ${key}    ${json}    ${container}=vpp_agent_ctl
    Log Many           ${key}    ${json}    ${container}
    ${command}=        Set Variable    echo '${json}' | vpp-agent-ctl ${AGENT_VPP_ETCD_CONF_PATH} -put ${key} -
    ${out}=            Write To Container Until Prompt    ${container}    ${command}
    [Return]           ${out}

vpp_ctl: Read Key
    [Arguments]        ${key}    ${container}=vpp_agent_ctl
    Log Many           ${key}    ${container}
    ${command}=        Set Variable    vpp-agent-ctl ${AGENT_VPP_ETCD_CONF_PATH} -get ${key}
#    ${out}=            Write To Container Until Prompt    ${container}    ${command}
    ${out}=            Execute In Container    ${container}    ${command}
    [Return]           ${out}

vpp_ctl: Read Key With Prefix
    [Arguments]        ${key}    ${container}=vpp_agent_ctl
    Log Many           ${key}    ${container}
    ${command}=        Set Variable    vpp-agent-ctl ${AGENT_VPP_ETCD_CONF_PATH} -get ${key}
#    ${out}=            Write To Container Until Prompt    ${container}    ${command}
    ${out}=            Execute In Container    ${container}    ${command}
    [Return]           ${out}

vpp_ctl: Put Memif Interface
    [Arguments]    ${node}    ${name}    ${mac}    ${master}    ${id}    ${socket}=memif.sock    ${mtu}=1500    ${vrf}=0    ${enabled}=true
    Log Many    ${node}    ${name}    ${mac}    ${master}    ${id}    ${socket}    ${mtu}    ${vrf}    ${enabled}
    ${socket}=            Set Variable                  ${${node}_MEMIF_SOCKET_FOLDER}/${socket}
    Log                   ${socket}
    ${data}=              OperatingSystem.Get File      ${CURDIR}/../resources/memif_interface.json
    ${uri}=               Set Variable                  /vnf-agent/${node}/vpp/config/v1/interface/${name}
    Log Many              ${data}                       ${uri}
    ${data}=              Replace Variables             ${data}
    Log                   ${data}
    vpp_ctl: Put Json     ${uri}    ${data}

vpp_ctl: Put Memif Interface With IP
    [Arguments]    ${node}    ${name}    ${mac}    ${master}    ${id}    ${ip}    ${prefix}=24    ${socket}=memif.sock    ${mtu}=1500    ${vrf}=0    ${enabled}=true
    Log Many    ${node}    ${name}    ${mac}    ${master}    ${id}    ${ip}    ${prefix}    ${socket}    ${mtu}    ${vrf}    ${enabled}
    ${socket}=            Set Variable                  ${${node}_MEMIF_SOCKET_FOLDER}/${socket}
    Log                   ${socket}
    ${data}=              OperatingSystem.Get File      ${CURDIR}/../resources/memif_interface_with_ip.json
    ${uri}=               Set Variable                  /vnf-agent/${node}/vpp/config/v1/interface/${name}
    Log Many              ${data}                       ${uri}
    ${data}=              Replace Variables             ${data}
    Log                   ${data}
    vpp_ctl: Put Json     ${uri}    ${data}

vpp_ctl: Delete key
    [Arguments]     ${key}    ${container}=vpp_agent_ctl
    Log Many        ${container}    ${key}
    ${out}=         Write To Container Until Prompt    ${container}   vpp-agent-ctl ${AGENT_VPP_ETCD_CONF_PATH} -del ${key}
    Log Many        ${out}
    [Return]        ${out}

vpp_ctl: Put Veth Interface
    [Arguments]    ${node}    ${name}    ${mac}    ${peer}    ${enabled}=true
    Log Many    ${node}    ${name}    ${mac}    ${peer}    ${enabled}
    ${data}=              OperatingSystem.Get File      ${CURDIR}/../resources/veth_interface.json
    ${uri}=               Set Variable                  /vnf-agent/${node}/linux/config/v1/interface/${name}
    Log Many              ${data}                       ${uri}
    ${data}=              Replace Variables             ${data}
    Log                   ${data}
    vpp_ctl: Put Json     ${uri}    ${data}

vpp_ctl: Put Veth Interface With IP
    [Arguments]    ${node}    ${name}    ${mac}    ${peer}    ${ip}    ${prefix}=24    ${mtu}=1500    ${vrf}=0    ${enabled}=true
    Log Many    ${node}    ${name}    ${mac}    ${peer}    ${ip}    ${prefix}    ${mtu}    ${vrf}    ${enabled}
    ${data}=              OperatingSystem.Get File      ${CURDIR}/../resources/veth_interface_with_ip.json
    ${uri}=               Set Variable                  /vnf-agent/${node}/linux/config/v1/interface/${name}
    Log Many              ${data}                       ${uri}
    ${data}=              Replace Variables             ${data}
    Log                   ${data}
    vpp_ctl: Put Json     ${uri}    ${data}

vpp_ctl: Put Afpacket Interface
    [Arguments]    ${node}    ${name}    ${mac}    ${host_int}    ${mtu}=1500    ${enabled}=true    ${vrf}=0
    Log Many    ${node}    ${name}    ${mac}    ${host_int}    ${mtu}    ${vrf}    ${enabled}
    ${data}=              OperatingSystem.Get File      ${CURDIR}/../resources/afpacket_interface.json
    ${uri}=               Set Variable                  /vnf-agent/${node}/vpp/config/v1/interface/${name}
    Log Many              ${data}                       ${uri}
    ${data}=              Replace Variables             ${data}
    Log                   ${data}
    vpp_ctl: Put Json     ${uri}    ${data}

vpp_ctl: Put VXLan Interface
    [Arguments]    ${node}    ${name}    ${src}    ${dst}    ${vni}    ${enabled}=true    ${vrf}=0
    Log Many    ${node}    ${name}    ${src}    ${dst}    ${vni}    ${enabled}    ${vrf}
    ${data}=              OperatingSystem.Get File      ${CURDIR}/../resources/vxlan_interface.json
    ${uri}=               Set Variable                  /vnf-agent/${node}/vpp/config/v1/interface/${name}
    Log Many              ${data}                       ${uri}
    ${data}=              Replace Variables             ${data}
    Log                   ${data}
    vpp_ctl: Put Json     ${uri}    ${data}

vpp_ctl: Put Bridge Domain
    [Arguments]    ${node}    ${name}    ${ints}    ${flood}=true    ${unicast}=true    ${forward}=true    ${learn}=true    ${arp_term}=true
    Log Many    ${node}    ${name}    ${ints}    ${flood}    ${unicast}    ${forward}    ${learn}    ${arp_term}
    ${interfaces}=        Create Interfaces Json From List    ${ints}
    Log                   ${interfaces}
    ${data}=              OperatingSystem.Get File      ${CURDIR}/../resources/bridge_domain.json
    ${uri}=               Set Variable                  /vnf-agent/${node}/vpp/config/v1/bd/${name}
    Log Many              ${data}                       ${uri}
    ${data}=              Replace Variables             ${data}
    Log                   ${data}
    vpp_ctl: Put Json     ${uri}    ${data}

vpp_ctl: Put Loopback Interface
    [Arguments]    ${node}    ${name}    ${mac}    ${mtu}=1500    ${enabled}=true
    Log Many    ${node}    ${name}    ${mac}    ${mtu}    ${enabled}
    ${data}=              OperatingSystem.Get File      ${CURDIR}/../resources/loopback_interface.json
    ${uri}=               Set Variable                  /vnf-agent/${node}/vpp/config/v1/interface/${name}
    Log Many              ${data}                       ${uri}
    ${data}=              Replace Variables             ${data}
    Log                   ${data}
    vpp_ctl: Put Json     ${uri}    ${data}

vpp_ctl: Put Loopback Interface With IP
    [Arguments]    ${node}    ${name}    ${mac}    ${ip}    ${prefix}=24    ${mtu}=1500    ${vrf}=0    ${enabled}=true
    Log Many    ${node}    ${name}    ${mac}    ${ip}    ${prefix}    ${mtu}    ${vrf}    ${enabled}
    ${data}=              OperatingSystem.Get File      ${CURDIR}/../resources/loopback_interface_with_ip.json
    ${uri}=               Set Variable                  /vnf-agent/${node}/vpp/config/v1/interface/${name}
    Log Many              ${data}                       ${uri}
    ${data}=              Replace Variables             ${data}
    Log                   ${data}
    vpp_ctl: Put Json     ${uri}    ${data}

vpp_ctl: Put Physical Interface With IP
    [Arguments]    ${node}    ${name}    ${ip}    ${prefix}=24    ${mtu}=1500    ${enabled}=true
    Log Many    ${node}    ${name}    ${ip}    ${prefix}    ${mtu}    ${enabled}
    ${data}=              OperatingSystem.Get File      ${CURDIR}/../resources/physical_interface_with_ip.json
    ${uri}=               Set Variable                  /vnf-agent/${node}/vpp/config/v1/interface/${name}
    Log Many              ${data}                       ${uri}
    ${data}=              Replace Variables             ${data}
    Log                   ${data}
    vpp_ctl: Put Json     ${uri}    ${data}

vpp_ctl: Get VPP Interface State
    [Arguments]    ${node}    ${interface}
    Log Many    ${node}    ${interface}
    ${key}=               Set Variable    /vnf-agent/${node}/vpp/status/v1/interface/${interface}
    Log                   ${key}
    ${out}=               vpp_ctl: Read Key    ${key}
    Log                   ${out}
    [Return]              ${out}

vpp_ctl: Get VPP Interface State As Json
    [Arguments]    ${node}    ${interface}
    Log Many    ${node}    ${interface}
    ${key}=               Set Variable    /vnf-agent/${node}/vpp/status/v1/interface/${interface}
    Log                   ${key}
    ${data}=              vpp_ctl: Read Key    ${key}
    Log                   ${data}
    ${data}=              Set Variable If      '''${data}'''==""    {}    ${data}
    Log                   ${data}
    ${output}=            Evaluate             json.loads('''${data}''')    json
    [Return]              ${output}

vpp_ctl: Get VPP Interface Config As Json
    [Arguments]    ${node}    ${interface}
    Log Many    ${node}    ${interface}
    ${key}=               Set Variable    /vnf-agent/${node}/vpp/config/v1/interface/${interface}
    Log                   ${key}
    ${data}=              vpp_ctl: Read Key    ${key}
    Log                   ${data}
    ${data}=              Set Variable If      '''${data}'''==""    {}    ${data}
    Log                   ${data}
    ${output}=            Evaluate             json.loads('''${data}''')    json
    [Return]              ${output}

vpp_ctl: Get Linux Interface Config As Json
    [Arguments]    ${node}    ${name}
    Log Many    ${node}    ${name}
    ${key}=               Set Variable    /vnf-agent/${node}/linux/config/v1/interface/${name}
    Log                   ${key}
    ${data}=              vpp_ctl: Read Key    ${key}
    Log                   ${data}
    ${data}=              Set Variable If      '''${data}'''==""    {}    ${data}
    Log                   ${data}
    ${output}=            Evaluate             json.loads('''${data}''')    json
    [Return]              ${output}

vpp_ctl: Get Bridge Domain State As Json
    [Arguments]    ${node}    ${bd}
    Log Many    ${node}    ${bd}
    ${key}=               Set Variable    /vnf-agent/${node}/vpp/status/v1/bd/${bd}
    Log                   ${key}
    ${data}=              vpp_ctl: Read Key    ${key}
    Log                   ${data}
    ${data}=              Set Variable If      '''${data}'''==""    {}    ${data}
    Log                   ${data}
    ${output}=            Evaluate             json.loads('''${data}''')    json
    [Return]              ${output}

vpp_ctl: Get Interface Internal Name
    [Arguments]    ${node}    ${interface}
    Log Many    ${node}    ${interface}
    ${name}=    Set Variable      ${EMPTY}
    ${empty_dict}=   Create Dictionary
    ${state}=    vpp_ctl: Get VPP Interface State As Json    ${node}    ${interface}
    Log         ${state}
    ${length}=   Get Length     ${state}
    Log         ${length}
    ${name}=    Run Keyword If      ${length} != 0     Set Variable    ${state["internal_name"]}
    Log    ${name}
    [Return]    ${name}

vpp_ctl: Get Interface Sw If Index
    [Arguments]    ${node}    ${interface}
    Log Many    ${node}    ${interface}
    ${state}=    vpp_ctl: Get VPP Interface State As Json    ${node}    ${interface}
    ${sw_if_index}=    Set Variable    ${state["if_index"]}
    Log    ${sw_if_index}
    [Return]    ${sw_if_index}

vpp_ctl: Get Bridge Domain ID
    [Arguments]    ${node}    ${bd}
    Log Many    ${node}    ${bd}
    ${state}=    vpp_ctl: Get Bridge Domain State As Json    ${node}    ${bd}
    ${bd_id}=    Set Variable    ${state["index"]}
    Log    ${bd_id}
    [Return]    ${bd_id}

vpp_ctl: Put TAP Interface With IP
    [Arguments]    ${node}    ${name}    ${mac}    ${ip}    ${host_if_name}    ${prefix}=24    ${mtu}=1500    ${enabled}=true    ${vrf}=0
    Log Many    ${node}    ${name}    ${mac}    ${ip}    ${host_if_name}    ${prefix}    ${mtu}    ${enabled}    ${vrf}
    ${data}=              OperatingSystem.Get File      ${CURDIR}/../resources/tap_interface_with_ip.json
    ${uri}=               Set Variable                  /vnf-agent/${node}/vpp/config/v1/interface/${name}
    Log Many              ${data}                       ${uri}
    ${data}=              Replace Variables             ${data}
    Log                   ${data}
    vpp_ctl: Put Json     ${uri}    ${data}
    Sleep                 10s    Time to let etcd to get state of newly setup tap interface.

vpp_ctl: Put Static Fib Entry
    [Arguments]    ${node}    ${bd_name}    ${mac}    ${outgoing_interface}    ${static}=true
    Log Many    ${node}    ${bd_name}    ${mac}    ${outgoing_interface}
    ${data}=              OperatingSystem.Get File      ${CURDIR}/../resources/static_fib.json
    ${uri}=               Set Variable                  /vnf-agent/${node}/vpp/config/v1/bd/${bd_name}/fib/${mac}
    Log Many              ${data}                       ${uri}
    ${data}=              Replace Variables             ${data}
    Log                   ${data}
    vpp_ctl: Put Json     ${uri}    ${data}

vpp_ctl: Delete Bridge Domain
    [Arguments]    ${node}    ${name}
    Log Many     ${node}    ${name}
    ${uri}=      Set Variable    /vnf-agent/${node}/vpp/config/v1/bd/${name}
    ${out}=      vpp_ctl: Delete key    ${uri}
    Log Many     ${out}
    [Return]    ${out}

vpp_ctl: Delete VPP Interface
    [Arguments]    ${node}    ${name}
    Log Many     ${node}    ${name}
    ${uri}=      Set Variable    /vnf-agent/${node}/vpp/config/v1/interface/${name}
    ${out}=      vpp_ctl: Delete key    ${uri}
    Log Many     ${out}
    [Return]    ${out}

vpp_ctl: Delete Linux Interface
    [Arguments]    ${node}    ${name}
    Log Many     ${node}    ${name}
    ${uri}=      Set Variable    /vnf-agent/${node}/linux/config/v1/interface/${name}
    ${out}=      vpp_ctl: Delete key    ${uri}
    Log Many     ${out}
    [Return]    ${out}

vpp_ctl: Delete Routes
    [Arguments]    ${node}    ${id}
    ${uri}=    Set Variable                /vnf-agent/${node}/vpp/config/v1/vrf/${id}/fib
    Log Many        ${uri}
    ${out}=         vpp_ctl: Delete key  ${uri}
    Log Many        ${out}
    [Return]       ${out}

vpp_ctl: Put BFD Session
    [Arguments]    ${node}    ${session_name}    ${min_tx_interval}    ${dest_adr}    ${detect_multiplier}    ${interface}    ${min_rx_interval}    ${source_adr}   ${enabled}    ${auth_key_id}=0    ${BFD_auth_key_id}=0
    Log Many    ${node}    ${session_name}    ${min_tx_interval}    ${dest_adr}    ${detect_multiplier}    ${interface}    ${min_rx_interval}    ${source_adr}   ${enabled}    ${auth_key_id}    ${BFD_auth_key_id}
    ${data}=              OperatingSystem.Get File      ${CURDIR}/../resources/bfd_session.json
    ${uri}=               Set Variable                  /vnf-agent/${node}/vpp/config/v1/bfd/session/${session_name}
    Log Many              ${data}                       ${uri}
    ${data}=              Replace Variables             ${data}
    Log                   ${data}
    vpp_ctl: Put Json     ${uri}    ${data}

vpp_ctl: Put BFD Authentication Key
    [Arguments]    ${node}    ${key_name}    ${auth_type}    ${id}    ${secret}
    Log Many    ${node}    ${key_name}    ${auth_type}    ${id}    ${secret}
    ${data}=              OperatingSystem.Get File      ${CURDIR}/../resources/bfd_key.json
    ${uri}=               Set Variable                  /vnf-agent/${node}/vpp/config/v1/bfd/auth-key/${key_name}
    Log Many              ${data}                       ${uri}
    ${data}=              Replace Variables             ${data}
    Log                   ${data}
    vpp_ctl: Put Json     ${uri}    ${data}

vpp_ctl: Put BFD Echo Function
    [Arguments]    ${node}    ${echo_func_name}    ${source_intf}
    Log Many    ${node}    ${echo_func_name}    ${source_intf}
    ${data}=              OperatingSystem.Get File      ${CURDIR}/../resources/bfd_echo_function.json
    ${uri}=               Set Variable                  /vnf-agent/${node}/vpp/config/v1/bfd/echo-function
    Log Many              ${data}                       ${uri}
    ${data}=              Replace Variables             ${data}
    Log                   ${data}
    vpp_ctl: Put Json     ${uri}    ${data}

vpp_ctl: Get BFD Session As Json
    [Arguments]    ${node}    ${session_name}
    Log Many    ${node}    ${session_name}
    ${key}=               Set Variable            /vnf-agent/${node}/vpp/config/v1/bfd/session/${session_name}
    Log                   ${key}
    ${data}=              vpp_ctl: Read Key    ${key}
    Log                   ${data}
    ${data}=              Set Variable If      '''${data}'''==""    {}    ${data}
    Log                   ${data}
    ${output}=            Evaluate             json.loads('''${data}''')    json
    Log                   ${output}
    [Return]              ${output}

vpp_ctl: Get BFD Authentication Key As Json
    [Arguments]    ${node}    ${key_name}
    Log Many    ${node}    ${key_name}
    ${key}=               Set Variable          /vnf-agent/${node}/vpp/config/v1/bfd/auth-key/${key_name}
    Log                   ${key}
    ${data}=              vpp_ctl: Read Key    ${key}
    Log                   ${data}
    ${data}=              Set Variable If      '''${data}'''==""    {}    ${data}
    Log                   ${data}
    ${output}=            Evaluate             json.loads('''${data}''')    json
    [Return]              ${output}

vpp_ctl: Get BFD Echo Function As Json
    [Arguments]    ${node}
    Log Many    ${node}
    ${key}=               Set Variable          /vnf-agent/${node}/vpp/config/v1/bfd/echo-function
    Log                   ${key}
    ${data}=              vpp_ctl: Read Key    ${key}
    Log                   ${data}
    ${data}=              Set Variable If      '''${data}'''==""    {}    ${data}
    Log                   ${data}
    ${output}=            Evaluate             json.loads('''${data}''')    json
    [Return]              ${output}

vpp_ctl: Put ACL TCP
    [Arguments]    ${node}    ${acl_name}    ${egr_intf1}   ${ingr_intf1}   ${rule_nm}    ${acl_action}    ${dest_ntw}    ${src_ntw}    ${dest_port_low}   ${dest_port_up}    ${src_port_low}    ${src_port_up}
    Log Many    ${node}    ${acl_name}    ${egr_intf1}   ${ingr_intf1}   ${rule_nm}    ${acl_action}    ${dest_ntw}    ${src_ntw}    ${dest_port_low}   ${dest_port_up}    ${src_port_low}    ${src_port_up}
    ${data}=              OperatingSystem.Get File      ${CURDIR}/../resources/acl_TCP.json
    ${uri}=               Set Variable          /vnf-agent/${node}/vpp/config/v1/acl/${acl_name}
    Log Many              ${data}                       ${uri}
    ${data}=              Replace Variables             ${data}
    Log                   ${data}
    #OperatingSystem.Create File   ${REPLY_DATA_FOLDER}/reply.json     ${data}
    vpp_ctl: Put Json     ${uri}    ${data}

vpp_ctl: Put ACL UDP
    [Arguments]    ${node}    ${acl_name}    ${egr_intf1}    ${ingr_intf1}     ${egr_intf2}    ${ingr_intf2}    ${rule_nm}   ${acl_action}    ${dest_ntw}   ${src_ntw}    ${dest_port_low}   ${dest_port_up}    ${src_port_low}    ${src_port_up}
    Log Many    ${node}    ${acl_name}    ${egr_intf1}    ${ingr_intf1}    ${egr_intf2}    ${ingr_intf2}   ${rule_nm}   ${acl_action}   ${dest_ntw}      ${src_ntw}    ${dest_port_low}   ${dest_port_up}    ${src_port_low}    ${src_port_up}
    ${data}=              OperatingSystem.Get File      ${CURDIR}/../resources/acl_UDP.json
    ${uri}=               Set Variable          /vnf-agent/${node}/vpp/config/v1/acl/${acl_name}
    Log Many              ${data}                       ${uri}
    ${data}=              Replace Variables             ${data}
    Log                   ${data}
    #OperatingSystem.Create File   ${REPLY_DATA_FOLDER}/reply.json     ${data}
    vpp_ctl: Put Json     ${uri}    ${data}

vpp_ctl: Put ACL MACIP
    [Arguments]    ${node}    ${acl_name}    ${egr_intf1}    ${ingr_intf1}    ${rule_nm}   ${acl_action}    ${src_addr}    ${src_addr_prefix}    ${src_mac_addr}   ${src_mac_addr_mask}
    Log Many    ${node}    ${acl_name}    ${egr_intf1}    ${ingr_intf1}    ${rule_nm}   ${acl_action}    ${src_addr}    ${src_addr_prefix}    ${src_mac_addr}   ${src_mac_addr_mask}
    ${data}=              OperatingSystem.Get File      ${CURDIR}/../resources/acl_MACIP.json
    ${uri}=               Set Variable          /vnf-agent/${node}/vpp/config/v1/acl/${acl_name}
    Log Many              ${data}                       ${uri}
    ${data}=              Replace Variables             ${data}
    Log                   ${data}
    #OperatingSystem.Create File   ${REPLY_DATA_FOLDER}/reply.json     ${data}
    vpp_ctl: Put Json     ${uri}    ${data}

vpp_ctl: Put ACL ICMP
    [Arguments]    ${node}    ${acl_name}    ${egr_intf1}   ${egr_intf2}    ${ingr_intf1}   ${ingr_intf2}   ${rule_nm}   ${acl_action}   ${dest_ntw}    ${src_ntw}    ${icmpv6}   ${code_range_low}   ${code_range_up}    ${type_range_low}   ${type_range_up}
    Log Many    ${node}    ${acl_name}    ${egr_intf1}   ${egr_intf2}    ${ingr_intf1}   ${ingr_intf2}   ${rule_nm}   ${acl_action}   ${dest_ntw}    ${src_ntw}    ${icmpv6}   ${code_range_low}   ${code_range_up}    ${type_range_low}   ${type_range_up}
    ${data}=              OperatingSystem.Get File      ${CURDIR}/../resources/acl_ICMP.json
    ${uri}=               Set Variable          /vnf-agent/${node}/vpp/config/v1/acl/${acl_name}
    Log Many              ${data}                       ${uri}
    ${data}=              Replace Variables             ${data}
    Log                   ${data}
    #OperatingSystem.Create File   ${REPLY_DATA_FOLDER}/reply.json     ${data}
    vpp_ctl: Put Json     ${uri}    ${data}

vpp_ctl: Get ACL As Json
    [Arguments]           ${node}  ${acl_name}
    Log Many              ${node}     ${acl_name}
    ${key}=               Set Variable          /vnf-agent/${node}/vpp/config/v1/acl/${acl_name}
    Log                   ${key}
    ${data}=              vpp_ctl: Read Key    ${key}
    Log                   ${data}
    ${data}=              Set Variable If      '''${data}'''==""    {}    ${data}
    Log                   ${data}
    #${output}=            Evaluate             json.loads('''${data}''')     json
    #log                   ${output}
    OperatingSystem.Create File   ${REPLY_DATA_FOLDER}/reply_${acl_name}.json    ${data}
    #[Return]              ${output}
    [Return]              ${data}

vpp_ctl: Get All ACL As Json
    [Arguments]           ${node}
    Log Many              ${node}
    ${key}=               Set Variable          /vnf-agent/${node}/vpp/config/v1/acl
    Log                   ${key}
    ${data}=              etcd: Get ETCD Tree    ${key}
    Log                   ${data}
    ${data}=              Set Variable If      '''${data}'''==""    {}    ${data}
    Log                   ${data}
    #${output}=            Evaluate             json.loads('''${data}''')     json
    #log                   ${output}
    OperatingSystem.Create File   ${REPLY_DATA_FOLDER}/reply_acl_all.json    ${data}
    #[Return]              ${output}
    [Return]              ${data}

etcd: Get ETCD Tree
    [Arguments]           ${key}
    Log Many              ${key}
    ${command}=         Set Variable    ${DOCKER_COMMAND} exec etcd etcdctl get --prefix="true" ${key}
    ${out}=             Execute On Machine    docker    ${command}    log=false
    [Return]            ${out}

vpp_ctl: Delete ACL
    [Arguments]    ${node}    ${name}
    Log Many     ${node}    ${name}
    ${uri}=      Set Variable    /vnf-agent/${node}/vpp/config/v1/acl/${name}
    ${out}=      vpp_ctl: Delete key    ${uri}
    Log Many     ${out}
    [Return]    ${out}

vpp_ctl: Put Veth Interface Via Linux Plugin
    [Arguments]    ${node}    ${namespace}    ${name}    ${host_if_name}    ${mac}    ${peer}    ${ip}    ${prefix}=24    ${mtu}=1500    ${enabled}=true
    Log Many    ${node}    ${namespace}    ${name}    ${host_if_name}    ${mac}    ${peer}    ${ip}    ${prefix}    ${mtu}    ${enabled}
    ${data}=              OperatingSystem.Get File      ${CURDIR}/../resources/linux_veth_interface.json
    ${uri}=               Set Variable                  /vnf-agent/${node}/linux/config/v1/interface/${name}
    Log Many              ${data}                       ${uri}
    ${data}=              Replace Variables             ${data}
    Log                   ${data}
    vpp_ctl: Put Json     ${uri}    ${data}

vpp_ctl: Put Linux Route
    [Arguments]    ${node}    ${namespace}    ${interface}    ${routename}    ${ip}    ${next_hop}    ${prefix}=24    ${metric}=100    ${isdefault}=false
    Log Many    ${node}    ${namespace}    ${interface}    ${routename}    ${ip}    ${prefix}    ${next_hop}    ${metric}    ${isdefault}
    ${data}=              OperatingSystem.Get File      ${CURDIR}/../resources/linux_static_route.json
    ${uri}=               Set Variable                  /vnf-agent/${node}/linux/config/v1/route/${routename}
    Log Many              ${data}                       ${uri}
    ${data}=              Replace Variables             ${data}
    Log                   ${data}
    vpp_ctl: Put Json     ${uri}    ${data}

vpp_ctl: Put Default Linux Route
    [Arguments]    ${node}    ${namespace}    ${interface}    ${routename}    ${next_hop}    ${metric}=100    ${isdefault}=true
    Log Many    ${node}    ${namespace}    ${interface}    ${routename}    ${next_hop}    ${metric}    ${isdefault}
    ${data}=              OperatingSystem.Get File      ${CURDIR}/../resources/linux_default_static_route.json
    ${uri}=               Set Variable                  /vnf-agent/${node}/linux/config/v1/route/${routename}
    Log Many              ${data}                       ${uri}
    ${data}=              Replace Variables             ${data}
    Log                   ${data}
    vpp_ctl: Put Json     ${uri}    ${data}

vpp_ctl: Put Linux Route Without Interface
    [Arguments]    ${node}    ${namespace}    ${routename}    ${ip}    ${next_hop}    ${prefix}=24    ${metric}=100
    Log Many    ${node}    ${namespace}    ${routename}    ${ip}    ${prefix}    ${next_hop}    ${metric}
    ${data}=              OperatingSystem.Get File      ${CURDIR}/../resources/linux_static_route_without_interface.json
    ${uri}=               Set Variable                  /vnf-agent/${node}/linux/config/v1/route/${routename}
    Log Many              ${data}                       ${uri}
    ${data}=              Replace Variables             ${data}
    Log                   ${data}
    vpp_ctl: Put Json     ${uri}    ${data}

vpp_ctl: Delete Linux Route
    [Arguments]    ${node}    ${routename}
    Log Many    ${node}    ${routename}
    ${uri}=               Set Variable                  /vnf-agent/${node}/linux/config/v1/route/${routename}
    ${out}=      vpp_ctl: Delete key    ${uri}
    Log Many     ${out}
    [Return]    ${out}

vpp_ctl: Get Linux Route As Json
    [Arguments]    ${node}    ${routename}
    Log Many    ${node}    ${routename}
    ${uri}=               Set Variable                  /vnf-agent/${node}/linux/config/v1/route/${routename}
    Log                   ${uri}
    ${data}=              vpp_ctl: Read Key    ${uri}
    Log                   ${data}
    ${data}=              Set Variable If      '''${data}'''==""    {}    ${data}
    Log                   ${data}
    ${output}=            Evaluate             json.loads('''${data}''')    json
    [Return]              ${output}

vpp_ctl: Check ACL Reply
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


 vpp_ctl: Put ARP
    [Arguments]    ${node}    ${interface}    ${ipv4}    ${MAC}    ${static}
    Log Many    ${node}    ${interface}    ${ipv4}    ${MAC}    ${static}
    ${data}=              OperatingSystem.Get File      ${CURDIR}/../resources/arp.json
    ${uri}=               Set Variable                  /vnf-agent/${node}/vpp/config/v1/arp/${interface}/${ipv4}
    Log Many              ${data}                       ${uri}
    ${data}=              Replace Variables             ${data}
    Log                   ${data}
    vpp_ctl: Put Json     ${uri}    ${data}

 vpp_ctl: Get ARP As Json
    [Arguments]           ${node}  ${interface}
    Log Many              ${node}     ${interface}
    ${key}=               Set Variable          /vnf-agent/${node}/vpp/config/v1/arp/${interface}
    Log                   ${key}
    ${data}=              vpp_ctl: Read Key    ${key}
    Log                   ${data}
    ${data}=              Set Variable If      '''${data}'''==""    {}    ${data}
    Log                   ${data}
    ${output}=            Evaluate             json.loads('''${data}''')     json
    log                   ${output}
    [Return]              ${output}

vpp_ctl: Set L4 Features On Node
    [Arguments]    ${node}    ${enabled}
    [Documentation]    Enable [disable] L4 features by setting ${enabled} to true [false].
    ${data}=              OperatingSystem.Get File      ${CURDIR}/../resources/enable-l4.json
    ${uri}=               Set Variable                  /vnf-agent/${node}/vpp/config/v1/l4/features/feature
    Log Many              ${data}                       ${uri}
    ${data}=              Replace Variables             ${data}
    Log                   ${data}
    vpp_ctl: Put Json     ${uri}    ${data}

vpp_ctl: Put Application Namespace
    [Arguments]    ${node}    ${id}    ${secret}    ${interface}
    [Documentation]    Put application namespace config json to etcd.
    Log Many    ${node}    ${id}    ${secret}    ${interface}
    ${data}=              OperatingSystem.Get File      ${CURDIR}/../resources/app_namespace.json
    ${uri}=               Set Variable                  /vnf-agent/${node}/vpp/config/v1/l4/namespaces/${id}
    Log Many              ${data}                       ${uri}
    ${data}=              Replace Variables             ${data}
    Log                   ${data}
    vpp_ctl: Put Json     ${uri}    ${data}


vpp_ctl: Delete ARP
    [Arguments]    ${node}    ${interface}    ${ipv4}
    Log Many    ${node}    ${interface}    ${ipv4}
    ${uri}=               Set Variable                  /vnf-agent/${node}/vpp/config/v1/arp/${interface}/${ipv4}
    ${out}=      vpp_ctl: Delete key    ${uri}
    Log Many     ${out}
    [Return]    ${out}

vpp_ctl: Put Linux ARP
    [Arguments]    ${node}    ${interface}    ${arp-name}    ${ipv4}    ${MAC}    ${static}
    Log Many    ${node}    ${interface}      ${arp-name}   ${ipv4}    ${MAC}    ${static}
    ${data}=              OperatingSystem.Get File      ${CURDIR}/../resources/arp.json
    ${uri}=               Set Variable                  /vnf-agent/${node}/vnf-agent/vpp1/linux/config/v1/arp/${arp-name}
    Log Many              ${data}                       ${uri}
    ${data}=              Replace Variables             ${data}
    Log                   ${data}
    vpp_ctl: Put Json     ${uri}    ${data}

