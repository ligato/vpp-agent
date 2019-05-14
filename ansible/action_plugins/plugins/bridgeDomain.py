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

import etcd3
from google.protobuf.json_format import MessageToJson, Parse

from action_plugins.pout.models.vpp.l2.bridge_domain_pb2 import BridgeDomain


def plugin_init(name, values, agent_name, ip, port):
    if name == 'bridge-domain':
        return BridgeDomainValidation(values, agent_name)
    elif name == 'add-bridge-domain-interface':
        return AddBridgeDomainInterfaceValidation(values, agent_name, ip, port)
    elif name == 'remove-bridge-domain-interface':
        return RemoveBridgeDomainInterfaceValidation(values, agent_name, ip, port)
    else:
        return False


class BridgeDomainValidation:

    def __init__(self, values, agent_name):
        self.values = values
        self.agent_name = agent_name

    def validate(self):
        bridgeDomain = BridgeDomain()
        Parse(json.dumps(self.values), bridgeDomain)
        return MessageToJson(bridgeDomain, indent=None)

    def create_key(self):
        return "/vnf-agent/{}/config/vpp/l2/v2/bridge-domain/{}".format(self.agent_name, self.values['name'])


class AddBridgeDomainInterfaceValidation:

    def __init__(self, values, agent_name, ip, port):
        self.values = values
        self.agent_name = agent_name
        host = ip
        port = port
        self.client = etcd3.client(host, port)

    def validate(self):
        etcd_values = self.client.get(self.create_key())
        val = {}
        if etcd_values[0] is None:
            val['interfaces'] = []
        else:
            val = json.loads(etcd_values[0])

        if val.get('interfaces') is None:
            val['interfaces'] = []
        val['interfaces'] += self.values['interfaces']

        bridgeDomain = BridgeDomain()
        Parse(json.dumps(val), bridgeDomain)
        return MessageToJson(bridgeDomain, indent=None)

    def create_key(self):
        return "/vnf-agent/{}/config/vpp/l2/v2/bridge-domain/{}".format(self.agent_name, self.values['name'])


class RemoveBridgeDomainInterfaceValidation:

    def __init__(self, values, agent_name, ip, port):
        self.values = values
        self.agent_name = agent_name
        host = ip
        port = int(port)
        self.client = etcd3.client(host, port)

    def validate(self):
        etcd_values = self.client.get(self.create_key())
        val = json.loads(etcd_values[0])
        try:
            val['interfaces'].remove(self.values['interfaces'][0])
        except:
            pass

        bridgeDomain = BridgeDomain()
        Parse(json.dumps(val), bridgeDomain)
        return MessageToJson(bridgeDomain, indent=None)

    def create_key(self):
        return "/vnf-agent/{}/config/vpp/l2/v2/bridge-domain/{}".format(self.agent_name, self.values['name'])
