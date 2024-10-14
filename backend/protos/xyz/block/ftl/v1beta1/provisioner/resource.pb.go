// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.1
// 	protoc        (unknown)
// source: xyz/block/ftl/v1beta1/provisioner/resource.proto

package provisioner

import (
	v1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	schema "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	structpb "google.golang.org/protobuf/types/known/structpb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Resource is an abstract resource extracted from FTL Schema.
type Resource struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// id unique within the module
	ResourceId string `protobuf:"bytes,1,opt,name=resource_id,json=resourceId,proto3" json:"resource_id,omitempty"`
	// Types that are assignable to Resource:
	//
	//	*Resource_Postgres
	//	*Resource_Mysql
	//	*Resource_Module
	Resource isResource_Resource `protobuf_oneof:"resource"`
}

func (x *Resource) Reset() {
	*x = Resource{}
	mi := &file_xyz_block_ftl_v1beta1_provisioner_resource_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Resource) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Resource) ProtoMessage() {}

func (x *Resource) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_v1beta1_provisioner_resource_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Resource.ProtoReflect.Descriptor instead.
func (*Resource) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_v1beta1_provisioner_resource_proto_rawDescGZIP(), []int{0}
}

func (x *Resource) GetResourceId() string {
	if x != nil {
		return x.ResourceId
	}
	return ""
}

func (m *Resource) GetResource() isResource_Resource {
	if m != nil {
		return m.Resource
	}
	return nil
}

func (x *Resource) GetPostgres() *PostgresResource {
	if x, ok := x.GetResource().(*Resource_Postgres); ok {
		return x.Postgres
	}
	return nil
}

func (x *Resource) GetMysql() *MysqlResource {
	if x, ok := x.GetResource().(*Resource_Mysql); ok {
		return x.Mysql
	}
	return nil
}

func (x *Resource) GetModule() *ModuleResource {
	if x, ok := x.GetResource().(*Resource_Module); ok {
		return x.Module
	}
	return nil
}

type isResource_Resource interface {
	isResource_Resource()
}

type Resource_Postgres struct {
	Postgres *PostgresResource `protobuf:"bytes,102,opt,name=postgres,proto3,oneof"`
}

type Resource_Mysql struct {
	Mysql *MysqlResource `protobuf:"bytes,103,opt,name=mysql,proto3,oneof"`
}

type Resource_Module struct {
	Module *ModuleResource `protobuf:"bytes,104,opt,name=module,proto3,oneof"`
}

func (*Resource_Postgres) isResource_Resource() {}

func (*Resource_Mysql) isResource_Resource() {}

func (*Resource_Module) isResource_Resource() {}

type PostgresResource struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Output *PostgresResource_PostgresResourceOutput `protobuf:"bytes,1,opt,name=output,proto3" json:"output,omitempty"`
}

func (x *PostgresResource) Reset() {
	*x = PostgresResource{}
	mi := &file_xyz_block_ftl_v1beta1_provisioner_resource_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PostgresResource) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PostgresResource) ProtoMessage() {}

func (x *PostgresResource) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_v1beta1_provisioner_resource_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PostgresResource.ProtoReflect.Descriptor instead.
func (*PostgresResource) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_v1beta1_provisioner_resource_proto_rawDescGZIP(), []int{1}
}

func (x *PostgresResource) GetOutput() *PostgresResource_PostgresResourceOutput {
	if x != nil {
		return x.Output
	}
	return nil
}

type MysqlResource struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Output *MysqlResource_MysqlResourceOutput `protobuf:"bytes,1,opt,name=output,proto3" json:"output,omitempty"`
}

func (x *MysqlResource) Reset() {
	*x = MysqlResource{}
	mi := &file_xyz_block_ftl_v1beta1_provisioner_resource_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *MysqlResource) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MysqlResource) ProtoMessage() {}

func (x *MysqlResource) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_v1beta1_provisioner_resource_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MysqlResource.ProtoReflect.Descriptor instead.
func (*MysqlResource) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_v1beta1_provisioner_resource_proto_rawDescGZIP(), []int{2}
}

