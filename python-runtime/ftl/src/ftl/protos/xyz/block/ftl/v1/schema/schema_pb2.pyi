from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class AliasKind(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    ALIAS_KIND_JSON: _ClassVar[AliasKind]
ALIAS_KIND_JSON: AliasKind

class Any(_message.Message):
    __slots__ = ("pos",)
    POS_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ...) -> None: ...

class Array(_message.Message):
    __slots__ = ("pos", "element")
    POS_FIELD_NUMBER: _ClassVar[int]
    ELEMENT_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    element: Type
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., element: _Optional[_Union[Type, _Mapping]] = ...) -> None: ...

class Bool(_message.Message):
    __slots__ = ("pos",)
    POS_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ...) -> None: ...

class Bytes(_message.Message):
    __slots__ = ("pos",)
    POS_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ...) -> None: ...

class Config(_message.Message):
    __slots__ = ("pos", "comments", "name", "type")
    POS_FIELD_NUMBER: _ClassVar[int]
    COMMENTS_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    comments: _containers.RepeatedScalarFieldContainer[str]
    name: str
    type: Type
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., comments: _Optional[_Iterable[str]] = ..., name: _Optional[str] = ..., type: _Optional[_Union[Type, _Mapping]] = ...) -> None: ...

class Data(_message.Message):
    __slots__ = ("pos", "comments", "export", "name", "type_parameters", "fields", "metadata")
    POS_FIELD_NUMBER: _ClassVar[int]
    COMMENTS_FIELD_NUMBER: _ClassVar[int]
    EXPORT_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    TYPE_PARAMETERS_FIELD_NUMBER: _ClassVar[int]
    FIELDS_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    comments: _containers.RepeatedScalarFieldContainer[str]
    export: bool
    name: str
    type_parameters: _containers.RepeatedCompositeFieldContainer[TypeParameter]
    fields: _containers.RepeatedCompositeFieldContainer[Field]
    metadata: _containers.RepeatedCompositeFieldContainer[Metadata]
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., comments: _Optional[_Iterable[str]] = ..., export: bool = ..., name: _Optional[str] = ..., type_parameters: _Optional[_Iterable[_Union[TypeParameter, _Mapping]]] = ..., fields: _Optional[_Iterable[_Union[Field, _Mapping]]] = ..., metadata: _Optional[_Iterable[_Union[Metadata, _Mapping]]] = ...) -> None: ...

class Database(_message.Message):
    __slots__ = ("pos", "runtime", "comments", "type", "name")
    POS_FIELD_NUMBER: _ClassVar[int]
    RUNTIME_FIELD_NUMBER: _ClassVar[int]
    COMMENTS_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    runtime: DatabaseRuntime
    comments: _containers.RepeatedScalarFieldContainer[str]
    type: str
    name: str
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., runtime: _Optional[_Union[DatabaseRuntime, _Mapping]] = ..., comments: _Optional[_Iterable[str]] = ..., type: _Optional[str] = ..., name: _Optional[str] = ...) -> None: ...

class DatabaseRuntime(_message.Message):
    __slots__ = ("dsn",)
    DSN_FIELD_NUMBER: _ClassVar[int]
    dsn: str
    def __init__(self, dsn: _Optional[str] = ...) -> None: ...

class Decl(_message.Message):
    __slots__ = ("config", "data", "database", "enum", "secret", "subscription", "topic", "type_alias", "verb")
    CONFIG_FIELD_NUMBER: _ClassVar[int]
    DATA_FIELD_NUMBER: _ClassVar[int]
    DATABASE_FIELD_NUMBER: _ClassVar[int]
    ENUM_FIELD_NUMBER: _ClassVar[int]
    SECRET_FIELD_NUMBER: _ClassVar[int]
    SUBSCRIPTION_FIELD_NUMBER: _ClassVar[int]
    TOPIC_FIELD_NUMBER: _ClassVar[int]
    TYPE_ALIAS_FIELD_NUMBER: _ClassVar[int]
    VERB_FIELD_NUMBER: _ClassVar[int]
    config: Config
    data: Data
    database: Database
    enum: Enum
    secret: Secret
    subscription: Subscription
    topic: Topic
    type_alias: TypeAlias
    verb: Verb
    def __init__(self, config: _Optional[_Union[Config, _Mapping]] = ..., data: _Optional[_Union[Data, _Mapping]] = ..., database: _Optional[_Union[Database, _Mapping]] = ..., enum: _Optional[_Union[Enum, _Mapping]] = ..., secret: _Optional[_Union[Secret, _Mapping]] = ..., subscription: _Optional[_Union[Subscription, _Mapping]] = ..., topic: _Optional[_Union[Topic, _Mapping]] = ..., type_alias: _Optional[_Union[TypeAlias, _Mapping]] = ..., verb: _Optional[_Union[Verb, _Mapping]] = ...) -> None: ...

