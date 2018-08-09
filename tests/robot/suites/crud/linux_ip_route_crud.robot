*** Settings ***
Library      OperatingSystem
#Library      RequestsLibrary
#Library      SSHLibrary      timeout=60s
#Library      String

Resource     ../../variables/${VARIABLES}_variables.robot

Resource     ../../libraries/all_libs.robot

Force Tags        crudIPv4
Suite Setup       Testsuite Setup
Suite Teardown    Testsuite Teardown
Test Setup        TestSetup
Test Teardown     TestTeardown

*** Variables ***
${VARIABLES}=          common
${ENV}=                common
${WAIT_TIMEOUT}=     20s
${SYNC_SLEEP}=       2s
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

    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    Check Linux Interfaces    node=agent_vpp_1    namespace=ns1    interface=ns1_veth1
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    Check Linux Interfaces    node=agent_vpp_1    namespace=ns2    interface=ns2_veth2

    # This should work by default after veth interface setup
    Ping in namespace    node=agent_vpp_1    namespace=ns1    ip=192.168.22.2
    Ping in namespace    node=agent_vpp_1    namespace=ns2    ip=192.168.22.1

Create Linux Routes
    vpp_ctl: Put Linux Route    node=agent_vpp_1    namespace=ns1    interface=ns1_veth1    routename=pingingveth2    ip=192.168.22.2    prefix=32    next_hop=192.168.22.1
    vpp_ctl: Put Linux Route    node=agent_vpp_1    namespace=ns2    interface=ns2_veth2    routename=pingingveth1    ip=192.168.22.1    prefix=32    next_hop=192.168.22.2
    vpp_ctl: Put Linux Route    node=agent_vpp_1    namespace=ns1    interface=ns1_veth1    routename=pinginggoogl    ip=8.8.8.8    prefix=32    next_hop=192.168.22.1
    vpp_ctl: Put Linux Route    node=agent_vpp_1    namespace=ns2    interface=ns2_veth2    routename=pinging9    ip=9.9.9.9    prefix=32    next_hop=192.168.22.2

    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    Check Linux Routes    node=agent_vpp_1    namespace=ns1    ip=192.168.22.2
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    Check Linux Routes    node=agent_vpp_1    namespace=ns2    ip=192.168.22.1
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    Check Linux Routes    node=agent_vpp_1    namespace=ns1    ip=8.8.8.8
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    Check Linux Routes    node=agent_vpp_1    namespace=ns2    ip=9.9.9.9

    # created routes should not exist in other namespace
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    Check Removed Linux Route    node=agent_vpp_1    namespace=ns2    ip=192.168.22.2
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    Check Removed Linux Route    node=agent_vpp_1    namespace=ns1    ip=192.168.22.1
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    Check Removed Linux Route    node=agent_vpp_1    namespace=ns2    ip=8.8.8.8
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    Check Removed Linux Route    node=agent_vpp_1    namespace=ns1    ip=9.9.9.9

Read Route Information From Setup Database
    vpp_ctl: Get Linux Route As Json    node=agent_vpp_1    routename=pingingveth2
    vpp_ctl: Get Linux Route As Json    node=agent_vpp_1    routename=pingingveth1
    vpp_ctl: Get Linux Route As Json    node=agent_vpp_1    routename=pinginggoogl
    vpp_ctl: Get Linux Route As Json    node=agent_vpp_1    routename=pinging9

Change Linux Routes Without Deleting Key (Changing Metric)
    # changing of gateway - this is incorrect/ the record would not be put in the database  - Let us change metric
    vpp_ctl: Put Linux Route    node=agent_vpp_1    namespace=ns1    interface=ns1_veth1    routename=pinginggoogl    ip=8.8.8.8    prefix=32    next_hop=192.168.22.1    metric=55

    # testing if there is the new metric
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    Check Linux Routes Metric    node=agent_vpp_1    namespace=ns1    ip=8.8.8.8    metric=55

