## What are these files?

These files are a backup of jenkins jobs done with the help of [jjwrecker][1]

```
jjwrecker -f /var/jenkins_home/jobs/01A.\ PREPARE\ REPOSITORY\ INSIDE\ JENKINS\ CONTAINER/config.xml -n 01A.\ PREPARE\ REPOSITORY\ INSIDE\ JENKINS\ CONTAINER
jjwrecker -f /var/jenkins_home/jobs/02A.\ BUILD\ ARM64\ DOCKER\ IMAGES\ for\ ligato\ vpp_agent/config.xml -n 02A.\ BUILD\ ARM64\ DOCKER\ IMAGES\ for\ ligato\ vpp_agent
jjwrecker -f /var/jenkins_home/jobs/02B.\ BUILD\ DOCKER\ IMAGE\ sfc_controller\ for\ arm64/config.xml -n 02B.\ BUILD\ DOCKER\ IMAGE\ sfc_controller\ for\ arm64
jjwrecker -f /var/jenkins_home/jobs/03A.\ GENERATE\ LIGATO-ROBOT-TEST\ JENKINS\ JOBS\ VIA\ JJB/config.xml -n 03A.\ GENERATE\ LIGATO-ROBOT-TEST\ JENKINS\ JOBS\ VIA\ JJB
jjwrecker -f /var/jenkins_home/jobs/M01A.\ Fix\ found\ problems/config.xml -n M01A.\ Fix\ found\ problems
jjwrecker -f /var/jenkins_home/jobs/M02C.\ BUILD\ DOCKER\ IMAGE\ kafka\ for\ arm64/config.xml -n M02C.\ BUILD\ DOCKER\ IMAGE\ kafka\ for\ arm64
jjwrecker -f /var/jenkins_home/jobs/M02D.\ BUILD\ DOCKER\ IMAGE\ libmemif\ for\ arm64/config.xml -n M02D.\ BUILD\ DOCKER\ IMAGE\ libmemif\ for\ arm64
jjwrecker -f /var/jenkins_home/jobs/Setup\ Jenkins\ container/config.xml -n Setup\ Jenkins\ container
```
## Description
These backuped jenkins jobs were prepared for tasks connected to build docker images for ARM64 platform, automatically generate jenkins jobs for robot tests present in [ligato/vpp-agent/robot/suites][2] folder (which is good for case that if some new robot test is added then automatically will be created respective jenkins jobs)

These files is possible to deploy on your jenkins server using jenkins-job-builder:
```
jenkins-jobs update vpp_agent_jobs/
``` 
After deploying on Jenkins server you will need to adjust scripts to your environment (such as servers, credentials, etc)

[1]: https://github.com/ktdreyer/jenkins-job-wrecker
[2]: ../test/robot/suites
