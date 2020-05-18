// Code generated by GoVPP's binapi-generator. DO NOT EDIT.
// source: /usr/share/vpp/api/core/ipfix_export.api.json

/*
Package ipfix_export is a generated VPP binary API for 'ipfix_export' module.

It consists of:
	 12 messages
	  6 services
*/
package ipfix_export

import (
	"bytes"
	"context"
	"io"
	"strconv"

	api "git.fd.io/govpp.git/api"
	struc "github.com/lunixbochs/struc"
)

const (
	// ModuleName is the name of this module.
	ModuleName = "ipfix_export"
	// APIVersion is the API version of this module.
	APIVersion = "1.0.0"
	// VersionCrc is the CRC of this module.
	VersionCrc = 0x3e08644f
)

// IpfixClassifyStreamDetails represents VPP binary API message 'ipfix_classify_stream_details'.
type IpfixClassifyStreamDetails struct {
	DomainID uint32
	SrcPort  uint16
}

func (m *IpfixClassifyStreamDetails) Reset()                        { *m = IpfixClassifyStreamDetails{} }
func (*IpfixClassifyStreamDetails) GetMessageName() string          { return "ipfix_classify_stream_details" }
func (*IpfixClassifyStreamDetails) GetCrcString() string            { return "2903539d" }
func (*IpfixClassifyStreamDetails) GetMessageType() api.MessageType { return api.ReplyMessage }

// IpfixClassifyStreamDump represents VPP binary API message 'ipfix_classify_stream_dump'.
type IpfixClassifyStreamDump struct{}

func (m *IpfixClassifyStreamDump) Reset()                        { *m = IpfixClassifyStreamDump{} }
func (*IpfixClassifyStreamDump) GetMessageName() string          { return "ipfix_classify_stream_dump" }
func (*IpfixClassifyStreamDump) GetCrcString() string            { return "51077d14" }
func (*IpfixClassifyStreamDump) GetMessageType() api.MessageType { return api.RequestMessage }

// IpfixClassifyTableAddDel represents VPP binary API message 'ipfix_classify_table_add_del'.
type IpfixClassifyTableAddDel struct {
	TableID           uint32
	IPVersion         uint8
	TransportProtocol uint8
	IsAdd             uint8
}

func (m *IpfixClassifyTableAddDel) Reset()                        { *m = IpfixClassifyTableAddDel{} }
func (*IpfixClassifyTableAddDel) GetMessageName() string          { return "ipfix_classify_table_add_del" }
func (*IpfixClassifyTableAddDel) GetCrcString() string            { return "48efe167" }
func (*IpfixClassifyTableAddDel) GetMessageType() api.MessageType { return api.RequestMessage }

// IpfixClassifyTableAddDelReply represents VPP binary API message 'ipfix_classify_table_add_del_reply'.
type IpfixClassifyTableAddDelReply struct {
	Retval int32
}

func (m *IpfixClassifyTableAddDelReply) Reset() { *m = IpfixClassifyTableAddDelReply{} }
func (*IpfixClassifyTableAddDelReply) GetMessageName() string {
	return "ipfix_classify_table_add_del_reply"
}
func (*IpfixClassifyTableAddDelReply) GetCrcString() string            { return "e8d4e804" }
func (*IpfixClassifyTableAddDelReply) GetMessageType() api.MessageType { return api.ReplyMessage }

// IpfixClassifyTableDetails represents VPP binary API message 'ipfix_classify_table_details'.
type IpfixClassifyTableDetails struct {
	TableID           uint32
	IPVersion         uint8
	TransportProtocol uint8
}

func (m *IpfixClassifyTableDetails) Reset()                        { *m = IpfixClassifyTableDetails{} }
func (*IpfixClassifyTableDetails) GetMessageName() string          { return "ipfix_classify_table_details" }
func (*IpfixClassifyTableDetails) GetCrcString() string            { return "973d0d5b" }
func (*IpfixClassifyTableDetails) GetMessageType() api.MessageType { return api.ReplyMessage }

