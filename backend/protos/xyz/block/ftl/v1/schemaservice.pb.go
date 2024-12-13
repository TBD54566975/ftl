// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.2
// 	protoc        (unknown)
// source: xyz/block/ftl/v1/schemaservice.proto

package ftlv1

import (
	v1 "github.com/TBD54566975/ftl/common/protos/xyz/block/ftl/schema/v1"
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

type DeploymentChangeType int32

const (
	DeploymentChangeType_DEPLOYMENT_CHANGE_TYPE_UNSPECIFIED DeploymentChangeType = 0
	DeploymentChangeType_DEPLOYMENT_CHANGE_TYPE_ADDED       DeploymentChangeType = 1
	DeploymentChangeType_DEPLOYMENT_CHANGE_TYPE_REMOVED     DeploymentChangeType = 2
	DeploymentChangeType_DEPLOYMENT_CHANGE_TYPE_CHANGED     DeploymentChangeType = 3
)

// Enum value maps for DeploymentChangeType.
var (
	DeploymentChangeType_name = map[int32]string{
		0: "DEPLOYMENT_CHANGE_TYPE_UNSPECIFIED",
		1: "DEPLOYMENT_CHANGE_TYPE_ADDED",
		2: "DEPLOYMENT_CHANGE_TYPE_REMOVED",
		3: "DEPLOYMENT_CHANGE_TYPE_CHANGED",
	}
	DeploymentChangeType_value = map[string]int32{
		"DEPLOYMENT_CHANGE_TYPE_UNSPECIFIED": 0,
		"DEPLOYMENT_CHANGE_TYPE_ADDED":       1,
		"DEPLOYMENT_CHANGE_TYPE_REMOVED":     2,
		"DEPLOYMENT_CHANGE_TYPE_CHANGED":     3,
	}
)

func (x DeploymentChangeType) Enum() *DeploymentChangeType {
	p := new(DeploymentChangeType)
	*p = x
	return p
}

func (x DeploymentChangeType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (DeploymentChangeType) Descriptor() protoreflect.EnumDescriptor {
	return file_xyz_block_ftl_v1_schemaservice_proto_enumTypes[0].Descriptor()
}

func (DeploymentChangeType) Type() protoreflect.EnumType {
	return &file_xyz_block_ftl_v1_schemaservice_proto_enumTypes[0]
}

func (x DeploymentChangeType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use DeploymentChangeType.Descriptor instead.
func (DeploymentChangeType) EnumDescriptor() ([]byte, []int) {
	return file_xyz_block_ftl_v1_schemaservice_proto_rawDescGZIP(), []int{0}
}

type GetSchemaRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *GetSchemaRequest) Reset() {
	*x = GetSchemaRequest{}
	mi := &file_xyz_block_ftl_v1_schemaservice_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetSchemaRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetSchemaRequest) ProtoMessage() {}

func (x *GetSchemaRequest) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_v1_schemaservice_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetSchemaRequest.ProtoReflect.Descriptor instead.
func (*GetSchemaRequest) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_v1_schemaservice_proto_rawDescGZIP(), []int{0}
}

type GetSchemaResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Schema *v1.Schema `protobuf:"bytes,1,opt,name=schema,proto3" json:"schema,omitempty"`
}

func (x *GetSchemaResponse) Reset() {
	*x = GetSchemaResponse{}
	mi := &file_xyz_block_ftl_v1_schemaservice_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetSchemaResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetSchemaResponse) ProtoMessage() {}

func (x *GetSchemaResponse) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_v1_schemaservice_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetSchemaResponse.ProtoReflect.Descriptor instead.
func (*GetSchemaResponse) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_v1_schemaservice_proto_rawDescGZIP(), []int{1}
}

func (x *GetSchemaResponse) GetSchema() *v1.Schema {
	if x != nil {
		return x.Schema
	}
	return nil
}

type PullSchemaRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *PullSchemaRequest) Reset() {
	*x = PullSchemaRequest{}
	mi := &file_xyz_block_ftl_v1_schemaservice_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PullSchemaRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PullSchemaRequest) ProtoMessage() {}

func (x *PullSchemaRequest) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_v1_schemaservice_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PullSchemaRequest.ProtoReflect.Descriptor instead.
func (*PullSchemaRequest) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_v1_schemaservice_proto_rawDescGZIP(), []int{2}
}

type PullSchemaResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Will not be set for builtin modules.
	DeploymentKey *string `protobuf:"bytes,1,opt,name=deployment_key,json=deploymentKey,proto3,oneof" json:"deployment_key,omitempty"`
	ModuleName    string  `protobuf:"bytes,2,opt,name=module_name,json=moduleName,proto3" json:"module_name,omitempty"`
	// For deletes this will not be present.
	Schema *v1.Module `protobuf:"bytes,4,opt,name=schema,proto3,oneof" json:"schema,omitempty"`
	// If true there are more schema changes immediately following this one as part of the initial batch.
	// If false this is the last schema change in the initial batch, but others may follow later.
	More       bool                 `protobuf:"varint,3,opt,name=more,proto3" json:"more,omitempty"`
	ChangeType DeploymentChangeType `protobuf:"varint,5,opt,name=change_type,json=changeType,proto3,enum=xyz.block.ftl.v1.DeploymentChangeType" json:"change_type,omitempty"`
	// If this is true then the module was removed as well as the deployment. This is only set for DEPLOYMENT_REMOVED.
	ModuleRemoved bool `protobuf:"varint,6,opt,name=module_removed,json=moduleRemoved,proto3" json:"module_removed,omitempty"`
}

func (x *PullSchemaResponse) Reset() {
	*x = PullSchemaResponse{}
	mi := &file_xyz_block_ftl_v1_schemaservice_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PullSchemaResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PullSchemaResponse) ProtoMessage() {}

func (x *PullSchemaResponse) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_v1_schemaservice_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PullSchemaResponse.ProtoReflect.Descriptor instead.
func (*PullSchemaResponse) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_v1_schemaservice_proto_rawDescGZIP(), []int{3}
}

func (x *PullSchemaResponse) GetDeploymentKey() string {
	if x != nil && x.DeploymentKey != nil {
		return *x.DeploymentKey
	}
	return ""
}

func (x *PullSchemaResponse) GetModuleName() string {
	if x != nil {
		return x.ModuleName
	}
	return ""
}

func (x *PullSchemaResponse) GetSchema() *v1.Module {
	if x != nil {
		return x.Schema
	}
	return nil
}

func (x *PullSchemaResponse) GetMore() bool {
	if x != nil {
		return x.More
	}
	return false
}

func (x *PullSchemaResponse) GetChangeType() DeploymentChangeType {
	if x != nil {
		return x.ChangeType
	}
	return DeploymentChangeType_DEPLOYMENT_CHANGE_TYPE_UNSPECIFIED
}

func (x *PullSchemaResponse) GetModuleRemoved() bool {
	if x != nil {
		return x.ModuleRemoved
	}
	return false
}

type UpdateDeploymentRuntimeRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Deployment string                 `protobuf:"bytes,1,opt,name=deployment,proto3" json:"deployment,omitempty"`
	Event      *v1.ModuleRuntimeEvent `protobuf:"bytes,2,opt,name=event,proto3" json:"event,omitempty"`
}

func (x *UpdateDeploymentRuntimeRequest) Reset() {
	*x = UpdateDeploymentRuntimeRequest{}
	mi := &file_xyz_block_ftl_v1_schemaservice_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UpdateDeploymentRuntimeRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateDeploymentRuntimeRequest) ProtoMessage() {}

