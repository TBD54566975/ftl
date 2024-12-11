from xyz.block.ftl.schema.v1 import schema_pb2 as _schema_pb2
from xyz.block.ftl.timeline.v1 import timeline_pb2 as _timeline_pb2
from xyz.block.ftl.v1 import controller_pb2 as _controller_pb2
from xyz.block.ftl.v1 import ftl_pb2 as _ftl_pb2
from xyz.block.ftl.v1 import verb_pb2 as _verb_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Config(_message.Message):
    __slots__ = ("config", "references")
    CONFIG_FIELD_NUMBER: _ClassVar[int]
    REFERENCES_FIELD_NUMBER: _ClassVar[int]
    config: _schema_pb2.Config
    references: _containers.RepeatedCompositeFieldContainer[_schema_pb2.Ref]
    def __init__(self, config: _Optional[_Union[_schema_pb2.Config, _Mapping]] = ..., references: _Optional[_Iterable[_Union[_schema_pb2.Ref, _Mapping]]] = ...) -> None: ...

class Data(_message.Message):
    __slots__ = ("data", "schema", "references")
    DATA_FIELD_NUMBER: _ClassVar[int]
    SCHEMA_FIELD_NUMBER: _ClassVar[int]
    REFERENCES_FIELD_NUMBER: _ClassVar[int]
    data: _schema_pb2.Data
    schema: str
    references: _containers.RepeatedCompositeFieldContainer[_schema_pb2.Ref]
    def __init__(self, data: _Optional[_Union[_schema_pb2.Data, _Mapping]] = ..., schema: _Optional[str] = ..., references: _Optional[_Iterable[_Union[_schema_pb2.Ref, _Mapping]]] = ...) -> None: ...

class Database(_message.Message):
    __slots__ = ("database", "references")
    DATABASE_FIELD_NUMBER: _ClassVar[int]
    REFERENCES_FIELD_NUMBER: _ClassVar[int]
    database: _schema_pb2.Database
    references: _containers.RepeatedCompositeFieldContainer[_schema_pb2.Ref]
    def __init__(self, database: _Optional[_Union[_schema_pb2.Database, _Mapping]] = ..., references: _Optional[_Iterable[_Union[_schema_pb2.Ref, _Mapping]]] = ...) -> None: ...

class Enum(_message.Message):
    __slots__ = ("enum", "references")
    ENUM_FIELD_NUMBER: _ClassVar[int]
    REFERENCES_FIELD_NUMBER: _ClassVar[int]
    enum: _schema_pb2.Enum
    references: _containers.RepeatedCompositeFieldContainer[_schema_pb2.Ref]
    def __init__(self, enum: _Optional[_Union[_schema_pb2.Enum, _Mapping]] = ..., references: _Optional[_Iterable[_Union[_schema_pb2.Ref, _Mapping]]] = ...) -> None: ...

class Topic(_message.Message):
    __slots__ = ("topic", "references")
    TOPIC_FIELD_NUMBER: _ClassVar[int]
    REFERENCES_FIELD_NUMBER: _ClassVar[int]
    topic: _schema_pb2.Topic
    references: _containers.RepeatedCompositeFieldContainer[_schema_pb2.Ref]
    def __init__(self, topic: _Optional[_Union[_schema_pb2.Topic, _Mapping]] = ..., references: _Optional[_Iterable[_Union[_schema_pb2.Ref, _Mapping]]] = ...) -> None: ...

class TypeAlias(_message.Message):
    __slots__ = ("typealias", "references")
    TYPEALIAS_FIELD_NUMBER: _ClassVar[int]
    REFERENCES_FIELD_NUMBER: _ClassVar[int]
    typealias: _schema_pb2.TypeAlias
    references: _containers.RepeatedCompositeFieldContainer[_schema_pb2.Ref]
    def __init__(self, typealias: _Optional[_Union[_schema_pb2.TypeAlias, _Mapping]] = ..., references: _Optional[_Iterable[_Union[_schema_pb2.Ref, _Mapping]]] = ...) -> None: ...

