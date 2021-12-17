// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: cosmos/group/v1beta1/events.proto

package group

import (
	fmt "fmt"
	_ "github.com/cosmos/cosmos-proto"
	proto "github.com/gogo/protobuf/proto"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// EventCreateGroup is an event emitted when a group is created.
type EventCreateGroup struct {
	// group_id is the unique ID of the group.
	GroupId uint64 `protobuf:"varint,1,opt,name=group_id,json=groupId,proto3" json:"group_id,omitempty"`
}

func (m *EventCreateGroup) Reset()         { *m = EventCreateGroup{} }
func (m *EventCreateGroup) String() string { return proto.CompactTextString(m) }
func (*EventCreateGroup) ProtoMessage()    {}
func (*EventCreateGroup) Descriptor() ([]byte, []int) {
	return fileDescriptor_7879e051fb126fc0, []int{0}
}
func (m *EventCreateGroup) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *EventCreateGroup) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_EventCreateGroup.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *EventCreateGroup) XXX_Merge(src proto.Message) {
	xxx_messageInfo_EventCreateGroup.Merge(m, src)
}
func (m *EventCreateGroup) XXX_Size() int {
	return m.Size()
}
func (m *EventCreateGroup) XXX_DiscardUnknown() {
	xxx_messageInfo_EventCreateGroup.DiscardUnknown(m)
}

var xxx_messageInfo_EventCreateGroup proto.InternalMessageInfo

func (m *EventCreateGroup) GetGroupId() uint64 {
	if m != nil {
		return m.GroupId
	}
	return 0
}

// EventUpdateGroup is an event emitted when a group is updated.
type EventUpdateGroup struct {
	// group_id is the unique ID of the group.
	GroupId uint64 `protobuf:"varint,1,opt,name=group_id,json=groupId,proto3" json:"group_id,omitempty"`
}

func (m *EventUpdateGroup) Reset()         { *m = EventUpdateGroup{} }
func (m *EventUpdateGroup) String() string { return proto.CompactTextString(m) }
func (*EventUpdateGroup) ProtoMessage()    {}
func (*EventUpdateGroup) Descriptor() ([]byte, []int) {
	return fileDescriptor_7879e051fb126fc0, []int{1}
}
func (m *EventUpdateGroup) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *EventUpdateGroup) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_EventUpdateGroup.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *EventUpdateGroup) XXX_Merge(src proto.Message) {
	xxx_messageInfo_EventUpdateGroup.Merge(m, src)
}
func (m *EventUpdateGroup) XXX_Size() int {
	return m.Size()
}
func (m *EventUpdateGroup) XXX_DiscardUnknown() {
	xxx_messageInfo_EventUpdateGroup.DiscardUnknown(m)
}

var xxx_messageInfo_EventUpdateGroup proto.InternalMessageInfo

func (m *EventUpdateGroup) GetGroupId() uint64 {
	if m != nil {
		return m.GroupId
	}
	return 0
}

// EventCreateGroupAccount is an event emitted when a group account is created.
type EventCreateGroupAccount struct {
	// address is the address of the group account.
	Address string `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
}

func (m *EventCreateGroupAccount) Reset()         { *m = EventCreateGroupAccount{} }
func (m *EventCreateGroupAccount) String() string { return proto.CompactTextString(m) }
func (*EventCreateGroupAccount) ProtoMessage()    {}
func (*EventCreateGroupAccount) Descriptor() ([]byte, []int) {
	return fileDescriptor_7879e051fb126fc0, []int{2}
}
func (m *EventCreateGroupAccount) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *EventCreateGroupAccount) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_EventCreateGroupAccount.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *EventCreateGroupAccount) XXX_Merge(src proto.Message) {
	xxx_messageInfo_EventCreateGroupAccount.Merge(m, src)
}
func (m *EventCreateGroupAccount) XXX_Size() int {
	return m.Size()
}
func (m *EventCreateGroupAccount) XXX_DiscardUnknown() {
	xxx_messageInfo_EventCreateGroupAccount.DiscardUnknown(m)
}

var xxx_messageInfo_EventCreateGroupAccount proto.InternalMessageInfo

func (m *EventCreateGroupAccount) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

// EventUpdateGroupAccount is an event emitted when a group account is updated.
type EventUpdateGroupAccount struct {
	// address is the address of the group account.
	Address string `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
}

