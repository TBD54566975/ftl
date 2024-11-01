// @generated by protoc-gen-es v1.10.0 with parameter "target=ts"
// @generated from file xyz/block/ftl/v1/admin.proto (package xyz.block.ftl.v1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Message, proto3 } from "@bufbuild/protobuf";

/**
 * @generated from enum xyz.block.ftl.v1.ConfigProvider
 */
export enum ConfigProvider {
  /**
   * Write values inline in the configuration file.
   *
   * @generated from enum value: CONFIG_INLINE = 0;
   */
  CONFIG_INLINE = 0,

  /**
   * Print configuration as environment variables.
   *
   * @generated from enum value: CONFIG_ENVAR = 1;
   */
  CONFIG_ENVAR = 1,

  /**
   * Use the database as a configuration store.
   *
   * @generated from enum value: CONFIG_DB = 2;
   */
  CONFIG_DB = 2,
}
// Retrieve enum metadata with: proto3.getEnumType(ConfigProvider)
proto3.util.setEnumType(ConfigProvider, "xyz.block.ftl.v1.ConfigProvider", [
  { no: 0, name: "CONFIG_INLINE" },
  { no: 1, name: "CONFIG_ENVAR" },
  { no: 2, name: "CONFIG_DB" },
]);

/**
 * @generated from enum xyz.block.ftl.v1.SecretProvider
 */
export enum SecretProvider {
  /**
   * Write values inline in the configuration file.
   *
   * @generated from enum value: SECRET_INLINE = 0;
   */
  SECRET_INLINE = 0,

  /**
   * Print configuration as environment variables.
   *
   * @generated from enum value: SECRET_ENVAR = 1;
   */
  SECRET_ENVAR = 1,

  /**
   * Write to the system keychain.
   *
   * @generated from enum value: SECRET_KEYCHAIN = 2;
   */
  SECRET_KEYCHAIN = 2,

  /**
   * Store a secret in the 1Password vault.
   *
   * @generated from enum value: SECRET_OP = 3;
   */
  SECRET_OP = 3,

  /**
   * Store a secret in the AWS Secrets Manager.
   *
   * @generated from enum value: SECRET_ASM = 4;
   */
  SECRET_ASM = 4,
}
// Retrieve enum metadata with: proto3.getEnumType(SecretProvider)
proto3.util.setEnumType(SecretProvider, "xyz.block.ftl.v1.SecretProvider", [
  { no: 0, name: "SECRET_INLINE" },
  { no: 1, name: "SECRET_ENVAR" },
  { no: 2, name: "SECRET_KEYCHAIN" },
  { no: 3, name: "SECRET_OP" },
  { no: 4, name: "SECRET_ASM" },
]);

/**
 * @generated from message xyz.block.ftl.v1.ConfigRef
 */
export class ConfigRef extends Message<ConfigRef> {
  /**
   * @generated from field: optional string module = 1;
   */
  module?: string;

  /**
   * @generated from field: string name = 2;
   */
  name = "";

