## What contains this folder


Here are templates for jenkins-job-builder.
Template [05test_template.yaml][6] is a template for jenkins job which basically runs pybot command to run one file with robot tests e.g.
```
pybot --loglevel ${LOGLEVEL} -v AGENT_VPP_IMAGE_NAME:${IMAGE_NAME} -v AGENT_IMAGE_NAME:${IMAGE_NAME} -v DOCKER_HOST_IP:${DOCKER_HOST_IP} -v VARIABLES:${VARIABLES_FILE}    /root/vpp-agent/tests/robot/suites/crud/tap_crud.robot

```
Note: the variables enclosed in curly braces will be replaced in time of execution of jenkins job
Note: This procedure suppose that some preliminary steps are done - preparation of repository to the folder /root. This is done by other jenkins job ([vpp_agent_jobs/'01A. PREPARE REPOSITORY INSIDE JENKINS CONTAINER.yml)][4]. This is done to prevent recurrent downloading of repository for each single robot file.

Template [04pipeline_template.yaml][5] is a template which amasses the jobs to the groups to run them together.


Project ligato/vpp-agent contains [robot tests][1]

Script [script.sh][2]
* will collect all these robot test to the file `list_of_all_robot_tests2`
* sort them
* will run python script [script.py][3] which will prepare the data for templates (outputed to `p.yaml`) 

Example of p.yaml
```
- project:
    name: ligato/vpp-agent all tests on arm64
    jobs:
      - '05{inode_of_folder}_{name_of_test}_job':
          HOWTOBUILD_INCLTAGPRESENT: ' '
          HOWTOBUILD_EXCLTAGPRESENT: ' '
          inode_of_folder: API________
          date_of_jjb_generation: 2019-02-22 05:45:55
          name_of_test: bfd_api
          path_to_test: /root/vpp-agent/tests/robot/suites/api/BFD/bfd_api.robot
          local_variables_file: arm64_local
          arm64_node: 147.75.72.194
...
...
      - '05{inode_of_folder}_{name_of_test}_job':
          HOWTOBUILD_INCLTAGPRESENT: ' '
          HOWTOBUILD_EXCLTAGPRESENT: ' '
          inode_of_folder: CRUDIPV6___
          date_of_jjb_generation: 2019-02-22 05:45:55
          name_of_test: loopback_crudIPv6
          path_to_test: /root/vpp-agent/tests/robot/suites/crudIPv6/loopback_crudIPv6.robot
          local_variables_file: arm64_local
          arm64_node: 147.75.72.194
...
...
      - '04{name_of_pipeline}_pipeline':
          name_of_pipeline: IPv4_arm64_node_I
          date_of_jjb_generation: 2019-02-22 05:45:55
          local_variables_file: 'arm64_local'
          arm64_node: '147.75.72.194'
          list_of_jenkins_jobs: |-
              stage 'test'
              build job: '05CRUD________acl_crud_job', parameters: [string(name: 'HOWTOBUILD', value: "${{HOWTOBUILD}}"), string(name: 'LOGLEVEL', value: "${{LOGLEVEL}}"), string(name: 'VARIABLES_FIL
E', value: "${{VARIABLES_FILE}}"), string(name: 'DOCKER_HOST_IP', value: "${{DOCKER_HOST_IP}}"), string(name: 'IMAGE_NAME', value: "${{IMAGE_NAME}}")], propagate: false
              build job: '05CRUD________afpacket_crud_job', parameters: [string(name: 'HOWTOBUILD', value: "${{HOWTOBUILD}}"), string(name: 'LOGLEVEL', value: "${{LOGLEVEL}}"), string(name: 'VARIABLE
S_FILE', value: "${{VARIABLES_FILE}}"), string(name: 'DOCKER_HOST_IP', value: "${{DOCKER_HOST_IP}}"), string(name: 'IMAGE_NAME', value: "${{IMAGE_NAME}}")], propagate: false
              build job: '05CRUD________app_namespaces_crud_job', parameters: [string(name: 'HOWTOBUILD', value: "${{HOWTOBUILD}}"), string(name: 'LOGLEVEL', value: "${{LOGLEVEL}}"), string(name: 'VA
RIABLES_FILE', value: "${{VARIABLES_FILE}}"), string(name: 'DOCKER_HOST_IP', value: "${{DOCKER_HOST_IP}}"), string(name: 'IMAGE_NAME', value: "${{IMAGE_NAME}}")], propagate: false
              build job: '05CRUD________arp_crud_job', parameters: [string(name: 'HOWTOBUILD', value: "${{HOWTOBUILD}}"), string(name: 'LOGLEVEL', value: "${{LOGLEVEL}}"), string(name: 'VARIABLES_FIL
E', value: "${{VARIABLES_FILE}}"), string(name: 'DOCKER_HOST_IP', value: "${{DOCKER_HOST_IP}}"), string(name: 'IMAGE_NAME', value: "${{IMAGE_NAME}}")], propagate: false
              build job: '05CRUD________bd_crud_job', parameters: [string(name: 'HOWTOBUILD', value: "${{HOWTOBUILD}}"), string(name: 'LOGLEVEL', value: "${{LOGLEVEL}}"), string(name: 'VARIABLES_FILE
', value: "${{VARIABLES_FILE}}"), string(name: 'DOCKER_HOST_IP', value: "${{DOCKER_HOST_IP}}"), string(name: 'IMAGE_NAME', value: "${{IMAGE_NAME}}")], propagate: false
...
...

```

* Prepared templates and data will be merged by jenkins-job-builder to deploy jenkins jobs
* defaults.yaml is a file with setting for all jobs 


[1]: https://github.com/ligato/vpp-agent/tree/master/tests/robot/suites
[2]: script.sh
[3]: script.py
[4]: vpp_agent_jobs/01A.%20PREPARE%20REPOSITORY%20INSIDE%20JENKINS%20CONTAINER.yml
[5]: pipeline_template.yaml
[6]: 05test_template.yaml
[7]: defaults.yaml
