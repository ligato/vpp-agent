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
from enum import Enum

from google.protobuf.json_format import MessageToJson, Parse

from action_plugins.pout.models.vpp.l3.route_pb2 import Route


def plugin_init(name, values, agent_name, ip, port):
    if name == 'route':
        return RouteValidation(values, agent_name)
    else:
        return False


class RouteValidation:

    def __init__(self, values, agent_name):
        self.values = values
        self.agent_name = agent_name

    def validate(self):
        route = Route()
        Parse(json.dumps(self.values), route)
        return MessageToJson(route, indent=None)

    def create_key(self):
        return "/vnf-agent/{}/config/vpp/v2/route/vrf/{}/dst/{}/gw/{}".format(self.agent_name,
                                                                              self.values.get('name', 0),
                                                                              self.values['dst_network'],
                                                                              self.values.get('next_hop_addr', ''))
