// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.1
// 	protoc        (unknown)
// source: xyz/block/ftl/v1/module.proto

package ftlv1

import (
	schema "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	durationpb "google.golang.org/protobuf/types/known/durationpb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type ModuleContextResponse_DBType int32

const (
	ModuleContextResponse_POSTGRES ModuleContextResponse_DBType = 0
)

// Enum value maps for ModuleContextResponse_DBType.
var (
	ModuleContextResponse_DBType_name = map[int32]string{
		0: "POSTGRES",
	}
	ModuleContextResponse_DBType_value = map[string]int32{
		"POSTGRES": 0,
	}
)

func (x ModuleContextResponse_DBType) Enum() *ModuleContextResponse_DBType {
	p := new(ModuleContextResponse_DBType)
	*p = x
	return p
}

func (x ModuleContextResponse_DBType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ModuleContextResponse_DBType) Descriptor() protoreflect.EnumDescriptor {
	return file_xyz_block_ftl_v1_module_proto_enumTypes[0].Descriptor()
}

func (ModuleContextResponse_DBType) Type() protoreflect.EnumType {
	return &file_xyz_block_ftl_v1_module_proto_enumTypes[0]
}

func (x ModuleContextResponse_DBType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ModuleContextResponse_DBType.Descriptor instead.
func (ModuleContextResponse_DBType) EnumDescriptor() ([]byte, []int) {
	return file_xyz_block_ftl_v1_module_proto_rawDescGZIP(), []int{5, 0}
}

type AcquireLeaseRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Module string               `protobuf:"bytes,1,opt,name=module,proto3" json:"module,omitempty"`
	Key    []string             `protobuf:"bytes,2,rep,name=key,proto3" json:"key,omitempty"`
	Ttl    *durationpb.Duration `protobuf:"bytes,3,opt,name=ttl,proto3" json:"ttl,omitempty"`
}

func (x *AcquireLeaseRequest) Reset() {
	*x = AcquireLeaseRequest{}
	mi := &file_xyz_block_ftl_v1_module_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AcquireLeaseRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AcquireLeaseRequest) ProtoMessage() {}

func (x *AcquireLeaseRequest) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_v1_module_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AcquireLeaseRequest.ProtoReflect.Descriptor instead.
func (*AcquireLeaseRequest) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_v1_module_proto_rawDescGZIP(), []int{0}
}

func (x *AcquireLeaseRequest) GetModule() string {
	if x != nil {
		return x.Module
	}
	return ""
}

func (x *AcquireLeaseRequest) GetKey() []string {
	if x != nil {
		return x.Key
	}
	return nil
}

func (x *AcquireLeaseRequest) GetTtl() *durationpb.Duration {
	if x != nil {
		return x.Ttl
	}
	return nil
}

type AcquireLeaseResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *AcquireLeaseResponse) Reset() {
	*x = AcquireLeaseResponse{}
	mi := &file_xyz_block_ftl_v1_module_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AcquireLeaseResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AcquireLeaseResponse) ProtoMessage() {}

func (x *AcquireLeaseResponse) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_v1_module_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AcquireLeaseResponse.ProtoReflect.Descriptor instead.
func (*AcquireLeaseResponse) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_v1_module_proto_rawDescGZIP(), []int{1}
}

type PublishEventRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Topic *schema.Ref `protobuf:"bytes,1,opt,name=topic,proto3" json:"topic,omitempty"`
	Body  []byte      `protobuf:"bytes,2,opt,name=body,proto3" json:"body,omitempty"`
	// Only verb name is included because this verb will be in the same module as topic
	Caller string `protobuf:"bytes,3,opt,name=caller,proto3" json:"caller,omitempty"`
}

func (x *PublishEventRequest) Reset() {
	*x = PublishEventRequest{}
	mi := &file_xyz_block_ftl_v1_module_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PublishEventRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PublishEventRequest) ProtoMessage() {}

func (x *PublishEventRequest) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_v1_module_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PublishEventRequest.ProtoReflect.Descriptor instead.
func (*PublishEventRequest) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_v1_module_proto_rawDescGZIP(), []int{2}
}