// IpfixClassifyTableDump represents VPP binary API message 'ipfix_classify_table_dump'.
type IpfixClassifyTableDump struct{}

func (m *IpfixClassifyTableDump) Reset()                        { *m = IpfixClassifyTableDump{} }
func (*IpfixClassifyTableDump) GetMessageName() string          { return "ipfix_classify_table_dump" }
func (*IpfixClassifyTableDump) GetCrcString() string            { return "51077d14" }
func (*IpfixClassifyTableDump) GetMessageType() api.MessageType { return api.RequestMessage }

// IpfixExporterDetails represents VPP binary API message 'ipfix_exporter_details'.
type IpfixExporterDetails struct {
	CollectorAddress []byte `struc:"[16]byte"`
	CollectorPort    uint16
	SrcAddress       []byte `struc:"[16]byte"`
	VrfID            uint32
	PathMtu          uint32
	TemplateInterval uint32
	UDPChecksum      uint8
}

func (m *IpfixExporterDetails) Reset()                        { *m = IpfixExporterDetails{} }
func (*IpfixExporterDetails) GetMessageName() string          { return "ipfix_exporter_details" }
func (*IpfixExporterDetails) GetCrcString() string            { return "742dddee" }
func (*IpfixExporterDetails) GetMessageType() api.MessageType { return api.ReplyMessage }

// IpfixExporterDump represents VPP binary API message 'ipfix_exporter_dump'.
type IpfixExporterDump struct{}

func (m *IpfixExporterDump) Reset()                        { *m = IpfixExporterDump{} }
func (*IpfixExporterDump) GetMessageName() string          { return "ipfix_exporter_dump" }
func (*IpfixExporterDump) GetCrcString() string            { return "51077d14" }
func (*IpfixExporterDump) GetMessageType() api.MessageType { return api.RequestMessage }

// SetIpfixClassifyStream represents VPP binary API message 'set_ipfix_classify_stream'.
type SetIpfixClassifyStream struct {
	DomainID uint32
	SrcPort  uint16
}

func (m *SetIpfixClassifyStream) Reset()                        { *m = SetIpfixClassifyStream{} }
func (*SetIpfixClassifyStream) GetMessageName() string          { return "set_ipfix_classify_stream" }
func (*SetIpfixClassifyStream) GetCrcString() string            { return "c9cbe053" }
func (*SetIpfixClassifyStream) GetMessageType() api.MessageType { return api.RequestMessage }

// SetIpfixClassifyStreamReply represents VPP binary API message 'set_ipfix_classify_stream_reply'.
type SetIpfixClassifyStreamReply struct {
	Retval int32
}

func (m *SetIpfixClassifyStreamReply) Reset()                        { *m = SetIpfixClassifyStreamReply{} }
func (*SetIpfixClassifyStreamReply) GetMessageName() string          { return "set_ipfix_classify_stream_reply" }
func (*SetIpfixClassifyStreamReply) GetCrcString() string            { return "e8d4e804" }
func (*SetIpfixClassifyStreamReply) GetMessageType() api.MessageType { return api.ReplyMessage }

// SetIpfixExporter represents VPP binary API message 'set_ipfix_exporter'.
type SetIpfixExporter struct {
	CollectorAddress []byte `struc:"[16]byte"`
	CollectorPort    uint16
	SrcAddress       []byte `struc:"[16]byte"`
	VrfID            uint32
	PathMtu          uint32
	TemplateInterval uint32
	UDPChecksum      uint8
}

func (m *SetIpfixExporter) Reset()                        { *m = SetIpfixExporter{} }
func (*SetIpfixExporter) GetMessageName() string          { return "set_ipfix_exporter" }
func (*SetIpfixExporter) GetCrcString() string            { return "4ff71dea" }
func (*SetIpfixExporter) GetMessageType() api.MessageType { return api.RequestMessage }

// SetIpfixExporterReply represents VPP binary API message 'set_ipfix_exporter_reply'.
type SetIpfixExporterReply struct {
	Retval int32
}

