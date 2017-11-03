*** Settings ***

Library     OperatingSystem
Library     String
#Library     RequestsLibrary

Resource     ../../../variables/${VARIABLES}_variables.robot
Resource    ../../../libraries/all_libs.robot
Resource    ../../../libraries/pretty_keywords.robot

Suite Setup       Run Keywords    Discard old results

*** Variables ***
${VARIABLES}=          common
${ENV}=                common

*** Test Cases ***
# Default VRF table ...
Start Three Agents And Then Configure
    [Setup]         Test Setup
    [Teardown]      Test Teardown
    Add Agent VPP Node    agent_vpp_1
    Add Agent VPP Node    agent_vpp_2
    Add Agent VPP Node    agent_vpp_3
    #setup one side with agent2
    Create loopback interface bvi_loop0 on agent_vpp_1 with ip 10.1.1.1/24 and mac 8a:f1:be:90:00:00
    Create Master memif0 on agent_vpp_1 with MAC 02:f1:be:90:00:00, key 1 and m0.sock socket
    Create bridge domain bd1 With Autolearn on agent_vpp_1 with interfaces bvi_loop0, memif0
    #setup second side with agent3
    Create loopback interface bvi_loop1 on agent_vpp_1 with ip 20.1.1.1/24 and mac 8a:f1:be:90:02:00
    Create Master memif1 on agent_vpp_1 with MAC 02:f1:be:90:02:00, key 2 and m1.sock socket
    Create bridge domain bd2 With Autolearn on agent_vpp_1 with interfaces bvi_loop1, memif1
    # prepare second agent
    Create loopback interface bvi_loop0 on agent_vpp_2 with ip 10.1.1.2/24 and mac 8a:f1:be:90:00:02
    Create Slave memif0 on agent_vpp_2 with MAC 02:f1:be:90:00:02, key 1 and m0.sock socket
    Create bridge domain bd1 With Autolearn on agent_vpp_2 with interfaces bvi_loop0, memif0
    # prepare third agent
    Create loopback interface bvi_loop0 on agent_vpp_3 with ip 20.1.1.2/24 and mac 8a:f1:be:90:00:03
    Create Slave memif0 on agent_vpp_3 with MAC 02:f1:be:90:00:03, key 2 and m1.sock socket
    Create bridge domain bd1 With Autolearn on agent_vpp_3 with interfaces bvi_loop0, memif0
    # setup routes
    Create Route On agent_vpp_2 With IP 20.1.1.0/24 With Next Hop 10.1.1.1 And Vrf Id 0
    Create Route On agent_vpp_3 With IP 10.1.1.0/24 With Next Hop 20.1.1.1 And Vrf Id 0
    # try ping
    Ping From agent_vpp_1 To 10.1.1.2
    Ping From agent_vpp_1 To 20.1.1.2
    Ping From agent_vpp_2 To 20.1.1.2
    Ping From agent_vpp_3 To 10.1.1.2

First Configure Three Agents And Then Start Agents
    [Setup]         Test Setup
    [Teardown]      Test Teardown
    #prepare first agent
    Create loopback interface bvi_loop0 on agent_vpp_1 with ip 10.1.1.1/24 and mac 8a:f1:be:90:00:00
    Create Master memif0 on agent_vpp_1 with MAC 02:f1:be:90:00:00, key 1 and m0.sock socket
    Create loopback interface bvi_loop1 on agent_vpp_1 with ip 20.1.1.1/24 and mac 8a:f1:be:90:02:00
    Create Master memif1 on agent_vpp_1 with MAC 02:f1:be:90:02:00, key 2 and m1.sock socket
    Create bridge domain bd1 With Autolearn on agent_vpp_1 with interfaces bvi_loop0, memif0
    Create bridge domain bd2 With Autolearn on agent_vpp_1 with interfaces bvi_loop1, memif1
    #prepare second agent
    Create loopback interface bvi_loop0 on agent_vpp_2 with ip 10.1.1.2/24 and mac 8a:f1:be:90:00:02
    Create Slave memif0 on agent_vpp_2 with MAC 02:f1:be:90:00:02, key 1 and m0.sock socket
    Create bridge domain bd1 With Autolearn on agent_vpp_2 with interfaces bvi_loop0, memif0
    #prepare third agent
    Create loopback interface bvi_loop0 on agent_vpp_3 with ip 20.1.1.2/24 and mac 8a:f1:be:90:00:03
    Create Slave memif0 on agent_vpp_3 with MAC 02:f1:be:90:00:03, key 2 and m1.sock socket
    Create bridge domain bd1 With Autolearn on agent_vpp_3 with interfaces bvi_loop0, memif0
    #setup routes
    Create Route On agent_vpp_2 With IP 20.1.1.0/24 With Next Hop 10.1.1.1 And Vrf Id 0
    Create Route On agent_vpp_3 With IP 10.1.1.0/24 With Next Hop 20.1.1.1 And Vrf Id 0
    #start agents
    Add Agent VPP Node    agent_vpp_1
    Add Agent VPP Node    agent_vpp_2
    Add Agent VPP Node    agent_vpp_3
    #check ping
    Ping From agent_vpp_1 To 10.1.1.2
    Ping From agent_vpp_1 To 20.1.1.2
    Ping From agent_vpp_2 To 20.1.1.2
    Ping From agent_vpp_3 To 10.1.1.2