func (x *PublishEventRequest) GetTopic() *schema.Ref {
	if x != nil {
		return x.Topic
	}
	return nil
}

func (x *PublishEventRequest) GetBody() []byte {
	if x != nil {
		return x.Body
	}
	return nil
}

func (x *PublishEventRequest) GetCaller() string {
	if x != nil {
		return x.Caller
	}
	return ""
}

type PublishEventResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *PublishEventResponse) Reset() {
	*x = PublishEventResponse{}
	mi := &file_xyz_block_ftl_v1_module_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PublishEventResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PublishEventResponse) ProtoMessage() {}

func (x *PublishEventResponse) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_v1_module_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PublishEventResponse.ProtoReflect.Descriptor instead.
func (*PublishEventResponse) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_v1_module_proto_rawDescGZIP(), []int{3}
}

type ModuleContextRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Module string `protobuf:"bytes,1,opt,name=module,proto3" json:"module,omitempty"`
}

func (x *ModuleContextRequest) Reset() {
	*x = ModuleContextRequest{}
	mi := &file_xyz_block_ftl_v1_module_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ModuleContextRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ModuleContextRequest) ProtoMessage() {}

func (x *ModuleContextRequest) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_v1_module_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ModuleContextRequest.ProtoReflect.Descriptor instead.
func (*ModuleContextRequest) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_v1_module_proto_rawDescGZIP(), []int{4}
}

func (x *ModuleContextRequest) GetModule() string {
	if x != nil {
		return x.Module
	}
	return ""
}

type ModuleContextResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Module    string                       `protobuf:"bytes,1,opt,name=module,proto3" json:"module,omitempty"`
	Configs   map[string][]byte            `protobuf:"bytes,2,rep,name=configs,proto3" json:"configs,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Secrets   map[string][]byte            `protobuf:"bytes,3,rep,name=secrets,proto3" json:"secrets,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Databases []*ModuleContextResponse_DSN `protobuf:"bytes,4,rep,name=databases,proto3" json:"databases,omitempty"`
}

func (x *ModuleContextResponse) Reset() {
	*x = ModuleContextResponse{}
	mi := &file_xyz_block_ftl_v1_module_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ModuleContextResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ModuleContextResponse) ProtoMessage() {}

func (x *ModuleContextResponse) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_v1_module_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ModuleContextResponse.ProtoReflect.Descriptor instead.
func (*ModuleContextResponse) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_v1_module_proto_rawDescGZIP(), []int{5}
}

func (x *ModuleContextResponse) GetModule() string {
	if x != nil {
		return x.Module
	}
	return ""
}

func (x *ModuleContextResponse) GetConfigs() map[string][]byte {
	if x != nil {
		return x.Configs
	}
	return nil
}

func (x *ModuleContextResponse) GetSecrets() map[string][]byte {
	if x != nil {
		return x.Secrets
	}
	return nil
}

func (x *ModuleContextResponse) GetDatabases() []*ModuleContextResponse_DSN {
	if x != nil {
		return x.Databases
	}
	return nil
}

type ModuleContextResponse_Ref struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Module *string `protobuf:"bytes,1,opt,name=module,proto3,oneof" json:"module,omitempty"`
	Name   string  `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *ModuleContextResponse_Ref) Reset() {
	*x = ModuleContextResponse_Ref{}
	mi := &file_xyz_block_ftl_v1_module_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ModuleContextResponse_Ref) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ModuleContextResponse_Ref) ProtoMessage() {}

func (x *ModuleContextResponse_Ref) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_v1_module_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ModuleContextResponse_Ref.ProtoReflect.Descriptor instead.
func (*ModuleContextResponse_Ref) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_v1_module_proto_rawDescGZIP(), []int{5, 0}
}

func (x *ModuleContextResponse_Ref) GetModule() string {
	if x != nil && x.Module != nil {
		return *x.Module
	}
	return ""
}

func (x *ModuleContextResponse_Ref) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

type ModuleContextResponse_DSN struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name string                       `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Type ModuleContextResponse_DBType `protobuf:"varint,2,opt,name=type,proto3,enum=xyz.block.ftl.v1.ModuleContextResponse_DBType" json:"type,omitempty"`
	Dsn  string                       `protobuf:"bytes,3,opt,name=dsn,proto3" json:"dsn,omitempty"`
}

