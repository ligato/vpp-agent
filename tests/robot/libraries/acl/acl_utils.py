from ipaddress import ip_address, IPv6Address, AddressValueError

from robot.api import logger

acl_actions = {
    0: "deny",
    1: "permit",
    2: "permit+reflect"
}

protocol_ids = {
    "ICMP":   1,
    "TCP":    6,
    "UDP":    17,
    "ICMPV6": 58
}


def filter_acl_by_name(acl_data, acl_name):
    """Find and return the specified ACL entry by name.

    :param acl_data: Reply from VPP terminal
    :param acl_name: Name of the requested ACL.
    :type acl_data: str
    :type acl_name: str
    :returns: Data for a single ACL entry.
    :rtype: str
    :raises RuntimeError: If the ACL name is not present.
    """
    if len(acl_data) > 0 and not acl_data.startswith("acl-index "):
        raise RuntimeError("Unexpected format of ACL data.")

    entries = acl_data.split("acl-index ")

    for entry in entries:
        if acl_name in entry:
            data = entry
            data = data.replace("\r", "")
            data = data.strip()
            return "{0}".format(data[2:])
        else:
            logger.trace(
                "ACL name '{name}' not found in entry '''{entry}'''".format(
                    name=acl_name, entry=entry))

    else:
        logger.debug("Response data: '''{data}'''".format(data=acl_data))
        raise RuntimeError("ACL name {name} not found in response data.".format(name=acl_name))


def translate_acl_action(acl_action_id):

    logger.trace(acl_action_id)
    try:
        return acl_actions[acl_action_id]
    except KeyError:
        raise NotImplementedError("We only know the basic ACL actions: 0,1,2")


def translate_protocol_name(protocol_name):

    logger.trace(protocol_name)
    try:
        return protocol_ids[protocol_name.upper()]
    except KeyError:
        raise NotImplementedError("Unknown IP protocol name. Options: TCP, UDP, ICMP, ICMPv6")


def expand_ipv6_network(network):

    logger.trace(network)
    if "/" not in network:
        raise RuntimeError("Network address format unrecognized.")

    address, prefix = network.split("/")
    address = ip_address(address)

    return "{address}/{prefix}".format(address=address, prefix=prefix)


def prepare_acl_variables(
        ingress_interfaces, egress_interfaces,
        acl_action, protocol, source_network, destination_network):

    if ingress_interfaces or egress_interfaces:
        # TODO: add support for ingress and egress interfaces, add test
        raise NotImplementedError("Ingress/Egress interface format in dump unknown.")

    try:
        int(acl_action)
        acl_action = translate_acl_action(int(acl_action))
    except ValueError:
        # assume it's in string form already
        pass

    try:
        IPv6Address(source_network.split("/")[0])
        ip_version = "ipv6"
    except AddressValueError:
        ip_version = "ipv4"

    if ip_version == "ipv6" and protocol == "ICMP":
        protocol = "ICMPV6"

    protocol = protocol_ids[protocol.upper()]

    if ip_version == "ipv6":
        source_network = expand_ipv6_network(source_network)
        destination_network = expand_ipv6_network(destination_network)

    return acl_action, protocol, ip_version, source_network, destination_network


def replace_acl_variables_tcp(
        template, acl_name,
        ingress_interfaces, egress_interfaces,
        acl_action,
        destination_network, source_network,
        destination_port_min, destination_port_max,
        source_port_min, source_port_max,
        tcp_flags_mask, tcp_flags_value
        ):

    acl_action, protocol, ip_version, source_network, destination_network = prepare_acl_variables(
        ingress_interfaces, egress_interfaces, acl_action, "TCP", source_network, destination_network)

    try:
        data = template.format(
            acl_name=acl_name, acl_action=acl_action,
            src_net=source_network, dst_net=destination_network,
            src_port_range="-".join([source_port_min, source_port_max]),
            dst_port_range="-".join([destination_port_min, destination_port_max]),
            ip_version=ip_version, protocol=protocol,
            flags="tcpflags "+tcp_flags_value, mask="mask "+tcp_flags_mask
        )
    except KeyError:
        logger.warn("Template requires additional variables.")
        raise

    return data.strip()


def replace_acl_variables_udp(
        template, acl_name,
        ingress_interfaces, egress_interfaces,
        acl_action,
        destination_network, source_network,
        destination_port_min, destination_port_max,
        source_port_min, source_port_max,
):

    acl_action, protocol, ip_version, source_network, destination_network = prepare_acl_variables(
        ingress_interfaces, egress_interfaces, acl_action, "UDP", source_network, destination_network)

    try:
        data = template.format(
            acl_name=acl_name, acl_action=acl_action,
            src_net=source_network, dst_net=destination_network,
            src_port_range="-".join([source_port_min, source_port_max]),
            dst_port_range="-".join([destination_port_min, destination_port_max]),
            ip_version=ip_version, protocol=protocol
        )

    except KeyError:
        logger.warn("Template requires additional variables.")
        raise

    return data.strip()


def replace_acl_variables_icmp(
        template, acl_name,
        ingress_interfaces, egress_interfaces,
        acl_action,
        destination_network, source_network,
        ipv6, icmp_code_min, icmp_code_max,
        icmp_type_min, icmp_type_max
):

    acl_action, protocol, ip_version, source_network, destination_network = prepare_acl_variables(
        ingress_interfaces, egress_interfaces, acl_action, "ICMP", source_network, destination_network)

    try:
        data = template.format(
            acl_name=acl_name, acl_action=acl_action,
            src_net=source_network, dst_net=destination_network,
            ip_version=ip_version, protocol=protocol, ipv6=ipv6,
            icmp_code_range="-".join([icmp_code_min, icmp_code_max]),
            icmp_type_range="-".join([icmp_type_min, icmp_type_max])
        )

    except KeyError:
        logger.warn("Template requires additional variables.")
        raise

    return data.strip()
