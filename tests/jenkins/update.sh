#!/bin/bash
mkdir -p /etc/jenkins_jobs
cp jenkins_jobs.ini /etc/jenkins_jobs
#jenkins-jobs update --delete-old jobs5/  &> update.log
jenkins-jobs update jobs5/  &> update.log
