[Documentation]     ENV specific configurations

*** Settings ***

*** Keywords ***

Configure Environment 1
    Add Agent VPP Node    agent_vpp_1
    Add Agent VPP Node    agent_vpp_2
    Execute In Container    agent_vpp_1    echo $MICROSERVICE_LABEL
    Execute In Container    agent_vpp_1    ls -al
    Execute On Machine    docker    ${DOCKER_COMMAND} images
    Execute On Machine    docker    ${DOCKER_COMMAND} ps -as

Configure Environment 2
    Add Agent VPP Node    agent_vpp_1       vswitch=${TRUE}
    Add Agent Node        agent_1
    Add Agent Node        agent_2
    Execute In Container    agent_vpp_1    echo $MICROSERVICE_LABEL
    Execute In Container    agent_vpp_1    ls -al
    Execute On Machine    docker    ${DOCKER_COMMAND} images
    Execute On Machine    docker    ${DOCKER_COMMAND} ps -as
    Start SFC Controller Container With Own Config    acl_basic.conf
    Sleep    ${SYNC_SLEEP}