Change Linux Routes At First Deleting Key And Putting The Same Secondly Deleting Key Then Putting It To Other Namespace
    vpp_ctl: Delete Linux Route    node=agent_vpp_1    routename=pinging9


    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    Check Removed Linux Route    node=agent_vpp_1    namespace=ns2    ip=9.9.9.9

    # we create exactly the same as deleted route
    vpp_ctl: Put Linux Route    node=agent_vpp_1    namespace=ns2    interface=ns2_veth2    routename=pinging9    ip=9.9.9.9    prefix=32    next_hop=192.168.22.2

    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    Check Linux Routes    node=agent_vpp_1    namespace=ns2    ip=9.9.9.9

    # delete again
    vpp_ctl: Delete Linux Route    node=agent_vpp_1    routename=pinging9

    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    Check Removed Linux Route    node=agent_vpp_1    namespace=ns2    ip=9.9.9.9

    # we try to transfer route to other namespace - there is also need to change appropriately gateway
    vpp_ctl: Put Linux Route    node=agent_vpp_1    namespace=ns1    interface=ns1_veth1    routename=pinging9    ip=9.9.9.9    prefix=32    next_hop=192.168.22.1


    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    Check Removed Linux Route    node=agent_vpp_1    namespace=ns2    ip=9.9.9.9
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    Check Linux Routes Gateway    node=agent_vpp_1    namespace=ns1    ip=9.9.9.9    next_hop=192.168.22.1

At first create route and after that create inteface in namespace 3
    vpp_ctl: Put Linux Route    node=agent_vpp_1    namespace=ns3    interface=ns3_veth3    routename=pingingns2_veth3    ip=192.169.22.22    prefix=32    next_hop=192.169.22.3
    vpp_ctl: Put Linux Route    node=agent_vpp_1    namespace=ns3    interface=ns3_veth3    routename=pingingns2_veth2    ip=192.168.22.2    prefix=32    next_hop=192.169.22.3
    vpp_ctl: Put Linux Route    node=agent_vpp_1    namespace=ns3    interface=ns3_veth3    routename=pingingns1_veth1    ip=192.168.22.1    prefix=32    next_hop=192.169.22.3
    vpp_ctl: Put Linux Route    node=agent_vpp_1    namespace=ns2    interface=ns2_veth3    routename=pingingns3_veth3    ip=192.169.22.3    prefix=32    next_hop=192.169.22.22

    vpp_ctl: Put Veth Interface Via Linux Plugin    node=agent_vpp_1    namespace=ns3    name=ns3_veth3    host_if_name=ns3_veth3_linux    mac=92:c7:42:67:ab:ce    peer=ns2_veth3    ip=192.169.22.3
    vpp_ctl: Put Veth Interface Via Linux Plugin    node=agent_vpp_1    namespace=ns2    name=ns2_veth3    host_if_name=ns2_veth3_linux    mac=92:c7:42:67:ab:cf    peer=ns3_veth3    ip=192.169.22.22

    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    Check Linux Interfaces    node=agent_vpp_1    namespace=ns3    interface=ns3_veth3
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    Check Linux Interfaces    node=agent_vpp_1    namespace=ns2    interface=ns2_veth3

    Ping in namespace    node=agent_vpp_1    namespace=ns2    ip=192.169.22.3
    Ping in namespace    node=agent_vpp_1    namespace=ns3    ip=192.169.22.22

    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    Check Linux Routes    node=agent_vpp_1    namespace=ns3    ip=192.168.22.1
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    Check Linux Routes    node=agent_vpp_1    namespace=ns3    ip=192.168.22.2
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    Check Linux Routes    node=agent_vpp_1    namespace=ns3    ip=192.169.22.22
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    Check Linux Routes    node=agent_vpp_1    namespace=ns2    ip=192.169.22.3

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
    # Ping in namespace    node=agent_vpp_1    namespace=ns1    ip=192.169.22.3

