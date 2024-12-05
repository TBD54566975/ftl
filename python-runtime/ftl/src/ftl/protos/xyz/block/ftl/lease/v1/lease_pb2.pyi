from google.protobuf import duration_pb2 as _duration_pb2
from xyz.block.ftl.v1 import ftl_pb2 as _ftl_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class AcquireLeaseRequest(_message.Message):
    __slots__ = ("key", "ttl")
    KEY_FIELD_NUMBER: _ClassVar[int]
    TTL_FIELD_NUMBER: _ClassVar[int]
    key: _containers.RepeatedScalarFieldContainer[str]
    ttl: _duration_pb2.Duration
    def __init__(self, key: _Optional[_Iterable[str]] = ..., ttl: _Optional[_Union[_duration_pb2.Duration, _Mapping]] = ...) -> None: ...

class AcquireLeaseResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...
