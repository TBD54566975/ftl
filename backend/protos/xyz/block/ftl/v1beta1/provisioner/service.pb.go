// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.1
// 	protoc        (unknown)
// source: xyz/block/ftl/v1beta1/provisioner/service.proto

package provisioner

import (
	v1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

var File_xyz_block_ftl_v1beta1_provisioner_service_proto protoreflect.FileDescriptor

var file_xyz_block_ftl_v1beta1_provisioner_service_proto_rawDesc = []byte{
	0x0a, 0x2f, 0x78, 0x79, 0x7a, 0x2f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2f, 0x66, 0x74, 0x6c, 0x2f,
	0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2f, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x73, 0x69, 0x6f,
	0x6e, 0x65, 0x72, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x21, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c,
	0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x73, 0x69,
	0x6f, 0x6e, 0x65, 0x72, 0x1a, 0x1a, 0x78, 0x79, 0x7a, 0x2f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2f,
	0x66, 0x74, 0x6c, 0x2f, 0x76, 0x31, 0x2f, 0x66, 0x74, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x32, 0xda, 0x06, 0x0a, 0x12, 0x50, 0x72, 0x6f, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x65, 0x72,
	0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x4a, 0x0a, 0x04, 0x50, 0x69, 0x6e, 0x67, 0x12,
	0x1d, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e,
	0x76, 0x31, 0x2e, 0x50, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1e,
	0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76,
	0x31, 0x2e, 0x50, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x03,
	0x90, 0x02, 0x01, 0x12, 0x4b, 0x0a, 0x06, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x1f, 0x2e,
	0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76, 0x31,
	0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x20,
	0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76,
	0x31, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x69, 0x0a, 0x10, 0x47, 0x65, 0x74, 0x41, 0x72, 0x74, 0x65, 0x66, 0x61, 0x63, 0x74, 0x44,
	0x69, 0x66, 0x66, 0x73, 0x12, 0x29, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b,
	0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x41, 0x72, 0x74, 0x65, 0x66,
	0x61, 0x63, 0x74, 0x44, 0x69, 0x66, 0x66, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x2a, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e,
	0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x41, 0x72, 0x74, 0x65, 0x66, 0x61, 0x63, 0x74, 0x44, 0x69,
	0x66, 0x66, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x63, 0x0a, 0x0e, 0x55,
	0x70, 0x6c, 0x6f, 0x61, 0x64, 0x41, 0x72, 0x74, 0x65, 0x66, 0x61, 0x63, 0x74, 0x12, 0x27, 0x2e,
	0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76, 0x31,
	0x2e, 0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x41, 0x72, 0x74, 0x65, 0x66, 0x61, 0x63, 0x74, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x28, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f,
	0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64,
	0x41, 0x72, 0x74, 0x65, 0x66, 0x61, 0x63, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x69, 0x0a, 0x10, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x44, 0x65, 0x70, 0x6c, 0x6f, 0x79,
	0x6d, 0x65, 0x6e, 0x74, 0x12, 0x29, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b,
	0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x44, 0x65,
	0x70, 0x6c, 0x6f, 0x79, 0x6d, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x2a, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e,
	0x76, 0x31, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x44, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x6d,
	0x65, 0x6e, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x5d, 0x0a, 0x0c, 0x55,
	0x70, 0x64, 0x61, 0x74, 0x65, 0x44, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x12, 0x25, 0x2e, 0x78, 0x79,
	0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x55,
	0x70, 0x64, 0x61, 0x74, 0x65, 0x44, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x26, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66,
	0x74, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x44, 0x65, 0x70, 0x6c,
	0x6f, 0x79, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x60, 0x0a, 0x0d, 0x52, 0x65,
	0x70, 0x6c, 0x61, 0x63, 0x65, 0x44, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x12, 0x26, 0x2e, 0x78, 0x79,
	0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x52,
	0x65, 0x70, 0x6c, 0x61, 0x63, 0x65, 0x44, 0x65, 0x70, 0x6c, 0x6f, 0x79, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x27, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e,
	0x66, 0x74, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x70, 0x6c, 0x61, 0x63, 0x65, 0x44, 0x65,
	0x70, 0x6c, 0x6f, 0x79, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x54, 0x0a, 0x09,
	0x47, 0x65, 0x74, 0x53, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x12, 0x22, 0x2e, 0x78, 0x79, 0x7a, 0x2e,
	0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74,
	0x53, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x23, 0x2e,
	0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76, 0x31,
	0x2e, 0x47, 0x65, 0x74, 0x53, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x59, 0x0a, 0x0a, 0x50, 0x75, 0x6c, 0x6c, 0x53, 0x63, 0x68, 0x65, 0x6d, 0x61,
	0x12, 0x23, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2e, 0x66, 0x74, 0x6c,
	0x2e, 0x76, 0x31, 0x2e, 0x50, 0x75, 0x6c, 0x6c, 0x53, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x24, 0x2e, 0x78, 0x79, 0x7a, 0x2e, 0x62, 0x6c, 0x6f, 0x63,
	0x6b, 0x2e, 0x66, 0x74, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x75, 0x6c, 0x6c, 0x53, 0x63, 0x68,
	0x65, 0x6d, 0x61, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x30, 0x01, 0x42, 0x5b, 0x50,
	0x01, 0x5a, 0x57, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x54, 0x42,
	0x44, 0x35, 0x34, 0x35, 0x36, 0x36, 0x39, 0x37, 0x35, 0x2f, 0x66, 0x74, 0x6c, 0x2f, 0x62, 0x61,
	0x63, 0x6b, 0x65, 0x6e, 0x64, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2f, 0x78, 0x79, 0x7a,
	0x2f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x2f, 0x66, 0x74, 0x6c, 0x2f, 0x76, 0x31, 0x62, 0x65, 0x74,
	0x61, 0x31, 0x2f, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x65, 0x72, 0x3b, 0x70,
	0x72, 0x6f, 0x76, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x65, 0x72, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var file_xyz_block_ftl_v1beta1_provisioner_service_proto_goTypes = []any{
	(*v1.PingRequest)(nil),              // 0: xyz.block.ftl.v1.PingRequest
	(*v1.StatusRequest)(nil),            // 1: xyz.block.ftl.v1.StatusRequest
	(*v1.GetArtefactDiffsRequest)(nil),  // 2: xyz.block.ftl.v1.GetArtefactDiffsRequest
	(*v1.UploadArtefactRequest)(nil),    // 3: xyz.block.ftl.v1.UploadArtefactRequest
	(*v1.CreateDeploymentRequest)(nil),  // 4: xyz.block.ftl.v1.CreateDeploymentRequest
	(*v1.UpdateDeployRequest)(nil),      // 5: xyz.block.ftl.v1.UpdateDeployRequest
	(*v1.ReplaceDeployRequest)(nil),     // 6: xyz.block.ftl.v1.ReplaceDeployRequest
	(*v1.GetSchemaRequest)(nil),         // 7: xyz.block.ftl.v1.GetSchemaRequest
	(*v1.PullSchemaRequest)(nil),        // 8: xyz.block.ftl.v1.PullSchemaRequest
	(*v1.PingResponse)(nil),             // 9: xyz.block.ftl.v1.PingResponse
	(*v1.StatusResponse)(nil),           // 10: xyz.block.ftl.v1.StatusResponse
	(*v1.GetArtefactDiffsResponse)(nil), // 11: xyz.block.ftl.v1.GetArtefactDiffsResponse
	(*v1.UploadArtefactResponse)(nil),   // 12: xyz.block.ftl.v1.UploadArtefactResponse
	(*v1.CreateDeploymentResponse)(nil), // 13: xyz.block.ftl.v1.CreateDeploymentResponse
	(*v1.UpdateDeployResponse)(nil),     // 14: xyz.block.ftl.v1.UpdateDeployResponse
	(*v1.ReplaceDeployResponse)(nil),    // 15: xyz.block.ftl.v1.ReplaceDeployResponse
	(*v1.GetSchemaResponse)(nil),        // 16: xyz.block.ftl.v1.GetSchemaResponse
	(*v1.PullSchemaResponse)(nil),       // 17: xyz.block.ftl.v1.PullSchemaResponse
}
var file_xyz_block_ftl_v1beta1_provisioner_service_proto_depIdxs = []int32{
	0,  // 0: xyz.block.ftl.v1beta1.provisioner.ProvisionerService.Ping:input_type -> xyz.block.ftl.v1.PingRequest
	1,  // 1: xyz.block.ftl.v1beta1.provisioner.ProvisionerService.Status:input_type -> xyz.block.ftl.v1.StatusRequest
	2,  // 2: xyz.block.ftl.v1beta1.provisioner.ProvisionerService.GetArtefactDiffs:input_type -> xyz.block.ftl.v1.GetArtefactDiffsRequest
	3,  // 3: xyz.block.ftl.v1beta1.provisioner.ProvisionerService.UploadArtefact:input_type -> xyz.block.ftl.v1.UploadArtefactRequest
	4,  // 4: xyz.block.ftl.v1beta1.provisioner.ProvisionerService.CreateDeployment:input_type -> xyz.block.ftl.v1.CreateDeploymentRequest
	5,  // 5: xyz.block.ftl.v1beta1.provisioner.ProvisionerService.UpdateDeploy:input_type -> xyz.block.ftl.v1.UpdateDeployRequest
	6,  // 6: xyz.block.ftl.v1beta1.provisioner.ProvisionerService.ReplaceDeploy:input_type -> xyz.block.ftl.v1.ReplaceDeployRequest
	7,  // 7: xyz.block.ftl.v1beta1.provisioner.ProvisionerService.GetSchema:input_type -> xyz.block.ftl.v1.GetSchemaRequest
	8,  // 8: xyz.block.ftl.v1beta1.provisioner.ProvisionerService.PullSchema:input_type -> xyz.block.ftl.v1.PullSchemaRequest
	9,  // 9: xyz.block.ftl.v1beta1.provisioner.ProvisionerService.Ping:output_type -> xyz.block.ftl.v1.PingResponse
	10, // 10: xyz.block.ftl.v1beta1.provisioner.ProvisionerService.Status:output_type -> xyz.block.ftl.v1.StatusResponse
	11, // 11: xyz.block.ftl.v1beta1.provisioner.ProvisionerService.GetArtefactDiffs:output_type -> xyz.block.ftl.v1.GetArtefactDiffsResponse
	12, // 12: xyz.block.ftl.v1beta1.provisioner.ProvisionerService.UploadArtefact:output_type -> xyz.block.ftl.v1.UploadArtefactResponse
	13, // 13: xyz.block.ftl.v1beta1.provisioner.ProvisionerService.CreateDeployment:output_type -> xyz.block.ftl.v1.CreateDeploymentResponse
	14, // 14: xyz.block.ftl.v1beta1.provisioner.ProvisionerService.UpdateDeploy:output_type -> xyz.block.ftl.v1.UpdateDeployResponse
	15, // 15: xyz.block.ftl.v1beta1.provisioner.ProvisionerService.ReplaceDeploy:output_type -> xyz.block.ftl.v1.ReplaceDeployResponse
	16, // 16: xyz.block.ftl.v1beta1.provisioner.ProvisionerService.GetSchema:output_type -> xyz.block.ftl.v1.GetSchemaResponse
	17, // 17: xyz.block.ftl.v1beta1.provisioner.ProvisionerService.PullSchema:output_type -> xyz.block.ftl.v1.PullSchemaResponse
	9,  // [9:18] is the sub-list for method output_type
	0,  // [0:9] is the sub-list for method input_type
	0,  // [0:0] is the sub-list for extension type_name
	0,  // [0:0] is the sub-list for extension extendee
	0,  // [0:0] is the sub-list for field type_name
}

func init() { file_xyz_block_ftl_v1beta1_provisioner_service_proto_init() }
func file_xyz_block_ftl_v1beta1_provisioner_service_proto_init() {
	if File_xyz_block_ftl_v1beta1_provisioner_service_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_xyz_block_ftl_v1beta1_provisioner_service_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_xyz_block_ftl_v1beta1_provisioner_service_proto_goTypes,
		DependencyIndexes: file_xyz_block_ftl_v1beta1_provisioner_service_proto_depIdxs,
	}.Build()
	File_xyz_block_ftl_v1beta1_provisioner_service_proto = out.File
	file_xyz_block_ftl_v1beta1_provisioner_service_proto_rawDesc = nil
	file_xyz_block_ftl_v1beta1_provisioner_service_proto_goTypes = nil
	file_xyz_block_ftl_v1beta1_provisioner_service_proto_depIdxs = nil
}
