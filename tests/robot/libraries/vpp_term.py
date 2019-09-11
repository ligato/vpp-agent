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
def Parse_ARP(info, intf, ip, mac):
    for line in info.splitlines():
        if intf in line and ip in line and mac in line:
            print "ARP Found:"+line
            return True
    print "ARP Found"
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

def verify_api_response(data, response):
    """Verify VAT terminal response against expected data.

    :param data: Expected data.
    :param response: Data in API response.
    :type data: str
    :type response: str
    :raises RuntimeError: If the response does not match expected data.
    """

    data_lines = data.splitlines()
    response_lines = response.splitlines()

    # Strip API name and handler ID from the response
    for index, line in enumerate(response_lines):
        if line.startswith("vl_api"):
            parts = line.split(":")
            response_lines[index] = ":".join(parts[2:])

    # Strip API name and handler ID from expected data
    for index, line in enumerate(data_lines):
        if line.startswith("vl_api"):
            parts = line.split(":")
            data_lines[index] = ":".join(parts[2:])

    # Strip whitespaces
    data_lines = [x.strip() for x in data_lines]
    response_lines = [x.strip() for x in response_lines]

    for data_line in data_lines:
        if data_line in response_lines:
            logger.trace("Data line '''{line}''' matched to response line.".format(line=data_line))
        else:
            logger.debug("full response: '''{response}'''".format(response=response_lines))
            raise RuntimeError(
                "Expected data line '''{line}''' not present in response."
                    .format(line=data_line)
            )
