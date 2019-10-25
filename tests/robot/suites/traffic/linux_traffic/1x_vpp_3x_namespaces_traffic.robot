*** Settings ***
Library      OperatingSystem
#Library      RequestsLibrary
#Library      SSHLibrary      timeout=60s
#Library      String

Resource     ../../../variables/${VARIABLES}_variables.robot

Resource     ../../../libraries/all_libs.robot

Force Tags        traffic     IPv4
Suite Setup       Testsuite Setup
Suite Teardown    Testsuite Teardown
Test Setup        TestSetup
Test Teardown     TestTeardown

*** Variables ***
${VARIABLES}=          common
${ENV}=                common
${WAIT_TIMEOUT}=       20s
${SYNC_SLEEP}=         3s
${RESYNC_SLEEP}=       1s
# wait for resync vpps after restart
${RESYNC_WAIT}=        30s
${IP_1}=               192.168.22.1
${IP_2}=               192.168.22.2
${IP_3}=               192.168.22.5
${IP_4}=               192.168.22.6
${PREFIX}=             30

*** Test Cases ***
Configure Environment
    [Tags]    setup
    Add Agent VPP Node    agent_vpp_1

Show Interfaces Before Setup
    vpp_term: Show Interfaces    agent_vpp_1

Setup Interfaces
    Put Veth Interface Via Linux Plugin    node=agent_vpp_1    namespace=ns1
    ...    name=ns1_veth1    host_if_name=ns1_veth1_linux    mac=d2:74:8c:12:67:d2
    ...    peer=ns2_veth2    ip=${IP_1}    prefix=${PREFIX}
    Put Veth Interface Via Linux Plugin    node=agent_vpp_1    namespace=ns2
    ...    name=ns2_veth2    host_if_name=ns2_veth2_linux    mac=92:c7:42:67:ab:cd
    ...    peer=ns1_veth1    ip=${IP_2}    prefix=${PREFIX}

    Put Veth Interface Via Linux Plugin    node=agent_vpp_1    namespace=ns2
    ...    name=ns2_veth3    host_if_name=ns2_veth3_linux    mac=92:c7:42:67:ab:cf
    ...    peer=ns3_veth3    ip=${IP_3}    prefix=${PREFIX}
    Put Veth Interface Via Linux Plugin    node=agent_vpp_1    namespace=ns3
    ...    name=ns3_veth3    host_if_name=ns3_veth3_linux    mac=92:c7:42:67:ab:ce
    ...    peer=ns2_veth3    ip=${IP_4}    prefix=${PREFIX}
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    Check Linux Interfaces    node=agent_vpp_1    namespace=ns1    interface=ns1_veth1
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    Check Linux Interfaces    node=agent_vpp_1    namespace=ns2    interface=ns2_veth2
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    Check Linux Interfaces    node=agent_vpp_1    namespace=ns2    interface=ns2_veth3
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    Check Linux Interfaces    node=agent_vpp_1    namespace=ns3    interface=ns3_veth3

Ping Within The Same Namespace
    Ping in namespace    node=agent_vpp_1    namespace=ns1    ip=${IP_2}
    Ping in namespace    node=agent_vpp_1    namespace=ns2    ip=${IP_1}
    Ping in namespace    node=agent_vpp_1    namespace=ns2    ip=${IP_4}
    Ping in namespace    node=agent_vpp_1    namespace=ns3    ip=${IP_3}

Create Linux Routes
    Put Linux Route    node=agent_vpp_1    namespace=ns1    interface=ns1_veth1
    ...    ip=${IP_4}    prefix=32    next_hop=${IP_2}
    Put Linux Route    node=agent_vpp_1    namespace=ns3    interface=ns3_veth3
    ...    ip=${IP_1}    prefix=32    next_hop=${IP_3}
    # Enable forwarding in namespace ns2
    Execute In Container    agent_vpp_1    ip netns exec ns2 sysctl -w net.ipv4.ip_forward=1

Ping Across Namespaces
    Ping in namespace    node=agent_vpp_1    namespace=ns1    ip=${IP_4}
    Ping in namespace    node=agent_vpp_1    namespace=ns3    ip=${IP_1}

Create Linux Default Routes
    Put Default Linux Route    node=agent_vpp_1    namespace=ns1    interface=ns1_veth1
    ...    next_hop=${IP_2}
    Put Default Linux Route    node=agent_vpp_1    namespace=ns3    interface=ns3_veth3
    ...    next_hop=${IP_3}
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    Check Linux Default Routes    node=agent_vpp_1    namespace=ns1    next_hop=${IP_2}
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}
    ...    Check Linux Default Routes    node=agent_vpp_1    namespace=ns3    next_hop=${IP_3}

