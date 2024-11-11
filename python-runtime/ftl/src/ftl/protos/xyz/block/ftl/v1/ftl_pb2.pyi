from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class PingRequest(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class PingResponse(_message.Message):
    __slots__ = ("not_ready",)
    NOT_READY_FIELD_NUMBER: _ClassVar[int]
    not_ready: str
    def __init__(self, not_ready: _Optional[str] = ...) -> None: ...

class Metadata(_message.Message):
    __slots__ = ("values",)
    class Pair(_message.Message):
        __slots__ = ("key", "value")
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    VALUES_FIELD_NUMBER: _ClassVar[int]
    values: _containers.RepeatedCompositeFieldContainer[Metadata.Pair]
    def __init__(self, values: _Optional[_Iterable[_Union[Metadata.Pair, _Mapping]]] = ...) -> None: ...
