*** Settings ***
Library      OperatingSystem
#Library      RequestsLibrary
#Library      SSHLibrary      timeout=60s
#Library      String

Resource     ../../variables/${VARIABLES}_variables.robot

Resource     ../../libraries/all_libs.robot

Suite Setup       Testsuite Setup
# Suite Teardown    Testsuite Teardown

*** Variables ***
${VARIABLES}=          common
${ENV}=                common
${CONFIG_SLEEP}=       1s
${RESYNC_SLEEP}=       1s
# wait for resync vpps after restart
${RESYNC_WAIT}=        30s

*** Test Cases ***
Configure Environment
    [Tags]    setup
    Add Agent VPP Node    agent_vpp_1

Show Interfaces Before Setup
    vpp_term: Show Interfaces    agent_vpp_1
    Write To Machine    vpp_agent_ctl    vpp-agent-ctl ${AGENT_VPP_ETCD_CONF_PATH} -ps

Setup Interfaces
    vpp_ctl: Put Veth Interface Via Linux Plugin    node=agent_vpp_1    namespace=ns1    name=ns1_veth1    host_if_name=ns1_veth1_linux    mac=d2:74:8c:12:67:d2    peer=ns2_veth2    ip=192.168.22.1
    vpp_ctl: Put Veth Interface Via Linux Plugin    node=agent_vpp_1    namespace=ns2    name=ns2_veth2    host_if_name=ns2_veth2_linux    mac=92:c7:42:67:ab:cd    peer=ns1_veth1    ip=192.168.22.2

    Check Linux Interfaces    node=agent_vpp_1    namespace=ns1    interface=ns1_veth1
    Check Linux Interfaces    node=agent_vpp_1    namespace=ns2    interface=ns2_veth2

    Check Linux Routes    node=agent_vpp_1    namespace=ns1    ip=192.168.22.0
    Check Linux Routes    node=agent_vpp_1    namespace=ns2    ip=192.168.22.0

    Ping in namespace    node=agent_vpp_1    namespace=ns1    ip=192.168.22.2
    Ping in namespace    node=agent_vpp_1    namespace=ns2    ip=192.168.22.1

    Sleep    10

Create Linux Routes
    vpp_ctl: Put Linux Route    node=agent_vpp_1    namespace=ns1    interface=ns1_veth1    routename=pingingveth2    ip=192.168.22.2    prefix=32    next_hop=192.168.22.1
    vpp_ctl: Put Linux Route    node=agent_vpp_1    namespace=ns2    interface=ns2_veth2    routename=pingingveth1    ip=192.168.22.1    prefix=32    next_hop=192.168.22.2
    vpp_ctl: Put Linux Route    node=agent_vpp_1    namespace=ns1    interface=ns1_veth1    routename=pinginggoogl    ip=8.8.8.8    prefix=32    next_hop=192.168.22.1
    vpp_ctl: Put Linux Route    node=agent_vpp_1    namespace=ns2    interface=ns2_veth2    routename=pinging9    ip=9.9.9.9    prefix=32    next_hop=192.168.22.2

    Sleep    10

    Check Linux Routes    node=agent_vpp_1    namespace=ns1    ip=192.168.22.2
    Check Linux Routes    node=agent_vpp_1    namespace=ns2    ip=192.168.22.1
    Check Linux Routes    node=agent_vpp_1    namespace=ns1    ip=8.8.8.8
    Check Linux Routes    node=agent_vpp_1    namespace=ns2    ip=9.9.9.9

    # # created routes should not exist in other namespace
    # Check Removed Linux Route    node=agent_vpp_1    namespace=ns2    ip=192.168.22.2 - su obsihnute v gateway
    # Check Removed Linux Route    node=agent_vpp_1    namespace=ns1    ip=192.168.22.1 - su obsihnute v gateway
    # Check Removed Linux Route    node=agent_vpp_1    namespace=ns2    ip=8.8.8.8
    # Check Removed Linux Route    node=agent_vpp_1    namespace=ns1    ip=9.9.9.9

Change Linux Routes Without Deleting Key (Changing Gateway)
    # changing of gateway
    vpp_ctl: Put Linux Route    node=agent_vpp_1    namespace=ns1    interface=ns1_veth1    routename=pinginggoogl    ip=8.8.8.8    prefix=32    next_hop=192.168.22.55
    Sleep    10

    Check Linux Routes    node=agent_vpp_1    namespace=ns1    ip=8.8.8.8
    # testing if there is the new gateway
    Check Linux Routes    node=agent_vpp_1    namespace=ns1    ip=192.168.22.55

