// Code generated by protoc-gen-go. DO NOT EDIT.
// source: ligato/vpp-agent/kvscheduler/value_status.proto

package kvscheduler

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
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

type ValueState int32

const (
	// ValueState_NONEXISTENT is assigned to value that was deleted or has never
	// existed.
	ValueState_NONEXISTENT ValueState = 0
	// ValueState_MISSING is assigned to NB value that was configured but refresh
	// found it to be missing.
	ValueState_MISSING ValueState = 1
	// ValueState_UNIMPLEMENTED marks value received from NB that cannot
	// be configured because there is no registered descriptor associated
	// with it.
	ValueState_UNIMPLEMENTED ValueState = 2
	// ValueState_REMOVED is assigned to NB value after it was removed or when
	// it is being re-created. The state is only temporary: for re-create, the
	// value transits to whatever state the following Create operation produces,
	// and delete values are removed from the graph (go to the NONEXISTENT state)
	// immediately after the notification about the state change is sent.
	ValueState_REMOVED ValueState = 3
	// ValueState_CONFIGURED marks value defined by NB and successfully configured.
	ValueState_CONFIGURED ValueState = 4
	// ValueState_OBTAINED marks value not managed by NB, instead created
	// automatically or externally in SB. The KVScheduler learns about the value
	// either using Retrieve() or through a SB notification.
	ValueState_OBTAINED ValueState = 5
	// ValueState_DISCOVERED marks NB value that was found (=retrieved) by refresh
	// but not actually configured by the agent in this run.
	ValueState_DISCOVERED ValueState = 6
	// ValueState_PENDING represents (NB) value that cannot be configured yet
	// due to missing dependencies.
	ValueState_PENDING ValueState = 7
	// ValueState_INVALID represents (NB) value that will not be configured
	// because it has a logically invalid content as declared by the Validate
	// method of the associated descriptor.
	// The corresponding error and the list of affected fields are stored
	// in the <InvalidValueDetails> structure available via <details> for invalid
	// value.
	ValueState_INVALID ValueState = 8
	// ValueState_FAILED marks (NB) value for which the last executed operation
	// returned an error.
	// The error and the type of the operation which caused the error are stored
	// in the <FailedValueDetails> structure available via <details> for failed
	// value.
	ValueState_FAILED ValueState = 9
	// ValueState_RETRYING marks unsucessfully applied (NB) value, for which,
	// however, one or more attempts to fix the error by repeating the last
	// operation are planned, and only if all the retries fail, the value will
	// then transit to the FAILED state.
	ValueState_RETRYING ValueState = 10
)

var ValueState_name = map[int32]string{
	0:  "NONEXISTENT",
	1:  "MISSING",
	2:  "UNIMPLEMENTED",
	3:  "REMOVED",
	4:  "CONFIGURED",
	5:  "OBTAINED",
	6:  "DISCOVERED",
	7:  "PENDING",
	8:  "INVALID",
	9:  "FAILED",
	10: "RETRYING",
}

var ValueState_value = map[string]int32{
	"NONEXISTENT":   0,
	"MISSING":       1,
	"UNIMPLEMENTED": 2,
	"REMOVED":       3,
	"CONFIGURED":    4,
	"OBTAINED":      5,
	"DISCOVERED":    6,
	"PENDING":       7,
	"INVALID":       8,
	"FAILED":        9,
	"RETRYING":      10,
}

func (x ValueState) String() string {
	return proto.EnumName(ValueState_name, int32(x))
}

func (ValueState) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_676435f6cd35671d, []int{0}
}

type TxnOperation int32

const (
	TxnOperation_UNDEFINED TxnOperation = 0
	TxnOperation_VALIDATE  TxnOperation = 1
	TxnOperation_CREATE    TxnOperation = 2
	TxnOperation_UPDATE    TxnOperation = 3
	TxnOperation_DELETE    TxnOperation = 4
)

