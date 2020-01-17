
def get_interface_index_from_api(data, name):
    """Process data from sw_interface_dump API and return index
     of the interface specified by name.

     :param data: Output of interface dump API call.
     :param name: Name of the interface to find.
     :type data: list
     :type name: str
     :returns: Index of the specified interface.
     :rtype: int
     :raises RuntimeError: If the interface is not found.
     """

    for iface in data:
        if iface["sw_interface_details"]["interface_name"] == name:
            return iface["sw_interface_details"]["sw_if_index"]
    else:
        raise RuntimeError(
            "Interface with name {name} not found in dump. "
            "Dumped data: {data}".format(name=name, data=data))


def get_interface_name_from_api(data, index):
    """Process data from sw_interface_dump API and return index
     of the interface specified by name.

     :param data: Output of interface dump API call.
     :param index: Index of the interface to find.
     :type data: list
     :type index: int
     :returns: Name of the specified interface.
     :rtype: str
     :raises RuntimeError: If the interface is not found.
     """

    for iface in data:
        if iface["sw_interface_details"]["sw_if_index"] == index:
            return iface["sw_interface_details"]["interface_name"]
    else:
        raise RuntimeError(
            "Interface with index {index} not found in dump. "
            "Dumped data: {data}".format(index=index, data=data))


def get_interface_state_from_api(data, index=-1, name="_"):
    """Process data from sw_interface_dump API and return state
    of the interface specified by name or index.

    :param data: Output of interface dump API call.
    :param index: Index of the interface to find.
    :param name: Name of the interface to find.
    :type data: list
    :type index: int
    :type name: str
    :returns: State of the specified interface.
    :rtype: str
    :raises RuntimeError: If the interface is not found.
    """

    if index == -1 and name == "_":
        raise ValueError("Provide either an interface index or a name.")

    for iface in data:
        if iface["sw_interface_details"]["sw_if_index"] == int(index)\
                or iface["sw_interface_details"]["interface_name"] == name:
            return iface["sw_interface_details"]
    else:
        raise RuntimeError(
            "Interface with index {index} or name {name} not found in dump.\n "
            "Dumped data:\n {data}".format(index=index, name=name, data=data))


def convert_str_to_mac_address(hex_mac):
    """Just add colons in the right places."""

    hex_pairs = [hex_mac[(x*2):((x+1)*2)] for x in range(6)]

    mac = "{0}:{1}:{2}:{3}:{4}:{5}".format(*hex_pairs)

    return mac
