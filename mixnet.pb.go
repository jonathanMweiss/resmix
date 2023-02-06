// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.21.0-devel
// 	protoc        v3.7.1
// source: mixnet.proto

package resmix

import (
	context "context"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

type NewRoundRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Round uint32 `protobuf:"varint,1,opt,name=round,proto3" json:"round,omitempty"`
}

func (x *NewRoundRequest) Reset() {
	*x = NewRoundRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mixnet_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NewRoundRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NewRoundRequest) ProtoMessage() {}

func (x *NewRoundRequest) ProtoReflect() protoreflect.Message {
	mi := &file_mixnet_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NewRoundRequest.ProtoReflect.Descriptor instead.
func (*NewRoundRequest) Descriptor() ([]byte, []int) {
	return file_mixnet_proto_rawDescGZIP(), []int{0}
}

func (x *NewRoundRequest) GetRound() uint32 {
	if x != nil {
		return x.Round
	}
	return 0
}

type NewRoundResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *NewRoundResponse) Reset() {
	*x = NewRoundResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mixnet_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NewRoundResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NewRoundResponse) ProtoMessage() {}

func (x *NewRoundResponse) ProtoReflect() protoreflect.Message {
	mi := &file_mixnet_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NewRoundResponse.ProtoReflect.Descriptor instead.
func (*NewRoundResponse) Descriptor() ([]byte, []int) {
	return file_mixnet_proto_rawDescGZIP(), []int{1}
}

type EndRoundRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Round uint32 `protobuf:"varint,1,opt,name=round,proto3" json:"round,omitempty"`
}

func (x *EndRoundRequest) Reset() {
	*x = EndRoundRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mixnet_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EndRoundRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EndRoundRequest) ProtoMessage() {}

func (x *EndRoundRequest) ProtoReflect() protoreflect.Message {
	mi := &file_mixnet_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EndRoundRequest.ProtoReflect.Descriptor instead.
func (*EndRoundRequest) Descriptor() ([]byte, []int) {
	return file_mixnet_proto_rawDescGZIP(), []int{2}
}

func (x *EndRoundRequest) GetRound() uint32 {
	if x != nil {
		return x.Round
	}
	return 0
}

type EndRoundResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *EndRoundResponse) Reset() {
	*x = EndRoundResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mixnet_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EndRoundResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EndRoundResponse) ProtoMessage() {}

func (x *EndRoundResponse) ProtoReflect() protoreflect.Message {
	mi := &file_mixnet_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EndRoundResponse.ProtoReflect.Descriptor instead.
func (*EndRoundResponse) Descriptor() ([]byte, []int) {
	return file_mixnet_proto_rawDescGZIP(), []int{3}
}

type Messages struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Messages [][]byte `protobuf:"bytes,1,rep,name=messages,proto3" json:"messages,omitempty"`
	// represents the logical mix (which layer) that sent the messages.
	PhysicalSender  []byte `protobuf:"bytes,2,opt,name=physicalSender,proto3" json:"physicalSender,omitempty"`
	LogicalSender   []byte `protobuf:"bytes,3,opt,name=logicalSender,proto3" json:"logicalSender,omitempty"`
	LogicalReceiver []byte `protobuf:"bytes,4,opt,name=logicalReceiver,proto3" json:"logicalReceiver,omitempty"`
}

func (x *Messages) Reset() {
	*x = Messages{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mixnet_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Messages) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Messages) ProtoMessage() {}

func (x *Messages) ProtoReflect() protoreflect.Message {
	mi := &file_mixnet_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Messages.ProtoReflect.Descriptor instead.
func (*Messages) Descriptor() ([]byte, []int) {
	return file_mixnet_proto_rawDescGZIP(), []int{4}
}

func (x *Messages) GetMessages() [][]byte {
	if x != nil {
		return x.Messages
	}
	return nil
}

func (x *Messages) GetPhysicalSender() []byte {
	if x != nil {
		return x.PhysicalSender
	}
	return nil
}

func (x *Messages) GetLogicalSender() []byte {
	if x != nil {
		return x.LogicalSender
	}
	return nil
}

func (x *Messages) GetLogicalReceiver() []byte {
	if x != nil {
		return x.LogicalReceiver
	}
	return nil
}

type AddMessagesRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Round    uint32      `protobuf:"varint,1,opt,name=round,proto3" json:"round,omitempty"`
	Messages []*Messages `protobuf:"bytes,2,rep,name=messages,proto3" json:"messages,omitempty"`
}

func (x *AddMessagesRequest) Reset() {
	*x = AddMessagesRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mixnet_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AddMessagesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AddMessagesRequest) ProtoMessage() {}

