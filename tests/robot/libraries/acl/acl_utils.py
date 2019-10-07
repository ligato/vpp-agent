import vpp_api

from robot.api import logger


def acl_dump(host, username, password, node):

    # Use max uint32 value to dump all ACLs
    int_max = 4294967295

    data = vpp_api.vpp_api.execute_api(
        host, username, password, node, "acl_dump", acl_index=int_max)

    acls = []
    for item in data[0]["api_reply"]:
        acls.append(process_acl_dump(item))

    return acls


def process_acl_dump(data):
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

    data = data["acl_details"]

    ipv6 = int(data["r"][0]["is_ipv6"])
    protocol = int(data["r"][0]["proto"])

    destination_prefix = data["r"][0]["dst_ip_prefix_len"]
    source_prefix = data["r"][0]["src_ip_prefix_len"]

    if ipv6:
        destination_address = data["r"][0]["dst_ip_addr"]["ipv6"]
        source_address = data["r"][0]["src_ip_addr"]["ipv6"]
    else:
        destination_address = data["r"][0]["dst_ip_addr"]["ipv4"]
        source_address = data["r"][0]["src_ip_addr"]["ipv4"]

    destination_network = "/".join([
        str(destination_address),
        str(destination_prefix)])
    source_network = "/".join([
        str(source_address),
        str(source_prefix)])

    output = {
        "acl_name": data["tag"],
        "acl_action": data["r"][0]["is_permit"],
        "ipv6": ipv6,
        "protocol": protocol,
        "destination_network": destination_network,
        "source_network": source_network,
        "destination_port_low": data["r"][0]["dstport_or_icmpcode_first"],
        "destination_port_high": data["r"][0]["dstport_or_icmpcode_last"],
        "source_port_low": data["r"][0]["srcport_or_icmptype_first"],
        "source_port_high": data["r"][0]["srcport_or_icmptype_last"],
        "icmp_code_low": data["r"][0]["dstport_or_icmpcode_first"],
        "icmp_code_high": data["r"][0]["dstport_or_icmpcode_last"],
        "icmp_type_low": data["r"][0]["srcport_or_icmptype_first"],
        "icmp_type_high": data["r"][0]["srcport_or_icmptype_last"]
    }

    if protocol == 6:
        try:
            output["tcp_flags_mask"] = data["r"][0]["tcp_flags_mask"]
            output["tcp_flags_value"] = data["r"][0]["tcp_flags_value"]
        except KeyError:
            pass

    return output


def filter_acl_dump_by_name(data, name):
    for item in data:
        if str(item["acl_name"]) == str(name):
            return item
    else:
        raise RuntimeError("ACL not found by name {name}.".format(name=name))