func (x *UpdateDeploymentRuntimeRequest) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_v1_schemaservice_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateDeploymentRuntimeRequest.ProtoReflect.Descriptor instead.
func (*UpdateDeploymentRuntimeRequest) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_v1_schemaservice_proto_rawDescGZIP(), []int{4}
}

func (x *UpdateDeploymentRuntimeRequest) GetDeployment() string {
	if x != nil {
		return x.Deployment
	}
	return ""
}

func (x *UpdateDeploymentRuntimeRequest) GetEvent() *v1.ModuleRuntimeEvent {
	if x != nil {
		return x.Event
	}
	return nil
}

type UpdateDeploymentRuntimeResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *UpdateDeploymentRuntimeResponse) Reset() {
	*x = UpdateDeploymentRuntimeResponse{}
	mi := &file_xyz_block_ftl_v1_schemaservice_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UpdateDeploymentRuntimeResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateDeploymentRuntimeResponse) ProtoMessage() {}

func (x *UpdateDeploymentRuntimeResponse) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_v1_schemaservice_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateDeploymentRuntimeResponse.ProtoReflect.Descriptor instead.
func (*UpdateDeploymentRuntimeResponse) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_v1_schemaservice_proto_rawDescGZIP(), []int{5}
}

var File_xyz_block_ftl_v1_schemaservice_proto protoreflect.FileDescriptor

