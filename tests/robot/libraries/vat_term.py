import json

# input - json output from vxlan_tunnel_dump, src ip, dst ip, vni
# output - true if tunnel exists, false if not, interface index
def Check_VXLan_Tunnel_Presence(out, src, dst, vni):
    out =  out[out.find('['):out.rfind(']')+1]
    data = json.loads(out)
    present = False
    if_index = -1
    for iface in data:
        if iface["src_address"] == src and iface["dst_address"] == dst and iface["vni"] == int(vni):
            present = True
            if_index  = iface["sw_if_index"]
    return present, if_index

# input - json output from sw_interface_dump, index
# output - interface name
def Get_Interface_Name(out, index):
    out =  out[out.find('['):out.rfind(']')+1]
    data = json.loads(out)
    name = "x"
    for iface in data:
        if iface["sw_if_index"] == int(index):
            name = iface["interface_name"]
    return name

# input - json output from sw_interface_dump, interface name
# output - index
def Get_Interface_Index(out, name):
    out =  out[out.find('['):out.rfind(']')+1]
    data = json.loads(out)
    index = -1
    for iface in data:
        if iface["interface_name"] == name:
            index = iface["sw_if_index"]
    return index

# input - json output from sw_interface_dump, index
# output - whole interface state
def Get_Interface_State(out, index):
    out =  out[out.find('['):out.rfind(']')+1]
    data = json.loads(out)
    state = -1
    for iface in data:
        if iface["sw_if_index"] == int(index):
            state = iface
    return state

# input - mac in dec from sw_interface_dump
# output - regular mac in hex
def Convert_Dec_MAC_To_Hex(mac):
    hexmac=[]
    for num in mac[:6]:
        hexmac.append("%02x" % num)
    hexmac = ":".join(hexmac)
    return hexmac

# input - output from show memif intf command
# output - state info list
def Parse_Memif_Info(info):
    state = []
    for line in info.splitlines():
        if (line.strip().split()[0] == "flags"):
            if "admin-up" in line:
                state.append("enabled=1")
            if "slave" in line:
                state.append("role=slave")
            if "connected" in line:
                state.append("connected=1")
        if (line.strip().split()[0] == "id"):
            state.append("id="+line.strip().split()[1])
            state.append("socket="+line.strip().split()[-1])
    if "enabled=1" not in state:
        state.append("enabled=0")
    if "role=slave" not in state:
        state.append("role=master")
    if "connected=1" not in state:
        state.append("connected=0")
    return state

# input - output from show br br_id detail command
# output - state info list
def Parse_BD_Details(details):
    state = []
    line = details.splitlines()[1]
    if (line.strip().split()[6]) == "on":
        state.append("uuflood=1")
    else:
        state.append("uuflood=0")
    if (line.strip().split()[8]) == "on":
        state.append("arp_term=1")
    else:
        state.append("arp_term=0")
    return state

# input - etcd dump
# output - etcd dump converted to json + key, node, name, type atributes
def Convert_ETCD_Dump_To_JSON(dump):
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
                data += line 
    if not firstline:
        etcd_json += '{"key":"'+key+'","node":"'+node+'","name":"'+name+'","type":"'+type+'","data":'+data+'}'
    etcd_json += ']'
    return etcd_json

def Parse_BD_Interfaces(node, bd, etcd_json, bd_dump):
    interfaces = []
    bd_dump = json.loads(bd_dump)
    etcd_json = json.loads(etcd_json)
    for int in bd_dump[0]["sw_if"]:
        bd_sw_if_index =  int["sw_if_index"]
        etcd_name = "none"
        for key_data in etcd_json:
            if key_data["node"] and key_data["type"] == "status" and "/interface/" in key_data["key"]:
                if "if_index" in key_data["data"]:
                    if key_data["data"]["if_index"] == bd_sw_if_index:
                        etcd_name = key_data["data"]["name"]
        interfaces.append("interface="+etcd_name)
    if bd_dump[0]["bvi_sw_if_index"] != 4294967295:
        bvi_sw_if_index = bd_dump[0]["bvi_sw_if_index"]
        etcd_name = "none"
        for key_data in etcd_json:
            if key_data["node"] and key_data["type"] == "status" and "/interface/" in key_data["key"]:
                if "if_index" in key_data["data"]:
                    if key_data["data"]["if_index"] == bvi_sw_if_index:
                        etcd_name = key_data["data"]["name"]
        interfaces.append("bvi_int="+etcd_name)
    return interfaces

