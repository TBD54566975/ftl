from xyz.block.ftl.v1 import ftl_pb2 as _ftl_pb2
from xyz.block.ftl.v1.schema import schema_pb2 as _schema_pb2
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class CallRequest(_message.Message):
    __slots__ = ("metadata", "verb", "body")
    METADATA_FIELD_NUMBER: _ClassVar[int]
    VERB_FIELD_NUMBER: _ClassVar[int]
    BODY_FIELD_NUMBER: _ClassVar[int]
    metadata: _ftl_pb2.Metadata
    verb: _schema_pb2.Ref
    body: bytes
    def __init__(self, metadata: _Optional[_Union[_ftl_pb2.Metadata, _Mapping]] = ..., verb: _Optional[_Union[_schema_pb2.Ref, _Mapping]] = ..., body: _Optional[bytes] = ...) -> None: ...

class CallResponse(_message.Message):
    __slots__ = ("body", "error")
    class Error(_message.Message):
        __slots__ = ("message", "stack")
        MESSAGE_FIELD_NUMBER: _ClassVar[int]
        STACK_FIELD_NUMBER: _ClassVar[int]
        message: str
        stack: str
        def __init__(self, message: _Optional[str] = ..., stack: _Optional[str] = ...) -> None: ...
    BODY_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    body: bytes
    error: CallResponse.Error
    def __init__(self, body: _Optional[bytes] = ..., error: _Optional[_Union[CallResponse.Error, _Mapping]] = ...) -> None: ...
