// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.2
// 	protoc        (unknown)
// source: xyz/block/ftl/provisioner/v1beta1/plugin.proto

package provisionerpb

import (
	v1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
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

type ProvisionResponse_ProvisionResponseStatus int32

const (
	ProvisionResponse_PROVISION_RESPONSE_STATUS_UNSPECIFIED ProvisionResponse_ProvisionResponseStatus = 0
	ProvisionResponse_PROVISION_RESPONSE_STATUS_SUBMITTED   ProvisionResponse_ProvisionResponseStatus = 1
)

// Enum value maps for ProvisionResponse_ProvisionResponseStatus.
var (
	ProvisionResponse_ProvisionResponseStatus_name = map[int32]string{
		0: "PROVISION_RESPONSE_STATUS_UNSPECIFIED",
		1: "PROVISION_RESPONSE_STATUS_SUBMITTED",
	}
	ProvisionResponse_ProvisionResponseStatus_value = map[string]int32{
		"PROVISION_RESPONSE_STATUS_UNSPECIFIED": 0,
		"PROVISION_RESPONSE_STATUS_SUBMITTED":   1,
	}
)

func (x ProvisionResponse_ProvisionResponseStatus) Enum() *ProvisionResponse_ProvisionResponseStatus {
	p := new(ProvisionResponse_ProvisionResponseStatus)
	*p = x
	return p
}

func (x ProvisionResponse_ProvisionResponseStatus) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ProvisionResponse_ProvisionResponseStatus) Descriptor() protoreflect.EnumDescriptor {
	return file_xyz_block_ftl_provisioner_v1beta1_plugin_proto_enumTypes[0].Descriptor()
}

func (ProvisionResponse_ProvisionResponseStatus) Type() protoreflect.EnumType {
	return &file_xyz_block_ftl_provisioner_v1beta1_plugin_proto_enumTypes[0]
}

func (x ProvisionResponse_ProvisionResponseStatus) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ProvisionResponse_ProvisionResponseStatus.Descriptor instead.
func (ProvisionResponse_ProvisionResponseStatus) EnumDescriptor() ([]byte, []int) {
	return file_xyz_block_ftl_provisioner_v1beta1_plugin_proto_rawDescGZIP(), []int{2, 0}
}

// ResourceContext is the context used to create a new resource
// This includes the direct dependencies of the new resource, that can impact
// the resource creation.
type ResourceContext struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Resource     *Resource   `protobuf:"bytes,1,opt,name=resource,proto3" json:"resource,omitempty"`
	Dependencies []*Resource `protobuf:"bytes,2,rep,name=dependencies,proto3" json:"dependencies,omitempty"`
}

func (x *ResourceContext) Reset() {
	*x = ResourceContext{}
	mi := &file_xyz_block_ftl_provisioner_v1beta1_plugin_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ResourceContext) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ResourceContext) ProtoMessage() {}

func (x *ResourceContext) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_provisioner_v1beta1_plugin_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ResourceContext.ProtoReflect.Descriptor instead.
func (*ResourceContext) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_provisioner_v1beta1_plugin_proto_rawDescGZIP(), []int{0}
}

func (x *ResourceContext) GetResource() *Resource {
	if x != nil {
		return x.Resource
	}
	return nil
}

func (x *ResourceContext) GetDependencies() []*Resource {
	if x != nil {
		return x.Dependencies
	}
	return nil
}

type ProvisionRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	FtlClusterId string `protobuf:"bytes,1,opt,name=ftl_cluster_id,json=ftlClusterId,proto3" json:"ftl_cluster_id,omitempty"`
	Module       string `protobuf:"bytes,2,opt,name=module,proto3" json:"module,omitempty"`
	// The resource FTL thinks exists currently
	ExistingResources []*Resource `protobuf:"bytes,3,rep,name=existing_resources,json=existingResources,proto3" json:"existing_resources,omitempty"`
	// The resource FTL would like to exist after this provisioning run.
	// This includes all new, existing, and changes resources in this change.
	DesiredResources []*ResourceContext `protobuf:"bytes,4,rep,name=desired_resources,json=desiredResources,proto3" json:"desired_resources,omitempty"`
}

func (x *ProvisionRequest) Reset() {
	*x = ProvisionRequest{}
	mi := &file_xyz_block_ftl_provisioner_v1beta1_plugin_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ProvisionRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProvisionRequest) ProtoMessage() {}

