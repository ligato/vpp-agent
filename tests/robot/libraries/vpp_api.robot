[Documentation]     Keywords for working with VPP API

*** Settings ***
Library      Collections
Library      vpp_api.py

*** Variables ***
${interface_timeout}=     15s
${terminal_timeout}=      30s

*** Keywords ***
Execute API Command
    [Arguments]    ${node}    ${command}    &{arguments}
    ${out}=    execute api    ${DOCKER_HOST_IP}    ${DOCKER_HOST_USER}    ${DOCKER_HOST_PSWD}    ${node}    ${command}    &{arguments}
    [Return]    ${out}