from google.protobuf import struct_pb2 as _struct_pb2
from xyz.block.ftl.v1 import controller_pb2 as _controller_pb2
from xyz.block.ftl.v1.schema import schema_pb2 as _schema_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Resource(_message.Message):
    __slots__ = ("resource_id", "postgres", "mysql", "module")
    RESOURCE_ID_FIELD_NUMBER: _ClassVar[int]
    POSTGRES_FIELD_NUMBER: _ClassVar[int]
    MYSQL_FIELD_NUMBER: _ClassVar[int]
    MODULE_FIELD_NUMBER: _ClassVar[int]
    resource_id: str
    postgres: PostgresResource
    mysql: MysqlResource
    module: ModuleResource
    def __init__(self, resource_id: _Optional[str] = ..., postgres: _Optional[_Union[PostgresResource, _Mapping]] = ..., mysql: _Optional[_Union[MysqlResource, _Mapping]] = ..., module: _Optional[_Union[ModuleResource, _Mapping]] = ...) -> None: ...

class PostgresResource(_message.Message):
    __slots__ = ("output",)
    class PostgresResourceOutput(_message.Message):
        __slots__ = ("read_dsn", "write_dsn")
        READ_DSN_FIELD_NUMBER: _ClassVar[int]
        WRITE_DSN_FIELD_NUMBER: _ClassVar[int]
        read_dsn: str
        write_dsn: str
        def __init__(self, read_dsn: _Optional[str] = ..., write_dsn: _Optional[str] = ...) -> None: ...
    OUTPUT_FIELD_NUMBER: _ClassVar[int]
    output: PostgresResource.PostgresResourceOutput
    def __init__(self, output: _Optional[_Union[PostgresResource.PostgresResourceOutput, _Mapping]] = ...) -> None: ...

class MysqlResource(_message.Message):
    __slots__ = ("output",)
    class MysqlResourceOutput(_message.Message):
        __slots__ = ("read_dsn", "write_dsn")
        READ_DSN_FIELD_NUMBER: _ClassVar[int]
        WRITE_DSN_FIELD_NUMBER: _ClassVar[int]
        read_dsn: str
        write_dsn: str
        def __init__(self, read_dsn: _Optional[str] = ..., write_dsn: _Optional[str] = ...) -> None: ...
    OUTPUT_FIELD_NUMBER: _ClassVar[int]
    output: MysqlResource.MysqlResourceOutput
    def __init__(self, output: _Optional[_Union[MysqlResource.MysqlResourceOutput, _Mapping]] = ...) -> None: ...

class ModuleResource(_message.Message):
    __slots__ = ("output", "schema", "artefacts", "labels")
    class ModuleResourceOutput(_message.Message):
        __slots__ = ("deployment_key",)
        DEPLOYMENT_KEY_FIELD_NUMBER: _ClassVar[int]
        deployment_key: str
        def __init__(self, deployment_key: _Optional[str] = ...) -> None: ...
    OUTPUT_FIELD_NUMBER: _ClassVar[int]
    SCHEMA_FIELD_NUMBER: _ClassVar[int]
    ARTEFACTS_FIELD_NUMBER: _ClassVar[int]
    LABELS_FIELD_NUMBER: _ClassVar[int]
    output: ModuleResource.ModuleResourceOutput
    schema: _schema_pb2.Module
    artefacts: _containers.RepeatedCompositeFieldContainer[_controller_pb2.DeploymentArtefact]
    labels: _struct_pb2.Struct
    def __init__(self, output: _Optional[_Union[ModuleResource.ModuleResourceOutput, _Mapping]] = ..., schema: _Optional[_Union[_schema_pb2.Module, _Mapping]] = ..., artefacts: _Optional[_Iterable[_Union[_controller_pb2.DeploymentArtefact, _Mapping]]] = ..., labels: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ...) -> None: ...
