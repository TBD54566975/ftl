from google.protobuf import duration_pb2 as _duration_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from xyz.block.ftl.schema.v1 import schema_pb2 as _schema_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class EventType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    EVENT_TYPE_UNSPECIFIED: _ClassVar[EventType]
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
    ASYNC_EXECUTE_EVENT_TYPE_UNSPECIFIED: _ClassVar[AsyncExecuteEventType]
    ASYNC_EXECUTE_EVENT_TYPE_CRON: _ClassVar[AsyncExecuteEventType]
    ASYNC_EXECUTE_EVENT_TYPE_PUBSUB: _ClassVar[AsyncExecuteEventType]

class LogLevel(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    LOG_LEVEL_UNSPECIFIED: _ClassVar[LogLevel]
    LOG_LEVEL_TRACE: _ClassVar[LogLevel]
    LOG_LEVEL_DEBUG: _ClassVar[LogLevel]
    LOG_LEVEL_INFO: _ClassVar[LogLevel]
    LOG_LEVEL_WARN: _ClassVar[LogLevel]
    LOG_LEVEL_ERROR: _ClassVar[LogLevel]
EVENT_TYPE_UNSPECIFIED: EventType
EVENT_TYPE_LOG: EventType
EVENT_TYPE_CALL: EventType
EVENT_TYPE_DEPLOYMENT_CREATED: EventType
EVENT_TYPE_DEPLOYMENT_UPDATED: EventType
EVENT_TYPE_INGRESS: EventType
EVENT_TYPE_CRON_SCHEDULED: EventType
EVENT_TYPE_ASYNC_EXECUTE: EventType
EVENT_TYPE_PUBSUB_PUBLISH: EventType
EVENT_TYPE_PUBSUB_CONSUME: EventType
ASYNC_EXECUTE_EVENT_TYPE_UNSPECIFIED: AsyncExecuteEventType
ASYNC_EXECUTE_EVENT_TYPE_CRON: AsyncExecuteEventType
ASYNC_EXECUTE_EVENT_TYPE_PUBSUB: AsyncExecuteEventType
LOG_LEVEL_UNSPECIFIED: LogLevel
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
