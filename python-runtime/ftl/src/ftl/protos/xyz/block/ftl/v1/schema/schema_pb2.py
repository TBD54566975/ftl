# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# NO CHECKED-IN PROTOBUF GENCODE
# source: xyz/block/ftl/v1/schema/schema.proto
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
    'xyz/block/ftl/v1/schema/schema.proto'
)
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from google.protobuf import timestamp_pb2 as google_dot_protobuf_dot_timestamp__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n$xyz/block/ftl/v1/schema/schema.proto\x12\x17xyz.block.ftl.v1.schema\x1a\x1fgoogle/protobuf/timestamp.proto\"G\n\x03\x41ny\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.v1.schema.PositionH\x00R\x03pos\x88\x01\x01\x42\x06\n\x04_pos\"\x82\x01\n\x05\x41rray\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.v1.schema.PositionH\x00R\x03pos\x88\x01\x01\x12\x37\n\x07\x65lement\x18\x02 \x01(\x0b\x32\x1d.xyz.block.ftl.v1.schema.TypeR\x07\x65lementB\x06\n\x04_pos\"H\n\x04\x42ool\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.v1.schema.PositionH\x00R\x03pos\x88\x01\x01\x42\x06\n\x04_pos\"I\n\x05\x42ytes\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.v1.schema.PositionH\x00R\x03pos\x88\x01\x01\x42\x06\n\x04_pos\"\xad\x01\n\x06\x43onfig\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.v1.schema.PositionH\x00R\x03pos\x88\x01\x01\x12\x1a\n\x08\x63omments\x18\x02 \x03(\tR\x08\x63omments\x12\x12\n\x04name\x18\x03 \x01(\tR\x04name\x12\x31\n\x04type\x18\x04 \x01(\x0b\x32\x1d.xyz.block.ftl.v1.schema.TypeR\x04typeB\x06\n\x04_pos\"\xd8\x02\n\x04\x44\x61ta\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.v1.schema.PositionH\x00R\x03pos\x88\x01\x01\x12\x1a\n\x08\x63omments\x18\x02 \x03(\tR\x08\x63omments\x12\x16\n\x06\x65xport\x18\x03 \x01(\x08R\x06\x65xport\x12\x12\n\x04name\x18\x04 \x01(\tR\x04name\x12O\n\x0ftype_parameters\x18\x05 \x03(\x0b\x32&.xyz.block.ftl.v1.schema.TypeParameterR\x0etypeParameters\x12\x36\n\x06\x66ields\x18\x06 \x03(\x0b\x32\x1e.xyz.block.ftl.v1.schema.FieldR\x06\x66ields\x12=\n\x08metadata\x18\x07 \x03(\x0b\x32!.xyz.block.ftl.v1.schema.MetadataR\x08metadataB\x06\n\x04_pos\"\xe7\x01\n\x08\x44\x61tabase\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.v1.schema.PositionH\x00R\x03pos\x88\x01\x01\x12I\n\x07runtime\x18\x92\xf7\x01 \x01(\x0b\x32(.xyz.block.ftl.v1.schema.DatabaseRuntimeH\x01R\x07runtime\x88\x01\x01\x12\x1a\n\x08\x63omments\x18\x02 \x03(\tR\x08\x63omments\x12\x12\n\x04type\x18\x04 \x01(\tR\x04type\x12\x12\n\x04name\x18\x03 \x01(\tR\x04nameB\x06\n\x04_posB\n\n\x08_runtime\"#\n\x0f\x44\x61tabaseRuntime\x12\x10\n\x03\x64sn\x18\x01 \x01(\tR\x03\x64sn\"\xaf\x04\n\x04\x44\x65\x63l\x12\x39\n\x06\x63onfig\x18\x06 \x01(\x0b\x32\x1f.xyz.block.ftl.v1.schema.ConfigH\x00R\x06\x63onfig\x12\x33\n\x04\x64\x61ta\x18\x01 \x01(\x0b\x32\x1d.xyz.block.ftl.v1.schema.DataH\x00R\x04\x64\x61ta\x12?\n\x08\x64\x61tabase\x18\x03 \x01(\x0b\x32!.xyz.block.ftl.v1.schema.DatabaseH\x00R\x08\x64\x61tabase\x12\x33\n\x04\x65num\x18\x04 \x01(\x0b\x32\x1d.xyz.block.ftl.v1.schema.EnumH\x00R\x04\x65num\x12\x39\n\x06secret\x18\x07 \x01(\x0b\x32\x1f.xyz.block.ftl.v1.schema.SecretH\x00R\x06secret\x12K\n\x0csubscription\x18\n \x01(\x0b\x32%.xyz.block.ftl.v1.schema.SubscriptionH\x00R\x0csubscription\x12\x36\n\x05topic\x18\t \x01(\x0b\x32\x1e.xyz.block.ftl.v1.schema.TopicH\x00R\x05topic\x12\x43\n\ntype_alias\x18\x05 \x01(\x0b\x32\".xyz.block.ftl.v1.schema.TypeAliasH\x00R\ttypeAlias\x12\x33\n\x04verb\x18\x02 \x01(\x0b\x32\x1d.xyz.block.ftl.v1.schema.VerbH\x00R\x04verbB\x07\n\x05value\"\x93\x02\n\x04\x45num\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.v1.schema.PositionH\x00R\x03pos\x88\x01\x01\x12\x1a\n\x08\x63omments\x18\x02 \x03(\tR\x08\x63omments\x12\x16\n\x06\x65xport\x18\x03 \x01(\x08R\x06\x65xport\x12\x12\n\x04name\x18\x04 \x01(\tR\x04name\x12\x36\n\x04type\x18\x05 \x01(\x0b\x32\x1d.xyz.block.ftl.v1.schema.TypeH\x01R\x04type\x88\x01\x01\x12@\n\x08variants\x18\x06 \x03(\x0b\x32$.xyz.block.ftl.v1.schema.EnumVariantR\x08variantsB\x06\n\x04_posB\x07\n\x05_type\"\xb5\x01\n\x0b\x45numVariant\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.v1.schema.PositionH\x00R\x03pos\x88\x01\x01\x12\x1a\n\x08\x63omments\x18\x02 \x03(\tR\x08\x63omments\x12\x12\n\x04name\x18\x03 \x01(\tR\x04name\x12\x34\n\x05value\x18\x04 \x01(\x0b\x32\x1e.xyz.block.ftl.v1.schema.ValueR\x05valueB\x06\n\x04_pos\"\xeb\x01\n\x05\x46ield\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.v1.schema.PositionH\x00R\x03pos\x88\x01\x01\x12\x1a\n\x08\x63omments\x18\x03 \x03(\tR\x08\x63omments\x12\x12\n\x04name\x18\x02 \x01(\tR\x04name\x12\x31\n\x04type\x18\x04 \x01(\x0b\x32\x1d.xyz.block.ftl.v1.schema.TypeR\x04type\x12=\n\x08metadata\x18\x05 \x03(\x0b\x32!.xyz.block.ftl.v1.schema.MetadataR\x08metadataB\x06\n\x04_pos\"I\n\x05\x46loat\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.v1.schema.PositionH\x00R\x03pos\x88\x01\x01\x42\x06\n\x04_pos\"\xe7\x01\n\x14IngressPathComponent\x12_\n\x14ingress_path_literal\x18\x01 \x01(\x0b\x32+.xyz.block.ftl.v1.schema.IngressPathLiteralH\x00R\x12ingressPathLiteral\x12\x65\n\x16ingress_path_parameter\x18\x02 \x01(\x0b\x32-.xyz.block.ftl.v1.schema.IngressPathParameterH\x00R\x14ingressPathParameterB\x07\n\x05value\"j\n\x12IngressPathLiteral\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.v1.schema.PositionH\x00R\x03pos\x88\x01\x01\x12\x12\n\x04text\x18\x02 \x01(\tR\x04textB\x06\n\x04_pos\"l\n\x14IngressPathParameter\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.v1.schema.PositionH\x00R\x03pos\x88\x01\x01\x12\x12\n\x04name\x18\x02 \x01(\tR\x04nameB\x06\n\x04_pos\"G\n\x03Int\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.v1.schema.PositionH\x00R\x03pos\x88\x01\x01\x42\x06\n\x04_pos\"b\n\x08IntValue\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.v1.schema.PositionH\x00R\x03pos\x88\x01\x01\x12\x14\n\x05value\x18\x02 \x01(\x03R\x05valueB\x06\n\x04_pos\"\xad\x01\n\x03Map\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.v1.schema.PositionH\x00R\x03pos\x88\x01\x01\x12/\n\x03key\x18\x02 \x01(\x0b\x32\x1d.xyz.block.ftl.v1.schema.TypeR\x03key\x12\x33\n\x05value\x18\x03 \x01(\x0b\x32\x1d.xyz.block.ftl.v1.schema.TypeR\x05valueB\x06\n\x04_pos\"\x94\x06\n\x08Metadata\x12>\n\x05\x61lias\x18\x05 \x01(\x0b\x32&.xyz.block.ftl.v1.schema.MetadataAliasH\x00R\x05\x61lias\x12>\n\x05\x63\x61lls\x18\x01 \x01(\x0b\x32&.xyz.block.ftl.v1.schema.MetadataCallsH\x00R\x05\x63\x61lls\x12\x41\n\x06\x63onfig\x18\n \x01(\x0b\x32\'.xyz.block.ftl.v1.schema.MetadataConfigH\x00R\x06\x63onfig\x12\x45\n\x08\x63ron_job\x18\x03 \x01(\x0b\x32(.xyz.block.ftl.v1.schema.MetadataCronJobH\x00R\x07\x63ronJob\x12J\n\tdatabases\x18\x04 \x01(\x0b\x32*.xyz.block.ftl.v1.schema.MetadataDatabasesH\x00R\tdatabases\x12G\n\x08\x65ncoding\x18\t \x01(\x0b\x32).xyz.block.ftl.v1.schema.MetadataEncodingH\x00R\x08\x65ncoding\x12\x44\n\x07ingress\x18\x02 \x01(\x0b\x32(.xyz.block.ftl.v1.schema.MetadataIngressH\x00R\x07ingress\x12>\n\x05retry\x18\x06 \x01(\x0b\x32&.xyz.block.ftl.v1.schema.MetadataRetryH\x00R\x05retry\x12\x44\n\x07secrets\x18\x0b \x01(\x0b\x32(.xyz.block.ftl.v1.schema.MetadataSecretsH\x00R\x07secrets\x12M\n\nsubscriber\x18\x07 \x01(\x0b\x32+.xyz.block.ftl.v1.schema.MetadataSubscriberH\x00R\nsubscriber\x12\x45\n\x08type_map\x18\x08 \x01(\x0b\x32(.xyz.block.ftl.v1.schema.MetadataTypeMapH\x00R\x07typeMapB\x07\n\x05value\"\x9f\x01\n\rMetadataAlias\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.v1.schema.PositionH\x00R\x03pos\x88\x01\x01\x12\x36\n\x04kind\x18\x02 \x01(\x0e\x32\".xyz.block.ftl.v1.schema.AliasKindR\x04kind\x12\x14\n\x05\x61lias\x18\x03 \x01(\tR\x05\x61liasB\x06\n\x04_pos\"\x85\x01\n\rMetadataCalls\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.v1.schema.PositionH\x00R\x03pos\x88\x01\x01\x12\x32\n\x05\x63\x61lls\x18\x02 \x03(\x0b\x32\x1c.xyz.block.ftl.v1.schema.RefR\x05\x63\x61llsB\x06\n\x04_pos\"\x88\x01\n\x0eMetadataConfig\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.v1.schema.PositionH\x00R\x03pos\x88\x01\x01\x12\x34\n\x06\x63onfig\x18\x02 \x03(\x0b\x32\x1c.xyz.block.ftl.v1.schema.RefR\x06\x63onfigB\x06\n\x04_pos\"g\n\x0fMetadataCronJob\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.v1.schema.PositionH\x00R\x03pos\x88\x01\x01\x12\x12\n\x04\x63ron\x18\x02 \x01(\tR\x04\x63ronB\x06\n\x04_pos\"\x89\x01\n\x11MetadataDatabases\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.v1.schema.PositionH\x00R\x03pos\x88\x01\x01\x12\x32\n\x05\x63\x61lls\x18\x02 \x03(\x0b\x32\x1c.xyz.block.ftl.v1.schema.RefR\x05\x63\x61llsB\x06\n\x04_pos\"\x82\x01\n\x10MetadataEncoding\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.v1.schema.PositionH\x00R\x03pos\x88\x01\x01\x12\x12\n\x04type\x18\x02 \x01(\tR\x04type\x12\x18\n\x07lenient\x18\x03 \x01(\x08R\x07lenientB\x06\n\x04_pos\"\xc2\x01\n\x0fMetadataIngress\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.v1.schema.PositionH\x00R\x03pos\x88\x01\x01\x12\x12\n\x04type\x18\x02 \x01(\tR\x04type\x12\x16\n\x06method\x18\x03 \x01(\tR\x06method\x12\x41\n\x04path\x18\x04 \x03(\x0b\x32-.xyz.block.ftl.v1.schema.IngressPathComponentR\x04pathB\x06\n\x04_pos\"\xfb\x01\n\rMetadataRetry\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.v1.schema.PositionH\x00R\x03pos\x88\x01\x01\x12\x19\n\x05\x63ount\x18\x02 \x01(\x03H\x01R\x05\x63ount\x88\x01\x01\x12\x1f\n\x0bmin_backoff\x18\x03 \x01(\tR\nminBackoff\x12\x1f\n\x0bmax_backoff\x18\x04 \x01(\tR\nmaxBackoff\x12\x37\n\x05\x63\x61tch\x18\x05 \x01(\x0b\x32\x1c.xyz.block.ftl.v1.schema.RefH\x02R\x05\x63\x61tch\x88\x01\x01\x42\x06\n\x04_posB\x08\n\x06_countB\x08\n\x06_catch\"\x8b\x01\n\x0fMetadataSecrets\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.v1.schema.PositionH\x00R\x03pos\x88\x01\x01\x12\x36\n\x07secrets\x18\x02 \x03(\x0b\x32\x1c.xyz.block.ftl.v1.schema.RefR\x07secretsB\x06\n\x04_pos\"j\n\x12MetadataSubscriber\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.v1.schema.PositionH\x00R\x03pos\x88\x01\x01\x12\x12\n\x04name\x18\x02 \x01(\tR\x04nameB\x06\n\x04_pos\"\x8e\x01\n\x0fMetadataTypeMap\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.v1.schema.PositionH\x00R\x03pos\x88\x01\x01\x12\x18\n\x07runtime\x18\x02 \x01(\tR\x07runtime\x12\x1f\n\x0bnative_name\x18\x03 \x01(\tR\nnativeNameB\x06\n\x04_pos\"\x9e\x02\n\x06Module\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.v1.schema.PositionH\x00R\x03pos\x88\x01\x01\x12\x1a\n\x08\x63omments\x18\x02 \x03(\tR\x08\x63omments\x12\x18\n\x07\x62uiltin\x18\x03 \x01(\x08R\x07\x62uiltin\x12\x12\n\x04name\x18\x04 \x01(\tR\x04name\x12\x33\n\x05\x64\x65\x63ls\x18\x05 \x03(\x0b\x32\x1d.xyz.block.ftl.v1.schema.DeclR\x05\x64\x65\x63ls\x12G\n\x07runtime\x18\x92\xf7\x01 \x01(\x0b\x32&.xyz.block.ftl.v1.schema.ModuleRuntimeH\x01R\x07runtime\x88\x01\x01\x42\x06\n\x04_posB\n\n\x08_runtime\"\xdf\x01\n\rModuleRuntime\x12;\n\x0b\x63reate_time\x18\x01 \x01(\x0b\x32\x1a.google.protobuf.TimestampR\ncreateTime\x12\x1a\n\x08language\x18\x02 \x01(\tR\x08language\x12!\n\x0cmin_replicas\x18\x03 \x01(\x05R\x0bminReplicas\x12\x13\n\x02os\x18\x04 \x01(\tH\x00R\x02os\x88\x01\x01\x12\x17\n\x04\x61rch\x18\x05 \x01(\tH\x01R\x04\x61rch\x88\x01\x01\x12\x14\n\x05image\x18\x06 \x01(\tR\x05imageB\x05\n\x03_osB\x07\n\x05_arch\"\x8d\x01\n\x08Optional\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.v1.schema.PositionH\x00R\x03pos\x88\x01\x01\x12\x36\n\x04type\x18\x02 \x01(\x0b\x32\x1d.xyz.block.ftl.v1.schema.TypeH\x01R\x04type\x88\x01\x01\x42\x06\n\x04_posB\x07\n\x05_type\"R\n\x08Position\x12\x1a\n\x08\x66ilename\x18\x01 \x01(\tR\x08\x66ilename\x12\x12\n\x04line\x18\x02 \x01(\x03R\x04line\x12\x16\n\x06\x63olumn\x18\x03 \x01(\x03R\x06\x63olumn\"\xbb\x01\n\x03Ref\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.v1.schema.PositionH\x00R\x03pos\x88\x01\x01\x12\x16\n\x06module\x18\x03 \x01(\tR\x06module\x12\x12\n\x04name\x18\x02 \x01(\tR\x04name\x12\x46\n\x0ftype_parameters\x18\x04 \x03(\x0b\x32\x1d.xyz.block.ftl.v1.schema.TypeR\x0etypeParametersB\x06\n\x04_pos\"\x85\x01\n\x06Schema\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.v1.schema.PositionH\x00R\x03pos\x88\x01\x01\x12\x39\n\x07modules\x18\x02 \x03(\x0b\x32\x1f.xyz.block.ftl.v1.schema.ModuleR\x07modulesB\x06\n\x04_pos\"\xad\x01\n\x06Secret\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.v1.schema.PositionH\x00R\x03pos\x88\x01\x01\x12\x1a\n\x08\x63omments\x18\x02 \x03(\tR\x08\x63omments\x12\x12\n\x04name\x18\x03 \x01(\tR\x04name\x12\x31\n\x04type\x18\x04 \x01(\x0b\x32\x1d.xyz.block.ftl.v1.schema.TypeR\x04typeB\x06\n\x04_pos\"J\n\x06String\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.v1.schema.PositionH\x00R\x03pos\x88\x01\x01\x42\x06\n\x04_pos\"e\n\x0bStringValue\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.v1.schema.PositionH\x00R\x03pos\x88\x01\x01\x12\x14\n\x05value\x18\x02 \x01(\tR\x05valueB\x06\n\x04_pos\"\xb4\x01\n\x0cSubscription\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.v1.schema.PositionH\x00R\x03pos\x88\x01\x01\x12\x1a\n\x08\x63omments\x18\x02 \x03(\tR\x08\x63omments\x12\x12\n\x04name\x18\x03 \x01(\tR\x04name\x12\x32\n\x05topic\x18\x04 \x01(\x0b\x32\x1c.xyz.block.ftl.v1.schema.RefR\x05topicB\x06\n\x04_pos\"H\n\x04Time\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.v1.schema.PositionH\x00R\x03pos\x88\x01\x01\x42\x06\n\x04_pos\"\xc6\x01\n\x05Topic\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.v1.schema.PositionH\x00R\x03pos\x88\x01\x01\x12\x1a\n\x08\x63omments\x18\x02 \x03(\tR\x08\x63omments\x12\x16\n\x06\x65xport\x18\x03 \x01(\x08R\x06\x65xport\x12\x12\n\x04name\x18\x04 \x01(\tR\x04name\x12\x33\n\x05\x65vent\x18\x05 \x01(\x0b\x32\x1d.xyz.block.ftl.v1.schema.TypeR\x05\x65ventB\x06\n\x04_pos\"\x9a\x05\n\x04Type\x12\x30\n\x03\x61ny\x18\t \x01(\x0b\x32\x1c.xyz.block.ftl.v1.schema.AnyH\x00R\x03\x61ny\x12\x36\n\x05\x61rray\x18\x07 \x01(\x0b\x32\x1e.xyz.block.ftl.v1.schema.ArrayH\x00R\x05\x61rray\x12\x33\n\x04\x62ool\x18\x05 \x01(\x0b\x32\x1d.xyz.block.ftl.v1.schema.BoolH\x00R\x04\x62ool\x12\x36\n\x05\x62ytes\x18\x04 \x01(\x0b\x32\x1e.xyz.block.ftl.v1.schema.BytesH\x00R\x05\x62ytes\x12\x36\n\x05\x66loat\x18\x02 \x01(\x0b\x32\x1e.xyz.block.ftl.v1.schema.FloatH\x00R\x05\x66loat\x12\x30\n\x03int\x18\x01 \x01(\x0b\x32\x1c.xyz.block.ftl.v1.schema.IntH\x00R\x03int\x12\x30\n\x03map\x18\x08 \x01(\x0b\x32\x1c.xyz.block.ftl.v1.schema.MapH\x00R\x03map\x12?\n\x08optional\x18\x0c \x01(\x0b\x32!.xyz.block.ftl.v1.schema.OptionalH\x00R\x08optional\x12\x30\n\x03ref\x18\x0b \x01(\x0b\x32\x1c.xyz.block.ftl.v1.schema.RefH\x00R\x03ref\x12\x39\n\x06string\x18\x03 \x01(\x0b\x32\x1f.xyz.block.ftl.v1.schema.StringH\x00R\x06string\x12\x33\n\x04time\x18\x06 \x01(\x0b\x32\x1d.xyz.block.ftl.v1.schema.TimeH\x00R\x04time\x12\x33\n\x04unit\x18\n \x01(\x0b\x32\x1d.xyz.block.ftl.v1.schema.UnitH\x00R\x04unitB\x07\n\x05value\"\x87\x02\n\tTypeAlias\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.v1.schema.PositionH\x00R\x03pos\x88\x01\x01\x12\x1a\n\x08\x63omments\x18\x02 \x03(\tR\x08\x63omments\x12\x16\n\x06\x65xport\x18\x03 \x01(\x08R\x06\x65xport\x12\x12\n\x04name\x18\x04 \x01(\tR\x04name\x12\x31\n\x04type\x18\x05 \x01(\x0b\x32\x1d.xyz.block.ftl.v1.schema.TypeR\x04type\x12=\n\x08metadata\x18\x06 \x03(\x0b\x32!.xyz.block.ftl.v1.schema.MetadataR\x08metadataB\x06\n\x04_pos\"e\n\rTypeParameter\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.v1.schema.PositionH\x00R\x03pos\x88\x01\x01\x12\x12\n\x04name\x18\x02 \x01(\tR\x04nameB\x06\n\x04_pos\"\x82\x01\n\tTypeValue\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.v1.schema.PositionH\x00R\x03pos\x88\x01\x01\x12\x33\n\x05value\x18\x02 \x01(\x0b\x32\x1d.xyz.block.ftl.v1.schema.TypeR\x05valueB\x06\n\x04_pos\"H\n\x04Unit\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.v1.schema.PositionH\x00R\x03pos\x88\x01\x01\x42\x06\n\x04_pos\"\xe2\x01\n\x05Value\x12@\n\tint_value\x18\x02 \x01(\x0b\x32!.xyz.block.ftl.v1.schema.IntValueH\x00R\x08intValue\x12I\n\x0cstring_value\x18\x01 \x01(\x0b\x32$.xyz.block.ftl.v1.schema.StringValueH\x00R\x0bstringValue\x12\x43\n\ntype_value\x18\x03 \x01(\x0b\x32\".xyz.block.ftl.v1.schema.TypeValueH\x00R\ttypeValueB\x07\n\x05value\"\x96\x03\n\x04Verb\x12\x38\n\x03pos\x18\x01 \x01(\x0b\x32!.xyz.block.ftl.v1.schema.PositionH\x00R\x03pos\x88\x01\x01\x12\x1a\n\x08\x63omments\x18\x02 \x03(\tR\x08\x63omments\x12\x16\n\x06\x65xport\x18\x03 \x01(\x08R\x06\x65xport\x12\x12\n\x04name\x18\x04 \x01(\tR\x04name\x12\x37\n\x07request\x18\x05 \x01(\x0b\x32\x1d.xyz.block.ftl.v1.schema.TypeR\x07request\x12\x39\n\x08response\x18\x06 \x01(\x0b\x32\x1d.xyz.block.ftl.v1.schema.TypeR\x08response\x12=\n\x08metadata\x18\x07 \x03(\x0b\x32!.xyz.block.ftl.v1.schema.MetadataR\x08metadata\x12\x45\n\x07runtime\x18\x92\xf7\x01 \x01(\x0b\x32$.xyz.block.ftl.v1.schema.VerbRuntimeH\x01R\x07runtime\x88\x01\x01\x42\x06\n\x04_posB\n\n\x08_runtime\"\x85\x01\n\x0bVerbRuntime\x12;\n\x0b\x63reate_time\x18\x01 \x01(\x0b\x32\x1a.google.protobuf.TimestampR\ncreateTime\x12\x39\n\nstart_time\x18\x02 \x01(\x0b\x32\x1a.google.protobuf.TimestampR\tstartTime* \n\tAliasKind\x12\x13\n\x0f\x41LIAS_KIND_JSON\x10\x00\x42NP\x01ZJgithub.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema;schemapbb\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'xyz.block.ftl.v1.schema.schema_pb2', _globals)
if not _descriptor._USE_C_DESCRIPTORS:
  _globals['DESCRIPTOR']._loaded_options = None
  _globals['DESCRIPTOR']._serialized_options = b'P\001ZJgithub.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema;schemapb'
  _globals['_ALIASKIND']._serialized_start=9738
  _globals['_ALIASKIND']._serialized_end=9770
  _globals['_ANY']._serialized_start=98
  _globals['_ANY']._serialized_end=169
  _globals['_ARRAY']._serialized_start=172
  _globals['_ARRAY']._serialized_end=302
  _globals['_BOOL']._serialized_start=304
  _globals['_BOOL']._serialized_end=376
  _globals['_BYTES']._serialized_start=378
  _globals['_BYTES']._serialized_end=451
  _globals['_CONFIG']._serialized_start=454
  _globals['_CONFIG']._serialized_end=627
  _globals['_DATA']._serialized_start=630
  _globals['_DATA']._serialized_end=974
  _globals['_DATABASE']._serialized_start=977
  _globals['_DATABASE']._serialized_end=1208
  _globals['_DATABASERUNTIME']._serialized_start=1210
  _globals['_DATABASERUNTIME']._serialized_end=1245
  _globals['_DECL']._serialized_start=1248
  _globals['_DECL']._serialized_end=1807
  _globals['_ENUM']._serialized_start=1810
  _globals['_ENUM']._serialized_end=2085
  _globals['_ENUMVARIANT']._serialized_start=2088
  _globals['_ENUMVARIANT']._serialized_end=2269
  _globals['_FIELD']._serialized_start=2272
  _globals['_FIELD']._serialized_end=2507
  _globals['_FLOAT']._serialized_start=2509
  _globals['_FLOAT']._serialized_end=2582
  _globals['_INGRESSPATHCOMPONENT']._serialized_start=2585
  _globals['_INGRESSPATHCOMPONENT']._serialized_end=2816
  _globals['_INGRESSPATHLITERAL']._serialized_start=2818
  _globals['_INGRESSPATHLITERAL']._serialized_end=2924
  _globals['_INGRESSPATHPARAMETER']._serialized_start=2926
  _globals['_INGRESSPATHPARAMETER']._serialized_end=3034
  _globals['_INT']._serialized_start=3036
  _globals['_INT']._serialized_end=3107
  _globals['_INTVALUE']._serialized_start=3109
  _globals['_INTVALUE']._serialized_end=3207
  _globals['_MAP']._serialized_start=3210
  _globals['_MAP']._serialized_end=3383
  _globals['_METADATA']._serialized_start=3386
  _globals['_METADATA']._serialized_end=4174
  _globals['_METADATAALIAS']._serialized_start=4177
  _globals['_METADATAALIAS']._serialized_end=4336
  _globals['_METADATACALLS']._serialized_start=4339
  _globals['_METADATACALLS']._serialized_end=4472
  _globals['_METADATACONFIG']._serialized_start=4475
  _globals['_METADATACONFIG']._serialized_end=4611
  _globals['_METADATACRONJOB']._serialized_start=4613
  _globals['_METADATACRONJOB']._serialized_end=4716
  _globals['_METADATADATABASES']._serialized_start=4719
  _globals['_METADATADATABASES']._serialized_end=4856
  _globals['_METADATAENCODING']._serialized_start=4859
  _globals['_METADATAENCODING']._serialized_end=4989
  _globals['_METADATAINGRESS']._serialized_start=4992
  _globals['_METADATAINGRESS']._serialized_end=5186
  _globals['_METADATARETRY']._serialized_start=5189
  _globals['_METADATARETRY']._serialized_end=5440
  _globals['_METADATASECRETS']._serialized_start=5443
  _globals['_METADATASECRETS']._serialized_end=5582
  _globals['_METADATASUBSCRIBER']._serialized_start=5584
  _globals['_METADATASUBSCRIBER']._serialized_end=5690
  _globals['_METADATATYPEMAP']._serialized_start=5693
  _globals['_METADATATYPEMAP']._serialized_end=5835
  _globals['_MODULE']._serialized_start=5838
  _globals['_MODULE']._serialized_end=6124
  _globals['_MODULERUNTIME']._serialized_start=6127
  _globals['_MODULERUNTIME']._serialized_end=6350
  _globals['_OPTIONAL']._serialized_start=6353
  _globals['_OPTIONAL']._serialized_end=6494
  _globals['_POSITION']._serialized_start=6496
  _globals['_POSITION']._serialized_end=6578
  _globals['_REF']._serialized_start=6581
  _globals['_REF']._serialized_end=6768
  _globals['_SCHEMA']._serialized_start=6771
  _globals['_SCHEMA']._serialized_end=6904
  _globals['_SECRET']._serialized_start=6907
  _globals['_SECRET']._serialized_end=7080
  _globals['_STRING']._serialized_start=7082
  _globals['_STRING']._serialized_end=7156
  _globals['_STRINGVALUE']._serialized_start=7158
  _globals['_STRINGVALUE']._serialized_end=7259
  _globals['_SUBSCRIPTION']._serialized_start=7262
  _globals['_SUBSCRIPTION']._serialized_end=7442
  _globals['_TIME']._serialized_start=7444
  _globals['_TIME']._serialized_end=7516
  _globals['_TOPIC']._serialized_start=7519
  _globals['_TOPIC']._serialized_end=7717
  _globals['_TYPE']._serialized_start=7720
  _globals['_TYPE']._serialized_end=8386
  _globals['_TYPEALIAS']._serialized_start=8389
  _globals['_TYPEALIAS']._serialized_end=8652
  _globals['_TYPEPARAMETER']._serialized_start=8654
  _globals['_TYPEPARAMETER']._serialized_end=8755
  _globals['_TYPEVALUE']._serialized_start=8758
  _globals['_TYPEVALUE']._serialized_end=8888
  _globals['_UNIT']._serialized_start=8890
  _globals['_UNIT']._serialized_end=8962
  _globals['_VALUE']._serialized_start=8965
  _globals['_VALUE']._serialized_end=9191
  _globals['_VERB']._serialized_start=9194
  _globals['_VERB']._serialized_end=9600
  _globals['_VERBRUNTIME']._serialized_start=9603
  _globals['_VERBRUNTIME']._serialized_end=9736
# @@protoc_insertion_point(module_scope)
