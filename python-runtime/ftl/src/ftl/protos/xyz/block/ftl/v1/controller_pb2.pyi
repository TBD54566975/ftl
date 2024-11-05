from google.protobuf import struct_pb2 as _struct_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from xyz.block.ftl.v1 import ftl_pb2 as _ftl_pb2
from xyz.block.ftl.v1.schema import schema_pb2 as _schema_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class DeploymentChangeType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    DEPLOYMENT_ADDED: _ClassVar[DeploymentChangeType]
    DEPLOYMENT_REMOVED: _ClassVar[DeploymentChangeType]
    DEPLOYMENT_CHANGED: _ClassVar[DeploymentChangeType]
DEPLOYMENT_ADDED: DeploymentChangeType
DEPLOYMENT_REMOVED: DeploymentChangeType
DEPLOYMENT_CHANGED: DeploymentChangeType

class GetCertificationRequest(_message.Message):
    __slots__ = ("request", "signature")
    REQUEST_FIELD_NUMBER: _ClassVar[int]
    SIGNATURE_FIELD_NUMBER: _ClassVar[int]
    request: CertificateContent
    signature: bytes
    def __init__(self, request: _Optional[_Union[CertificateContent, _Mapping]] = ..., signature: _Optional[bytes] = ...) -> None: ...

class GetCertificationResponse(_message.Message):
    __slots__ = ("certificate",)
    CERTIFICATE_FIELD_NUMBER: _ClassVar[int]
    certificate: Certificate
    def __init__(self, certificate: _Optional[_Union[Certificate, _Mapping]] = ...) -> None: ...

class CertificateContent(_message.Message):
    __slots__ = ("identity", "public_key")
    IDENTITY_FIELD_NUMBER: _ClassVar[int]
    PUBLIC_KEY_FIELD_NUMBER: _ClassVar[int]
    identity: str
    public_key: bytes
    def __init__(self, identity: _Optional[str] = ..., public_key: _Optional[bytes] = ...) -> None: ...

class Certificate(_message.Message):
    __slots__ = ("content", "controller_signature")
    CONTENT_FIELD_NUMBER: _ClassVar[int]
    CONTROLLER_SIGNATURE_FIELD_NUMBER: _ClassVar[int]
    content: CertificateContent
    controller_signature: bytes
    def __init__(self, content: _Optional[_Union[CertificateContent, _Mapping]] = ..., controller_signature: _Optional[bytes] = ...) -> None: ...

