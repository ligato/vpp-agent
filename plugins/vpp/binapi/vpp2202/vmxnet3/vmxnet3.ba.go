// Code generated by GoVPP's binapi-generator. DO NOT EDIT.

// Package vmxnet3 contains generated bindings for API file vmxnet3.api.
//
// Contents:
// -  2 structs
// -  8 messages
package vmxnet3

import (
	api "go.fd.io/govpp/api"
	codec "go.fd.io/govpp/codec"
	ethernet_types "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/ethernet_types"
	interface_types "go.ligato.io/vpp-agent/v3/plugins/vpp/binapi/vpp2202/interface_types"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the GoVPP api package it is being compiled against.
// A compilation error at this line likely means your copy of the
// GoVPP api package needs to be updated.
const _ = api.GoVppAPIPackageIsVersion2

const (
	APIFile    = "vmxnet3"
	APIVersion = "1.2.0"
	VersionCrc = 0x609454ea
)

// Vmxnet3RxList defines type 'vmxnet3_rx_list'.
type Vmxnet3RxList struct {
	RxQsize   uint16   `binapi:"u16,name=rx_qsize" json:"rx_qsize,omitempty"`
	RxFill    []uint16 `binapi:"u16[2],name=rx_fill" json:"rx_fill,omitempty"`
	RxNext    uint16   `binapi:"u16,name=rx_next" json:"rx_next,omitempty"`
	RxProduce []uint16 `binapi:"u16[2],name=rx_produce" json:"rx_produce,omitempty"`
	RxConsume []uint16 `binapi:"u16[2],name=rx_consume" json:"rx_consume,omitempty"`
}

// Vmxnet3TxList defines type 'vmxnet3_tx_list'.
type Vmxnet3TxList struct {
	TxQsize   uint16 `binapi:"u16,name=tx_qsize" json:"tx_qsize,omitempty"`
	TxNext    uint16 `binapi:"u16,name=tx_next" json:"tx_next,omitempty"`
	TxProduce uint16 `binapi:"u16,name=tx_produce" json:"tx_produce,omitempty"`
	TxConsume uint16 `binapi:"u16,name=tx_consume" json:"tx_consume,omitempty"`
}

// SwVmxnet3InterfaceDetails defines message 'sw_vmxnet3_interface_details'.
type SwVmxnet3InterfaceDetails struct {
	SwIfIndex   interface_types.InterfaceIndex `binapi:"interface_index,name=sw_if_index" json:"sw_if_index,omitempty"`
	IfName      string                         `binapi:"string[64],name=if_name" json:"if_name,omitempty"`
	HwAddr      ethernet_types.MacAddress      `binapi:"mac_address,name=hw_addr" json:"hw_addr,omitempty"`
	PciAddr     uint32                         `binapi:"u32,name=pci_addr" json:"pci_addr,omitempty"`
	Version     uint8                          `binapi:"u8,name=version" json:"version,omitempty"`
	AdminUpDown bool                           `binapi:"bool,name=admin_up_down" json:"admin_up_down,omitempty"`
	RxCount     uint8                          `binapi:"u8,name=rx_count" json:"rx_count,omitempty"`
	RxList      [16]Vmxnet3RxList              `binapi:"vmxnet3_rx_list[16],name=rx_list" json:"rx_list,omitempty"`
	TxCount     uint8                          `binapi:"u8,name=tx_count" json:"tx_count,omitempty"`
	TxList      [8]Vmxnet3TxList               `binapi:"vmxnet3_tx_list[8],name=tx_list" json:"tx_list,omitempty"`
}

func (m *SwVmxnet3InterfaceDetails) Reset()               { *m = SwVmxnet3InterfaceDetails{} }
func (*SwVmxnet3InterfaceDetails) GetMessageName() string { return "sw_vmxnet3_interface_details" }
func (*SwVmxnet3InterfaceDetails) GetCrcString() string   { return "6a1a5498" }
func (*SwVmxnet3InterfaceDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *SwVmxnet3InterfaceDetails) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4     // m.SwIfIndex
	size += 64    // m.IfName
	size += 1 * 6 // m.HwAddr
	size += 4     // m.PciAddr
	size += 1     // m.Version
	size += 1     // m.AdminUpDown
	size += 1     // m.RxCount
	for j1 := 0; j1 < 16; j1++ {
		size += 2     // m.RxList[j1].RxQsize
		size += 2 * 2 // m.RxList[j1].RxFill
		size += 2     // m.RxList[j1].RxNext
		size += 2 * 2 // m.RxList[j1].RxProduce
		size += 2 * 2 // m.RxList[j1].RxConsume
	}
	size += 1 // m.TxCount
	for j1 := 0; j1 < 8; j1++ {
		size += 2 // m.TxList[j1].TxQsize
		size += 2 // m.TxList[j1].TxNext
		size += 2 // m.TxList[j1].TxProduce
		size += 2 // m.TxList[j1].TxConsume
	}
	return size
}
func (m *SwVmxnet3InterfaceDetails) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeUint32(uint32(m.SwIfIndex))
	buf.EncodeString(m.IfName, 64)
	buf.EncodeBytes(m.HwAddr[:], 6)
	buf.EncodeUint32(m.PciAddr)
	buf.EncodeUint8(m.Version)
	buf.EncodeBool(m.AdminUpDown)
	buf.EncodeUint8(m.RxCount)
	for j0 := 0; j0 < 16; j0++ {
		buf.EncodeUint16(m.RxList[j0].RxQsize)
		for i := 0; i < 2; i++ {
			var x uint16
			if i < len(m.RxList[j0].RxFill) {
				x = uint16(m.RxList[j0].RxFill[i])
			}
			buf.EncodeUint16(x)
		}
		buf.EncodeUint16(m.RxList[j0].RxNext)
		for i := 0; i < 2; i++ {
			var x uint16
			if i < len(m.RxList[j0].RxProduce) {
				x = uint16(m.RxList[j0].RxProduce[i])
			}
			buf.EncodeUint16(x)
		}
		for i := 0; i < 2; i++ {
			var x uint16
			if i < len(m.RxList[j0].RxConsume) {
				x = uint16(m.RxList[j0].RxConsume[i])
			}
			buf.EncodeUint16(x)
		}
	}
	buf.EncodeUint8(m.TxCount)
	for j0 := 0; j0 < 8; j0++ {
		buf.EncodeUint16(m.TxList[j0].TxQsize)
		buf.EncodeUint16(m.TxList[j0].TxNext)
		buf.EncodeUint16(m.TxList[j0].TxProduce)
		buf.EncodeUint16(m.TxList[j0].TxConsume)
	}
	return buf.Bytes(), nil
}
func (m *SwVmxnet3InterfaceDetails) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	m.IfName = buf.DecodeString(64)
	copy(m.HwAddr[:], buf.DecodeBytes(6))
	m.PciAddr = buf.DecodeUint32()
	m.Version = buf.DecodeUint8()
	m.AdminUpDown = buf.DecodeBool()
	m.RxCount = buf.DecodeUint8()
	for j0 := 0; j0 < 16; j0++ {
		m.RxList[j0].RxQsize = buf.DecodeUint16()
		m.RxList[j0].RxFill = make([]uint16, 2)
		for i := 0; i < len(m.RxList[j0].RxFill); i++ {
			m.RxList[j0].RxFill[i] = buf.DecodeUint16()
		}
		m.RxList[j0].RxNext = buf.DecodeUint16()
		m.RxList[j0].RxProduce = make([]uint16, 2)
		for i := 0; i < len(m.RxList[j0].RxProduce); i++ {
			m.RxList[j0].RxProduce[i] = buf.DecodeUint16()
		}
		m.RxList[j0].RxConsume = make([]uint16, 2)
		for i := 0; i < len(m.RxList[j0].RxConsume); i++ {
			m.RxList[j0].RxConsume[i] = buf.DecodeUint16()
		}
	}
	m.TxCount = buf.DecodeUint8()
	for j0 := 0; j0 < 8; j0++ {
		m.TxList[j0].TxQsize = buf.DecodeUint16()
		m.TxList[j0].TxNext = buf.DecodeUint16()
		m.TxList[j0].TxProduce = buf.DecodeUint16()
		m.TxList[j0].TxConsume = buf.DecodeUint16()
	}
	return nil
}

