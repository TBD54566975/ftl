# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# NO CHECKED-IN PROTOBUF GENCODE
# source: xyz/block/ftl/v1/verb.proto
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
    'xyz/block/ftl/v1/verb.proto'
)
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from xyz.block.ftl.schema.v1 import schema_pb2 as xyz_dot_block_dot_ftl_dot_schema_dot_v1_dot_schema__pb2
from xyz.block.ftl.v1 import ftl_pb2 as xyz_dot_block_dot_ftl_dot_v1_dot_ftl__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n\x1bxyz/block/ftl/v1/verb.proto\x12\x10xyz.block.ftl.v1\x1a$xyz/block/ftl/schema/v1/schema.proto\x1a\x1axyz/block/ftl/v1/ftl.proto\"\x8b\x01\n\x0b\x43\x61llRequest\x12\x36\n\x08metadata\x18\x01 \x01(\x0b\x32\x1a.xyz.block.ftl.v1.MetadataR\x08metadata\x12\x30\n\x04verb\x18\x02 \x01(\x0b\x32\x1c.xyz.block.ftl.schema.v1.RefR\x04verb\x12\x12\n\x04\x62ody\x18\x03 \x01(\x0cR\x04\x62ody\"\xb6\x01\n\x0c\x43\x61llResponse\x12\x14\n\x04\x62ody\x18\x01 \x01(\x0cH\x00R\x04\x62ody\x12<\n\x05\x65rror\x18\x02 \x01(\x0b\x32$.xyz.block.ftl.v1.CallResponse.ErrorH\x00R\x05\x65rror\x1a\x46\n\x05\x45rror\x12\x18\n\x07message\x18\x01 \x01(\tR\x07message\x12\x19\n\x05stack\x18\x02 \x01(\tH\x00R\x05stack\x88\x01\x01\x42\x08\n\x06_stackB\n\n\x08response2\xa0\x01\n\x0bVerbService\x12J\n\x04Ping\x12\x1d.xyz.block.ftl.v1.PingRequest\x1a\x1e.xyz.block.ftl.v1.PingResponse\"\x03\x90\x02\x01\x12\x45\n\x04\x43\x61ll\x12\x1d.xyz.block.ftl.v1.CallRequest\x1a\x1e.xyz.block.ftl.v1.CallResponseBDP\x01Z@github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1;ftlv1b\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'xyz.block.ftl.v1.verb_pb2', _globals)
if not _descriptor._USE_C_DESCRIPTORS:
  _globals['DESCRIPTOR']._loaded_options = None
  _globals['DESCRIPTOR']._serialized_options = b'P\001Z@github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1;ftlv1'
  _globals['_VERBSERVICE'].methods_by_name['Ping']._loaded_options = None
  _globals['_VERBSERVICE'].methods_by_name['Ping']._serialized_options = b'\220\002\001'
  _globals['_CALLREQUEST']._serialized_start=116
  _globals['_CALLREQUEST']._serialized_end=255
  _globals['_CALLRESPONSE']._serialized_start=258
  _globals['_CALLRESPONSE']._serialized_end=440
  _globals['_CALLRESPONSE_ERROR']._serialized_start=358
  _globals['_CALLRESPONSE_ERROR']._serialized_end=428
  _globals['_VERBSERVICE']._serialized_start=443
  _globals['_VERBSERVICE']._serialized_end=603
# @@protoc_insertion_point(module_scope)