def Check_BD_Presence(bd_dump, indexes):
    bd_dump = json.loads(bd_dump)
    present = True
    for bd in bd_dump:
        for index in indexes:
            int_present = False
            for bd_int in bd["sw_if"]:
                if bd_int["sw_if_index"] == index:
                    int_present = True
            if int_present == False:
                present = False
    return present


x='''[ {
    "bd_id": 2,
    "flood": 1,
    "forward": 1,
    "learn": 1,
    "bvi_sw_if_index": 4,
    "n_sw_ifs": 2,
    "sw_if": [ 
      {
        "sw_if_index": 4,
        "shg": 0
      },
      {
        "sw_if_index": 5,
        "shg": 0
      }
    ]
  }
]
'''

y='''[{"key":"/vnf-agent/agent_vpp_1/check/status/v1/agent","node":"agent_vpp_1","name":"agent","type":"status","data":{"build_version":"644272b05b52f73057b2f838ab504399057df132","build_date":"2017-09-06T08:40+00:00","state":1,"start_time":1504878512,"last_change":1504878520,"last_update":1504878665}},{"key":"/vnf-agent/agent_vpp_1/check/status/v1/plugin/etcdv3","node":"agent_vpp_1","name":"etcdv3","type":"status","data":{"state":1,"last_change":1504878520,"last_update":1504878665}},{"key":"/vnf-agent/agent_vpp_1/check/status/v1/plugin/govpp","node":"agent_vpp_1","name":"govpp","type":"status","data":{"state":1,"last_change":1504878515,"last_update":1504878665}},{"key":"/vnf-agent/agent_vpp_1/check/status/v1/plugin/kafka","node":"agent_vpp_1","name":"kafka","type":"status","data":{"state":1,"last_change":1504878520,"last_update":1504878665}},{"key":"/vnf-agent/agent_vpp_1/linux/config/v1/interface/vpp1_veth1","node":"agent_vpp_1","name":"vpp1_veth1","type":"config","data":{"name":"vpp1_veth1","type":0,"enabled":true,"phys_address":"12:11:11:11:11:11","veth":{"peer_if_name":"vpp1_veth2"},"mtu":1500,"ip_addresses":["10.10.1.1/24"]}},{"key":"/vnf-agent/agent_vpp_1/linux/config/v1/interface/vpp1_veth2","node":"agent_vpp_1","name":"vpp1_veth2","type":"config","data":{"name":"vpp1_veth2","type":0,"enabled":true,"phys_address":"12:12:12:12:12:12","veth":{"peer_if_name":"vpp1_veth1"}}},{"key":"/vnf-agent/agent_vpp_1/vpp/config/v1/bd/vpp1_bd1","node":"agent_vpp_1","name":"vpp1_bd1","type":"config","data":{"name":"vpp1_bd1","flood":true,"unknown_unicast_flood":true,"forward":true,"learn":true,"arp_termination":true,"interfaces":[{"name":"vpp1_vxlan1"},{"name":"vpp1_afpacket1"}]}},{"key":"/vnf-agent/agent_vpp_1/vpp/config/v1/interface/vpp1_afpacket1","node":"agent_vpp_1","name":"vpp1_afpacket1","type":"config","data":{"name":"vpp1_afpacket1","type":4,"enabled":true,"phys_address":"a2:a1:a1:a1:a1:a1","afpacket":{"host_if_name":"vpp1_veth2"}}},{"key":"/vnf-agent/agent_vpp_1/vpp/config/v1/interface/vpp1_loop1","node":"agent_vpp_1","name":"vpp1_loop1","type":"config","data":{"name":"vpp1_loop1","type":0,"enabled":true,"phys_address":"12:21:21:11:11:11","mtu":1500,"ip_addresses":["20.20.1.1/24"]}},{"key":"/vnf-agent/agent_vpp_1/vpp/config/v1/interface/vpp1_memif1","node":"agent_vpp_1","name":"vpp1_memif1","type":"config","data":{"name":"vpp1_memif1","type":2,"enabled":true,"phys_address":"62:61:61:61:61:61","memif":{"master":true,"id":1,"socket_filename":"/tmp/default.sock"},"mtu":1500,"ip_addresses":["192.168.1.1/24"]}},{"key":"/vnf-agent/agent_vpp_1/vpp/config/v1/interface/vpp1_tap1","node":"agent_vpp_1","name":"vpp1_tap1","type":"config","data":{"name":"vpp1_tap1","type":3,"enabled":true,"phys_address":"32:21:21:11:11:11","mtu":1500,"ip_addresses":["30.30.1.1/24"],"tap":{"host_if_name":"linux_vpp1_tap1"}}},{"key":"/vnf-agent/agent_vpp_1/vpp/config/v1/interface/vpp1_vxlan1","node":"agent_vpp_1","name":"vpp1_vxlan1","type":"config","data":{"name":"vpp1_vxlan1","type":5,"enabled":true,"vxlan":{"src_address":"192.168.1.1","dst_address":"192.168.1.2","vni":5}}},{"key":"/vnf-agent/agent_vpp_1/vpp/status/v1/bd/vpp1_bd1","node":"agent_vpp_1","name":"vpp1_bd1","type":"status","data":{"index":1,"internal_name":"vpp1_bd1","bvi_interface":"\u003cnot_set\u003e","interface_count":2,"last_change":1504878585,"l2_params":{"flood":true,"unknown_unicast_flood":true,"forward":true,"learn":true,"arp_termination":true},"interfaces":[{"name":"vpp1_afpacket1","sw_if_index":2},{"name":"vpp1_vxlan1","sw_if_index":3}]}},{"key":"/vnf-agent/agent_vpp_1/vpp/status/v1/interface/local0","node":"agent_vpp_1","name":"local0","type":"status","data":{"name":"local0","internal_name":"local0","admin_status":2,"oper_status":2,"statistics":{}}},{"key":"/vnf-agent/agent_vpp_1/vpp/status/v1/interface/vpp1_afpacket1","node":"agent_vpp_1","name":"vpp1_afpacket1","type":"status","data":{"name":"vpp1_afpacket1","internal_name":"host-vpp1_veth2","if_index":2,"admin_status":1,"oper_status":1,"last_change":1504878579,"phys_address":"a2:a1:a1:a1:a1:a1","mtu":9216,"statistics":{"in_packets":4,"in_bytes":280,"drop_packets":4,"ipv6_packets":2}}},{"key":"/vnf-agent/agent_vpp_1/vpp/status/v1/interface/vpp1_loop1","node":"agent_vpp_1","name":"vpp1_loop1","type":"status","data":{"name":"vpp1_loop1","internal_name":"loop0","if_index":4,"admin_status":1,"oper_status":1,"last_change":1504878588,"phys_address":"12:21:21:11:11:11","mtu":9216,"statistics":{}}},{"key":"/vnf-agent/agent_vpp_1/vpp/status/v1/interface/vpp1_memif1","node":"agent_vpp_1","name":"vpp1_memif1","type":"status","data":{"name":"vpp1_memif1","internal_name":"memif0/1","if_index":1,"admin_status":1,"oper_status":2,"last_change":1504878570,"phys_address":"62:61:61:61:61:61","mtu":9216,"statistics":{"drop_packets":3,"out_error_packets":3}}},{"key":"/vnf-agent/agent_vpp_1/vpp/status/v1/interface/vpp1_tap1","node":"agent_vpp_1","name":"vpp1_tap1","type":"status","data":{"name":"vpp1_tap1","internal_name":"tap-0","if_index":5,"admin_status":1,"oper_status":1,"last_change":1504878591,"phys_address":"32:21:21:11:11:11","speed":1000000000,"mtu":9216,"statistics":{"in_packets":6,"in_bytes":480,"drop_packets":8,"ipv6_packets":6}}},{"key":"/vnf-agent/agent_vpp_1/vpp/status/v1/interface/vpp1_vxlan1","node":"agent_vpp_1","name":"vpp1_vxlan1","type":"status","data":{"name":"vpp1_vxlan1","internal_name":"vxlan_tunnel0","if_index":3,"admin_status":1,"oper_status":1,"last_change":1504878582,"statistics":{"out_packets":2,"out_bytes":212}}},{"key":"/vnf-agent/agent_vpp_2/check/status/v1/agent","node":"agent_vpp_2","name":"agent","type":"status","data":{"build_version":"644272b05b52f73057b2f838ab504399057df132","build_date":"2017-09-06T08:40+00:00","state":1,"start_time":1504878540,"last_change":1504878545,"last_update":1504878660}},{"key":"/vnf-agent/agent_vpp_2/check/status/v1/plugin/etcdv3","node":"agent_vpp_2","name":"etcdv3","type":"status","data":{"state":1,"last_change":1504878545,"last_update":1504878660}},{"key":"/vnf-agent/agent_vpp_2/check/status/v1/plugin/govpp","node":"agent_vpp_2","name":"govpp","type":"status","data":{"state":1,"last_change":1504878540,"last_update":1504878660}},{"key":"/vnf-agent/agent_vpp_2/check/status/v1/plugin/kafka","node":"agent_vpp_2","name":"kafka","type":"status","data":{"state":1,"last_change":1504878545,"last_update":1504878660}},{"key":"/vnf-agent/agent_vpp_2/vpp/status/v1/interface/local0","node":"agent_vpp_2","name":"local0","type":"status","data":{"name":"local0","statistics":{}}}]'''

