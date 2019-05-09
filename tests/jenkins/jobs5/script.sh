#!/bin/bash
find /root/vpp-agent/tests/robot/suites/ -name *.robot -type f  > list_of_all_robot_tests2
sort list_of_all_robot_tests2 > list_of_all_robot_tests
rm list_of_all_robot_tests2
rm p.yaml
python script.py ${1:-51}
more p.yaml
#cd ..
#jenkins-jobs update --delete-old jobs5
#cd jobs5
