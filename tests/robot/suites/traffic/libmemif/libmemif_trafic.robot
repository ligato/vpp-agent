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
${SYNC_SLEEP}=       10s
${RESYNC_SLEEP}=       1s
# wait for resync vpps after restart
${RESYNC_WAIT}=        30s

*** Test Cases ***
Configure Environment
    [Tags]    setup
    Configure Environment 3

Show Interfaces Before Setup
    vpp_term: Show Interfaces    agent_vpp_1
    Sleep    ${SYNC_SLEEP}

Add Memif Interface
    vpp_ctl: Put Memif Interface With IP    node=agent_vpp_1    name=vpp1_memif1    mac=62:61:61:61:61:61    master=true    id=1    ip=192.168.1.1    prefix=24    socket=default.sock

Check Memif Interface Created
    vpp_term: Interface Is Created    node=agent_vpp_1    mac=62:61:61:61:61:61
    vat_term: Check Memif Interface State     agent_vpp_1  vpp1_memif1  mac=62:61:61:61:61:61  role=master  id=1  ipv4=192.168.1.1/24  connected=0  enabled=1  socket=${AGENT_LIBMEMIF_1_MEMIF_SOCKET_FOLDER}/default.sock

Create Memif On Agent Libmemif 1
    lmterm: Issue Command    agent_libmemif_1   conn 0 0

Show Status On Agent LibMemif 1
    lmterm: Issue Command    agent_libmemif_1    show

*** Keywords ***
TestSetup
    Make Datastore Snapshots    ${TEST_NAME}_test_setup

TestTeardown
    Make Datastore Snapshots    ${TEST_NAME}_test_teardown