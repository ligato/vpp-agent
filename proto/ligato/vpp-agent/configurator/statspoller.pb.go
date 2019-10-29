// Code generated by protoc-gen-go. DO NOT EDIT.
// source: ligato/vpp-agent/configurator/statspoller.proto

package configurator

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	vpp "go.ligato.io/vpp-agent/v2/proto/ligato/vpp-agent/vpp"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type Stats struct {
	// Types that are valid to be assigned to Stats:
	//	*Stats_VppStats
	Stats                isStats_Stats `protobuf_oneof:"stats"`
	XXX_NoUnkeyedLiteral struct{}      `json:"-"`
	XXX_unrecognized     []byte        `json:"-"`
	XXX_sizecache        int32         `json:"-"`
}

func (m *Stats) Reset()         { *m = Stats{} }
func (m *Stats) String() string { return proto.CompactTextString(m) }
func (*Stats) ProtoMessage()    {}
func (*Stats) Descriptor() ([]byte, []int) {
	return fileDescriptor_fa2922171d6be150, []int{0}
}

func (m *Stats) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Stats.Unmarshal(m, b)
}
func (m *Stats) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Stats.Marshal(b, m, deterministic)
}
func (m *Stats) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Stats.Merge(m, src)
}
func (m *Stats) XXX_Size() int {
	return xxx_messageInfo_Stats.Size(m)
}
func (m *Stats) XXX_DiscardUnknown() {
	xxx_messageInfo_Stats.DiscardUnknown(m)
}

var xxx_messageInfo_Stats proto.InternalMessageInfo

type isStats_Stats interface {
	isStats_Stats()
}

type Stats_VppStats struct {
	VppStats *vpp.Stats `protobuf:"bytes,1,opt,name=vpp_stats,json=vppStats,proto3,oneof"`
}

func (*Stats_VppStats) isStats_Stats() {}

func (m *Stats) GetStats() isStats_Stats {
	if m != nil {
		return m.Stats
	}
	return nil
}

func (m *Stats) GetVppStats() *vpp.Stats {
	if x, ok := m.GetStats().(*Stats_VppStats); ok {
		return x.VppStats
	}
	return nil
}

// XXX_OneofWrappers is for the internal use of the proto package.
func (*Stats) XXX_OneofWrappers() []interface{} {
	return []interface{}{
		(*Stats_VppStats)(nil),
	}
}

