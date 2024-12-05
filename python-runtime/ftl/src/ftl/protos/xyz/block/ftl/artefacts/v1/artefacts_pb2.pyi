from xyz.block.ftl.v1 import ftl_pb2 as _ftl_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class UploadArtefactRequest(_message.Message):
    __slots__ = ("content",)
    CONTENT_FIELD_NUMBER: _ClassVar[int]
    content: bytes
    def __init__(self, content: _Optional[bytes] = ...) -> None: ...

class UploadArtefactResponse(_message.Message):
    __slots__ = ("digest",)
    DIGEST_FIELD_NUMBER: _ClassVar[int]
    digest: bytes
    def __init__(self, digest: _Optional[bytes] = ...) -> None: ...

class DeploymentArtefact(_message.Message):
    __slots__ = ("digest", "path", "executable")
    DIGEST_FIELD_NUMBER: _ClassVar[int]
    PATH_FIELD_NUMBER: _ClassVar[int]
    EXECUTABLE_FIELD_NUMBER: _ClassVar[int]
    digest: str
    path: str
    executable: bool
    def __init__(self, digest: _Optional[str] = ..., path: _Optional[str] = ..., executable: bool = ...) -> None: ...

class GetArtefactDiffsRequest(_message.Message):
    __slots__ = ("client_digests",)
    CLIENT_DIGESTS_FIELD_NUMBER: _ClassVar[int]
    client_digests: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, client_digests: _Optional[_Iterable[str]] = ...) -> None: ...

class GetArtefactDiffsResponse(_message.Message):
    __slots__ = ("missing_digests", "client_artefacts")
    MISSING_DIGESTS_FIELD_NUMBER: _ClassVar[int]
    CLIENT_ARTEFACTS_FIELD_NUMBER: _ClassVar[int]
    missing_digests: _containers.RepeatedScalarFieldContainer[str]
    client_artefacts: _containers.RepeatedCompositeFieldContainer[DeploymentArtefact]
    def __init__(self, missing_digests: _Optional[_Iterable[str]] = ..., client_artefacts: _Optional[_Iterable[_Union[DeploymentArtefact, _Mapping]]] = ...) -> None: ...
