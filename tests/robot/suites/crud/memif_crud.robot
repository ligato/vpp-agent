*** Settings ***
Library      OperatingSystem
#Library      RequestsLibrary
#Library      SSHLibrary      timeout=60s
#Library      String

Resource     ../../variables/${VARIABLES}_variables.robot

Resource     ../../libraries/all_libs.robot

Suite Setup       Testsuite Setup
Suite Teardown    Testsuite Teardown
Test Setup        TestSetup
Test Teardown     TestTeardown

*** Variables ***
${VARIABLES}=          common
${ENV}=                common
${MEMIF11_MAC}=          1a:00:00:11:11:11
${MEMIF11_SEC_MAC}=      1a:00:00:11:11:12
${MEMIF21_MAC}=          2a:00:00:22:22:22
${MEMIF21_SEC_MAC}=      2a:00:00:22:22:23
${MEMIF12_MAC}=          3a:00:00:33:33:33
${MEMIF22_MAC}=          4a:00:00:44:44:44

*** Test Cases ***
Configure Environment
    [Tags]    setup
    Configure Environment 1

Show Interfaces Before Setup
    vpp_term: Show Interfaces    agent_vpp_1

Add VPP1_memif1 Interface
    vpp_term: Interface Not Exists    node=agent_vpp_1    mac=${MEMIF11_MAC}
    vpp_ctl: Put Memif Interface With IP    node=agent_vpp_1    name=vpp1_memif1    mac=${MEMIF11_MAC}    master=true    id=1    ip=192.168.1.1    prefix=24    socket=default.sock

Check That VPP1_memif1 Is Created But Not Connected
    vpp_term: Interface Is Created    node=agent_vpp_1    mac=${MEMIF11_MAC}
    vat_term: Check Memif Interface State     agent_vpp_1  vpp1_memif1  mac=${MEMIF11_MAC}  role=master  id=1  ipv4=192.168.1.1/24  connected=0  enabled=1  socket=${AGENT_VPP_1_SOCKET_FOLDER}/default.sock

Add VPP2_memif1 Interface
    vpp_term: Interface Not Exists    node=agent_vpp_2    mac=${MEMIF21_MAC}
    vpp_ctl: Put Memif Interface With IP    node=agent_vpp_2    name=vpp2_memif1    mac=${MEMIF21_MAC}    master=false    id=1    ip=192.168.1.2    prefix=28    socket=default.sock

Check That VPP2_memif1 Is Created And Connected With VPP1_memif1
    vpp_term: Interface Is Created    node=agent_vpp_2    mac=${MEMIF21_MAC}
    vat_term: Check Memif Interface State     agent_vpp_2  vpp2_memif1  mac=${MEMIF21_MAC}  role=slave  id=1  ipv4=192.168.1.2/28  connected=1  enabled=1  socket=${AGENT_VPP_2_SOCKET_FOLDER}/default.sock
    vat_term: Check Memif Interface State     agent_vpp_1  vpp1_memif1  mac=${MEMIF11_MAC}  role=master  id=1  ipv4=192.168.1.1/24  connected=1  enabled=1  socket=${AGENT_VPP_1_SOCKET_FOLDER}/default.sock

Add VPP1_memif2 Interface
    vpp_term: Interface Not Exists    node=agent_vpp_1    mac=${MEMIF12_MAC}
    vpp_ctl: Put Memif Interface With IP    node=agent_vpp_1    name=vpp1_memif2    mac=${MEMIF12_MAC}    master=true    id=2    ip=192.168.2.1    prefix=26    socket=default.sock

Check That VPP1_memif2 Is Created But Not Connected
    vpp_term: Interface Is Created    node=agent_vpp_1    mac=${MEMIF12_MAC}
    vat_term: Check Memif Interface State     agent_vpp_1  vpp1_memif2  mac=${MEMIF12_MAC}  role=master  id=2  ipv4=192.168.2.1/26  connected=0  enabled=1  socket=${AGENT_VPP_1_SOCKET_FOLDER}/default.sock

Add VPP2_memif2 Interface
    vpp_term: Interface Not Exists    node=agent_vpp_2    mac=${MEMIF22_MAC}
    vpp_ctl: Put Memif Interface With IP    node=agent_vpp_2    name=vpp2_memif2    mac=${MEMIF22_MAC}    master=false    id=2    ip=192.168.2.2    prefix=28    socket=default.sock

