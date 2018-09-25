find /root/vpp-agent/tests/robot/suites/ -name *.robot -type f  > vpp_agent_templates/list_of_all_robot_tests
rm -f vpp_agent_templates/robot_tests.yaml
python vpp_agent_templates/generate_robot_tests.py
cat vpp_agent_templates/robot_tests.yaml
/root/.local/bin/jenkins-jobs update --delete-old vpp_agent_templates