func (x *MysqlResource) GetOutput() *MysqlResource_MysqlResourceOutput {
	if x != nil {
		return x.Output
	}
	return nil
}

type ModuleResource struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Output    *ModuleResource_ModuleResourceOutput `protobuf:"bytes,1,opt,name=output,proto3" json:"output,omitempty"`
	Schema    *schema.Module                       `protobuf:"bytes,2,opt,name=schema,proto3" json:"schema,omitempty"`
	Artefacts []*v1.DeploymentArtefact             `protobuf:"bytes,3,rep,name=artefacts,proto3" json:"artefacts,omitempty"`
	// Runner labels required to run this deployment.
	Labels *structpb.Struct `protobuf:"bytes,4,opt,name=labels,proto3,oneof" json:"labels,omitempty"`
}

func (x *ModuleResource) Reset() {
	*x = ModuleResource{}
	mi := &file_xyz_block_ftl_v1beta1_provisioner_resource_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ModuleResource) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ModuleResource) ProtoMessage() {}

func (x *ModuleResource) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_v1beta1_provisioner_resource_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ModuleResource.ProtoReflect.Descriptor instead.
func (*ModuleResource) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_v1beta1_provisioner_resource_proto_rawDescGZIP(), []int{3}
}

func (x *ModuleResource) GetOutput() *ModuleResource_ModuleResourceOutput {
	if x != nil {
		return x.Output
	}
	return nil
}

func (x *ModuleResource) GetSchema() *schema.Module {
	if x != nil {
		return x.Schema
	}
	return nil
}

func (x *ModuleResource) GetArtefacts() []*v1.DeploymentArtefact {
	if x != nil {
		return x.Artefacts
	}
	return nil
}

func (x *ModuleResource) GetLabels() *structpb.Struct {
	if x != nil {
		return x.Labels
	}
	return nil
}

type PostgresResource_PostgresResourceOutput struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ReadEndpoint  string `protobuf:"bytes,1,opt,name=read_endpoint,json=readEndpoint,proto3" json:"read_endpoint,omitempty"`
	WriteEndpoint string `protobuf:"bytes,2,opt,name=write_endpoint,json=writeEndpoint,proto3" json:"write_endpoint,omitempty"`
}

func (x *PostgresResource_PostgresResourceOutput) Reset() {
	*x = PostgresResource_PostgresResourceOutput{}
	mi := &file_xyz_block_ftl_v1beta1_provisioner_resource_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PostgresResource_PostgresResourceOutput) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PostgresResource_PostgresResourceOutput) ProtoMessage() {}

func (x *PostgresResource_PostgresResourceOutput) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_v1beta1_provisioner_resource_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PostgresResource_PostgresResourceOutput.ProtoReflect.Descriptor instead.
func (*PostgresResource_PostgresResourceOutput) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_v1beta1_provisioner_resource_proto_rawDescGZIP(), []int{1, 0}
}

func (x *PostgresResource_PostgresResourceOutput) GetReadEndpoint() string {
	if x != nil {
		return x.ReadEndpoint
	}
	return ""
}

func (x *PostgresResource_PostgresResourceOutput) GetWriteEndpoint() string {
	if x != nil {
		return x.WriteEndpoint
	}
	return ""
}

type MysqlResource_MysqlResourceOutput struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ReadEndpoint  string `protobuf:"bytes,1,opt,name=read_endpoint,json=readEndpoint,proto3" json:"read_endpoint,omitempty"`
	WriteEndpoint string `protobuf:"bytes,2,opt,name=write_endpoint,json=writeEndpoint,proto3" json:"write_endpoint,omitempty"`
}

func (x *MysqlResource_MysqlResourceOutput) Reset() {
	*x = MysqlResource_MysqlResourceOutput{}
	mi := &file_xyz_block_ftl_v1beta1_provisioner_resource_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *MysqlResource_MysqlResourceOutput) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MysqlResource_MysqlResourceOutput) ProtoMessage() {}

