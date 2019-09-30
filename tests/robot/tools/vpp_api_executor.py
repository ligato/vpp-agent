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

import os
import sys
import json
import fnmatch
import argparse
import binascii
import ipaddress

from vpp_papi import VPP

vpp_json_dir_core = "/usr/share/vpp/api/core"
vpp_json_dir_plugins = "/usr/share/vpp/api/plugins"


def _convert_reply(api_r):
    """Process API reply / a part of API reply for smooth converting to
    JSON string.

    It is used only with 'request' and 'dump' methods.

    Apply binascii.hexlify() method for string values.

    TODO: Implement complex solution to process of replies.

    :param api_r: API reply.
    :type api_r: Vpp_serializer reply object (named tuple)
    :returns: Processed API reply / a part of API reply.
    :rtype: dict
    """
    unwanted_fields = ['count', 'index', 'context']

    def process_value(val):
        """Process value.

        :param val: Value to be processed.
        :type val: object
        :returns: Processed value.
        :rtype: dict or str or int
        """

        # with dict or list just recursively iterate through all elements
        if isinstance(val, dict):
            for val_k, val_v in val.items():
                val[str(val_k)] = process_value(val_v)
            return val
        elif isinstance(val, list):
            for idx, val_l in enumerate(val):
                val[idx] = process_value(val_l)
            return val
        # no processing for int
        elif hasattr(val, '__int__'):
            return int(val)
        elif isinstance(val, bytes):
            # if exactly 16 bytes it's probably an IP address
            if len(val) == 16:
                try:
                    # without context we don't know if it's IPv4 or IPv6, return both forms
                    ipv4 = ipaddress.IPv4Address(val[:4])
                    ipv6 = ipaddress.IPv6Address(val)
                    return {"ipv4": str(ipv4), "ipv6": str(ipv6)}
                except ipaddress.AddressValueError:
                    # maybe it's not an IP address after all
                    pass
            elif len(val) in (6, 8):
                # Probably a padded MAC address(8) or "Dmac, Smac, etc."??(6)
                return val.hex()

            # strip null byte padding from some fields, such as tag or name
            while val.endswith(b"\x00"):
                val = val[:-1]
            return str(val, "ascii")
        elif hasattr(val, '__str__'):

            if "(" in repr(val):
                # it's another vpp-internal object
                item_dict = dict()
                for item in dir(val):
                    if not item.startswith("_") and item not in unwanted_fields:
                        item_dict[item] = process_value(getattr(val, item))
                return item_dict
            else:
                # just a simple string
                return str(val)
        # Next handles parameters not supporting preferred integer or string
        # representation to get it logged
        elif hasattr(val, '__repr__'):
            return repr(val)
        else:
            return val

    reply_dict = dict()
    reply_key = repr(api_r).split('(')[0]
    reply_value = dict()
    for item in dir(api_r):
        if not item.startswith('_') and item not in unwanted_fields:
            reply_value[item] = process_value(getattr(api_r, item))
    reply_dict[reply_key] = reply_value
    return reply_dict


class VppApi(object):
    def __init__(self):
        self.vpp = None

        jsonfiles = []
        for root, dirnames, filenames in os.walk(vpp_json_dir_core):
            for filename in fnmatch.filter(filenames, '*.api.json'):
                jsonfiles.append(os.path.join(vpp_json_dir_core, filename))
        for root, dirnames, filenames in os.walk(vpp_json_dir_plugins):
            for filename in fnmatch.filter(filenames, '*.api.json'):
                jsonfiles.append(os.path.join(vpp_json_dir_plugins, filename))

        self.vpp = VPP(jsonfiles)

    def connect(self):
        resp = self.vpp.connect("ligato-test-api")
        if resp != 0:
            raise RuntimeError("VPP papi connection failed.")

    def list_capabilities(self):
        print(dir(self.vpp.api))

    def disconnect(self):
        resp = self.vpp.disconnect()
        if resp != 0:
            print("Warning: VPP papi disconnect failed.")

    def show_version(self):
        print(self.vpp.api.show_version())

    def process_json_request(self, args):
        """Process the request/reply and dump classes of VPP API methods.

        :param args: Command line arguments passed to VPP PAPI Provider.
        :type args: ArgumentParser
        :returns: JSON formatted string.
        :rtype: str
        :raises RuntimeError: If PAPI command error occurs.
        """

        vpp = self.vpp

        reply = list()

        def process_value(val):
            """Process value.

            :param val: Value to be processed.
            :type val: object
            :returns: Processed value.
            :rtype: dict or str or int
            """
            if isinstance(val, dict):
                for val_k, val_v in val.items():
                    val[str(val_k)] = process_value(val_v)
                return val
            elif isinstance(val, list):
                for idx, val_l in enumerate(val):
                    val[idx] = process_value(val_l)
                return val
            elif isinstance(val, str):
                return binascii.unhexlify(val)
            elif isinstance(val, int):
                return val
            else:
                return str(val)

        self.connect()
        json_data = json.loads(args.data)

        for data in json_data:
            api_name = data['api_name']
            api_args_unicode = data['api_args']
            api_reply = dict(api_name=api_name)
            api_args = dict()
            for a_k, a_v in api_args_unicode.items():
                api_args[str(a_k)] = process_value(a_v)
            try:
                papi_fn = getattr(vpp.api, api_name)
                rep = papi_fn(**api_args)

                if isinstance(rep, list):
                    converted_reply = list()
                    for r in rep:
                        converted_reply.append(_convert_reply(r))
                else:
                    converted_reply = _convert_reply(rep)

                api_reply['api_reply'] = converted_reply
                reply.append(api_reply)
            except (AttributeError, ValueError) as err:
                vpp.disconnect()
                raise RuntimeError('PAPI command {api}({args}) input error:\n{err}'.
                                   format(api=api_name,
                                          args=api_args,
                                          err=repr(err)))
            except Exception as err:
                vpp.disconnect()
                raise RuntimeError('PAPI command {api}({args}) error:\n{exc}'.
                                   format(api=api_name,
                                          args=api_args,
                                          exc=repr(err)))
        self.disconnect()

        return json.dumps(reply)


def main():
    """Main function for the Python API provider.
    """

    # The functions which process different types of VPP Python API methods.

    parser = argparse.ArgumentParser(
        formatter_class=argparse.RawDescriptionHelpFormatter,
        description=__doc__)
    parser.add_argument("-d", "--data",
                        required=True,
                        help="Data is a JSON string (list) containing API name(s)"
                             "and its/their input argument(s).")

    args = parser.parse_args()

    vpp = VppApi()
    return VppApi.process_json_request(vpp, args)


if __name__ == '__main__':
    sys.stdout.write(main())
    sys.stdout.flush()
    sys.exit(0)
