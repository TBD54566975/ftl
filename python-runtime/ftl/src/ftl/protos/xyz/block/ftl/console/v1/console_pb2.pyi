from google.protobuf import duration_pb2 as _duration_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from xyz.block.ftl.schema.v1 import schema_pb2 as _schema_pb2
from xyz.block.ftl.v1 import event_pb2 as _event_pb2
from xyz.block.ftl.v1 import ftl_pb2 as _ftl_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
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

class GetEventsRequest(_message.Message):
    __slots__ = ("filters", "limit", "order")
    class Order(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = ()
        ORDER_UNSPECIFIED: _ClassVar[GetEventsRequest.Order]
        ORDER_ASC: _ClassVar[GetEventsRequest.Order]
        ORDER_DESC: _ClassVar[GetEventsRequest.Order]
    ORDER_UNSPECIFIED: GetEventsRequest.Order
    ORDER_ASC: GetEventsRequest.Order
    ORDER_DESC: GetEventsRequest.Order
    class LimitFilter(_message.Message):
        __slots__ = ("limit",)
        LIMIT_FIELD_NUMBER: _ClassVar[int]
        limit: int
        def __init__(self, limit: _Optional[int] = ...) -> None: ...
    class LogLevelFilter(_message.Message):
        __slots__ = ("log_level",)
        LOG_LEVEL_FIELD_NUMBER: _ClassVar[int]
        log_level: _event_pb2.LogLevel
        def __init__(self, log_level: _Optional[_Union[_event_pb2.LogLevel, str]] = ...) -> None: ...
    class DeploymentFilter(_message.Message):
        __slots__ = ("deployments",)
        DEPLOYMENTS_FIELD_NUMBER: _ClassVar[int]
        deployments: _containers.RepeatedScalarFieldContainer[str]
        def __init__(self, deployments: _Optional[_Iterable[str]] = ...) -> None: ...
    class RequestFilter(_message.Message):
        __slots__ = ("requests",)
        REQUESTS_FIELD_NUMBER: _ClassVar[int]
        requests: _containers.RepeatedScalarFieldContainer[str]
        def __init__(self, requests: _Optional[_Iterable[str]] = ...) -> None: ...
    class EventTypeFilter(_message.Message):
        __slots__ = ("event_types",)
        EVENT_TYPES_FIELD_NUMBER: _ClassVar[int]
        event_types: _containers.RepeatedScalarFieldContainer[_event_pb2.EventType]
        def __init__(self, event_types: _Optional[_Iterable[_Union[_event_pb2.EventType, str]]] = ...) -> None: ...
    class TimeFilter(_message.Message):
        __slots__ = ("older_than", "newer_than")
        OLDER_THAN_FIELD_NUMBER: _ClassVar[int]
        NEWER_THAN_FIELD_NUMBER: _ClassVar[int]
        older_than: _timestamp_pb2.Timestamp
        newer_than: _timestamp_pb2.Timestamp
        def __init__(self, older_than: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., newer_than: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...
    class IDFilter(_message.Message):
        __slots__ = ("lower_than", "higher_than")
        LOWER_THAN_FIELD_NUMBER: _ClassVar[int]
        HIGHER_THAN_FIELD_NUMBER: _ClassVar[int]
        lower_than: int
        higher_than: int
        def __init__(self, lower_than: _Optional[int] = ..., higher_than: _Optional[int] = ...) -> None: ...
    class CallFilter(_message.Message):
        __slots__ = ("dest_module", "dest_verb", "source_module")
        DEST_MODULE_FIELD_NUMBER: _ClassVar[int]
        DEST_VERB_FIELD_NUMBER: _ClassVar[int]
        SOURCE_MODULE_FIELD_NUMBER: _ClassVar[int]
        dest_module: str
        dest_verb: str
        source_module: str
        def __init__(self, dest_module: _Optional[str] = ..., dest_verb: _Optional[str] = ..., source_module: _Optional[str] = ...) -> None: ...
    class ModuleFilter(_message.Message):
        __slots__ = ("module", "verb")
        MODULE_FIELD_NUMBER: _ClassVar[int]
        VERB_FIELD_NUMBER: _ClassVar[int]
        module: str
        verb: str
        def __init__(self, module: _Optional[str] = ..., verb: _Optional[str] = ...) -> None: ...
    class Filter(_message.Message):
        __slots__ = ("limit", "log_level", "deployments", "requests", "event_types", "time", "id", "call", "module")
        LIMIT_FIELD_NUMBER: _ClassVar[int]
        LOG_LEVEL_FIELD_NUMBER: _ClassVar[int]
        DEPLOYMENTS_FIELD_NUMBER: _ClassVar[int]
        REQUESTS_FIELD_NUMBER: _ClassVar[int]
        EVENT_TYPES_FIELD_NUMBER: _ClassVar[int]
        TIME_FIELD_NUMBER: _ClassVar[int]
        ID_FIELD_NUMBER: _ClassVar[int]
        CALL_FIELD_NUMBER: _ClassVar[int]
        MODULE_FIELD_NUMBER: _ClassVar[int]
        limit: GetEventsRequest.LimitFilter
        log_level: GetEventsRequest.LogLevelFilter
        deployments: GetEventsRequest.DeploymentFilter
        requests: GetEventsRequest.RequestFilter
        event_types: GetEventsRequest.EventTypeFilter
        time: GetEventsRequest.TimeFilter
        id: GetEventsRequest.IDFilter
        call: GetEventsRequest.CallFilter
        module: GetEventsRequest.ModuleFilter
        def __init__(self, limit: _Optional[_Union[GetEventsRequest.LimitFilter, _Mapping]] = ..., log_level: _Optional[_Union[GetEventsRequest.LogLevelFilter, _Mapping]] = ..., deployments: _Optional[_Union[GetEventsRequest.DeploymentFilter, _Mapping]] = ..., requests: _Optional[_Union[GetEventsRequest.RequestFilter, _Mapping]] = ..., event_types: _Optional[_Union[GetEventsRequest.EventTypeFilter, _Mapping]] = ..., time: _Optional[_Union[GetEventsRequest.TimeFilter, _Mapping]] = ..., id: _Optional[_Union[GetEventsRequest.IDFilter, _Mapping]] = ..., call: _Optional[_Union[GetEventsRequest.CallFilter, _Mapping]] = ..., module: _Optional[_Union[GetEventsRequest.ModuleFilter, _Mapping]] = ...) -> None: ...
    FILTERS_FIELD_NUMBER: _ClassVar[int]
    LIMIT_FIELD_NUMBER: _ClassVar[int]
    ORDER_FIELD_NUMBER: _ClassVar[int]
    filters: _containers.RepeatedCompositeFieldContainer[GetEventsRequest.Filter]
    limit: int
    order: GetEventsRequest.Order
    def __init__(self, filters: _Optional[_Iterable[_Union[GetEventsRequest.Filter, _Mapping]]] = ..., limit: _Optional[int] = ..., order: _Optional[_Union[GetEventsRequest.Order, str]] = ...) -> None: ...

class GetEventsResponse(_message.Message):
    __slots__ = ("events", "cursor")
    EVENTS_FIELD_NUMBER: _ClassVar[int]
    CURSOR_FIELD_NUMBER: _ClassVar[int]
    events: _containers.RepeatedCompositeFieldContainer[_event_pb2.Event]
    cursor: int
    def __init__(self, events: _Optional[_Iterable[_Union[_event_pb2.Event, _Mapping]]] = ..., cursor: _Optional[int] = ...) -> None: ...

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

class StreamEventsRequest(_message.Message):
    __slots__ = ("update_interval", "query")
    UPDATE_INTERVAL_FIELD_NUMBER: _ClassVar[int]
    QUERY_FIELD_NUMBER: _ClassVar[int]
    update_interval: _duration_pb2.Duration
    query: GetEventsRequest
    def __init__(self, update_interval: _Optional[_Union[_duration_pb2.Duration, _Mapping]] = ..., query: _Optional[_Union[GetEventsRequest, _Mapping]] = ...) -> None: ...

class StreamEventsResponse(_message.Message):
    __slots__ = ("events",)
    EVENTS_FIELD_NUMBER: _ClassVar[int]
    events: _containers.RepeatedCompositeFieldContainer[_event_pb2.Event]
    def __init__(self, events: _Optional[_Iterable[_Union[_event_pb2.Event, _Mapping]]] = ...) -> None: ...