type PollStatsRequest struct {
	// PeriodSec defines polling period (in seconds)
	PeriodSec            uint32   `protobuf:"varint,1,opt,name=period_sec,json=periodSec,proto3" json:"period_sec,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PollStatsRequest) Reset()         { *m = PollStatsRequest{} }
func (m *PollStatsRequest) String() string { return proto.CompactTextString(m) }
func (*PollStatsRequest) ProtoMessage()    {}
func (*PollStatsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_fa2922171d6be150, []int{1}
}

func (m *PollStatsRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PollStatsRequest.Unmarshal(m, b)
}
func (m *PollStatsRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PollStatsRequest.Marshal(b, m, deterministic)
}
func (m *PollStatsRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PollStatsRequest.Merge(m, src)
}
func (m *PollStatsRequest) XXX_Size() int {
	return xxx_messageInfo_PollStatsRequest.Size(m)
}
func (m *PollStatsRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_PollStatsRequest.DiscardUnknown(m)
}

var xxx_messageInfo_PollStatsRequest proto.InternalMessageInfo

func (m *PollStatsRequest) GetPeriodSec() uint32 {
	if m != nil {
		return m.PeriodSec
	}
	return 0
}

type PollStatsResponse struct {
	PollSeq              uint32   `protobuf:"varint,1,opt,name=poll_seq,json=pollSeq,proto3" json:"poll_seq,omitempty"`
	Stats                *Stats   `protobuf:"bytes,2,opt,name=stats,proto3" json:"stats,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PollStatsResponse) Reset()         { *m = PollStatsResponse{} }
func (m *PollStatsResponse) String() string { return proto.CompactTextString(m) }
func (*PollStatsResponse) ProtoMessage()    {}
func (*PollStatsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_fa2922171d6be150, []int{2}
}

func (m *PollStatsResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PollStatsResponse.Unmarshal(m, b)
}
func (m *PollStatsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PollStatsResponse.Marshal(b, m, deterministic)
}
func (m *PollStatsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PollStatsResponse.Merge(m, src)
}
func (m *PollStatsResponse) XXX_Size() int {
	return xxx_messageInfo_PollStatsResponse.Size(m)
}
func (m *PollStatsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_PollStatsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_PollStatsResponse proto.InternalMessageInfo

func (m *PollStatsResponse) GetPollSeq() uint32 {
	if m != nil {
		return m.PollSeq
	}
	return 0
}

func (m *PollStatsResponse) GetStats() *Stats {
	if m != nil {
		return m.Stats
	}
	return nil
}

func init() {
	proto.RegisterType((*Stats)(nil), "ligato.vpp_agent.configurator.Stats")
	proto.RegisterType((*PollStatsRequest)(nil), "ligato.vpp_agent.configurator.PollStatsRequest")
	proto.RegisterType((*PollStatsResponse)(nil), "ligato.vpp_agent.configurator.PollStatsResponse")
}

func init() {
	proto.RegisterFile("ligato/vpp-agent/configurator/statspoller.proto", fileDescriptor_fa2922171d6be150)
}

var fileDescriptor_fa2922171d6be150 = []byte{
	// 285 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x92, 0xc1, 0x4b, 0xc3, 0x30,
	0x14, 0xc6, 0x57, 0x61, 0x6e, 0x7d, 0x43, 0xd0, 0x9c, 0xe6, 0x64, 0x22, 0xc5, 0x83, 0x17, 0x9b,
	0x59, 0x6f, 0xf3, 0xb6, 0x93, 0xc8, 0x0e, 0xa3, 0xbd, 0x79, 0x29, 0xb5, 0xc6, 0x52, 0x09, 0xcd,
	0x6b, 0x92, 0xf5, 0xee, 0x7f, 0x2e, 0x79, 0x99, 0x52, 0x1d, 0x4c, 0x0f, 0x85, 0xbe, 0x47, 0x7e,
	0xdf, 0xf7, 0xbe, 0xbc, 0x00, 0x97, 0x75, 0x55, 0x58, 0xc5, 0x3b, 0xc4, 0xdb, 0xa2, 0x12, 0x8d,
	0xe5, 0xa5, 0x6a, 0xde, 0xea, 0x6a, 0xab, 0x0b, 0xab, 0x34, 0x37, 0xb6, 0xb0, 0x06, 0x95, 0x94,
	0x42, 0xc7, 0xa8, 0x95, 0x55, 0x6c, 0xee, 0x81, 0xb8, 0x43, 0xcc, 0x09, 0x88, 0xfb, 0xc0, 0xec,
	0x72, 0x4f, 0xaf, 0x43, 0x74, 0x9f, 0xc7, 0xa3, 0x35, 0x0c, 0x33, 0xa7, 0xc9, 0x96, 0x10, 0x3a,
	0x09, 0x32, 0x98, 0x06, 0x57, 0xc1, 0xcd, 0x24, 0xb9, 0x88, 0xf7, 0xb4, 0x1d, 0x48, 0xe7, 0x1f,
	0x07, 0xe9, 0xb8, 0x43, 0xa4, 0xff, 0xd5, 0x08, 0x86, 0xc4, 0x45, 0x77, 0x70, 0xba, 0x51, 0x52,
	0x52, 0x37, 0x15, 0xed, 0x56, 0x18, 0xcb, 0xe6, 0x00, 0x28, 0x74, 0xad, 0x5e, 0x73, 0x23, 0x4a,
	0x52, 0x3e, 0x49, 0x43, 0xdf, 0xc9, 0x44, 0x19, 0xbd, 0xc3, 0x59, 0x0f, 0x31, 0xa8, 0x1a, 0x23,
	0xd8, 0x39, 0x8c, 0x5d, 0xc8, 0xdc, 0x88, 0x76, 0x47, 0x8c, 0x5c, 0x9d, 0x89, 0x96, 0x2d, 0x77,
	0x5e, 0xd3, 0x23, 0x9a, 0xf1, 0x3a, 0x3e, 0x98, 0xdf, 0x0f, 0x9b, 0x7a, 0x24, 0xf9, 0x08, 0x60,
	0x42, 0x8d, 0x0d, 0xdd, 0x20, 0xd3, 0x10, 0x7e, 0x7b, 0x33, 0xfe, 0x87, 0xd2, 0xef, 0x60, 0xb3,
	0xc5, 0xff, 0x01, 0x1f, 0x2b, 0x1a, 0x2c, 0x82, 0xd5, 0xfa, 0xf9, 0xa9, 0x52, 0x5f, 0x64, 0xfd,
	0x63, 0x31, 0x09, 0xa7, 0x9d, 0x1c, 0x7e, 0x02, 0x0f, 0xfd, 0xe2, 0xe5, 0x98, 0x88, 0xfb, 0xcf,
	0x00, 0x00, 0x00, 0xff, 0xff, 0xa8, 0x89, 0xfa, 0x8e, 0x37, 0x02, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// StatsPollerClient is the client API for StatsPoller service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type StatsPollerClient interface {
	// PollStats is used for polling metrics using poll period.
	PollStats(ctx context.Context, in *PollStatsRequest, opts ...grpc.CallOption) (StatsPoller_PollStatsClient, error)
}

type statsPollerClient struct {
	cc *grpc.ClientConn
}

func NewStatsPollerClient(cc *grpc.ClientConn) StatsPollerClient {
	return &statsPollerClient{cc}
}

func (c *statsPollerClient) PollStats(ctx context.Context, in *PollStatsRequest, opts ...grpc.CallOption) (StatsPoller_PollStatsClient, error) {
	stream, err := c.cc.NewStream(ctx, &_StatsPoller_serviceDesc.Streams[0], "/ligato.vpp_agent.configurator.StatsPoller/PollStats", opts...)
	if err != nil {
		return nil, err
	}
	x := &statsPollerPollStatsClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type StatsPoller_PollStatsClient interface {
	Recv() (*PollStatsResponse, error)
	grpc.ClientStream
}

type statsPollerPollStatsClient struct {
	grpc.ClientStream
}

func (x *statsPollerPollStatsClient) Recv() (*PollStatsResponse, error) {
	m := new(PollStatsResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// StatsPollerServer is the server API for StatsPoller service.
type StatsPollerServer interface {
	// PollStats is used for polling metrics using poll period.
	PollStats(*PollStatsRequest, StatsPoller_PollStatsServer) error
}

// UnimplementedStatsPollerServer can be embedded to have forward compatible implementations.
type UnimplementedStatsPollerServer struct {
}

func (*UnimplementedStatsPollerServer) PollStats(req *PollStatsRequest, srv StatsPoller_PollStatsServer) error {
	return status.Errorf(codes.Unimplemented, "method PollStats not implemented")
}

func RegisterStatsPollerServer(s *grpc.Server, srv StatsPollerServer) {
	s.RegisterService(&_StatsPoller_serviceDesc, srv)
}

func _StatsPoller_PollStats_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(PollStatsRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(StatsPollerServer).PollStats(m, &statsPollerPollStatsServer{stream})
}

type StatsPoller_PollStatsServer interface {
	Send(*PollStatsResponse) error
	grpc.ServerStream
}

type statsPollerPollStatsServer struct {
	grpc.ServerStream
}

func (x *statsPollerPollStatsServer) Send(m *PollStatsResponse) error {
	return x.ServerStream.SendMsg(m)
}

var _StatsPoller_serviceDesc = grpc.ServiceDesc{
	ServiceName: "ligato.vpp_agent.configurator.StatsPoller",
	HandlerType: (*StatsPollerServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "PollStats",
			Handler:       _StatsPoller_PollStats_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "ligato/vpp-agent/configurator/statspoller.proto",
}