# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# NO CHECKED-IN PROTOBUF GENCODE
# source: xyz/block/ftl/v1/module.proto
# Protobuf Python Version: 5.28.3
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import runtime_version as _runtime_version
from google.protobuf import symbol_database as _symbol_database
from google.protobuf.internal import builder as _builder
_runtime_version.ValidateProtobufRuntimeVersion(
    _runtime_version.Domain.PUBLIC,
    5,
    28,
    3,
    '',
    'xyz/block/ftl/v1/module.proto'
)
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from google.protobuf import duration_pb2 as google_dot_protobuf_dot_duration__pb2
from xyz.block.ftl.schema.v1 import schema_pb2 as xyz_dot_block_dot_ftl_dot_schema_dot_v1_dot_schema__pb2
from xyz.block.ftl.v1 import ftl_pb2 as xyz_dot_block_dot_ftl_dot_v1_dot_ftl__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n\x1dxyz/block/ftl/v1/module.proto\x12\x10xyz.block.ftl.v1\x1a\x1egoogle/protobuf/duration.proto\x1a$xyz/block/ftl/schema/v1/schema.proto\x1a\x1axyz/block/ftl/v1/ftl.proto\"l\n\x13\x41\x63quireLeaseRequest\x12\x16\n\x06module\x18\x01 \x01(\tR\x06module\x12\x10\n\x03key\x18\x02 \x03(\tR\x03key\x12+\n\x03ttl\x18\x03 \x01(\x0b\x32\x19.google.protobuf.DurationR\x03ttl\"\x16\n\x14\x41\x63quireLeaseResponse\"u\n\x13PublishEventRequest\x12\x32\n\x05topic\x18\x01 \x01(\x0b\x32\x1c.xyz.block.ftl.schema.v1.RefR\x05topic\x12\x12\n\x04\x62ody\x18\x02 \x01(\x0cR\x04\x62ody\x12\x16\n\x06\x63\x61ller\x18\x03 \x01(\tR\x06\x63\x61ller\"\x16\n\x14PublishEventResponse\"1\n\x17GetModuleContextRequest\x12\x16\n\x06module\x18\x01 \x01(\tR\x06module\"\x9e\x06\n\x18GetModuleContextResponse\x12\x16\n\x06module\x18\x01 \x01(\tR\x06module\x12Q\n\x07\x63onfigs\x18\x02 \x03(\x0b\x32\x37.xyz.block.ftl.v1.GetModuleContextResponse.ConfigsEntryR\x07\x63onfigs\x12Q\n\x07secrets\x18\x03 \x03(\x0b\x32\x37.xyz.block.ftl.v1.GetModuleContextResponse.SecretsEntryR\x07secrets\x12L\n\tdatabases\x18\x04 \x03(\x0b\x32..xyz.block.ftl.v1.GetModuleContextResponse.DSNR\tdatabases\x12H\n\x06routes\x18\x05 \x03(\x0b\x32\x30.xyz.block.ftl.v1.GetModuleContextResponse.RouteR\x06routes\x1a\x41\n\x03Ref\x12\x1b\n\x06module\x18\x01 \x01(\tH\x00R\x06module\x88\x01\x01\x12\x12\n\x04name\x18\x02 \x01(\tR\x04nameB\t\n\x07_module\x1ar\n\x03\x44SN\x12\x12\n\x04name\x18\x01 \x01(\tR\x04name\x12\x45\n\x04type\x18\x02 \x01(\x0e\x32\x31.xyz.block.ftl.v1.GetModuleContextResponse.DbTypeR\x04type\x12\x10\n\x03\x64sn\x18\x03 \x01(\tR\x03\x64sn\x1a\x31\n\x05Route\x12\x16\n\x06module\x18\x01 \x01(\tR\x06module\x12\x10\n\x03uri\x18\x02 \x01(\tR\x03uri\x1a:\n\x0c\x43onfigsEntry\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n\x05value\x18\x02 \x01(\x0cR\x05value:\x02\x38\x01\x1a:\n\x0cSecretsEntry\x12\x10\n\x03key\x18\x01 \x01(\tR\x03key\x12\x14\n\x05value\x18\x02 \x01(\x0cR\x05value:\x02\x38\x01\"J\n\x06\x44\x62Type\x12\x17\n\x13\x44\x42_TYPE_UNSPECIFIED\x10\x00\x12\x14\n\x10\x44\x42_TYPE_POSTGRES\x10\x01\x12\x11\n\rDB_TYPE_MYSQL\x10\x02\x32\x8a\x03\n\rModuleService\x12J\n\x04Ping\x12\x1d.xyz.block.ftl.v1.PingRequest\x1a\x1e.xyz.block.ftl.v1.PingResponse\"\x03\x90\x02\x01\x12k\n\x10GetModuleContext\x12).xyz.block.ftl.v1.GetModuleContextRequest\x1a*.xyz.block.ftl.v1.GetModuleContextResponse0\x01\x12\x61\n\x0c\x41\x63quireLease\x12%.xyz.block.ftl.v1.AcquireLeaseRequest\x1a&.xyz.block.ftl.v1.AcquireLeaseResponse(\x01\x30\x01\x12]\n\x0cPublishEvent\x12%.xyz.block.ftl.v1.PublishEventRequest\x1a&.xyz.block.ftl.v1.PublishEventResponseBDP\x01Z@github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1;ftlv1b\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'xyz.block.ftl.v1.module_pb2', _globals)
if not _descriptor._USE_C_DESCRIPTORS:
  _globals['DESCRIPTOR']._loaded_options = None
  _globals['DESCRIPTOR']._serialized_options = b'P\001Z@github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1;ftlv1'
  _globals['_GETMODULECONTEXTRESPONSE_CONFIGSENTRY']._loaded_options = None
  _globals['_GETMODULECONTEXTRESPONSE_CONFIGSENTRY']._serialized_options = b'8\001'
  _globals['_GETMODULECONTEXTRESPONSE_SECRETSENTRY']._loaded_options = None
  _globals['_GETMODULECONTEXTRESPONSE_SECRETSENTRY']._serialized_options = b'8\001'
  _globals['_MODULESERVICE'].methods_by_name['Ping']._loaded_options = None
  _globals['_MODULESERVICE'].methods_by_name['Ping']._serialized_options = b'\220\002\001'
  _globals['_ACQUIRELEASEREQUEST']._serialized_start=149
  _globals['_ACQUIRELEASEREQUEST']._serialized_end=257
  _globals['_ACQUIRELEASERESPONSE']._serialized_start=259
  _globals['_ACQUIRELEASERESPONSE']._serialized_end=281
  _globals['_PUBLISHEVENTREQUEST']._serialized_start=283
  _globals['_PUBLISHEVENTREQUEST']._serialized_end=400
  _globals['_PUBLISHEVENTRESPONSE']._serialized_start=402
  _globals['_PUBLISHEVENTRESPONSE']._serialized_end=424
  _globals['_GETMODULECONTEXTREQUEST']._serialized_start=426
  _globals['_GETMODULECONTEXTREQUEST']._serialized_end=475
  _globals['_GETMODULECONTEXTRESPONSE']._serialized_start=478
  _globals['_GETMODULECONTEXTRESPONSE']._serialized_end=1276
  _globals['_GETMODULECONTEXTRESPONSE_REF']._serialized_start=848
  _globals['_GETMODULECONTEXTRESPONSE_REF']._serialized_end=913
  _globals['_GETMODULECONTEXTRESPONSE_DSN']._serialized_start=915
  _globals['_GETMODULECONTEXTRESPONSE_DSN']._serialized_end=1029
  _globals['_GETMODULECONTEXTRESPONSE_ROUTE']._serialized_start=1031
  _globals['_GETMODULECONTEXTRESPONSE_ROUTE']._serialized_end=1080
  _globals['_GETMODULECONTEXTRESPONSE_CONFIGSENTRY']._serialized_start=1082
  _globals['_GETMODULECONTEXTRESPONSE_CONFIGSENTRY']._serialized_end=1140
  _globals['_GETMODULECONTEXTRESPONSE_SECRETSENTRY']._serialized_start=1142
  _globals['_GETMODULECONTEXTRESPONSE_SECRETSENTRY']._serialized_end=1200
  _globals['_GETMODULECONTEXTRESPONSE_DBTYPE']._serialized_start=1202
  _globals['_GETMODULECONTEXTRESPONSE_DBTYPE']._serialized_end=1276
  _globals['_MODULESERVICE']._serialized_start=1279
  _globals['_MODULESERVICE']._serialized_end=1673
# @@protoc_insertion_point(module_scope)