class Enum(_message.Message):
    __slots__ = ("pos", "comments", "export", "name", "type", "variants")
    POS_FIELD_NUMBER: _ClassVar[int]
    COMMENTS_FIELD_NUMBER: _ClassVar[int]
    EXPORT_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    VARIANTS_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    comments: _containers.RepeatedScalarFieldContainer[str]
    export: bool
    name: str
    type: Type
    variants: _containers.RepeatedCompositeFieldContainer[EnumVariant]
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., comments: _Optional[_Iterable[str]] = ..., export: bool = ..., name: _Optional[str] = ..., type: _Optional[_Union[Type, _Mapping]] = ..., variants: _Optional[_Iterable[_Union[EnumVariant, _Mapping]]] = ...) -> None: ...

class EnumVariant(_message.Message):
    __slots__ = ("pos", "comments", "name", "value")
    POS_FIELD_NUMBER: _ClassVar[int]
    COMMENTS_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    comments: _containers.RepeatedScalarFieldContainer[str]
    name: str
    value: Value
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., comments: _Optional[_Iterable[str]] = ..., name: _Optional[str] = ..., value: _Optional[_Union[Value, _Mapping]] = ...) -> None: ...

class Field(_message.Message):
    __slots__ = ("pos", "comments", "name", "type", "metadata")
    POS_FIELD_NUMBER: _ClassVar[int]
    COMMENTS_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    comments: _containers.RepeatedScalarFieldContainer[str]
    name: str
    type: Type
    metadata: _containers.RepeatedCompositeFieldContainer[Metadata]
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., comments: _Optional[_Iterable[str]] = ..., name: _Optional[str] = ..., type: _Optional[_Union[Type, _Mapping]] = ..., metadata: _Optional[_Iterable[_Union[Metadata, _Mapping]]] = ...) -> None: ...

class Float(_message.Message):
    __slots__ = ("pos",)
    POS_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ...) -> None: ...

class IngressPathComponent(_message.Message):
    __slots__ = ("ingress_path_literal", "ingress_path_parameter")
    INGRESS_PATH_LITERAL_FIELD_NUMBER: _ClassVar[int]
    INGRESS_PATH_PARAMETER_FIELD_NUMBER: _ClassVar[int]
    ingress_path_literal: IngressPathLiteral
    ingress_path_parameter: IngressPathParameter
    def __init__(self, ingress_path_literal: _Optional[_Union[IngressPathLiteral, _Mapping]] = ..., ingress_path_parameter: _Optional[_Union[IngressPathParameter, _Mapping]] = ...) -> None: ...

class IngressPathLiteral(_message.Message):
    __slots__ = ("pos", "text")
    POS_FIELD_NUMBER: _ClassVar[int]
    TEXT_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    text: str
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., text: _Optional[str] = ...) -> None: ...

class IngressPathParameter(_message.Message):
    __slots__ = ("pos", "name")
    POS_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    name: str
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., name: _Optional[str] = ...) -> None: ...

class Int(_message.Message):
    __slots__ = ("pos",)
    POS_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ...) -> None: ...

class IntValue(_message.Message):
    __slots__ = ("pos", "value")
    POS_FIELD_NUMBER: _ClassVar[int]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    value: int
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., value: _Optional[int] = ...) -> None: ...

class Map(_message.Message):
    __slots__ = ("pos", "key", "value")
    POS_FIELD_NUMBER: _ClassVar[int]
    KEY_FIELD_NUMBER: _ClassVar[int]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    key: Type
    value: Type
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., key: _Optional[_Union[Type, _Mapping]] = ..., value: _Optional[_Union[Type, _Mapping]] = ...) -> None: ...

