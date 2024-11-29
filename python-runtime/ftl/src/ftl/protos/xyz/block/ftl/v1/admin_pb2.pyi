from xyz.block.ftl.v1 import ftl_pb2 as _ftl_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class ConfigProvider(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    CONFIG_PROVIDER_UNSPECIFIED: _ClassVar[ConfigProvider]
    CONFIG_PROVIDER_INLINE: _ClassVar[ConfigProvider]
    CONFIG_PROVIDER_ENVAR: _ClassVar[ConfigProvider]
    CONFIG_PROVIDER_DB: _ClassVar[ConfigProvider]

class SecretProvider(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    SECRET_PROVIDER_UNSPECIFIED: _ClassVar[SecretProvider]
    SECRET_PROVIDER_INLINE: _ClassVar[SecretProvider]
    SECRET_PROVIDER_ENVAR: _ClassVar[SecretProvider]
    SECRET_PROVIDER_KEYCHAIN: _ClassVar[SecretProvider]
    SECRET_PROVIDER_OP: _ClassVar[SecretProvider]
    SECRET_PROVIDER_ASM: _ClassVar[SecretProvider]
CONFIG_PROVIDER_UNSPECIFIED: ConfigProvider
CONFIG_PROVIDER_INLINE: ConfigProvider
CONFIG_PROVIDER_ENVAR: ConfigProvider
CONFIG_PROVIDER_DB: ConfigProvider
SECRET_PROVIDER_UNSPECIFIED: SecretProvider
SECRET_PROVIDER_INLINE: SecretProvider
SECRET_PROVIDER_ENVAR: SecretProvider
SECRET_PROVIDER_KEYCHAIN: SecretProvider
SECRET_PROVIDER_OP: SecretProvider
SECRET_PROVIDER_ASM: SecretProvider

class ConfigRef(_message.Message):
    __slots__ = ("module", "name")
    MODULE_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    module: str
    name: str
    def __init__(self, module: _Optional[str] = ..., name: _Optional[str] = ...) -> None: ...

class ListConfigRequest(_message.Message):
    __slots__ = ("module", "include_values", "provider")
    MODULE_FIELD_NUMBER: _ClassVar[int]
    INCLUDE_VALUES_FIELD_NUMBER: _ClassVar[int]
    PROVIDER_FIELD_NUMBER: _ClassVar[int]
    module: str
    include_values: bool
    provider: ConfigProvider
    def __init__(self, module: _Optional[str] = ..., include_values: bool = ..., provider: _Optional[_Union[ConfigProvider, str]] = ...) -> None: ...

class ListConfigResponse(_message.Message):
    __slots__ = ("configs",)
    class Config(_message.Message):
        __slots__ = ("ref_path", "value")
        REF_PATH_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        ref_path: str
        value: bytes
        def __init__(self, ref_path: _Optional[str] = ..., value: _Optional[bytes] = ...) -> None: ...
    CONFIGS_FIELD_NUMBER: _ClassVar[int]
    configs: _containers.RepeatedCompositeFieldContainer[ListConfigResponse.Config]
    def __init__(self, configs: _Optional[_Iterable[_Union[ListConfigResponse.Config, _Mapping]]] = ...) -> None: ...

class GetConfigRequest(_message.Message):
    __slots__ = ("ref",)
    REF_FIELD_NUMBER: _ClassVar[int]
    ref: ConfigRef
    def __init__(self, ref: _Optional[_Union[ConfigRef, _Mapping]] = ...) -> None: ...

class GetConfigResponse(_message.Message):
    __slots__ = ("value",)
    VALUE_FIELD_NUMBER: _ClassVar[int]
    value: bytes
    def __init__(self, value: _Optional[bytes] = ...) -> None: ...

class SetConfigRequest(_message.Message):
    __slots__ = ("provider", "ref", "value")
    PROVIDER_FIELD_NUMBER: _ClassVar[int]
    REF_FIELD_NUMBER: _ClassVar[int]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    provider: ConfigProvider
    ref: ConfigRef
    value: bytes
    def __init__(self, provider: _Optional[_Union[ConfigProvider, str]] = ..., ref: _Optional[_Union[ConfigRef, _Mapping]] = ..., value: _Optional[bytes] = ...) -> None: ...

class SetConfigResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class UnsetConfigRequest(_message.Message):
    __slots__ = ("provider", "ref")
    PROVIDER_FIELD_NUMBER: _ClassVar[int]
    REF_FIELD_NUMBER: _ClassVar[int]
    provider: ConfigProvider
    ref: ConfigRef
    def __init__(self, provider: _Optional[_Union[ConfigProvider, str]] = ..., ref: _Optional[_Union[ConfigRef, _Mapping]] = ...) -> None: ...

class UnsetConfigResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class ListSecretsRequest(_message.Message):
    __slots__ = ("module", "include_values", "provider")
    MODULE_FIELD_NUMBER: _ClassVar[int]
    INCLUDE_VALUES_FIELD_NUMBER: _ClassVar[int]
    PROVIDER_FIELD_NUMBER: _ClassVar[int]
    module: str
    include_values: bool
    provider: SecretProvider
    def __init__(self, module: _Optional[str] = ..., include_values: bool = ..., provider: _Optional[_Union[SecretProvider, str]] = ...) -> None: ...

class ListSecretsResponse(_message.Message):
    __slots__ = ("secrets",)
    class Secret(_message.Message):
        __slots__ = ("ref_path", "value")
        REF_PATH_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        ref_path: str
        value: bytes
        def __init__(self, ref_path: _Optional[str] = ..., value: _Optional[bytes] = ...) -> None: ...
    SECRETS_FIELD_NUMBER: _ClassVar[int]
    secrets: _containers.RepeatedCompositeFieldContainer[ListSecretsResponse.Secret]
    def __init__(self, secrets: _Optional[_Iterable[_Union[ListSecretsResponse.Secret, _Mapping]]] = ...) -> None: ...

class GetSecretRequest(_message.Message):
    __slots__ = ("ref",)
    REF_FIELD_NUMBER: _ClassVar[int]
    ref: ConfigRef
    def __init__(self, ref: _Optional[_Union[ConfigRef, _Mapping]] = ...) -> None: ...

class GetSecretResponse(_message.Message):
    __slots__ = ("value",)
    VALUE_FIELD_NUMBER: _ClassVar[int]
    value: bytes
    def __init__(self, value: _Optional[bytes] = ...) -> None: ...

class SetSecretRequest(_message.Message):
    __slots__ = ("provider", "ref", "value")
    PROVIDER_FIELD_NUMBER: _ClassVar[int]
    REF_FIELD_NUMBER: _ClassVar[int]
    VALUE_FIELD_NUMBER: _ClassVar[int]
    provider: SecretProvider
    ref: ConfigRef
    value: bytes
    def __init__(self, provider: _Optional[_Union[SecretProvider, str]] = ..., ref: _Optional[_Union[ConfigRef, _Mapping]] = ..., value: _Optional[bytes] = ...) -> None: ...

class SetSecretResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class UnsetSecretRequest(_message.Message):
    __slots__ = ("provider", "ref")
    PROVIDER_FIELD_NUMBER: _ClassVar[int]
    REF_FIELD_NUMBER: _ClassVar[int]
    provider: SecretProvider
    ref: ConfigRef
    def __init__(self, provider: _Optional[_Union[SecretProvider, str]] = ..., ref: _Optional[_Union[ConfigRef, _Mapping]] = ...) -> None: ...

class UnsetSecretResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...