func (x *MysqlResource_MysqlResourceOutput) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_v1beta1_provisioner_resource_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MysqlResource_MysqlResourceOutput.ProtoReflect.Descriptor instead.
func (*MysqlResource_MysqlResourceOutput) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_v1beta1_provisioner_resource_proto_rawDescGZIP(), []int{2, 0}
}

func (x *MysqlResource_MysqlResourceOutput) GetReadEndpoint() string {
	if x != nil {
		return x.ReadEndpoint
	}
	return ""
}

func (x *MysqlResource_MysqlResourceOutput) GetWriteEndpoint() string {
	if x != nil {
		return x.WriteEndpoint
	}
	return ""
}

type ModuleResource_ModuleResourceOutput struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	DeploymentKey string `protobuf:"bytes,1,opt,name=deployment_key,json=deploymentKey,proto3" json:"deployment_key,omitempty"`
}

func (x *ModuleResource_ModuleResourceOutput) Reset() {
	*x = ModuleResource_ModuleResourceOutput{}
	mi := &file_xyz_block_ftl_v1beta1_provisioner_resource_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ModuleResource_ModuleResourceOutput) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ModuleResource_ModuleResourceOutput) ProtoMessage() {}

func (x *ModuleResource_ModuleResourceOutput) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_v1beta1_provisioner_resource_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ModuleResource_ModuleResourceOutput.ProtoReflect.Descriptor instead.
func (*ModuleResource_ModuleResourceOutput) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_v1beta1_provisioner_resource_proto_rawDescGZIP(), []int{3, 0}
}

func (x *ModuleResource_ModuleResourceOutput) GetDeploymentKey() string {
	if x != nil {
		return x.DeploymentKey
	}
	return ""
}

var File_xyz_block_ftl_v1beta1_provisioner_resource_proto protoreflect.FileDescriptor

