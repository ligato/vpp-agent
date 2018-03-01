*** Settings ***
Library      OperatingSystem
#Library      RequestsLibrary
#Library      SSHLibrary      timeout=60s
#Library      String

Resource     ../../../variables/${VARIABLES}_variables.robot

Resource     ../../../libraries/all_libs.robot

Suite Setup       Testsuite Setup
Suite Teardown    Testsuite Teardown
Test Setup        TestSetup
Test Teardown     TestTeardown

*** Variables ***
${VARIABLES}=          common
${ENV}=                common
${SYNC_SLEEP}=         10s
${RESYNC_SLEEP}=       1s
${LIBMEMIF_IP1}=       192.168.1.2
${VPP2MEMIF_IP1}=      192.168.1.2
${VPP1MEMIF_IP1}=      192.168.1.1
${LIBMEMIF_IP2}=       192.168.2.2
${VPP2MEMIF_IP2}=       192.168.2.2
${VPP1MEMIF_IP2}=       192.168.2.1
# wait for resync vpps after restart
${RESYNC_WAIT}=        30s

*** Test Cases ***
Configure Environment
    [Tags]    setup
    Configure Environment 3

Show Interfaces Before Setup
    vpp_term: Show Interfaces    agent_vpp_1

Add Memif1 Interface On VPP1
    vpp_ctl: Put Memif Interface With IP    node=agent_vpp_1    name=vpp1_memif1    mac=62:61:61:61:61:61    master=false    id=0    ip=${VPP1MEMIF_IP1}    prefix=24    socket=memif.sock
    Sleep     ${SYNC_SLEEP}

Check Memif1 Interface Created On VPP1
    vpp_term: Interface Is Created    node=agent_vpp_1    mac=62:61:61:61:61:61
    vat_term: Check Memif Interface State     agent_vpp_1  vpp1_memif1  mac=62:61:61:61:61:61  role=slave  id=0  ipv4=${VPP1MEMIF_IP1}/24  connected=0  enabled=1  socket=${AGENT_LIBMEMIF_1_MEMIF_SOCKET_FOLDER}/memif.sock

Modify Memif1 Interface On VPP1
    vpp_ctl: Put Memif Interface With IP    node=agent_vpp_1    name=vpp1_memif1    mac=62:61:61:61:61:62    master=false    id=0    ip=${VPP1MEMIF_IP2}    prefix=24    socket=memif.sock
    Sleep     ${SYNC_SLEEP}

Check Memif1 Interface On VPP1 is Modified
    vpp_term: Interface Is Created    node=agent_vpp_1    mac=62:61:61:61:61:62
    vat_term: Check Memif Interface State     agent_vpp_1  vpp1_memif1  mac=62:61:61:61:61:62  role=slave  id=0  ipv4=${VPP1MEMIF_IP2}/24  connected=0  enabled=1  socket=${AGENT_LIBMEMIF_1_MEMIF_SOCKET_FOLDER}/memif.sock

Create And Chek Memif1 On Agent Libmemif 1
    ${out}=      lmterm: Issue Command    agent_libmemif_1   conn 0 1
    ${out}=      lmterm: Issue Command    agent_libmemif_1    show
    Should Contain     ${out}     interface ip: ${LIBMEMIF_IP1}
    Should Contain     ${out}     link: up

Check Memif1 Interface On VPP1 Connected To LibMemif
    vpp_term: Interface Is Created    node=agent_vpp_1    mac=62:61:61:61:61:62
    vat_term: Check Memif Interface State     agent_vpp_1  vpp1_memif1  mac=62:61:61:61:61:62  role=slave  id=0  ipv4=${VPP1MEMIF_IP2}/24  connected=1  enabled=1  socket=${AGENT_LIBMEMIF_1_MEMIF_SOCKET_FOLDER}/memif.sock

Modify Memif1 On VPP1 back
    vpp_ctl: Put Memif Interface With IP    node=agent_vpp_1    name=vpp1_memif1    mac=62:61:61:61:61:61    master=false    id=0    ip=${VPP1MEMIF_IP1}    prefix=24    socket=memif.sock
    Sleep     ${SYNC_SLEEP}