// SwVmxnet3InterfaceDump defines message 'sw_vmxnet3_interface_dump'.
type SwVmxnet3InterfaceDump struct {
	SwIfIndex interface_types.InterfaceIndex `binapi:"interface_index,name=sw_if_index,default=4294967295" json:"sw_if_index,omitempty"`
}

func (m *SwVmxnet3InterfaceDump) Reset()               { *m = SwVmxnet3InterfaceDump{} }
func (*SwVmxnet3InterfaceDump) GetMessageName() string { return "sw_vmxnet3_interface_dump" }
func (*SwVmxnet3InterfaceDump) GetCrcString() string   { return "f9e6675e" }
func (*SwVmxnet3InterfaceDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *SwVmxnet3InterfaceDump) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.SwIfIndex
	return size
}
func (m *SwVmxnet3InterfaceDump) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeUint32(uint32(m.SwIfIndex))
	return buf.Bytes(), nil
}
func (m *SwVmxnet3InterfaceDump) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	return nil
}

// Vmxnet3Create defines message 'vmxnet3_create'.
type Vmxnet3Create struct {
	PciAddr    uint32 `binapi:"u32,name=pci_addr" json:"pci_addr,omitempty"`
	EnableElog int32  `binapi:"i32,name=enable_elog" json:"enable_elog,omitempty"`
	RxqSize    uint16 `binapi:"u16,name=rxq_size" json:"rxq_size,omitempty"`
	RxqNum     uint16 `binapi:"u16,name=rxq_num" json:"rxq_num,omitempty"`
	TxqSize    uint16 `binapi:"u16,name=txq_size" json:"txq_size,omitempty"`
	TxqNum     uint16 `binapi:"u16,name=txq_num" json:"txq_num,omitempty"`
	Bind       uint8  `binapi:"u8,name=bind" json:"bind,omitempty"`
	EnableGso  bool   `binapi:"bool,name=enable_gso" json:"enable_gso,omitempty"`
}

