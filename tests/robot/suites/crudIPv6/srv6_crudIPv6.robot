*** Settings ***
Library      OperatingSystem
#Library      RequestsLibrary
#Library      SSHLibrary      timeout=60s
#Library      String

Resource     ../../variables/${VARIABLES}_variables.robot

Resource     ../../libraries/all_libs.robot

# not implemented in v2
Force Tags        crud     IPv4     ExpectedFailure
Suite Setup       Testsuite Setup
Suite Teardown    Testsuite Teardown
Test Setup        TestSetup
Test Teardown     TestTeardown

*** Variables ***
${VARIABLES}=          common
${ENV}=                common
${WAIT_TIMEOUT}=       20s
${SYNC_SLEEP}=         3s
# wait for resync vpps after restart
${RESYNC_WAIT}=        30s
@{segmentList1}    b::    c::    d::
@{segmentList2}    c::    d::    e::
@{segmentList3}    d::    e::    a::
@{segmentList4}    e::    a::    b::
@{segmentList1weight1}    1    @{segmentList1}    # segment list's weight and segments
@{segmentList2weight2}    2    @{segmentList2}    # segment list's weight and segments
@{segmentList3weight3}    3    @{segmentList3}    # segment list's weight and segments
@{segmentList3weight4}    4    @{segmentList3}    # segment list's weight and segments
@{segmentList4weight4}    4    @{segmentList4}    # segment list's weight and segments
@{segmentLists1weight1}        ${segmentList1weight1}
@{segmentLists3weight3}        ${segmentList3weight3}
@{segmentLists3weight4}        ${segmentList3weight4}
@{segmentLists4weight4}        ${segmentList4weight4}
@{segmentLists12weight12}      ${segmentList1weight1}    ${segmentList2weight2}
@{segmentLists23weight23}      ${segmentList2weight2}    ${segmentList3weight3}

*** Test Cases ***
Configure Environment
    [Tags]    setup
    Add Agent VPP Node            agent_vpp_1
    Put Veth Interface With IP    node=agent_vpp_1    name=vpp1_veth1        mac=12:11:11:11:11:11    peer=vpp1_veth2    ip=10.10.1.1
    Put Veth Interface            node=agent_vpp_1    name=vpp1_veth2        mac=12:12:12:12:12:12    peer=vpp1_veth1
    Put Afpacket Interface        node=agent_vpp_1    name=vpp1_afpacket1    mac=a2:a1:a1:a1:a1:a1    host_int=vpp1_veth2
    Put Veth Interface With IP    node=agent_vpp_1    name=vpp1_veth3        mac=12:13:13:13:13:13    peer=vpp1_veth4    ip=10.10.1.2
    Put Veth Interface            node=agent_vpp_1    name=vpp1_veth4        mac=12:14:14:14:14:14    peer=vpp1_veth3
    Put Afpacket Interface        node=agent_vpp_1    name=vpp1_afpacket2    mac=a2:a2:a2:a2:a2:a2    host_int=vpp1_veth4

Check Local SID CRUD
    Put Local SID With End.DX6 function    node=agent_vpp_1    localsidName=A    sidAddress=a::    fibtable=0          outinterface=vpp1_afpacket1           nexthop=a::1
    Wait Until Keyword Succeeds        ${WAIT_TIMEOUT}     ${SYNC_SLEEP}     vpp_term: Check Local SID Presence    node=agent_vpp_1    sidAddress=a::    interface=host-vpp1_veth2    nexthop=a::1
    Put Local SID With End.DX6 function    node=agent_vpp_1    localsidName=A    sidAddress=a::    fibtable=0          outinterface=vpp1_afpacket1           nexthop=c::1   #modification
    Wait Until Keyword Succeeds        ${WAIT_TIMEOUT}     ${SYNC_SLEEP}     vpp_term: Check Local SID Presence    node=agent_vpp_1    sidAddress=a::    interface=host-vpp1_veth2    nexthop=c::1
    Delete Local SID                       node=agent_vpp_1    localsidName=A
    Wait Until Keyword Succeeds        ${WAIT_TIMEOUT}     ${SYNC_SLEEP}     vpp_term: Check Local SID Deleted     node=agent_vpp_1    sidAddress=a::