Check That VPP2_memif2 Is Created And Connected With VPP1_memif2
    vpp_term: Interface Is Created    node=agent_vpp_2    mac=${MEMIF22_MAC}
    vat_term: Check Memif Interface State     agent_vpp_2  vpp2_memif2  mac=${MEMIF22_MAC}  role=slave  id=2  ipv4=192.168.2.2/28  connected=1  enabled=1  socket=${AGENT_VPP_2_SOCKET_FOLDER}/default.sock
    vat_term: Check Memif Interface State     agent_vpp_1  vpp1_memif2  mac=${MEMIF12_MAC}  role=master  id=2  ipv4=192.168.2.1/26  connected=1  enabled=1  socket=${AGENT_VPP_1_SOCKET_FOLDER}/default.sock

Check That VPP1_memif1 And VPP2_memif1 Interfaces Are Not Affected By VPP1_memif2 And VPP2_memif2 Interfaces
    vat_term: Check Memif Interface State     agent_vpp_1  vpp1_memif1  mac=${MEMIF11_MAC}  role=master  id=1  ipv4=192.168.1.1/24  connected=1  enabled=1  socket=${AGENT_VPP_1_SOCKET_FOLDER}/default.sock
    vat_term: Check Memif Interface State     agent_vpp_2  vpp2_memif1  mac=${MEMIF21_MAC}  role=slave  id=1  ipv4=192.168.1.2/28  connected=1  enabled=1  socket=${AGENT_VPP_2_SOCKET_FOLDER}/default.sock

Update VPP1_memif1 Interface
    vpp_ctl: Put Memif Interface With IP    node=agent_vpp_1    name=vpp1_memif1    mac=${MEMIF11_SEC_MAC}    master=true    id=1    ip=192.168.10.1    prefix=30    socket=default.sock
    vpp_term: Interface Is Deleted    node=agent_vpp_1    mac=${MEMIF11_MAC}
    vpp_term: Interface Is Created    node=agent_vpp_1    mac=${MEMIF11_SEC_MAC}
    vat_term: Check Memif Interface State     agent_vpp_1  vpp1_memif1  mac=${MEMIF11_SEC_MAC}  role=master  id=1  ipv4=192.168.10.1/30  connected=1  enabled=1  socket=${AGENT_VPP_1_SOCKET_FOLDER}/default.sock

Check That VPP2_memif1 Is Still Configured And Connected
    vat_term: Check Memif Interface State     agent_vpp_2  vpp2_memif1  mac=${MEMIF21_MAC}  role=slave  id=1  ipv4=192.168.1.2/28  connected=1  enabled=1  socket=${AGENT_VPP_2_SOCKET_FOLDER}/default.sock

Check That VPP1_memif2 And VPP2_memif2 Are Not Affected By VPP1_memif1 Update
    vat_term: Check Memif Interface State     agent_vpp_2  vpp2_memif2  mac=${MEMIF22_MAC}  role=slave  id=2  ipv4=192.168.2.2/28  connected=1  enabled=1  socket=${AGENT_VPP_2_SOCKET_FOLDER}/default.sock
    vat_term: Check Memif Interface State     agent_vpp_1  vpp1_memif2  mac=${MEMIF12_MAC}  role=master  id=2  ipv4=192.168.2.1/26  connected=1  enabled=1  socket=${AGENT_VPP_1_SOCKET_FOLDER}/default.sock

Update VPP2_memif1 Interface
    vpp_ctl: Put Memif Interface With IP    node=agent_vpp_2    name=vpp2_memif1    mac=${MEMIF21_SEC_MAC}    master=false    id=1    ip=192.168.10.2    prefix=24    socket=default.sock
    vpp_term: Interface Is Deleted    node=agent_vpp_2    mac=${MEMIF21_MAC}
    vpp_term: Interface Is Created    node=agent_vpp_2    mac=${MEMIF21_SEC_MAC}
    vat_term: Check Memif Interface State     agent_vpp_2  vpp2_memif1  mac=${MEMIF21_SEC_MAC}  role=slave  id=1  ipv4=192.168.10.2/24  connected=1  enabled=1  socket=${AGENT_VPP_2_SOCKET_FOLDER}/default.sock

Check That VPP1_memif1 Is Still Configured And Connected
    vat_term: Check Memif Interface State     agent_vpp_1  vpp1_memif1  mac=${MEMIF11_SEC_MAC}  role=master  id=1  ipv4=192.168.10.1/30  connected=1  enabled=1  socket=${AGENT_VPP_1_SOCKET_FOLDER}/default.sock