class Metadata(_message.Message):
    __slots__ = ("alias", "calls", "config", "cron_job", "databases", "encoding", "ingress", "retry", "secrets", "subscriber", "type_map")
    ALIAS_FIELD_NUMBER: _ClassVar[int]
    CALLS_FIELD_NUMBER: _ClassVar[int]
    CONFIG_FIELD_NUMBER: _ClassVar[int]
    CRON_JOB_FIELD_NUMBER: _ClassVar[int]
    DATABASES_FIELD_NUMBER: _ClassVar[int]
    ENCODING_FIELD_NUMBER: _ClassVar[int]
    INGRESS_FIELD_NUMBER: _ClassVar[int]
    RETRY_FIELD_NUMBER: _ClassVar[int]
    SECRETS_FIELD_NUMBER: _ClassVar[int]
    SUBSCRIBER_FIELD_NUMBER: _ClassVar[int]
    TYPE_MAP_FIELD_NUMBER: _ClassVar[int]
    alias: MetadataAlias
    calls: MetadataCalls
    config: MetadataConfig
    cron_job: MetadataCronJob
    databases: MetadataDatabases
    encoding: MetadataEncoding
    ingress: MetadataIngress
    retry: MetadataRetry
    secrets: MetadataSecrets
    subscriber: MetadataSubscriber
    type_map: MetadataTypeMap
    def __init__(self, alias: _Optional[_Union[MetadataAlias, _Mapping]] = ..., calls: _Optional[_Union[MetadataCalls, _Mapping]] = ..., config: _Optional[_Union[MetadataConfig, _Mapping]] = ..., cron_job: _Optional[_Union[MetadataCronJob, _Mapping]] = ..., databases: _Optional[_Union[MetadataDatabases, _Mapping]] = ..., encoding: _Optional[_Union[MetadataEncoding, _Mapping]] = ..., ingress: _Optional[_Union[MetadataIngress, _Mapping]] = ..., retry: _Optional[_Union[MetadataRetry, _Mapping]] = ..., secrets: _Optional[_Union[MetadataSecrets, _Mapping]] = ..., subscriber: _Optional[_Union[MetadataSubscriber, _Mapping]] = ..., type_map: _Optional[_Union[MetadataTypeMap, _Mapping]] = ...) -> None: ...

class MetadataAlias(_message.Message):
    __slots__ = ("pos", "kind", "alias")
    POS_FIELD_NUMBER: _ClassVar[int]
    KIND_FIELD_NUMBER: _ClassVar[int]
    ALIAS_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    kind: AliasKind
    alias: str
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., kind: _Optional[_Union[AliasKind, str]] = ..., alias: _Optional[str] = ...) -> None: ...

class MetadataCalls(_message.Message):
    __slots__ = ("pos", "calls")
    POS_FIELD_NUMBER: _ClassVar[int]
    CALLS_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    calls: _containers.RepeatedCompositeFieldContainer[Ref]
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., calls: _Optional[_Iterable[_Union[Ref, _Mapping]]] = ...) -> None: ...

class MetadataConfig(_message.Message):
    __slots__ = ("pos", "config")
    POS_FIELD_NUMBER: _ClassVar[int]
    CONFIG_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    config: _containers.RepeatedCompositeFieldContainer[Ref]
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., config: _Optional[_Iterable[_Union[Ref, _Mapping]]] = ...) -> None: ...

class MetadataCronJob(_message.Message):
    __slots__ = ("pos", "cron")
    POS_FIELD_NUMBER: _ClassVar[int]
    CRON_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    cron: str
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., cron: _Optional[str] = ...) -> None: ...

class MetadataDatabases(_message.Message):
    __slots__ = ("pos", "calls")
    POS_FIELD_NUMBER: _ClassVar[int]
    CALLS_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    calls: _containers.RepeatedCompositeFieldContainer[Ref]
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., calls: _Optional[_Iterable[_Union[Ref, _Mapping]]] = ...) -> None: ...

class MetadataEncoding(_message.Message):
    __slots__ = ("pos", "type", "lenient")
    POS_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    LENIENT_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    type: str
    lenient: bool
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., type: _Optional[str] = ..., lenient: bool = ...) -> None: ...

