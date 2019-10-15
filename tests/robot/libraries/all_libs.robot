*** Settings ***
Documentation     Library which includes all other libs
...
...
...

Library     basic_operations.py
Resource    setup-teardown.robot
Resource    ssh.robot
Resource    docker.robot
Resource    configurations.robot
Resource    vat_term.robot
Resource    vpp_term.robot
Resource    lm_term.robot
Resource    vpp.robot
Resource    etcdctl.robot
Resource    rest_api.robot
Resource    linux.robot
Resource    vpp_api.robot
# Resource    kubernetes/all_kube_libs.robot
Resource    SshCommons.robot