func (m *EventUpdateGroupAccount) Reset()         { *m = EventUpdateGroupAccount{} }
func (m *EventUpdateGroupAccount) String() string { return proto.CompactTextString(m) }
func (*EventUpdateGroupAccount) ProtoMessage()    {}
func (*EventUpdateGroupAccount) Descriptor() ([]byte, []int) {
	return fileDescriptor_7879e051fb126fc0, []int{3}
}
func (m *EventUpdateGroupAccount) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *EventUpdateGroupAccount) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_EventUpdateGroupAccount.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *EventUpdateGroupAccount) XXX_Merge(src proto.Message) {
	xxx_messageInfo_EventUpdateGroupAccount.Merge(m, src)
}
func (m *EventUpdateGroupAccount) XXX_Size() int {
	return m.Size()
}
func (m *EventUpdateGroupAccount) XXX_DiscardUnknown() {
	xxx_messageInfo_EventUpdateGroupAccount.DiscardUnknown(m)
}

var xxx_messageInfo_EventUpdateGroupAccount proto.InternalMessageInfo

func (m *EventUpdateGroupAccount) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

// EventCreateProposal is an event emitted when a proposal is created.
type EventCreateProposal struct {
	// proposal_id is the unique ID of the proposal.
	ProposalId uint64 `protobuf:"varint,1,opt,name=proposal_id,json=proposalId,proto3" json:"proposal_id,omitempty"`
}

func (m *EventCreateProposal) Reset()         { *m = EventCreateProposal{} }
func (m *EventCreateProposal) String() string { return proto.CompactTextString(m) }
func (*EventCreateProposal) ProtoMessage()    {}
func (*EventCreateProposal) Descriptor() ([]byte, []int) {
	return fileDescriptor_7879e051fb126fc0, []int{4}
}
func (m *EventCreateProposal) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *EventCreateProposal) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_EventCreateProposal.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *EventCreateProposal) XXX_Merge(src proto.Message) {
	xxx_messageInfo_EventCreateProposal.Merge(m, src)
}
func (m *EventCreateProposal) XXX_Size() int {
	return m.Size()
}
func (m *EventCreateProposal) XXX_DiscardUnknown() {
	xxx_messageInfo_EventCreateProposal.DiscardUnknown(m)
}

var xxx_messageInfo_EventCreateProposal proto.InternalMessageInfo

func (m *EventCreateProposal) GetProposalId() uint64 {
	if m != nil {
		return m.ProposalId
	}
	return 0
}

// EventVote is an event emitted when a voter votes on a proposal.
type EventVote struct {
	// proposal_id is the unique ID of the proposal.
	ProposalId uint64 `protobuf:"varint,1,opt,name=proposal_id,json=proposalId,proto3" json:"proposal_id,omitempty"`
}

func (m *EventVote) Reset()         { *m = EventVote{} }
func (m *EventVote) String() string { return proto.CompactTextString(m) }
func (*EventVote) ProtoMessage()    {}
func (*EventVote) Descriptor() ([]byte, []int) {
	return fileDescriptor_7879e051fb126fc0, []int{5}
}
func (m *EventVote) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *EventVote) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_EventVote.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *EventVote) XXX_Merge(src proto.Message) {
	xxx_messageInfo_EventVote.Merge(m, src)
}
func (m *EventVote) XXX_Size() int {
	return m.Size()
}
func (m *EventVote) XXX_DiscardUnknown() {
	xxx_messageInfo_EventVote.DiscardUnknown(m)
}

var xxx_messageInfo_EventVote proto.InternalMessageInfo

func (m *EventVote) GetProposalId() uint64 {
	if m != nil {
		return m.ProposalId
	}
	return 0
}

// EventExec is an event emitted when a proposal is executed.
type EventExec struct {
	// proposal_id is the unique ID of the proposal.
	ProposalId uint64 `protobuf:"varint,1,opt,name=proposal_id,json=proposalId,proto3" json:"proposal_id,omitempty"`
}

func (m *EventExec) Reset()         { *m = EventExec{} }
func (m *EventExec) String() string { return proto.CompactTextString(m) }
func (*EventExec) ProtoMessage()    {}
func (*EventExec) Descriptor() ([]byte, []int) {
	return fileDescriptor_7879e051fb126fc0, []int{6}
}
func (m *EventExec) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *EventExec) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_EventExec.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *EventExec) XXX_Merge(src proto.Message) {
	xxx_messageInfo_EventExec.Merge(m, src)
}
func (m *EventExec) XXX_Size() int {
	return m.Size()
}
func (m *EventExec) XXX_DiscardUnknown() {
	xxx_messageInfo_EventExec.DiscardUnknown(m)
}

var xxx_messageInfo_EventExec proto.InternalMessageInfo

func (m *EventExec) GetProposalId() uint64 {
	if m != nil {
		return m.ProposalId
	}
	return 0
}