var file_xyz_block_ftl_v1beta1_provisioner_resource_proto_rawDesc = []byte{
	0x0a, 0x30, 0x78, 0x79, 0x7a, 0x2f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2f, 0x66, 0x74, 0x6c, 0x2f,
	0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2f, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x73, 0x69, 0x6f,
	0x6e, 0x65, 0x72, 0x2f, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x12, 0x21, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74,
	0x6c, 0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x73,
	0x69, 0x6f, 0x6e, 0x65, 0x72, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x73, 0x74, 0x72, 0x75, 0x63, 0x74, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x1a, 0x1a, 0x78, 0x79, 0x7a, 0x2f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2f, 0x66,
	0x74, 0x6c, 0x2f, 0x76, 0x31, 0x2f, 0x66, 0x74, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x24, 0x78, 0x79, 0x7a, 0x2f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2f, 0x66, 0x74, 0x6c, 0x2f, 0x76,
	0x31, 0x2f, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x2f, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xa1, 0x02, 0x0a, 0x08, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72,
	0x63, 0x65, 0x12, 0x1f, 0x0a, 0x0b, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x5f, 0x69,
	0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63,
	0x65, 0x49, 0x64, 0x12, 0x51, 0x0a, 0x08, 0x70, 0x6f, 0x73, 0x74, 0x67, 0x72, 0x65, 0x73, 0x18,
	0x66, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x33, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63,
	0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x70, 0x72,
	0x6f, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x65, 0x72, 0x2e, 0x50, 0x6f, 0x73, 0x74, 0x67, 0x72,
	0x65, 0x73, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x48, 0x00, 0x52, 0x08, 0x70, 0x6f,
	0x73, 0x74, 0x67, 0x72, 0x65, 0x73, 0x12, 0x48, 0x0a, 0x05, 0x6d, 0x79, 0x73, 0x71, 0x6c, 0x18,
	0x67, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x30, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63,
	0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x70, 0x72,
	0x6f, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x65, 0x72, 0x2e, 0x4d, 0x79, 0x73, 0x71, 0x6c, 0x52,
	0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x48, 0x00, 0x52, 0x05, 0x6d, 0x79, 0x73, 0x71, 0x6c,
	0x12, 0x4b, 0x0a, 0x06, 0x6d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x18, 0x68, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x31, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c,
	0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x73, 0x69,
	0x6f, 0x6e, 0x65, 0x72, 0x2e, 0x4d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x52, 0x65, 0x73, 0x6f, 0x75,
	0x72, 0x63, 0x65, 0x48, 0x00, 0x52, 0x06, 0x6d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x42, 0x0a, 0x0a,
	0x08, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x22, 0xdc, 0x01, 0x0a, 0x10, 0x50, 0x6f,
	0x73, 0x74, 0x67, 0x72, 0x65, 0x73, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x12, 0x62,
	0x0a, 0x06, 0x6f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x4a,
	0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76,
	0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e,
	0x65, 0x72, 0x2e, 0x50, 0x6f, 0x73, 0x74, 0x67, 0x72, 0x65, 0x73, 0x52, 0x65, 0x73, 0x6f, 0x75,
	0x72, 0x63, 0x65, 0x2e, 0x50, 0x6f, 0x73, 0x74, 0x67, 0x72, 0x65, 0x73, 0x52, 0x65, 0x73, 0x6f,
	0x75, 0x72, 0x63, 0x65, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x52, 0x06, 0x6f, 0x75, 0x74, 0x70,
	0x75, 0x74, 0x1a, 0x64, 0x0a, 0x16, 0x50, 0x6f, 0x73, 0x74, 0x67, 0x72, 0x65, 0x73, 0x52, 0x65,
	0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x12, 0x23, 0x0a, 0x0d,
	0x72, 0x65, 0x61, 0x64, 0x5f, 0x65, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x0c, 0x72, 0x65, 0x61, 0x64, 0x45, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e,
	0x74, 0x12, 0x25, 0x0a, 0x0e, 0x77, 0x72, 0x69, 0x74, 0x65, 0x5f, 0x65, 0x6e, 0x64, 0x70, 0x6f,
	0x69, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x77, 0x72, 0x69, 0x74, 0x65,
	0x45, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x22, 0xd0, 0x01, 0x0a, 0x0d, 0x4d, 0x79, 0x73,
	0x71, 0x6c, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x12, 0x5c, 0x0a, 0x06, 0x6f, 0x75,
	0x74, 0x70, 0x75, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x44, 0x2e, 0x78, 0x79, 0x7a,
	0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76, 0x31, 0x62, 0x65, 0x74,
	0x61, 0x31, 0x2e, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x65, 0x72, 0x2e, 0x4d,
	0x79, 0x73, 0x71, 0x6c, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x2e, 0x4d, 0x79, 0x73,
	0x71, 0x6c, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74,
	0x52, 0x06, 0x6f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x1a, 0x61, 0x0a, 0x13, 0x4d, 0x79, 0x73, 0x71,
	0x6c, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x12,
	0x23, 0x0a, 0x0d, 0x72, 0x65, 0x61, 0x64, 0x5f, 0x65, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x72, 0x65, 0x61, 0x64, 0x45, 0x6e, 0x64, 0x70,
	0x6f, 0x69, 0x6e, 0x74, 0x12, 0x25, 0x0a, 0x0e, 0x77, 0x72, 0x69, 0x74, 0x65, 0x5f, 0x65, 0x6e,
	0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x77, 0x72,
	0x69, 0x74, 0x65, 0x45, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x22, 0xed, 0x02, 0x0a, 0x0e,
	0x4d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x12, 0x5e,
	0x0a, 0x06, 0x6f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x46,
	0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76,
	0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e,
	0x65, 0x72, 0x2e, 0x4d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63,
	0x65, 0x2e, 0x4d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65,
	0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x52, 0x06, 0x6f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x12, 0x37,
	0x0a, 0x06, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1f,
	0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76,
	0x31, 0x2e, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x2e, 0x4d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x52,
	0x06, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x12, 0x42, 0x0a, 0x09, 0x61, 0x72, 0x74, 0x65, 0x66,
	0x61, 0x63, 0x74, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x24, 0x2e, 0x78, 0x79, 0x7a,
	0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x44, 0x65,
	0x70, 0x6c, 0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x41, 0x72, 0x74, 0x65, 0x66, 0x61, 0x63, 0x74,
	0x52, 0x09, 0x61, 0x72, 0x74, 0x65, 0x66, 0x61, 0x63, 0x74, 0x73, 0x12, 0x34, 0x0a, 0x06, 0x6c,
	0x61, 0x62, 0x65, 0x6c, 0x73, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x53, 0x74,
	0x72, 0x75, 0x63, 0x74, 0x48, 0x00, 0x52, 0x06, 0x6c, 0x61, 0x62, 0x65, 0x6c, 0x73, 0x88, 0x01,
	0x01, 0x1a, 0x3d, 0x0a, 0x14, 0x4d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x52, 0x65, 0x73, 0x6f, 0x75,
	0x72, 0x63, 0x65, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x12, 0x25, 0x0a, 0x0e, 0x64, 0x65, 0x70,
	0x6c, 0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x0d, 0x64, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x4b, 0x65, 0x79,
	0x42, 0x09, 0x0a, 0x07, 0x5f, 0x6c, 0x61, 0x62, 0x65, 0x6c, 0x73, 0x42, 0x5b, 0x50, 0x01, 0x5a,
	0x57, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x54, 0x42, 0x44, 0x35,
	0x34, 0x35, 0x36, 0x36, 0x39, 0x37, 0x35, 0x2f, 0x66, 0x74, 0x6c, 0x2f, 0x62, 0x61, 0x63, 0x6b,
	0x65, 0x6e, 0x64, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2f, 0x78, 0x79, 0x7a, 0x2f, 0x62,
	0x6c, 0x6f, 0x63, 0x6b, 0x2f, 0x66, 0x74, 0x6c, 0x2f, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31,
	0x2f, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x65, 0x72, 0x3b, 0x70, 0x72, 0x6f,
	0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x65, 0x72, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_xyz_block_ftl_v1beta1_provisioner_resource_proto_rawDescOnce sync.Once
	file_xyz_block_ftl_v1beta1_provisioner_resource_proto_rawDescData = file_xyz_block_ftl_v1beta1_provisioner_resource_proto_rawDesc
)

