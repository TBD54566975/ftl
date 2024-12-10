from xyz.block.ftl.v1 import ftl_pb2 as _ftl_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class GetDeploymentContextRequest(_message.Message):
    __slots__ = ("deployment",)
    DEPLOYMENT_FIELD_NUMBER: _ClassVar[int]
    deployment: str
    def __init__(self, deployment: _Optional[str] = ...) -> None: ...

class GetDeploymentContextResponse(_message.Message):
    __slots__ = ("module", "deployment", "configs", "secrets", "databases", "routes")
    class DbType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = ()
        DB_TYPE_UNSPECIFIED: _ClassVar[GetDeploymentContextResponse.DbType]
        DB_TYPE_POSTGRES: _ClassVar[GetDeploymentContextResponse.DbType]
        DB_TYPE_MYSQL: _ClassVar[GetDeploymentContextResponse.DbType]
    DB_TYPE_UNSPECIFIED: GetDeploymentContextResponse.DbType
    DB_TYPE_POSTGRES: GetDeploymentContextResponse.DbType
    DB_TYPE_MYSQL: GetDeploymentContextResponse.DbType
    class DSN(_message.Message):
        __slots__ = ("name", "type", "dsn")
        NAME_FIELD_NUMBER: _ClassVar[int]
        TYPE_FIELD_NUMBER: _ClassVar[int]
        DSN_FIELD_NUMBER: _ClassVar[int]
        name: str
        type: GetDeploymentContextResponse.DbType
        dsn: str
        def __init__(self, name: _Optional[str] = ..., type: _Optional[_Union[GetDeploymentContextResponse.DbType, str]] = ..., dsn: _Optional[str] = ...) -> None: ...
    class Route(_message.Message):
        __slots__ = ("module", "uri")
        MODULE_FIELD_NUMBER: _ClassVar[int]
        URI_FIELD_NUMBER: _ClassVar[int]
        module: str
        uri: str
        def __init__(self, module: _Optional[str] = ..., uri: _Optional[str] = ...) -> None: ...
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
    DEPLOYMENT_FIELD_NUMBER: _ClassVar[int]
    CONFIGS_FIELD_NUMBER: _ClassVar[int]
    SECRETS_FIELD_NUMBER: _ClassVar[int]
    DATABASES_FIELD_NUMBER: _ClassVar[int]
    ROUTES_FIELD_NUMBER: _ClassVar[int]
    module: str
    deployment: str
    configs: _containers.ScalarMap[str, bytes]
    secrets: _containers.ScalarMap[str, bytes]
    databases: _containers.RepeatedCompositeFieldContainer[GetDeploymentContextResponse.DSN]
    routes: _containers.RepeatedCompositeFieldContainer[GetDeploymentContextResponse.Route]
    def __init__(self, module: _Optional[str] = ..., deployment: _Optional[str] = ..., configs: _Optional[_Mapping[str, bytes]] = ..., secrets: _Optional[_Mapping[str, bytes]] = ..., databases: _Optional[_Iterable[_Union[GetDeploymentContextResponse.DSN, _Mapping]]] = ..., routes: _Optional[_Iterable[_Union[GetDeploymentContextResponse.Route, _Mapping]]] = ...) -> None: ...