Check linux Routes On VPP1
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    Check Linux Routes    node=agent_vpp_1    namespace=ns1    ip=192.168.22.2
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    Check Linux Routes    node=agent_vpp_1    namespace=ns2    ip=192.168.22.1
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    Check Linux Routes Gateway    node=agent_vpp_1    namespace=ns1    ip=8.8.8.8    next_hop=192.168.22.1
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    Check Linux Routes Gateway    node=agent_vpp_1    namespace=ns1    ip=9.9.9.9    next_hop=192.168.22.1
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    Check Linux Routes    node=agent_vpp_1    namespace=ns3    ip=192.168.22.1
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    Check Linux Routes    node=agent_vpp_1    namespace=ns3    ip=192.168.22.2
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    Check Linux Routes    node=agent_vpp_1    namespace=ns3    ip=192.169.22.22
    Wait Until Keyword Succeeds   ${WAIT_TIMEOUT}   ${SYNC_SLEEP}    Check Linux Routes    node=agent_vpp_1    namespace=ns2    ip=192.169.22.3

Remove VPP Nodes
    Remove All Nodes
    Sleep    ${RESYNC_SLEEP}

Start VPP1 Again
    Add Agent VPP Node    agent_vpp_1
    Sleep    ${RESYNC_WAIT}

Check Linux Interfaces On VPP1 After Resync
    ${out}=    Execute In Container    agent_vpp_1    ip netns exec ns1 ip a
    Log    ${out}
    Should Contain    ${out}    ns1_veth1_linux

    ${out}=    Execute In Container    agent_vpp_1    ip netns exec ns2 ip a
    Log    ${out}
    Should Contain    ${out}    ns2_veth2_linux
    Should Contain    ${out}    ns2_veth3_linux

    ${out}=    Execute In Container    agent_vpp_1    ip netns exec ns3 ip a
    Log    ${out}
    Should Contain    ${out}    ns3_veth3_linux

Check linux Routes On VPP1 After Resync
    Check Linux Routes    node=agent_vpp_1    namespace=ns1    ip=192.168.22.2
    Check Linux Routes    node=agent_vpp_1    namespace=ns2    ip=192.168.22.1
    Check Linux Routes Gateway    node=agent_vpp_1    namespace=ns1    ip=8.8.8.8    next_hop=192.168.22.1
    Check Linux Routes Gateway    node=agent_vpp_1    namespace=ns1    ip=9.9.9.9    next_hop=192.168.22.1
    Check Linux Routes    node=agent_vpp_1    namespace=ns3    ip=192.168.22.1
    Check Linux Routes    node=agent_vpp_1    namespace=ns3    ip=192.168.22.2
    Check Linux Routes    node=agent_vpp_1    namespace=ns3    ip=192.169.22.22
    Check Linux Routes    node=agent_vpp_1    namespace=ns2    ip=192.169.22.3


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
    Should Contain    ${out}    ${ip} via

Check Linux Routes Gateway
    [Arguments]    ${node}    ${namespace}    ${ip}    ${next_hop}=${EMPTY}
    Log Many    ${node}    ${namespace}
    ${out}=    Execute In Container    ${node}    ip netns exec ${namespace} ip route show
    Log    ${out}
    Should Contain    ${out}    ${ip} via ${next_hop}

Check Linux Routes Metric
    [Arguments]    ${node}    ${namespace}    ${ip}    ${metric}
    Log Many    ${node}    ${namespace}
    ${out}=    Execute In Container    ${node}    ip netns exec ${namespace} ip route show
    Log    ${out}
    Should Match Regexp    ${out}    ${ip} via.*metric ${metric}\\s

Check Removed Linux Route
    [Arguments]    ${node}    ${namespace}    ${ip}
    Log Many    ${node}    ${namespace}
    ${out}=    Execute In Container    ${node}    ip netns exec ${namespace} ip route show
    Log    ${out}
    Should Not Contain    ${out}    ${ip} via

Ping in namespace
    [Arguments]    ${node}    ${namespace}    ${ip}
    ${out}=    Execute In Container    ${node}    ip netns exec ${namespace} ping -c 5 ${ip}
    Log    ${out}
    Should Contain     ${out}    from ${ip}
    Should Not Contain    ${out}    100% packet loss

*** Keywords ***
TestSetup
    Make Datastore Snapshots    ${TEST_NAME}_test_setup

TestTeardown
    Make Datastore Snapshots    ${TEST_NAME}_test_teardown
