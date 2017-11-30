*** Settings ***
Library      OperatingSystem
#Library      RequestsLibrary
#Library      SSHLibrary      timeout=60s
#Library      String

Resource     ../../../variables/${VARIABLES}_variables.robot
Resource     ../../../libraries/all_libs.robot
Resource    ../../../libraries/pretty_keywords.robot

Suite Setup       Testsuite Setup
Suite Teardown    Testsuite Teardown

*** Variables ***
${VARIABLES}=               common
${ENV}=                     common
${SYNC_SLEEP}=         10s
# wait for resync vpps after restart
${RESYNC_WAIT}=        50s

*** Test Cases ***
Configure Environment
    [Tags]    setup
    Configure Environment 1