class MetadataIngress(_message.Message):
    __slots__ = ("pos", "type", "method", "path")
    POS_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    METHOD_FIELD_NUMBER: _ClassVar[int]
    PATH_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    type: str
    method: str
    path: _containers.RepeatedCompositeFieldContainer[IngressPathComponent]
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., type: _Optional[str] = ..., method: _Optional[str] = ..., path: _Optional[_Iterable[_Union[IngressPathComponent, _Mapping]]] = ...) -> None: ...

class MetadataRetry(_message.Message):
    __slots__ = ("pos", "count", "min_backoff", "max_backoff", "catch")
    POS_FIELD_NUMBER: _ClassVar[int]
    COUNT_FIELD_NUMBER: _ClassVar[int]
    MIN_BACKOFF_FIELD_NUMBER: _ClassVar[int]
    MAX_BACKOFF_FIELD_NUMBER: _ClassVar[int]
    CATCH_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    count: int
    min_backoff: str
    max_backoff: str
    catch: Ref
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., count: _Optional[int] = ..., min_backoff: _Optional[str] = ..., max_backoff: _Optional[str] = ..., catch: _Optional[_Union[Ref, _Mapping]] = ...) -> None: ...

class MetadataSecrets(_message.Message):
    __slots__ = ("pos", "secrets")
    POS_FIELD_NUMBER: _ClassVar[int]
    SECRETS_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    secrets: _containers.RepeatedCompositeFieldContainer[Ref]
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., secrets: _Optional[_Iterable[_Union[Ref, _Mapping]]] = ...) -> None: ...

class MetadataSubscriber(_message.Message):
    __slots__ = ("pos", "name")
    POS_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    name: str
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., name: _Optional[str] = ...) -> None: ...

class MetadataTypeMap(_message.Message):
    __slots__ = ("pos", "runtime", "native_name")
    POS_FIELD_NUMBER: _ClassVar[int]
    RUNTIME_FIELD_NUMBER: _ClassVar[int]
    NATIVE_NAME_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    runtime: str
    native_name: str
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., runtime: _Optional[str] = ..., native_name: _Optional[str] = ...) -> None: ...

class Module(_message.Message):
    __slots__ = ("pos", "comments", "builtin", "name", "decls", "runtime")
    POS_FIELD_NUMBER: _ClassVar[int]
    COMMENTS_FIELD_NUMBER: _ClassVar[int]
    BUILTIN_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    DECLS_FIELD_NUMBER: _ClassVar[int]
    RUNTIME_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    comments: _containers.RepeatedScalarFieldContainer[str]
    builtin: bool
    name: str
    decls: _containers.RepeatedCompositeFieldContainer[Decl]
    runtime: ModuleRuntime
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., comments: _Optional[_Iterable[str]] = ..., builtin: bool = ..., name: _Optional[str] = ..., decls: _Optional[_Iterable[_Union[Decl, _Mapping]]] = ..., runtime: _Optional[_Union[ModuleRuntime, _Mapping]] = ...) -> None: ...

class ModuleRuntime(_message.Message):
    __slots__ = ("create_time", "language", "min_replicas", "os", "arch")
    CREATE_TIME_FIELD_NUMBER: _ClassVar[int]
    LANGUAGE_FIELD_NUMBER: _ClassVar[int]
    MIN_REPLICAS_FIELD_NUMBER: _ClassVar[int]
    OS_FIELD_NUMBER: _ClassVar[int]
    ARCH_FIELD_NUMBER: _ClassVar[int]
    create_time: _timestamp_pb2.Timestamp
    language: str
    min_replicas: int
    os: str
    arch: str
    def __init__(self, create_time: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., language: _Optional[str] = ..., min_replicas: _Optional[int] = ..., os: _Optional[str] = ..., arch: _Optional[str] = ...) -> None: ...

class Optional(_message.Message):
    __slots__ = ("pos", "type")
    POS_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    type: Type
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., type: _Optional[_Union[Type, _Mapping]] = ...) -> None: ...

