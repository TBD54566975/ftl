from xyz.block.ftl.v1 import ftl_pb2 as _ftl_pb2
from xyz.block.ftl.v1beta1.provisioner import resource_pb2 as _resource_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class ResourceContext(_message.Message):
    __slots__ = ("resource", "dependencies")
    RESOURCE_FIELD_NUMBER: _ClassVar[int]
    DEPENDENCIES_FIELD_NUMBER: _ClassVar[int]
    resource: _resource_pb2.Resource
    dependencies: _containers.RepeatedCompositeFieldContainer[_resource_pb2.Resource]
    def __init__(self, resource: _Optional[_Union[_resource_pb2.Resource, _Mapping]] = ..., dependencies: _Optional[_Iterable[_Union[_resource_pb2.Resource, _Mapping]]] = ...) -> None: ...

class ProvisionRequest(_message.Message):
    __slots__ = ("ftl_cluster_id", "module", "existing_resources", "desired_resources")
    FTL_CLUSTER_ID_FIELD_NUMBER: _ClassVar[int]
    MODULE_FIELD_NUMBER: _ClassVar[int]
    EXISTING_RESOURCES_FIELD_NUMBER: _ClassVar[int]
    DESIRED_RESOURCES_FIELD_NUMBER: _ClassVar[int]
    ftl_cluster_id: str
    module: str
    existing_resources: _containers.RepeatedCompositeFieldContainer[_resource_pb2.Resource]
    desired_resources: _containers.RepeatedCompositeFieldContainer[ResourceContext]
    def __init__(self, ftl_cluster_id: _Optional[str] = ..., module: _Optional[str] = ..., existing_resources: _Optional[_Iterable[_Union[_resource_pb2.Resource, _Mapping]]] = ..., desired_resources: _Optional[_Iterable[_Union[ResourceContext, _Mapping]]] = ...) -> None: ...

class ProvisionResponse(_message.Message):
    __slots__ = ("provisioning_token", "status")
    class ProvisionResponseStatus(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = ()
        UNKNOWN: _ClassVar[ProvisionResponse.ProvisionResponseStatus]
        SUBMITTED: _ClassVar[ProvisionResponse.ProvisionResponseStatus]
    UNKNOWN: ProvisionResponse.ProvisionResponseStatus
    SUBMITTED: ProvisionResponse.ProvisionResponseStatus
    PROVISIONING_TOKEN_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    provisioning_token: str
    status: ProvisionResponse.ProvisionResponseStatus
    def __init__(self, provisioning_token: _Optional[str] = ..., status: _Optional[_Union[ProvisionResponse.ProvisionResponseStatus, str]] = ...) -> None: ...

class StatusRequest(_message.Message):
    __slots__ = ("provisioning_token", "desired_resources")
    PROVISIONING_TOKEN_FIELD_NUMBER: _ClassVar[int]
    DESIRED_RESOURCES_FIELD_NUMBER: _ClassVar[int]
    provisioning_token: str
    desired_resources: _containers.RepeatedCompositeFieldContainer[_resource_pb2.Resource]
    def __init__(self, provisioning_token: _Optional[str] = ..., desired_resources: _Optional[_Iterable[_Union[_resource_pb2.Resource, _Mapping]]] = ...) -> None: ...

class StatusResponse(_message.Message):
    __slots__ = ("running", "success")
    class ProvisioningRunning(_message.Message):
        __slots__ = ()
        def __init__(self) -> None: ...
    class ProvisioningFailed(_message.Message):
        __slots__ = ("error_message",)
        ERROR_MESSAGE_FIELD_NUMBER: _ClassVar[int]
        error_message: str
        def __init__(self, error_message: _Optional[str] = ...) -> None: ...
    class ProvisioningSuccess(_message.Message):
        __slots__ = ("updated_resources",)
        UPDATED_RESOURCES_FIELD_NUMBER: _ClassVar[int]
        updated_resources: _containers.RepeatedCompositeFieldContainer[_resource_pb2.Resource]
        def __init__(self, updated_resources: _Optional[_Iterable[_Union[_resource_pb2.Resource, _Mapping]]] = ...) -> None: ...
    RUNNING_FIELD_NUMBER: _ClassVar[int]
    SUCCESS_FIELD_NUMBER: _ClassVar[int]
    running: StatusResponse.ProvisioningRunning
    success: StatusResponse.ProvisioningSuccess
    def __init__(self, running: _Optional[_Union[StatusResponse.ProvisioningRunning, _Mapping]] = ..., success: _Optional[_Union[StatusResponse.ProvisioningSuccess, _Mapping]] = ...) -> None: ...

class PlanRequest(_message.Message):
    __slots__ = ("provisioning",)
    PROVISIONING_FIELD_NUMBER: _ClassVar[int]
    provisioning: ProvisionRequest
    def __init__(self, provisioning: _Optional[_Union[ProvisionRequest, _Mapping]] = ...) -> None: ...

class PlanResponse(_message.Message):
    __slots__ = ("plan",)
    PLAN_FIELD_NUMBER: _ClassVar[int]
    plan: str
    def __init__(self, plan: _Optional[str] = ...) -> None: ...
