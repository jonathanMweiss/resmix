// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.21.0-devel
// 	protoc        v3.7.1
// source: config.proto

package config

import (
	proto "github.com/golang/protobuf/proto"
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

type IBEConfigs struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	VSSPolynomial                   []byte            `protobuf:"bytes,1,opt,name=VSSPolynomial,proto3" json:"VSSPolynomial,omitempty"`
	VSSExponentPolynomials          map[string][]byte `protobuf:"bytes,2,rep,name=VSSExponentPolynomials,proto3" json:"VSSExponentPolynomials,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	AddressOfNodeToSecretShare      map[string][]byte `protobuf:"bytes,3,rep,name=AddressOfNodeToSecretShare,proto3" json:"AddressOfNodeToSecretShare,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	AddressOfNodeToMasterPublicKeys map[string][]byte `protobuf:"bytes,4,rep,name=AddressOfNodeToMasterPublicKeys,proto3" json:"AddressOfNodeToMasterPublicKeys,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *IBEConfigs) Reset() {
	*x = IBEConfigs{}
	if protoimpl.UnsafeEnabled {
		mi := &file_config_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *IBEConfigs) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*IBEConfigs) ProtoMessage() {}

func (x *IBEConfigs) ProtoReflect() protoreflect.Message {
	mi := &file_config_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use IBEConfigs.ProtoReflect.Descriptor instead.
func (*IBEConfigs) Descriptor() ([]byte, []int) {
	return file_config_proto_rawDescGZIP(), []int{0}
}

func (x *IBEConfigs) GetVSSPolynomial() []byte {
	if x != nil {
		return x.VSSPolynomial
	}
	return nil
}

func (x *IBEConfigs) GetVSSExponentPolynomials() map[string][]byte {
	if x != nil {
		return x.VSSExponentPolynomials
	}
	return nil
}

func (x *IBEConfigs) GetAddressOfNodeToSecretShare() map[string][]byte {
	if x != nil {
		return x.AddressOfNodeToSecretShare
	}
	return nil
}

func (x *IBEConfigs) GetAddressOfNodeToMasterPublicKeys() map[string][]byte {
	if x != nil {
		return x.AddressOfNodeToMasterPublicKeys
	}
	return nil
}

type SystemConfig struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ServerConfigs []*ServerConfig `protobuf:"bytes,1,rep,name=ServerConfigs,proto3" json:"ServerConfigs,omitempty"`
	LogicalMixes  *Topology       `protobuf:"bytes,2,opt,name=LogicalMixes,proto3" json:"LogicalMixes,omitempty"`
}