func (x *ModuleContextResponse_DSN) Reset() {
	*x = ModuleContextResponse_DSN{}
	mi := &file_xyz_block_ftl_v1_module_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ModuleContextResponse_DSN) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ModuleContextResponse_DSN) ProtoMessage() {}

func (x *ModuleContextResponse_DSN) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_v1_module_proto_msgTypes[7]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ModuleContextResponse_DSN.ProtoReflect.Descriptor instead.
func (*ModuleContextResponse_DSN) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_v1_module_proto_rawDescGZIP(), []int{5, 1}
}

func (x *ModuleContextResponse_DSN) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *ModuleContextResponse_DSN) GetType() ModuleContextResponse_DBType {
	if x != nil {
		return x.Type
	}
	return ModuleContextResponse_POSTGRES
}

func (x *ModuleContextResponse_DSN) GetDsn() string {
	if x != nil {
		return x.Dsn
	}
	return ""
}

var File_xyz_block_ftl_v1_module_proto protoreflect.FileDescriptor

var file_xyz_block_ftl_v1_module_proto_rawDesc = []byte{
	0x0a, 0x1d, 0x78, 0x79, 0x7a, 0x2f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2f, 0x66, 0x74, 0x6c, 0x2f,
	0x76, 0x31, 0x2f, 0x6d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x10, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76,
	0x31, 0x1a, 0x1e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x75, 0x66, 0x2f, 0x64, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x1a, 0x1a, 0x78, 0x79, 0x7a, 0x2f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2f, 0x66, 0x74, 0x6c,
	0x2f, 0x76, 0x31, 0x2f, 0x66, 0x74, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x24, 0x78,
	0x79, 0x7a, 0x2f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2f, 0x66, 0x74, 0x6c, 0x2f, 0x76, 0x31, 0x2f,
	0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x2f, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x22, 0x6c, 0x0a, 0x13, 0x41, 0x63, 0x71, 0x75, 0x69, 0x72, 0x65, 0x4c, 0x65,
	0x61, 0x73, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x6d, 0x6f,
	0x64, 0x75, 0x6c, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x6d, 0x6f, 0x64, 0x75,
	0x6c, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x02, 0x20, 0x03, 0x28, 0x09, 0x52,
	0x03, 0x6b, 0x65, 0x79, 0x12, 0x2b, 0x0a, 0x03, 0x74, 0x74, 0x6c, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x19, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2e, 0x44, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x03, 0x74, 0x74,
	0x6c, 0x22, 0x16, 0x0a, 0x14, 0x41, 0x63, 0x71, 0x75, 0x69, 0x72, 0x65, 0x4c, 0x65, 0x61, 0x73,
	0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x75, 0x0a, 0x13, 0x50, 0x75, 0x62,
	0x6c, 0x69, 0x73, 0x68, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x12, 0x32, 0x0a, 0x05, 0x74, 0x6f, 0x70, 0x69, 0x63, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x1c, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e,
	0x76, 0x31, 0x2e, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x2e, 0x52, 0x65, 0x66, 0x52, 0x05, 0x74,
	0x6f, 0x70, 0x69, 0x63, 0x12, 0x12, 0x0a, 0x04, 0x62, 0x6f, 0x64, 0x79, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0c, 0x52, 0x04, 0x62, 0x6f, 0x64, 0x79, 0x12, 0x16, 0x0a, 0x06, 0x63, 0x61, 0x6c, 0x6c,
	0x65, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x63, 0x61, 0x6c, 0x6c, 0x65, 0x72,
	0x22, 0x16, 0x0a, 0x14, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x73, 0x68, 0x45, 0x76, 0x65, 0x6e, 0x74,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x2e, 0x0a, 0x14, 0x4d, 0x6f, 0x64, 0x75,
	0x6c, 0x65, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x12, 0x16, 0x0a, 0x06, 0x6d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x06, 0x6d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x22, 0xde, 0x04, 0x0a, 0x15, 0x4d, 0x6f, 0x64,
	0x75, 0x6c, 0x65, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x6d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x06, 0x6d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x12, 0x4e, 0x0a, 0x07, 0x63, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x34, 0x2e, 0x78, 0x79,
	0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x4d,
	0x6f, 0x64, 0x75, 0x6c, 0x65, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x2e, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x73, 0x45, 0x6e, 0x74, 0x72,
	0x79, 0x52, 0x07, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x73, 0x12, 0x4e, 0x0a, 0x07, 0x73, 0x65,
	0x63, 0x72, 0x65, 0x74, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x34, 0x2e, 0x78, 0x79,
	0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x4d,
	0x6f, 0x64, 0x75, 0x6c, 0x65, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x2e, 0x53, 0x65, 0x63, 0x72, 0x65, 0x74, 0x73, 0x45, 0x6e, 0x74, 0x72,
	0x79, 0x52, 0x07, 0x73, 0x65, 0x63, 0x72, 0x65, 0x74, 0x73, 0x12, 0x49, 0x0a, 0x09, 0x64, 0x61,
	0x74, 0x61, 0x62, 0x61, 0x73, 0x65, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x2b, 0x2e,
	0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76, 0x31,
	0x2e, 0x4d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x2e, 0x44, 0x53, 0x4e, 0x52, 0x09, 0x64, 0x61, 0x74, 0x61,
	0x62, 0x61, 0x73, 0x65, 0x73, 0x1a, 0x41, 0x0a, 0x03, 0x52, 0x65, 0x66, 0x12, 0x1b, 0x0a, 0x06,
	0x6d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x06,
	0x6d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x88, 0x01, 0x01, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d,
	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x42, 0x09, 0x0a,
	0x07, 0x5f, 0x6d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x1a, 0x6f, 0x0a, 0x03, 0x44, 0x53, 0x4e, 0x12,
	0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e,
	0x61, 0x6d, 0x65, 0x12, 0x42, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0e, 0x32, 0x2e, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74,
	0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x4d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x43, 0x6f, 0x6e, 0x74, 0x65,
	0x78, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x2e, 0x44, 0x42, 0x54, 0x79, 0x70,
	0x65, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x64, 0x73, 0x6e, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x64, 0x73, 0x6e, 0x1a, 0x3a, 0x0a, 0x0c, 0x43, 0x6f, 0x6e,
	0x66, 0x69, 0x67, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76,
	0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x3a, 0x02, 0x38, 0x01, 0x1a, 0x3a, 0x0a, 0x0c, 0x53, 0x65, 0x63, 0x72, 0x65, 0x74, 0x73,
	0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38,
	0x01, 0x22, 0x16, 0x0a, 0x06, 0x44, 0x42, 0x54, 0x79, 0x70, 0x65, 0x12, 0x0c, 0x0a, 0x08, 0x50,
	0x4f, 0x53, 0x54, 0x47, 0x52, 0x45, 0x53, 0x10, 0x00, 0x32, 0x84, 0x03, 0x0a, 0x0d, 0x4d, 0x6f,
	0x64, 0x75, 0x6c, 0x65, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x4a, 0x0a, 0x04, 0x50,
	0x69, 0x6e, 0x67, 0x12, 0x1d, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e,
	0x66, 0x74, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x1e, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66,
	0x74, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x22, 0x03, 0x90, 0x02, 0x01, 0x12, 0x65, 0x0a, 0x10, 0x47, 0x65, 0x74, 0x4d, 0x6f,
	0x64, 0x75, 0x6c, 0x65, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x12, 0x26, 0x2e, 0x78, 0x79,
	0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x4d,
	0x6f, 0x64, 0x75, 0x6c, 0x65, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x27, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e,
	0x66, 0x74, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x4d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x43, 0x6f, 0x6e,
	0x74, 0x65, 0x78, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x30, 0x01, 0x12, 0x61,
	0x0a, 0x0c, 0x41, 0x63, 0x71, 0x75, 0x69, 0x72, 0x65, 0x4c, 0x65, 0x61, 0x73, 0x65, 0x12, 0x25,
	0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76,
	0x31, 0x2e, 0x41, 0x63, 0x71, 0x75, 0x69, 0x72, 0x65, 0x4c, 0x65, 0x61, 0x73, 0x65, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x26, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63,
	0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x41, 0x63, 0x71, 0x75, 0x69, 0x72, 0x65,
	0x4c, 0x65, 0x61, 0x73, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x28, 0x01, 0x30,
	0x01, 0x12, 0x5d, 0x0a, 0x0c, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x73, 0x68, 0x45, 0x76, 0x65, 0x6e,
	0x74, 0x12, 0x25, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74,
	0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x73, 0x68, 0x45, 0x76, 0x65, 0x6e,
	0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x26, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62,
	0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x75, 0x62, 0x6c,
	0x69, 0x73, 0x68, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x42, 0x44, 0x50, 0x01, 0x5a, 0x40, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d,
	0x2f, 0x54, 0x42, 0x44, 0x35, 0x34, 0x35, 0x36, 0x36, 0x39, 0x37, 0x35, 0x2f, 0x66, 0x74, 0x6c,
	0x2f, 0x62, 0x61, 0x63, 0x6b, 0x65, 0x6e, 0x64, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2f,
	0x78, 0x79, 0x7a, 0x2f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2f, 0x66, 0x74, 0x6c, 0x2f, 0x76, 0x31,
	0x3b, 0x66, 0x74, 0x6c, 0x76, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_xyz_block_ftl_v1_module_proto_rawDescOnce sync.Once
	file_xyz_block_ftl_v1_module_proto_rawDescData = file_xyz_block_ftl_v1_module_proto_rawDesc
)

