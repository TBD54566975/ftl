from google.protobuf import struct_pb2 as _struct_pb2
from xyz.block.ftl.v1 import ftl_pb2 as _ftl_pb2
from xyz.block.ftl.v1.schema import schema_pb2 as _schema_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class ModuleConfig(_message.Message):
    __slots__ = ("name", "dir", "language", "deploy_dir", "build", "build_lock", "generated_schema_dir", "watch", "language_config")
    NAME_FIELD_NUMBER: _ClassVar[int]
    DIR_FIELD_NUMBER: _ClassVar[int]
    LANGUAGE_FIELD_NUMBER: _ClassVar[int]
    DEPLOY_DIR_FIELD_NUMBER: _ClassVar[int]
    BUILD_FIELD_NUMBER: _ClassVar[int]
    BUILD_LOCK_FIELD_NUMBER: _ClassVar[int]
    GENERATED_SCHEMA_DIR_FIELD_NUMBER: _ClassVar[int]
    WATCH_FIELD_NUMBER: _ClassVar[int]
    LANGUAGE_CONFIG_FIELD_NUMBER: _ClassVar[int]
    name: str
    dir: str
    language: str
    deploy_dir: str
    build: str
    build_lock: str
    generated_schema_dir: str
    watch: _containers.RepeatedScalarFieldContainer[str]
    language_config: _struct_pb2.Struct
    def __init__(self, name: _Optional[str] = ..., dir: _Optional[str] = ..., language: _Optional[str] = ..., deploy_dir: _Optional[str] = ..., build: _Optional[str] = ..., build_lock: _Optional[str] = ..., generated_schema_dir: _Optional[str] = ..., watch: _Optional[_Iterable[str]] = ..., language_config: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ...) -> None: ...

class ProjectConfig(_message.Message):
    __slots__ = ("dir", "name", "no_git", "hermit")
    DIR_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    NO_GIT_FIELD_NUMBER: _ClassVar[int]
    HERMIT_FIELD_NUMBER: _ClassVar[int]
    dir: str
    name: str
    no_git: bool
    hermit: bool
    def __init__(self, dir: _Optional[str] = ..., name: _Optional[str] = ..., no_git: bool = ..., hermit: bool = ...) -> None: ...

class GetCreateModuleFlagsRequest(_message.Message):
    __slots__ = ("language",)
    LANGUAGE_FIELD_NUMBER: _ClassVar[int]
    language: str
    def __init__(self, language: _Optional[str] = ...) -> None: ...

class GetCreateModuleFlagsResponse(_message.Message):
    __slots__ = ("flags",)
    class Flag(_message.Message):
        __slots__ = ("name", "help", "envar", "short", "placeholder", "default")
        NAME_FIELD_NUMBER: _ClassVar[int]
        HELP_FIELD_NUMBER: _ClassVar[int]
        ENVAR_FIELD_NUMBER: _ClassVar[int]
        SHORT_FIELD_NUMBER: _ClassVar[int]
        PLACEHOLDER_FIELD_NUMBER: _ClassVar[int]
        DEFAULT_FIELD_NUMBER: _ClassVar[int]
        name: str
        help: str
        envar: str
        short: str
        placeholder: str
        default: str
        def __init__(self, name: _Optional[str] = ..., help: _Optional[str] = ..., envar: _Optional[str] = ..., short: _Optional[str] = ..., placeholder: _Optional[str] = ..., default: _Optional[str] = ...) -> None: ...
    FLAGS_FIELD_NUMBER: _ClassVar[int]
    flags: _containers.RepeatedCompositeFieldContainer[GetCreateModuleFlagsResponse.Flag]
    def __init__(self, flags: _Optional[_Iterable[_Union[GetCreateModuleFlagsResponse.Flag, _Mapping]]] = ...) -> None: ...

class CreateModuleRequest(_message.Message):
    __slots__ = ("name", "dir", "project_config", "Flags")
    NAME_FIELD_NUMBER: _ClassVar[int]
    DIR_FIELD_NUMBER: _ClassVar[int]
    PROJECT_CONFIG_FIELD_NUMBER: _ClassVar[int]
    FLAGS_FIELD_NUMBER: _ClassVar[int]
    name: str
    dir: str
    project_config: ProjectConfig
    Flags: _struct_pb2.Struct
    def __init__(self, name: _Optional[str] = ..., dir: _Optional[str] = ..., project_config: _Optional[_Union[ProjectConfig, _Mapping]] = ..., Flags: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ...) -> None: ...

class CreateModuleResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class ModuleConfigDefaultsRequest(_message.Message):
    __slots__ = ("dir",)
    DIR_FIELD_NUMBER: _ClassVar[int]
    dir: str
    def __init__(self, dir: _Optional[str] = ...) -> None: ...