var TxnOperation_name = map[int32]string{
	0: "UNDEFINED",
	1: "VALIDATE",
	2: "CREATE",
	3: "UPDATE",
	4: "DELETE",
}

var TxnOperation_value = map[string]int32{
	"UNDEFINED": 0,
	"VALIDATE":  1,
	"CREATE":    2,
	"UPDATE":    3,
	"DELETE":    4,
}

func (x TxnOperation) String() string {
	return proto.EnumName(TxnOperation_name, int32(x))
}

func (TxnOperation) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_676435f6cd35671d, []int{1}
}

type ValueStatus struct {
	Key           string       `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	State         ValueState   `protobuf:"varint,2,opt,name=state,proto3,enum=ligato.vpp_agent.kvscheduler.ValueState" json:"state,omitempty"`
	Error         string       `protobuf:"bytes,3,opt,name=error,proto3" json:"error,omitempty"`
	LastOperation TxnOperation `protobuf:"varint,4,opt,name=last_operation,json=lastOperation,proto3,enum=ligato.vpp_agent.kvscheduler.TxnOperation" json:"last_operation,omitempty"`
	// - for invalid value, details is a list of invalid fields
	// - for pending value, details is a list of missing dependencies (labels)
	Details              []string `protobuf:"bytes,5,rep,name=details,proto3" json:"details,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ValueStatus) Reset()         { *m = ValueStatus{} }
func (m *ValueStatus) String() string { return proto.CompactTextString(m) }
func (*ValueStatus) ProtoMessage()    {}
func (*ValueStatus) Descriptor() ([]byte, []int) {
	return fileDescriptor_676435f6cd35671d, []int{0}
}

func (m *ValueStatus) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ValueStatus.Unmarshal(m, b)
}
func (m *ValueStatus) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ValueStatus.Marshal(b, m, deterministic)
}
func (m *ValueStatus) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ValueStatus.Merge(m, src)
}
func (m *ValueStatus) XXX_Size() int {
	return xxx_messageInfo_ValueStatus.Size(m)
}
func (m *ValueStatus) XXX_DiscardUnknown() {
	xxx_messageInfo_ValueStatus.DiscardUnknown(m)
}

var xxx_messageInfo_ValueStatus proto.InternalMessageInfo

func (m *ValueStatus) GetKey() string {
	if m != nil {
		return m.Key
	}
	return ""
}

func (m *ValueStatus) GetState() ValueState {
	if m != nil {
		return m.State
	}
	return ValueState_NONEXISTENT
}

func (m *ValueStatus) GetError() string {
	if m != nil {
		return m.Error
	}
	return ""
}

func (m *ValueStatus) GetLastOperation() TxnOperation {
	if m != nil {
		return m.LastOperation
	}
	return TxnOperation_UNDEFINED
}

func (m *ValueStatus) GetDetails() []string {
	if m != nil {
		return m.Details
	}
	return nil
}

