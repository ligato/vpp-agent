import json
import re

# input - output from sh int addr
# output - list of words containing ip/prefix
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
def Parse_ARP(info, intf, ip, mac):
    for line in info.splitlines():
        if intf in line and ip in line and mac in line:
            print "ARP Found:"+line
            return True
    print "ARP Found"
    return False

# input - json output from sw_interface_dump, index
# output - whole interface state
def Get_Interface_State(out, index):
    out =  out[out.find('['):out.rfind(']')+1]
    data = json.loads(out)
    state = -1
    for iface in data:
        if iface["sw_if_index"] == int(index):
            state = iface
    return state

# input - mac in dec from sw_interface_dump
# output - regular mac in hex
def Convert_Dec_MAC_To_Hex(mac):
    hexmac=[]
    for num in mac[:6]:
        hexmac.append("%02x" % num)
    hexmac = ":".join(hexmac)
    return hexmac

def replace_rrn(mytext):
    if mytext=="":
        return ""
    mytext=mytext.replace("\r\r\n", "")
    print mytext
    return mytext

def replace_spaces_to_space(mytext):
    mytext = re.sub(
           r" +",
           " ", mytext
           )
    return mytext

def replace_slash_to_space(mytext):
    mytext = re.sub(
           r"/",
           " ", mytext
           )
    return mytext

# input - output from sh int, interface name
# output - index
def Vpp_Get_Interface_Index(out, name):
    out = replace_rrn(out)
    out = replace_spaces_to_space(out)
    data =  out[out.rfind(name):out.rfind(name)+10]
    index = -1
    numbers = [int(s) for s in data.split() if s.isdigit()]
    print data
    print numbers
    if len(numbers) > 0:
       index = numbers[0]
    else:
       print "Index Not Found"
    return index

# input - output from sh int, interface name
# output - whole interface state
def Vpp_Get_Interface_State(out, name):
    out = replace_rrn(out)
    out = replace_spaces_to_space(out)
    data = out[out.rfind(name):out.rfind(name) + 15]
    state = [str(s) for s in data.split()]
    print state
    if state[2] == 'up':
        interfacestate = 1
    elif state[2] == 'down':
        interfacestate = 0
    else:
        print "State Not Found"
    return interfacestate

# input - output from sh int, interface name
# output - mtu
def Vpp_Get_Interface_Mtu(out, name):
    out = replace_rrn(out)
    out = replace_spaces_to_space(out)
    data = out[out.rfind(name):out.rfind(name) + 25]
    state = replace_slash_to_space(data)
    state = [str(s) for s in state.split()]
    if len(data) > 0:
        interfacemtu = state[3]
    else:
        print "Mtu Not Found"
    return interfacemtu

# input - output from sh h, interface name
# output - mac address
def Vpp_Get_Mac_Address(out, name):
    ethadd = "Ethernet address"
    out = replace_rrn(out)
    out = replace_spaces_to_space(out)
    data = out[out.find(name):out.rfind(name) + 70]
    print data
    address = data[data.find(ethadd):data.find(ethadd)+35]
    if len(data) > 0:
        macaddress = re.search(r'([0-9A-F]{2}[:-]){5}([0-9A-F]{2})', address, re.I).group()
        print macaddress
    else:
        print "Mac Address Not Found"
    return macaddress

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
