*** Keywords ***
Ping From ${node} To ${ip}
    vpp_term: Check Ping    ${node}    ${ip}

Ping6 From ${node} To ${ip}
    vpp_term: Check Ping6    ${node}    ${ip}


Ping On ${node} With IP ${ip}, Source ${source}
    vpp_term: Check Ping Within Interface    ${node}    ${ip}    ${source}

Create Loopback Interface ${name} On ${node} With Ip ${ip}/${prefix} And Mac ${mac}
    Log Many    ${name}    ${node}    ${ip}    ${prefix}    ${mac}
    vpp_ctl: Put Loopback Interface With IP    ${node}    ${name}   ${mac}   ${ip}   ${prefix}

Create Loopback Interface ${name} On ${node} With Mac ${mac}
    Log Many    ${name}    ${node}    ${mac}
    vpp_ctl: Put Loopback Interface    ${node}    ${name}   ${mac}

Create Loopback Interface ${name} On ${node} With VRF ${vrf}, Ip ${ip}/${prefix} And Mac ${mac}
    Log Many    ${name}    ${node}    ${ip}    ${prefix}    ${mac}    ${vrf}
    vpp_ctl: Put Loopback Interface With IP    ${node}    ${name}   ${mac}   ${ip}   ${prefix}    vrf=${vrf}

Create ${type} ${name} On ${node} With MAC ${mac}, Key ${key} And ${sock} Socket
    Log Many    ${type}    ${name}    ${node}    ${mac}    ${key}    ${sock}
    ${type}=    Set Variable if    "${type}"=="Master"    true    false
    Log Many    ${type}
    vpp_ctl: put memif interface   ${node}   ${name}   ${mac}   ${type}   ${key}   ${sock}

Create ${type} ${name} On ${node} With IP ${ip}, MAC ${mac}, Key ${key} And ${socket} Socket
    Log Many   ${type}    ${name}    ${node}    ${ip}    ${key}    ${socket}
    ${type}=   Set Variable if    "${type}"=="Master"    true    false
    Log Many   ${type}
    ${out}=    vpp_ctl: Put Memif Interface With IP    ${node}    ${name}   ${mac}    ${type}   ${key}    ${ip}    socket=${socket}
    Log Many   ${out}

Create ${type} ${name} On ${node} With Vrf ${vrf}, IP ${ip}, MAC ${mac}, Key ${key} And ${socket} Socket
    Log Many   ${type}    ${name}    ${node}    ${ip}    ${key}    ${vrf}    ${socket}
    ${type}=   Set Variable if    "${type}"=="Master"    true    false
    Log Many   ${type}
    ${out}=    vpp_ctl: Put Memif Interface With IP    ${node}    ${name}   ${mac}    ${type}   ${key}    ${ip}    socket=${socket}    vrf=${vrf}
    Log Many   ${out}

Create Tap Interface ${name} On ${node} With Vrf ${vrf}, IP ${ip}, MAC ${mac} And HostIfName ${host_if_name}
    Log Many   ${name}    ${node}    ${ip}    ${vrf}
    ${out}=    vpp_ctl: Put TAP Interface With IP    ${node}    ${name}   ${mac}    ${ip}    ${host_if_name}    vrf=${vrf}
    Log Many   ${out}

Create Bridge Domain ${name} with Autolearn On ${node} With Interfaces ${interfaces}
    Log Many    ${name}    ${node}    ${interfaces}
    @{ints}=    Split String   ${interfaces}    separator=,${space}
    Log Many    @{ints}
    vpp_ctl: put bridge domain    ${node}    ${name}   ${ints}

Create Bridge Domain ${name} Without Autolearn On ${node} With Interfaces ${interfaces}
    Log Many    ${name}    ${node}    ${interfaces}
    @{ints}=    Split String   ${interfaces}    separator=,${space}
    Log Many    @{ints}
    vpp_ctl: put bridge domain    ${node}    ${name}   ${ints}    unicast=false    learn=false

Create Route On ${node} With IP ${ip}/${prefix} With Next Hop ${next_hop} And Vrf Id ${id}
    Log Many        ${node}    ${ip}   ${prefix}    ${next_hop}    ${id}
    ${data}=        OperatingSystem.Get File    ${CURDIR}/../../robot/resources/static_route.json
    Log Many        ${data}
    ${data}=        replace variables           ${data}
    Log Many        ${data}
    ${uri}=         Set Variable                /vnf-agent/${node}/vpp/config/v1/vrf/${id}/fib/${ip}/${prefix}/${next_hop}
    Log Many        ${uri}
    ${out}=         vpp_ctl: Put Json    ${uri}   ${data}
    Log Many        ${out}

Delete IPsec On ${node} With Prefix ${prefix} And Name ${name}
    Log Many        ${node}    ${prefix}    ${name}
    vpp_ctl: Delete IPsec    ${node}    ${prefix}    ${name}

Create VXLan ${name} From ${src_ip} To ${dst_ip} With Vni ${vni} On ${node}
    Log Many    ${name}    ${src_ip}    ${dst_ip}    ${vni}    ${node}
    vpp_ctl: Put VXLan Interface    ${node}    ${name}    ${src_ip}    ${dst_ip}    ${vni}

Delete Routes On ${node} And Vrf Id ${id}
    Log Many        ${node}    ${id}
    vpp_ctl: Delete routes    ${node}    ${id}