Check SR-Proxy CRUD
    # SR-proxy is actual a Local SID with End.AD end function, but VPP is configured in this case differently in compare to other local SIDs (VPP CLI(using VPE binary API) vs binary VPP API) -> need to test this
    Put Local SID With End.AD function     node=agent_vpp_1    localsidName=A    sidAddress=a::    l3serviceaddress=10.10.1.2    outinterface=vpp1_afpacket1    ininterface=vpp1_afpacket2
    Wait Until Keyword Succeeds        ${WAIT_TIMEOUT}     ${SYNC_SLEEP}     vpp_term: Check SR-Proxy Local SID Presence    node=agent_vpp_1    sidAddress=a::    serviceaddress=10.10.1.2    outinterface=host-vpp1_veth2    ininterface=host-vpp1_veth4
    Put Local SID With End.AD function     node=agent_vpp_1    localsidName=A    sidAddress=a::    l3serviceaddress=10.10.1.2    outinterface=vpp1_afpacket2    ininterface=vpp1_afpacket1   #modification
    Wait Until Keyword Succeeds        ${WAIT_TIMEOUT}     ${SYNC_SLEEP}     vpp_term: Check SR-Proxy Local SID Presence    node=agent_vpp_1    sidAddress=a::    serviceaddress=10.10.1.2    outinterface=host-vpp1_veth4    ininterface=host-vpp1_veth2
    Delete Local SID                       node=agent_vpp_1    localsidName=A
    Wait Until Keyword Succeeds        ${WAIT_TIMEOUT}     ${SYNC_SLEEP}     vpp_term: Check Local SID Deleted     node=agent_vpp_1    sidAddress=a::

Check Policy CRUD
    Put SRv6 Policy                    node=agent_vpp_1    bsid=a::e    segmentlists=${segmentLists1weight1}    fibtable=0         srhEncapsulation=true      sprayBehaviour=true
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    vpp_term: Check SRv6 Policy Presence        node=agent_vpp_1    bsid=a::e    segmentlists=${segmentLists1weight1}    fibtable=0    behaviour=Encapsulation    type=Spray    index=0
    Put SRv6 Policy                    node=agent_vpp_1    bsid=a::e    segmentlists=${segmentLists1weight1}    fibtable=0         srhEncapsulation=false      sprayBehaviour=true    # modification of non-segment list part
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    vpp_term: Check SRv6 Policy Presence        node=agent_vpp_1    bsid=a::e    segmentlists=${segmentLists1weight1}    fibtable=0    behaviour=SRH insertion    type=Spray    index=0
    Put SRv6 Policy                    node=agent_vpp_1    bsid=a::e    segmentlists=${segmentLists12weight12}    fibtable=0         srhEncapsulation=false      sprayBehaviour=true    # modification - adding one new segment list (+preserving one existing segment key)
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    vpp_term: Check SRv6 Policy Presence        node=agent_vpp_1    bsid=a::e    segmentlists=${segmentLists12weight12}    fibtable=0    behaviour=SRH insertion    type=Spray    index=0
    Put SRv6 Policy                    node=agent_vpp_1    bsid=a::e    segmentlists=${segmentLists23weight23}    fibtable=0         srhEncapsulation=false      sprayBehaviour=true    # modification - mixing addition of mew segment list with removal of existing segment list (+preserving one existing segment key)
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    vpp_term: Check SRv6 Policy Presence        node=agent_vpp_1    bsid=a::e    segmentlists=${segmentLists23weight23}    fibtable=0    behaviour=SRH insertion    type=Spray    index=0
    Put SRv6 Policy                    node=agent_vpp_1    bsid=a::e    segmentlists=${segmentLists3weight3}    fibtable=0         srhEncapsulation=false      sprayBehaviour=true    # modification - removing of existing segment list (+preserving one existing segment key)
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    vpp_term: Check SRv6 Policy Presence        node=agent_vpp_1    bsid=a::e    segmentlists=${segmentLists3weight3}    fibtable=0    behaviour=SRH insertion    type=Spray    index=0
    Put SRv6 Policy                    node=agent_vpp_1    bsid=a::e    segmentlists=${segmentLists3weight4}    fibtable=0         srhEncapsulation=false      sprayBehaviour=true    # modification - modify weight of existing segment list
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    vpp_term: Check SRv6 Policy Presence        node=agent_vpp_1    bsid=a::e    segmentlists=${segmentLists3weight4}    fibtable=0    behaviour=SRH insertion    type=Spray    index=0
    Put SRv6 Policy                    node=agent_vpp_1    bsid=a::e    segmentlists=${segmentLists4weight4}    fibtable=0         srhEncapsulation=false      sprayBehaviour=true    # modification - modify segments of existing segment list
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    vpp_term: Check SRv6 Policy Presence        node=agent_vpp_1    bsid=a::e    segmentlists=${segmentLists4weight4}    fibtable=0    behaviour=SRH insertion    type=Spray    index=0
    Delete SRv6 Policy                 node=agent_vpp_1    bsid=a::e
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    vpp_term: Check SRv6 Policy Nonexistence    node=agent_vpp_1    bsid=a::e

