[Documentation]     Keywords for working with interfaces using VPP API

*** Settings ***
Library     interface_generic.py

Resource    ../vpp_api.robot

*** Variables ***
${terminal_timeout}=      30s
${bd_timeout}=            15s

*** Keywords ***

vpp_api: Interfaces Dump
    [Arguments]        ${node}
    ${out}=            Execute API Command    ${node}    sw_interface_dump
    [Return]           ${out}

vpp_api: Get Interface Index
    [Arguments]        ${node}    ${name}
    [Documentation]    Return interface index with specified name
    ${out}=            vpp_api: Interfaces Dump    ${node}
    ${index}=          Get Interface Index From API    ${out[0]["api_reply"]}    ${name}
    [Return]           ${index}

vpp_api: Get Interface Name
    [Arguments]        ${node}    ${index}
    [Documentation]    Return interface index with specified name
    ${out}=            vpp_api: Interfaces Dump    ${node}
    ${index}=          Get Interface Name From API    ${out[0]["api_reply"]}    ${index}
    [Return]           ${index}

vpp_api: Get Interface State By Name
    [Arguments]        ${node}    ${name}
    ${out}=            vpp_api: Interfaces Dump    ${node}
    ${state}=          Get Interface State From API    ${out[0]["api_reply"]}    name=${name}
    [Return]           ${state}

vpp_api: Get Interface State By Index
    [Arguments]        ${node}    ${index}
    ${out}=            vpp_api: Interfaces Dump    ${node}
    ${state}=          Get Interface State From API    ${out[0]["api_reply"]}    index=${index}
    [Return]           ${state}