  constructor(data?: PartialMessage<ConfigRef>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1.ConfigRef";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "module", kind: "scalar", T: 9 /* ScalarType.STRING */, opt: true },
    { no: 2, name: "name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ConfigRef {
    return new ConfigRef().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ConfigRef {
    return new ConfigRef().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ConfigRef {
    return new ConfigRef().fromJsonString(jsonString, options);
  }

  static equals(a: ConfigRef | PlainMessage<ConfigRef> | undefined, b: ConfigRef | PlainMessage<ConfigRef> | undefined): boolean {
    return proto3.util.equals(ConfigRef, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1.ListConfigRequest
 */
export class ListConfigRequest extends Message<ListConfigRequest> {
  /**
   * @generated from field: optional string module = 1;
   */
  module?: string;

  /**
   * @generated from field: optional bool include_values = 2;
   */
  includeValues?: boolean;

  /**
   * @generated from field: optional xyz.block.ftl.v1.ConfigProvider provider = 3;
   */
  provider?: ConfigProvider;

  constructor(data?: PartialMessage<ListConfigRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1.ListConfigRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "module", kind: "scalar", T: 9 /* ScalarType.STRING */, opt: true },
    { no: 2, name: "include_values", kind: "scalar", T: 8 /* ScalarType.BOOL */, opt: true },
    { no: 3, name: "provider", kind: "enum", T: proto3.getEnumType(ConfigProvider), opt: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ListConfigRequest {
    return new ListConfigRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ListConfigRequest {
    return new ListConfigRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ListConfigRequest {
    return new ListConfigRequest().fromJsonString(jsonString, options);
  }

  static equals(a: ListConfigRequest | PlainMessage<ListConfigRequest> | undefined, b: ListConfigRequest | PlainMessage<ListConfigRequest> | undefined): boolean {
    return proto3.util.equals(ListConfigRequest, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1.ListConfigResponse
 */
export class ListConfigResponse extends Message<ListConfigResponse> {
  /**
   * @generated from field: repeated xyz.block.ftl.v1.ListConfigResponse.Config configs = 1;
   */
  configs: ListConfigResponse_Config[] = [];

  constructor(data?: PartialMessage<ListConfigResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1.ListConfigResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "configs", kind: "message", T: ListConfigResponse_Config, repeated: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ListConfigResponse {
    return new ListConfigResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ListConfigResponse {
    return new ListConfigResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ListConfigResponse {
    return new ListConfigResponse().fromJsonString(jsonString, options);
  }

  static equals(a: ListConfigResponse | PlainMessage<ListConfigResponse> | undefined, b: ListConfigResponse | PlainMessage<ListConfigResponse> | undefined): boolean {
    return proto3.util.equals(ListConfigResponse, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1.ListConfigResponse.Config
 */
export class ListConfigResponse_Config extends Message<ListConfigResponse_Config> {
  /**
   * @generated from field: string refPath = 1;
   */
  refPath = "";

  /**
   * @generated from field: optional bytes value = 2;
   */
  value?: Uint8Array;

  constructor(data?: PartialMessage<ListConfigResponse_Config>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1.ListConfigResponse.Config";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "refPath", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "value", kind: "scalar", T: 12 /* ScalarType.BYTES */, opt: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ListConfigResponse_Config {
    return new ListConfigResponse_Config().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ListConfigResponse_Config {
    return new ListConfigResponse_Config().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ListConfigResponse_Config {
    return new ListConfigResponse_Config().fromJsonString(jsonString, options);
  }

  static equals(a: ListConfigResponse_Config | PlainMessage<ListConfigResponse_Config> | undefined, b: ListConfigResponse_Config | PlainMessage<ListConfigResponse_Config> | undefined): boolean {
    return proto3.util.equals(ListConfigResponse_Config, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1.GetConfigRequest
 */
export class GetConfigRequest extends Message<GetConfigRequest> {
  /**
   * @generated from field: xyz.block.ftl.v1.ConfigRef ref = 1;
   */
  ref?: ConfigRef;

  constructor(data?: PartialMessage<GetConfigRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1.GetConfigRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "ref", kind: "message", T: ConfigRef },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetConfigRequest {
    return new GetConfigRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetConfigRequest {
    return new GetConfigRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetConfigRequest {
    return new GetConfigRequest().fromJsonString(jsonString, options);
  }

  static equals(a: GetConfigRequest | PlainMessage<GetConfigRequest> | undefined, b: GetConfigRequest | PlainMessage<GetConfigRequest> | undefined): boolean {
    return proto3.util.equals(GetConfigRequest, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1.GetConfigResponse
 */
export class GetConfigResponse extends Message<GetConfigResponse> {
  /**
   * @generated from field: bytes value = 1;
   */
  value = new Uint8Array(0);

  constructor(data?: PartialMessage<GetConfigResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1.GetConfigResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "value", kind: "scalar", T: 12 /* ScalarType.BYTES */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetConfigResponse {
    return new GetConfigResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetConfigResponse {
    return new GetConfigResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetConfigResponse {
    return new GetConfigResponse().fromJsonString(jsonString, options);
  }

  static equals(a: GetConfigResponse | PlainMessage<GetConfigResponse> | undefined, b: GetConfigResponse | PlainMessage<GetConfigResponse> | undefined): boolean {
    return proto3.util.equals(GetConfigResponse, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1.SetConfigRequest
 */
export class SetConfigRequest extends Message<SetConfigRequest> {
  /**
   * @generated from field: optional xyz.block.ftl.v1.ConfigProvider provider = 1;
   */
  provider?: ConfigProvider;

  /**
   * @generated from field: xyz.block.ftl.v1.ConfigRef ref = 2;
   */
  ref?: ConfigRef;

  /**
   * @generated from field: bytes value = 3;
   */
  value = new Uint8Array(0);

  constructor(data?: PartialMessage<SetConfigRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1.SetConfigRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "provider", kind: "enum", T: proto3.getEnumType(ConfigProvider), opt: true },
    { no: 2, name: "ref", kind: "message", T: ConfigRef },
    { no: 3, name: "value", kind: "scalar", T: 12 /* ScalarType.BYTES */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): SetConfigRequest {
    return new SetConfigRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): SetConfigRequest {
    return new SetConfigRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): SetConfigRequest {
    return new SetConfigRequest().fromJsonString(jsonString, options);
  }

  static equals(a: SetConfigRequest | PlainMessage<SetConfigRequest> | undefined, b: SetConfigRequest | PlainMessage<SetConfigRequest> | undefined): boolean {
    return proto3.util.equals(SetConfigRequest, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1.SetConfigResponse
 */
export class SetConfigResponse extends Message<SetConfigResponse> {
  constructor(data?: PartialMessage<SetConfigResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1.SetConfigResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): SetConfigResponse {
    return new SetConfigResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): SetConfigResponse {
    return new SetConfigResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): SetConfigResponse {
    return new SetConfigResponse().fromJsonString(jsonString, options);
  }

  static equals(a: SetConfigResponse | PlainMessage<SetConfigResponse> | undefined, b: SetConfigResponse | PlainMessage<SetConfigResponse> | undefined): boolean {
    return proto3.util.equals(SetConfigResponse, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1.UnsetConfigRequest
 */
export class UnsetConfigRequest extends Message<UnsetConfigRequest> {
  /**
   * @generated from field: optional xyz.block.ftl.v1.ConfigProvider provider = 1;
   */
  provider?: ConfigProvider;

  /**
   * @generated from field: xyz.block.ftl.v1.ConfigRef ref = 2;
   */
  ref?: ConfigRef;

  constructor(data?: PartialMessage<UnsetConfigRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1.UnsetConfigRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "provider", kind: "enum", T: proto3.getEnumType(ConfigProvider), opt: true },
    { no: 2, name: "ref", kind: "message", T: ConfigRef },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): UnsetConfigRequest {
    return new UnsetConfigRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): UnsetConfigRequest {
    return new UnsetConfigRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): UnsetConfigRequest {
    return new UnsetConfigRequest().fromJsonString(jsonString, options);
  }

  static equals(a: UnsetConfigRequest | PlainMessage<UnsetConfigRequest> | undefined, b: UnsetConfigRequest | PlainMessage<UnsetConfigRequest> | undefined): boolean {
    return proto3.util.equals(UnsetConfigRequest, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1.UnsetConfigResponse
 */
export class UnsetConfigResponse extends Message<UnsetConfigResponse> {
  constructor(data?: PartialMessage<UnsetConfigResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1.UnsetConfigResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): UnsetConfigResponse {
    return new UnsetConfigResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): UnsetConfigResponse {
    return new UnsetConfigResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): UnsetConfigResponse {
    return new UnsetConfigResponse().fromJsonString(jsonString, options);
  }

  static equals(a: UnsetConfigResponse | PlainMessage<UnsetConfigResponse> | undefined, b: UnsetConfigResponse | PlainMessage<UnsetConfigResponse> | undefined): boolean {
    return proto3.util.equals(UnsetConfigResponse, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1.ListSecretsRequest
 */
export class ListSecretsRequest extends Message<ListSecretsRequest> {
  /**
   * @generated from field: optional string module = 1;
   */
  module?: string;

  /**
   * @generated from field: optional bool include_values = 2;
   */
  includeValues?: boolean;

  /**
   * @generated from field: optional xyz.block.ftl.v1.SecretProvider provider = 3;
   */
  provider?: SecretProvider;

  constructor(data?: PartialMessage<ListSecretsRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1.ListSecretsRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "module", kind: "scalar", T: 9 /* ScalarType.STRING */, opt: true },
    { no: 2, name: "include_values", kind: "scalar", T: 8 /* ScalarType.BOOL */, opt: true },
    { no: 3, name: "provider", kind: "enum", T: proto3.getEnumType(SecretProvider), opt: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ListSecretsRequest {
    return new ListSecretsRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ListSecretsRequest {
    return new ListSecretsRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ListSecretsRequest {
    return new ListSecretsRequest().fromJsonString(jsonString, options);
  }

  static equals(a: ListSecretsRequest | PlainMessage<ListSecretsRequest> | undefined, b: ListSecretsRequest | PlainMessage<ListSecretsRequest> | undefined): boolean {
    return proto3.util.equals(ListSecretsRequest, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1.ListSecretsResponse
 */
export class ListSecretsResponse extends Message<ListSecretsResponse> {
  /**
   * @generated from field: repeated xyz.block.ftl.v1.ListSecretsResponse.Secret secrets = 1;
   */
  secrets: ListSecretsResponse_Secret[] = [];

  constructor(data?: PartialMessage<ListSecretsResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1.ListSecretsResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "secrets", kind: "message", T: ListSecretsResponse_Secret, repeated: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ListSecretsResponse {
    return new ListSecretsResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ListSecretsResponse {
    return new ListSecretsResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ListSecretsResponse {
    return new ListSecretsResponse().fromJsonString(jsonString, options);
  }

  static equals(a: ListSecretsResponse | PlainMessage<ListSecretsResponse> | undefined, b: ListSecretsResponse | PlainMessage<ListSecretsResponse> | undefined): boolean {
    return proto3.util.equals(ListSecretsResponse, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1.ListSecretsResponse.Secret
 */
export class ListSecretsResponse_Secret extends Message<ListSecretsResponse_Secret> {
  /**
   * @generated from field: string refPath = 1;
   */
  refPath = "";

  /**
   * @generated from field: optional bytes value = 2;
   */
  value?: Uint8Array;

  constructor(data?: PartialMessage<ListSecretsResponse_Secret>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1.ListSecretsResponse.Secret";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "refPath", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "value", kind: "scalar", T: 12 /* ScalarType.BYTES */, opt: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ListSecretsResponse_Secret {
    return new ListSecretsResponse_Secret().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ListSecretsResponse_Secret {
    return new ListSecretsResponse_Secret().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ListSecretsResponse_Secret {
    return new ListSecretsResponse_Secret().fromJsonString(jsonString, options);
  }

  static equals(a: ListSecretsResponse_Secret | PlainMessage<ListSecretsResponse_Secret> | undefined, b: ListSecretsResponse_Secret | PlainMessage<ListSecretsResponse_Secret> | undefined): boolean {
    return proto3.util.equals(ListSecretsResponse_Secret, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1.GetSecretRequest
 */
export class GetSecretRequest extends Message<GetSecretRequest> {
  /**
   * @generated from field: xyz.block.ftl.v1.ConfigRef ref = 1;
   */
  ref?: ConfigRef;

  constructor(data?: PartialMessage<GetSecretRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1.GetSecretRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "ref", kind: "message", T: ConfigRef },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetSecretRequest {
    return new GetSecretRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetSecretRequest {
    return new GetSecretRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetSecretRequest {
    return new GetSecretRequest().fromJsonString(jsonString, options);
  }

  static equals(a: GetSecretRequest | PlainMessage<GetSecretRequest> | undefined, b: GetSecretRequest | PlainMessage<GetSecretRequest> | undefined): boolean {
    return proto3.util.equals(GetSecretRequest, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1.GetSecretResponse
 */
export class GetSecretResponse extends Message<GetSecretResponse> {
  /**
   * @generated from field: bytes value = 1;
   */
  value = new Uint8Array(0);

  constructor(data?: PartialMessage<GetSecretResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1.GetSecretResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "value", kind: "scalar", T: 12 /* ScalarType.BYTES */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetSecretResponse {
    return new GetSecretResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetSecretResponse {
    return new GetSecretResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetSecretResponse {
    return new GetSecretResponse().fromJsonString(jsonString, options);
  }

  static equals(a: GetSecretResponse | PlainMessage<GetSecretResponse> | undefined, b: GetSecretResponse | PlainMessage<GetSecretResponse> | undefined): boolean {
    return proto3.util.equals(GetSecretResponse, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1.SetSecretRequest
 */
export class SetSecretRequest extends Message<SetSecretRequest> {
  /**
   * @generated from field: optional xyz.block.ftl.v1.SecretProvider provider = 1;
   */
  provider?: SecretProvider;

  /**
   * @generated from field: xyz.block.ftl.v1.ConfigRef ref = 2;
   */
  ref?: ConfigRef;

  /**
   * @generated from field: bytes value = 3;
   */
  value = new Uint8Array(0);

  constructor(data?: PartialMessage<SetSecretRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1.SetSecretRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "provider", kind: "enum", T: proto3.getEnumType(SecretProvider), opt: true },
    { no: 2, name: "ref", kind: "message", T: ConfigRef },
    { no: 3, name: "value", kind: "scalar", T: 12 /* ScalarType.BYTES */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): SetSecretRequest {
    return new SetSecretRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): SetSecretRequest {
    return new SetSecretRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): SetSecretRequest {
    return new SetSecretRequest().fromJsonString(jsonString, options);
  }

  static equals(a: SetSecretRequest | PlainMessage<SetSecretRequest> | undefined, b: SetSecretRequest | PlainMessage<SetSecretRequest> | undefined): boolean {
    return proto3.util.equals(SetSecretRequest, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1.SetSecretResponse
 */
export class SetSecretResponse extends Message<SetSecretResponse> {
  constructor(data?: PartialMessage<SetSecretResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1.SetSecretResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): SetSecretResponse {
    return new SetSecretResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): SetSecretResponse {
    return new SetSecretResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): SetSecretResponse {
    return new SetSecretResponse().fromJsonString(jsonString, options);
  }

  static equals(a: SetSecretResponse | PlainMessage<SetSecretResponse> | undefined, b: SetSecretResponse | PlainMessage<SetSecretResponse> | undefined): boolean {
    return proto3.util.equals(SetSecretResponse, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1.UnsetSecretRequest
 */
export class UnsetSecretRequest extends Message<UnsetSecretRequest> {
  /**
   * @generated from field: optional xyz.block.ftl.v1.SecretProvider provider = 1;
   */
  provider?: SecretProvider;

  /**
   * @generated from field: xyz.block.ftl.v1.ConfigRef ref = 2;
   */
  ref?: ConfigRef;

  constructor(data?: PartialMessage<UnsetSecretRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1.UnsetSecretRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "provider", kind: "enum", T: proto3.getEnumType(SecretProvider), opt: true },
    { no: 2, name: "ref", kind: "message", T: ConfigRef },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): UnsetSecretRequest {
    return new UnsetSecretRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): UnsetSecretRequest {
    return new UnsetSecretRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): UnsetSecretRequest {
    return new UnsetSecretRequest().fromJsonString(jsonString, options);
  }

  static equals(a: UnsetSecretRequest | PlainMessage<UnsetSecretRequest> | undefined, b: UnsetSecretRequest | PlainMessage<UnsetSecretRequest> | undefined): boolean {
    return proto3.util.equals(UnsetSecretRequest, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1.UnsetSecretResponse
 */
export class UnsetSecretResponse extends Message<UnsetSecretResponse> {
  constructor(data?: PartialMessage<UnsetSecretResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1.UnsetSecretResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): UnsetSecretResponse {
    return new UnsetSecretResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): UnsetSecretResponse {
    return new UnsetSecretResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): UnsetSecretResponse {
    return new UnsetSecretResponse().fromJsonString(jsonString, options);
  }

  static equals(a: UnsetSecretResponse | PlainMessage<UnsetSecretResponse> | undefined, b: UnsetSecretResponse | PlainMessage<UnsetSecretResponse> | undefined): boolean {
    return proto3.util.equals(UnsetSecretResponse, a, b);
  }
}
