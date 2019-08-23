# Copyright (c) 2019 PANTHEON.tech
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at:
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import json
import ipaddress
from google.protobuf.json_format import MessageToJson, Parse

from action_plugins.pout.models.vpp.l3.l3_pb2 import DHCPProxy


def plugin_init(name, values, agent_name, ip, port):
    if name == 'dhcp-proxy':
        return DHCPProxyValidation(values, agent_name)
    else:
        return False


class DHCPProxyValidation:

    def __init__(self, values, agent_name):
        self.values = values
        self.agent_name =agent_name
        version = ipaddress.ip_address(self.values['source_ip_address']).version
        self.protocol = 'IPv{}'.format(version)

    def validate(self):
        dhcp_proxy = DHCPProxy()
        Parse(json.dumps(self.values), dhcp_proxy)
        return MessageToJson(dhcp_proxy, preserving_proto_field_name=True, indent=None)

    def create_key(self):
        return "/vnf-agent/{}/config/vpp/v2/dhcp-proxy/{}".format(self.agent_name,
                                                               self.protocol)