class Position(_message.Message):
    __slots__ = ("filename", "line", "column")
    FILENAME_FIELD_NUMBER: _ClassVar[int]
    LINE_FIELD_NUMBER: _ClassVar[int]
    COLUMN_FIELD_NUMBER: _ClassVar[int]
    filename: str
    line: int
    column: int
    def __init__(self, filename: _Optional[str] = ..., line: _Optional[int] = ..., column: _Optional[int] = ...) -> None: ...

class Ref(_message.Message):
    __slots__ = ("pos", "module", "name", "type_parameters")
    POS_FIELD_NUMBER: _ClassVar[int]
    MODULE_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    TYPE_PARAMETERS_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    module: str
    name: str
    type_parameters: _containers.RepeatedCompositeFieldContainer[Type]
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., module: _Optional[str] = ..., name: _Optional[str] = ..., type_parameters: _Optional[_Iterable[_Union[Type, _Mapping]]] = ...) -> None: ...

class Schema(_message.Message):
    __slots__ = ("pos", "modules")
    POS_FIELD_NUMBER: _ClassVar[int]
    MODULES_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    modules: _containers.RepeatedCompositeFieldContainer[Module]
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., modules: _Optional[_Iterable[_Union[Module, _Mapping]]] = ...) -> None: ...

class Secret(_message.Message):
    __slots__ = ("pos", "comments", "name", "type")
    POS_FIELD_NUMBER: _ClassVar[int]
    COMMENTS_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    comments: _containers.RepeatedScalarFieldContainer[str]
    name: str
    type: Type
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., comments: _Optional[_Iterable[str]] = ..., name: _Optional[str] = ..., type: _Optional[_Union[Type, _Mapping]] = ...) -> None: ...

class String(_message.Message):
    __slots__ = ("pos",)
    POS_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ...) -> None: ...

class StringValue(_message.Message):
    __slots__ = ("pos", "value")
    POS_FIELD_NUMBER: _ClassVar[int]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    value: str
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., value: _Optional[str] = ...) -> None: ...

class Subscription(_message.Message):
    __slots__ = ("pos", "comments", "name", "topic")
    POS_FIELD_NUMBER: _ClassVar[int]
    COMMENTS_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    TOPIC_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    comments: _containers.RepeatedScalarFieldContainer[str]
    name: str
    topic: Ref
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., comments: _Optional[_Iterable[str]] = ..., name: _Optional[str] = ..., topic: _Optional[_Union[Ref, _Mapping]] = ...) -> None: ...

class Time(_message.Message):
    __slots__ = ("pos",)
    POS_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ...) -> None: ...

class Topic(_message.Message):
    __slots__ = ("pos", "comments", "export", "name", "event")
    POS_FIELD_NUMBER: _ClassVar[int]
    COMMENTS_FIELD_NUMBER: _ClassVar[int]
    EXPORT_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    EVENT_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    comments: _containers.RepeatedScalarFieldContainer[str]
    export: bool
    name: str
    event: Type
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., comments: _Optional[_Iterable[str]] = ..., export: bool = ..., name: _Optional[str] = ..., event: _Optional[_Union[Type, _Mapping]] = ...) -> None: ...

class Type(_message.Message):
    __slots__ = ("any", "array", "bool", "bytes", "float", "int", "map", "optional", "ref", "string", "time", "unit")
    ANY_FIELD_NUMBER: _ClassVar[int]
    ARRAY_FIELD_NUMBER: _ClassVar[int]
    BOOL_FIELD_NUMBER: _ClassVar[int]
    BYTES_FIELD_NUMBER: _ClassVar[int]
    FLOAT_FIELD_NUMBER: _ClassVar[int]
    INT_FIELD_NUMBER: _ClassVar[int]
    MAP_FIELD_NUMBER: _ClassVar[int]
    OPTIONAL_FIELD_NUMBER: _ClassVar[int]
    REF_FIELD_NUMBER: _ClassVar[int]
    STRING_FIELD_NUMBER: _ClassVar[int]
    TIME_FIELD_NUMBER: _ClassVar[int]
    UNIT_FIELD_NUMBER: _ClassVar[int]
    any: Any
    array: Array
    bool: Bool
    bytes: Bytes
    float: Float
    int: Int
    map: Map
    optional: Optional
    ref: Ref
    string: String
    time: Time
    unit: Unit
    def __init__(self, any: _Optional[_Union[Any, _Mapping]] = ..., array: _Optional[_Union[Array, _Mapping]] = ..., bool: _Optional[_Union[Bool, _Mapping]] = ..., bytes: _Optional[_Union[Bytes, _Mapping]] = ..., float: _Optional[_Union[Float, _Mapping]] = ..., int: _Optional[_Union[Int, _Mapping]] = ..., map: _Optional[_Union[Map, _Mapping]] = ..., optional: _Optional[_Union[Optional, _Mapping]] = ..., ref: _Optional[_Union[Ref, _Mapping]] = ..., string: _Optional[_Union[String, _Mapping]] = ..., time: _Optional[_Union[Time, _Mapping]] = ..., unit: _Optional[_Union[Unit, _Mapping]] = ...) -> None: ...

