[Documentation] Keywords for ssh sessions

*** Settings ***
#Library       String
#Library       RequestsLibrary
Library       SSHLibrary            timeout=60 seconds

*** Keywords ***
Execute On Machine     [Arguments]              ${machine}               ${command}               ${log}=true
                       [Documentation]          *Execute On Machine ${machine} ${command}*
                       ...                      Executing ${command} on connection with name ${machine}
                       ...                      Output log is added to machine output log
                       Log Many                 ${machine}               ${command}               ${log}
                       Switch Connection        ${machine}
                       ${out}   ${stderr}=      Execute Command          ${command}    return_stderr=True
                       Log Many                 ${out}                   ${stderr}
                       ${status}=               Run Keyword And Return Status    Should Be Empty    ${stderr}
                       Run Keyword If           ${status}==False         Log     One or more error occured during execution of a command ${command} on ${machine}    level=WARN
                       Run Keyword If           '${log}'=='true'         Append To File    ${RESULTS_FOLDER}/output_${machine}.log    *** Command: ${command}${\n}${out}${\n}*** Error: ${stderr}${\n}
                       [Return]                 ${out}

Write To Machine       [Arguments]              ${machine}               ${command}               ${delay}=${SSH_READ_DELAY}s
                       [Documentation]          *Write Machine ${machine} ${command}*
                       ...                      Writing ${command} to connection with name ${machine}
                       ...                      Output log is added to machine output log
                       Log Many                 ${machine}               ${command}               ${delay}
                       Switch Connection        ${machine}
                       Write                    ${command}
                       ${out}=                  Read                     delay=${delay}
                       Log                      ${out}
                       Append To File           ${RESULTS_FOLDER}/output_${machine}.log    *** Command: ${command}${\n}${out}${\n}
                       [Return]                 ${out}

Write To Machine Until Prompt
                       [Arguments]              ${machine}    ${command}    ${prompt}=root@    ${delay}=${SSH_READ_DELAY}
                       [Documentation]          *Write Machine ${machine} ${command}*
                       ...                      Writing ${command} to connection with name ${machine} and reading until prompt
                       ...                      Output log is added to machine output log
                       Log                      Use 'Write To Container Until Prompt' instead of this kw    level=WARN
                       Log Many                 ${machine}    ${command}    ${prompt}    ${delay}
                       Switch Connection        ${machine}
                       Write                    ${command}
                       ${out}=                  Read Until               ${prompt}${${machine}_HOSTNAME}
                       Log                      ${out}
                       ${out2}=                 Read                     delay=${delay}
                       Log                      ${out2}
                       Append To File           ${RESULTS_FOLDER}/output_${machine}.log    *** Command: ${command}${\n}${out}${out2}${\n}
                       [Return]                 ${out}${out2}

Write To Machine Until String
                       [Arguments]              ${machine}    ${command}    ${string}    ${delay}=${SSH_READ_DELAY}
                       [Documentation]          *Write Machine ${machine} ${command}*
                       ...                      Writing ${command} to connection with name ${machine} and reading until specified string
                       ...                      Output log is added to machine output log
                       Log Many                 ${machine}    ${command}    ${string}     ${delay}
                       Switch Connection        ${machine}
                       Write                    ${command}
                       ${out}=                  Read Until               ${string}
                       Log                      ${out}
                       ${out2}=                 Read                     delay=${delay}
                       Log                      ${out2}
                       Append To File           ${RESULTS_FOLDER}/output_${machine}.log    *** Command: ${command}${\n}${out}${out2}${\n}
                       [Return]                 ${out}${out2}

