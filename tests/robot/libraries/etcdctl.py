import json


def create_interfaces_json_from_list(interfaces):
    ints_json = ""
    for interface in interfaces:
        if interface[:4] == 'bvi_':
            ints_json += '{ "name": "' + interface + '", "bridged_virtual_interface": true },'
        else:
            ints_json += '{ "name": "' + interface + '" },'
    ints_json = ints_json[:-1]
    return ints_json


def remove_empty_lines(lines):
    out_lines = ""
    for line in lines:
        if line.strip():
            out_lines += line
    return out_lines


def remove_keys(lines):
    out_lines = ""
    for line in lines:
        if line[0] != '/':
            out_lines += line + '\n'
    return out_lines


# input - etcd dump
# output - etcd dump converted to json + key, node, name, type atributes
def convert_etcd_dump_to_json(dump):
    etcd_json = '['
    key = ''
    data = ''
    firstline = True
    for line in dump.splitlines():
        if line.strip() != '':
            if line[0] == '/':
                if not firstline:
                    etcd_json += '{"key":"'+key+'","node":"'+node+'","name":"'+name+'","type":"'+type+'","data":'+data+'},'
                key = line
                node = key.split('/')[2]
                name = key.split('/')[-1]
                type = key.split('/')[4]
                data = ''
                firstline = False
            else:
                if line == "null":
                    line = '{"error":"null"}'
                data += line
    if not firstline:
        etcd_json += '{"key":"'+key+'","node":"'+node+'","name":"'+name+'","type":"'+type+'","data":'+data+'}'
    etcd_json += ']'
    return etcd_json