Check That VPP1_memif2 And VPP2_memif2 Are Not Affected By VPP2_memif1 Update
    vat_term: Check Memif Interface State     agent_vpp_2  vpp2_memif2  mac=${MEMIF22_MAC}  role=slave  id=2  ipv4=192.168.2.2/28  connected=1  enabled=1  socket=${AGENT_VPP_2_SOCKET_FOLDER}/default.sock
    vat_term: Check Memif Interface State     agent_vpp_1  vpp1_memif2  mac=${MEMIF12_MAC}  role=master  id=2  ipv4=192.168.2.1/26  connected=1  enabled=1  socket=${AGENT_VPP_1_SOCKET_FOLDER}/default.sock

Delete VPP1_memif2 Interface
    vpp_ctl: Delete VPP Interface    node=agent_vpp_1    name=vpp1_memif2
    vpp_term: Interface Is Deleted    node=agent_vpp_1    mac=${MEMIF12_MAC}

Check That VPP2_memif2 Interface Is Disconnected
    vat_term: Check Memif Interface State     agent_vpp_2  vpp2_memif2  mac=${MEMIF22_MAC}  role=slave  id=2  ipv4=192.168.2.2/28  connected=0  enabled=1  socket=${AGENT_VPP_2_SOCKET_FOLDER}/default.sock

Check That VPP1_memif1 And VPP2_memif1 Are Not Affected By VPP1_memif2 Delete
    vat_term: Check Memif Interface State     agent_vpp_1  vpp1_memif1  mac=${MEMIF11_SEC_MAC}  role=master  id=1  ipv4=192.168.10.1/30  connected=1  enabled=1  socket=${AGENT_VPP_1_SOCKET_FOLDER}/default.sock
    vat_term: Check Memif Interface State     agent_vpp_2  vpp2_memif1  mac=${MEMIF21_SEC_MAC}  role=slave  id=1  ipv4=192.168.10.2/24  connected=1  enabled=1  socket=${AGENT_VPP_2_SOCKET_FOLDER}/default.sock

Delete VPP2_memif2 Interface
    vpp_ctl: Delete VPP Interface    node=agent_vpp_2    name=vpp2_memif2
    vpp_term: Interface Is Deleted    node=agent_vpp_2    mac=${MEMIF22_MAC}

Check That VPP1_memif1 And VPP2_memif1 Are Not Affected By VPP2_memif2 Delete
    vat_term: Check Memif Interface State     agent_vpp_1  vpp1_memif1  mac=${MEMIF11_SEC_MAC}  role=master  id=1  ipv4=192.168.10.1/30  connected=1  enabled=1  socket=${AGENT_VPP_1_SOCKET_FOLDER}/default.sock
    vat_term: Check Memif Interface State     agent_vpp_2  vpp2_memif1  mac=${MEMIF21_SEC_MAC}  role=slave  id=1  ipv4=192.168.10.2/24  connected=1  enabled=1  socket=${AGENT_VPP_2_SOCKET_FOLDER}/default.sock

Show Interfaces And Other Objects After Setup
    vpp_term: Show Interfaces    agent_vpp_1
    vpp_term: Show Interfaces    agent_vpp_2
    Write To Machine    agent_vpp_1_term    show int addr
    Write To Machine    agent_vpp_2_term    show int addr
    Write To Machine    agent_vpp_1_term    show h
    Write To Machine    agent_vpp_2_term    show h
    Write To Machine    agent_vpp_1_term    show memif
    Write To Machine    agent_vpp_2_term    show memif
    Write To Machine    agent_vpp_1_term    show err
    Write To Machine    agent_vpp_2_term    show err
    vat_term: Interfaces Dump    agent_vpp_1
    vat_term: Interfaces Dump    agent_vpp_2
    Write To Machine    vpp_agent_ctl    vpp-agent-ctl ${AGENT_VPP_ETCD_CONF_PATH} -ps
    Execute In Container    agent_vpp_1    ip a
    Execute In Container    agent_vpp_2    ip a

*** Keywords ***
TestSetup
    Make Datastore Snapshots    ${TEST_NAME}_test_setup

TestTeardown
    Make Datastore Snapshots    ${TEST_NAME}_test_teardown

