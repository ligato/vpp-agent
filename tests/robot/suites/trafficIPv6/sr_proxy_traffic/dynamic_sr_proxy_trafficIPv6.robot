*** Settings ***

Library     OperatingSystem
Library     String

Resource     ../../../variables/${VARIABLES}_variables.robot
Resource    ../../../libraries/all_libs.robot
Resource    ../../../libraries/pretty_keywords.robot

Suite Setup       Testsuite Setup
Suite Teardown    Testsuite Teardown
Test Setup        TestSetup
Test Teardown     TestTeardown

*** Variables ***
${VARIABLES}=                     common
${ENV}=                           common
${WAIT_TIMEOUT}=                  20s
${SYNC_SLEEP}=                    3s
${TRACE_WAIT_TIMEOUT}=            6s
${TRACE_SYNC_SLEEP}=              1s
${PING_WAIT_TIMEOUT}=             15s
${PING_SLEEP}=                    1s

@{segments}                       b::    c::
@{segmentsweight}                 1    @{segments}    # segment list's weight and segments
@{segmentList}                    ${segmentsweight}
${vpp1_tap_ipv6}=                 a::a
${linux_vpp1_tap_ipv6}=           a::1
${linux_vpp1_tap_ipv6_subnet}=    a::
${vpp3_tap_ipv6}=                 c::c
${linux_vpp3_tap_ipv6}=           c::1
${linux_vpp3_tap_ipv6_subnet}=    c::
${linux_vpp1_tap_ipv4_subnet}     10.1.1.0  # 24-bit netmask, IPv4 pattern: 10.<vpp number>.x.x
${linux_vpp1_tap_ipv4}=           10.1.1.1
${vpp1_tap_ipv4}=                 10.1.1.2
${vpp1_memif2_ipv4}=              10.1.3.1
${srproxy_out_memif_ipv4}=        10.2.1.1
${srproxy_in_memif_ipv4}=         10.2.2.2
${linux_vpp3_tap_ipv4_subnet}=    10.3.1.0  # 24-bit netmask
${linux_vpp3_tap_ipv4}=           10.3.1.1
${vpp3_tap_ipv4}=                 10.3.1.2
${vpp3_memif2_ipv4}=              10.3.3.1
${service_out_memif_ipv4}=        10.4.1.2
${service_in_memif_ipv4}=         10.4.2.1
${vpp1_tap_mac}=                  11:11:11:11:11:11
${linux_vpp1_tap_mac}=            22:22:22:22:22:22
${vpp3_tap_mac}=                  33:33:33:33:33:33
${linux_vpp3_tap_mac}=            44:44:44:44:44:44
${vpp1_memif1_mac}=               02:f1:be:90:00:01
${vpp2_memif1_mac}=               02:f1:be:90:00:02
${vpp2_memif2_mac}=               02:f1:be:90:02:02
${vpp3_memif1_mac}=               02:f1:be:90:00:03
${vpp3_memif2_mac}=               02:f1:be:90:02:03
${vpp1_memif2_mac}=               02:f1:be:90:02:01
${srproxy_out_memif_mac}=         02:f1:be:90:03:02
${service_in_memif_mac}=          02:f1:be:90:00:04
${service_out_memif_mac}=         02:f1:be:90:02:04
${srproxy_in_memif_mac}=          02:f1:be:90:04:02

