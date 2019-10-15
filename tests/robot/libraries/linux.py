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
            if "inet6" in line and "scope link" not in line:
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
    for iface in data:
        if data[iface]["mac"] == mac:
            present = True
    return present

# input - json output from Parse_Linux_Interfaces
# output - true if interface exist, false if not
def Check_Linux_Interface_IP_Presence(data, mac, ip):
    present_mac = False
    present_ip = False
    for iface in data:
        if  "mac" in  data[iface]:
           if data[iface]["mac"] == mac:
              present_mac = True
        if "ipv4" in data[iface]:
           if data[iface]["ipv4"] == ip:
              present_ip = True
        if "ipv6" in data[iface]:
           if data[iface]["ipv6"] == ip:
              present_ip = True
    if present_mac == True and present_ip == True:
        return True
    else:
        return False


def parse_linux_arp_entries(data):
    """Parse output of arp command and return list of ARP entries.

    :param data: output of 'arp' command.
    :type data: str
    :returns: Parsed ARP entries.
    :rtype: list of dict
    """
    entries = []

    for line in data.splitlines():
        if "ip address" in line.lower():
            # skip column headers line
            continue
        entry_data = line.split()
        entry = {
            "interface": entry_data[5],
            "ip_addr": entry_data[0],
            "mac_addr": entry_data[3]
        }
        entries.append(entry)

    return entries


def parse_linux_ipv6_neighbor_entries(data):
    """Parse output of ip neighbor command and return list of neighbor entries.

    :param data: output of 'ip neighbor' command.
    :type data: str
    :returns: Parsed neighbor entries.
    :rtype: list of dict
    """
    entries = []

    for line in data.splitlines():
        if "ip address" in line.lower():
            # skip column headers line
            continue
        entry_data = line.split()
        entry = {
            "interface": entry_data[2],
            "ip_addr": entry_data[0],
            "mac_addr": entry_data[4]
        }
        entries.append(entry)

    return entries