Remove Interface ${name} On ${node}
    Log Many    ${name}    ${node}
    vpp_ctl: Delete VPP Interface    ${node}    ${name}

Remove Bridge Domain ${name} On ${node}
    Log Many    ${name}    ${node}
    vpp_ctl: Delete Bridge Domain    ${node}    ${name}

Add fib entry for ${mac} in ${name} over ${outgoing} on ${node}
    Log Many    ${mac}    ${name}    ${outgoing}    ${node}
    vpp_ctl: Put Static Fib Entry    ${node}    ${name}    ${mac}    ${outgoing}

Command: ${cmd} should ${expected}
    Log Many    ${cmd}   ${expected}
    ${out}=         Run Keyword And Ignore Error    ${cmd}
    Log Many    @{out}
    Should Be Equal    @{out}[0]    ${expected}    ignore_case=True
    [Return]     ${out}

IP Fib Table ${id} On ${node} Should Be Empty
    Log many    ${node} ${id}
    ${out}=    vpp_term: Show IP Fib Table    ${node}   ${id}
    log many    ${out}
    Should Be Equal    ${out}   vpp#${SPACE}

IP6 Fib Table ${id} On ${node} Should Be Empty
    Log many    ${node} ${id}
    ${out}=    vpp_term: Show IP6 Fib Table    ${node}   ${id}
    log many    ${out}
    Should Be Equal    ${out}   vpp#${SPACE}

IP Fib Table ${id} On ${node} Should Not Be Empty
    Log many    ${node} ${id}
    ${out}=    vpp_term: Show IP Fib Table    ${node}   ${id}
    log many    ${out}
    Should Not Be Equal    ${out}   vpp#${SPACE}

IP Fib Table ${id} On ${node} Should Contain Route With IP ${ip}/${prefix}
    Log many    ${node} ${id}
    ${out}=    vpp_term: Show IP Fib Table    ${node}   ${id}
    log many    ${out}
    Should Match Regexp        ${out}  ${ip}\\/${prefix}\\s*unicast\\-ip4-chain\\s*\\[\\@0\\]:\\ dpo-load-balance:\\ \\[proto:ip4\\ index:\\d+\\ buckets:\\d+\\ uRPF:\\d+\\ to:\\[0:0\\]\\]

IP Fib Table ${id} On ${node} Should Not Contain Route With IP ${ip}/${prefix}
    Log many    ${node} ${id}
    ${out}=    vpp_term: Show IP Fib Table    ${node}   ${id}
    log many    ${out}
    Should Not Match Regexp        ${out}  ${ip}\\/${prefix}\\s*unicast\\-ip4-chain\\s*\\[\\@0\\]:\\ dpo-load-balance:\\ \\[proto:ip4\\ index:\\d+\\ buckets:\\d+\\ uRPF:\\d+\\ to:\\[0:0\\]\\]

IP6 Fib Table ${id} On ${node} Should Contain Route With IP ${ip}/${prefix}
    Log many    ${node} ${id}
    ${out}=    vpp_term: Show IP6 Fib Table    ${node}   ${id}
    log many    ${out}
    Should Match Regexp        ${out}  ${ip}\\/${prefix}\\s*unicast\\-ip6-chain\\s*\\[\\@0\\]:\\ dpo-load-balance:\\ \\[proto:ip6\\ index:\\d+\\ buckets:\\d+\\ uRPF:\\d+\\ to:\\[0:0\\]\\]

IP6 Fib Table ${id} On ${node} Should Not Contain Route With IP ${ip}/${prefix}
    Log many    ${node} ${id}
    ${out}=    vpp_term: Show IP6 Fib Table    ${node}   ${id}
    log many    ${out}
    Should Not Match Regexp        ${out}  ${ip}\\/${prefix}\\s*unicast\\-ip6-chain\\s*\\[\\@0\\]:\\ dpo-load-balance:\\ \\[proto:ip6\\ index:\\d+\\ buckets:\\d+\\ uRPF:\\d+\\ to:\\[0:0\\]\\]


Show IP Fib On ${node}
    Log Many    ${node}
    ${out}=     vpp_term: Show IP Fib    ${node}
    Log Many    ${out}

Show IP6 Fib On ${node}
    Log Many    ${node}
    ${out}=     vpp_term: Show IP6 Fib    ${node}
    Log Many    ${out}

Show Interfaces On ${node}
    ${out}=   vpp_term: Show Interfaces    ${node}
    Log Many  ${out}

Show Interfaces Address On ${node}
    ${out}=   vpp_term: Show Interfaces Address    ${node}
    Log Many  ${out}

Create Linux Route On ${node} With IP ${ip}/${prefix} With Next Hop ${next_hop} And Vrf Id ${id}
    Log Many        ${node}    ${ip}   ${prefix}    ${next_hop}
    ${data}=        OperatingSystem.Get File    ${CURDIR}/../../robot/resources/linux_static_route.json
    Log Many        ${data}
    ${data}=        replace variables           ${data}
    Log Many        ${data}
    ${uri}=         Set Variable                /vnf-agent/${node}/vpp/config/v1/vrf/${id}/fib/${ip}/${prefix}/${next_hop}
    Log Many        ${uri}
    ${out}=         vpp_ctl: Put Json    ${uri}   ${data}
    Log Many        ${out}