Change Linux Routes At First Deleting Key And Putting The Same Secondly Deleting Key Then Putting It To Other Namespace
    vpp_ctl: Delete Linux Route    node=agent_vpp_1    routename=pinging9
    Check Removed Linux Route    node=agent_vpp_1    namespace=ns2    ip=9.9.9.9
    # we create exactly the same as deleted route
    vpp_ctl: Put Linux Route    node=agent_vpp_1    namespace=ns2    interface=ns2_veth2    routename=pinging9    ip=9.9.9.9    prefix=32    next_hop=192.168.22.2
    Sleep    10

    Check Linux Routes    node=agent_vpp_1    namespace=ns2    ip=9.9.9.9
    # delete again
    vpp_ctl: Delete Linux Route    node=agent_vpp_1    routename=pinging9
    Check Removed Linux Route    node=agent_vpp_1    namespace=ns2    ip=9.9.9.9
    # we try to transfer route to other namespace
    vpp_ctl: Put Linux Route    node=agent_vpp_1    namespace=ns1    interface=ns1_veth1    routename=pinging9    ip=9.9.9.9    prefix=32    next_hop=192.168.22.2
    Sleep    10

    Check Removed Linux Route    node=agent_vpp_1    namespace=ns2    ip=9.9.9.9
    Check Linux Routes    node=agent_vpp_1    namespace=ns1    ip=9.9.9.9

At first create route and after that create inteface in namespace 3
    vpp_ctl: Put Linux Route    node=agent_vpp_1    namespace=ns3    interface=ns3_veth3    routename=pingingns2_veth3    ip=192.169.22.22    prefix=32    next_hop=192.169.22.3
    vpp_ctl: Put Linux Route    node=agent_vpp_1    namespace=ns3    interface=ns3_veth3    routename=pingingns2_veth2    ip=192.168.22.2    prefix=32    next_hop=192.169.22.3
    vpp_ctl: Put Linux Route    node=agent_vpp_1    namespace=ns3    interface=ns3_veth3    routename=pingingns1_veth1    ip=192.168.22.1    prefix=32    next_hop=192.169.22.3
    vpp_ctl: Put Linux Route    node=agent_vpp_1    namespace=ns2    interface=ns2_veth3    routename=pingingns3_veth3    ip=192.169.22.3    prefix=32    next_hop=192.169.22.22

    Sleep    10

    vpp_ctl: Put Veth Interface Via Linux Plugin    node=agent_vpp_1    namespace=ns3    name=ns3_veth3    host_if_name=ns3_veth3_linux    mac=92:c7:42:67:ab:ce    peer=ns2_veth3    ip=192.169.22.3
    vpp_ctl: Put Veth Interface Via Linux Plugin    node=agent_vpp_1    namespace=ns2    name=ns2_veth3    host_if_name=ns2_veth3_linux    mac=92:c7:42:67:ab:cf    peer=ns3_veth3    ip=192.169.22.22

    Sleep    10

    Check Linux Interfaces    node=agent_vpp_1    namespace=ns3    interface=ns3_veth3
    Check Linux Interfaces    node=agent_vpp_1    namespace=ns2    interface=ns2_veth3

    Check Linux Routes    node=agent_vpp_1    namespace=ns2    ip=192.169.22.0
    Check Linux Routes    node=agent_vpp_1    namespace=ns3    ip=192.169.22.0

    Ping in namespace    node=agent_vpp_1    namespace=ns2    ip=192.169.22.3
    Ping in namespace    node=agent_vpp_1    namespace=ns3    ip=192.169.22.22

    Check Linux Routes    node=agent_vpp_1    namespace=ns3    ip=192.168.22.1
    Check Linux Routes    node=agent_vpp_1    namespace=ns3    ip=192.168.22.2
    Check Linux Routes    node=agent_vpp_1    namespace=ns3    ip=192.169.22.22
    Check Linux Routes    node=agent_vpp_1    namespace=ns2    ip=192.169.22.3

    # tested also above, but repeat after giving exact routes
    Ping in namespace    node=agent_vpp_1    namespace=ns3    ip=192.169.22.22
    Ping in namespace    node=agent_vpp_1    namespace=ns2    ip=192.169.22.3
    # this works
    Ping in namespace    node=agent_vpp_1    namespace=ns3    ip=192.168.22.2

