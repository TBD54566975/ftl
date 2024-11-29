from google.protobuf import duration_pb2 as _duration_pb2
from xyz.block.ftl.v1 import ftl_pb2 as _ftl_pb2
from xyz.block.ftl.v1.schema import schema_pb2 as _schema_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class AcquireLeaseRequest(_message.Message):
    __slots__ = ("module", "key", "ttl")
    MODULE_FIELD_NUMBER: _ClassVar[int]
    KEY_FIELD_NUMBER: _ClassVar[int]
    TTL_FIELD_NUMBER: _ClassVar[int]
    module: str
    key: _containers.RepeatedScalarFieldContainer[str]
    ttl: _duration_pb2.Duration
    def __init__(self, module: _Optional[str] = ..., key: _Optional[_Iterable[str]] = ..., ttl: _Optional[_Union[_duration_pb2.Duration, _Mapping]] = ...) -> None: ...

class AcquireLeaseResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class PublishEventRequest(_message.Message):
    __slots__ = ("topic", "body", "caller")
    TOPIC_FIELD_NUMBER: _ClassVar[int]
    BODY_FIELD_NUMBER: _ClassVar[int]
    CALLER_FIELD_NUMBER: _ClassVar[int]
    topic: _schema_pb2.Ref
    body: bytes
    caller: str
    def __init__(self, topic: _Optional[_Union[_schema_pb2.Ref, _Mapping]] = ..., body: _Optional[bytes] = ..., caller: _Optional[str] = ...) -> None: ...

class PublishEventResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class ModuleContextRequest(_message.Message):
    __slots__ = ("module",)
    MODULE_FIELD_NUMBER: _ClassVar[int]
    module: str
    def __init__(self, module: _Optional[str] = ...) -> None: ...

class GetModuleContextResponse(_message.Message):
    __slots__ = ("module", "configs", "secrets", "databases")
    class DbType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = ()
        DB_TYPE_UNSPECIFIED: _ClassVar[GetModuleContextResponse.DbType]
        DB_TYPE_POSTGRES: _ClassVar[GetModuleContextResponse.DbType]
        DB_TYPE_MYSQL: _ClassVar[GetModuleContextResponse.DbType]
    DB_TYPE_UNSPECIFIED: GetModuleContextResponse.DbType
    DB_TYPE_POSTGRES: GetModuleContextResponse.DbType
    DB_TYPE_MYSQL: GetModuleContextResponse.DbType
    class Ref(_message.Message):
        __slots__ = ("module", "name")
        MODULE_FIELD_NUMBER: _ClassVar[int]
        NAME_FIELD_NUMBER: _ClassVar[int]
        module: str
        name: str
        def __init__(self, module: _Optional[str] = ..., name: _Optional[str] = ...) -> None: ...
    class DSN(_message.Message):
        __slots__ = ("name", "type", "dsn")
        NAME_FIELD_NUMBER: _ClassVar[int]
        TYPE_FIELD_NUMBER: _ClassVar[int]
        DSN_FIELD_NUMBER: _ClassVar[int]
        name: str
        type: GetModuleContextResponse.DbType
        dsn: str
        def __init__(self, name: _Optional[str] = ..., type: _Optional[_Union[GetModuleContextResponse.DbType, str]] = ..., dsn: _Optional[str] = ...) -> None: ...
    class ConfigsEntry(_message.Message):
        __slots__ = ("key", "value")
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: bytes
        def __init__(self, key: _Optional[str] = ..., value: _Optional[bytes] = ...) -> None: ...
    class SecretsEntry(_message.Message):
        __slots__ = ("key", "value")
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: bytes
        def __init__(self, key: _Optional[str] = ..., value: _Optional[bytes] = ...) -> None: ...
    MODULE_FIELD_NUMBER: _ClassVar[int]
    CONFIGS_FIELD_NUMBER: _ClassVar[int]
    SECRETS_FIELD_NUMBER: _ClassVar[int]
    DATABASES_FIELD_NUMBER: _ClassVar[int]
    module: str
    configs: _containers.ScalarMap[str, bytes]
    secrets: _containers.ScalarMap[str, bytes]
    databases: _containers.RepeatedCompositeFieldContainer[GetModuleContextResponse.DSN]
    def __init__(self, module: _Optional[str] = ..., configs: _Optional[_Mapping[str, bytes]] = ..., secrets: _Optional[_Mapping[str, bytes]] = ..., databases: _Optional[_Iterable[_Union[GetModuleContextResponse.DSN, _Mapping]]] = ...) -> None: ...
