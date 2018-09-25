#find /root/vpp-agent/tests/robot/suites/ -name *.robot -type f  > list_of_all_robot_tests
import sys
import os
import string
jenkins_project=''
jenkins_project=jenkins_project+'- project:\n' 
jenkins_project=jenkins_project+'    name: ligato/vpp-agent all tests on arm64\n'
jenkins_project=jenkins_project+'    jobs:\n'
with open("vpp_agent_templates/list_of_all_robot_tests") as f:
  for idx,i in enumerate(f):
    print(i)
    if (i == '/root/vpp-agent/tests/robot/suites/trafficIPv6/veth_afpacket_memif_vxlan_traffic/veth_afpacket_memif_vxlan_trafficIPv6.robot'):
      print(idx + '--------------------------')
      filefound=idx
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
    #jenkins_project=jenkins_project+'          inode_of_folder: ' + str(os.stat(odloz2[0]).st_ino) + '\n'
    jenkins_project=jenkins_project+'          inode_of_folder: ' + format(os.path.basename(os.path.normpath(odloz2[0])).upper(), "_<11s") + '\n'
    #jenkins_project=jenkins_project+'          inode_of_folder: 1 \n'
    jenkins_project=jenkins_project+'          name_of_test: '    + os.path.basename(i[:-7]) + '\n' 
    jenkins_project=jenkins_project+'          path_to_test: '    + i[:-1] + '\n'
    jenkins_project=jenkins_project+'          local_variables_file: '    + ( 'arm64_local' if idx < 51 else 'arm64contiv_local' )  + '\n'
    jenkins_project=jenkins_project+'          arm64_node: '    + ( '147.75.98.202' if idx < 51 else '147.75.72.194' )  + '\n'

#jenkins_project=jenkins_project+"      - '04{meno}':\n"
#jenkins_project=jenkins_project+'          meno: ' + 'pokus' + '\n'
g = open('vpp_agent_templates/robot_tests.yaml', 'w')
g.write(jenkins_project)
g.close()
