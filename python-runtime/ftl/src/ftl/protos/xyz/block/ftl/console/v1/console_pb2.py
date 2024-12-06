# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# NO CHECKED-IN PROTOBUF GENCODE
# source: xyz/block/ftl/console/v1/console.proto
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
    'xyz/block/ftl/console/v1/console.proto'
)
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from xyz.block.ftl.schema.v1 import schema_pb2 as xyz_dot_block_dot_ftl_dot_schema_dot_v1_dot_schema__pb2
from xyz.block.ftl.v1 import ftl_pb2 as xyz_dot_block_dot_ftl_dot_v1_dot_ftl__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n&xyz/block/ftl/console/v1/console.proto\x12\x18xyz.block.ftl.console.v1\x1a$xyz/block/ftl/schema/v1/schema.proto\x1a\x1axyz/block/ftl/v1/ftl.proto\"\x7f\n\x06\x43onfig\x12\x37\n\x06\x63onfig\x18\x01 \x01(\x0b\x32\x1f.xyz.block.ftl.schema.v1.ConfigR\x06\x63onfig\x12<\n\nreferences\x18\x02 \x03(\x0b\x32\x1c.xyz.block.ftl.schema.v1.RefR\nreferences\"\x8f\x01\n\x04\x44\x61ta\x12\x31\n\x04\x64\x61ta\x18\x01 \x01(\x0b\x32\x1d.xyz.block.ftl.schema.v1.DataR\x04\x64\x61ta\x12\x16\n\x06schema\x18\x02 \x01(\tR\x06schema\x12<\n\nreferences\x18\x03 \x03(\x0b\x32\x1c.xyz.block.ftl.schema.v1.RefR\nreferences\"\x87\x01\n\x08\x44\x61tabase\x12=\n\x08\x64\x61tabase\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.schema.v1.DatabaseR\x08\x64\x61tabase\x12<\n\nreferences\x18\x02 \x03(\x0b\x32\x1c.xyz.block.ftl.schema.v1.RefR\nreferences\"w\n\x04\x45num\x12\x31\n\x04\x65num\x18\x01 \x01(\x0b\x32\x1d.xyz.block.ftl.schema.v1.EnumR\x04\x65num\x12<\n\nreferences\x18\x02 \x03(\x0b\x32\x1c.xyz.block.ftl.schema.v1.RefR\nreferences\"{\n\x05Topic\x12\x34\n\x05topic\x18\x01 \x01(\x0b\x32\x1e.xyz.block.ftl.schema.v1.TopicR\x05topic\x12<\n\nreferences\x18\x02 \x03(\x0b\x32\x1c.xyz.block.ftl.schema.v1.RefR\nreferences\"\x8b\x01\n\tTypeAlias\x12@\n\ttypealias\x18\x01 \x01(\x0b\x32\".xyz.block.ftl.schema.v1.TypeAliasR\ttypealias\x12<\n\nreferences\x18\x02 \x03(\x0b\x32\x1c.xyz.block.ftl.schema.v1.RefR\nreferences\"\x7f\n\x06Secret\x12\x37\n\x06secret\x18\x01 \x01(\x0b\x32\x1f.xyz.block.ftl.schema.v1.SecretR\x06secret\x12<\n\nreferences\x18\x02 \x03(\x0b\x32\x1c.xyz.block.ftl.schema.v1.RefR\nreferences\"\xbf\x01\n\x04Verb\x12\x31\n\x04verb\x18\x01 \x01(\x0b\x32\x1d.xyz.block.ftl.schema.v1.VerbR\x04verb\x12\x16\n\x06schema\x18\x02 \x01(\tR\x06schema\x12.\n\x13json_request_schema\x18\x03 \x01(\tR\x11jsonRequestSchema\x12<\n\nreferences\x18\x04 \x03(\x0b\x32\x1c.xyz.block.ftl.schema.v1.RefR\nreferences\"\xd1\x04\n\x06Module\x12\x12\n\x04name\x18\x01 \x01(\tR\x04name\x12%\n\x0e\x64\x65ployment_key\x18\x02 \x01(\tR\rdeploymentKey\x12\x1a\n\x08language\x18\x03 \x01(\tR\x08language\x12\x16\n\x06schema\x18\x04 \x01(\tR\x06schema\x12\x34\n\x05verbs\x18\x05 \x03(\x0b\x32\x1e.xyz.block.ftl.console.v1.VerbR\x05verbs\x12\x32\n\x04\x64\x61ta\x18\x06 \x03(\x0b\x32\x1e.xyz.block.ftl.console.v1.DataR\x04\x64\x61ta\x12:\n\x07secrets\x18\x07 \x03(\x0b\x32 .xyz.block.ftl.console.v1.SecretR\x07secrets\x12:\n\x07\x63onfigs\x18\x08 \x03(\x0b\x32 .xyz.block.ftl.console.v1.ConfigR\x07\x63onfigs\x12@\n\tdatabases\x18\t \x03(\x0b\x32\".xyz.block.ftl.console.v1.DatabaseR\tdatabases\x12\x34\n\x05\x65nums\x18\n \x03(\x0b\x32\x1e.xyz.block.ftl.console.v1.EnumR\x05\x65nums\x12\x37\n\x06topics\x18\x0b \x03(\x0b\x32\x1f.xyz.block.ftl.console.v1.TopicR\x06topics\x12\x45\n\x0btypealiases\x18\x0c \x03(\x0b\x32#.xyz.block.ftl.console.v1.TypeAliasR\x0btypealiases\")\n\rTopologyGroup\x12\x18\n\x07modules\x18\x01 \x03(\tR\x07modules\"K\n\x08Topology\x12?\n\x06levels\x18\x01 \x03(\x0b\x32\'.xyz.block.ftl.console.v1.TopologyGroupR\x06levels\"\x13\n\x11GetModulesRequest\"\x90\x01\n\x12GetModulesResponse\x12:\n\x07modules\x18\x01 \x03(\x0b\x32 .xyz.block.ftl.console.v1.ModuleR\x07modules\x12>\n\x08topology\x18\x02 \x01(\x0b\x32\".xyz.block.ftl.console.v1.TopologyR\x08topology\"\x16\n\x14StreamModulesRequest\"\x93\x01\n\x15StreamModulesResponse\x12:\n\x07modules\x18\x01 \x03(\x0b\x32 .xyz.block.ftl.console.v1.ModuleR\x07modules\x12>\n\x08topology\x18\x02 \x01(\x0b\x32\".xyz.block.ftl.console.v1.TopologyR\x08topology\"N\n\x10GetConfigRequest\x12\x12\n\x04name\x18\x01 \x01(\tR\x04name\x12\x1b\n\x06module\x18\x02 \x01(\tH\x00R\x06module\x88\x01\x01\x42\t\n\x07_module\")\n\x11GetConfigResponse\x12\x14\n\x05value\x18\x01 \x01(\x0cR\x05value\"d\n\x10SetConfigRequest\x12\x12\n\x04name\x18\x01 \x01(\tR\x04name\x12\x1b\n\x06module\x18\x02 \x01(\tH\x00R\x06module\x88\x01\x01\x12\x14\n\x05value\x18\x03 \x01(\x0cR\x05valueB\t\n\x07_module\")\n\x11SetConfigResponse\x12\x14\n\x05value\x18\x01 \x01(\x0cR\x05value\"N\n\x10GetSecretRequest\x12\x12\n\x04name\x18\x01 \x01(\tR\x04name\x12\x1b\n\x06module\x18\x02 \x01(\tH\x00R\x06module\x88\x01\x01\x42\t\n\x07_module\")\n\x11GetSecretResponse\x12\x14\n\x05value\x18\x01 \x01(\x0cR\x05value\"d\n\x10SetSecretRequest\x12\x12\n\x04name\x18\x01 \x01(\tR\x04name\x12\x1b\n\x06module\x18\x02 \x01(\tH\x00R\x06module\x88\x01\x01\x12\x14\n\x05value\x18\x03 \x01(\x0cR\x05valueB\t\n\x07_module\")\n\x11SetSecretResponse\x12\x14\n\x05value\x18\x01 \x01(\x0cR\x05value2\xd1\x05\n\x0e\x43onsoleService\x12J\n\x04Ping\x12\x1d.xyz.block.ftl.v1.PingRequest\x1a\x1e.xyz.block.ftl.v1.PingResponse\"\x03\x90\x02\x01\x12g\n\nGetModules\x12+.xyz.block.ftl.console.v1.GetModulesRequest\x1a,.xyz.block.ftl.console.v1.GetModulesResponse\x12r\n\rStreamModules\x12..xyz.block.ftl.console.v1.StreamModulesRequest\x1a/.xyz.block.ftl.console.v1.StreamModulesResponse0\x01\x12\x64\n\tGetConfig\x12*.xyz.block.ftl.console.v1.GetConfigRequest\x1a+.xyz.block.ftl.console.v1.GetConfigResponse\x12\x64\n\tSetConfig\x12*.xyz.block.ftl.console.v1.SetConfigRequest\x1a+.xyz.block.ftl.console.v1.SetConfigResponse\x12\x64\n\tGetSecret\x12*.xyz.block.ftl.console.v1.GetSecretRequest\x1a+.xyz.block.ftl.console.v1.GetSecretResponse\x12\x64\n\tSetSecret\x12*.xyz.block.ftl.console.v1.SetSecretRequest\x1a+.xyz.block.ftl.console.v1.SetSecretResponseBPP\x01ZLgithub.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/console/v1;pbconsoleb\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'xyz.block.ftl.console.v1.console_pb2', _globals)
if not _descriptor._USE_C_DESCRIPTORS:
  _globals['DESCRIPTOR']._loaded_options = None
  _globals['DESCRIPTOR']._serialized_options = b'P\001ZLgithub.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/console/v1;pbconsole'
  _globals['_CONSOLESERVICE'].methods_by_name['Ping']._loaded_options = None
  _globals['_CONSOLESERVICE'].methods_by_name['Ping']._serialized_options = b'\220\002\001'
  _globals['_CONFIG']._serialized_start=134
  _globals['_CONFIG']._serialized_end=261
  _globals['_DATA']._serialized_start=264
  _globals['_DATA']._serialized_end=407
  _globals['_DATABASE']._serialized_start=410
  _globals['_DATABASE']._serialized_end=545
  _globals['_ENUM']._serialized_start=547
  _globals['_ENUM']._serialized_end=666
  _globals['_TOPIC']._serialized_start=668
  _globals['_TOPIC']._serialized_end=791
  _globals['_TYPEALIAS']._serialized_start=794
  _globals['_TYPEALIAS']._serialized_end=933
  _globals['_SECRET']._serialized_start=935
  _globals['_SECRET']._serialized_end=1062
  _globals['_VERB']._serialized_start=1065
  _globals['_VERB']._serialized_end=1256
  _globals['_MODULE']._serialized_start=1259
  _globals['_MODULE']._serialized_end=1852
  _globals['_TOPOLOGYGROUP']._serialized_start=1854
  _globals['_TOPOLOGYGROUP']._serialized_end=1895
  _globals['_TOPOLOGY']._serialized_start=1897
  _globals['_TOPOLOGY']._serialized_end=1972
  _globals['_GETMODULESREQUEST']._serialized_start=1974
  _globals['_GETMODULESREQUEST']._serialized_end=1993
  _globals['_GETMODULESRESPONSE']._serialized_start=1996
  _globals['_GETMODULESRESPONSE']._serialized_end=2140
  _globals['_STREAMMODULESREQUEST']._serialized_start=2142
  _globals['_STREAMMODULESREQUEST']._serialized_end=2164
  _globals['_STREAMMODULESRESPONSE']._serialized_start=2167
  _globals['_STREAMMODULESRESPONSE']._serialized_end=2314
  _globals['_GETCONFIGREQUEST']._serialized_start=2316
  _globals['_GETCONFIGREQUEST']._serialized_end=2394
  _globals['_GETCONFIGRESPONSE']._serialized_start=2396
  _globals['_GETCONFIGRESPONSE']._serialized_end=2437
  _globals['_SETCONFIGREQUEST']._serialized_start=2439
  _globals['_SETCONFIGREQUEST']._serialized_end=2539
  _globals['_SETCONFIGRESPONSE']._serialized_start=2541
  _globals['_SETCONFIGRESPONSE']._serialized_end=2582
  _globals['_GETSECRETREQUEST']._serialized_start=2584
  _globals['_GETSECRETREQUEST']._serialized_end=2662
  _globals['_GETSECRETRESPONSE']._serialized_start=2664
  _globals['_GETSECRETRESPONSE']._serialized_end=2705
  _globals['_SETSECRETREQUEST']._serialized_start=2707
  _globals['_SETSECRETREQUEST']._serialized_end=2807
  _globals['_SETSECRETRESPONSE']._serialized_start=2809
  _globals['_SETSECRETRESPONSE']._serialized_end=2850
  _globals['_CONSOLESERVICE']._serialized_start=2853
  _globals['_CONSOLESERVICE']._serialized_end=3574
# @@protoc_insertion_point(module_scope)
