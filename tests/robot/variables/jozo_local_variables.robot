*** Settings ***
Resource                          common_variables.robot

*** Variables ***
${DOCKER_HOST_IP}                 192.168.200.11
${DOCKER_HOST_USER}               robot
${DOCKER_HOST_PSWD}               robot

${AGENT_VPP_IMAGE_NAME}           containers.cisco.com/odpm_jenkins_gen/prod_vpp_agent:pt-dev

${vpp1_DOCKER_IMAGE}              ${AGENT_VPP_IMAGE_NAME}
${vpp1_VPP_PORT}                  5002
${vpp1_VPP_HOST_PORT}             5004
${vpp1_SOCKET_FOLDER}             /tmp
${vpp1_VPP_TERM_PROMPT}           vpp#
${vpp1_VPP_VAT_PROMPT}            vat#