Create inteface Then Routes in namespace 3
    vpp_ctl: Put Veth Interface Via Linux Plugin    node=agent_vpp_1    namespace=ns3    name=ns3_veth3    host_if_name=ns3_veth3_linux    mac=92:c7:42:67:ab:ce    peer=ns2_veth3    ip=192.169.22.3
    vpp_ctl: Put Veth Interface Via Linux Plugin    node=agent_vpp_1    namespace=ns2    name=ns2_veth3    host_if_name=ns2_veth3_linux    mac=92:c7:42:67:ab:cf    peer=ns3_veth3    ip=192.169.22.22

    Check Linux Interfaces    node=agent_vpp_1    namespace=ns3    interface=ns3_veth3
    Check Linux Interfaces    node=agent_vpp_1    namespace=ns2    interface=ns2_veth3

    Check Linux Routes    node=agent_vpp_1    namespace=ns2    ip=192.169.22.0
    Check Linux Routes    node=agent_vpp_1    namespace=ns3    ip=192.169.22.0

    Ping in namespace    node=agent_vpp_1    namespace=ns2    ip=192.169.22.3
    Ping in namespace    node=agent_vpp_1    namespace=ns3    ip=192.169.22.22

    Sleep    10
    vpp_ctl: Put Linux Route    node=agent_vpp_1    namespace=ns3    interface=ns3_veth3    routename=pingingns2_veth3    ip=192.169.22.22    prefix=32    next_hop=192.169.22.3
    Sleep    10
    vpp_ctl: Put Linux Route    node=agent_vpp_1    namespace=ns3    interface=ns3_veth3    routename=pingingns2_veth2    ip=192.168.22.2    prefix=32    next_hop=192.169.22.3
    Sleep    10
    #vpp_ctl: Put Linux Route    node=agent_vpp_1    namespace=ns3    interface=ns3_veth3    routename=pingingns1_veth1    ip=192.168.22.1    prefix=32    next_hop=192.169.22.22
    vpp_ctl: Put Linux Route    node=agent_vpp_1    namespace=ns3    interface=ns3_veth3    routename=pingingns1_veth1    ip=192.168.22.1    prefix=32    next_hop=192.169.22.3
    Sleep    10
    #vpp_ctl: Put Linux Route    node=agent_vpp_1    namespace=ns2    interface=ns2_veth3    routename=pingingns3_veth3    ip=192.169.22.3    prefix=32    next_hop=192.169.22.22
    vpp_ctl: Put Linux Route    node=agent_vpp_1    namespace=ns2    interface=ns2_veth3    routename=pingingns3_veth3    ip=192.169.22.3    prefix=32    next_hop=192.169.22.22
    Sleep    10

    Check Linux Routes    node=agent_vpp_1    namespace=ns3    ip=192.168.22.1
    Check Linux Routes    node=agent_vpp_1    namespace=ns3    ip=192.168.22.2
    Check Linux Routes    node=agent_vpp_1    namespace=ns3    ip=192.169.22.22
    Check Linux Routes    node=agent_vpp_1    namespace=ns2    ip=192.169.22.3

    # tested also above, but repeat after giving exact routes
    Ping in namespace    node=agent_vpp_1    namespace=ns3    ip=192.169.22.22
    Ping in namespace    node=agent_vpp_1    namespace=ns2    ip=192.169.22.3
    # this works
    Ping in namespace    node=agent_vpp_1    namespace=ns3    ip=192.168.22.2
    # this does not work
    # https://serverfault.com/questions/568839/linux-network-namespaces-ping-fails-on-specific-veth
    # https://unix.stackexchange.com/questions/391193/how-to-forward-traffic-between-linux-network-namespaces
    #Ping in namespace    node=agent_vpp_1    namespace=ns3    ip=192.168.22.1


    # routy sa zalozia po uspesnom pingu zo ns3 ?! or ping fails
    Ping in namespace    node=agent_vpp_1    namespace=ns1    ip=192.169.22.3



# Config Done
#     No Operation
#
# Final Sleep After Config For Manual Checking
#     Sleep   ${CONFIG_SLEEP}
#
# Remove VPP Nodes
#     Remove All Nodes
#
# Start VPP1 And VPP2 Again
#     Add Agent VPP Node    agent_vpp_1
#     Sleep    ${RESYNC_WAIT}
#
# Check Linux Interfaces On VPP1 After Resync
#     ${out}=    Execute In Container    agent_vpp_1    ip netns exec ns1 ip a
#     Log    ${out}
#     Should Contain    ${out}    ns1_veth1_linux
#
#     ${out}=    Execute In Container    agent_vpp_1    ip netns exec ns2 ip a
#     Log    ${out}
#     Should Contain    ${out}    ns2_veth2_linux


*** Keywords ***
Check Linux Interfaces
    [Arguments]    ${node}    ${namespace}    ${interface}
    Log Many    ${node}    ${namespace}    ${interface}
    ${out}=    Execute In Container    ${node}    ip netns exec ${namespace} ip a
    Log    ${out}
    Should Contain    ${out}    ${interface}

Check Linux Routes
    [Arguments]    ${node}    ${namespace}    ${ip}
    Log Many    ${node}    ${namespace}
    ${out}=    Execute In Container    ${node}    ip netns exec ${namespace} ip route show
    Log    ${out}
    Should Contain    ${out}    ${ip}

Check Removed Linux Route
    [Arguments]    ${node}    ${namespace}    ${ip}
    Log Many    ${node}    ${namespace}
    ${out}=    Execute In Container    ${node}    ip netns exec ${namespace} ip route show
    Log    ${out}
    Should Not Contain    ${out}    ${ip}

Ping in namespace
    [Arguments]    ${node}    ${namespace}    ${ip}
    ${out}=    Execute In Container    ${node}    ip netns exec ${namespace} ping -c 5 ${ip}
    Log    ${out}
    Should Contain     ${out}    from ${ip}
    Should Not Contain    ${out}    100% packet loss
