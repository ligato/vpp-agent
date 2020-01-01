# input - output from sh int addr
# output - list of words containing ip/prefix
from robot.api import logger

def Find_IPV4_In_Text(text):
    ipv4 = []
    for word in text.split():
        if (word.count('.') == 3) and (word.count('/') == 1):
            ipv4.append(word)
    return ipv4


def Find_IPV6_In_Text(text):
    """Find and return all IPv6 addresses in the given string.

    :param text: string to search.
    :type text: str

    :return: IPv6 addresses found in string.
    :rtype: list of str
    """

    ipv6 = []
    for word in text.split():
        if (word.count(':') >= 2) and (word.count('/') == 1):
            ipv6.append(word)
    return ipv6


# input - output from sh hardware interface_name
# output - list of words containing mac
def Find_MAC_In_Text(text):
    mac = ''
    for word in text.split():
        if (word.count(':') == 5):
            mac = word
            break
    return mac


# input - output from sh ip arp command
# output - state info list
def parse_arp(info, intf, ip, mac):
    """Parse ARP list from vpp console and find a specific entry using the provided arguments.

    :param info: ARP list from VPP console.
    :param intf: VPP-internal name of the interface configured with this ARP entry.
    :param ip: IP address of the ARP entry.
    :param mac: MAC address of the ARP entry.
    :type info: str
    :type intf: str
    :type ip: str
    :type mac: str
    :returns: True if a matching entry is found, else False
    :rtype: bool
    """

    for line in info.splitlines():
            if intf in line and ip in line and mac in line:
                print("ARP Found:"+line)
                return True
    logger.debug("ARP not Found")
    return False


# input - output from sh ip arp command
# output - state info list
def parse_neighbor(info, intf, ip, mac):
    """Parse neighbor list from vpp console and find a specific entry using the provided arguments.

    :param info: Neighbor list from VPP console.
    :param intf: VPP-internal name of the interface configured with this neighbor.
    :param ip: IP address of the neighbor entry.
    :param mac: MAC address of the neighbor entry.
    :type info: str
    :type intf: str
    :type ip: str
    :type mac: str
    :returns: True if a matching entry is found, else False
    :rtype: bool
    """

    for line in info.splitlines():
        if intf in line and ip in line and mac in line:
            print("Neighbor Found:"+line)
            return True
    logger.debug("Neighbor not Found")
    return False


# input - output from sh ip arp command
# output - state info list
def parse_stn_rule(info):
    state = {}
    for line in info.splitlines():
        try:
            if "address" in line.strip().split()[0]:
                state['ip_address'] = line.strip().split()[1]
            elif "iface" in line.strip().split()[0]:
                state['iface'] = line.strip().split()[1]
            elif "next_node" in line.strip().split()[0]:
                state['next_node'] = line.strip().split()[1]
        except IndexError:
            pass

    return state['ip_address'], state['iface'], state['next_node']


def parse_memif_info(info):
    state = []
    sockets_line = []
    for line in info.splitlines():
        if line:
            try:
                _ = int(line.strip().split()[0])
                sockets_line.append(line)
            except ValueError:
                pass
            if line.strip().split()[0] == "flags":
                if "admin-up" in line:
                    state.append("enabled=1")
                if "slave" in line:
                    state.append("role=slave")
                if "connected" in line:
                    state.append("connected=1")
            if line.strip().split()[0] == "socket-id":
                try:
                    socket_id = int(line.strip().split()[1])
                    state.append("id="+line.strip().split()[3])
                    for sock_line in sockets_line:
                        try:
                            num = int(sock_line.strip().split()[0])
                            if num == socket_id:
                                state.append("socket=" + sock_line.strip().split()[-1])
                        except ValueError:
                            pass
                except ValueError:
                    pass
    if "enabled=1" not in state:
        state.append("enabled=0")
    if "role=slave" not in state:
        state.append("role=master")
    if "connected=1" not in state:
        state.append("connected=0")
    return state
