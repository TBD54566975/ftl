from google.protobuf import duration_pb2 as _duration_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from xyz.block.ftl.v1 import ftl_pb2 as _ftl_pb2
from xyz.block.ftl.v1.schema import schema_pb2 as _schema_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class EventType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    EVENT_TYPE_UNKNOWN: _ClassVar[EventType]
    EVENT_TYPE_LOG: _ClassVar[EventType]
    EVENT_TYPE_CALL: _ClassVar[EventType]
    EVENT_TYPE_DEPLOYMENT_CREATED: _ClassVar[EventType]
    EVENT_TYPE_DEPLOYMENT_UPDATED: _ClassVar[EventType]
    EVENT_TYPE_INGRESS: _ClassVar[EventType]
    EVENT_TYPE_CRON_SCHEDULED: _ClassVar[EventType]
    EVENT_TYPE_ASYNC_EXECUTE: _ClassVar[EventType]
    EVENT_TYPE_PUBSUB_PUBLISH: _ClassVar[EventType]
    EVENT_TYPE_PUBSUB_CONSUME: _ClassVar[EventType]

class AsyncExecuteEventType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    ASYNC_EXECUTE_EVENT_TYPE_UNKNOWN: _ClassVar[AsyncExecuteEventType]
    ASYNC_EXECUTE_EVENT_TYPE_CRON: _ClassVar[AsyncExecuteEventType]
    ASYNC_EXECUTE_EVENT_TYPE_PUBSUB: _ClassVar[AsyncExecuteEventType]