func file_xyz_block_ftl_v1beta1_provisioner_resource_proto_rawDescGZIP() []byte {
	file_xyz_block_ftl_v1beta1_provisioner_resource_proto_rawDescOnce.Do(func() {
		file_xyz_block_ftl_v1beta1_provisioner_resource_proto_rawDescData = protoimpl.X.CompressGZIP(file_xyz_block_ftl_v1beta1_provisioner_resource_proto_rawDescData)
	})
	return file_xyz_block_ftl_v1beta1_provisioner_resource_proto_rawDescData
}

var file_xyz_block_ftl_v1beta1_provisioner_resource_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_xyz_block_ftl_v1beta1_provisioner_resource_proto_goTypes = []any{
	(*Resource)(nil),                                // 0: xyz.block.ftl.v1beta1.provisioner.Resource
	(*PostgresResource)(nil),                        // 1: xyz.block.ftl.v1beta1.provisioner.PostgresResource
	(*MysqlResource)(nil),                           // 2: xyz.block.ftl.v1beta1.provisioner.MysqlResource
	(*ModuleResource)(nil),                          // 3: xyz.block.ftl.v1beta1.provisioner.ModuleResource
	(*PostgresResource_PostgresResourceOutput)(nil), // 4: xyz.block.ftl.v1beta1.provisioner.PostgresResource.PostgresResourceOutput
	(*MysqlResource_MysqlResourceOutput)(nil),       // 5: xyz.block.ftl.v1beta1.provisioner.MysqlResource.MysqlResourceOutput
	(*ModuleResource_ModuleResourceOutput)(nil),     // 6: xyz.block.ftl.v1beta1.provisioner.ModuleResource.ModuleResourceOutput
	(*schema.Module)(nil),                           // 7: xyz.block.ftl.v1.schema.Module
	(*v1.DeploymentArtefact)(nil),                   // 8: xyz.block.ftl.v1.DeploymentArtefact
	(*structpb.Struct)(nil),                         // 9: google.protobuf.Struct
}
var file_xyz_block_ftl_v1beta1_provisioner_resource_proto_depIdxs = []int32{
	1, // 0: xyz.block.ftl.v1beta1.provisioner.Resource.postgres:type_name -> xyz.block.ftl.v1beta1.provisioner.PostgresResource
	2, // 1: xyz.block.ftl.v1beta1.provisioner.Resource.mysql:type_name -> xyz.block.ftl.v1beta1.provisioner.MysqlResource
	3, // 2: xyz.block.ftl.v1beta1.provisioner.Resource.module:type_name -> xyz.block.ftl.v1beta1.provisioner.ModuleResource
	4, // 3: xyz.block.ftl.v1beta1.provisioner.PostgresResource.output:type_name -> xyz.block.ftl.v1beta1.provisioner.PostgresResource.PostgresResourceOutput
	5, // 4: xyz.block.ftl.v1beta1.provisioner.MysqlResource.output:type_name -> xyz.block.ftl.v1beta1.provisioner.MysqlResource.MysqlResourceOutput
	6, // 5: xyz.block.ftl.v1beta1.provisioner.ModuleResource.output:type_name -> xyz.block.ftl.v1beta1.provisioner.ModuleResource.ModuleResourceOutput
	7, // 6: xyz.block.ftl.v1beta1.provisioner.ModuleResource.schema:type_name -> xyz.block.ftl.v1.schema.Module
	8, // 7: xyz.block.ftl.v1beta1.provisioner.ModuleResource.artefacts:type_name -> xyz.block.ftl.v1.DeploymentArtefact
	9, // 8: xyz.block.ftl.v1beta1.provisioner.ModuleResource.labels:type_name -> google.protobuf.Struct
	9, // [9:9] is the sub-list for method output_type
	9, // [9:9] is the sub-list for method input_type
	9, // [9:9] is the sub-list for extension type_name
	9, // [9:9] is the sub-list for extension extendee
	0, // [0:9] is the sub-list for field type_name
}

