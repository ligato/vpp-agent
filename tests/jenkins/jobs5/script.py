#find /root/vpp-agent/tests/robot/suites/ -name *.robot -type f  > list_of_all_robot_tests
# todo : slovak rozdel_na_pozicii
# todo : input parameter - to help divide the tests to two group, even non evenly
# todo : define agent (declarative) or node (scripted) ... preferably ARM64 node
import sys
import os
import string
import datetime

print(sys.argv[1])
rozdel_na_pozicii=int(sys.argv[1])

running_at_arm64_node_I=''
running_at_arm64_node_I=running_at_arm64_node_I+'              '+'stage \'test\'\n'
running_at_arm64_node_II=''
running_at_arm64_node_II=running_at_arm64_node_II+'              '+'stage \'test\'\n'
running_at_arm64_node_I_IPv4=''
running_at_arm64_node_I_IPv4=running_at_arm64_node_I_IPv4+'              '+'stage \'test\'\n'
running_at_arm64_node_II_IPv6=''
running_at_arm64_node_II_IPv6=running_at_arm64_node_II_IPv6+'              '+'stage \'test\'\n'
running_at_arm64_node_II_OTHER=''
running_at_arm64_node_II_OTHER=running_at_arm64_node_II_OTHER+'              '+'stage \'test\'\n'
running_at_arm64_node_I_SFCIPv4=''
running_at_arm64_node_I_SFCIPv4=running_at_arm64_node_I_SFCIPv4+'              '+'stage \'test\'\n'
running_at_arm64_node_II_SFCIPv6=''
running_at_arm64_node_II_SFCIPv6=running_at_arm64_node_II_SFCIPv6+'              '+'stage \'test\'\n'

