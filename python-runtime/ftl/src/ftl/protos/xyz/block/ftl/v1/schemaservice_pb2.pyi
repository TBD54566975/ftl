from xyz.block.ftl.v1 import ftl_pb2 as _ftl_pb2
from xyz.block.ftl.v1.schema import schema_pb2 as _schema_pb2
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class DeploymentChangeType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    DEPLOYMENT_CHANGE_TYPE_UNSPECIFIED: _ClassVar[DeploymentChangeType]
    DEPLOYMENT_CHANGE_TYPE_ADDED: _ClassVar[DeploymentChangeType]
    DEPLOYMENT_CHANGE_TYPE_REMOVED: _ClassVar[DeploymentChangeType]
    DEPLOYMENT_CHANGE_TYPE_CHANGED: _ClassVar[DeploymentChangeType]
DEPLOYMENT_CHANGE_TYPE_UNSPECIFIED: DeploymentChangeType
DEPLOYMENT_CHANGE_TYPE_ADDED: DeploymentChangeType
DEPLOYMENT_CHANGE_TYPE_REMOVED: DeploymentChangeType
DEPLOYMENT_CHANGE_TYPE_CHANGED: DeploymentChangeType

class GetSchemaRequest(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class GetSchemaResponse(_message.Message):
    __slots__ = ("schema",)
    SCHEMA_FIELD_NUMBER: _ClassVar[int]
    schema: _schema_pb2.Schema
    def __init__(self, schema: _Optional[_Union[_schema_pb2.Schema, _Mapping]] = ...) -> None: ...

class PullSchemaRequest(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class PullSchemaResponse(_message.Message):
    __slots__ = ("deployment_key", "module_name", "schema", "more", "change_type", "module_removed")
    DEPLOYMENT_KEY_FIELD_NUMBER: _ClassVar[int]
    MODULE_NAME_FIELD_NUMBER: _ClassVar[int]
    SCHEMA_FIELD_NUMBER: _ClassVar[int]
    MORE_FIELD_NUMBER: _ClassVar[int]
    CHANGE_TYPE_FIELD_NUMBER: _ClassVar[int]
    MODULE_REMOVED_FIELD_NUMBER: _ClassVar[int]
    deployment_key: str
    module_name: str
    schema: _schema_pb2.Module
    more: bool
    change_type: DeploymentChangeType
    module_removed: bool
    def __init__(self, deployment_key: _Optional[str] = ..., module_name: _Optional[str] = ..., schema: _Optional[_Union[_schema_pb2.Module, _Mapping]] = ..., more: bool = ..., change_type: _Optional[_Union[DeploymentChangeType, str]] = ..., module_removed: bool = ...) -> None: ...