class TypeAlias(_message.Message):
    __slots__ = ("pos", "comments", "export", "name", "type", "metadata")
    POS_FIELD_NUMBER: _ClassVar[int]
    COMMENTS_FIELD_NUMBER: _ClassVar[int]
    EXPORT_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    comments: _containers.RepeatedScalarFieldContainer[str]
    export: bool
    name: str
    type: Type
    metadata: _containers.RepeatedCompositeFieldContainer[Metadata]
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., comments: _Optional[_Iterable[str]] = ..., export: bool = ..., name: _Optional[str] = ..., type: _Optional[_Union[Type, _Mapping]] = ..., metadata: _Optional[_Iterable[_Union[Metadata, _Mapping]]] = ...) -> None: ...

class TypeParameter(_message.Message):
    __slots__ = ("pos", "name")
    POS_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    name: str
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., name: _Optional[str] = ...) -> None: ...

class TypeValue(_message.Message):
    __slots__ = ("pos", "value")
    POS_FIELD_NUMBER: _ClassVar[int]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    value: Type
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., value: _Optional[_Union[Type, _Mapping]] = ...) -> None: ...

class Unit(_message.Message):
    __slots__ = ("pos",)
    POS_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ...) -> None: ...

class Value(_message.Message):
    __slots__ = ("int_value", "string_value", "type_value")
    INT_VALUE_FIELD_NUMBER: _ClassVar[int]
    STRING_VALUE_FIELD_NUMBER: _ClassVar[int]
    TYPE_VALUE_FIELD_NUMBER: _ClassVar[int]
    int_value: IntValue
    string_value: StringValue
    type_value: TypeValue
    def __init__(self, int_value: _Optional[_Union[IntValue, _Mapping]] = ..., string_value: _Optional[_Union[StringValue, _Mapping]] = ..., type_value: _Optional[_Union[TypeValue, _Mapping]] = ...) -> None: ...

class Verb(_message.Message):
    __slots__ = ("pos", "comments", "export", "name", "request", "response", "metadata", "runtime")
    POS_FIELD_NUMBER: _ClassVar[int]
    COMMENTS_FIELD_NUMBER: _ClassVar[int]
    EXPORT_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    REQUEST_FIELD_NUMBER: _ClassVar[int]
    RESPONSE_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    RUNTIME_FIELD_NUMBER: _ClassVar[int]
    pos: Position
    comments: _containers.RepeatedScalarFieldContainer[str]
    export: bool
    name: str
    request: Type
    response: Type
    metadata: _containers.RepeatedCompositeFieldContainer[Metadata]
    runtime: VerbRuntime
    def __init__(self, pos: _Optional[_Union[Position, _Mapping]] = ..., comments: _Optional[_Iterable[str]] = ..., export: bool = ..., name: _Optional[str] = ..., request: _Optional[_Union[Type, _Mapping]] = ..., response: _Optional[_Union[Type, _Mapping]] = ..., metadata: _Optional[_Iterable[_Union[Metadata, _Mapping]]] = ..., runtime: _Optional[_Union[VerbRuntime, _Mapping]] = ...) -> None: ...

class VerbRuntime(_message.Message):
    __slots__ = ("create_time", "start_time")
    CREATE_TIME_FIELD_NUMBER: _ClassVar[int]
    START_TIME_FIELD_NUMBER: _ClassVar[int]
    create_time: _timestamp_pb2.Timestamp
    start_time: _timestamp_pb2.Timestamp
    def __init__(self, create_time: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., start_time: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...
