[Documentation]     Restconf api specific configurations

*** Settings ***
Library        rest_api.py

*** Keywords ***

rest_api: Get
    [Arguments]      ${node}    ${uri}    ${expected_code}=200
    ${response}=      Get Request          ${node}    ${uri}
    Log Many          ${response.text}     ${response.status_code}
#    ${pretty}=        Ordered Json         ${response.text}
#    log               ${pretty}
    Sleep             ${REST_CALL_SLEEP}
    Run Keyword If    '${expected_code}'!='0'       Should Be Equal As Integers    ${response.status_code}    ${expected_code}
    [Return]         ${response.text}


rest_api: Put
    [Arguments]      ${node}    ${uri}    ${expected_code}=200
    Log Many          ${node}    ${uri}
    ${response}=      Put Request          ${node}    ${uri}
    Log Many          ${response.text}     ${response.status_code}
    ${pretty}=        Ordered Json         ${response.text}
    log               ${pretty}
    Sleep             ${REST_CALL_SLEEP}
    Run Keyword If    '${expected_code}'!='0'       Should Be Equal As Integers    ${response.status_code}    ${expected_code}
    [Return]         ${response.text}

rest_api: Get Loggers List
    [Arguments]      ${node}
    Log Many          ${node}
    ${uri}=           Set Variable     log/list
    Log               ${uri}
    ${out}=           rest_api: Get    ${node}    ${uri}
    [Return]         ${out}

rest_api: Change Logger Level
    [Arguments]     ${node}    ${logger}    ${log_level}
    Log Many         ${node}    ${logger}    ${log_level}
    ${uri}=          Set variable      /log/${logger}/${log_level}
    Log              ${uri}
    ${out}=          rest_api: Put     ${node}    ${uri}
    [Return]        ${out}