func (m *Vmxnet3Create) Reset()               { *m = Vmxnet3Create{} }
func (*Vmxnet3Create) GetMessageName() string { return "vmxnet3_create" }
func (*Vmxnet3Create) GetCrcString() string   { return "71a07314" }
func (*Vmxnet3Create) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *Vmxnet3Create) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.PciAddr
	size += 4 // m.EnableElog
	size += 2 // m.RxqSize
	size += 2 // m.RxqNum
	size += 2 // m.TxqSize
	size += 2 // m.TxqNum
	size += 1 // m.Bind
	size += 1 // m.EnableGso
	return size
}
func (m *Vmxnet3Create) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeUint32(m.PciAddr)
	buf.EncodeInt32(m.EnableElog)
	buf.EncodeUint16(m.RxqSize)
	buf.EncodeUint16(m.RxqNum)
	buf.EncodeUint16(m.TxqSize)
	buf.EncodeUint16(m.TxqNum)
	buf.EncodeUint8(m.Bind)
	buf.EncodeBool(m.EnableGso)
	return buf.Bytes(), nil
}
func (m *Vmxnet3Create) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.PciAddr = buf.DecodeUint32()
	m.EnableElog = buf.DecodeInt32()
	m.RxqSize = buf.DecodeUint16()
	m.RxqNum = buf.DecodeUint16()
	m.TxqSize = buf.DecodeUint16()
	m.TxqNum = buf.DecodeUint16()
	m.Bind = buf.DecodeUint8()
	m.EnableGso = buf.DecodeBool()
	return nil
}

// Vmxnet3CreateReply defines message 'vmxnet3_create_reply'.
type Vmxnet3CreateReply struct {
	Retval    int32                          `binapi:"i32,name=retval" json:"retval,omitempty"`
	SwIfIndex interface_types.InterfaceIndex `binapi:"interface_index,name=sw_if_index" json:"sw_if_index,omitempty"`
}

func (m *Vmxnet3CreateReply) Reset()               { *m = Vmxnet3CreateReply{} }
func (*Vmxnet3CreateReply) GetMessageName() string { return "vmxnet3_create_reply" }
func (*Vmxnet3CreateReply) GetCrcString() string   { return "5383d31f" }
func (*Vmxnet3CreateReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *Vmxnet3CreateReply) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.Retval
	size += 4 // m.SwIfIndex
	return size
}
func (m *Vmxnet3CreateReply) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeInt32(m.Retval)
	buf.EncodeUint32(uint32(m.SwIfIndex))
	return buf.Bytes(), nil
}
func (m *Vmxnet3CreateReply) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Retval = buf.DecodeInt32()
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	return nil
}

// Vmxnet3Delete defines message 'vmxnet3_delete'.
type Vmxnet3Delete struct {
	SwIfIndex interface_types.InterfaceIndex `binapi:"interface_index,name=sw_if_index" json:"sw_if_index,omitempty"`
}

func (m *Vmxnet3Delete) Reset()               { *m = Vmxnet3Delete{} }
func (*Vmxnet3Delete) GetMessageName() string { return "vmxnet3_delete" }
func (*Vmxnet3Delete) GetCrcString() string   { return "f9e6675e" }
func (*Vmxnet3Delete) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *Vmxnet3Delete) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.SwIfIndex
	return size
}
func (m *Vmxnet3Delete) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeUint32(uint32(m.SwIfIndex))
	return buf.Bytes(), nil
}
func (m *Vmxnet3Delete) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	return nil
}