z='''
/vnf-agent/agent_vpp_1/check/status/v1/agent
{"build_version":"644272b05b52f73057b2f838ab504399057df132","build_date":"2017-09-06T08:40+00:00","state":1,"start_time":1504878512,"last_change":1504878520,"last_update":1504878665}
/vnf-agent/agent_vpp_1/check/status/v1/plugin/etcdv3
{"state":1,"last_change":1504878520,"last_update":1504878665}
/vnf-agent/agent_vpp_1/check/status/v1/plugin/govpp
{"state":1,"last_change":1504878515,"last_update":1504878665}
/vnf-agent/agent_vpp_1/check/status/v1/plugin/kafka
{"state":1,"last_change":1504878520,"last_update":1504878665}
/vnf-agent/agent_vpp_1/linux/config/v1/interface/vpp1_veth1
{
  "name": "vpp1_veth1",
  "type": 0,
  "enabled": true,
  "phys_address": "12:11:11:11:11:11",
  "veth": {
    "peer_if_name": "vpp1_veth2"
  },
  "mtu": 1500,
  "ip_addresses": [
    "10.10.1.1/24"
  ]
}


/vnf-agent/agent_vpp_1/linux/config/v1/interface/vpp1_veth2
{
  "name": "vpp1_veth2",
  "type": 0,
  "enabled": true,
  "phys_address": "12:12:12:12:12:12",
  "veth": {
    "peer_if_name": "vpp1_veth1"
  }
}


/vnf-agent/agent_vpp_1/vpp/config/v1/bd/vpp1_bd1
{
  "name": "vpp1_bd1",
  "flood": true,
  "unknown_unicast_flood": true,
  "forward": true,
  "learn": true,
  "arp_termination": true,
  "interfaces": [
    { "name": "vpp1_vxlan1" },{ "name": "vpp1_afpacket1" }
  ]
}


/vnf-agent/agent_vpp_1/vpp/config/v1/interface/vpp1_afpacket1
{
  "name": "vpp1_afpacket1",
  "type": 4,
  "enabled": true,
  "phys_address": "a2:a1:a1:a1:a1:a1",
  "afpacket": {
    "host_if_name": "vpp1_veth2"
  }
}


/vnf-agent/agent_vpp_1/vpp/config/v1/interface/vpp1_loop1
{
  "name": "vpp1_loop1",
  "type": 0,
  "enabled": true,
  "phys_address": "12:21:21:11:11:11",
  "mtu": 1500,
  "ip_addresses": [
    "20.20.1.1/24"
  ]
}


/vnf-agent/agent_vpp_1/vpp/config/v1/interface/vpp1_memif1
{
  "name": "vpp1_memif1",
  "type": 2,
  "enabled": true,
  "phys_address": "62:61:61:61:61:61",
  "memif": {
    "master": true,
    "id": 1,
    "socket_filename": "/tmp/default.sock"
  },
  "mtu": 1500,
  "ip_addresses": [
    "192.168.1.1/24"
  ]
}



/vnf-agent/agent_vpp_1/vpp/config/v1/interface/vpp1_tap1
{
  "name": "vpp1_tap1",
  "type": 3,
  "enabled": true,
  "phys_address": "32:21:21:11:11:11",
  "mtu": 1500,
  "ip_addresses": [
    "30.30.1.1/24"
  ],
  "tap": {
    "host_if_name": "linux_vpp1_tap1"
  }
}


/vnf-agent/agent_vpp_1/vpp/config/v1/interface/vpp1_vxlan1
{
  "name": "vpp1_vxlan1",
  "type": 5,
  "enabled": true,
  "vxlan": {
    "src_address": "192.168.1.1",
    "dst_address": "192.168.1.2",
    "vni": 5
  }
}


/vnf-agent/agent_vpp_1/vpp/status/v1/bd/vpp1_bd1
{"index":1,"internal_name":"vpp1_bd1","bvi_interface":"\u003cnot_set\u003e","interface_count":2,"last_change":1504878585,"l2_params":{"flood":true,"unknown_unicast_flood":true,"forward":true,"learn":true,"arp_termination":true},"interfaces":[{"name":"vpp1_afpacket1","sw_if_index":2},{"name":"vpp1_vxlan1","sw_if_index":3}]}
/vnf-agent/agent_vpp_1/vpp/status/v1/interface/local0
{"name":"local0","internal_name":"local0","admin_status":2,"oper_status":2,"statistics":{}}
/vnf-agent/agent_vpp_1/vpp/status/v1/interface/vpp1_afpacket1
{"name":"vpp1_afpacket1","internal_name":"host-vpp1_veth2","if_index":2,"admin_status":1,"oper_status":1,"last_change":1504878579,"phys_address":"a2:a1:a1:a1:a1:a1","mtu":9216,"statistics":{"in_packets":4,"in_bytes":280,"drop_packets":4,"ipv6_packets":2}}
/vnf-agent/agent_vpp_1/vpp/status/v1/interface/vpp1_loop1
{"name":"vpp1_loop1","internal_name":"loop0","if_index":4,"admin_status":1,"oper_status":1,"last_change":1504878588,"phys_address":"12:21:21:11:11:11","mtu":9216,"statistics":{}}
/vnf-agent/agent_vpp_1/vpp/status/v1/interface/vpp1_memif1
{"name":"vpp1_memif1","internal_name":"memif0/1","if_index":1,"admin_status":1,"oper_status":2,"last_change":1504878570,"phys_address":"62:61:61:61:61:61","mtu":9216,"statistics":{"drop_packets":3,"out_error_packets":3}}
/vnf-agent/agent_vpp_1/vpp/status/v1/interface/vpp1_tap1
{"name":"vpp1_tap1","internal_name":"tap-0","if_index":5,"admin_status":1,"oper_status":1,"last_change":1504878591,"phys_address":"32:21:21:11:11:11","speed":1000000000,"mtu":9216,"statistics":{"in_packets":6,"in_bytes":480,"drop_packets":8,"ipv6_packets":6}}
/vnf-agent/agent_vpp_1/vpp/status/v1/interface/vpp1_vxlan1
{"name":"vpp1_vxlan1","internal_name":"vxlan_tunnel0","if_index":3,"admin_status":1,"oper_status":1,"last_change":1504878582,"statistics":{"out_packets":2,"out_bytes":212}}
/vnf-agent/agent_vpp_2/check/status/v1/agent
{"build_version":"644272b05b52f73057b2f838ab504399057df132","build_date":"2017-09-06T08:40+00:00","state":1,"start_time":1504878540,"last_change":1504878545,"last_update":1504878660}
/vnf-agent/agent_vpp_2/check/status/v1/plugin/etcdv3
{"state":1,"last_change":1504878545,"last_update":1504878660}
/vnf-agent/agent_vpp_2/check/status/v1/plugin/govpp
{"state":1,"last_change":1504878540,"last_update":1504878660}
/vnf-agent/agent_vpp_2/check/status/v1/plugin/kafka
{"state":1,"last_change":1504878545,"last_update":1504878660}
/vnf-agent/agent_vpp_2/vpp/status/v1/interface/local0
{"name":"local0","statistics":{}}'''

#print Parse_BD_Interfaces("agent_vpp_1","vpp1_bd1",y,x)

#print Convert_ETCD_Dump_To_JSON(z)
