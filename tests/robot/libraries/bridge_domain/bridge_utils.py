import vpp_api

from robot.api import logger


def bridge_domain_dump(host, username, password, node):
    """Execute API command bridge_domain_dump on the specified host/node and return the API reply.

    :param host: docker host name or ip
    :param username: username for the docker host
    :param password: password for the docker host
    :param node: VPP node name in docker
    :type host: str
    :type username: str
    :type password: str
    :type node: str
    :returns: API reply data
    :rtype: list
    """

    # Use max uint32 value to dump all ACLs
    int_max = 4294967295

    data = vpp_api.vpp_api.execute_api(
        host, username, password, node, "bridge_domain_dump", bd_id=int_max)

    bridges = []
    for item in data[0]["api_reply"]:
        bridges.append(process_bridge_domain_dump(item))

    return bridges


def process_bridge_domain_dump(data):
    """Process API reply bridge_domain_dump and return dictionary of usable values.

    :param data: API reply from bridge_domain_dump call,
    :type data: dict
    :return: Values ready for comparison with Agent or ETCD values.
    :rtype: dict
    :raises RuntimeError: If the data is in an unexpceted format,
    """

    if len(data) > 1:
        logger.debug(len(data))
        logger.trace(data)
        raise RuntimeError("Data contains more than one API reply.")

    data = data["bridge_domain_details"]

    return data


def filter_bridge_domain_dump_by_id(data, bd_id):
    """Find bridge domain entry in provided dump by the specified bridge domain index.

    :param data: API reply from vridge_domain_dump
    :param bd_id: Bridge domain index.
    :type data: list
    :type bd_id: int
    :returns: Single bridge domain entry.
    :rtype: dict
    :raises RuntimeError: If the bridge domain is not present in API dump."""

    for item in data:
        if str(item["bd_id"]) == str(bd_id):
            return item
    else:
        raise RuntimeError("Bridge domain not found by id {id}.".format(id=bd_id))
