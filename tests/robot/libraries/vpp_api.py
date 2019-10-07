#!/usr/bin/env python3

# Copyright (c) 2019 Cisco and/or its affiliates.
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

import binascii
import json

from paramiko import SSHClient, AutoAddPolicy

from robot.api import logger

CLIENT_NAME = 'ligato_papi'


class vpp_api(object):
    @staticmethod
    def execute_api(host, username, password, node, command, **arguments):
        with PapiExecutor(host, username, password, node) as papi_exec:
            papi_resp = papi_exec.add(command, **arguments).get_replies()

        return papi_resp.reply


class PapiResponse(object):
    """Class for metadata specifying the Papi reply, stdout, stderr and return
    code.
    """

    def __init__(self, papi_reply=None, stdout="", stderr="", requests=None):
        """Construct the Papi response by setting the values needed.

        :param papi_reply: API reply from last executed PAPI command(s).
        :param stdout: stdout from last executed PAPI command(s).
        :param stderr: stderr from last executed PAPI command(s).
        :param requests: List of used PAPI requests. It is used while verifying
            replies. If None, expected replies must be provided for verify_reply
            and verify_replies methods.
        :type papi_reply: list or None
        :type stdout: str
        :type stderr: str
        :type requests: list
        """

        # API reply from last executed PAPI command(s).
        self.reply = papi_reply

        # stdout from last executed PAPI command(s).
        self.stdout = stdout

        # stderr from last executed PAPI command(s).
        self.stderr = stderr

        # List of used PAPI requests.
        self.requests = requests

        # List of expected PAPI replies. It is used while verifying replies.
        if self.requests:
            self.expected_replies = \
                ["{rqst}_reply".format(rqst=rqst) for rqst in self.requests]

    def __str__(self):
        """Return string with human readable description of the PapiResponse.

        :returns: Readable description.
        :rtype: str
        """
        return (
            "papi_reply={papi_reply},stdout={stdout},stderr={stderr},"
            "requests={requests}").format(
            papi_reply=self.reply, stdout=self.stdout, stderr=self.stderr,
            requests=self.requests)

    def __repr__(self):
        """Return string executable as Python constructor call.

        :returns: Executable constructor call.
        :rtype: str
        """
        return "PapiResponse({str})".format(str=str(self))