var file_xyz_block_ftl_v1_schemaservice_proto_rawDesc = []byte{
	0x0a, 0x24, 0x78, 0x79, 0x7a, 0x2f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2f, 0x66, 0x74, 0x6c, 0x2f,
	0x76, 0x31, 0x2f, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x10, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63,
	0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76, 0x31, 0x1a, 0x24, 0x78, 0x79, 0x7a, 0x2f, 0x62, 0x6c,
	0x6f, 0x63, 0x6b, 0x2f, 0x66, 0x74, 0x6c, 0x2f, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x2f, 0x76,
	0x31, 0x2f, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1a,
	0x78, 0x79, 0x7a, 0x2f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2f, 0x66, 0x74, 0x6c, 0x2f, 0x76, 0x31,
	0x2f, 0x66, 0x74, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x12, 0x0a, 0x10, 0x47, 0x65,
	0x74, 0x53, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x4c,
	0x0a, 0x11, 0x47, 0x65, 0x74, 0x53, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x12, 0x37, 0x0a, 0x06, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x1f, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e,
	0x66, 0x74, 0x6c, 0x2e, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x63,
	0x68, 0x65, 0x6d, 0x61, 0x52, 0x06, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x22, 0x13, 0x0a, 0x11,
	0x50, 0x75, 0x6c, 0x6c, 0x53, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x22, 0xc1, 0x02, 0x0a, 0x12, 0x50, 0x75, 0x6c, 0x6c, 0x53, 0x63, 0x68, 0x65, 0x6d, 0x61,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2a, 0x0a, 0x0e, 0x64, 0x65, 0x70, 0x6c,
	0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x48, 0x00, 0x52, 0x0d, 0x64, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x4b, 0x65,
	0x79, 0x88, 0x01, 0x01, 0x12, 0x1f, 0x0a, 0x0b, 0x6d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x5f, 0x6e,
	0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x6d, 0x6f, 0x64, 0x75, 0x6c,
	0x65, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x3c, 0x0a, 0x06, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x18,
	0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1f, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63,
	0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x2e, 0x76, 0x31, 0x2e,
	0x4d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x48, 0x01, 0x52, 0x06, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61,
	0x88, 0x01, 0x01, 0x12, 0x12, 0x0a, 0x04, 0x6d, 0x6f, 0x72, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x08, 0x52, 0x04, 0x6d, 0x6f, 0x72, 0x65, 0x12, 0x47, 0x0a, 0x0b, 0x63, 0x68, 0x61, 0x6e, 0x67,
	0x65, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x26, 0x2e, 0x78,
	0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76, 0x31, 0x2e,
	0x44, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65,
	0x54, 0x79, 0x70, 0x65, 0x52, 0x0a, 0x63, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x54, 0x79, 0x70, 0x65,
	0x12, 0x25, 0x0a, 0x0e, 0x6d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x5f, 0x72, 0x65, 0x6d, 0x6f, 0x76,
	0x65, 0x64, 0x18, 0x06, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0d, 0x6d, 0x6f, 0x64, 0x75, 0x6c, 0x65,
	0x52, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x64, 0x42, 0x11, 0x0a, 0x0f, 0x5f, 0x64, 0x65, 0x70, 0x6c,
	0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x5f, 0x6b, 0x65, 0x79, 0x42, 0x09, 0x0a, 0x07, 0x5f, 0x73,
	0x63, 0x68, 0x65, 0x6d, 0x61, 0x22, 0x83, 0x01, 0x0a, 0x1e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65,
	0x44, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x52, 0x75, 0x6e, 0x74, 0x69, 0x6d,
	0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1e, 0x0a, 0x0a, 0x64, 0x65, 0x70, 0x6c,
	0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x64, 0x65,
	0x70, 0x6c, 0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x12, 0x41, 0x0a, 0x05, 0x65, 0x76, 0x65, 0x6e,
	0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2b, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c,
	0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x2e, 0x76,
	0x31, 0x2e, 0x4d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x52, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x45,
	0x76, 0x65, 0x6e, 0x74, 0x52, 0x05, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x22, 0x21, 0x0a, 0x1f, 0x55,
	0x70, 0x64, 0x61, 0x74, 0x65, 0x44, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x52,
	0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x2a, 0xa8,
	0x01, 0x0a, 0x14, 0x44, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x43, 0x68, 0x61,
	0x6e, 0x67, 0x65, 0x54, 0x79, 0x70, 0x65, 0x12, 0x26, 0x0a, 0x22, 0x44, 0x45, 0x50, 0x4c, 0x4f,
	0x59, 0x4d, 0x45, 0x4e, 0x54, 0x5f, 0x43, 0x48, 0x41, 0x4e, 0x47, 0x45, 0x5f, 0x54, 0x59, 0x50,
	0x45, 0x5f, 0x55, 0x4e, 0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12,
	0x20, 0x0a, 0x1c, 0x44, 0x45, 0x50, 0x4c, 0x4f, 0x59, 0x4d, 0x45, 0x4e, 0x54, 0x5f, 0x43, 0x48,
	0x41, 0x4e, 0x47, 0x45, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x41, 0x44, 0x44, 0x45, 0x44, 0x10,
	0x01, 0x12, 0x22, 0x0a, 0x1e, 0x44, 0x45, 0x50, 0x4c, 0x4f, 0x59, 0x4d, 0x45, 0x4e, 0x54, 0x5f,
	0x43, 0x48, 0x41, 0x4e, 0x47, 0x45, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x52, 0x45, 0x4d, 0x4f,
	0x56, 0x45, 0x44, 0x10, 0x02, 0x12, 0x22, 0x0a, 0x1e, 0x44, 0x45, 0x50, 0x4c, 0x4f, 0x59, 0x4d,
	0x45, 0x4e, 0x54, 0x5f, 0x43, 0x48, 0x41, 0x4e, 0x47, 0x45, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f,
	0x43, 0x48, 0x41, 0x4e, 0x47, 0x45, 0x44, 0x10, 0x03, 0x32, 0x96, 0x03, 0x0a, 0x0d, 0x53, 0x63,
	0x68, 0x65, 0x6d, 0x61, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x4a, 0x0a, 0x04, 0x50,
	0x69, 0x6e, 0x67, 0x12, 0x1d, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e,
	0x66, 0x74, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x1e, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66,
	0x74, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x22, 0x03, 0x90, 0x02, 0x01, 0x12, 0x59, 0x0a, 0x09, 0x47, 0x65, 0x74, 0x53, 0x63,
	0x68, 0x65, 0x6d, 0x61, 0x12, 0x22, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b,
	0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x53, 0x63, 0x68, 0x65, 0x6d,
	0x61, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x23, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62,
	0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x53,
	0x63, 0x68, 0x65, 0x6d, 0x61, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x03, 0x90,
	0x02, 0x01, 0x12, 0x5e, 0x0a, 0x0a, 0x50, 0x75, 0x6c, 0x6c, 0x53, 0x63, 0x68, 0x65, 0x6d, 0x61,
	0x12, 0x23, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c,
	0x2e, 0x76, 0x31, 0x2e, 0x50, 0x75, 0x6c, 0x6c, 0x53, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x24, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63,
	0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x75, 0x6c, 0x6c, 0x53, 0x63, 0x68,
	0x65, 0x6d, 0x61, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x03, 0x90, 0x02, 0x01,
	0x30, 0x01, 0x12, 0x7e, 0x0a, 0x17, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x44, 0x65, 0x70, 0x6c,
	0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x52, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x12, 0x30, 0x2e,
	0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76, 0x31,
	0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x44, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x6d, 0x65, 0x6e,
	0x74, 0x52, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x31, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e,
	0x76, 0x31, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x44, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x6d,
	0x65, 0x6e, 0x74, 0x52, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x42, 0x44, 0x50, 0x01, 0x5a, 0x40, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63,
	0x6f, 0x6d, 0x2f, 0x54, 0x42, 0x44, 0x35, 0x34, 0x35, 0x36, 0x36, 0x39, 0x37, 0x35, 0x2f, 0x66,
	0x74, 0x6c, 0x2f, 0x62, 0x61, 0x63, 0x6b, 0x65, 0x6e, 0x64, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x73, 0x2f, 0x78, 0x79, 0x7a, 0x2f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2f, 0x66, 0x74, 0x6c, 0x2f,
	0x76, 0x31, 0x3b, 0x66, 0x74, 0x6c, 0x76, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_xyz_block_ftl_v1_schemaservice_proto_rawDescOnce sync.Once
	file_xyz_block_ftl_v1_schemaservice_proto_rawDescData = file_xyz_block_ftl_v1_schemaservice_proto_rawDesc
)