func file_xyz_block_ftl_v1_module_proto_rawDescGZIP() []byte {
	file_xyz_block_ftl_v1_module_proto_rawDescOnce.Do(func() {
		file_xyz_block_ftl_v1_module_proto_rawDescData = protoimpl.X.CompressGZIP(file_xyz_block_ftl_v1_module_proto_rawDescData)
	})
	return file_xyz_block_ftl_v1_module_proto_rawDescData
}

var file_xyz_block_ftl_v1_module_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_xyz_block_ftl_v1_module_proto_msgTypes = make([]protoimpl.MessageInfo, 10)
var file_xyz_block_ftl_v1_module_proto_goTypes = []any{
	(ModuleContextResponse_DBType)(0), // 0: xyz.block.ftl.v1.ModuleContextResponse.DBType
	(*AcquireLeaseRequest)(nil),       // 1: xyz.block.ftl.v1.AcquireLeaseRequest
	(*AcquireLeaseResponse)(nil),      // 2: xyz.block.ftl.v1.AcquireLeaseResponse
	(*PublishEventRequest)(nil),       // 3: xyz.block.ftl.v1.PublishEventRequest
	(*PublishEventResponse)(nil),      // 4: xyz.block.ftl.v1.PublishEventResponse
	(*ModuleContextRequest)(nil),      // 5: xyz.block.ftl.v1.ModuleContextRequest
	(*ModuleContextResponse)(nil),     // 6: xyz.block.ftl.v1.ModuleContextResponse
	(*ModuleContextResponse_Ref)(nil), // 7: xyz.block.ftl.v1.ModuleContextResponse.Ref
	(*ModuleContextResponse_DSN)(nil), // 8: xyz.block.ftl.v1.ModuleContextResponse.DSN
	nil,                               // 9: xyz.block.ftl.v1.ModuleContextResponse.ConfigsEntry
	nil,                               // 10: xyz.block.ftl.v1.ModuleContextResponse.SecretsEntry
	(*durationpb.Duration)(nil),       // 11: google.protobuf.Duration
	(*schema.Ref)(nil),                // 12: xyz.block.ftl.v1.schema.Ref
	(*PingRequest)(nil),               // 13: xyz.block.ftl.v1.PingRequest
	(*PingResponse)(nil),              // 14: xyz.block.ftl.v1.PingResponse
}
var file_xyz_block_ftl_v1_module_proto_depIdxs = []int32{
	11, // 0: xyz.block.ftl.v1.AcquireLeaseRequest.ttl:type_name -> google.protobuf.Duration
	12, // 1: xyz.block.ftl.v1.PublishEventRequest.topic:type_name -> xyz.block.ftl.v1.schema.Ref
	9,  // 2: xyz.block.ftl.v1.ModuleContextResponse.configs:type_name -> xyz.block.ftl.v1.ModuleContextResponse.ConfigsEntry
	10, // 3: xyz.block.ftl.v1.ModuleContextResponse.secrets:type_name -> xyz.block.ftl.v1.ModuleContextResponse.SecretsEntry
	8,  // 4: xyz.block.ftl.v1.ModuleContextResponse.databases:type_name -> xyz.block.ftl.v1.ModuleContextResponse.DSN
	0,  // 5: xyz.block.ftl.v1.ModuleContextResponse.DSN.type:type_name -> xyz.block.ftl.v1.ModuleContextResponse.DBType
	13, // 6: xyz.block.ftl.v1.ModuleService.Ping:input_type -> xyz.block.ftl.v1.PingRequest
	5,  // 7: xyz.block.ftl.v1.ModuleService.GetModuleContext:input_type -> xyz.block.ftl.v1.ModuleContextRequest
	1,  // 8: xyz.block.ftl.v1.ModuleService.AcquireLease:input_type -> xyz.block.ftl.v1.AcquireLeaseRequest
	3,  // 9: xyz.block.ftl.v1.ModuleService.PublishEvent:input_type -> xyz.block.ftl.v1.PublishEventRequest
	14, // 10: xyz.block.ftl.v1.ModuleService.Ping:output_type -> xyz.block.ftl.v1.PingResponse
	6,  // 11: xyz.block.ftl.v1.ModuleService.GetModuleContext:output_type -> xyz.block.ftl.v1.ModuleContextResponse
	2,  // 12: xyz.block.ftl.v1.ModuleService.AcquireLease:output_type -> xyz.block.ftl.v1.AcquireLeaseResponse
	4,  // 13: xyz.block.ftl.v1.ModuleService.PublishEvent:output_type -> xyz.block.ftl.v1.PublishEventResponse
	10, // [10:14] is the sub-list for method output_type
	6,  // [6:10] is the sub-list for method input_type
	6,  // [6:6] is the sub-list for extension type_name
	6,  // [6:6] is the sub-list for extension extendee
	0,  // [0:6] is the sub-list for field type_name
}

func init() { file_xyz_block_ftl_v1_module_proto_init() }
func file_xyz_block_ftl_v1_module_proto_init() {
	if File_xyz_block_ftl_v1_module_proto != nil {
		return
	}
	file_xyz_block_ftl_v1_ftl_proto_init()
	file_xyz_block_ftl_v1_module_proto_msgTypes[6].OneofWrappers = []any{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_xyz_block_ftl_v1_module_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   10,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_xyz_block_ftl_v1_module_proto_goTypes,
		DependencyIndexes: file_xyz_block_ftl_v1_module_proto_depIdxs,
		EnumInfos:         file_xyz_block_ftl_v1_module_proto_enumTypes,
		MessageInfos:      file_xyz_block_ftl_v1_module_proto_msgTypes,
	}.Build()
	File_xyz_block_ftl_v1_module_proto = out.File
	file_xyz_block_ftl_v1_module_proto_rawDesc = nil
	file_xyz_block_ftl_v1_module_proto_goTypes = nil
	file_xyz_block_ftl_v1_module_proto_depIdxs = nil
}