func init() {
	proto.RegisterType((*EventCreateGroup)(nil), "cosmos.group.v1beta1.EventCreateGroup")
	proto.RegisterType((*EventUpdateGroup)(nil), "cosmos.group.v1beta1.EventUpdateGroup")
	proto.RegisterType((*EventCreateGroupAccount)(nil), "cosmos.group.v1beta1.EventCreateGroupAccount")
	proto.RegisterType((*EventUpdateGroupAccount)(nil), "cosmos.group.v1beta1.EventUpdateGroupAccount")
	proto.RegisterType((*EventCreateProposal)(nil), "cosmos.group.v1beta1.EventCreateProposal")
	proto.RegisterType((*EventVote)(nil), "cosmos.group.v1beta1.EventVote")
	proto.RegisterType((*EventExec)(nil), "cosmos.group.v1beta1.EventExec")
}

func init() { proto.RegisterFile("cosmos/group/v1beta1/events.proto", fileDescriptor_7879e051fb126fc0) }

var fileDescriptor_7879e051fb126fc0 = []byte{
	// 281 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x52, 0x4c, 0xce, 0x2f, 0xce,
	0xcd, 0x2f, 0xd6, 0x4f, 0x2f, 0xca, 0x2f, 0x2d, 0xd0, 0x2f, 0x33, 0x4c, 0x4a, 0x2d, 0x49, 0x34,
	0xd4, 0x4f, 0x2d, 0x4b, 0xcd, 0x2b, 0x29, 0xd6, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x12, 0x81,
	0x28, 0xd1, 0x03, 0x2b, 0xd1, 0x83, 0x2a, 0x91, 0x92, 0x84, 0x88, 0xc6, 0x83, 0xd5, 0xe8, 0x43,
	0x95, 0x80, 0x39, 0x4a, 0xba, 0x5c, 0x02, 0xae, 0x20, 0x03, 0x9c, 0x8b, 0x52, 0x13, 0x4b, 0x52,
	0xdd, 0x41, 0xda, 0x84, 0x24, 0xb9, 0x38, 0xc0, 0xfa, 0xe3, 0x33, 0x53, 0x24, 0x18, 0x15, 0x18,
	0x35, 0x58, 0x82, 0xd8, 0xc1, 0x7c, 0xcf, 0x14, 0xb8, 0xf2, 0xd0, 0x82, 0x14, 0x62, 0x94, 0xfb,
	0x72, 0x89, 0xa3, 0x9b, 0xee, 0x98, 0x9c, 0x9c, 0x5f, 0x9a, 0x57, 0x22, 0x64, 0xc4, 0xc5, 0x9e,
	0x98, 0x92, 0x52, 0x94, 0x5a, 0x5c, 0x0c, 0xd6, 0xc4, 0xe9, 0x24, 0x71, 0x69, 0x8b, 0x2e, 0xcc,
	0xf9, 0x8e, 0x10, 0x99, 0xe0, 0x92, 0xa2, 0xcc, 0xbc, 0xf4, 0x20, 0x98, 0x42, 0xb8, 0x71, 0x48,
	0xb6, 0x53, 0x62, 0x9c, 0x19, 0x97, 0x30, 0x92, 0xeb, 0x02, 0x8a, 0xf2, 0x0b, 0xf2, 0x8b, 0x13,
	0x73, 0x84, 0xe4, 0xb9, 0xb8, 0x0b, 0xa0, 0x6c, 0x84, 0x97, 0xb8, 0x60, 0x42, 0x9e, 0x29, 0x4a,
	0x3a, 0x5c, 0x9c, 0x60, 0x7d, 0x61, 0xf9, 0x25, 0xa9, 0xc4, 0xab, 0x76, 0xad, 0x48, 0x4d, 0x26,
	0xa8, 0xda, 0xc9, 0xee, 0xc4, 0x23, 0x39, 0xc6, 0x0b, 0x8f, 0xe4, 0x18, 0x1f, 0x3c, 0x92, 0x63,
	0x9c, 0xf0, 0x58, 0x8e, 0xe1, 0xc2, 0x63, 0x39, 0x86, 0x1b, 0x8f, 0xe5, 0x18, 0xa2, 0x54, 0xd2,
	0x33, 0x4b, 0x32, 0x4a, 0x93, 0xf4, 0x92, 0xf3, 0x73, 0xa1, 0x51, 0x08, 0xa5, 0x74, 0x8b, 0x53,
	0xb2, 0xf5, 0x2b, 0x20, 0xa9, 0x22, 0x89, 0x0d, 0x1c, 0xad, 0xc6, 0x80, 0x00, 0x00, 0x00, 0xff,
	0xff, 0x90, 0x54, 0xa1, 0x1a, 0x2c, 0x02, 0x00, 0x00,
}