class Secret(_message.Message):
    __slots__ = ("secret", "references")
    SECRET_FIELD_NUMBER: _ClassVar[int]
    REFERENCES_FIELD_NUMBER: _ClassVar[int]
    secret: _schema_pb2.Secret
    references: _containers.RepeatedCompositeFieldContainer[_schema_pb2.Ref]
    def __init__(self, secret: _Optional[_Union[_schema_pb2.Secret, _Mapping]] = ..., references: _Optional[_Iterable[_Union[_schema_pb2.Ref, _Mapping]]] = ...) -> None: ...

class Verb(_message.Message):
    __slots__ = ("verb", "schema", "json_request_schema", "references")
    VERB_FIELD_NUMBER: _ClassVar[int]
    SCHEMA_FIELD_NUMBER: _ClassVar[int]
    JSON_REQUEST_SCHEMA_FIELD_NUMBER: _ClassVar[int]
    REFERENCES_FIELD_NUMBER: _ClassVar[int]
    verb: _schema_pb2.Verb
    schema: str
    json_request_schema: str
    references: _containers.RepeatedCompositeFieldContainer[_schema_pb2.Ref]
    def __init__(self, verb: _Optional[_Union[_schema_pb2.Verb, _Mapping]] = ..., schema: _Optional[str] = ..., json_request_schema: _Optional[str] = ..., references: _Optional[_Iterable[_Union[_schema_pb2.Ref, _Mapping]]] = ...) -> None: ...

class Module(_message.Message):
    __slots__ = ("name", "deployment_key", "language", "schema", "verbs", "data", "secrets", "configs", "databases", "enums", "topics", "typealiases")
    NAME_FIELD_NUMBER: _ClassVar[int]
    DEPLOYMENT_KEY_FIELD_NUMBER: _ClassVar[int]
    LANGUAGE_FIELD_NUMBER: _ClassVar[int]
    SCHEMA_FIELD_NUMBER: _ClassVar[int]
    VERBS_FIELD_NUMBER: _ClassVar[int]
    DATA_FIELD_NUMBER: _ClassVar[int]
    SECRETS_FIELD_NUMBER: _ClassVar[int]
    CONFIGS_FIELD_NUMBER: _ClassVar[int]
    DATABASES_FIELD_NUMBER: _ClassVar[int]
    ENUMS_FIELD_NUMBER: _ClassVar[int]
    TOPICS_FIELD_NUMBER: _ClassVar[int]
    TYPEALIASES_FIELD_NUMBER: _ClassVar[int]
    name: str
    deployment_key: str
    language: str
    schema: str
    verbs: _containers.RepeatedCompositeFieldContainer[Verb]
    data: _containers.RepeatedCompositeFieldContainer[Data]
    secrets: _containers.RepeatedCompositeFieldContainer[Secret]
    configs: _containers.RepeatedCompositeFieldContainer[Config]
    databases: _containers.RepeatedCompositeFieldContainer[Database]
    enums: _containers.RepeatedCompositeFieldContainer[Enum]
    topics: _containers.RepeatedCompositeFieldContainer[Topic]
    typealiases: _containers.RepeatedCompositeFieldContainer[TypeAlias]
    def __init__(self, name: _Optional[str] = ..., deployment_key: _Optional[str] = ..., language: _Optional[str] = ..., schema: _Optional[str] = ..., verbs: _Optional[_Iterable[_Union[Verb, _Mapping]]] = ..., data: _Optional[_Iterable[_Union[Data, _Mapping]]] = ..., secrets: _Optional[_Iterable[_Union[Secret, _Mapping]]] = ..., configs: _Optional[_Iterable[_Union[Config, _Mapping]]] = ..., databases: _Optional[_Iterable[_Union[Database, _Mapping]]] = ..., enums: _Optional[_Iterable[_Union[Enum, _Mapping]]] = ..., topics: _Optional[_Iterable[_Union[Topic, _Mapping]]] = ..., typealiases: _Optional[_Iterable[_Union[TypeAlias, _Mapping]]] = ...) -> None: ...