// Vmxnet3DeleteReply defines message 'vmxnet3_delete_reply'.
type Vmxnet3DeleteReply struct {
	Retval int32 `binapi:"i32,name=retval" json:"retval,omitempty"`
}

func (m *Vmxnet3DeleteReply) Reset()               { *m = Vmxnet3DeleteReply{} }
func (*Vmxnet3DeleteReply) GetMessageName() string { return "vmxnet3_delete_reply" }
func (*Vmxnet3DeleteReply) GetCrcString() string   { return "e8d4e804" }
func (*Vmxnet3DeleteReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *Vmxnet3DeleteReply) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.Retval
	return size
}
func (m *Vmxnet3DeleteReply) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeInt32(m.Retval)
	return buf.Bytes(), nil
}
func (m *Vmxnet3DeleteReply) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Retval = buf.DecodeInt32()
	return nil
}

// Vmxnet3Details defines message 'vmxnet3_details'.
type Vmxnet3Details struct {
	SwIfIndex   interface_types.InterfaceIndex `binapi:"interface_index,name=sw_if_index" json:"sw_if_index,omitempty"`
	IfName      string                         `binapi:"string[64],name=if_name" json:"if_name,omitempty"`
	HwAddr      ethernet_types.MacAddress      `binapi:"mac_address,name=hw_addr" json:"hw_addr,omitempty"`
	PciAddr     uint32                         `binapi:"u32,name=pci_addr" json:"pci_addr,omitempty"`
	Version     uint8                          `binapi:"u8,name=version" json:"version,omitempty"`
	AdminUpDown bool                           `binapi:"bool,name=admin_up_down" json:"admin_up_down,omitempty"`
	RxCount     uint8                          `binapi:"u8,name=rx_count" json:"rx_count,omitempty"`
	RxList      [16]Vmxnet3RxList              `binapi:"vmxnet3_rx_list[16],name=rx_list" json:"rx_list,omitempty"`
	TxCount     uint8                          `binapi:"u8,name=tx_count" json:"tx_count,omitempty"`
	TxList      [8]Vmxnet3TxList               `binapi:"vmxnet3_tx_list[8],name=tx_list" json:"tx_list,omitempty"`
}

func (m *Vmxnet3Details) Reset()               { *m = Vmxnet3Details{} }
func (*Vmxnet3Details) GetMessageName() string { return "vmxnet3_details" }
func (*Vmxnet3Details) GetCrcString() string   { return "6a1a5498" }
func (*Vmxnet3Details) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *Vmxnet3Details) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4     // m.SwIfIndex
	size += 64    // m.IfName
	size += 1 * 6 // m.HwAddr
	size += 4     // m.PciAddr
	size += 1     // m.Version
	size += 1     // m.AdminUpDown
	size += 1     // m.RxCount
	for j1 := 0; j1 < 16; j1++ {
		size += 2     // m.RxList[j1].RxQsize
		size += 2 * 2 // m.RxList[j1].RxFill
		size += 2     // m.RxList[j1].RxNext
		size += 2 * 2 // m.RxList[j1].RxProduce
		size += 2 * 2 // m.RxList[j1].RxConsume
	}
	size += 1 // m.TxCount
	for j1 := 0; j1 < 8; j1++ {
		size += 2 // m.TxList[j1].TxQsize
		size += 2 // m.TxList[j1].TxNext
		size += 2 // m.TxList[j1].TxProduce
		size += 2 // m.TxList[j1].TxConsume
	}
	return size
}
func (m *Vmxnet3Details) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeUint32(uint32(m.SwIfIndex))
	buf.EncodeString(m.IfName, 64)
	buf.EncodeBytes(m.HwAddr[:], 6)
	buf.EncodeUint32(m.PciAddr)
	buf.EncodeUint8(m.Version)
	buf.EncodeBool(m.AdminUpDown)
	buf.EncodeUint8(m.RxCount)
	for j0 := 0; j0 < 16; j0++ {
		buf.EncodeUint16(m.RxList[j0].RxQsize)
		for i := 0; i < 2; i++ {
			var x uint16
			if i < len(m.RxList[j0].RxFill) {
				x = uint16(m.RxList[j0].RxFill[i])
			}
			buf.EncodeUint16(x)
		}
		buf.EncodeUint16(m.RxList[j0].RxNext)
		for i := 0; i < 2; i++ {
			var x uint16
			if i < len(m.RxList[j0].RxProduce) {
				x = uint16(m.RxList[j0].RxProduce[i])
			}
			buf.EncodeUint16(x)
		}
		for i := 0; i < 2; i++ {
			var x uint16
			if i < len(m.RxList[j0].RxConsume) {
				x = uint16(m.RxList[j0].RxConsume[i])
			}
			buf.EncodeUint16(x)
		}
	}
	buf.EncodeUint8(m.TxCount)
	for j0 := 0; j0 < 8; j0++ {
		buf.EncodeUint16(m.TxList[j0].TxQsize)
		buf.EncodeUint16(m.TxList[j0].TxNext)
		buf.EncodeUint16(m.TxList[j0].TxProduce)
		buf.EncodeUint16(m.TxList[j0].TxConsume)
	}
	return buf.Bytes(), nil
}
func (m *Vmxnet3Details) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	m.IfName = buf.DecodeString(64)
	copy(m.HwAddr[:], buf.DecodeBytes(6))
	m.PciAddr = buf.DecodeUint32()
	m.Version = buf.DecodeUint8()
	m.AdminUpDown = buf.DecodeBool()
	m.RxCount = buf.DecodeUint8()
	for j0 := 0; j0 < 16; j0++ {
		m.RxList[j0].RxQsize = buf.DecodeUint16()
		m.RxList[j0].RxFill = make([]uint16, 2)
		for i := 0; i < len(m.RxList[j0].RxFill); i++ {
			m.RxList[j0].RxFill[i] = buf.DecodeUint16()
		}
		m.RxList[j0].RxNext = buf.DecodeUint16()
		m.RxList[j0].RxProduce = make([]uint16, 2)
		for i := 0; i < len(m.RxList[j0].RxProduce); i++ {
			m.RxList[j0].RxProduce[i] = buf.DecodeUint16()
		}
		m.RxList[j0].RxConsume = make([]uint16, 2)
		for i := 0; i < len(m.RxList[j0].RxConsume); i++ {
			m.RxList[j0].RxConsume[i] = buf.DecodeUint16()
		}
	}
	m.TxCount = buf.DecodeUint8()
	for j0 := 0; j0 < 8; j0++ {
		m.TxList[j0].TxQsize = buf.DecodeUint16()
		m.TxList[j0].TxNext = buf.DecodeUint16()
		m.TxList[j0].TxProduce = buf.DecodeUint16()
		m.TxList[j0].TxConsume = buf.DecodeUint16()
	}
	return nil
}

