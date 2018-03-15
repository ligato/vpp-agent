*** Settings ***
Resource                          common_variables.robot

*** Variables ***
${DOCKER_HOST_IP}                 localhost
${DOCKER_HOST_USER}               jenkins_ccmts
${DOCKER_HOST_PSWD}               ccmts_jenkins
${DOCKER_COMMAND}                 docker

${AGENT_VPP_IMAGE_NAME}           ligato/vpp-agent:pantheon-dev

${vpp1_DOCKER_IMAGE}              ${AGENT_VPP_IMAGE_NAME}