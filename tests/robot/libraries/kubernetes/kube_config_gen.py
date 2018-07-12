import yaml
import os
import math
"""Generates YAML config files for use with kubernetes."""


def mac_hex(number):
    """Convert integer to hexadecimal for incrementing MAC addresses."""
    temp = hex(number)[2:]
    if number < 16:
        temp = "0{0}".format(temp)
    elif number > 99:
        raise NotImplementedError(
            "Incrementing MAC addresses only implemented up to 99.")
    else:
        pass
    return temp


def yaml_replace_line(yaml_string, line_identifier, replacement):
    """Replace a single line in the specified file."""
    for line in yaml_string.splitlines():
        if line_identifier in line:
            whitespace = len(line) - len(line.lstrip(" "))
            return yaml_string.replace(line, "{spaces}{content}".format(
                spaces=" " * whitespace,
                content=replacement
            ))


class YamlConfigGenerator(object):
    def __init__(self, vnf_count, novpp_count, memif_per_vnf, template_folder):
        """Initialize config generator with topology parameters.

        :param vnf_count: Number of VNF nodes.
        :param novpp_count: Number of non-VPP nodes.
        :param template_folder: Path to .yaml config templates.
        :type vnf_count: int
        :type novpp_count: int
        :type template_folder: str
        """

        self.vnf_count = int(vnf_count)
        self.novpp_count = int(novpp_count)
        self.memif_per_vnf = int(memif_per_vnf)
        self.templates = {}
        self.output = {}
        self.load_templates(template_folder)

        if self.novpp_count % self.memif_per_vnf != 0:
            raise NotImplementedError("Number of non-VPP containers must be"
                                      " a multiple of bridge domain count.")

    def load_templates(self, template_folder):
        with open("{0}/sfc-k8.yaml".format(template_folder), "r") as sfc:
            self.templates["sfc"] = sfc.read()
        with open("{0}/vnf-vpp.yaml".format(template_folder), "r") as vnf:
            self.templates["vnf"] = vnf.read()
        with open("{0}/novpp.yaml".format(template_folder), "r") as novpp:
            self.templates["novpp"] = novpp.read()

    def generate_config(self, output_path):
        """Generate topology config .yaml files for Kubernetes.

        :param output_path: Path to where the output files should be placed.
        :type output_path: str
        """
        self.generate_sfc_config()
        self.generate_vnf_config()
        self.generate_novpp_config()
        self.write_config_files(output_path)

    def generate_sfc_config(self):

        entities_list = []

        for bridge_index in range(self.memif_per_vnf):
            entity = {
                "name": "L2Bridge-{0}".format(bridge_index),
                "description": "TODO",
                "type": 3,
                "bd_parms": {
                    "learn": True,
                    "flood": True,
                    "forward": True,
                    "unknown_unicast_flood": True
                },
                "elements": [{
                    "container": "agent_vpp_vswitch",
                    "port_label": "L2Bridge-{0}".format(bridge_index),
                    "etcd_vpp_switch_key": "agent_vpp_vswitch",
                    "type": 5
                }]
            }

            for vnf_index in range(self.vnf_count):
                new_element = {
                    "container": "vnf-vpp-{index}".format(index=vnf_index),
                    "port_label": "vnf{0}_memif{1}".format(vnf_index,
                                                           bridge_index),
                    "mac_addr": "02:01:01:01:{0}:{1}".format(
                        mac_hex(bridge_index + 1),
                        mac_hex(vnf_index + 1)),
                    "ipv4_addr": "192.168.{0}.{1}".format(
                        bridge_index + 1,
                        vnf_index + 1),
                    "l2fib_macs": [
                        "192.168.{0}.{1}".format(
                            bridge_index + 1,
                            vnf_index + 1)
                    ],
                    "type": 2,
                    "etcd_vpp_switch_key": "agent_vpp_vswitch"
                }

                entity["elements"].append(new_element)
            novpp_range = int(math.ceil(
                float(self.novpp_count) / float(self.memif_per_vnf)
            ))

            bridge_novpp_index = self.vnf_count + 1
            for novpp_index in range(
                            novpp_range * bridge_index,
                            (novpp_range * bridge_index) + novpp_range):
                new_element = {
                    "container": "novpp-{index}".format(index=novpp_index),
                    "port_label": "veth_novpp{index}".format(index=novpp_index),
                    "mac_addr": "02:01:01:01:{0}:{1}".format(
                        mac_hex(bridge_index + 1),
                        mac_hex(novpp_index + self.vnf_count + 1)),
                    "ipv4_addr": "192.168.{0}.{1}".format(
                        bridge_index + 1,
                        bridge_novpp_index),
                    "type": 3,
                    "etcd_vpp_switch_key": "agent_vpp_vswitch"
                }
                entity["elements"].append(new_element)
                bridge_novpp_index += 1

            entities_list.append(entity)

        output = ""
        for line in yaml.dump(
                entities_list,
                default_flow_style=False
        ).splitlines():
            output += " "*6 + line + "\n"

        template = self.templates["sfc"]
        if "---" in template:
            sections = template.split("---")
            for section in sections:
                if "sfc_entities:" in section:
                    output = template.replace(section, section + output)
                    self.output["sfc"] = output
        else:
            self.output["sfc"] = template + output

    def generate_vnf_config(self):
        template = self.templates["vnf"]
        output = yaml_replace_line(
            template,
            "replicas:",
            "replicas: {0}".format(self.vnf_count))
        self.output["vnf"] = output

    def generate_novpp_config(self):
        template = self.templates["novpp"]
        output = yaml_replace_line(
            template,
            "replicas:",
            "replicas: {0}".format(self.novpp_count))
        self.output["novpp"] = output

    def write_config_files(self, output_path):
        if not os.path.exists(output_path):
            os.makedirs(output_path)

        with open("{0}/sfc.yaml".format(output_path), "w") as sfc:
            sfc.write(self.output["sfc"])
        with open("{0}/vnf.yaml".format(output_path), "w") as vnf:
            vnf.write(self.output["vnf"])
        with open("{0}/novpp.yaml".format(output_path), "w") as novpp:
            novpp.write(self.output["novpp"])


def generate_config(
        vnf_count, novpp_count, memif_per_vnf, template_path, output_path):
    generator = YamlConfigGenerator(
        vnf_count,
        novpp_count,
        memif_per_vnf,
        template_path)
    generator.generate_config(output_path)