func (x *AddMessagesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_mixnet_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AddMessagesRequest.ProtoReflect.Descriptor instead.
func (*AddMessagesRequest) Descriptor() ([]byte, []int) {
	return file_mixnet_proto_rawDescGZIP(), []int{5}
}

func (x *AddMessagesRequest) GetRound() uint32 {
	if x != nil {
		return x.Round
	}
	return 0
}

func (x *AddMessagesRequest) GetMessages() []*Messages {
	if x != nil {
		return x.Messages
	}
	return nil
}

type Vote struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// sig over some ID. the sig is created using some part of a DKG share.
	Signature      []byte `protobuf:"bytes,1,opt,name=signature,proto3" json:"signature,omitempty"`
	SignerPosition uint32 `protobuf:"varint,2,opt,name=signerPosition,proto3" json:"signerPosition,omitempty"`
}

func (x *Vote) Reset() {
	*x = Vote{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mixnet_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Vote) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Vote) ProtoMessage() {}

func (x *Vote) ProtoReflect() protoreflect.Message {
	mi := &file_mixnet_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Vote.ProtoReflect.Descriptor instead.
func (*Vote) Descriptor() ([]byte, []int) {
	return file_mixnet_proto_rawDescGZIP(), []int{6}
}

func (x *Vote) GetSignature() []byte {
	if x != nil {
		return x.Signature
	}
	return nil
}

func (x *Vote) GetSignerPosition() uint32 {
	if x != nil {
		return x.SignerPosition
	}
	return 0
}

type KeyReleaseMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Votes []*Vote `protobuf:"bytes,1,rep,name=votes,proto3" json:"votes,omitempty"`
	// all votes are targeting this id.
	ServerId []byte `protobuf:"bytes,2,opt,name=serverId,proto3" json:"serverId,omitempty"`
}

func (x *KeyReleaseMessage) Reset() {
	*x = KeyReleaseMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mixnet_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *KeyReleaseMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*KeyReleaseMessage) ProtoMessage() {}