func (m *SetIpfixExporterReply) Reset()                        { *m = SetIpfixExporterReply{} }
func (*SetIpfixExporterReply) GetMessageName() string          { return "set_ipfix_exporter_reply" }
func (*SetIpfixExporterReply) GetCrcString() string            { return "e8d4e804" }
func (*SetIpfixExporterReply) GetMessageType() api.MessageType { return api.ReplyMessage }

func init() {
	api.RegisterMessage((*IpfixClassifyStreamDetails)(nil), "ipfix_export.IpfixClassifyStreamDetails")
	api.RegisterMessage((*IpfixClassifyStreamDump)(nil), "ipfix_export.IpfixClassifyStreamDump")
	api.RegisterMessage((*IpfixClassifyTableAddDel)(nil), "ipfix_export.IpfixClassifyTableAddDel")
	api.RegisterMessage((*IpfixClassifyTableAddDelReply)(nil), "ipfix_export.IpfixClassifyTableAddDelReply")
	api.RegisterMessage((*IpfixClassifyTableDetails)(nil), "ipfix_export.IpfixClassifyTableDetails")
	api.RegisterMessage((*IpfixClassifyTableDump)(nil), "ipfix_export.IpfixClassifyTableDump")
	api.RegisterMessage((*IpfixExporterDetails)(nil), "ipfix_export.IpfixExporterDetails")
	api.RegisterMessage((*IpfixExporterDump)(nil), "ipfix_export.IpfixExporterDump")
	api.RegisterMessage((*SetIpfixClassifyStream)(nil), "ipfix_export.SetIpfixClassifyStream")
	api.RegisterMessage((*SetIpfixClassifyStreamReply)(nil), "ipfix_export.SetIpfixClassifyStreamReply")
	api.RegisterMessage((*SetIpfixExporter)(nil), "ipfix_export.SetIpfixExporter")
	api.RegisterMessage((*SetIpfixExporterReply)(nil), "ipfix_export.SetIpfixExporterReply")
}

// Messages returns list of all messages in this module.
func AllMessages() []api.Message {
	return []api.Message{
		(*IpfixClassifyStreamDetails)(nil),
		(*IpfixClassifyStreamDump)(nil),
		(*IpfixClassifyTableAddDel)(nil),
		(*IpfixClassifyTableAddDelReply)(nil),
		(*IpfixClassifyTableDetails)(nil),
		(*IpfixClassifyTableDump)(nil),
		(*IpfixExporterDetails)(nil),
		(*IpfixExporterDump)(nil),
		(*SetIpfixClassifyStream)(nil),
		(*SetIpfixClassifyStreamReply)(nil),
		(*SetIpfixExporter)(nil),
		(*SetIpfixExporterReply)(nil),
	}
}

// RPCService represents RPC service API for ipfix_export module.
type RPCService interface {
	DumpIpfixClassifyStream(ctx context.Context, in *IpfixClassifyStreamDump) (RPCService_DumpIpfixClassifyStreamClient, error)
	DumpIpfixClassifyTable(ctx context.Context, in *IpfixClassifyTableDump) (RPCService_DumpIpfixClassifyTableClient, error)
	DumpIpfixExporter(ctx context.Context, in *IpfixExporterDump) (RPCService_DumpIpfixExporterClient, error)
	IpfixClassifyTableAddDel(ctx context.Context, in *IpfixClassifyTableAddDel) (*IpfixClassifyTableAddDelReply, error)
	SetIpfixClassifyStream(ctx context.Context, in *SetIpfixClassifyStream) (*SetIpfixClassifyStreamReply, error)
	SetIpfixExporter(ctx context.Context, in *SetIpfixExporter) (*SetIpfixExporterReply, error)
}

type serviceClient struct {
	ch api.Channel
}

func NewServiceClient(ch api.Channel) RPCService {
	return &serviceClient{ch}
}