class GetSchemaRequest(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class GetSchemaResponse(_message.Message):
    __slots__ = ("schema",)
    SCHEMA_FIELD_NUMBER: _ClassVar[int]
    schema: _schema_pb2.Schema
    def __init__(self, schema: _Optional[_Union[_schema_pb2.Schema, _Mapping]] = ...) -> None: ...

class PullSchemaRequest(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class PullSchemaResponse(_message.Message):
    __slots__ = ("deployment_key", "module_name", "schema", "more", "change_type")
    DEPLOYMENT_KEY_FIELD_NUMBER: _ClassVar[int]
    MODULE_NAME_FIELD_NUMBER: _ClassVar[int]
    SCHEMA_FIELD_NUMBER: _ClassVar[int]
    MORE_FIELD_NUMBER: _ClassVar[int]
    CHANGE_TYPE_FIELD_NUMBER: _ClassVar[int]
    deployment_key: str
    module_name: str
    schema: _schema_pb2.Module
    more: bool
    change_type: DeploymentChangeType
    def __init__(self, deployment_key: _Optional[str] = ..., module_name: _Optional[str] = ..., schema: _Optional[_Union[_schema_pb2.Module, _Mapping]] = ..., more: bool = ..., change_type: _Optional[_Union[DeploymentChangeType, str]] = ...) -> None: ...

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

class CreateDeploymentRequest(_message.Message):
    __slots__ = ("schema", "artefacts", "labels")
    SCHEMA_FIELD_NUMBER: _ClassVar[int]
    ARTEFACTS_FIELD_NUMBER: _ClassVar[int]
    LABELS_FIELD_NUMBER: _ClassVar[int]
    schema: _schema_pb2.Module
    artefacts: _containers.RepeatedCompositeFieldContainer[DeploymentArtefact]
    labels: _struct_pb2.Struct
    def __init__(self, schema: _Optional[_Union[_schema_pb2.Module, _Mapping]] = ..., artefacts: _Optional[_Iterable[_Union[DeploymentArtefact, _Mapping]]] = ..., labels: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ...) -> None: ...

class CreateDeploymentResponse(_message.Message):
    __slots__ = ("deployment_key", "active_deployment_key")
    DEPLOYMENT_KEY_FIELD_NUMBER: _ClassVar[int]
    ACTIVE_DEPLOYMENT_KEY_FIELD_NUMBER: _ClassVar[int]
    deployment_key: str
    active_deployment_key: str
    def __init__(self, deployment_key: _Optional[str] = ..., active_deployment_key: _Optional[str] = ...) -> None: ...

class GetDeploymentRequest(_message.Message):
    __slots__ = ("deployment_key",)
    DEPLOYMENT_KEY_FIELD_NUMBER: _ClassVar[int]
    deployment_key: str
    def __init__(self, deployment_key: _Optional[str] = ...) -> None: ...

class GetDeploymentResponse(_message.Message):
    __slots__ = ("schema", "artefacts")
    SCHEMA_FIELD_NUMBER: _ClassVar[int]
    ARTEFACTS_FIELD_NUMBER: _ClassVar[int]
    schema: _schema_pb2.Module
    artefacts: _containers.RepeatedCompositeFieldContainer[DeploymentArtefact]
    def __init__(self, schema: _Optional[_Union[_schema_pb2.Module, _Mapping]] = ..., artefacts: _Optional[_Iterable[_Union[DeploymentArtefact, _Mapping]]] = ...) -> None: ...

class RegisterRunnerRequest(_message.Message):
    __slots__ = ("key", "endpoint", "deployment", "labels")
    KEY_FIELD_NUMBER: _ClassVar[int]
    ENDPOINT_FIELD_NUMBER: _ClassVar[int]
    DEPLOYMENT_FIELD_NUMBER: _ClassVar[int]
    LABELS_FIELD_NUMBER: _ClassVar[int]
    key: str
    endpoint: str
    deployment: str
    labels: _struct_pb2.Struct
    def __init__(self, key: _Optional[str] = ..., endpoint: _Optional[str] = ..., deployment: _Optional[str] = ..., labels: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ...) -> None: ...

class RegisterRunnerResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class UpdateDeployRequest(_message.Message):
    __slots__ = ("deployment_key", "min_replicas")
    DEPLOYMENT_KEY_FIELD_NUMBER: _ClassVar[int]
    MIN_REPLICAS_FIELD_NUMBER: _ClassVar[int]
    deployment_key: str
    min_replicas: int
    def __init__(self, deployment_key: _Optional[str] = ..., min_replicas: _Optional[int] = ...) -> None: ...

class UpdateDeployResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class ReplaceDeployRequest(_message.Message):
    __slots__ = ("deployment_key", "min_replicas")
    DEPLOYMENT_KEY_FIELD_NUMBER: _ClassVar[int]
    MIN_REPLICAS_FIELD_NUMBER: _ClassVar[int]
    deployment_key: str
    min_replicas: int
    def __init__(self, deployment_key: _Optional[str] = ..., min_replicas: _Optional[int] = ...) -> None: ...

class ReplaceDeployResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class StreamDeploymentLogsRequest(_message.Message):
    __slots__ = ("deployment_key", "request_key", "time_stamp", "log_level", "attributes", "message", "error")
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
    deployment_key: str
    request_key: str
    time_stamp: _timestamp_pb2.Timestamp
    log_level: int
    attributes: _containers.ScalarMap[str, str]
    message: str
    error: str
    def __init__(self, deployment_key: _Optional[str] = ..., request_key: _Optional[str] = ..., time_stamp: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., log_level: _Optional[int] = ..., attributes: _Optional[_Mapping[str, str]] = ..., message: _Optional[str] = ..., error: _Optional[str] = ...) -> None: ...

class StreamDeploymentLogsResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class StatusRequest(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class StatusResponse(_message.Message):
    __slots__ = ("controllers", "runners", "deployments", "ingress_routes", "routes")
    class Controller(_message.Message):
        __slots__ = ("key", "endpoint", "version")
        KEY_FIELD_NUMBER: _ClassVar[int]
        ENDPOINT_FIELD_NUMBER: _ClassVar[int]
        VERSION_FIELD_NUMBER: _ClassVar[int]
        key: str
        endpoint: str
        version: str
        def __init__(self, key: _Optional[str] = ..., endpoint: _Optional[str] = ..., version: _Optional[str] = ...) -> None: ...
    class Runner(_message.Message):
        __slots__ = ("key", "endpoint", "deployment", "labels")
        KEY_FIELD_NUMBER: _ClassVar[int]
        ENDPOINT_FIELD_NUMBER: _ClassVar[int]
        DEPLOYMENT_FIELD_NUMBER: _ClassVar[int]
        LABELS_FIELD_NUMBER: _ClassVar[int]
        key: str
        endpoint: str
        deployment: str
        labels: _struct_pb2.Struct
        def __init__(self, key: _Optional[str] = ..., endpoint: _Optional[str] = ..., deployment: _Optional[str] = ..., labels: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ...) -> None: ...
    class Deployment(_message.Message):
        __slots__ = ("key", "language", "name", "min_replicas", "replicas", "labels", "schema")
        KEY_FIELD_NUMBER: _ClassVar[int]
        LANGUAGE_FIELD_NUMBER: _ClassVar[int]
        NAME_FIELD_NUMBER: _ClassVar[int]
        MIN_REPLICAS_FIELD_NUMBER: _ClassVar[int]
        REPLICAS_FIELD_NUMBER: _ClassVar[int]
        LABELS_FIELD_NUMBER: _ClassVar[int]
        SCHEMA_FIELD_NUMBER: _ClassVar[int]
        key: str
        language: str
        name: str
        min_replicas: int
        replicas: int
        labels: _struct_pb2.Struct
        schema: _schema_pb2.Module
        def __init__(self, key: _Optional[str] = ..., language: _Optional[str] = ..., name: _Optional[str] = ..., min_replicas: _Optional[int] = ..., replicas: _Optional[int] = ..., labels: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., schema: _Optional[_Union[_schema_pb2.Module, _Mapping]] = ...) -> None: ...
    class IngressRoute(_message.Message):
        __slots__ = ("deployment_key", "verb", "method", "path")
        DEPLOYMENT_KEY_FIELD_NUMBER: _ClassVar[int]
        VERB_FIELD_NUMBER: _ClassVar[int]
        METHOD_FIELD_NUMBER: _ClassVar[int]
        PATH_FIELD_NUMBER: _ClassVar[int]
        deployment_key: str
        verb: _schema_pb2.Ref
        method: str
        path: str
        def __init__(self, deployment_key: _Optional[str] = ..., verb: _Optional[_Union[_schema_pb2.Ref, _Mapping]] = ..., method: _Optional[str] = ..., path: _Optional[str] = ...) -> None: ...
    class Route(_message.Message):
        __slots__ = ("module", "deployment", "endpoint")
        MODULE_FIELD_NUMBER: _ClassVar[int]
        DEPLOYMENT_FIELD_NUMBER: _ClassVar[int]
        ENDPOINT_FIELD_NUMBER: _ClassVar[int]
        module: str
        deployment: str
        endpoint: str
        def __init__(self, module: _Optional[str] = ..., deployment: _Optional[str] = ..., endpoint: _Optional[str] = ...) -> None: ...
    CONTROLLERS_FIELD_NUMBER: _ClassVar[int]
    RUNNERS_FIELD_NUMBER: _ClassVar[int]
    DEPLOYMENTS_FIELD_NUMBER: _ClassVar[int]
    INGRESS_ROUTES_FIELD_NUMBER: _ClassVar[int]
    ROUTES_FIELD_NUMBER: _ClassVar[int]
    controllers: _containers.RepeatedCompositeFieldContainer[StatusResponse.Controller]
    runners: _containers.RepeatedCompositeFieldContainer[StatusResponse.Runner]
    deployments: _containers.RepeatedCompositeFieldContainer[StatusResponse.Deployment]
    ingress_routes: _containers.RepeatedCompositeFieldContainer[StatusResponse.IngressRoute]
    routes: _containers.RepeatedCompositeFieldContainer[StatusResponse.Route]
    def __init__(self, controllers: _Optional[_Iterable[_Union[StatusResponse.Controller, _Mapping]]] = ..., runners: _Optional[_Iterable[_Union[StatusResponse.Runner, _Mapping]]] = ..., deployments: _Optional[_Iterable[_Union[StatusResponse.Deployment, _Mapping]]] = ..., ingress_routes: _Optional[_Iterable[_Union[StatusResponse.IngressRoute, _Mapping]]] = ..., routes: _Optional[_Iterable[_Union[StatusResponse.Route, _Mapping]]] = ...) -> None: ...

class ProcessListRequest(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class ProcessListResponse(_message.Message):
    __slots__ = ("processes",)
    class ProcessRunner(_message.Message):
        __slots__ = ("key", "endpoint", "labels")
        KEY_FIELD_NUMBER: _ClassVar[int]
        ENDPOINT_FIELD_NUMBER: _ClassVar[int]
        LABELS_FIELD_NUMBER: _ClassVar[int]
        key: str
        endpoint: str
        labels: _struct_pb2.Struct
        def __init__(self, key: _Optional[str] = ..., endpoint: _Optional[str] = ..., labels: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ...) -> None: ...
    class Process(_message.Message):
        __slots__ = ("deployment", "min_replicas", "labels", "runner")
        DEPLOYMENT_FIELD_NUMBER: _ClassVar[int]
        MIN_REPLICAS_FIELD_NUMBER: _ClassVar[int]
        LABELS_FIELD_NUMBER: _ClassVar[int]
        RUNNER_FIELD_NUMBER: _ClassVar[int]
        deployment: str
        min_replicas: int
        labels: _struct_pb2.Struct
        runner: ProcessListResponse.ProcessRunner
        def __init__(self, deployment: _Optional[str] = ..., min_replicas: _Optional[int] = ..., labels: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ..., runner: _Optional[_Union[ProcessListResponse.ProcessRunner, _Mapping]] = ...) -> None: ...
    PROCESSES_FIELD_NUMBER: _ClassVar[int]
    processes: _containers.RepeatedCompositeFieldContainer[ProcessListResponse.Process]
    def __init__(self, processes: _Optional[_Iterable[_Union[ProcessListResponse.Process, _Mapping]]] = ...) -> None: ...

class ResetSubscriptionRequest(_message.Message):
    __slots__ = ("subscription",)
    SUBSCRIPTION_FIELD_NUMBER: _ClassVar[int]
    subscription: _schema_pb2.Ref
    def __init__(self, subscription: _Optional[_Union[_schema_pb2.Ref, _Mapping]] = ...) -> None: ...

class ResetSubscriptionResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...
