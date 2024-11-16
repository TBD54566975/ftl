from xyz.block.ftl.v1 import ftl_pb2 as _ftl_pb2
from xyz.block.ftl.v1.schema import schema_pb2 as _schema_pb2
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

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
    __slots__ = ("deleted", "schema", "initial_batch")
    DELETED_FIELD_NUMBER: _ClassVar[int]
    SCHEMA_FIELD_NUMBER: _ClassVar[int]
    INITIAL_BATCH_FIELD_NUMBER: _ClassVar[int]
    deleted: bool
    schema: _schema_pb2.Module
    initial_batch: bool
    def __init__(self, deleted: bool = ..., schema: _Optional[_Union[_schema_pb2.Module, _Mapping]] = ..., initial_batch: bool = ...) -> None: ...

class UpsertModuleRequest(_message.Message):
    __slots__ = ("schema",)
    SCHEMA_FIELD_NUMBER: _ClassVar[int]
    schema: _schema_pb2.Module
    def __init__(self, schema: _Optional[_Union[_schema_pb2.Module, _Mapping]] = ...) -> None: ...

class UpsertModuleResponse(_message.Message):
    __slots__ = ("deployment_key",)
    DEPLOYMENT_KEY_FIELD_NUMBER: _ClassVar[int]
    deployment_key: str
    def __init__(self, deployment_key: _Optional[str] = ...) -> None: ...

class DeleteDeploymentRequest(_message.Message):
    __slots__ = ("deployment_key",)
    DEPLOYMENT_KEY_FIELD_NUMBER: _ClassVar[int]
    deployment_key: str
    def __init__(self, deployment_key: _Optional[str] = ...) -> None: ...

class DeleteDeploymentResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class DeleteModuleRequest(_message.Message):
    __slots__ = ("module_name",)
    MODULE_NAME_FIELD_NUMBER: _ClassVar[int]
    module_name: str
    def __init__(self, module_name: _Optional[str] = ...) -> None: ...

class DeleteModuleResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...
