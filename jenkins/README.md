## What contains this folder

```
-rw-rw-r--  1 stanislav stanislav    20 Mar 29 10:23 README.md
-rw-rw-r--  1 stanislav stanislav   116 Mar 29 10:23 jenkins_jobs.ini
drwxrwxr-x  2 stanislav stanislav  4096 Mar 29 11:47 jobs5
-rwxrwxr-x  1 stanislav stanislav   175 Mar 29 10:23 update.sh
drwxrwxr-x  2 stanislav stanislav  4096 Mar 29 11:10 vpp_agent_jobs

```

* folder `vpp_agent_jobs` contains jenkins jobs yaml definition files for jenkins-job-builder (prepare repository, build docker images, generate single jenkins jobs for respective robot files)
* folder `jobs5` contains templates for single robot tests run inside jenkins jobi (names of jobs beginning by 05), template for pipelines which groups together these jobs (names of jobs beginnig by 04) and python script for processing the list of robot files to data file to be merged with templates . Execute script.sh to get it prepared...
* file jenkins_jobs.ini is a configuration file for jenkins-job-builder`
* script `update.sh` executes jenkins-job-builder to generate jobs beginning by 05 and 04 into jenkins (it merges the data present in p.yaml file with templates)