Check Memif1 on Vpp1 is connected
    vat_term: Check Memif Interface State     agent_vpp_1  vpp1_memif1  mac=62:61:61:61:61:61  role=slave  id=0  ipv4=${VPP1MEMIF_IP1}/24  connected=1  enabled=1  socket=${AGENT_LIBMEMIF_1_MEMIF_SOCKET_FOLDER}/memif.sock
    Sleep     ${SYNC_SLEEP}

Check Ping VPP1 -> Agent Libmemif 1
    vpp_term: Check Ping    agent_vpp_1    ${LIBMEMIF_IP1}


Remove VPP Nodes
    Remove All VPP Nodes
    Sleep    ${SYNC_SLEEP}
    Add Agent VPP Node    agent_vpp_1
    #Add Agent VPP Node    agent_vpp_2
    Sleep    ${RESYNC_WAIT}

Check Memif1 Interface On VPP1 Connected To LibMemif After Resync
    vpp_term: Interface Is Created    node=agent_vpp_1    mac=62:61:61:61:61:61
    vat_term: Check Memif Interface State     agent_vpp_1  vpp1_memif1  mac=62:61:61:61:61:61  role=slave  id=0  ipv4=${VPP1MEMIF_IP1}/24  connected=1  enabled=1  socket=${AGENT_LIBMEMIF_1_MEMIF_SOCKET_FOLDER}/memif.sock

Check Ping VPP1 -> Agent Libmemif 1 After Resync
    vpp_term: Check Ping    agent_vpp_1    ${LIBMEMIF_IP1}

##############################################################################


Delete Memif On Agent Libmemif 1
    ${out}=      lmterm: Issue Command    agent_libmemif_1   del 0
    Sleep     ${SYNC_SLEEP}

Check Memif1 Interface On VPP1 Disconnected After Master Deleted
    vpp_term: Interface Is Created    node=agent_vpp_1    mac=62:61:61:61:61:61
    vat_term: Check Memif Interface State     agent_vpp_1  vpp1_memif1  mac=62:61:61:61:61:61  role=slave  id=0  ipv4=${VPP1MEMIF_IP1}/24  connected=0  enabled=1  socket=${AGENT_LIBMEMIF_1_MEMIF_SOCKET_FOLDER}/memif.sock

Create Memif1 On Agent Libmemif 1 Again
    ${out}=      lmterm: Issue Command    agent_libmemif_1   conn 0 1
    Sleep     ${SYNC_SLEEP}

Check Memif1 Interface On VPP1 Connected After Master Deleted and Created
    vpp_term: Interface Is Created    node=agent_vpp_1    mac=62:61:61:61:61:61
    vat_term: Check Memif Interface State     agent_vpp_1  vpp1_memif1  mac=62:61:61:61:61:61  role=slave  id=0  ipv4=${VPP1MEMIF_IP1}/24  connected=1  enabled=1  socket=${AGENT_LIBMEMIF_1_MEMIF_SOCKET_FOLDER}/memif.sock

Check Ping VPP1 -> Agent Libmemif 1 After Delete and Create
    vpp_term: Check Ping    agent_vpp_1    ${LIBMEMIF_IP1}
    Sleep    850s

###### Here VPP crashes
Modify Memif1 Interface On VPP1 While Connected
    vpp_ctl: Put Memif Interface With IP    node=agent_vpp_1    name=vpp1_memif1    mac=62:61:61:61:61:62    master=false    id=0    ip=${VPP1MEMIF_IP2}    prefix=24    socket=memif.sock
    Sleep     ${SYNC_SLEEP}

Check Memif1 Interface On VPP1 Modified
    vpp_term: Interface Is Created    node=agent_vpp_1    mac=62:61:61:61:61:62
    vat_term: Check Memif Interface State     agent_vpp_1  vpp1_memif1  mac=62:61:61:61:61:62  role=slave  id=0  ipv4=${VPP1MEMIF_IP2}/24  connected=1  enabled=1  socket=${AGENT_LIBMEMIF_1_MEMIF_SOCKET_FOLDER}/memif.sock

Final Sleep
    Sleep    250s
###########################################################


*** Keywords ***
TestSetup
    Make Datastore Snapshots    ${TEST_NAME}_test_setup

TestTeardown
    Make Datastore Snapshots    ${TEST_NAME}_test_teardown