func (x *SystemConfig) Reset() {
	*x = SystemConfig{}
	if protoimpl.UnsafeEnabled {
		mi := &file_config_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SystemConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SystemConfig) ProtoMessage() {}

func (x *SystemConfig) ProtoReflect() protoreflect.Message {
	mi := &file_config_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SystemConfig.ProtoReflect.Descriptor instead.
func (*SystemConfig) Descriptor() ([]byte, []int) {
	return file_config_proto_rawDescGZIP(), []int{1}
}

func (x *SystemConfig) GetServerConfigs() []*ServerConfig {
	if x != nil {
		return x.ServerConfigs
	}
	return nil
}

func (x *SystemConfig) GetLogicalMixes() *Topology {
	if x != nil {
		return x.LogicalMixes
	}
	return nil
}

type Layer struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	LogicalMixes []*LogicalMix `protobuf:"bytes,1,rep,name=logicalMixes,proto3" json:"logicalMixes,omitempty"`
}

func (x *Layer) Reset() {
	*x = Layer{}
	if protoimpl.UnsafeEnabled {
		mi := &file_config_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Layer) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Layer) ProtoMessage() {}

func (x *Layer) ProtoReflect() protoreflect.Message {
	mi := &file_config_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Layer.ProtoReflect.Descriptor instead.
func (*Layer) Descriptor() ([]byte, []int) {
	return file_config_proto_rawDescGZIP(), []int{2}
}

func (x *Layer) GetLogicalMixes() []*LogicalMix {
	if x != nil {
		return x.LogicalMixes
	}
	return nil
}

type Topology struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Layers []*Layer               `protobuf:"bytes,1,rep,name=layers,proto3" json:"layers,omitempty"`
	Mixes  map[string]*LogicalMix `protobuf:"bytes,2,rep,name=mixes,proto3" json:"mixes,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *Topology) Reset() {
	*x = Topology{}
	if protoimpl.UnsafeEnabled {
		mi := &file_config_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Topology) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Topology) ProtoMessage() {}

func (x *Topology) ProtoReflect() protoreflect.Message {
	mi := &file_config_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Topology.ProtoReflect.Descriptor instead.
func (*Topology) Descriptor() ([]byte, []int) {
	return file_config_proto_rawDescGZIP(), []int{3}
}

func (x *Topology) GetLayers() []*Layer {
	if x != nil {
		return x.Layers
	}
	return nil
}

func (x *Topology) GetMixes() map[string]*LogicalMix {
	if x != nil {
		return x.Mixes
	}
	return nil
}

type ServerConfig struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Hostname string `protobuf:"bytes,1,opt,name=Hostname,proto3" json:"Hostname,omitempty"`
	// Id is the id of the server, represent it's position in the DKG/VSS.
	Id        uint32 `protobuf:"varint,2,opt,name=Id,proto3" json:"Id,omitempty"`
	Threshold uint32 `protobuf:"varint,3,opt,name=threshold,proto3" json:"threshold,omitempty"`
	// DKGPolynomialShare is the secret share this server holds from participating in DKG.
	SecretDKGShare []byte      `protobuf:"bytes,4,opt,name=SecretDKGShare,proto3" json:"SecretDKGShare,omitempty"`
	DKGPublicKeys  [][]byte    `protobuf:"bytes,5,rep,name=DKGPublicKeys,proto3" json:"DKGPublicKeys,omitempty"`
	IBEConfigs     *IBEConfigs `protobuf:"bytes,6,opt,name=IBEConfigs,proto3" json:"IBEConfigs,omitempty"`
	Mixes          []string    `protobuf:"bytes,7,rep,name=Mixes,proto3" json:"Mixes,omitempty"`
	RrpcSecretKey  []byte      `protobuf:"bytes,8,opt,name=rrpcSecretKey,proto3" json:"rrpcSecretKey,omitempty"`
	RrpcPublicKey  []byte      `protobuf:"bytes,9,opt,name=rrpcPublicKey,proto3" json:"rrpcPublicKey,omitempty"`
}

func (x *ServerConfig) Reset() {
	*x = ServerConfig{}
	if protoimpl.UnsafeEnabled {
		mi := &file_config_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ServerConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServerConfig) ProtoMessage() {}

func (x *ServerConfig) ProtoReflect() protoreflect.Message {
	mi := &file_config_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ServerConfig.ProtoReflect.Descriptor instead.
func (*ServerConfig) Descriptor() ([]byte, []int) {
	return file_config_proto_rawDescGZIP(), []int{4}
}

func (x *ServerConfig) GetHostname() string {
	if x != nil {
		return x.Hostname
	}
	return ""
}

func (x *ServerConfig) GetId() uint32 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *ServerConfig) GetThreshold() uint32 {
	if x != nil {
		return x.Threshold
	}
	return 0
}

func (x *ServerConfig) GetSecretDKGShare() []byte {
	if x != nil {
		return x.SecretDKGShare
	}
	return nil
}

func (x *ServerConfig) GetDKGPublicKeys() [][]byte {
	if x != nil {
		return x.DKGPublicKeys
	}
	return nil
}

func (x *ServerConfig) GetIBEConfigs() *IBEConfigs {
	if x != nil {
		return x.IBEConfigs
	}
	return nil
}

func (x *ServerConfig) GetMixes() []string {
	if x != nil {
		return x.Mixes
	}
	return nil
}

func (x *ServerConfig) GetRrpcSecretKey() []byte {
	if x != nil {
		return x.RrpcSecretKey
	}
	return nil
}

func (x *ServerConfig) GetRrpcPublicKey() []byte {
	if x != nil {
		return x.RrpcPublicKey
	}
	return nil
}

type LogicalMix struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// who owns this mix
	Hostname     string   `protobuf:"bytes,1,opt,name=Hostname,proto3" json:"Hostname,omitempty"`
	Name         string   `protobuf:"bytes,2,opt,name=Name,proto3" json:"Name,omitempty"`
	ServerIndex  int32    `protobuf:"varint,3,opt,name=ServerIndex,proto3" json:"ServerIndex,omitempty"`
	Layer        int32    `protobuf:"varint,4,opt,name=layer,proto3" json:"layer,omitempty"`
	Predecessors []string `protobuf:"bytes,5,rep,name=predecessors,proto3" json:"predecessors,omitempty"`
	Successors   []string `protobuf:"bytes,6,rep,name=successors,proto3" json:"successors,omitempty"`
}

func (x *LogicalMix) Reset() {
	*x = LogicalMix{}
	if protoimpl.UnsafeEnabled {
		mi := &file_config_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LogicalMix) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LogicalMix) ProtoMessage() {}

func (x *LogicalMix) ProtoReflect() protoreflect.Message {
	mi := &file_config_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LogicalMix.ProtoReflect.Descriptor instead.
func (*LogicalMix) Descriptor() ([]byte, []int) {
	return file_config_proto_rawDescGZIP(), []int{5}
}

func (x *LogicalMix) GetHostname() string {
	if x != nil {
		return x.Hostname
	}
	return ""
}

func (x *LogicalMix) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *LogicalMix) GetServerIndex() int32 {
	if x != nil {
		return x.ServerIndex
	}
	return 0
}

func (x *LogicalMix) GetLayer() int32 {
	if x != nil {
		return x.Layer
	}
	return 0
}

func (x *LogicalMix) GetPredecessors() []string {
	if x != nil {
		return x.Predecessors
	}
	return nil
}

func (x *LogicalMix) GetSuccessors() []string {
	if x != nil {
		return x.Successors
	}
	return nil
}

var File_config_proto protoreflect.FileDescriptor

var file_config_proto_rawDesc = []byte{
	0x0a, 0x0c, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06,
	0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x22, 0x80, 0x05, 0x0a, 0x0a, 0x49, 0x42, 0x45, 0x43, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x73, 0x12, 0x24, 0x0a, 0x0d, 0x56, 0x53, 0x53, 0x50, 0x6f, 0x6c, 0x79,
	0x6e, 0x6f, 0x6d, 0x69, 0x61, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0d, 0x56, 0x53,
	0x53, 0x50, 0x6f, 0x6c, 0x79, 0x6e, 0x6f, 0x6d, 0x69, 0x61, 0x6c, 0x12, 0x66, 0x0a, 0x16, 0x56,
	0x53, 0x53, 0x45, 0x78, 0x70, 0x6f, 0x6e, 0x65, 0x6e, 0x74, 0x50, 0x6f, 0x6c, 0x79, 0x6e, 0x6f,
	0x6d, 0x69, 0x61, 0x6c, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x2e, 0x2e, 0x63, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x2e, 0x49, 0x42, 0x45, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x73, 0x2e,
	0x56, 0x53, 0x53, 0x45, 0x78, 0x70, 0x6f, 0x6e, 0x65, 0x6e, 0x74, 0x50, 0x6f, 0x6c, 0x79, 0x6e,
	0x6f, 0x6d, 0x69, 0x61, 0x6c, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x16, 0x56, 0x53, 0x53,
	0x45, 0x78, 0x70, 0x6f, 0x6e, 0x65, 0x6e, 0x74, 0x50, 0x6f, 0x6c, 0x79, 0x6e, 0x6f, 0x6d, 0x69,
	0x61, 0x6c, 0x73, 0x12, 0x72, 0x0a, 0x1a, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x4f, 0x66,
	0x4e, 0x6f, 0x64, 0x65, 0x54, 0x6f, 0x53, 0x65, 0x63, 0x72, 0x65, 0x74, 0x53, 0x68, 0x61, 0x72,
	0x65, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x32, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67,
	0x2e, 0x49, 0x42, 0x45, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x73, 0x2e, 0x41, 0x64, 0x64, 0x72,
	0x65, 0x73, 0x73, 0x4f, 0x66, 0x4e, 0x6f, 0x64, 0x65, 0x54, 0x6f, 0x53, 0x65, 0x63, 0x72, 0x65,
	0x74, 0x53, 0x68, 0x61, 0x72, 0x65, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x1a, 0x41, 0x64, 0x64,
	0x72, 0x65, 0x73, 0x73, 0x4f, 0x66, 0x4e, 0x6f, 0x64, 0x65, 0x54, 0x6f, 0x53, 0x65, 0x63, 0x72,
	0x65, 0x74, 0x53, 0x68, 0x61, 0x72, 0x65, 0x12, 0x81, 0x01, 0x0a, 0x1f, 0x41, 0x64, 0x64, 0x72,
	0x65, 0x73, 0x73, 0x4f, 0x66, 0x4e, 0x6f, 0x64, 0x65, 0x54, 0x6f, 0x4d, 0x61, 0x73, 0x74, 0x65,
	0x72, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x4b, 0x65, 0x79, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x37, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x49, 0x42, 0x45, 0x43, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x73, 0x2e, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x4f, 0x66, 0x4e,
	0x6f, 0x64, 0x65, 0x54, 0x6f, 0x4d, 0x61, 0x73, 0x74, 0x65, 0x72, 0x50, 0x75, 0x62, 0x6c, 0x69,
	0x63, 0x4b, 0x65, 0x79, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x1f, 0x41, 0x64, 0x64, 0x72,
	0x65, 0x73, 0x73, 0x4f, 0x66, 0x4e, 0x6f, 0x64, 0x65, 0x54, 0x6f, 0x4d, 0x61, 0x73, 0x74, 0x65,
	0x72, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x4b, 0x65, 0x79, 0x73, 0x1a, 0x49, 0x0a, 0x1b, 0x56,
	0x53, 0x53, 0x45, 0x78, 0x70, 0x6f, 0x6e, 0x65, 0x6e, 0x74, 0x50, 0x6f, 0x6c, 0x79, 0x6e, 0x6f,
	0x6d, 0x69, 0x61, 0x6c, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65,
	0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05,
	0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x05, 0x76, 0x61, 0x6c,
	0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x1a, 0x4d, 0x0a, 0x1f, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73,
	0x73, 0x4f, 0x66, 0x4e, 0x6f, 0x64, 0x65, 0x54, 0x6f, 0x53, 0x65, 0x63, 0x72, 0x65, 0x74, 0x53,
	0x68, 0x61, 0x72, 0x65, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76,
	0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x3a, 0x02, 0x38, 0x01, 0x1a, 0x52, 0x0a, 0x24, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73,
	0x4f, 0x66, 0x4e, 0x6f, 0x64, 0x65, 0x54, 0x6f, 0x4d, 0x61, 0x73, 0x74, 0x65, 0x72, 0x50, 0x75,
	0x62, 0x6c, 0x69, 0x63, 0x4b, 0x65, 0x79, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a,
	0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12,
	0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x05,
	0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0x80, 0x01, 0x0a, 0x0c, 0x53, 0x79,
	0x73, 0x74, 0x65, 0x6d, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x3a, 0x0a, 0x0d, 0x53, 0x65,
	0x72, 0x76, 0x65, 0x72, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x14, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x53, 0x65, 0x72, 0x76, 0x65,
	0x72, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x0d, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x43,
	0x6f, 0x6e, 0x66, 0x69, 0x67, 0x73, 0x12, 0x34, 0x0a, 0x0c, 0x4c, 0x6f, 0x67, 0x69, 0x63, 0x61,
	0x6c, 0x4d, 0x69, 0x78, 0x65, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x63,
	0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x54, 0x6f, 0x70, 0x6f, 0x6c, 0x6f, 0x67, 0x79, 0x52, 0x0c,
	0x4c, 0x6f, 0x67, 0x69, 0x63, 0x61, 0x6c, 0x4d, 0x69, 0x78, 0x65, 0x73, 0x22, 0x3f, 0x0a, 0x05,
	0x4c, 0x61, 0x79, 0x65, 0x72, 0x12, 0x36, 0x0a, 0x0c, 0x6c, 0x6f, 0x67, 0x69, 0x63, 0x61, 0x6c,
	0x4d, 0x69, 0x78, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x63, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x2e, 0x4c, 0x6f, 0x67, 0x69, 0x63, 0x61, 0x6c, 0x4d, 0x69, 0x78, 0x52,
	0x0c, 0x6c, 0x6f, 0x67, 0x69, 0x63, 0x61, 0x6c, 0x4d, 0x69, 0x78, 0x65, 0x73, 0x22, 0xb2, 0x01,
	0x0a, 0x08, 0x54, 0x6f, 0x70, 0x6f, 0x6c, 0x6f, 0x67, 0x79, 0x12, 0x25, 0x0a, 0x06, 0x6c, 0x61,
	0x79, 0x65, 0x72, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x63, 0x6f, 0x6e,
	0x66, 0x69, 0x67, 0x2e, 0x4c, 0x61, 0x79, 0x65, 0x72, 0x52, 0x06, 0x6c, 0x61, 0x79, 0x65, 0x72,
	0x73, 0x12, 0x31, 0x0a, 0x05, 0x6d, 0x69, 0x78, 0x65, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x1b, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x54, 0x6f, 0x70, 0x6f, 0x6c, 0x6f,
	0x67, 0x79, 0x2e, 0x4d, 0x69, 0x78, 0x65, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x05, 0x6d,
	0x69, 0x78, 0x65, 0x73, 0x1a, 0x4c, 0x0a, 0x0a, 0x4d, 0x69, 0x78, 0x65, 0x73, 0x45, 0x6e, 0x74,
	0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x03, 0x6b, 0x65, 0x79, 0x12, 0x28, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x4c, 0x6f, 0x67,
	0x69, 0x63, 0x61, 0x6c, 0x4d, 0x69, 0x78, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02,
	0x38, 0x01, 0x22, 0xbc, 0x02, 0x0a, 0x0c, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x43, 0x6f, 0x6e,
	0x66, 0x69, 0x67, 0x12, 0x1a, 0x0a, 0x08, 0x48, 0x6f, 0x73, 0x74, 0x6e, 0x61, 0x6d, 0x65, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x48, 0x6f, 0x73, 0x74, 0x6e, 0x61, 0x6d, 0x65, 0x12,
	0x0e, 0x0a, 0x02, 0x49, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x02, 0x49, 0x64, 0x12,
	0x1c, 0x0a, 0x09, 0x74, 0x68, 0x72, 0x65, 0x73, 0x68, 0x6f, 0x6c, 0x64, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x0d, 0x52, 0x09, 0x74, 0x68, 0x72, 0x65, 0x73, 0x68, 0x6f, 0x6c, 0x64, 0x12, 0x26, 0x0a,
	0x0e, 0x53, 0x65, 0x63, 0x72, 0x65, 0x74, 0x44, 0x4b, 0x47, 0x53, 0x68, 0x61, 0x72, 0x65, 0x18,
	0x04, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0e, 0x53, 0x65, 0x63, 0x72, 0x65, 0x74, 0x44, 0x4b, 0x47,
	0x53, 0x68, 0x61, 0x72, 0x65, 0x12, 0x24, 0x0a, 0x0d, 0x44, 0x4b, 0x47, 0x50, 0x75, 0x62, 0x6c,
	0x69, 0x63, 0x4b, 0x65, 0x79, 0x73, 0x18, 0x05, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x0d, 0x44, 0x4b,
	0x47, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x4b, 0x65, 0x79, 0x73, 0x12, 0x32, 0x0a, 0x0a, 0x49,
	0x42, 0x45, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x73, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x12, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x49, 0x42, 0x45, 0x43, 0x6f, 0x6e, 0x66,
	0x69, 0x67, 0x73, 0x52, 0x0a, 0x49, 0x42, 0x45, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x73, 0x12,
	0x14, 0x0a, 0x05, 0x4d, 0x69, 0x78, 0x65, 0x73, 0x18, 0x07, 0x20, 0x03, 0x28, 0x09, 0x52, 0x05,
	0x4d, 0x69, 0x78, 0x65, 0x73, 0x12, 0x24, 0x0a, 0x0d, 0x72, 0x72, 0x70, 0x63, 0x53, 0x65, 0x63,
	0x72, 0x65, 0x74, 0x4b, 0x65, 0x79, 0x18, 0x08, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0d, 0x72, 0x72,
	0x70, 0x63, 0x53, 0x65, 0x63, 0x72, 0x65, 0x74, 0x4b, 0x65, 0x79, 0x12, 0x24, 0x0a, 0x0d, 0x72,
	0x72, 0x70, 0x63, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x4b, 0x65, 0x79, 0x18, 0x09, 0x20, 0x01,
	0x28, 0x0c, 0x52, 0x0d, 0x72, 0x72, 0x70, 0x63, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x4b, 0x65,
	0x79, 0x22, 0xb8, 0x01, 0x0a, 0x0a, 0x4c, 0x6f, 0x67, 0x69, 0x63, 0x61, 0x6c, 0x4d, 0x69, 0x78,
	0x12, 0x1a, 0x0a, 0x08, 0x48, 0x6f, 0x73, 0x74, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x08, 0x48, 0x6f, 0x73, 0x74, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x12, 0x0a, 0x04,
	0x4e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x4e, 0x61, 0x6d, 0x65,
	0x12, 0x20, 0x0a, 0x0b, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x49, 0x6e, 0x64, 0x65, 0x78, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0b, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x49, 0x6e, 0x64,
	0x65, 0x78, 0x12, 0x14, 0x0a, 0x05, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x18, 0x04, 0x20, 0x01, 0x28,
	0x05, 0x52, 0x05, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x12, 0x22, 0x0a, 0x0c, 0x70, 0x72, 0x65, 0x64,
	0x65, 0x63, 0x65, 0x73, 0x73, 0x6f, 0x72, 0x73, 0x18, 0x05, 0x20, 0x03, 0x28, 0x09, 0x52, 0x0c,
	0x70, 0x72, 0x65, 0x64, 0x65, 0x63, 0x65, 0x73, 0x73, 0x6f, 0x72, 0x73, 0x12, 0x1e, 0x0a, 0x0a,
	0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x6f, 0x72, 0x73, 0x18, 0x06, 0x20, 0x03, 0x28, 0x09,
	0x52, 0x0a, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x6f, 0x72, 0x73, 0x42, 0x29, 0x5a, 0x27,
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6a, 0x6f, 0x6e, 0x61, 0x74,
	0x68, 0x61, 0x6e, 0x4d, 0x77, 0x65, 0x69, 0x73, 0x73, 0x2f, 0x72, 0x65, 0x73, 0x6d, 0x69, 0x78,
	0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_config_proto_rawDescOnce sync.Once
	file_config_proto_rawDescData = file_config_proto_rawDesc
)

func file_config_proto_rawDescGZIP() []byte {
	file_config_proto_rawDescOnce.Do(func() {
		file_config_proto_rawDescData = protoimpl.X.CompressGZIP(file_config_proto_rawDescData)
	})
	return file_config_proto_rawDescData
}

var file_config_proto_msgTypes = make([]protoimpl.MessageInfo, 10)
var file_config_proto_goTypes = []interface{}{
	(*IBEConfigs)(nil),   // 0: config.IBEConfigs
	(*SystemConfig)(nil), // 1: config.SystemConfig
	(*Layer)(nil),        // 2: config.Layer
	(*Topology)(nil),     // 3: config.Topology
	(*ServerConfig)(nil), // 4: config.ServerConfig
	(*LogicalMix)(nil),   // 5: config.LogicalMix
	nil,                  // 6: config.IBEConfigs.VSSExponentPolynomialsEntry
	nil,                  // 7: config.IBEConfigs.AddressOfNodeToSecretShareEntry
	nil,                  // 8: config.IBEConfigs.AddressOfNodeToMasterPublicKeysEntry
	nil,                  // 9: config.Topology.MixesEntry
}
var file_config_proto_depIdxs = []int32{
	6,  // 0: config.IBEConfigs.VSSExponentPolynomials:type_name -> config.IBEConfigs.VSSExponentPolynomialsEntry
	7,  // 1: config.IBEConfigs.AddressOfNodeToSecretShare:type_name -> config.IBEConfigs.AddressOfNodeToSecretShareEntry
	8,  // 2: config.IBEConfigs.AddressOfNodeToMasterPublicKeys:type_name -> config.IBEConfigs.AddressOfNodeToMasterPublicKeysEntry
	4,  // 3: config.SystemConfig.ServerConfigs:type_name -> config.ServerConfig
	3,  // 4: config.SystemConfig.LogicalMixes:type_name -> config.Topology
	5,  // 5: config.Layer.logicalMixes:type_name -> config.LogicalMix
	2,  // 6: config.Topology.layers:type_name -> config.Layer
	9,  // 7: config.Topology.mixes:type_name -> config.Topology.MixesEntry
	0,  // 8: config.ServerConfig.IBEConfigs:type_name -> config.IBEConfigs
	5,  // 9: config.Topology.MixesEntry.value:type_name -> config.LogicalMix
	10, // [10:10] is the sub-list for method output_type
	10, // [10:10] is the sub-list for method input_type
	10, // [10:10] is the sub-list for extension type_name
	10, // [10:10] is the sub-list for extension extendee
	0,  // [0:10] is the sub-list for field type_name
}

func init() { file_config_proto_init() }
func file_config_proto_init() {
	if File_config_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_config_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*IBEConfigs); i {
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
		file_config_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SystemConfig); i {
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
		file_config_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Layer); i {
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
		file_config_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Topology); i {
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
		file_config_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ServerConfig); i {
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
		file_config_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LogicalMix); i {
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
			RawDescriptor: file_config_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   10,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_config_proto_goTypes,
		DependencyIndexes: file_config_proto_depIdxs,
		MessageInfos:      file_config_proto_msgTypes,
	}.Build()
	File_config_proto = out.File
	file_config_proto_rawDesc = nil
	file_config_proto_goTypes = nil
	file_config_proto_depIdxs = nil
}
