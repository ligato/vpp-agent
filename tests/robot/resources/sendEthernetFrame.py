#!/usr/bin/env python
from socket import socket, AF_PACKET, SOCK_RAW
s = socket(AF_PACKET, SOCK_RAW)
s.bind(("${out_interface}", 0))

src_mac_addr = "${source_address}".replace(":","").decode("hex")
dst_mac_addr = "${destination_address}".replace(":","").decode("hex")
payload = (${payload})
checksum = "${checksum}".decode("hex")
ethertype = "${ethernet_type}".decode("hex")

s.send(dst_mac_addr+src_mac_addr+ethertype+payload+checksum)