Check Steering CRUD
    Put SRv6 Policy                    node=agent_vpp_1    bsid=a::e       fibtable=0         srhEncapsulation=true    sprayBehaviour=true    segmentlists=${segmentLists1weight1}
    Put SRv6 L3 Steering                  node=agent_vpp_1    name=toE        bsid=a::e          fibtable=0               prefixAddress=e::/64
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    vpp_term: Check SRv6 Steering Presence      node=agent_vpp_1    bsid=a::e    prefixAddress=e::/64
    Put SRv6 L3 Steering               node=agent_vpp_1    name=toE        bsid=a::e          fibtable=0               prefixAddress=d::/64   # modification
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    vpp_term: Check SRv6 Steering Presence      node=agent_vpp_1    bsid=a::e    prefixAddress=d::/64
    Delete SRv6 Steering               node=agent_vpp_1    name=toE
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    vpp_term: Check SRv6 Steering NonExistence  node=agent_vpp_1    bsid=a::e    prefixAddress=d::/64
    Delete SRv6 Policy                 node=agent_vpp_1    bsid=a::e       #cleanup

#TODO Steering can reference policy also by index -> add test (currently NOT WORKING on VPP side!)

Check Steering Creation By Using Reversed-Ordered Steering And Policy Setting (Delayed Configuration)
    Put SRv6 L3 Steering                  node=agent_vpp_1    name=toE        bsid=a::e          fibtable=0               prefixAddress=e::/64
    Put SRv6 Policy                    node=agent_vpp_1    bsid=a::e       fibtable=0         srhEncapsulation=true    sprayBehaviour=true    segmentlists=${segmentLists1weight1}
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    vpp_term: Check SRv6 Steering Presence      node=agent_vpp_1    bsid=a::e    prefixAddress=e::/64
    Delete SRv6 Steering               node=agent_vpp_1    name=toE        #cleanup
    Delete SRv6 Policy                 node=agent_vpp_1    bsid=a::e       #cleanup