// Vmxnet3Dump defines message 'vmxnet3_dump'.
// Deprecated: the message will be removed in the future versions
type Vmxnet3Dump struct{}

func (m *Vmxnet3Dump) Reset()               { *m = Vmxnet3Dump{} }
func (*Vmxnet3Dump) GetMessageName() string { return "vmxnet3_dump" }
func (*Vmxnet3Dump) GetCrcString() string   { return "51077d14" }
func (*Vmxnet3Dump) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *Vmxnet3Dump) Size() (size int) {
	if m == nil {
		return 0
	}
	return size
}
func (m *Vmxnet3Dump) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	return buf.Bytes(), nil
}
func (m *Vmxnet3Dump) Unmarshal(b []byte) error {
	return nil
}

func init() { file_vmxnet3_binapi_init() }
func file_vmxnet3_binapi_init() {
	api.RegisterMessage((*SwVmxnet3InterfaceDetails)(nil), "sw_vmxnet3_interface_details_6a1a5498")
	api.RegisterMessage((*SwVmxnet3InterfaceDump)(nil), "sw_vmxnet3_interface_dump_f9e6675e")
	api.RegisterMessage((*Vmxnet3Create)(nil), "vmxnet3_create_71a07314")
	api.RegisterMessage((*Vmxnet3CreateReply)(nil), "vmxnet3_create_reply_5383d31f")
	api.RegisterMessage((*Vmxnet3Delete)(nil), "vmxnet3_delete_f9e6675e")
	api.RegisterMessage((*Vmxnet3DeleteReply)(nil), "vmxnet3_delete_reply_e8d4e804")
	api.RegisterMessage((*Vmxnet3Details)(nil), "vmxnet3_details_6a1a5498")
	api.RegisterMessage((*Vmxnet3Dump)(nil), "vmxnet3_dump_51077d14")
}

// Messages returns list of all messages in this module.
func AllMessages() []api.Message {
	return []api.Message{
		(*SwVmxnet3InterfaceDetails)(nil),
		(*SwVmxnet3InterfaceDump)(nil),
		(*Vmxnet3Create)(nil),
		(*Vmxnet3CreateReply)(nil),
		(*Vmxnet3Delete)(nil),
		(*Vmxnet3DeleteReply)(nil),
		(*Vmxnet3Details)(nil),
		(*Vmxnet3Dump)(nil),
	}
}