func init() { file_xyz_block_ftl_v1beta1_provisioner_resource_proto_init() }
func file_xyz_block_ftl_v1beta1_provisioner_resource_proto_init() {
	if File_xyz_block_ftl_v1beta1_provisioner_resource_proto != nil {
		return
	}
	file_xyz_block_ftl_v1beta1_provisioner_resource_proto_msgTypes[0].OneofWrappers = []any{
		(*Resource_Postgres)(nil),
		(*Resource_Mysql)(nil),
		(*Resource_Module)(nil),
	}
	file_xyz_block_ftl_v1beta1_provisioner_resource_proto_msgTypes[3].OneofWrappers = []any{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_xyz_block_ftl_v1beta1_provisioner_resource_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_xyz_block_ftl_v1beta1_provisioner_resource_proto_goTypes,
		DependencyIndexes: file_xyz_block_ftl_v1beta1_provisioner_resource_proto_depIdxs,
		MessageInfos:      file_xyz_block_ftl_v1beta1_provisioner_resource_proto_msgTypes,
	}.Build()
	File_xyz_block_ftl_v1beta1_provisioner_resource_proto = out.File
	file_xyz_block_ftl_v1beta1_provisioner_resource_proto_rawDesc = nil
	file_xyz_block_ftl_v1beta1_provisioner_resource_proto_goTypes = nil
	file_xyz_block_ftl_v1beta1_provisioner_resource_proto_depIdxs = nil
}