# ethernet frame sending variables (used as values in sending python script or in validation)
${out_interface}=                 linux_vpp1_tap
${source_mac_address}=            01:02:03:04:05:06
${destination_mac_address}=       01:02:03:04:05:06
${ethernet_type}                  88b5                                # using public ethernet type for prototype and vendor-specific protocol development to not explicitly say what to expect in payload (=general frame) (http://standards-oui.ieee.org/ethertype/eth.txt)
${payload}                        "["*30)+"PAYLOAD"+("]"*30           # custom payload (= general frame) (inserted directly "as is" in python script)
${payload_hex_prefix}=            5b5b5b5b5b5b5b5b5b5b5b5b5b5b5b5b    # partial prefix of payload in hexadecimal form (used for validation)
${checksum}                       1a2b3c4d                            # checksum is not controlled anywhere, but needed for correct construction of frame structure
${srproxy_out_memif_vpp_name}=    memif3/3                            # used for validation
${srproxy_in_memif_vpp_index}=    4                                   # used for validation

*** Test Cases ***
Common Setup Used Across All Tests
    [Tags]    setup
    # create nodes
    Add Agent VPP Node    agent_vpp_1
    Add Agent VPP Node    agent_vpp_2
    Add Agent VPP Node    agent_vpp_3
    Add Agent VPP Node    agent_vpp_4

    # creating TAP tunnels between linux and VPP (in containers)
    Put TAPv2 Interface With 2 IPs    node=agent_vpp_1    name=vpp1_tap                  ip=${vpp1_tap_ipv6}          prefix=64    second_ip=${vpp1_tap_ipv4}    second_prefix=24    host_if_name=linux_vpp1_tap    mac=${vpp1_tap_mac}
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    vpp_term: Interface Is Created    node=agent_vpp_1    mac=${vpp1_tap_mac}
    linux: Set Host TAP Interface     node=agent_vpp_1    host_if_name=linux_vpp1_tap    ip=${linux_vpp1_tap_ipv6}    prefix=64    mac=${linux_vpp1_tap_mac}     second_ip=${linux_vpp1_tap_ipv4}    second_prefix=24
    Put TAPv2 Interface With 2 IPs    node=agent_vpp_3    name=vpp3_tap                  ip=${vpp3_tap_ipv6}          prefix=64    second_ip=${vpp3_tap_ipv4}    second_prefix=24    host_if_name=linux_vpp3_tap    mac=${vpp3_tap_mac}
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    vpp_term: Interface Is Created    node=agent_vpp_3    mac=${vpp3_tap_mac}
    linux: Set Host TAP Interface     node=agent_vpp_3    host_if_name=linux_vpp3_tap    ip=${linux_vpp3_tap_ipv6}    prefix=64    mac=${linux_vpp3_tap_mac}     second_ip=${linux_vpp3_tap_ipv4}    second_prefix=24
    # creating VPP (memif) tunnels between nodes (for IPv6 address purposes: agent_vpp_1 = node A, agent_vpp_2 = node B (SR proxy), agent_vpp_3 = node C, agent_vpp_4 = node D (SR-unaware service))
    Create Master vpp1_memif1 on agent_vpp_1 with IP ab::a, MAC ${vpp1_memif1_mac}, key 1 and m1.sock socket
    Create Slave vpp2_memif1 on agent_vpp_2 with IP ab::b, MAC ${vpp2_memif1_mac}, key 1 and m1.sock socket
    Create Master vpp2_memif2 on agent_vpp_2 with IP bc::b, MAC ${vpp2_memif2_mac}, key 2 and m2.sock socket
    Create Slave vpp3_memif1 on agent_vpp_3 with IP bc::c, MAC ${vpp3_memif1_mac}, key 2 and m2.sock socket
    # creating routes between nodes (using memifs)
    Create Route On agent_vpp_1 With IP b::/64 With Vrf Id 0 With Interface vpp1_memif1 And Next Hop ab::b
    Create Route On agent_vpp_2 With IP c::/64 With Vrf Id 0 With Interface vpp2_memif2 And Next Hop bc::c
    # configure segment routing that is common for all tests
    Put SRv6 Policy                 node=agent_vpp_1    bsid=a::c            fibtable=0         srhEncapsulation=true    sprayBehaviour=false    segmentlists=${segmentList}
    # preventing packet drops due to unresolved ipv6 neighbor discovery
    vpp_term: Set IPv6 neighbor  agent_vpp_1    memif1/1    ab::b    ${vpp2_memif1_mac}
    vpp_term: Set IPv6 neighbor  agent_vpp_2    memif1/1    ab::a    ${vpp1_memif1_mac}
    vpp_term: Set IPv6 neighbor  agent_vpp_2    memif2/2    bc::c    ${vpp3_memif1_mac}
    vpp_term: Set IPv6 neighbor  agent_vpp_3    memif1/2    bc::b    ${vpp2_memif2_mac}

Dynamic SR Proxy with L3-IPv6 SR-unaware service
    # Testing IPv6 traffic(ping packet) going through SR-proxy (IPv6 segment routing) connected to IPv6 SR-unaware service.
    # Desired ping packet path:
    ## Container agent_vpp_1:                          linux environment -> linux_vpp1_tap interface -(tap tunnel to VPP1)-> vpp1_tap interface
    ##                                                 -> steering to segment routing (segment list="b::, c::") -(segment routing to b::)
    ##                                                 -> vpp1_memif1 interface (memif tunnel to VPP2)
    ## Container agent_vpp_2 (SR proxy node):          vpp2_memif1 interface -(SR-proxy functionality)-> srproxy_out_memif interface
    ## Container agent_vpp_4 (SR-unware service node): service_in_memif interface -(just forwarding incomming packet)-> service_out_memif interface
    ## Container agent_vpp_2 (SR proxy node):          srproxy_in_memif interface -(segment routing to c::)-> vpp2_memif2 interface
    ## Container agent_vpp_3:                          vpp3_memif1 interface -(DX6 decapsulation from segment routing)-> vpp3_tap interface
    ##                                                 -> linux_vpp3_tap(ping reached destination) -(ping reply)-> vpp3_tap interface
    ##                                                 -> vpp3_memif2 interface (memif tunnel to VPP1)
    ## Container agent_vpp_1:                          vpp1_memif2 interface -> vpp1_tap interface -> linux_vpp1_tap interface -> linux environment

    # creating path for ping packet (path is already partially done in common setup)
    ## steering trafic from linux to TAPs tunnel leading to VPP
    linux: Add Route     node=agent_vpp_1    destination_ip=${linux_vpp3_tap_ipv6}    prefix=64    next_hop_ip=${vpp1_tap_ipv6}
    ## steering traffic to segment routing
    Put SRv6 L3 Steering    node=agent_vpp_1    name=toC    bsid=a::c    fibtable=0    prefixAddress=c::/64
    ## creating sr-proxy in and out interfaces (using memifs)
    Create Master srproxy_out_memif on agent_vpp_2 with Prefixed IP bd:1::b/32, MAC ${srproxy_out_memif_mac}, key 3 and m3.sock socket
    Create Slave service_in_memif on agent_vpp_4 with Prefixed IP bd:1::d/32, MAC ${service_in_memif_mac}, key 3 and m3.sock socket
    Create Master service_out_memif on agent_vpp_4 with Prefixed IP bd:2::d/32, MAC ${service_out_memif_mac}, key 4 and m4.sock socket
    Create Slave srproxy_in_memif on agent_vpp_2 with Prefixed IP bd:2::b/32, MAC ${srproxy_in_memif_mac}, key 4 and m4.sock socket
    ## configure SR-proxy
    Put Local SID With End.AD function    node=agent_vpp_2    sidAddress=b::    l3serviceaddress=bd:1::d    outinterface=srproxy_out_memif    ininterface=srproxy_in_memif
    ## creating service routes (Service just sends received packets back)
    Create Route On agent_vpp_4 With IP ${linux_vpp3_tap_ipv6_subnet}/64 With Vrf Id 0 With Interface service_out_memif And Next Hop bd:2::b
    ## configure exit from segment routing
    Put Local SID With End.DX6 function    node=agent_vpp_3    sidAddress=c::     fibtable=0         outinterface=vpp3_tap    nexthop=${linux_vpp3_tap_ipv6}
    ## path for ping packet returning back to source (no segment routing, but just plain IPv6 route):
    ## create route for ping echo to get back to vpp3 from linux enviroment in agent_vpp_3 container
    linux: Add Route    node=agent_vpp_3    destination_ip=${linux_vpp1_tap_ipv6}    prefix=64    next_hop_ip=${vpp3_tap_ipv6}
    ## creating path for ping echo from vpp3 to vpp1
    Create Master vpp3_memif2 on agent_vpp_3 with IP ac::c, MAC ${vpp3_memif2_mac}, key 5 and m5.sock socket
    Create Slave vpp1_memif2 on agent_vpp_1 with IP ac::a, MAC ${vpp1_memif2_mac}, key 5 and m5.sock socket
    Create Route On agent_vpp_3 With IP ${linux_vpp1_tap_ipv6_subnet}/64 With Vrf Id 0 With Interface vpp3_memif2 And Next Hop ac::a

    # preventing packet drops due to unresolved ipv6 neighbor discovery
    vpp_term: Set IPv6 neighbor  agent_vpp_1    memif2/5    ac::c                     ${vpp3_memif2_mac}
    vpp_term: Set IPv6 neighbor  agent_vpp_1    tap0        ${linux_vpp1_tap_ipv6}    ${linux_vpp1_tap_mac}
    vpp_term: Set IPv6 neighbor  agent_vpp_2    memif3/3    bd:1::d                   ${service_in_memif_mac}
    vpp_term: Set IPv6 neighbor  agent_vpp_2    memif4/4    bd:2::d                   ${service_out_memif_mac}
    vpp_term: Set IPv6 neighbor  agent_vpp_3    memif2/5    ac::a                     ${vpp1_memif2_mac}
    vpp_term: Set IPv6 neighbor  agent_vpp_3    tap0        ${linux_vpp3_tap_ipv6}    ${linux_vpp3_tap_mac}
    vpp_term: Set IPv6 neighbor  agent_vpp_4    memif1/3    bd:1::b                   ${srproxy_out_memif_mac}
    vpp_term: Set IPv6 neighbor  agent_vpp_4    memif2/4    bd:2::b                   ${srproxy_in_memif_mac}
    # add packet tracing
    vpp_term: Add Trace Memif       agent_vpp_2    100
    # ping from agent_vpp_1 to agent_vpp_3's tap interface  (despite efforts to eliminite first packet drop by setting ipv6 neighbor, sometimes it is still happening -> timeoutable pinging repeat until first ping success)
    Wait Until Keyword Succeeds    ${PING_WAIT_TIMEOUT}   .${PING_SLEEP}    linux: Check Ping6    node=agent_vpp_1    ip=${linux_vpp3_tap_ipv6}    count=1
    # check that packet is processed by SR-proxy, send to SR-unware service using correct interface and in the process decapsulated(checked only source and destination address and not if SR header extension is missing because that is not visible in trace)
    ${vpp2trace}=    vpp_term: Show Trace    agent_vpp_2
    Packet Trace ${vpp2trace} Should Contain One Packet Trace That Contains These Ordered Substrings srv6-ad-localsid, IP6: ${srproxy_out_memif_mac} -> ${service_in_memif_mac}, ICMP6: ${linux_vpp1_tap_ipv6} -> ${linux_vpp3_tap_ipv6}, ., .  # using only 3 substrings to match packet trace
    # check that packet has returned from SR-unware service, gets rewritten by proxy (SR encapsulation) and send to another SR segment (correct interface and correct source and destination address)
    Packet Trace ${vpp2trace} Should Contain One Packet Trace That Contains These Ordered Substrings ${linux_vpp1_tap_ipv6} -> ${linux_vpp3_tap_ipv6}, SRv6-AD-rewrite: src :: dst c::, IP6: ${vpp2_memif2_mac} -> ${vpp3_memif1_mac}, IPV6_ROUTE: :: -> c::, .  # using only 4 substrings to match packet trace

    # cleanup (for next test)
    Delete VPP Interface     agent_vpp_2         srproxy_in_memif
    Delete VPP Interface     agent_vpp_2         srproxy_out_memif
    Delete VPP Interface     agent_vpp_4         service_in_memif
    Delete VPP Interface     agent_vpp_4         service_out_memif
    Delete VPP Interface     agent_vpp_3         vpp3_memif2
    Delete VPP Interface     agent_vpp_1         vpp1_memif2
    Delete Local SID         node=agent_vpp_2    sidAddress=b::
    Delete Local SID         node=agent_vpp_3    sidAddress=c::
    vpp_term: Clear Trace    agent_vpp_2

Dynamic SR Proxy with L3-IPv4 SR-unaware service
    # Testing IPv4 traffic(ping packet) going through SR-proxy (IPv6 segment routing) connected to IPv4 SR-unaware service.
    # Desired ping packet path is basically the same as in test for IPv6 SR-unaware service. The difference is that we
    # use IPv4 addresses and IPv4 routes everywhere except of IPv6 segment routing.

    # creating path for ping packet (path is already partially done in common setup)
    ## steering trafic from linux to TAPs tunnel leading to VPP
    linux: Add Route     node=agent_vpp_1    destination_ip=${linux_vpp3_tap_ipv4_subnet}    prefix=24    next_hop_ip=${vpp1_tap_ipv4}
    ## steering traffic to segment routing
    Put SRv6 L3 Steering    node=agent_vpp_1    name=toC    bsid=a::c    fibtable=0    prefixAddress=${linux_vpp3_tap_ipv4_subnet}/24
    ## creating sr-proxy in and out interfaces (using memifs)
    Create Master srproxy_out_memif on agent_vpp_2 with Prefixed IP ${srproxy_out_memif_ipv4}/24, MAC ${srproxy_out_memif_mac}, key 3 and m3.sock socket
    Create Slave service_in_memif on agent_vpp_4 with Prefixed IP ${service_in_memif_ipv4}/24, MAC ${service_in_memif_mac}, key 3 and m3.sock socket
    Create Master service_out_memif on agent_vpp_4 with Prefixed IP ${service_out_memif_ipv4}/24, MAC ${service_out_memif_mac}, key 4 and m4.sock socket
    Create Slave srproxy_in_memif on agent_vpp_2 with Prefixed IP ${srproxy_in_memif_ipv4}/24, MAC ${srproxy_in_memif_mac}, key 4 and m4.sock socket
    ## configure SR-proxy
    Put Local SID With End.AD function     node=agent_vpp_2    sidAddress=b::    l3serviceaddress=${service_in_memif_ipv4}    outinterface=srproxy_out_memif    ininterface=srproxy_in_memif
    ## creating service routes (Service just sends received packets back)
    Create Route On agent_vpp_4 With IP ${linux_vpp3_tap_ipv4_subnet}/24 With Vrf Id 0 With Interface service_out_memif And Next Hop ${srproxy_in_memif_ipv4}
    ## configure exit from segment routing
    Put Local SID With End.DX4 function    node=agent_vpp_3    sidAddress=c::    fibtable=0    outinterface=vpp3_tap    nexthop=${linux_vpp3_tap_ipv4}
    ## path for ping packet returning back to source (no segment routing, but just plain IPv4 route):
    ## create route for ping echo to get back to vpp3 from linux enviroment in agent_vpp_3 container
    linux: Add Route    node=agent_vpp_3    destination_ip=${linux_vpp1_tap_ipv4_subnet}    prefix=24    next_hop_ip=${vpp3_tap_ipv4}
    ## creating path for ping echo from vpp3 to vpp1
    Create Master vpp3_memif2 on agent_vpp_3 with IP ${vpp3_memif2_ipv4}, MAC ${vpp3_memif2_mac}, key 5 and m5.sock socket
    Create Slave vpp1_memif2 on agent_vpp_1 with IP ${vpp1_memif2_ipv4}, MAC ${vpp1_memif2_mac}, key 5 and m5.sock socket
    Create Route On agent_vpp_3 With IP ${linux_vpp1_tap_ipv4_subnet}/24 With Vrf Id 0 With Interface vpp3_memif2 And Next Hop ${vpp1_memif2_ipv4}

    # preventing packet drops due to unresolved ipv6 neighbor discovery (some won't resolve properly, so it is not only about good traffic from first ping)
    vpp_term: Set ARP    agent_vpp_2    memif3/3    ${service_in_memif_ipv4}    ${service_in_memif_mac}
    vpp_term: Set ARP    agent_vpp_3    memif2/5    ${vpp1_memif2_ipv4}         ${vpp1_memif2_mac}
    vpp_term: Set ARP    agent_vpp_4    memif2/4    ${srproxy_in_memif_ipv4}    ${srproxy_in_memif_mac}

    # add packet tracing
    vpp_term: Add Trace Memif       agent_vpp_2    100
    # ping from agent_vpp_1 to agent_vpp_3's tap interface  (despite efforts to eliminite first packet drop by setting arp, sometimes it is still happening -> timeoutable pinging repeat until first ping success)
    Wait Until Keyword Succeeds    ${PING_WAIT_TIMEOUT}   .${PING_SLEEP}    linux: Check Ping    node=agent_vpp_1    ip=${linux_vpp3_tap_ipv4}    count=1
    # check that packet is processed by SR-proxy, send to SR-unware service using correct interface and in the process decapsulated(checked only source and destination address and not if SR header extension is missing because that is not visible in trace)
    ${vpp2trace}=    vpp_term: Show Trace    agent_vpp_2
    Packet Trace ${vpp2trace} Should Contain One Packet Trace That Contains These Ordered Substrings srv6-ad-localsid, IP4: ${srproxy_out_memif_mac} -> ${service_in_memif_mac}, ICMP: ${linux_vpp1_tap_ipv4} -> ${linux_vpp3_tap_ipv4}, ., .  # using only 3 substrings to match packet trace
    # check that packet has returned from SR-unware service, gets rewritten by proxy (SR encapsulation) and send to another SR segment (correct interface and correct source and destination address)
    Packet Trace ${vpp2trace} Should Contain One Packet Trace That Contains These Ordered Substrings ${linux_vpp1_tap_ipv4} -> ${linux_vpp3_tap_ipv4}, SRv6-AD-rewrite: src :: dst c::, IP6: ${vpp2_memif2_mac} -> ${vpp3_memif1_mac}, IPV6_ROUTE: :: -> c::, .  # using only 4 substrings to match packet trace

    # cleanup (for next test)
    Delete Local SID         node=agent_vpp_2    sidAddress=b::
    vpp_term: Clear Trace    agent_vpp_2

Dynamic SR Proxy with L2 SR-unaware service
    # Testing L2 traffic(sending custom Ethernet frame) going through SR-proxy (IPv6 segment routing) connected to
    # L2 SR-unaware service. Desired frame path starts identically to the path in IPv6 SR-unaware service test, but
    # ends right after SR-proxy (at least it is not further checked). We don't do the full packet/frame path as for
    # IPv4/IPv6 SR-unaware services, because sending ethernet frame is not like ping that has echo and nice ping tool
    # in linux that tells you whether echo was received or not. For doing the same for ethernet frame, there have to be
    # some traffic catching tool and that is too complicated for something that is basically not the aim of test.
    # The aim is to check SR-proxy functionality.

    # creating path for frame (path is already partially done in common setup+using memifs between sr-proxy and service from ipv4 test)
    ## steering traffic to segment routing
    Put SRv6 L2 Steering    node=agent_vpp_1    name=toC    bsid=a::c    interfaceName=vpp1_tap
    ## configure SR-proxy
    Put Local SID With End.AD function    node=agent_vpp_2    sidAddress=b::    outinterface=srproxy_out_memif    ininterface=srproxy_in_memif    # L2 SR-unware service
    ## creating service paths (Service just sends received frame back)
    Create Bridge Domain bd1 Without Autolearn On agent_vpp_4 With Interfaces service_in_memif, service_out_memif
    Add fib entry for 01:02:03:04:05:06 in bd1 over service_out_memif on agent_vpp_4

    # add packet tracing
    vpp_term: Add Trace Memif     agent_vpp_2    100
    # sending ethernet frame
    linux: Send Ethernet Frame    agent_vpp_1    ${out_interface}    ${source_mac_address}    ${destination_mac_address}    ${ethernet_type}    ${payload}    ${checksum}
    Wait Until Keyword Succeeds   ${TRACE_WAIT_TIMEOUT}   ${TRACE_SYNC_SLEEP}    Trace on agent_vpp_2 has at least 2 packets
    # check that packet is processed by SR-proxy, send to SR-unware service using correct interface and in the process decapsulated(checked only ethernet frame
    # type and source/destination address and not if SR header extension is missing because that is not visible in trace)
    ${vpp2trace}=    vpp_term: Show Trace    agent_vpp_2
    Packet Trace ${vpp2trace} Should Contain One Packet Trace That Contains These Ordered Substrings srv6-ad-localsid, ${srproxy_out_memif_vpp_name}, 0x${ethernet_type}: ${source_mac_address} -> ${destination_mac_address}, ., .  # using only 3 substrings to match packet trace
    # check that packet has returned from SR-unware service (something received from correct interface and later checked its rewritten payload to be sure it is
    # the right frame), gets rewritten by proxy (SR encapsulation) and send to another SR segment (correct interface and correct source and destination address)
    Packet Trace ${vpp2trace} Should Contain One Packet Trace That Contains These Ordered Substrings memif: hw_if_index ${srproxy_in_memif_vpp_index}, SRv6-AD-rewrite: src :: dst c::, ${ethernet_type}${payload_hex_prefix}, IP6: ${vpp2_memif2_mac} -> ${vpp3_memif1_mac}, IPV6_ROUTE: :: -> c::


*** Keywords ***
Packet Trace ${packettrace} Should Contain One Packet Trace That Contains These Ordered Substrings ${substr1}, ${substr2}, ${substr3}, ${substr4}, ${substr5}
    ${packetsplit}=       String.Split String    ${packettrace}    Packet
    Should Contain Match    ${packetsplit}    regexp=.*${substr1}.*${substr2}.*${substr3}.*${substr4}.*${substr5}.*    case_insensitive=True

Trace on ${node} has at least 2 packets
    ${trace}=    vpp_term: Show Trace    ${node}
    Should Contain    ${trace}    Packet 2

TestSetup
    Make Datastore Snapshots    ${TEST_NAME}_test_setup

TestTeardown
    Make Datastore Snapshots    ${TEST_NAME}_test_teardown