class TopologyGroup(_message.Message):
    __slots__ = ("modules",)
    MODULES_FIELD_NUMBER: _ClassVar[int]
    modules: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, modules: _Optional[_Iterable[str]] = ...) -> None: ...

class Topology(_message.Message):
    __slots__ = ("levels",)
    LEVELS_FIELD_NUMBER: _ClassVar[int]
    levels: _containers.RepeatedCompositeFieldContainer[TopologyGroup]
    def __init__(self, levels: _Optional[_Iterable[_Union[TopologyGroup, _Mapping]]] = ...) -> None: ...

class GetModulesRequest(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class GetModulesResponse(_message.Message):
    __slots__ = ("modules", "topology")
    MODULES_FIELD_NUMBER: _ClassVar[int]
    TOPOLOGY_FIELD_NUMBER: _ClassVar[int]
    modules: _containers.RepeatedCompositeFieldContainer[Module]
    topology: Topology
    def __init__(self, modules: _Optional[_Iterable[_Union[Module, _Mapping]]] = ..., topology: _Optional[_Union[Topology, _Mapping]] = ...) -> None: ...

class StreamModulesRequest(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class StreamModulesResponse(_message.Message):
    __slots__ = ("modules", "topology")
    MODULES_FIELD_NUMBER: _ClassVar[int]
    TOPOLOGY_FIELD_NUMBER: _ClassVar[int]
    modules: _containers.RepeatedCompositeFieldContainer[Module]
    topology: Topology
    def __init__(self, modules: _Optional[_Iterable[_Union[Module, _Mapping]]] = ..., topology: _Optional[_Union[Topology, _Mapping]] = ...) -> None: ...

class GetConfigRequest(_message.Message):
    __slots__ = ("name", "module")
    NAME_FIELD_NUMBER: _ClassVar[int]
    MODULE_FIELD_NUMBER: _ClassVar[int]
    name: str
    module: str
    def __init__(self, name: _Optional[str] = ..., module: _Optional[str] = ...) -> None: ...

class GetConfigResponse(_message.Message):
    __slots__ = ("value",)
    VALUE_FIELD_NUMBER: _ClassVar[int]
    value: bytes
    def __init__(self, value: _Optional[bytes] = ...) -> None: ...

class SetConfigRequest(_message.Message):
    __slots__ = ("name", "module", "value")
    NAME_FIELD_NUMBER: _ClassVar[int]
    MODULE_FIELD_NUMBER: _ClassVar[int]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    name: str
    module: str
    value: bytes
    def __init__(self, name: _Optional[str] = ..., module: _Optional[str] = ..., value: _Optional[bytes] = ...) -> None: ...

class SetConfigResponse(_message.Message):
    __slots__ = ("value",)
    VALUE_FIELD_NUMBER: _ClassVar[int]
    value: bytes
    def __init__(self, value: _Optional[bytes] = ...) -> None: ...

class GetSecretRequest(_message.Message):
    __slots__ = ("name", "module")
    NAME_FIELD_NUMBER: _ClassVar[int]
    MODULE_FIELD_NUMBER: _ClassVar[int]
    name: str
    module: str
    def __init__(self, name: _Optional[str] = ..., module: _Optional[str] = ...) -> None: ...

class GetSecretResponse(_message.Message):
    __slots__ = ("value",)
    VALUE_FIELD_NUMBER: _ClassVar[int]
    value: bytes
    def __init__(self, value: _Optional[bytes] = ...) -> None: ...

class SetSecretRequest(_message.Message):
    __slots__ = ("name", "module", "value")
    NAME_FIELD_NUMBER: _ClassVar[int]
    MODULE_FIELD_NUMBER: _ClassVar[int]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    name: str
    module: str
    value: bytes
    def __init__(self, name: _Optional[str] = ..., module: _Optional[str] = ..., value: _Optional[bytes] = ...) -> None: ...

class SetSecretResponse(_message.Message):
    __slots__ = ("value",)
    VALUE_FIELD_NUMBER: _ClassVar[int]
    value: bytes
    def __init__(self, value: _Optional[bytes] = ...) -> None: ...