class ModuleConfigDefaultsResponse(_message.Message):
    __slots__ = ("deploy_dir", "build", "build_lock", "generated_schema_dir", "watch", "language_config")
    DEPLOY_DIR_FIELD_NUMBER: _ClassVar[int]
    BUILD_FIELD_NUMBER: _ClassVar[int]
    BUILD_LOCK_FIELD_NUMBER: _ClassVar[int]
    GENERATED_SCHEMA_DIR_FIELD_NUMBER: _ClassVar[int]
    WATCH_FIELD_NUMBER: _ClassVar[int]
    LANGUAGE_CONFIG_FIELD_NUMBER: _ClassVar[int]
    deploy_dir: str
    build: str
    build_lock: str
    generated_schema_dir: str
    watch: _containers.RepeatedScalarFieldContainer[str]
    language_config: _struct_pb2.Struct
    def __init__(self, deploy_dir: _Optional[str] = ..., build: _Optional[str] = ..., build_lock: _Optional[str] = ..., generated_schema_dir: _Optional[str] = ..., watch: _Optional[_Iterable[str]] = ..., language_config: _Optional[_Union[_struct_pb2.Struct, _Mapping]] = ...) -> None: ...

class DependenciesRequest(_message.Message):
    __slots__ = ("module_config",)
    MODULE_CONFIG_FIELD_NUMBER: _ClassVar[int]
    module_config: ModuleConfig
    def __init__(self, module_config: _Optional[_Union[ModuleConfig, _Mapping]] = ...) -> None: ...

class DependenciesResponse(_message.Message):
    __slots__ = ("modules",)
    MODULES_FIELD_NUMBER: _ClassVar[int]
    modules: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, modules: _Optional[_Iterable[str]] = ...) -> None: ...

class BuildContext(_message.Message):
    __slots__ = ("id", "module_config", "schema", "dependencies", "build_env")
    ID_FIELD_NUMBER: _ClassVar[int]
    MODULE_CONFIG_FIELD_NUMBER: _ClassVar[int]
    SCHEMA_FIELD_NUMBER: _ClassVar[int]
    DEPENDENCIES_FIELD_NUMBER: _ClassVar[int]
    BUILD_ENV_FIELD_NUMBER: _ClassVar[int]
    id: str
    module_config: ModuleConfig
    schema: _schema_pb2.Schema
    dependencies: _containers.RepeatedScalarFieldContainer[str]
    build_env: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, id: _Optional[str] = ..., module_config: _Optional[_Union[ModuleConfig, _Mapping]] = ..., schema: _Optional[_Union[_schema_pb2.Schema, _Mapping]] = ..., dependencies: _Optional[_Iterable[str]] = ..., build_env: _Optional[_Iterable[str]] = ...) -> None: ...

class BuildContextUpdatedRequest(_message.Message):
    __slots__ = ("buildContext",)
    BUILDCONTEXT_FIELD_NUMBER: _ClassVar[int]
    buildContext: BuildContext
    def __init__(self, buildContext: _Optional[_Union[BuildContext, _Mapping]] = ...) -> None: ...

class BuildContextUpdatedResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class Error(_message.Message):
    __slots__ = ("msg", "level", "pos", "type")
    class ErrorLevel(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = ()
        INFO: _ClassVar[Error.ErrorLevel]
        WARN: _ClassVar[Error.ErrorLevel]
        ERROR: _ClassVar[Error.ErrorLevel]
    INFO: Error.ErrorLevel
    WARN: Error.ErrorLevel
    ERROR: Error.ErrorLevel
    class ErrorType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = ()
        FTL: _ClassVar[Error.ErrorType]
        COMPILER: _ClassVar[Error.ErrorType]
    FTL: Error.ErrorType
    COMPILER: Error.ErrorType
    MSG_FIELD_NUMBER: _ClassVar[int]
    LEVEL_FIELD_NUMBER: _ClassVar[int]
    POS_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    msg: str
    level: Error.ErrorLevel
    pos: Position
    type: Error.ErrorType
    def __init__(self, msg: _Optional[str] = ..., level: _Optional[_Union[Error.ErrorLevel, str]] = ..., pos: _Optional[_Union[Position, _Mapping]] = ..., type: _Optional[_Union[Error.ErrorType, str]] = ...) -> None: ...

class Position(_message.Message):
    __slots__ = ("filename", "line", "startColumn", "endColumn")
    FILENAME_FIELD_NUMBER: _ClassVar[int]
    LINE_FIELD_NUMBER: _ClassVar[int]
    STARTCOLUMN_FIELD_NUMBER: _ClassVar[int]
    ENDCOLUMN_FIELD_NUMBER: _ClassVar[int]
    filename: str
    line: int
    startColumn: int
    endColumn: int
    def __init__(self, filename: _Optional[str] = ..., line: _Optional[int] = ..., startColumn: _Optional[int] = ..., endColumn: _Optional[int] = ...) -> None: ...

class ErrorList(_message.Message):
    __slots__ = ("errors",)
    ERRORS_FIELD_NUMBER: _ClassVar[int]
    errors: _containers.RepeatedCompositeFieldContainer[Error]
    def __init__(self, errors: _Optional[_Iterable[_Union[Error, _Mapping]]] = ...) -> None: ...

class BuildRequest(_message.Message):
    __slots__ = ("project_root", "stubs_root", "rebuild_automatically", "build_context")
    PROJECT_ROOT_FIELD_NUMBER: _ClassVar[int]
    STUBS_ROOT_FIELD_NUMBER: _ClassVar[int]
    REBUILD_AUTOMATICALLY_FIELD_NUMBER: _ClassVar[int]
    BUILD_CONTEXT_FIELD_NUMBER: _ClassVar[int]
    project_root: str
    stubs_root: str
    rebuild_automatically: bool
    build_context: BuildContext
    def __init__(self, project_root: _Optional[str] = ..., stubs_root: _Optional[str] = ..., rebuild_automatically: bool = ..., build_context: _Optional[_Union[BuildContext, _Mapping]] = ...) -> None: ...

class AutoRebuildStarted(_message.Message):
    __slots__ = ("context_id",)
    CONTEXT_ID_FIELD_NUMBER: _ClassVar[int]
    context_id: str
    def __init__(self, context_id: _Optional[str] = ...) -> None: ...

class BuildSuccess(_message.Message):
    __slots__ = ("context_id", "is_automatic_rebuild", "module", "deploy", "errors")
    CONTEXT_ID_FIELD_NUMBER: _ClassVar[int]
    IS_AUTOMATIC_REBUILD_FIELD_NUMBER: _ClassVar[int]
    MODULE_FIELD_NUMBER: _ClassVar[int]
    DEPLOY_FIELD_NUMBER: _ClassVar[int]
    ERRORS_FIELD_NUMBER: _ClassVar[int]
    context_id: str
    is_automatic_rebuild: bool
    module: _schema_pb2.Module
    deploy: _containers.RepeatedScalarFieldContainer[str]
    errors: ErrorList
    def __init__(self, context_id: _Optional[str] = ..., is_automatic_rebuild: bool = ..., module: _Optional[_Union[_schema_pb2.Module, _Mapping]] = ..., deploy: _Optional[_Iterable[str]] = ..., errors: _Optional[_Union[ErrorList, _Mapping]] = ...) -> None: ...

class BuildFailure(_message.Message):
    __slots__ = ("context_id", "is_automatic_rebuild", "errors", "invalidate_dependencies")
    CONTEXT_ID_FIELD_NUMBER: _ClassVar[int]
    IS_AUTOMATIC_REBUILD_FIELD_NUMBER: _ClassVar[int]
    ERRORS_FIELD_NUMBER: _ClassVar[int]
    INVALIDATE_DEPENDENCIES_FIELD_NUMBER: _ClassVar[int]
    context_id: str
    is_automatic_rebuild: bool
    errors: ErrorList
    invalidate_dependencies: bool
    def __init__(self, context_id: _Optional[str] = ..., is_automatic_rebuild: bool = ..., errors: _Optional[_Union[ErrorList, _Mapping]] = ..., invalidate_dependencies: bool = ...) -> None: ...

class BuildEvent(_message.Message):
    __slots__ = ("auto_rebuild_started", "build_success", "build_failure")
    AUTO_REBUILD_STARTED_FIELD_NUMBER: _ClassVar[int]
    BUILD_SUCCESS_FIELD_NUMBER: _ClassVar[int]
    BUILD_FAILURE_FIELD_NUMBER: _ClassVar[int]
    auto_rebuild_started: AutoRebuildStarted
    build_success: BuildSuccess
    build_failure: BuildFailure
    def __init__(self, auto_rebuild_started: _Optional[_Union[AutoRebuildStarted, _Mapping]] = ..., build_success: _Optional[_Union[BuildSuccess, _Mapping]] = ..., build_failure: _Optional[_Union[BuildFailure, _Mapping]] = ...) -> None: ...

class GenerateStubsRequest(_message.Message):
    __slots__ = ("dir", "module", "module_config", "native_module_config")
    DIR_FIELD_NUMBER: _ClassVar[int]
    MODULE_FIELD_NUMBER: _ClassVar[int]
    MODULE_CONFIG_FIELD_NUMBER: _ClassVar[int]
    NATIVE_MODULE_CONFIG_FIELD_NUMBER: _ClassVar[int]
    dir: str
    module: _schema_pb2.Module
    module_config: ModuleConfig
    native_module_config: ModuleConfig
    def __init__(self, dir: _Optional[str] = ..., module: _Optional[_Union[_schema_pb2.Module, _Mapping]] = ..., module_config: _Optional[_Union[ModuleConfig, _Mapping]] = ..., native_module_config: _Optional[_Union[ModuleConfig, _Mapping]] = ...) -> None: ...

class GenerateStubsResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class SyncStubReferencesRequest(_message.Message):
    __slots__ = ("module_config", "stubs_root", "modules")
    MODULE_CONFIG_FIELD_NUMBER: _ClassVar[int]
    STUBS_ROOT_FIELD_NUMBER: _ClassVar[int]
    MODULES_FIELD_NUMBER: _ClassVar[int]
    module_config: ModuleConfig
    stubs_root: str
    modules: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, module_config: _Optional[_Union[ModuleConfig, _Mapping]] = ..., stubs_root: _Optional[str] = ..., modules: _Optional[_Iterable[str]] = ...) -> None: ...

class SyncStubReferencesResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...