func file_xyz_block_ftl_v1_schemaservice_proto_rawDescGZIP() []byte {
	file_xyz_block_ftl_v1_schemaservice_proto_rawDescOnce.Do(func() {
		file_xyz_block_ftl_v1_schemaservice_proto_rawDescData = protoimpl.X.CompressGZIP(file_xyz_block_ftl_v1_schemaservice_proto_rawDescData)
	})
	return file_xyz_block_ftl_v1_schemaservice_proto_rawDescData
}

var file_xyz_block_ftl_v1_schemaservice_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_xyz_block_ftl_v1_schemaservice_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_xyz_block_ftl_v1_schemaservice_proto_goTypes = []any{
	(DeploymentChangeType)(0),               // 0: xyz.block.ftl.v1.DeploymentChangeType
	(*GetSchemaRequest)(nil),                // 1: xyz.block.ftl.v1.GetSchemaRequest
	(*GetSchemaResponse)(nil),               // 2: xyz.block.ftl.v1.GetSchemaResponse
	(*PullSchemaRequest)(nil),               // 3: xyz.block.ftl.v1.PullSchemaRequest
	(*PullSchemaResponse)(nil),              // 4: xyz.block.ftl.v1.PullSchemaResponse
	(*UpdateDeploymentRuntimeRequest)(nil),  // 5: xyz.block.ftl.v1.UpdateDeploymentRuntimeRequest
	(*UpdateDeploymentRuntimeResponse)(nil), // 6: xyz.block.ftl.v1.UpdateDeploymentRuntimeResponse
	(*v1.Schema)(nil),                       // 7: xyz.block.ftl.schema.v1.Schema
	(*v1.Module)(nil),                       // 8: xyz.block.ftl.schema.v1.Module
	(*v1.ModuleRuntimeEvent)(nil),           // 9: xyz.block.ftl.schema.v1.ModuleRuntimeEvent
	(*PingRequest)(nil),                     // 10: xyz.block.ftl.v1.PingRequest
	(*PingResponse)(nil),                    // 11: xyz.block.ftl.v1.PingResponse
}
var file_xyz_block_ftl_v1_schemaservice_proto_depIdxs = []int32{
	7,  // 0: xyz.block.ftl.v1.GetSchemaResponse.schema:type_name -> xyz.block.ftl.schema.v1.Schema
	8,  // 1: xyz.block.ftl.v1.PullSchemaResponse.schema:type_name -> xyz.block.ftl.schema.v1.Module
	0,  // 2: xyz.block.ftl.v1.PullSchemaResponse.change_type:type_name -> xyz.block.ftl.v1.DeploymentChangeType
	9,  // 3: xyz.block.ftl.v1.UpdateDeploymentRuntimeRequest.event:type_name -> xyz.block.ftl.schema.v1.ModuleRuntimeEvent
	10, // 4: xyz.block.ftl.v1.SchemaService.Ping:input_type -> xyz.block.ftl.v1.PingRequest
	1,  // 5: xyz.block.ftl.v1.SchemaService.GetSchema:input_type -> xyz.block.ftl.v1.GetSchemaRequest
	3,  // 6: xyz.block.ftl.v1.SchemaService.PullSchema:input_type -> xyz.block.ftl.v1.PullSchemaRequest
	5,  // 7: xyz.block.ftl.v1.SchemaService.UpdateDeploymentRuntime:input_type -> xyz.block.ftl.v1.UpdateDeploymentRuntimeRequest
	11, // 8: xyz.block.ftl.v1.SchemaService.Ping:output_type -> xyz.block.ftl.v1.PingResponse
	2,  // 9: xyz.block.ftl.v1.SchemaService.GetSchema:output_type -> xyz.block.ftl.v1.GetSchemaResponse
	4,  // 10: xyz.block.ftl.v1.SchemaService.PullSchema:output_type -> xyz.block.ftl.v1.PullSchemaResponse
	6,  // 11: xyz.block.ftl.v1.SchemaService.UpdateDeploymentRuntime:output_type -> xyz.block.ftl.v1.UpdateDeploymentRuntimeResponse
	8,  // [8:12] is the sub-list for method output_type
	4,  // [4:8] is the sub-list for method input_type
	4,  // [4:4] is the sub-list for extension type_name
	4,  // [4:4] is the sub-list for extension extendee
	0,  // [0:4] is the sub-list for field type_name
}

func init() { file_xyz_block_ftl_v1_schemaservice_proto_init() }
func file_xyz_block_ftl_v1_schemaservice_proto_init() {
	if File_xyz_block_ftl_v1_schemaservice_proto != nil {
		return
	}
	file_xyz_block_ftl_v1_ftl_proto_init()
	file_xyz_block_ftl_v1_schemaservice_proto_msgTypes[3].OneofWrappers = []any{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_xyz_block_ftl_v1_schemaservice_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_xyz_block_ftl_v1_schemaservice_proto_goTypes,
		DependencyIndexes: file_xyz_block_ftl_v1_schemaservice_proto_depIdxs,
		EnumInfos:         file_xyz_block_ftl_v1_schemaservice_proto_enumTypes,
		MessageInfos:      file_xyz_block_ftl_v1_schemaservice_proto_msgTypes,
	}.Build()
	File_xyz_block_ftl_v1_schemaservice_proto = out.File
	file_xyz_block_ftl_v1_schemaservice_proto_rawDesc = nil
	file_xyz_block_ftl_v1_schemaservice_proto_goTypes = nil
	file_xyz_block_ftl_v1_schemaservice_proto_depIdxs = nil
}
