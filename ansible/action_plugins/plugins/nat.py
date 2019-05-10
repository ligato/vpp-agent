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

from google.protobuf.json_format import MessageToJson, Parse

from action_plugins.pout.models.vpp.nat.nat_pb2 import Nat44Global


def plugin_init(name, values, agent_name, ip, port):
    if name == 'nat':
        return NatValidation(values, agent_name)
    else:
        return False


class NatValidation:

    def __init__(self, values, agent_name):
        self.values = values
        self.agent_name =agent_name

    def validate(self):
        nat = Nat44Global()
        Parse(json.dumps(self.values), nat)
        return MessageToJson(nat, indent=None)

    def create_key(self):
        return "/vnf-agent/{}/config/vpp/nat/v2/nat44-global".format(self.agent_name)