func (x *KeyReleaseMessage) ProtoReflect() protoreflect.Message {
	mi := &file_mixnet_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use KeyReleaseMessage.ProtoReflect.Descriptor instead.
func (*KeyReleaseMessage) Descriptor() ([]byte, []int) {
	return file_mixnet_proto_rawDescGZIP(), []int{7}
}

func (x *KeyReleaseMessage) GetVotes() []*Vote {
	if x != nil {
		return x.Votes
	}
	return nil
}

func (x *KeyReleaseMessage) GetServerId() []byte {
	if x != nil {
		return x.ServerId
	}
	return nil
}

var File_mixnet_proto protoreflect.FileDescriptor

var file_mixnet_proto_rawDesc = []byte{
	0x0a, 0x0c, 0x6d, 0x69, 0x78, 0x6e, 0x65, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06,
	0x72, 0x65, 0x73, 0x6d, 0x69, 0x78, 0x22, 0x27, 0x0a, 0x0f, 0x4e, 0x65, 0x77, 0x52, 0x6f, 0x75,
	0x6e, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x72, 0x6f, 0x75,
	0x6e, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x05, 0x72, 0x6f, 0x75, 0x6e, 0x64, 0x22,
	0x12, 0x0a, 0x10, 0x4e, 0x65, 0x77, 0x52, 0x6f, 0x75, 0x6e, 0x64, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x27, 0x0a, 0x0f, 0x45, 0x6e, 0x64, 0x52, 0x6f, 0x75, 0x6e, 0x64, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x72, 0x6f, 0x75, 0x6e, 0x64, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x05, 0x72, 0x6f, 0x75, 0x6e, 0x64, 0x22, 0x12, 0x0a, 0x10,
	0x45, 0x6e, 0x64, 0x52, 0x6f, 0x75, 0x6e, 0x64, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x22, 0x9e, 0x01, 0x0a, 0x08, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x12, 0x1a, 0x0a,
	0x08, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0c, 0x52,
	0x08, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x12, 0x26, 0x0a, 0x0e, 0x70, 0x68, 0x79,
	0x73, 0x69, 0x63, 0x61, 0x6c, 0x53, 0x65, 0x6e, 0x64, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0c, 0x52, 0x0e, 0x70, 0x68, 0x79, 0x73, 0x69, 0x63, 0x61, 0x6c, 0x53, 0x65, 0x6e, 0x64, 0x65,
	0x72, 0x12, 0x24, 0x0a, 0x0d, 0x6c, 0x6f, 0x67, 0x69, 0x63, 0x61, 0x6c, 0x53, 0x65, 0x6e, 0x64,
	0x65, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0d, 0x6c, 0x6f, 0x67, 0x69, 0x63, 0x61,
	0x6c, 0x53, 0x65, 0x6e, 0x64, 0x65, 0x72, 0x12, 0x28, 0x0a, 0x0f, 0x6c, 0x6f, 0x67, 0x69, 0x63,
	0x61, 0x6c, 0x52, 0x65, 0x63, 0x65, 0x69, 0x76, 0x65, 0x72, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0c,
	0x52, 0x0f, 0x6c, 0x6f, 0x67, 0x69, 0x63, 0x61, 0x6c, 0x52, 0x65, 0x63, 0x65, 0x69, 0x76, 0x65,
	0x72, 0x22, 0x58, 0x0a, 0x12, 0x41, 0x64, 0x64, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x72, 0x6f, 0x75, 0x6e, 0x64,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x05, 0x72, 0x6f, 0x75, 0x6e, 0x64, 0x12, 0x2c, 0x0a,
	0x08, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x10, 0x2e, 0x72, 0x65, 0x73, 0x6d, 0x69, 0x78, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x73, 0x52, 0x08, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x22, 0x4c, 0x0a, 0x04, 0x56,
	0x6f, 0x74, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72,
	0x65, 0x12, 0x26, 0x0a, 0x0e, 0x73, 0x69, 0x67, 0x6e, 0x65, 0x72, 0x50, 0x6f, 0x73, 0x69, 0x74,
	0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x0e, 0x73, 0x69, 0x67, 0x6e, 0x65,
	0x72, 0x50, 0x6f, 0x73, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0x53, 0x0a, 0x11, 0x4b, 0x65, 0x79,
	0x52, 0x65, 0x6c, 0x65, 0x61, 0x73, 0x65, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x22,
	0x0a, 0x05, 0x76, 0x6f, 0x74, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0c, 0x2e,
	0x72, 0x65, 0x73, 0x6d, 0x69, 0x78, 0x2e, 0x56, 0x6f, 0x74, 0x65, 0x52, 0x05, 0x76, 0x6f, 0x74,
	0x65, 0x73, 0x12, 0x1a, 0x0a, 0x08, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x49, 0x64, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x08, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x49, 0x64, 0x32, 0x87,
	0x01, 0x0a, 0x03, 0x4d, 0x69, 0x78, 0x12, 0x3f, 0x0a, 0x08, 0x4e, 0x65, 0x77, 0x52, 0x6f, 0x75,
	0x6e, 0x64, 0x12, 0x17, 0x2e, 0x72, 0x65, 0x73, 0x6d, 0x69, 0x78, 0x2e, 0x4e, 0x65, 0x77, 0x52,
	0x6f, 0x75, 0x6e, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x18, 0x2e, 0x72, 0x65,
	0x73, 0x6d, 0x69, 0x78, 0x2e, 0x4e, 0x65, 0x77, 0x52, 0x6f, 0x75, 0x6e, 0x64, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x3f, 0x0a, 0x08, 0x45, 0x6e, 0x64, 0x52, 0x6f,
	0x75, 0x6e, 0x64, 0x12, 0x17, 0x2e, 0x72, 0x65, 0x73, 0x6d, 0x69, 0x78, 0x2e, 0x45, 0x6e, 0x64,
	0x52, 0x6f, 0x75, 0x6e, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x18, 0x2e, 0x72,
	0x65, 0x73, 0x6d, 0x69, 0x78, 0x2e, 0x45, 0x6e, 0x64, 0x52, 0x6f, 0x75, 0x6e, 0x64, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x22, 0x5a, 0x20, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6a, 0x6f, 0x6e, 0x61, 0x74, 0x68, 0x61, 0x6e, 0x4d,
	0x77, 0x65, 0x69, 0x73, 0x73, 0x2f, 0x72, 0x65, 0x73, 0x6d, 0x69, 0x78, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_mixnet_proto_rawDescOnce sync.Once
	file_mixnet_proto_rawDescData = file_mixnet_proto_rawDesc
)

func file_mixnet_proto_rawDescGZIP() []byte {
	file_mixnet_proto_rawDescOnce.Do(func() {
		file_mixnet_proto_rawDescData = protoimpl.X.CompressGZIP(file_mixnet_proto_rawDescData)
	})
	return file_mixnet_proto_rawDescData
}

var file_mixnet_proto_msgTypes = make([]protoimpl.MessageInfo, 8)
var file_mixnet_proto_goTypes = []interface{}{
	(*NewRoundRequest)(nil),    // 0: resmix.NewRoundRequest
	(*NewRoundResponse)(nil),   // 1: resmix.NewRoundResponse
	(*EndRoundRequest)(nil),    // 2: resmix.EndRoundRequest
	(*EndRoundResponse)(nil),   // 3: resmix.EndRoundResponse
	(*Messages)(nil),           // 4: resmix.Messages
	(*AddMessagesRequest)(nil), // 5: resmix.AddMessagesRequest
	(*Vote)(nil),               // 6: resmix.Vote
	(*KeyReleaseMessage)(nil),  // 7: resmix.KeyReleaseMessage
}
var file_mixnet_proto_depIdxs = []int32{
	4, // 0: resmix.AddMessagesRequest.messages:type_name -> resmix.Messages
	6, // 1: resmix.KeyReleaseMessage.votes:type_name -> resmix.Vote
	0, // 2: resmix.Mix.NewRound:input_type -> resmix.NewRoundRequest
	2, // 3: resmix.Mix.EndRound:input_type -> resmix.EndRoundRequest
	1, // 4: resmix.Mix.NewRound:output_type -> resmix.NewRoundResponse
	3, // 5: resmix.Mix.EndRound:output_type -> resmix.EndRoundResponse
	4, // [4:6] is the sub-list for method output_type
	2, // [2:4] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_mixnet_proto_init() }
func file_mixnet_proto_init() {
	if File_mixnet_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_mixnet_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NewRoundRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_mixnet_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NewRoundResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_mixnet_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EndRoundRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_mixnet_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EndRoundResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_mixnet_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Messages); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_mixnet_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AddMessagesRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_mixnet_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Vote); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_mixnet_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*KeyReleaseMessage); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_mixnet_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   8,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_mixnet_proto_goTypes,
		DependencyIndexes: file_mixnet_proto_depIdxs,
		MessageInfos:      file_mixnet_proto_msgTypes,
	}.Build()
	File_mixnet_proto = out.File
	file_mixnet_proto_rawDesc = nil
	file_mixnet_proto_goTypes = nil
	file_mixnet_proto_depIdxs = nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// MixClient is the client API for Mix service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type MixClient interface {
	// setting up:
	NewRound(ctx context.Context, in *NewRoundRequest, opts ...grpc.CallOption) (*NewRoundResponse, error)
	EndRound(ctx context.Context, in *EndRoundRequest, opts ...grpc.CallOption) (*EndRoundResponse, error)
}