func (m *EventCreateGroup) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *EventCreateGroup) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *EventCreateGroup) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.GroupId != 0 {
		i = encodeVarintEvents(dAtA, i, uint64(m.GroupId))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *EventUpdateGroup) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *EventUpdateGroup) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *EventUpdateGroup) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.GroupId != 0 {
		i = encodeVarintEvents(dAtA, i, uint64(m.GroupId))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *EventCreateGroupAccount) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *EventCreateGroupAccount) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *EventCreateGroupAccount) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Address) > 0 {
		i -= len(m.Address)
		copy(dAtA[i:], m.Address)
		i = encodeVarintEvents(dAtA, i, uint64(len(m.Address)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *EventUpdateGroupAccount) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *EventUpdateGroupAccount) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *EventUpdateGroupAccount) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Address) > 0 {
		i -= len(m.Address)
		copy(dAtA[i:], m.Address)
		i = encodeVarintEvents(dAtA, i, uint64(len(m.Address)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *EventCreateProposal) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *EventCreateProposal) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *EventCreateProposal) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.ProposalId != 0 {
		i = encodeVarintEvents(dAtA, i, uint64(m.ProposalId))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *EventVote) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *EventVote) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *EventVote) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.ProposalId != 0 {
		i = encodeVarintEvents(dAtA, i, uint64(m.ProposalId))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *EventExec) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *EventExec) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *EventExec) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.ProposalId != 0 {
		i = encodeVarintEvents(dAtA, i, uint64(m.ProposalId))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func encodeVarintEvents(dAtA []byte, offset int, v uint64) int {
	offset -= sovEvents(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *EventCreateGroup) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.GroupId != 0 {
		n += 1 + sovEvents(uint64(m.GroupId))
	}
	return n
}

func (m *EventUpdateGroup) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.GroupId != 0 {
		n += 1 + sovEvents(uint64(m.GroupId))
	}
	return n
}

func (m *EventCreateGroupAccount) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Address)
	if l > 0 {
		n += 1 + l + sovEvents(uint64(l))
	}
	return n
}

func (m *EventUpdateGroupAccount) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Address)
	if l > 0 {
		n += 1 + l + sovEvents(uint64(l))
	}
	return n
}

func (m *EventCreateProposal) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.ProposalId != 0 {
		n += 1 + sovEvents(uint64(m.ProposalId))
	}
	return n
}

func (m *EventVote) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.ProposalId != 0 {
		n += 1 + sovEvents(uint64(m.ProposalId))
	}
	return n
}

func (m *EventExec) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.ProposalId != 0 {
		n += 1 + sovEvents(uint64(m.ProposalId))
	}
	return n
}

func sovEvents(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozEvents(x uint64) (n int) {
	return sovEvents(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *EventCreateGroup) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowEvents
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: EventCreateGroup: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: EventCreateGroup: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field GroupId", wireType)
			}
			m.GroupId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowEvents
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.GroupId |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipEvents(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthEvents
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *EventUpdateGroup) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowEvents
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: EventUpdateGroup: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: EventUpdateGroup: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field GroupId", wireType)
			}
			m.GroupId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowEvents
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.GroupId |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipEvents(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthEvents
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *EventCreateGroupAccount) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowEvents
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: EventCreateGroupAccount: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: EventCreateGroupAccount: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Address", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowEvents
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthEvents
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthEvents
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Address = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipEvents(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthEvents
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *EventUpdateGroupAccount) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowEvents
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: EventUpdateGroupAccount: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: EventUpdateGroupAccount: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Address", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowEvents
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthEvents
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthEvents
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Address = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipEvents(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthEvents
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *EventCreateProposal) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowEvents
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: EventCreateProposal: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: EventCreateProposal: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ProposalId", wireType)
			}
			m.ProposalId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowEvents
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ProposalId |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipEvents(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthEvents
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *EventVote) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowEvents
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: EventVote: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: EventVote: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ProposalId", wireType)
			}
			m.ProposalId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowEvents
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ProposalId |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipEvents(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthEvents
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *EventExec) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowEvents
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: EventExec: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: EventExec: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ProposalId", wireType)
			}
			m.ProposalId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowEvents
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ProposalId |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipEvents(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthEvents
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipEvents(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowEvents
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowEvents
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowEvents
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthEvents
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupEvents
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthEvents
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthEvents        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowEvents          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupEvents = fmt.Errorf("proto: unexpected end of group")
)