func (c *serviceClient) DumpIpfixClassifyStream(ctx context.Context, in *IpfixClassifyStreamDump) (RPCService_DumpIpfixClassifyStreamClient, error) {
	stream := c.ch.SendMultiRequest(in)
	x := &serviceClient_DumpIpfixClassifyStreamClient{stream}
	return x, nil
}

type RPCService_DumpIpfixClassifyStreamClient interface {
	Recv() (*IpfixClassifyStreamDetails, error)
}

type serviceClient_DumpIpfixClassifyStreamClient struct {
	api.MultiRequestCtx
}

func (c *serviceClient_DumpIpfixClassifyStreamClient) Recv() (*IpfixClassifyStreamDetails, error) {
	m := new(IpfixClassifyStreamDetails)
	stop, err := c.MultiRequestCtx.ReceiveReply(m)
	if err != nil {
		return nil, err
	}
	if stop {
		return nil, io.EOF
	}
	return m, nil
}

func (c *serviceClient) DumpIpfixClassifyTable(ctx context.Context, in *IpfixClassifyTableDump) (RPCService_DumpIpfixClassifyTableClient, error) {
	stream := c.ch.SendMultiRequest(in)
	x := &serviceClient_DumpIpfixClassifyTableClient{stream}
	return x, nil
}

type RPCService_DumpIpfixClassifyTableClient interface {
	Recv() (*IpfixClassifyTableDetails, error)
}

type serviceClient_DumpIpfixClassifyTableClient struct {
	api.MultiRequestCtx
}

func (c *serviceClient_DumpIpfixClassifyTableClient) Recv() (*IpfixClassifyTableDetails, error) {
	m := new(IpfixClassifyTableDetails)
	stop, err := c.MultiRequestCtx.ReceiveReply(m)
	if err != nil {
		return nil, err
	}
	if stop {
		return nil, io.EOF
	}
	return m, nil
}

func (c *serviceClient) DumpIpfixExporter(ctx context.Context, in *IpfixExporterDump) (RPCService_DumpIpfixExporterClient, error) {
	stream := c.ch.SendMultiRequest(in)
	x := &serviceClient_DumpIpfixExporterClient{stream}
	return x, nil
}

type RPCService_DumpIpfixExporterClient interface {
	Recv() (*IpfixExporterDetails, error)
}

type serviceClient_DumpIpfixExporterClient struct {
	api.MultiRequestCtx
}

func (c *serviceClient_DumpIpfixExporterClient) Recv() (*IpfixExporterDetails, error) {
	m := new(IpfixExporterDetails)
	stop, err := c.MultiRequestCtx.ReceiveReply(m)
	if err != nil {
		return nil, err
	}
	if stop {
		return nil, io.EOF
	}
	return m, nil
}

func (c *serviceClient) IpfixClassifyTableAddDel(ctx context.Context, in *IpfixClassifyTableAddDel) (*IpfixClassifyTableAddDelReply, error) {
	out := new(IpfixClassifyTableAddDelReply)
	err := c.ch.SendRequest(in).ReceiveReply(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) SetIpfixClassifyStream(ctx context.Context, in *SetIpfixClassifyStream) (*SetIpfixClassifyStreamReply, error) {
	out := new(SetIpfixClassifyStreamReply)
	err := c.ch.SendRequest(in).ReceiveReply(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *serviceClient) SetIpfixExporter(ctx context.Context, in *SetIpfixExporter) (*SetIpfixExporterReply, error) {
	out := new(SetIpfixExporterReply)
	err := c.ch.SendRequest(in).ReceiveReply(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// This is a compile-time assertion to ensure that this generated file
// is compatible with the GoVPP api package it is being compiled against.
// A compilation error at this line likely means your copy of the
// GoVPP api package needs to be updated.
const _ = api.GoVppAPIPackageIsVersion1 // please upgrade the GoVPP api package

// Reference imports to suppress errors if they are not otherwise used.
var _ = api.RegisterMessage
var _ = bytes.NewBuffer
var _ = context.Background
var _ = io.Copy
var _ = strconv.Itoa
var _ = struc.Pack
