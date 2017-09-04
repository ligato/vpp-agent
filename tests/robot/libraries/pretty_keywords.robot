*** Keywords ***
Ping From ${node} To ${ip}
    vpp_term: Check Ping    ${node}    ${ip}

Create Loopback Interface ${name} On ${node} With Ip ${ip}/${prefix} And Mac ${mac}
    Log Many    ${name}    ${node}    ${ip}    ${prefix}    ${mac}
    vpp_ctl: Put Loopback Interface With IP    ${node}    ${name}   ${mac}   ${ip}   ${prefix}

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