Ping Across Namespaces Through Default Route
    Ping in namespace via interface    node=agent_vpp_1    namespace=ns1    ip=${IP_3}    interface=ns1_veth1_linux
    Ping in namespace via interface    node=agent_vpp_1    namespace=ns3    ip=${IP_2}    interface=ns3_veth3_linux

Restart VPP Node
    Remove All Nodes
    Sleep   ${RESYNC_SLEEP}
    Add Agent VPP Node    agent_vpp_1
    Sleep    ${RESYNC_WAIT}
    Execute In Container    agent_vpp_1    ip netns exec ns2 sysctl -w net.ipv4.ip_forward=1


Check Linux Interfaces On VPP1 After Resync
    ${out}=    Execute In Container    agent_vpp_1    ip netns exec ns1 ip address
    Should Contain    ${out}    ns1_veth1_linux

    ${out}=    Execute In Container    agent_vpp_1    ip netns exec ns2 ip address
    Should Contain    ${out}    ns2_veth2_linux
    Should Contain    ${out}    ns2_veth3_linux

    ${out}=    Execute In Container    agent_vpp_1    ip netns exec ns3 ip address
    Should Contain    ${out}    ns3_veth3_linux

    Check Linux Interfaces    node=agent_vpp_1    namespace=ns1    interface=ns1_veth1
    Check Linux Interfaces    node=agent_vpp_1    namespace=ns2    interface=ns2_veth2

    Check Linux Interfaces    node=agent_vpp_1    namespace=ns2    interface=ns2_veth3
    Check Linux Interfaces    node=agent_vpp_1    namespace=ns3    interface=ns3_veth3

Retry All Pings After Resync
    Ping in namespace    node=agent_vpp_1    namespace=ns1    ip=${IP_2}
    Ping in namespace    node=agent_vpp_1    namespace=ns2    ip=${IP_1}

    Ping in namespace    node=agent_vpp_1    namespace=ns2    ip=${IP_4}
    Ping in namespace    node=agent_vpp_1    namespace=ns3    ip=${IP_3}

    Ping in namespace    node=agent_vpp_1    namespace=ns1    ip=${IP_3}
    Ping in namespace    node=agent_vpp_1    namespace=ns3    ip=${IP_2}

    Ping in namespace    node=agent_vpp_1    namespace=ns1    ip=${IP_4}
    Ping in namespace    node=agent_vpp_1    namespace=ns3    ip=${IP_1}

*** Keywords ***
Check Linux Interfaces
    [Arguments]    ${node}    ${namespace}    ${interface}
    ${out}=    Execute In Container    ${node}    ip netns exec ${namespace} ip address
    Should Contain    ${out}    ${interface}

Check Linux Routes
    [Arguments]    ${node}    ${namespace}    ${ip}
    ${out}=    Execute In Container    ${node}    ip netns exec ${namespace} ip route show
    Should Contain    ${out}    ${ip} via

Check Linux Routes Gateway
    [Arguments]    ${node}    ${namespace}    ${ip}    ${next_hop}=${EMPTY}
    ${out}=    Execute In Container    ${node}    ip netns exec ${namespace} ip route show
    Should Contain    ${out}    ${ip} via ${next_hop}

Check Linux Default Routes
    [Arguments]    ${node}    ${namespace}    ${next_hop}
    ${out}=    Execute In Container    ${node}    ip netns exec ${namespace} ip route show
    Should Contain    ${out}    default via ${next_hop}

Check Linux Routes Metric
    [Arguments]    ${node}    ${namespace}    ${ip}    ${metric}
    ${out}=    Execute In Container    ${node}    ip netns exec ${namespace} ip route show
    Should Match Regexp    ${out}    ${ip} via.*metric ${metric}\\s

Check Removed Linux Route
    [Arguments]    ${node}    ${namespace}    ${ip}
    ${out}=    Execute In Container    ${node}    ip netns exec ${namespace} ip route show
    Should Not Contain    ${out}    ${ip} via

Ping in namespace
    [Arguments]    ${node}    ${namespace}    ${ip}
    ${out}=    Execute In Container    ${node}    ip netns exec ${namespace} ping -c 5 ${ip}
    Should Contain     ${out}    from ${ip}
    Should Not Contain    ${out}    100% packet loss

Ping in namespace via interface
    [Arguments]    ${node}    ${namespace}    ${ip}    ${interface}
    ${out}=    Execute In Container    ${node}    ip netns exec ${namespace} ping -c 5 -I ${interface} ${ip}
    Should Contain     ${out}    from ${ip}
    Should Not Contain    ${out}    100% packet loss

TestSetup
    Make Datastore Snapshots    ${TEST_NAME}_test_setup

TestTeardown
    Make Datastore Snapshots    ${TEST_NAME}_test_teardown