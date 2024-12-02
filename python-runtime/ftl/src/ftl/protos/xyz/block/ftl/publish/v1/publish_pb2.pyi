from xyz.block.ftl.schema.v1 import schema_pb2 as _schema_pb2
from xyz.block.ftl.v1 import ftl_pb2 as _ftl_pb2
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class PublishEventRequest(_message.Message):
    __slots__ = ("topic", "body", "key", "caller")
    TOPIC_FIELD_NUMBER: _ClassVar[int]
    BODY_FIELD_NUMBER: _ClassVar[int]
    KEY_FIELD_NUMBER: _ClassVar[int]
    CALLER_FIELD_NUMBER: _ClassVar[int]
    topic: _schema_pb2.Ref
    body: bytes
    key: str
    caller: str
    def __init__(self, topic: _Optional[_Union[_schema_pb2.Ref, _Mapping]] = ..., body: _Optional[bytes] = ..., key: _Optional[str] = ..., caller: _Optional[str] = ...) -> None: ...

class PublishEventResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...
