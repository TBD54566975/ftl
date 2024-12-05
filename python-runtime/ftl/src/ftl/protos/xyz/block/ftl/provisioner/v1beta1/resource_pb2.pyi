from xyz.block.ftl.artefacts.v1 import artefacts_pb2 as _artefacts_pb2
from xyz.block.ftl.schema.v1 import schema_pb2 as _schema_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Resource(_message.Message):
    __slots__ = ("resource_id", "postgres", "mysql", "module", "sql_migration", "topic", "subscription", "runner")
    RESOURCE_ID_FIELD_NUMBER: _ClassVar[int]
    POSTGRES_FIELD_NUMBER: _ClassVar[int]
    MYSQL_FIELD_NUMBER: _ClassVar[int]
    MODULE_FIELD_NUMBER: _ClassVar[int]
    SQL_MIGRATION_FIELD_NUMBER: _ClassVar[int]
    TOPIC_FIELD_NUMBER: _ClassVar[int]
    SUBSCRIPTION_FIELD_NUMBER: _ClassVar[int]
    RUNNER_FIELD_NUMBER: _ClassVar[int]
    resource_id: str
    postgres: PostgresResource
    mysql: MysqlResource
    module: ModuleResource
    sql_migration: SqlMigrationResource
    topic: TopicResource
    subscription: SubscriptionResource
    runner: RunnerResource
    def __init__(self, resource_id: _Optional[str] = ..., postgres: _Optional[_Union[PostgresResource, _Mapping]] = ..., mysql: _Optional[_Union[MysqlResource, _Mapping]] = ..., module: _Optional[_Union[ModuleResource, _Mapping]] = ..., sql_migration: _Optional[_Union[SqlMigrationResource, _Mapping]] = ..., topic: _Optional[_Union[TopicResource, _Mapping]] = ..., subscription: _Optional[_Union[SubscriptionResource, _Mapping]] = ..., runner: _Optional[_Union[RunnerResource, _Mapping]] = ...) -> None: ...

class PostgresResource(_message.Message):
    __slots__ = ("output",)
    OUTPUT_FIELD_NUMBER: _ClassVar[int]
    output: _schema_pb2.DatabaseRuntime
    def __init__(self, output: _Optional[_Union[_schema_pb2.DatabaseRuntime, _Mapping]] = ...) -> None: ...

class MysqlResource(_message.Message):
    __slots__ = ("output",)
    OUTPUT_FIELD_NUMBER: _ClassVar[int]
    output: _schema_pb2.DatabaseRuntime
    def __init__(self, output: _Optional[_Union[_schema_pb2.DatabaseRuntime, _Mapping]] = ...) -> None: ...

class SqlMigrationResource(_message.Message):
    __slots__ = ("output", "digest")
    class SqlMigrationResourceOutput(_message.Message):
        __slots__ = ()
        def __init__(self) -> None: ...
    OUTPUT_FIELD_NUMBER: _ClassVar[int]
    DIGEST_FIELD_NUMBER: _ClassVar[int]
    output: SqlMigrationResource.SqlMigrationResourceOutput
    digest: str
    def __init__(self, output: _Optional[_Union[SqlMigrationResource.SqlMigrationResourceOutput, _Mapping]] = ..., digest: _Optional[str] = ...) -> None: ...

class ModuleResource(_message.Message):
    __slots__ = ("output", "schema", "artefacts")
    class ModuleResourceOutput(_message.Message):
        __slots__ = ("deployment_key",)
        DEPLOYMENT_KEY_FIELD_NUMBER: _ClassVar[int]
        deployment_key: str
        def __init__(self, deployment_key: _Optional[str] = ...) -> None: ...
    OUTPUT_FIELD_NUMBER: _ClassVar[int]
    SCHEMA_FIELD_NUMBER: _ClassVar[int]
    ARTEFACTS_FIELD_NUMBER: _ClassVar[int]
    output: ModuleResource.ModuleResourceOutput
    schema: _schema_pb2.Module
    artefacts: _containers.RepeatedCompositeFieldContainer[_artefacts_pb2.DeploymentArtefact]
    def __init__(self, output: _Optional[_Union[ModuleResource.ModuleResourceOutput, _Mapping]] = ..., schema: _Optional[_Union[_schema_pb2.Module, _Mapping]] = ..., artefacts: _Optional[_Iterable[_Union[_artefacts_pb2.DeploymentArtefact, _Mapping]]] = ...) -> None: ...

class RunnerResource(_message.Message):
    __slots__ = ("output",)
    class RunnerResourceOutput(_message.Message):
        __slots__ = ("runner_uri", "deployment_key")
        RUNNER_URI_FIELD_NUMBER: _ClassVar[int]
        DEPLOYMENT_KEY_FIELD_NUMBER: _ClassVar[int]
        runner_uri: str
        deployment_key: str
        def __init__(self, runner_uri: _Optional[str] = ..., deployment_key: _Optional[str] = ...) -> None: ...
    OUTPUT_FIELD_NUMBER: _ClassVar[int]
    output: RunnerResource.RunnerResourceOutput
    def __init__(self, output: _Optional[_Union[RunnerResource.RunnerResourceOutput, _Mapping]] = ...) -> None: ...

class TopicResource(_message.Message):
    __slots__ = ("output",)
    class TopicResourceOutput(_message.Message):
        __slots__ = ("kafka_brokers", "topic_id")
        KAFKA_BROKERS_FIELD_NUMBER: _ClassVar[int]
        TOPIC_ID_FIELD_NUMBER: _ClassVar[int]
        kafka_brokers: _containers.RepeatedScalarFieldContainer[str]
        topic_id: str
        def __init__(self, kafka_brokers: _Optional[_Iterable[str]] = ..., topic_id: _Optional[str] = ...) -> None: ...
    OUTPUT_FIELD_NUMBER: _ClassVar[int]
    output: TopicResource.TopicResourceOutput
    def __init__(self, output: _Optional[_Union[TopicResource.TopicResourceOutput, _Mapping]] = ...) -> None: ...

class SubscriptionResource(_message.Message):
    __slots__ = ("output", "topic")
    class SubscriptionResourceOutput(_message.Message):
        __slots__ = ("kafka_brokers", "topic_id", "consumer_group_id")
        KAFKA_BROKERS_FIELD_NUMBER: _ClassVar[int]
        TOPIC_ID_FIELD_NUMBER: _ClassVar[int]
        CONSUMER_GROUP_ID_FIELD_NUMBER: _ClassVar[int]
        kafka_brokers: _containers.RepeatedScalarFieldContainer[str]
        topic_id: str
        consumer_group_id: str
        def __init__(self, kafka_brokers: _Optional[_Iterable[str]] = ..., topic_id: _Optional[str] = ..., consumer_group_id: _Optional[str] = ...) -> None: ...
    OUTPUT_FIELD_NUMBER: _ClassVar[int]
    TOPIC_FIELD_NUMBER: _ClassVar[int]
    output: SubscriptionResource.SubscriptionResourceOutput
    topic: _schema_pb2.Ref
    def __init__(self, output: _Optional[_Union[SubscriptionResource.SubscriptionResourceOutput, _Mapping]] = ..., topic: _Optional[_Union[_schema_pb2.Ref, _Mapping]] = ...) -> None: ...