class PapiExecutor(object):
    """Contains methods for executing VPP Python API commands on DUTs.

    Note: Use only with "with" statement, e.g.:

        with PapiExecutor(node) as papi_exec:
            papi_resp = papi_exec.add('show_version').get_replies(err_msg)

    This class processes three classes of VPP PAPI methods:
    1. simple request / reply: method='request',
    2. dump functions: method='dump',
    3. vpp-stats: method='stats'.

    The recommended ways of use are (examples):

    1. Simple request / reply

    a. One request with no arguments:

        with PapiExecutor(node) as papi_exec:
            data = papi_exec.add('show_version').get_replies().\
                verify_reply()

    b. Three requests with arguments, the second and the third ones are the same
       but with different arguments.

        with PapiExecutor(node) as papi_exec:
            data = papi_exec.add(cmd1, **args1).add(cmd2, **args2).\
                add(cmd2, **args3).get_replies(err_msg).verify_replies()

    2. Dump functions

        cmd = 'sw_interface_rx_placement_dump'
        with PapiExecutor(node) as papi_exec:
            papi_resp = papi_exec.add(cmd, sw_if_index=ifc['vpp_sw_index']).\
                get_dump(err_msg)

    3. vpp-stats

        path = ['^/if', '/err/ip4-input', '/sys/node/ip4-input']

        with PapiExecutor(node) as papi_exec:
            data = papi_exec.add(api_name='vpp-stats', path=path).get_stats()

        print('RX interface core 0, sw_if_index 0:\n{0}'.\
            format(data[0]['/if/rx'][0][0]))

        or

        path_1 = ['^/if', ]
        path_2 = ['^/if', '/err/ip4-input', '/sys/node/ip4-input']

        with PapiExecutor(node) as papi_exec:
            data = papi_exec.add('vpp-stats', path=path_1).\
                add('vpp-stats', path=path_2).get_stats()

        print('RX interface core 0, sw_if_index 0:\n{0}'.\
            format(data[1]['/if/rx'][0][0]))

        Note: In this case, when PapiExecutor method 'add' is used:
        - its parameter 'csit_papi_command' is used only to keep information
          that vpp-stats are requested. It is not further processed but it is
          included in the PAPI history this way:
          vpp-stats(path=['^/if', '/err/ip4-input', '/sys/node/ip4-input'])
          Always use csit_papi_command="vpp-stats" if the VPP PAPI method
          is "stats".
        - the second parameter must be 'path' as it is used by PapiExecutor
          method 'add'.
    """

    def __init__(self, host, username, password, node):
        """Initialization.
        """

        # Node to run command(s) on.
        self.host = host
        self.node = node
        self.username = username
        self.password = password

        self._ssh = SSHClient()
        self._ssh.set_missing_host_key_policy(AutoAddPolicy())

        # The list of PAPI commands to be executed on the node.
        self._api_command_list = list()

    def __enter__(self):
        try:
            self._ssh.connect(self.host, username=self.username, password=self.password)
        except IOError:
            raise RuntimeError("Cannot open SSH connection to host {host} to "
                               "execute PAPI command(s)".
                               format(host=self.host))
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        self._ssh.close()

    def add(self, csit_papi_command="vpp-stats", **kwargs):
        """Add next command to internal command list; return self.

        The argument name 'csit_papi_command' must be unique enough as it cannot
        be repeated in kwargs.

        :param csit_papi_command: VPP API command.
        :param kwargs: Optional key-value arguments.
        :type csit_papi_command: str
        :type kwargs: dict
        :returns: self, so that method chaining is possible.
        :rtype: PapiExecutor
        """
        self._api_command_list.append(dict(api_name=csit_papi_command,
                                           api_args=kwargs))
        return self

    def get_replies(self,
                    process_reply=True, ignore_errors=False, timeout=120):
        """Get reply/replies from VPP Python API.

        :param process_reply: Process PAPI reply if True.
        :param ignore_errors: If true, the errors in the reply are ignored.
        :param timeout: Timeout in seconds.
        :type process_reply: bool
        :type ignore_errors: bool
        :type timeout: int
        :returns: Papi response including: papi reply, stdout, stderr and
            return code.
        :rtype: PapiResponse
        """
        return self._execute(
            method='request', process_reply=process_reply,
            ignore_errors=ignore_errors, timeout=timeout)

    @staticmethod
    def _process_api_data(api_d):
        """Process API data for smooth converting to JSON string.

        Apply binascii.hexlify() method for string values.

        :param api_d: List of APIs with their arguments.
        :type api_d: list
        :returns: List of APIs with arguments pre-processed for JSON.
        :rtype: list
        """

        def process_value(val):
            """Process value.

            :param val: Value to be processed.
            :type val: object
            :returns: Processed value.
            :rtype: dict or str or int
            """
            if isinstance(val, dict):
                val_dict = dict()
                for val_k, val_v in val.items():
                    val_dict[str(val_k)] = process_value(val_v)
                return val_dict
            else:
                return binascii.hexlify(val) if isinstance(val, str) else val

        api_data_processed = list()
        for api in api_d:
            api_args_processed = dict()
            for a_k, a_v in api["api_args"].iteritems():
                api_args_processed[str(a_k)] = process_value(a_v)
            api_data_processed.append(dict(api_name=api["api_name"],
                                           api_args=api_args_processed))
        return api_data_processed

    @staticmethod
    def _revert_api_reply(api_r):
        """Process API reply / a part of API reply.

        Apply binascii.unhexlify() method for unicode values.

        :param api_r: API reply.
        :type api_r: dict
        :returns: Processed API reply / a part of API reply.
        :rtype: dict
        """
        reply_dict = dict()
        reply_value = dict()
        for reply_key, reply_v in api_r.items():
            for a_k, a_v in reply_v.iteritems():
                reply_value[a_k] = binascii.unhexlify(a_v) \
                    if isinstance(a_v, str) else a_v
            reply_dict[reply_key] = reply_value
        return reply_dict

    def _process_reply(self, api_reply):
        """Process API reply.

        :param api_reply: API reply.
        :type api_reply: dict or list of dict
        :returns: Processed API reply.
        :rtype: list or dict
        """
        if isinstance(api_reply, list):
            reverted_reply = [self._revert_api_reply(a_r) for a_r in api_reply]
        else:
            reverted_reply = self._revert_api_reply(api_reply)
        return reverted_reply

    def _execute_papi(self, api_data, method='request', timeout=120):
        """Execute PAPI command(s) on remote node and store the result.

        :param api_data: List of APIs with their arguments.
        :param method: VPP Python API method. Supported methods are: 'request',
            'dump' and 'stats'.
        :param timeout: Timeout in seconds.
        :type api_data: list
        :type method: str
        :type timeout: int
        :returns: Stdout and stderr.
        :rtype: 2-tuple of str
        :raises SSHTimeout: If PAPI command(s) execution has timed out.
        :raises RuntimeError: If PAPI executor failed due to another reason.
        :raises AssertionError: If PAPI command(s) execution has failed.
        """

        if not api_data:
            RuntimeError("No API data provided.")

        json_data = json.dumps(api_data) \
            if method in ("stats", "stats_request") \
            else json.dumps(self._process_api_data(api_data))

        cmd = "docker exec {node} python3 {fw_dir}/{papi_provider} --data '{json}'". \
            format(node=self.node,
                   fw_dir="/opt",
                   papi_provider="vpp_api_executor.py",
                   json=json_data)
        logger.debug(cmd)
        stdin, stdout, stderr = self._ssh.exec_command(
            cmd, timeout=timeout)
        stdout = stdout.read()
        stderr = stderr.read()
        return stdout, stderr

    def _execute(self, method='request', process_reply=True,
                 ignore_errors=False, timeout=120):
        """Turn internal command list into proper data and execute; return
        PAPI response.

        This method also clears the internal command list.

        IMPORTANT!
        Do not use this method in L1 keywords. Use:
        - get_stats()
        - get_replies()
        - get_dump()

        :param method: VPP Python API method. Supported methods are: 'request',
            'dump' and 'stats'.
        :param process_reply: Process PAPI reply if True.
        :param ignore_errors: If true, the errors in the reply are ignored.
        :param timeout: Timeout in seconds.
        :type method: str
        :type process_reply: bool
        :type ignore_errors: bool
        :type timeout: int
        :returns: Papi response including: papi reply, stdout, stderr and
            return code.
        :rtype: PapiResponse
        :raises KeyError: If the reply is not correct.
        """

        local_list = self._api_command_list

        # Clear first as execution may fail.
        self._api_command_list = list()

        stdout, stderr = self._execute_papi(
            local_list, method=method, timeout=timeout)
        papi_reply = list()
        if process_reply:
            try:
                json_data = json.loads(stdout)
            except ValueError:
                logger.error(
                    "An error occured while processing the PAPI reply:\n"
                    "stdout: {stdout}\n"
                    "stderr: {stderr}".format(stdout=stdout, stderr=stderr))
                raise
            for data in json_data:
                try:
                    api_reply_processed = dict(
                        api_name=data["api_name"],
                        api_reply=self._process_reply(data["api_reply"]))
                except KeyError:
                    if ignore_errors:
                        continue
                    else:
                        raise
                papi_reply.append(api_reply_processed)

        # Log processed papi reply to be able to check API replies changes
        logger.debug("Processed PAPI reply: {reply}".format(reply=papi_reply))

        return PapiResponse(
            papi_reply=papi_reply, stdout=stdout, stderr=stderr,
            requests=[rqst["api_name"] for rqst in local_list])
