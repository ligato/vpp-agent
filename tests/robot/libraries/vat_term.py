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

