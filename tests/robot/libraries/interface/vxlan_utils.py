import vpp_api

from robot.api import logger


def check_vxlan_tunnel_presence_from_api(data, src, dst, vni):

    for iface in data:
        if iface["src_address"] == src and iface["dst_address"] == dst and iface["vni"] == int(vni):
            logger.trace("matched interface: {interface}".format(interface=iface))
            return True, iface["sw_if_index"]
    else:
        logger.debug(
            "interface with:\n"
            "src_addr: {src_address}\n"
            "dst_addr: {dst_address}\n"
            "vni: {vni}\n"
            "not found in dump. Full dump:\n"
            "{data}".format(
                src_address=src, dst_address=dst, vni=vni, data=data))


def vxlan_tunnel_dump(host, username, password, node):

    # Use max uint32 value to dump all tunnels
    int_max = 4294967295

    data = vpp_api.vpp_api.execute_api(
        host, username, password, node, "vxlan_tunnel_dump", sw_if_index=int_max)

    interfaces = []
    for interface in data[0]["api_reply"]:
        interfaces.append(process_vxlan_dump(interface))

    return interfaces


def process_vxlan_dump(data):
    """
    Process API reply acl_dump and return dictionary of usable values.

    :param data: API reply from acl_dump call,
    :type data: dict
    :return: Values ready for comparison with Agent or ETCD values.
    :rtype: dict
    """

    if len(data) > 1:
        logger.debug(len(data))
        logger.trace(data)
        raise RuntimeError("Data contains more than one API reply.")

    data = data["vxlan_tunnel_details"]

    index = int(data["sw_if_index"])
    mcast_index = int(data["mcast_sw_if_index"])

    ipv6 = int(data["is_ipv6"])

    if ipv6:
        destination_address = data["dst_address"]["ipv6"]
        source_address = data["src_address"]["ipv6"]
    else:
        destination_address = data["dst_address"]["ipv4"]
        source_address = data["src_address"]["ipv4"]

    vrf_id = int(data["encap_vrf_id"])
    vni = int(data["vni"])

    next_index = int(data["decap_next_index"])

    instance = int(data["instance"])

    output = {
        "sw_if_index": index,
        "ipv6": ipv6,
        "vrf": vrf_id,
        "vni": vni,
        "dst_address": destination_address,
        "src_address": source_address,
        "next_index": next_index,
        "mcast_index": mcast_index,
        "instance": instance
    }

    return output


def filter_vxlan_tunnel_dump_by_index(data, index):
    for item in data:
        if int(item["sw_if_index"]) == int(index):
            return item
    else:
        raise RuntimeError("ACL not found by sw_if_index {index}.".format(index=index))