jenkins_project=''
jenkins_project=jenkins_project+'- project:\n'
jenkins_project=jenkins_project+'    name: ligato/vpp-agent all tests on arm64\n'
jenkins_project=jenkins_project+'    jobs:\n'
with open("list_of_all_robot_tests") as f:
  for idx,i in enumerate(f):
    print(i)

    if (idx == rozdel_na_pozicii):
      print(str(idx) + '--------------------------')

    with open(i[:-1]) as myfile:
      if 'ExpectedFailure' in myfile.read():
        is_expected_failure=True
      else:
        is_expected_failure=False

    print(os.path.basename(i[:-7]))
    x = (i,'')
    while True:
      odloz=x
      x=os.path.split(x[0])
      if (x[1] == 'suites'):
        break
      else:
        odloz2=odloz

    print(odloz2[0])
    print(os.stat(odloz2[0]).st_ino)
    jenkins_project=jenkins_project+"      - '05{inode_of_folder}_{name_of_test}_job':\n"
    jenkins_project=jenkins_project+'          HOWTOBUILD_INCLTAGPRESENT: ' + ( '--include ExpectedFailure' if is_expected_failure else '\' \'' ) + '\n'
    jenkins_project=jenkins_project+'          HOWTOBUILD_EXCLTAGPRESENT: ' + ( '--exclude ExpectedFailure' if is_expected_failure else '\' \'' ) + '\n'
    jenkins_project=jenkins_project+'          inode_of_folder: ' + format(os.path.basename(os.path.normpath(odloz2[0])).upper(), "_<11s") + '\n'
    jenkins_project=jenkins_project+'          date_of_jjb_generation: ' + datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S") + '\n'
    jenkins_project=jenkins_project+'          name_of_test: '    + os.path.basename(i[:-7]) + '\n'
    jenkins_project=jenkins_project+'          path_to_test: '    + i[:-1] + '\n'
    jenkins_project=jenkins_project+'          local_variables_file: '    + ( 'arm64_local' if idx < rozdel_na_pozicii else 'arm64_II_local' )  + '\n'
    # following row is enough ti control the ARM64 node where test is going to be executed
    jenkins_project=jenkins_project+'          arm64_node: '    + ( '147.75.72.194' if idx < rozdel_na_pozicii else '147.75.98.202' )  + '\n'
    #I will divide the test to pipelines according IPv4/IPv6
    groupoftests=os.path.basename(os.path.normpath(odloz2[0])).upper()
    if (groupoftests=='CRUD') or (groupoftests=='TRAFFIC'):
        running_at_arm64_node_I_IPv4=running_at_arm64_node_I_IPv4+'              build job: \'05'+ format(os.path.basename(os.path.normpath(odloz2[0])).upper(), "_<11s") +'_'+ os.path.basename(i[:-7]) +'_job\', parameters: [string(name: \'HOWTOBUILD\', value: "${{HOWTOBUILD}}"), string(name: \'LOGLEVEL\', value: "${{LOGLEVEL}}"), string(name: \'VARIABLES_FILE\', value: "${{VARIABLES_FILE}}"), string(name: \'DOCKER_HOST_IP\', value: "${{DOCKER_HOST_IP}}"), string(name: \'IMAGE_NAME\', value: "${{IMAGE_NAME}}")], propagate: false, quietPeriod: 60\n'
    elif (groupoftests=='CRUDIPV6') or (groupoftests=='TRAFFICIPV6'):
        running_at_arm64_node_II_IPv6=running_at_arm64_node_II_IPv6+'              build job: \'05'+ format(os.path.basename(os.path.normpath(odloz2[0])).upper(), "_<11s") +'_'+ os.path.basename(i[:-7]) +'_job\', parameters: [string(name: \'HOWTOBUILD\', value: "${{HOWTOBUILD}}"), string(name: \'LOGLEVEL\', value: "${{LOGLEVEL}}"), string(name: \'VARIABLES_FILE\', value: "${{VARIABLES_FILE}}"), string(name: \'DOCKER_HOST_IP\', value: "${{DOCKER_HOST_IP}}"), string(name: \'IMAGE_NAME\', value: "${{IMAGE_NAME}}")], propagate: false, quietPeriod: 60\n'
    elif (groupoftests=='SFC'):
        running_at_arm64_node_I_SFCIPv4=running_at_arm64_node_I_SFCIPv4+'              build job: \'05'+ format(os.path.basename(os.path.normpath(odloz2[0])).upper(), "_<11s") +'_'+ os.path.basename(i[:-7]) +'_job\', parameters: [string(name: \'HOWTOBUILD\', value: "${{HOWTOBUILD}}"), string(name: \'LOGLEVEL\', value: "${{LOGLEVEL}}"), string(name: \'VARIABLES_FILE\', value: "${{VARIABLES_FILE}}"), string(name: \'DOCKER_HOST_IP\', value: "${{DOCKER_HOST_IP}}"), string(name: \'IMAGE_NAME\', value: "${{IMAGE_NAME}}")], propagate: false, quietPeriod: 60\n'
    elif (groupoftests=='SFCIPV6'):
        running_at_arm64_node_II_SFCIPv6=running_at_arm64_node_II_SFCIPv6+'              build job: \'05'+ format(os.path.basename(os.path.normpath(odloz2[0])).upper(), "_<11s") +'_'+ os.path.basename(i[:-7]) +'_job\', parameters: [string(name: \'HOWTOBUILD\', value: "${{HOWTOBUILD}}"), string(name: \'LOGLEVEL\', value: "${{LOGLEVEL}}"), string(name: \'VARIABLES_FILE\', value: "${{VARIABLES_FILE}}"), string(name: \'DOCKER_HOST_IP\', value: "${{DOCKER_HOST_IP}}"), string(name: \'IMAGE_NAME\', value: "${{IMAGE_NAME}}")], propagate: false, quietPeriod: 60\n'
    else:
        running_at_arm64_node_II_OTHER=running_at_arm64_node_II_OTHER+'              build job: \'05'+ format(os.path.basename(os.path.normpath(odloz2[0])).upper(), "_<11s") +'_'+ os.path.basename(i[:-7]) +'_job\', parameters: [string(name: \'HOWTOBUILD\', value: "${{HOWTOBUILD}}"), string(name: \'LOGLEVEL\', value: "${{LOGLEVEL}}"), string(name: \'VARIABLES_FILE\', value: "${{VARIABLES_FILE}}"), string(name: \'DOCKER_HOST_IP\', value: "${{DOCKER_HOST_IP}}"), string(name: \'IMAGE_NAME\', value: "${{IMAGE_NAME}}")], propagate: false, quietPeriod: 60\n'


    if idx < rozdel_na_pozicii:
        running_at_arm64_node_I=running_at_arm64_node_I+'              build job: \'05'+ format(os.path.basename(os.path.normpath(odloz2[0])).upper(), "_<11s") +'_'+ os.path.basename(i[:-7]) +'_job\', parameters: [string(name: \'HOWTOBUILD\', value: "${{HOWTOBUILD}}"), string(name: \'LOGLEVEL\', value: "${{LOGLEVEL}}"), string(name: \'VARIABLES_FILE\', value: "${{VARIABLES_FILE}}"), string(name: \'DOCKER_HOST_IP\', value: "${{DOCKER_HOST_IP}}"), string(name: \'IMAGE_NAME\', value: "${{IMAGE_NAME}}")], propagate: false, quietPeriod: 60\n'
    else:
        #running_at_arm64_node_II=running_at_arm64_node_II+'              build job: \'05'+ format(os.path.basename(os.path.normpath(odloz2[0])).upper(), "_<11s") +'_'+ os.path.basename(i[:-7]) +'_job\', propagate: false\n'
        running_at_arm64_node_II=running_at_arm64_node_II+'              build job: \'05'+ format(os.path.basename(os.path.normpath(odloz2[0])).upper(), "_<11s") +'_'+ os.path.basename(i[:-7]) +'_job\', parameters: [string(name: \'HOWTOBUILD\', value: "${{HOWTOBUILD}}"), string(name: \'LOGLEVEL\', value: "${{LOGLEVEL}}"), string(name: \'VARIABLES_FILE\', value: "${{VARIABLES_FILE}}"), string(name: \'DOCKER_HOST_IP\', value: "${{DOCKER_HOST_IP}}"), string(name: \'IMAGE_NAME\', value: "${{IMAGE_NAME}}")], propagate: false, quietPeriod: 60\n'