func (x *ProvisionRequest) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_provisioner_v1beta1_plugin_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProvisionRequest.ProtoReflect.Descriptor instead.
func (*ProvisionRequest) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_provisioner_v1beta1_plugin_proto_rawDescGZIP(), []int{1}
}

func (x *ProvisionRequest) GetFtlClusterId() string {
	if x != nil {
		return x.FtlClusterId
	}
	return ""
}

func (x *ProvisionRequest) GetModule() string {
	if x != nil {
		return x.Module
	}
	return ""
}

func (x *ProvisionRequest) GetExistingResources() []*Resource {
	if x != nil {
		return x.ExistingResources
	}
	return nil
}

func (x *ProvisionRequest) GetDesiredResources() []*ResourceContext {
	if x != nil {
		return x.DesiredResources
	}
	return nil
}

type ProvisionResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ProvisioningToken string                                    `protobuf:"bytes,1,opt,name=provisioning_token,json=provisioningToken,proto3" json:"provisioning_token,omitempty"`
	Status            ProvisionResponse_ProvisionResponseStatus `protobuf:"varint,2,opt,name=status,proto3,enum=xyz.block.ftl.provisioner.v1beta1.ProvisionResponse_ProvisionResponseStatus" json:"status,omitempty"`
}

func (x *ProvisionResponse) Reset() {
	*x = ProvisionResponse{}
	mi := &file_xyz_block_ftl_provisioner_v1beta1_plugin_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ProvisionResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProvisionResponse) ProtoMessage() {}

func (x *ProvisionResponse) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_provisioner_v1beta1_plugin_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProvisionResponse.ProtoReflect.Descriptor instead.
func (*ProvisionResponse) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_provisioner_v1beta1_plugin_proto_rawDescGZIP(), []int{2}
}

func (x *ProvisionResponse) GetProvisioningToken() string {
	if x != nil {
		return x.ProvisioningToken
	}
	return ""
}

func (x *ProvisionResponse) GetStatus() ProvisionResponse_ProvisionResponseStatus {
	if x != nil {
		return x.Status
	}
	return ProvisionResponse_PROVISION_RESPONSE_STATUS_UNSPECIFIED
}

type StatusRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ProvisioningToken string `protobuf:"bytes,1,opt,name=provisioning_token,json=provisioningToken,proto3" json:"provisioning_token,omitempty"`
	// The set of desired_resources used to initiate this provisioning request
	// We need this as input here, so we can populate any resource fields in them
	// when the provisioning finishes
	DesiredResources []*Resource `protobuf:"bytes,2,rep,name=desired_resources,json=desiredResources,proto3" json:"desired_resources,omitempty"`
}

func (x *StatusRequest) Reset() {
	*x = StatusRequest{}
	mi := &file_xyz_block_ftl_provisioner_v1beta1_plugin_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *StatusRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StatusRequest) ProtoMessage() {}

func (x *StatusRequest) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_provisioner_v1beta1_plugin_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StatusRequest.ProtoReflect.Descriptor instead.
func (*StatusRequest) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_provisioner_v1beta1_plugin_proto_rawDescGZIP(), []int{3}
}

func (x *StatusRequest) GetProvisioningToken() string {
	if x != nil {
		return x.ProvisioningToken
	}
	return ""
}

func (x *StatusRequest) GetDesiredResources() []*Resource {
	if x != nil {
		return x.DesiredResources
	}
	return nil
}

type StatusResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Status:
	//
	//	*StatusResponse_Running
	//	*StatusResponse_Success
	Status isStatusResponse_Status `protobuf_oneof:"status"`
}

func (x *StatusResponse) Reset() {
	*x = StatusResponse{}
	mi := &file_xyz_block_ftl_provisioner_v1beta1_plugin_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *StatusResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StatusResponse) ProtoMessage() {}

func (x *StatusResponse) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_provisioner_v1beta1_plugin_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StatusResponse.ProtoReflect.Descriptor instead.
func (*StatusResponse) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_provisioner_v1beta1_plugin_proto_rawDescGZIP(), []int{4}
}

func (m *StatusResponse) GetStatus() isStatusResponse_Status {
	if m != nil {
		return m.Status
	}
	return nil
}