class LogLevel(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    LOG_LEVEL_UNKNOWN: _ClassVar[LogLevel]
    LOG_LEVEL_TRACE: _ClassVar[LogLevel]
    LOG_LEVEL_DEBUG: _ClassVar[LogLevel]
    LOG_LEVEL_INFO: _ClassVar[LogLevel]
    LOG_LEVEL_WARN: _ClassVar[LogLevel]
    LOG_LEVEL_ERROR: _ClassVar[LogLevel]
EVENT_TYPE_UNKNOWN: EventType
EVENT_TYPE_LOG: EventType
EVENT_TYPE_CALL: EventType
EVENT_TYPE_DEPLOYMENT_CREATED: EventType
EVENT_TYPE_DEPLOYMENT_UPDATED: EventType
EVENT_TYPE_INGRESS: EventType
EVENT_TYPE_CRON_SCHEDULED: EventType
EVENT_TYPE_ASYNC_EXECUTE: EventType
EVENT_TYPE_PUBSUB_PUBLISH: EventType
EVENT_TYPE_PUBSUB_CONSUME: EventType
ASYNC_EXECUTE_EVENT_TYPE_UNKNOWN: AsyncExecuteEventType
ASYNC_EXECUTE_EVENT_TYPE_CRON: AsyncExecuteEventType
ASYNC_EXECUTE_EVENT_TYPE_PUBSUB: AsyncExecuteEventType
LOG_LEVEL_UNKNOWN: LogLevel
LOG_LEVEL_TRACE: LogLevel
LOG_LEVEL_DEBUG: LogLevel
LOG_LEVEL_INFO: LogLevel
LOG_LEVEL_WARN: LogLevel
LOG_LEVEL_ERROR: LogLevel

class LogEvent(_message.Message):
    __slots__ = ("deployment_key", "request_key", "time_stamp", "log_level", "attributes", "message", "error", "stack")
    class AttributesEntry(_message.Message):
        __slots__ = ("key", "value")
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    DEPLOYMENT_KEY_FIELD_NUMBER: _ClassVar[int]
    REQUEST_KEY_FIELD_NUMBER: _ClassVar[int]
    TIME_STAMP_FIELD_NUMBER: _ClassVar[int]
    LOG_LEVEL_FIELD_NUMBER: _ClassVar[int]
    ATTRIBUTES_FIELD_NUMBER: _ClassVar[int]
    MESSAGE_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    STACK_FIELD_NUMBER: _ClassVar[int]
    deployment_key: str
    request_key: str
    time_stamp: _timestamp_pb2.Timestamp
    log_level: int
    attributes: _containers.ScalarMap[str, str]
    message: str
    error: str
    stack: str
    def __init__(self, deployment_key: _Optional[str] = ..., request_key: _Optional[str] = ..., time_stamp: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., log_level: _Optional[int] = ..., attributes: _Optional[_Mapping[str, str]] = ..., message: _Optional[str] = ..., error: _Optional[str] = ..., stack: _Optional[str] = ...) -> None: ...

class CallEvent(_message.Message):
    __slots__ = ("request_key", "deployment_key", "time_stamp", "source_verb_ref", "destination_verb_ref", "duration", "request", "response", "error", "stack")
    REQUEST_KEY_FIELD_NUMBER: _ClassVar[int]
    DEPLOYMENT_KEY_FIELD_NUMBER: _ClassVar[int]
    TIME_STAMP_FIELD_NUMBER: _ClassVar[int]
    SOURCE_VERB_REF_FIELD_NUMBER: _ClassVar[int]
    DESTINATION_VERB_REF_FIELD_NUMBER: _ClassVar[int]
    DURATION_FIELD_NUMBER: _ClassVar[int]
    REQUEST_FIELD_NUMBER: _ClassVar[int]
    RESPONSE_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    STACK_FIELD_NUMBER: _ClassVar[int]
    request_key: str
    deployment_key: str
    time_stamp: _timestamp_pb2.Timestamp
    source_verb_ref: _schema_pb2.Ref
    destination_verb_ref: _schema_pb2.Ref
    duration: _duration_pb2.Duration
    request: str
    response: str
    error: str
    stack: str
    def __init__(self, request_key: _Optional[str] = ..., deployment_key: _Optional[str] = ..., time_stamp: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., source_verb_ref: _Optional[_Union[_schema_pb2.Ref, _Mapping]] = ..., destination_verb_ref: _Optional[_Union[_schema_pb2.Ref, _Mapping]] = ..., duration: _Optional[_Union[_duration_pb2.Duration, _Mapping]] = ..., request: _Optional[str] = ..., response: _Optional[str] = ..., error: _Optional[str] = ..., stack: _Optional[str] = ...) -> None: ...

class DeploymentCreatedEvent(_message.Message):
    __slots__ = ("key", "language", "module_name", "min_replicas", "replaced")
    KEY_FIELD_NUMBER: _ClassVar[int]
    LANGUAGE_FIELD_NUMBER: _ClassVar[int]
    MODULE_NAME_FIELD_NUMBER: _ClassVar[int]
    MIN_REPLICAS_FIELD_NUMBER: _ClassVar[int]
    REPLACED_FIELD_NUMBER: _ClassVar[int]
    key: str
    language: str
    module_name: str
    min_replicas: int
    replaced: str
    def __init__(self, key: _Optional[str] = ..., language: _Optional[str] = ..., module_name: _Optional[str] = ..., min_replicas: _Optional[int] = ..., replaced: _Optional[str] = ...) -> None: ...

class DeploymentUpdatedEvent(_message.Message):
    __slots__ = ("key", "min_replicas", "prev_min_replicas")
    KEY_FIELD_NUMBER: _ClassVar[int]
    MIN_REPLICAS_FIELD_NUMBER: _ClassVar[int]
    PREV_MIN_REPLICAS_FIELD_NUMBER: _ClassVar[int]
    key: str
    min_replicas: int
    prev_min_replicas: int
    def __init__(self, key: _Optional[str] = ..., min_replicas: _Optional[int] = ..., prev_min_replicas: _Optional[int] = ...) -> None: ...

class IngressEvent(_message.Message):
    __slots__ = ("deployment_key", "request_key", "verb_ref", "method", "path", "status_code", "time_stamp", "duration", "request", "request_header", "response", "response_header", "error")
    DEPLOYMENT_KEY_FIELD_NUMBER: _ClassVar[int]
    REQUEST_KEY_FIELD_NUMBER: _ClassVar[int]
    VERB_REF_FIELD_NUMBER: _ClassVar[int]
    METHOD_FIELD_NUMBER: _ClassVar[int]
    PATH_FIELD_NUMBER: _ClassVar[int]
    STATUS_CODE_FIELD_NUMBER: _ClassVar[int]
    TIME_STAMP_FIELD_NUMBER: _ClassVar[int]
    DURATION_FIELD_NUMBER: _ClassVar[int]
    REQUEST_FIELD_NUMBER: _ClassVar[int]
    REQUEST_HEADER_FIELD_NUMBER: _ClassVar[int]
    RESPONSE_FIELD_NUMBER: _ClassVar[int]
    RESPONSE_HEADER_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    deployment_key: str
    request_key: str
    verb_ref: _schema_pb2.Ref
    method: str
    path: str
    status_code: int
    time_stamp: _timestamp_pb2.Timestamp
    duration: _duration_pb2.Duration
    request: str
    request_header: str
    response: str
    response_header: str
    error: str
    def __init__(self, deployment_key: _Optional[str] = ..., request_key: _Optional[str] = ..., verb_ref: _Optional[_Union[_schema_pb2.Ref, _Mapping]] = ..., method: _Optional[str] = ..., path: _Optional[str] = ..., status_code: _Optional[int] = ..., time_stamp: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., duration: _Optional[_Union[_duration_pb2.Duration, _Mapping]] = ..., request: _Optional[str] = ..., request_header: _Optional[str] = ..., response: _Optional[str] = ..., response_header: _Optional[str] = ..., error: _Optional[str] = ...) -> None: ...

class CronScheduledEvent(_message.Message):
    __slots__ = ("deployment_key", "verb_ref", "time_stamp", "duration", "scheduled_at", "schedule", "error")
    DEPLOYMENT_KEY_FIELD_NUMBER: _ClassVar[int]
    VERB_REF_FIELD_NUMBER: _ClassVar[int]
    TIME_STAMP_FIELD_NUMBER: _ClassVar[int]
    DURATION_FIELD_NUMBER: _ClassVar[int]
    SCHEDULED_AT_FIELD_NUMBER: _ClassVar[int]
    SCHEDULE_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    deployment_key: str
    verb_ref: _schema_pb2.Ref
    time_stamp: _timestamp_pb2.Timestamp
    duration: _duration_pb2.Duration
    scheduled_at: _timestamp_pb2.Timestamp
    schedule: str
    error: str
    def __init__(self, deployment_key: _Optional[str] = ..., verb_ref: _Optional[_Union[_schema_pb2.Ref, _Mapping]] = ..., time_stamp: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., duration: _Optional[_Union[_duration_pb2.Duration, _Mapping]] = ..., scheduled_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., schedule: _Optional[str] = ..., error: _Optional[str] = ...) -> None: ...

class AsyncExecuteEvent(_message.Message):
    __slots__ = ("deployment_key", "request_key", "verb_ref", "time_stamp", "duration", "async_event_type", "error")
    DEPLOYMENT_KEY_FIELD_NUMBER: _ClassVar[int]
    REQUEST_KEY_FIELD_NUMBER: _ClassVar[int]
    VERB_REF_FIELD_NUMBER: _ClassVar[int]
    TIME_STAMP_FIELD_NUMBER: _ClassVar[int]
    DURATION_FIELD_NUMBER: _ClassVar[int]
    ASYNC_EVENT_TYPE_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    deployment_key: str
    request_key: str
    verb_ref: _schema_pb2.Ref
    time_stamp: _timestamp_pb2.Timestamp
    duration: _duration_pb2.Duration
    async_event_type: AsyncExecuteEventType
    error: str
    def __init__(self, deployment_key: _Optional[str] = ..., request_key: _Optional[str] = ..., verb_ref: _Optional[_Union[_schema_pb2.Ref, _Mapping]] = ..., time_stamp: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., duration: _Optional[_Union[_duration_pb2.Duration, _Mapping]] = ..., async_event_type: _Optional[_Union[AsyncExecuteEventType, str]] = ..., error: _Optional[str] = ...) -> None: ...

class PubSubPublishEvent(_message.Message):
    __slots__ = ("deployment_key", "request_key", "verb_ref", "time_stamp", "duration", "topic", "request", "error")
    DEPLOYMENT_KEY_FIELD_NUMBER: _ClassVar[int]
    REQUEST_KEY_FIELD_NUMBER: _ClassVar[int]
    VERB_REF_FIELD_NUMBER: _ClassVar[int]
    TIME_STAMP_FIELD_NUMBER: _ClassVar[int]
    DURATION_FIELD_NUMBER: _ClassVar[int]
    TOPIC_FIELD_NUMBER: _ClassVar[int]
    REQUEST_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    deployment_key: str
    request_key: str
    verb_ref: _schema_pb2.Ref
    time_stamp: _timestamp_pb2.Timestamp
    duration: _duration_pb2.Duration
    topic: str
    request: str
    error: str
    def __init__(self, deployment_key: _Optional[str] = ..., request_key: _Optional[str] = ..., verb_ref: _Optional[_Union[_schema_pb2.Ref, _Mapping]] = ..., time_stamp: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., duration: _Optional[_Union[_duration_pb2.Duration, _Mapping]] = ..., topic: _Optional[str] = ..., request: _Optional[str] = ..., error: _Optional[str] = ...) -> None: ...

class PubSubConsumeEvent(_message.Message):
    __slots__ = ("deployment_key", "request_key", "dest_verb_module", "dest_verb_name", "time_stamp", "duration", "topic", "error")
    DEPLOYMENT_KEY_FIELD_NUMBER: _ClassVar[int]
    REQUEST_KEY_FIELD_NUMBER: _ClassVar[int]
    DEST_VERB_MODULE_FIELD_NUMBER: _ClassVar[int]
    DEST_VERB_NAME_FIELD_NUMBER: _ClassVar[int]
    TIME_STAMP_FIELD_NUMBER: _ClassVar[int]
    DURATION_FIELD_NUMBER: _ClassVar[int]
    TOPIC_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    deployment_key: str
    request_key: str
    dest_verb_module: str
    dest_verb_name: str
    time_stamp: _timestamp_pb2.Timestamp
    duration: _duration_pb2.Duration
    topic: str
    error: str
    def __init__(self, deployment_key: _Optional[str] = ..., request_key: _Optional[str] = ..., dest_verb_module: _Optional[str] = ..., dest_verb_name: _Optional[str] = ..., time_stamp: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., duration: _Optional[_Union[_duration_pb2.Duration, _Mapping]] = ..., topic: _Optional[str] = ..., error: _Optional[str] = ...) -> None: ...

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

class Subscription(_message.Message):
    __slots__ = ("subscription", "references")
    SUBSCRIPTION_FIELD_NUMBER: _ClassVar[int]
    REFERENCES_FIELD_NUMBER: _ClassVar[int]
    subscription: _schema_pb2.Subscription
    references: _containers.RepeatedCompositeFieldContainer[_schema_pb2.Ref]
    def __init__(self, subscription: _Optional[_Union[_schema_pb2.Subscription, _Mapping]] = ..., references: _Optional[_Iterable[_Union[_schema_pb2.Ref, _Mapping]]] = ...) -> None: ...

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
    __slots__ = ("name", "deployment_key", "language", "schema", "verbs", "data", "secrets", "configs", "databases", "enums", "topics", "typealiases", "subscriptions")
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
    SUBSCRIPTIONS_FIELD_NUMBER: _ClassVar[int]
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
    subscriptions: _containers.RepeatedCompositeFieldContainer[Subscription]
    def __init__(self, name: _Optional[str] = ..., deployment_key: _Optional[str] = ..., language: _Optional[str] = ..., schema: _Optional[str] = ..., verbs: _Optional[_Iterable[_Union[Verb, _Mapping]]] = ..., data: _Optional[_Iterable[_Union[Data, _Mapping]]] = ..., secrets: _Optional[_Iterable[_Union[Secret, _Mapping]]] = ..., configs: _Optional[_Iterable[_Union[Config, _Mapping]]] = ..., databases: _Optional[_Iterable[_Union[Database, _Mapping]]] = ..., enums: _Optional[_Iterable[_Union[Enum, _Mapping]]] = ..., topics: _Optional[_Iterable[_Union[Topic, _Mapping]]] = ..., typealiases: _Optional[_Iterable[_Union[TypeAlias, _Mapping]]] = ..., subscriptions: _Optional[_Iterable[_Union[Subscription, _Mapping]]] = ...) -> None: ...

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
    __slots__ = ("modules",)
    MODULES_FIELD_NUMBER: _ClassVar[int]
    modules: _containers.RepeatedCompositeFieldContainer[Module]
    def __init__(self, modules: _Optional[_Iterable[_Union[Module, _Mapping]]] = ...) -> None: ...

class EventsQuery(_message.Message):
    __slots__ = ("filters", "limit", "order")
    class Order(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = ()
        ASC: _ClassVar[EventsQuery.Order]
        DESC: _ClassVar[EventsQuery.Order]
    ASC: EventsQuery.Order
    DESC: EventsQuery.Order
    class LimitFilter(_message.Message):
        __slots__ = ("limit",)
        LIMIT_FIELD_NUMBER: _ClassVar[int]
        limit: int
        def __init__(self, limit: _Optional[int] = ...) -> None: ...
    class LogLevelFilter(_message.Message):
        __slots__ = ("log_level",)
        LOG_LEVEL_FIELD_NUMBER: _ClassVar[int]
        log_level: LogLevel
        def __init__(self, log_level: _Optional[_Union[LogLevel, str]] = ...) -> None: ...
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
        event_types: _containers.RepeatedScalarFieldContainer[EventType]
        def __init__(self, event_types: _Optional[_Iterable[_Union[EventType, str]]] = ...) -> None: ...
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
        limit: EventsQuery.LimitFilter
        log_level: EventsQuery.LogLevelFilter
        deployments: EventsQuery.DeploymentFilter
        requests: EventsQuery.RequestFilter
        event_types: EventsQuery.EventTypeFilter
        time: EventsQuery.TimeFilter
        id: EventsQuery.IDFilter
        call: EventsQuery.CallFilter
        module: EventsQuery.ModuleFilter
        def __init__(self, limit: _Optional[_Union[EventsQuery.LimitFilter, _Mapping]] = ..., log_level: _Optional[_Union[EventsQuery.LogLevelFilter, _Mapping]] = ..., deployments: _Optional[_Union[EventsQuery.DeploymentFilter, _Mapping]] = ..., requests: _Optional[_Union[EventsQuery.RequestFilter, _Mapping]] = ..., event_types: _Optional[_Union[EventsQuery.EventTypeFilter, _Mapping]] = ..., time: _Optional[_Union[EventsQuery.TimeFilter, _Mapping]] = ..., id: _Optional[_Union[EventsQuery.IDFilter, _Mapping]] = ..., call: _Optional[_Union[EventsQuery.CallFilter, _Mapping]] = ..., module: _Optional[_Union[EventsQuery.ModuleFilter, _Mapping]] = ...) -> None: ...
    FILTERS_FIELD_NUMBER: _ClassVar[int]
    LIMIT_FIELD_NUMBER: _ClassVar[int]
    ORDER_FIELD_NUMBER: _ClassVar[int]
    filters: _containers.RepeatedCompositeFieldContainer[EventsQuery.Filter]
    limit: int
    order: EventsQuery.Order
    def __init__(self, filters: _Optional[_Iterable[_Union[EventsQuery.Filter, _Mapping]]] = ..., limit: _Optional[int] = ..., order: _Optional[_Union[EventsQuery.Order, str]] = ...) -> None: ...

class StreamEventsRequest(_message.Message):
    __slots__ = ("update_interval", "query")
    UPDATE_INTERVAL_FIELD_NUMBER: _ClassVar[int]
    QUERY_FIELD_NUMBER: _ClassVar[int]
    update_interval: _duration_pb2.Duration
    query: EventsQuery
    def __init__(self, update_interval: _Optional[_Union[_duration_pb2.Duration, _Mapping]] = ..., query: _Optional[_Union[EventsQuery, _Mapping]] = ...) -> None: ...

class StreamEventsResponse(_message.Message):
    __slots__ = ("events",)
    EVENTS_FIELD_NUMBER: _ClassVar[int]
    events: _containers.RepeatedCompositeFieldContainer[Event]
    def __init__(self, events: _Optional[_Iterable[_Union[Event, _Mapping]]] = ...) -> None: ...

class Event(_message.Message):
    __slots__ = ("time_stamp", "id", "log", "call", "deployment_created", "deployment_updated", "ingress", "cron_scheduled", "async_execute", "pubsub_publish", "pubsub_consume")
    TIME_STAMP_FIELD_NUMBER: _ClassVar[int]
    ID_FIELD_NUMBER: _ClassVar[int]
    LOG_FIELD_NUMBER: _ClassVar[int]
    CALL_FIELD_NUMBER: _ClassVar[int]
    DEPLOYMENT_CREATED_FIELD_NUMBER: _ClassVar[int]
    DEPLOYMENT_UPDATED_FIELD_NUMBER: _ClassVar[int]
    INGRESS_FIELD_NUMBER: _ClassVar[int]
    CRON_SCHEDULED_FIELD_NUMBER: _ClassVar[int]
    ASYNC_EXECUTE_FIELD_NUMBER: _ClassVar[int]
    PUBSUB_PUBLISH_FIELD_NUMBER: _ClassVar[int]
    PUBSUB_CONSUME_FIELD_NUMBER: _ClassVar[int]
    time_stamp: _timestamp_pb2.Timestamp
    id: int
    log: LogEvent
    call: CallEvent
    deployment_created: DeploymentCreatedEvent
    deployment_updated: DeploymentUpdatedEvent
    ingress: IngressEvent
    cron_scheduled: CronScheduledEvent
    async_execute: AsyncExecuteEvent
    pubsub_publish: PubSubPublishEvent
    pubsub_consume: PubSubConsumeEvent
    def __init__(self, time_stamp: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., id: _Optional[int] = ..., log: _Optional[_Union[LogEvent, _Mapping]] = ..., call: _Optional[_Union[CallEvent, _Mapping]] = ..., deployment_created: _Optional[_Union[DeploymentCreatedEvent, _Mapping]] = ..., deployment_updated: _Optional[_Union[DeploymentUpdatedEvent, _Mapping]] = ..., ingress: _Optional[_Union[IngressEvent, _Mapping]] = ..., cron_scheduled: _Optional[_Union[CronScheduledEvent, _Mapping]] = ..., async_execute: _Optional[_Union[AsyncExecuteEvent, _Mapping]] = ..., pubsub_publish: _Optional[_Union[PubSubPublishEvent, _Mapping]] = ..., pubsub_consume: _Optional[_Union[PubSubConsumeEvent, _Mapping]] = ...) -> None: ...

class GetEventsResponse(_message.Message):
    __slots__ = ("events", "cursor")
    EVENTS_FIELD_NUMBER: _ClassVar[int]
    CURSOR_FIELD_NUMBER: _ClassVar[int]
    events: _containers.RepeatedCompositeFieldContainer[Event]
    cursor: int
    def __init__(self, events: _Optional[_Iterable[_Union[Event, _Mapping]]] = ..., cursor: _Optional[int] = ...) -> None: ...
