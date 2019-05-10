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

import os
import sys

basedir = os.path.split(sys.modules['ansible.plugins.action.vpp_etcd'].__file__)[0]
sys.path = ['/'.join(basedir.split('/')[:-1]), '/'.join(basedir.split('/')[:-2])] + sys.path

from ansible.plugins.action import ActionBase


class ActionModule(ActionBase):

    def run(self, tmp=None, task_vars=None):

        if task_vars is None:
            task_vars = dict()
        plugin = None
        result = super(ActionModule, self).run(tmp, task_vars)

        args = self._task.args.copy()
        plugin_name = args.get('value_type')

        plugindir = basedir + '/plugins'
        poutdir = basedir + '/pout'
        syspath = sys.path
        sys.path = [basedir, plugindir, poutdir] + syspath

        fnames = os.listdir(plugindir)
        values = args.get('value')
        agent_name = args.get('agent_name', task_vars.get('agent'))

        if agent_name is None:
            return {'failed': True, 'msg': 'agent_name must be defined'}
        for fname in fnames:
            if not fname.startswith(".#") and fname.endswith(".py") and \
                    not fname.startswith('__'):
                pluginmod = __import__(fname[:-3])
                try:
                    plugin = pluginmod.plugin_init(plugin_name, values, agent_name, task_vars.get('bridge_connection',
                                                                                                  '172.0.0.1'),
                                                   task_vars.get('etcd_port', 12379))
                    if plugin:
                        break
                except AttributeError as s:
                    print(pluginmod.__dict__)
                    raise AttributeError(pluginmod.__file__ + ': ' + str(s))
        sys.path = syspath

        new_args = dict()
        new_args['state'] = args.get('state')

        values = plugin.validate()

        new_args['host'] = task_vars.get('bridge_connection', '172.0.0.1')    # bridged connection for awx otherwise "172.0.0.1"
        new_args['port'] = int(task_vars.get('etcd_port', 12379))
        new_args['key'] = plugin.create_key()

        new_args['value'] = values
        if args.get('secure_transport', task_vars.get('secureTransport')):
            new_args['ca_cert'] = "/tmp/certificates/ca.pem"
            new_args['client_cert'] = "/tmp/certificates/client.pem"
            new_args['client_key'] = "/tmp/certificates/client-key.pem"

        result.update(self._execute_module(module_name='etcd3', module_args=new_args, task_vars=task_vars))
        return result
