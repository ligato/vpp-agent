# input - output from sh int addr
# output - list of words containing ip/prefix
def Find_IPV4_In_Text(text):
    ipv4 = []
    for word in text.split():
        if (word.count('.') == 3) and (word.count('/') == 1):
            ipv4.append(word)
    return ipv4

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