func (x *StatusResponse) GetRunning() *StatusResponse_ProvisioningRunning {
	if x, ok := x.GetStatus().(*StatusResponse_Running); ok {
		return x.Running
	}
	return nil
}

func (x *StatusResponse) GetSuccess() *StatusResponse_ProvisioningSuccess {
	if x, ok := x.GetStatus().(*StatusResponse_Success); ok {
		return x.Success
	}
	return nil
}

type isStatusResponse_Status interface {
	isStatusResponse_Status()
}

type StatusResponse_Running struct {
	Running *StatusResponse_ProvisioningRunning `protobuf:"bytes,1,opt,name=running,proto3,oneof"`
}

type StatusResponse_Success struct {
	Success *StatusResponse_ProvisioningSuccess `protobuf:"bytes,2,opt,name=success,proto3,oneof"`
}

func (*StatusResponse_Running) isStatusResponse_Status() {}

func (*StatusResponse_Success) isStatusResponse_Status() {}

type PlanRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Provisioning *ProvisionRequest `protobuf:"bytes,1,opt,name=provisioning,proto3" json:"provisioning,omitempty"`
}

func (x *PlanRequest) Reset() {
	*x = PlanRequest{}
	mi := &file_xyz_block_ftl_provisioner_v1beta1_plugin_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PlanRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PlanRequest) ProtoMessage() {}

func (x *PlanRequest) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_provisioner_v1beta1_plugin_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PlanRequest.ProtoReflect.Descriptor instead.
func (*PlanRequest) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_provisioner_v1beta1_plugin_proto_rawDescGZIP(), []int{5}
}

func (x *PlanRequest) GetProvisioning() *ProvisionRequest {
	if x != nil {
		return x.Provisioning
	}
	return nil
}

type PlanResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// a detailed, implementation specific, plan of changes this deployment would do
	Plan string `protobuf:"bytes,1,opt,name=plan,proto3" json:"plan,omitempty"`
}

func (x *PlanResponse) Reset() {
	*x = PlanResponse{}
	mi := &file_xyz_block_ftl_provisioner_v1beta1_plugin_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PlanResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PlanResponse) ProtoMessage() {}

func (x *PlanResponse) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_provisioner_v1beta1_plugin_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PlanResponse.ProtoReflect.Descriptor instead.
func (*PlanResponse) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_provisioner_v1beta1_plugin_proto_rawDescGZIP(), []int{6}
}

func (x *PlanResponse) GetPlan() string {
	if x != nil {
		return x.Plan
	}
	return ""
}

type StatusResponse_ProvisioningRunning struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *StatusResponse_ProvisioningRunning) Reset() {
	*x = StatusResponse_ProvisioningRunning{}
	mi := &file_xyz_block_ftl_provisioner_v1beta1_plugin_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *StatusResponse_ProvisioningRunning) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StatusResponse_ProvisioningRunning) ProtoMessage() {}

func (x *StatusResponse_ProvisioningRunning) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_provisioner_v1beta1_plugin_proto_msgTypes[7]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StatusResponse_ProvisioningRunning.ProtoReflect.Descriptor instead.
func (*StatusResponse_ProvisioningRunning) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_provisioner_v1beta1_plugin_proto_rawDescGZIP(), []int{4, 0}
}

type StatusResponse_ProvisioningFailed struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ErrorMessage string `protobuf:"bytes,1,opt,name=error_message,json=errorMessage,proto3" json:"error_message,omitempty"`
}

func (x *StatusResponse_ProvisioningFailed) Reset() {
	*x = StatusResponse_ProvisioningFailed{}
	mi := &file_xyz_block_ftl_provisioner_v1beta1_plugin_proto_msgTypes[8]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *StatusResponse_ProvisioningFailed) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StatusResponse_ProvisioningFailed) ProtoMessage() {}

func (x *StatusResponse_ProvisioningFailed) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_provisioner_v1beta1_plugin_proto_msgTypes[8]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StatusResponse_ProvisioningFailed.ProtoReflect.Descriptor instead.
func (*StatusResponse_ProvisioningFailed) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_provisioner_v1beta1_plugin_proto_rawDescGZIP(), []int{4, 1}
}

func (x *StatusResponse_ProvisioningFailed) GetErrorMessage() string {
	if x != nil {
		return x.ErrorMessage
	}
	return ""
}