Check Steering Delete By Removal of Policy
    Put SRv6 Policy                    node=agent_vpp_1    bsid=a::e       fibtable=0         srhEncapsulation=true    sprayBehaviour=true    segmentlists=${segmentLists1weight1}
    Put SRv6 L3 Steering                  node=agent_vpp_1    name=toE        bsid=a::e          fibtable=0               prefixAddress=e::/64
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    vpp_term: Check SRv6 Steering Presence      node=agent_vpp_1    bsid=a::e    prefixAddress=e::/64
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    vpp_term: Check SRv6 Policy Presence        node=agent_vpp_1    bsid=a::e    fibtable=0    behaviour=Encapsulation    type=Spray    index=0    segmentlists=${segmentLists1weight1}
    Delete SRv6 Policy                 node=agent_vpp_1    bsid=a::e
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    vpp_term: Check SRv6 Policy Nonexistence    node=agent_vpp_1    bsid=a::e
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    vpp_term: Check SRv6 Steering NonExistence  node=agent_vpp_1    bsid=a::e    prefixAddress=e::/64
    Delete SRv6 Steering               node=agent_vpp_1    name=toE        #cleanup

Check Update Of Policy Should Not Trigger Cascade Delete Of Steering
    Put SRv6 L3 Steering                  node=agent_vpp_1    name=toE        bsid=a::e          fibtable=0               prefixAddress=e::/64
    Put SRv6 Policy                    node=agent_vpp_1    bsid=a::e       fibtable=0         srhEncapsulation=true    sprayBehaviour=true    segmentlists=${segmentLists1weight1}
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    vpp_term: Check SRv6 Steering Presence      node=agent_vpp_1    bsid=a::e    prefixAddress=e::/64
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    vpp_term: Check SRv6 Policy Presence        node=agent_vpp_1    bsid=a::e    fibtable=0    behaviour=Encapsulation    type=Spray    index=0    segmentlists=${segmentLists1weight1}
    Put SRv6 Policy                    node=agent_vpp_1    bsid=a::e       fibtable=0         srhEncapsulation=false    sprayBehaviour=true    segmentlists=${segmentLists1weight1}    #modification of non-segment attribute
    Sleep    5s    # waiting constant time is bad practice, but how to otherwise check that nothing has changed?
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    vpp_term: Check SRv6 Steering Presence      node=agent_vpp_1    bsid=a::e    prefixAddress=e::/64
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    vpp_term: Check SRv6 Policy Presence        node=agent_vpp_1    bsid=a::e    fibtable=0    behaviour=SRH insertion    type=Spray    index=0    segmentlists=${segmentLists1weight1}
    Delete SRv6 Steering               node=agent_vpp_1    name=toE        #cleanup
    Delete SRv6 Policy                 node=agent_vpp_1    bsid=a::e       #cleanup

Check Resynchronization for clean VPP start
    Put Local SID With End.DX6 function    node=agent_vpp_1    localsidName=A    sidAddress=a::     fibtable=0               outinterface=vpp1_afpacket1    nexthop=a::1
    Put SRv6 Policy                        node=agent_vpp_1    bsid=a::e         fibtable=0         srhEncapsulation=true    sprayBehaviour=true            segmentlists=${segmentLists1weight1}
    Put SRv6 L3 Steering                      node=agent_vpp_1    name=toE          bsid=a::e          fibtable=0               prefixAddress=e::/64
    Remove All VPP Nodes
    Sleep                                       3s
    Add Agent VPP Node                          agent_vpp_1
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    vpp_term: Check Local SID Presence          node=agent_vpp_1    sidAddress=a::    interface=host-vpp1_veth2    nexthop=a::1
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    vpp_term: Check SRv6 Policy Presence        node=agent_vpp_1    bsid=a::e         fibtable=0                   behaviour=Encapsulation    type=Spray    index=0    segmentlists=${segmentLists1weight1}
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    vpp_term: Check SRv6 Steering Presence      node=agent_vpp_1    bsid=a::e         prefixAddress=e::/64
    Delete SRv6 Steering                   node=agent_vpp_1    name=toE        #cleanup
    Delete SRv6 Policy                     node=agent_vpp_1    bsid=a::e       #cleanup
    Delete Local SID                       node=agent_vpp_1    localsidName=A

*** Keywords ***
TestSetup
    Make Datastore Snapshots    ${TEST_NAME}_test_setup

TestTeardown
    Make Datastore Snapshots    ${TEST_NAME}_test_teardown