type mixClient struct {
	cc grpc.ClientConnInterface
}

func NewMixClient(cc grpc.ClientConnInterface) MixClient {
	return &mixClient{cc}
}

func (c *mixClient) NewRound(ctx context.Context, in *NewRoundRequest, opts ...grpc.CallOption) (*NewRoundResponse, error) {
	out := new(NewRoundResponse)
	err := c.cc.Invoke(ctx, "/resmix.Mix/NewRound", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *mixClient) EndRound(ctx context.Context, in *EndRoundRequest, opts ...grpc.CallOption) (*EndRoundResponse, error) {
	out := new(EndRoundResponse)
	err := c.cc.Invoke(ctx, "/resmix.Mix/EndRound", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MixServer is the server API for Mix service.
type MixServer interface {
	// setting up:
	NewRound(context.Context, *NewRoundRequest) (*NewRoundResponse, error)
	EndRound(context.Context, *EndRoundRequest) (*EndRoundResponse, error)
}

// UnimplementedMixServer can be embedded to have forward compatible implementations.
type UnimplementedMixServer struct {
}

func (*UnimplementedMixServer) NewRound(context.Context, *NewRoundRequest) (*NewRoundResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NewRound not implemented")
}
func (*UnimplementedMixServer) EndRound(context.Context, *EndRoundRequest) (*EndRoundResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method EndRound not implemented")
}

func RegisterMixServer(s *grpc.Server, srv MixServer) {
	s.RegisterService(&_Mix_serviceDesc, srv)
}

func _Mix_NewRound_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NewRoundRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MixServer).NewRound(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/resmix.Mix/NewRound",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MixServer).NewRound(ctx, req.(*NewRoundRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Mix_EndRound_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EndRoundRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MixServer).EndRound(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/resmix.Mix/EndRound",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MixServer).EndRound(ctx, req.(*EndRoundRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Mix_serviceDesc = grpc.ServiceDesc{
	ServiceName: "resmix.Mix",
	HandlerType: (*MixServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "NewRound",
			Handler:    _Mix_NewRound_Handler,
		},
		{
			MethodName: "EndRound",
			Handler:    _Mix_EndRound_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "mixnet.proto",
}