type StatusResponse_ProvisioningSuccess struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Some fields in the resources might have been populated
	// during the provisioning. The new state is returned here
	UpdatedResources []*Resource `protobuf:"bytes,1,rep,name=updated_resources,json=updatedResources,proto3" json:"updated_resources,omitempty"`
}

func (x *StatusResponse_ProvisioningSuccess) Reset() {
	*x = StatusResponse_ProvisioningSuccess{}
	mi := &file_xyz_block_ftl_provisioner_v1beta1_plugin_proto_msgTypes[9]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *StatusResponse_ProvisioningSuccess) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StatusResponse_ProvisioningSuccess) ProtoMessage() {}

func (x *StatusResponse_ProvisioningSuccess) ProtoReflect() protoreflect.Message {
	mi := &file_xyz_block_ftl_provisioner_v1beta1_plugin_proto_msgTypes[9]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StatusResponse_ProvisioningSuccess.ProtoReflect.Descriptor instead.
func (*StatusResponse_ProvisioningSuccess) Descriptor() ([]byte, []int) {
	return file_xyz_block_ftl_provisioner_v1beta1_plugin_proto_rawDescGZIP(), []int{4, 2}
}

func (x *StatusResponse_ProvisioningSuccess) GetUpdatedResources() []*Resource {
	if x != nil {
		return x.UpdatedResources
	}
	return nil
}

var File_xyz_block_ftl_provisioner_v1beta1_plugin_proto protoreflect.FileDescriptor

var file_xyz_block_ftl_provisioner_v1beta1_plugin_proto_rawDesc = []byte{
	0x0a, 0x2e, 0x78, 0x79, 0x7a, 0x2f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2f, 0x66, 0x74, 0x6c, 0x2f,
	0x70, 0x72, 0x6f, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x65, 0x72, 0x2f, 0x76, 0x31, 0x62, 0x65,
	0x74, 0x61, 0x31, 0x2f, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x21, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e,
	0x70, 0x72, 0x6f, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x62, 0x65,
	0x74, 0x61, 0x31, 0x1a, 0x30, 0x78, 0x79, 0x7a, 0x2f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2f, 0x66,
	0x74, 0x6c, 0x2f, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x65, 0x72, 0x2f, 0x76,
	0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2f, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1a, 0x78, 0x79, 0x7a, 0x2f, 0x62, 0x6c, 0x6f, 0x63, 0x6b,
	0x2f, 0x66, 0x74, 0x6c, 0x2f, 0x76, 0x31, 0x2f, 0x66, 0x74, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x22, 0xab, 0x01, 0x0a, 0x0f, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x43, 0x6f,
	0x6e, 0x74, 0x65, 0x78, 0x74, 0x12, 0x47, 0x0a, 0x08, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63,
	0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2b, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c,
	0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x73, 0x69, 0x6f,
	0x6e, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x52, 0x65, 0x73, 0x6f,
	0x75, 0x72, 0x63, 0x65, 0x52, 0x08, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x12, 0x4f,
	0x0a, 0x0c, 0x64, 0x65, 0x70, 0x65, 0x6e, 0x64, 0x65, 0x6e, 0x63, 0x69, 0x65, 0x73, 0x18, 0x02,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x2b, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b,
	0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x65, 0x72,
	0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63,
	0x65, 0x52, 0x0c, 0x64, 0x65, 0x70, 0x65, 0x6e, 0x64, 0x65, 0x6e, 0x63, 0x69, 0x65, 0x73, 0x22,
	0x8d, 0x02, 0x0a, 0x10, 0x50, 0x72, 0x6f, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x12, 0x24, 0x0a, 0x0e, 0x66, 0x74, 0x6c, 0x5f, 0x63, 0x6c, 0x75, 0x73,
	0x74, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x66, 0x74,
	0x6c, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x49, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x6d, 0x6f,
	0x64, 0x75, 0x6c, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x6d, 0x6f, 0x64, 0x75,
	0x6c, 0x65, 0x12, 0x5a, 0x0a, 0x12, 0x65, 0x78, 0x69, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x5f, 0x72,
	0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x2b,
	0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x70,
	0x72, 0x6f, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x62, 0x65, 0x74,
	0x61, 0x31, 0x2e, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x52, 0x11, 0x65, 0x78, 0x69,
	0x73, 0x74, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73, 0x12, 0x5f,
	0x0a, 0x11, 0x64, 0x65, 0x73, 0x69, 0x72, 0x65, 0x64, 0x5f, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72,
	0x63, 0x65, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x32, 0x2e, 0x78, 0x79, 0x7a, 0x2e,
	0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x73,
	0x69, 0x6f, 0x6e, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x52, 0x65,
	0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x52, 0x10, 0x64,
	0x65, 0x73, 0x69, 0x72, 0x65, 0x64, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73, 0x22,
	0x97, 0x02, 0x0a, 0x11, 0x50, 0x72, 0x6f, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2d, 0x0a, 0x12, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x73, 0x69,
	0x6f, 0x6e, 0x69, 0x6e, 0x67, 0x5f, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x11, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x69, 0x6e, 0x67, 0x54,
	0x6f, 0x6b, 0x65, 0x6e, 0x12, 0x64, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0e, 0x32, 0x4c, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b,
	0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x65, 0x72,
	0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x50, 0x72, 0x6f, 0x76, 0x69, 0x73, 0x69,
	0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x2e, 0x50, 0x72, 0x6f, 0x76, 0x69,
	0x73, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x53, 0x74, 0x61, 0x74,
	0x75, 0x73, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x22, 0x6d, 0x0a, 0x17, 0x50, 0x72,
	0x6f, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x53,
	0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x29, 0x0a, 0x25, 0x50, 0x52, 0x4f, 0x56, 0x49, 0x53, 0x49,
	0x4f, 0x4e, 0x5f, 0x52, 0x45, 0x53, 0x50, 0x4f, 0x4e, 0x53, 0x45, 0x5f, 0x53, 0x54, 0x41, 0x54,
	0x55, 0x53, 0x5f, 0x55, 0x4e, 0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00,
	0x12, 0x27, 0x0a, 0x23, 0x50, 0x52, 0x4f, 0x56, 0x49, 0x53, 0x49, 0x4f, 0x4e, 0x5f, 0x52, 0x45,
	0x53, 0x50, 0x4f, 0x4e, 0x53, 0x45, 0x5f, 0x53, 0x54, 0x41, 0x54, 0x55, 0x53, 0x5f, 0x53, 0x55,
	0x42, 0x4d, 0x49, 0x54, 0x54, 0x45, 0x44, 0x10, 0x01, 0x22, 0x98, 0x01, 0x0a, 0x0d, 0x53, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x2d, 0x0a, 0x12, 0x70,
	0x72, 0x6f, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x69, 0x6e, 0x67, 0x5f, 0x74, 0x6f, 0x6b, 0x65,
	0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x11, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x73, 0x69,
	0x6f, 0x6e, 0x69, 0x6e, 0x67, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x12, 0x58, 0x0a, 0x11, 0x64, 0x65,
	0x73, 0x69, 0x72, 0x65, 0x64, 0x5f, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73, 0x18,
	0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x2b, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63,
	0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x65,
	0x72, 0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72,
	0x63, 0x65, 0x52, 0x10, 0x64, 0x65, 0x73, 0x69, 0x72, 0x65, 0x64, 0x52, 0x65, 0x73, 0x6f, 0x75,
	0x72, 0x63, 0x65, 0x73, 0x22, 0xa3, 0x03, 0x0a, 0x0e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x61, 0x0a, 0x07, 0x72, 0x75, 0x6e, 0x6e, 0x69,
	0x6e, 0x67, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x45, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62,
	0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x73, 0x69,
	0x6f, 0x6e, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x53, 0x74, 0x61,
	0x74, 0x75, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x2e, 0x50, 0x72, 0x6f, 0x76,
	0x69, 0x73, 0x69, 0x6f, 0x6e, 0x69, 0x6e, 0x67, 0x52, 0x75, 0x6e, 0x6e, 0x69, 0x6e, 0x67, 0x48,
	0x00, 0x52, 0x07, 0x72, 0x75, 0x6e, 0x6e, 0x69, 0x6e, 0x67, 0x12, 0x61, 0x0a, 0x07, 0x73, 0x75,
	0x63, 0x63, 0x65, 0x73, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x45, 0x2e, 0x78, 0x79,
	0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x76,
	0x69, 0x73, 0x69, 0x6f, 0x6e, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e,
	0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x2e, 0x50,
	0x72, 0x6f, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x69, 0x6e, 0x67, 0x53, 0x75, 0x63, 0x63, 0x65,
	0x73, 0x73, 0x48, 0x00, 0x52, 0x07, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x1a, 0x15, 0x0a,
	0x13, 0x50, 0x72, 0x6f, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x69, 0x6e, 0x67, 0x52, 0x75, 0x6e,
	0x6e, 0x69, 0x6e, 0x67, 0x1a, 0x39, 0x0a, 0x12, 0x50, 0x72, 0x6f, 0x76, 0x69, 0x73, 0x69, 0x6f,
	0x6e, 0x69, 0x6e, 0x67, 0x46, 0x61, 0x69, 0x6c, 0x65, 0x64, 0x12, 0x23, 0x0a, 0x0d, 0x65, 0x72,
	0x72, 0x6f, 0x72, 0x5f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x0c, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x1a,
	0x6f, 0x0a, 0x13, 0x50, 0x72, 0x6f, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x69, 0x6e, 0x67, 0x53,
	0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x12, 0x58, 0x0a, 0x11, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65,
	0x64, 0x5f, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x2b, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74,
	0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x65, 0x72, 0x2e, 0x76, 0x31,
	0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x52, 0x10,
	0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73,
	0x42, 0x08, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x22, 0x66, 0x0a, 0x0b, 0x50, 0x6c,
	0x61, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x57, 0x0a, 0x0c, 0x70, 0x72, 0x6f,
	0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x69, 0x6e, 0x67, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x33, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e,
	0x70, 0x72, 0x6f, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x62, 0x65,
	0x74, 0x61, 0x31, 0x2e, 0x50, 0x72, 0x6f, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x52, 0x0c, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x69,
	0x6e, 0x67, 0x22, 0x22, 0x0a, 0x0c, 0x50, 0x6c, 0x61, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x6c, 0x61, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x04, 0x70, 0x6c, 0x61, 0x6e, 0x32, 0xb1, 0x03, 0x0a, 0x18, 0x50, 0x72, 0x6f, 0x76, 0x69,
	0x73, 0x69, 0x6f, 0x6e, 0x65, 0x72, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x53, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x12, 0x45, 0x0a, 0x04, 0x50, 0x69, 0x6e, 0x67, 0x12, 0x1d, 0x2e, 0x78, 0x79,
	0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x50,
	0x69, 0x6e, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1e, 0x2e, 0x78, 0x79, 0x7a,
	0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x69,
	0x6e, 0x67, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x76, 0x0a, 0x09, 0x50, 0x72,
	0x6f, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x33, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c,
	0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x73, 0x69, 0x6f,
	0x6e, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x50, 0x72, 0x6f, 0x76,
	0x69, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x34, 0x2e, 0x78,
	0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x70, 0x72, 0x6f,
	0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31,
	0x2e, 0x50, 0x72, 0x6f, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x67, 0x0a, 0x04, 0x50, 0x6c, 0x61, 0x6e, 0x12, 0x2e, 0x2e, 0x78, 0x79, 0x7a,
	0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x76, 0x69,
	0x73, 0x69, 0x6f, 0x6e, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x50,
	0x6c, 0x61, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x2f, 0x2e, 0x78, 0x79, 0x7a,
	0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x76, 0x69,
	0x73, 0x69, 0x6f, 0x6e, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x50,
	0x6c, 0x61, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x6d, 0x0a, 0x06, 0x53,
	0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x30, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63,
	0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x65,
	0x72, 0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x31, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c,
	0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x73, 0x69, 0x6f,
	0x6e, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x53, 0x74, 0x61, 0x74,
	0x75, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x5d, 0x50, 0x01, 0x5a, 0x59,
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x54, 0x42, 0x44, 0x35, 0x34,
	0x35, 0x36, 0x36, 0x39, 0x37, 0x35, 0x2f, 0x66, 0x74, 0x6c, 0x2f, 0x62, 0x61, 0x63, 0x6b, 0x65,
	0x6e, 0x64, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2f, 0x78, 0x79, 0x7a, 0x2f, 0x62, 0x6c,
	0x6f, 0x63, 0x6b, 0x2f, 0x66, 0x74, 0x6c, 0x2f, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x73, 0x69, 0x6f,
	0x6e, 0x65, 0x72, 0x2f, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x3b, 0x70, 0x72, 0x6f, 0x76,
	0x69, 0x73, 0x69, 0x6f, 0x6e, 0x65, 0x72, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_xyz_block_ftl_provisioner_v1beta1_plugin_proto_rawDescOnce sync.Once
	file_xyz_block_ftl_provisioner_v1beta1_plugin_proto_rawDescData = file_xyz_block_ftl_provisioner_v1beta1_plugin_proto_rawDesc
)

func file_xyz_block_ftl_provisioner_v1beta1_plugin_proto_rawDescGZIP() []byte {
	file_xyz_block_ftl_provisioner_v1beta1_plugin_proto_rawDescOnce.Do(func() {
		file_xyz_block_ftl_provisioner_v1beta1_plugin_proto_rawDescData = protoimpl.X.CompressGZIP(file_xyz_block_ftl_provisioner_v1beta1_plugin_proto_rawDescData)
	})
	return file_xyz_block_ftl_provisioner_v1beta1_plugin_proto_rawDescData
}

var file_xyz_block_ftl_provisioner_v1beta1_plugin_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_xyz_block_ftl_provisioner_v1beta1_plugin_proto_msgTypes = make([]protoimpl.MessageInfo, 10)
var file_xyz_block_ftl_provisioner_v1beta1_plugin_proto_goTypes = []any{
	(ProvisionResponse_ProvisionResponseStatus)(0), // 0: xyz.block.ftl.provisioner.v1beta1.ProvisionResponse.ProvisionResponseStatus
	(*ResourceContext)(nil),                        // 1: xyz.block.ftl.provisioner.v1beta1.ResourceContext
	(*ProvisionRequest)(nil),                       // 2: xyz.block.ftl.provisioner.v1beta1.ProvisionRequest
	(*ProvisionResponse)(nil),                      // 3: xyz.block.ftl.provisioner.v1beta1.ProvisionResponse
	(*StatusRequest)(nil),                          // 4: xyz.block.ftl.provisioner.v1beta1.StatusRequest
	(*StatusResponse)(nil),                         // 5: xyz.block.ftl.provisioner.v1beta1.StatusResponse
	(*PlanRequest)(nil),                            // 6: xyz.block.ftl.provisioner.v1beta1.PlanRequest
	(*PlanResponse)(nil),                           // 7: xyz.block.ftl.provisioner.v1beta1.PlanResponse
	(*StatusResponse_ProvisioningRunning)(nil),     // 8: xyz.block.ftl.provisioner.v1beta1.StatusResponse.ProvisioningRunning
	(*StatusResponse_ProvisioningFailed)(nil),      // 9: xyz.block.ftl.provisioner.v1beta1.StatusResponse.ProvisioningFailed
	(*StatusResponse_ProvisioningSuccess)(nil),     // 10: xyz.block.ftl.provisioner.v1beta1.StatusResponse.ProvisioningSuccess
	(*Resource)(nil),                               // 11: xyz.block.ftl.provisioner.v1beta1.Resource
	(*v1.PingRequest)(nil),                         // 12: xyz.block.ftl.v1.PingRequest
	(*v1.PingResponse)(nil),                        // 13: xyz.block.ftl.v1.PingResponse
}
var file_xyz_block_ftl_provisioner_v1beta1_plugin_proto_depIdxs = []int32{
	11, // 0: xyz.block.ftl.provisioner.v1beta1.ResourceContext.resource:type_name -> xyz.block.ftl.provisioner.v1beta1.Resource
	11, // 1: xyz.block.ftl.provisioner.v1beta1.ResourceContext.dependencies:type_name -> xyz.block.ftl.provisioner.v1beta1.Resource
	11, // 2: xyz.block.ftl.provisioner.v1beta1.ProvisionRequest.existing_resources:type_name -> xyz.block.ftl.provisioner.v1beta1.Resource
	1,  // 3: xyz.block.ftl.provisioner.v1beta1.ProvisionRequest.desired_resources:type_name -> xyz.block.ftl.provisioner.v1beta1.ResourceContext
	0,  // 4: xyz.block.ftl.provisioner.v1beta1.ProvisionResponse.status:type_name -> xyz.block.ftl.provisioner.v1beta1.ProvisionResponse.ProvisionResponseStatus
	11, // 5: xyz.block.ftl.provisioner.v1beta1.StatusRequest.desired_resources:type_name -> xyz.block.ftl.provisioner.v1beta1.Resource
	8,  // 6: xyz.block.ftl.provisioner.v1beta1.StatusResponse.running:type_name -> xyz.block.ftl.provisioner.v1beta1.StatusResponse.ProvisioningRunning
	10, // 7: xyz.block.ftl.provisioner.v1beta1.StatusResponse.success:type_name -> xyz.block.ftl.provisioner.v1beta1.StatusResponse.ProvisioningSuccess
	2,  // 8: xyz.block.ftl.provisioner.v1beta1.PlanRequest.provisioning:type_name -> xyz.block.ftl.provisioner.v1beta1.ProvisionRequest
	11, // 9: xyz.block.ftl.provisioner.v1beta1.StatusResponse.ProvisioningSuccess.updated_resources:type_name -> xyz.block.ftl.provisioner.v1beta1.Resource
	12, // 10: xyz.block.ftl.provisioner.v1beta1.ProvisionerPluginService.Ping:input_type -> xyz.block.ftl.v1.PingRequest
	2,  // 11: xyz.block.ftl.provisioner.v1beta1.ProvisionerPluginService.Provision:input_type -> xyz.block.ftl.provisioner.v1beta1.ProvisionRequest
	6,  // 12: xyz.block.ftl.provisioner.v1beta1.ProvisionerPluginService.Plan:input_type -> xyz.block.ftl.provisioner.v1beta1.PlanRequest
	4,  // 13: xyz.block.ftl.provisioner.v1beta1.ProvisionerPluginService.Status:input_type -> xyz.block.ftl.provisioner.v1beta1.StatusRequest
	13, // 14: xyz.block.ftl.provisioner.v1beta1.ProvisionerPluginService.Ping:output_type -> xyz.block.ftl.v1.PingResponse
	3,  // 15: xyz.block.ftl.provisioner.v1beta1.ProvisionerPluginService.Provision:output_type -> xyz.block.ftl.provisioner.v1beta1.ProvisionResponse
	7,  // 16: xyz.block.ftl.provisioner.v1beta1.ProvisionerPluginService.Plan:output_type -> xyz.block.ftl.provisioner.v1beta1.PlanResponse
	5,  // 17: xyz.block.ftl.provisioner.v1beta1.ProvisionerPluginService.Status:output_type -> xyz.block.ftl.provisioner.v1beta1.StatusResponse
	14, // [14:18] is the sub-list for method output_type
	10, // [10:14] is the sub-list for method input_type
	10, // [10:10] is the sub-list for extension type_name
	10, // [10:10] is the sub-list for extension extendee
	0,  // [0:10] is the sub-list for field type_name
}

func init() { file_xyz_block_ftl_provisioner_v1beta1_plugin_proto_init() }
func file_xyz_block_ftl_provisioner_v1beta1_plugin_proto_init() {
	if File_xyz_block_ftl_provisioner_v1beta1_plugin_proto != nil {
		return
	}
	file_xyz_block_ftl_provisioner_v1beta1_resource_proto_init()
	file_xyz_block_ftl_provisioner_v1beta1_plugin_proto_msgTypes[4].OneofWrappers = []any{
		(*StatusResponse_Running)(nil),
		(*StatusResponse_Success)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_xyz_block_ftl_provisioner_v1beta1_plugin_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   10,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_xyz_block_ftl_provisioner_v1beta1_plugin_proto_goTypes,
		DependencyIndexes: file_xyz_block_ftl_provisioner_v1beta1_plugin_proto_depIdxs,
		EnumInfos:         file_xyz_block_ftl_provisioner_v1beta1_plugin_proto_enumTypes,
		MessageInfos:      file_xyz_block_ftl_provisioner_v1beta1_plugin_proto_msgTypes,
	}.Build()
	File_xyz_block_ftl_provisioner_v1beta1_plugin_proto = out.File
	file_xyz_block_ftl_provisioner_v1beta1_plugin_proto_rawDesc = nil
	file_xyz_block_ftl_provisioner_v1beta1_plugin_proto_goTypes = nil
	file_xyz_block_ftl_provisioner_v1beta1_plugin_proto_depIdxs = nil
}