type BaseValueStatus struct {
	Value                *ValueStatus   `protobuf:"bytes,1,opt,name=value,proto3" json:"value,omitempty"`
	DerivedValues        []*ValueStatus `protobuf:"bytes,2,rep,name=derived_values,json=derivedValues,proto3" json:"derived_values,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *BaseValueStatus) Reset()         { *m = BaseValueStatus{} }
func (m *BaseValueStatus) String() string { return proto.CompactTextString(m) }
func (*BaseValueStatus) ProtoMessage()    {}
func (*BaseValueStatus) Descriptor() ([]byte, []int) {
	return fileDescriptor_676435f6cd35671d, []int{1}
}

func (m *BaseValueStatus) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_BaseValueStatus.Unmarshal(m, b)
}
func (m *BaseValueStatus) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_BaseValueStatus.Marshal(b, m, deterministic)
}
func (m *BaseValueStatus) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BaseValueStatus.Merge(m, src)
}
func (m *BaseValueStatus) XXX_Size() int {
	return xxx_messageInfo_BaseValueStatus.Size(m)
}
func (m *BaseValueStatus) XXX_DiscardUnknown() {
	xxx_messageInfo_BaseValueStatus.DiscardUnknown(m)
}

var xxx_messageInfo_BaseValueStatus proto.InternalMessageInfo

func (m *BaseValueStatus) GetValue() *ValueStatus {
	if m != nil {
		return m.Value
	}
	return nil
}

func (m *BaseValueStatus) GetDerivedValues() []*ValueStatus {
	if m != nil {
		return m.DerivedValues
	}
	return nil
}

func init() {
	proto.RegisterEnum("ligato.vpp_agent.kvscheduler.ValueState", ValueState_name, ValueState_value)
	proto.RegisterEnum("ligato.vpp_agent.kvscheduler.TxnOperation", TxnOperation_name, TxnOperation_value)
	proto.RegisterType((*ValueStatus)(nil), "ligato.vpp_agent.kvscheduler.ValueStatus")
	proto.RegisterType((*BaseValueStatus)(nil), "ligato.vpp_agent.kvscheduler.BaseValueStatus")
}

func init() {
	proto.RegisterFile("ligato/vpp-agent/kvscheduler/value_status.proto", fileDescriptor_676435f6cd35671d)
}

var fileDescriptor_676435f6cd35671d = []byte{
	// 460 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x92, 0x4f, 0x8f, 0xd2, 0x40,
	0x14, 0xc0, 0xb7, 0x94, 0xc2, 0xf2, 0x58, 0xd8, 0x71, 0xe2, 0xa1, 0x07, 0x0f, 0x64, 0x4f, 0x48,
	0xb2, 0x6d, 0x82, 0x77, 0x0d, 0x30, 0xc3, 0x66, 0x12, 0x98, 0xe2, 0xb4, 0x10, 0xf5, 0x42, 0xaa,
	0x4c, 0xb0, 0xd9, 0x86, 0x36, 0x9d, 0xb6, 0xd1, 0xef, 0xe3, 0xd1, 0x4f, 0xe5, 0x27, 0x31, 0x33,
	0xc8, 0xda, 0xd3, 0x46, 0x6f, 0xef, 0xcf, 0xfc, 0x7e, 0xcd, 0x7b, 0x7d, 0xe0, 0xa7, 0xc9, 0x31,
	0x2e, 0x33, 0xbf, 0xce, 0xf3, 0xfb, 0xf8, 0x28, 0x4f, 0xa5, 0xff, 0x58, 0xab, 0x2f, 0x5f, 0xe5,
	0xa1, 0x4a, 0x65, 0xe1, 0xd7, 0x71, 0x5a, 0xc9, 0xbd, 0x2a, 0xe3, 0xb2, 0x52, 0x5e, 0x5e, 0x64,
	0x65, 0x86, 0x5f, 0x9d, 0x01, 0xaf, 0xce, 0xf3, 0xbd, 0x01, 0xbc, 0x06, 0x70, 0xf7, 0xcb, 0x82,
	0xfe, 0x4e, 0x43, 0xa1, 0x61, 0x30, 0x02, 0xfb, 0x51, 0x7e, 0x77, 0xad, 0x91, 0x35, 0xee, 0x09,
	0x1d, 0xe2, 0xb7, 0xe0, 0x68, 0x9f, 0x74, 0x5b, 0x23, 0x6b, 0x3c, 0x9c, 0x8e, 0xbd, 0xe7, 0x7c,
	0xde, 0x93, 0x4b, 0x8a, 0x33, 0x86, 0x5f, 0x82, 0x23, 0x8b, 0x22, 0x2b, 0x5c, 0xdb, 0x38, 0xcf,
	0x09, 0x7e, 0x0f, 0xc3, 0x34, 0x56, 0xe5, 0x3e, 0xcb, 0x65, 0x11, 0x97, 0x49, 0x76, 0x72, 0xdb,
	0x46, 0x3f, 0x79, 0x5e, 0x1f, 0x7d, 0x3b, 0x05, 0x17, 0x42, 0x0c, 0xb4, 0xe1, 0x29, 0xc5, 0x2e,
	0x74, 0x0f, 0xb2, 0x8c, 0x93, 0x54, 0xb9, 0xce, 0xc8, 0x1e, 0xf7, 0xc4, 0x25, 0xbd, 0xfb, 0x61,
	0xc1, 0xed, 0x3c, 0x56, 0xb2, 0x39, 0xe8, 0x3b, 0x70, 0xcc, 0xb2, 0xcc, 0xa8, 0xfd, 0xe9, 0xeb,
	0x7f, 0x1c, 0xab, 0x52, 0xe2, 0xcc, 0xe1, 0x0d, 0x0c, 0x0f, 0xb2, 0x48, 0x6a, 0x79, 0xd8, 0x9b,
	0x82, 0x72, 0x5b, 0x23, 0xfb, 0xff, 0x4c, 0x83, 0x3f, 0x02, 0x53, 0x53, 0x93, 0x9f, 0x16, 0xc0,
	0xdf, 0xfd, 0xe1, 0x5b, 0xe8, 0xf3, 0x80, 0xd3, 0x0f, 0x2c, 0x8c, 0x28, 0x8f, 0xd0, 0x15, 0xee,
	0x43, 0x77, 0xcd, 0xc2, 0x90, 0xf1, 0x07, 0x64, 0xe1, 0x17, 0x30, 0xd8, 0x72, 0xb6, 0xde, 0xac,
	0xe8, 0x9a, 0xf2, 0x88, 0x12, 0xd4, 0xd2, 0x7d, 0x41, 0xd7, 0xc1, 0x8e, 0x12, 0x64, 0xe3, 0x21,
	0xc0, 0x22, 0xe0, 0x4b, 0xf6, 0xb0, 0x15, 0x94, 0xa0, 0x36, 0xbe, 0x81, 0xeb, 0x60, 0x1e, 0xcd,
	0x18, 0xa7, 0x04, 0x39, 0xba, 0x4b, 0x58, 0xb8, 0x08, 0x76, 0x54, 0x77, 0x3b, 0x1a, 0xdd, 0x50,
	0x4e, 0xb4, 0xba, 0xab, 0x13, 0xc6, 0x77, 0xb3, 0x15, 0x23, 0xe8, 0x1a, 0x03, 0x74, 0x96, 0x33,
	0xb6, 0xa2, 0x04, 0xf5, 0xb4, 0x43, 0xd0, 0x48, 0x7c, 0xd4, 0xcf, 0x60, 0x12, 0xc0, 0x4d, 0xf3,
	0x77, 0xe0, 0x01, 0xf4, 0xb6, 0x9c, 0xd0, 0xa5, 0xf9, 0xc4, 0x95, 0x7e, 0x6c, 0x1c, 0xb3, 0x88,
	0x22, 0x4b, 0x6b, 0x16, 0x82, 0xea, 0xb8, 0xa5, 0xe3, 0xed, 0xc6, 0xd4, 0x6d, 0x1d, 0x13, 0xba,
	0xa2, 0x11, 0x45, 0xed, 0xb9, 0xff, 0xe9, 0xfe, 0x98, 0x5d, 0xb6, 0x97, 0x34, 0x4f, 0xbc, 0x9e,
	0xfa, 0x79, 0x5a, 0x1d, 0x93, 0x93, 0x6a, 0x5e, 0xfb, 0xe7, 0x8e, 0xb9, 0xf0, 0x37, 0xbf, 0x03,
	0x00, 0x00, 0xff, 0xff, 0x46, 0x86, 0x06, 0x92, 0x14, 0x03, 0x00, 0x00,
}