jenkins_project=jenkins_project+'      - \'04{name_of_pipeline}_pipeline\':\n'
jenkins_project=jenkins_project+'          name_of_pipeline: IPv4_arm64_node_I\n'
jenkins_project=jenkins_project+'          date_of_jjb_generation: ' + datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S") + '\n'
jenkins_project=jenkins_project+'          local_variables_file: \'arm64_local\'\n'
jenkins_project=jenkins_project+'          arm64_node: \'147.75.72.194\'\n'
jenkins_project=jenkins_project+'          list_of_jenkins_jobs: |-\n' +running_at_arm64_node_I_IPv4+ ' \n'

jenkins_project=jenkins_project+'      - \'04{name_of_pipeline}_pipeline\':\n'
jenkins_project=jenkins_project+'          name_of_pipeline: IPv6_arm64_node_II\n'
jenkins_project=jenkins_project+'          date_of_jjb_generation: ' + datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S") + '\n'
jenkins_project=jenkins_project+'          local_variables_file: \'arm64_II_local\'\n'
jenkins_project=jenkins_project+'          arm64_node: \'147.75.98.202\'\n'
jenkins_project=jenkins_project+'          list_of_jenkins_jobs: |-\n' +running_at_arm64_node_II_IPv6+ ' \n'

jenkins_project=jenkins_project+'      - \'04{name_of_pipeline}_pipeline\':\n'
jenkins_project=jenkins_project+'          name_of_pipeline: SFCIPv4_arm64_node_I\n'
jenkins_project=jenkins_project+'          date_of_jjb_generation: ' + datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S") + '\n'
jenkins_project=jenkins_project+'          local_variables_file: \'arm64_local\'\n'
jenkins_project=jenkins_project+'          arm64_node: \'147.75.72.194\'\n'
jenkins_project=jenkins_project+'          list_of_jenkins_jobs: |-\n' +running_at_arm64_node_I_SFCIPv4+ ' \n'

jenkins_project=jenkins_project+'      - \'04{name_of_pipeline}_pipeline\':\n'
jenkins_project=jenkins_project+'          name_of_pipeline: SFCIPv6_arm64_node_II\n'
jenkins_project=jenkins_project+'          date_of_jjb_generation: ' + datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S") + '\n'
jenkins_project=jenkins_project+'          local_variables_file: \'arm64_II_local\'\n'
jenkins_project=jenkins_project+'          arm64_node: \'147.75.98.202\'\n'
jenkins_project=jenkins_project+'          list_of_jenkins_jobs: |-\n' +running_at_arm64_node_II_SFCIPv6+ ' \n'

jenkins_project=jenkins_project+'      - \'04{name_of_pipeline}_pipeline\':\n'
jenkins_project=jenkins_project+'          name_of_pipeline: OTHER_arm64_node_II\n'
jenkins_project=jenkins_project+'          date_of_jjb_generation: ' + datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S") + '\n'
jenkins_project=jenkins_project+'          local_variables_file: \'arm64_II_local\'\n'
jenkins_project=jenkins_project+'          arm64_node: \'147.75.98.202\'\n'
jenkins_project=jenkins_project+'          list_of_jenkins_jobs: |-\n' +running_at_arm64_node_II_OTHER+ ' \n'

g = open('p.yaml', 'w')
g.write(jenkins_project)
g.close()

#g = open('r1', 'w')
#g.write(running_at_arm64_node_I)
#g.close()

#g = open('r2', 'w')
#g.write(running_at_arm64_node_II)
#g.close()