# Non default VRF table 2 used in Agent VPP Node agent_vpp_2
Start Two Agents And Then Configure
    [Setup]         Test Setup
    [Teardown]      Test Teardown
    Add Agent VPP Node    agent_vpp_1
    Add Agent VPP Node    agent_vpp_2
    Create Master memif0 on agent_vpp_1 with IP 10.1.1.1, MAC 02:f1:be:90:00:00, key 1 and m0.sock socket
    Create Slave memif0 on agent_vpp_2 with IP 10.1.1.2, MAC 02:f1:be:90:00:02, key 1 and m0.sock socket
    # to propagate configuration from etcd to vpp
    sleep    1
    # temporarily ip table add 2
    vpp_term: Issue Command    node=agent_vpp_2    command=ip table add 2
    # temporarily
    vpp_term: Issue Command    node=agent_vpp_2    command=set int ip address del memif0/1 all
    # temporarily set int ip table memif0/1 2
    vpp_term: Issue Command    node=agent_vpp_2    command=set int ip table memif0/1 2
    # temporarily
    vpp_term: Issue Command    node=agent_vpp_2    command=set int ip address memif0/1 10.1.1.2/24

    # try ping
    Ping From agent_vpp_1 To 10.1.1.2
    # this does not work for non default vrf: Ping From agent_vpp_2 To 10.1.1.1
    Ping On agent_vpp_2 With IP 10.1.1.1, Source memif0/1

# Non default VRF table 2 used in Agent VPP Node agent_vpp_2
# Non default VRF table 3 used in Agent VPP Node agent_vpp_3
Start Three Agents, Then Configure And Move Interfaces To Non Default VRF
    [Setup]         Test Setup
    #[Teardown]      Test Teardown
    Add Agent VPP Node    agent_vpp_1
    Add Agent VPP Node    agent_vpp_2
    Add Agent VPP Node    agent_vpp_3
    #setup one side with agent2
    Create loopback interface bvi_loop0 on agent_vpp_1 with ip 10.1.1.1/24 and mac 8a:f1:be:90:00:00
    Create Master memif0 on agent_vpp_1 with MAC 02:f1:be:90:00:00, key 1 and m0.sock socket
    Create bridge domain bd1 With Autolearn on agent_vpp_1 with interfaces bvi_loop0, memif0
    #setup second side with agent3
    Create loopback interface bvi_loop1 on agent_vpp_1 with ip 20.1.1.1/24 and mac 8a:f1:be:90:02:00
    Create Master memif1 on agent_vpp_1 with MAC 02:f1:be:90:02:00, key 2 and m1.sock socket
    Create bridge domain bd2 With Autolearn on agent_vpp_1 with interfaces bvi_loop1, memif1

    # prepare second agent
    Create loopback interface bvi_loop0 on agent_vpp_2 with ip 10.1.1.2/24 and mac 8a:f1:be:90:00:02
    Create Slave memif0 on agent_vpp_2 with MAC 02:f1:be:90:00:02, key 1 and m0.sock socket
    Create bridge domain bd1 With Autolearn on agent_vpp_2 with interfaces bvi_loop0, memif0
    # to propagate configuration from etcd to vpp
    sleep    1
    # temporarily
    vpp_term: Issue Command    node=agent_vpp_2    command=ip table add 2
    vpp_term: Issue Command    node=agent_vpp_2    command=set int ip address del loop0 all
    vpp_term: Issue Command    node=agent_vpp_2    command=set int ip table loop0 2
    vpp_term: Issue Command    node=agent_vpp_2    command=set int ip table memif0/1 2
    vpp_term: Issue Command    node=agent_vpp_2    command=set int ip address loop0 10.1.1.2/24
    # temp. will be replaced by: Create Vrf Table 2 On agent_vpp_2 With Interfaces memif0

    # prepare third agent
    Create loopback interface bvi_loop0 on agent_vpp_3 with ip 20.1.1.2/24 and mac 8a:f1:be:90:00:03
    Create Slave memif0 on agent_vpp_3 with MAC 02:f1:be:90:00:03, key 2 and m1.sock socket
    Create bridge domain bd1 With Autolearn on agent_vpp_3 with interfaces bvi_loop0, memif0
    # to propagate configuration from etcd to vpp
    sleep    1
    # temporarily
    vpp_term: Issue Command    node=agent_vpp_3    command=ip table add 3
    vpp_term: Issue Command    node=agent_vpp_3    command=set int ip address del loop0 all
    vpp_term: Issue Command    node=agent_vpp_3    command=set int ip table loop0 3
    vpp_term: Issue Command    node=agent_vpp_3    command=set int ip table memif0/1 3
    vpp_term: Issue Command    node=agent_vpp_3    command=set int ip address loop0 20.1.1.2/24
    # temp. will be replaced by: Create Vrf Table 3 On agent_vpp_3 With Interfaces memif0

    # setup routes
    Create Route On agent_vpp_2 With IP 20.1.1.0/24 With Next Hop 10.1.1.1 And Vrf Id 2
    Create Route On agent_vpp_3 With IP 10.1.1.0/24 With Next Hop 20.1.1.1 And Vrf Id 3

    # try ping
    Ping From agent_vpp_1 To 10.1.1.2
    Ping From agent_vpp_1 To 20.1.1.2
    #Ping From agent_vpp_2 To 20.1.1.2
    Ping On agent_vpp_2 With IP 20.1.1.2, Source loop0
    #Ping From agent_vpp_3 To 10.1.1.2
    Ping On agent_vpp_3 With IP 10.1.1.2, Source loop0
