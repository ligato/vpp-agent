import json

# input - output from 'ip a' command
# output - interfaces with parameters in json
def Parse_Linux_Interfaces(data):
    ints = {}
    for line in data.splitlines():
        if line[0] != ' ':
            if_name = line.split()[1][:-1]
            ints[if_name] = {}
            if "mtu" in line:
                ints[if_name]["mtu"] = line[line.find("mtu"):].split()[1]
            if "state" in line:
                ints[if_name]["state"] = line[line.find("state"):].split()[1].lower()
        else:
            line = line.strip()
            if "link/" in line:
                ints[if_name]["mac"] = line.split()[1]
            if "inet " in line:
                ints[if_name]["ipv4"] = line.split()[1]
            if "inet6" in line:
                ints[if_name]["ipv6"] = line.split()[1]
    return ints

def Pick_Linux_Interface(ints, name):
    int = []
    for key in ints[name]:
        int.append(key+"="+ints[name][key])
    return int


# input - json output from Parse_Linux_Interfaces
# output - true if interface exist, false if not
def Check_Linux_Interface_Presence(data, mac):
    present = False
    if_index = -1
    for iface in data:
        if data[iface]["mac"] == mac:
            present = True
    return present


x='''1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN group default qlen 1
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
    inet6 ::1/128 scope host 
       valid_lft forever preferred_lft forever
2: vpp1_veth2@vpp1_veth1: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP group default 
    link/ether a2:a1:a1:a1:a1:a1 brd ff:ff:ff:ff:ff:ff
    inet6 fe80::14f3:b3ff:fe13:a7cb/64 scope link 
       valid_lft forever preferred_lft forever
3: vpp1_veth1@vpp1_veth2: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP group default 
    link/ether 12:11:11:11:11:11 brd ff:ff:ff:ff:ff:ff
    inet 10.10.1.1/24 scope global vpp1_veth1
       valid_lft forever preferred_lft forever
    inet6 fe80::1011:11ff:fe11:1111/64 scope link 
       valid_lft forever preferred_lft forever
4: linux_vpp1_tap1: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc pfifo_fast state UNKNOWN group default qlen 1000
    link/ether e6:45:54:36:94:2d brd ff:ff:ff:ff:ff:ff
    inet6 fe80::e445:54ff:fe36:942d/64 scope link 
       valid_lft forever preferred_lft forever
2962: eth0@if2963: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP group default 
    link/ether 02:42:ac:11:00:05 brd ff:ff:ff:ff:ff:ff link-netnsid 0
    inet 172.17.0.5/16 scope global eth0
       valid_lft forever preferred_lft forever
    inet6 fe80::42:acff:fe11:5/64 scope link 
       valid_lft forever preferred_lft forever'''

#Parse_Linux_Interfaces(x)
#Pick_Linux_Interface(x,"vpp1_veth1@vpp1_veth2")
