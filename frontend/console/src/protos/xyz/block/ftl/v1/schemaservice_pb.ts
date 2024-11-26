// @generated by protoc-gen-es v1.10.0 with parameter "target=ts"
// @generated from file xyz/block/ftl/v1/schemaservice.proto (package xyz.block.ftl.v1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Message, proto3 } from "@bufbuild/protobuf";
import { Module, Schema } from "./schema/schema_pb.js";

/**
 * @generated from enum xyz.block.ftl.v1.DeploymentChangeType
 */
export enum DeploymentChangeType {
  /**
   * @generated from enum value: DEPLOYMENT_ADDED = 0;
   */
  DEPLOYMENT_ADDED = 0,

  /**
   * @generated from enum value: DEPLOYMENT_REMOVED = 1;
   */
  DEPLOYMENT_REMOVED = 1,

  /**
   * @generated from enum value: DEPLOYMENT_CHANGED = 2;
   */
  DEPLOYMENT_CHANGED = 2,
}
// Retrieve enum metadata with: proto3.getEnumType(DeploymentChangeType)
proto3.util.setEnumType(DeploymentChangeType, "xyz.block.ftl.v1.DeploymentChangeType", [
  { no: 0, name: "DEPLOYMENT_ADDED" },
  { no: 1, name: "DEPLOYMENT_REMOVED" },
  { no: 2, name: "DEPLOYMENT_CHANGED" },
]);

/**
 * @generated from message xyz.block.ftl.v1.GetSchemaRequest
 */
export class GetSchemaRequest extends Message<GetSchemaRequest> {
  constructor(data?: PartialMessage<GetSchemaRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1.GetSchemaRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetSchemaRequest {
    return new GetSchemaRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetSchemaRequest {
    return new GetSchemaRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetSchemaRequest {
    return new GetSchemaRequest().fromJsonString(jsonString, options);
  }

  static equals(a: GetSchemaRequest | PlainMessage<GetSchemaRequest> | undefined, b: GetSchemaRequest | PlainMessage<GetSchemaRequest> | undefined): boolean {
    return proto3.util.equals(GetSchemaRequest, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1.GetSchemaResponse
 */
export class GetSchemaResponse extends Message<GetSchemaResponse> {
  /**
   * @generated from field: xyz.block.ftl.v1.schema.Schema schema = 1;
   */
  schema?: Schema;

  constructor(data?: PartialMessage<GetSchemaResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1.GetSchemaResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "schema", kind: "message", T: Schema },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetSchemaResponse {
    return new GetSchemaResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetSchemaResponse {
    return new GetSchemaResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetSchemaResponse {
    return new GetSchemaResponse().fromJsonString(jsonString, options);
  }

  static equals(a: GetSchemaResponse | PlainMessage<GetSchemaResponse> | undefined, b: GetSchemaResponse | PlainMessage<GetSchemaResponse> | undefined): boolean {
    return proto3.util.equals(GetSchemaResponse, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1.PullSchemaRequest
 */
export class PullSchemaRequest extends Message<PullSchemaRequest> {
  constructor(data?: PartialMessage<PullSchemaRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1.PullSchemaRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): PullSchemaRequest {
    return new PullSchemaRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): PullSchemaRequest {
    return new PullSchemaRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): PullSchemaRequest {
    return new PullSchemaRequest().fromJsonString(jsonString, options);
  }

  static equals(a: PullSchemaRequest | PlainMessage<PullSchemaRequest> | undefined, b: PullSchemaRequest | PlainMessage<PullSchemaRequest> | undefined): boolean {
    return proto3.util.equals(PullSchemaRequest, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1.PullSchemaResponse
 */
export class PullSchemaResponse extends Message<PullSchemaResponse> {
  /**
   * Will not be set for builtin modules.
   *
   * @generated from field: optional string deployment_key = 1;
   */
  deploymentKey?: string;

  /**
   * @generated from field: string module_name = 2;
   */
  moduleName = "";

  /**
   * For deletes this will not be present.
   *
   * @generated from field: optional xyz.block.ftl.v1.schema.Module schema = 4;
   */
  schema?: Module;

  /**
   * If true there are more schema changes immediately following this one as part of the initial batch.
   * If false this is the last schema change in the initial batch, but others may follow later.
   *
   * @generated from field: bool more = 3;
   */
  more = false;

  /**
   * @generated from field: xyz.block.ftl.v1.DeploymentChangeType change_type = 5;
   */
  changeType = DeploymentChangeType.DEPLOYMENT_ADDED;

  /**
   * If this is true then the module was removed as well as the deployment. This is only set for DEPLOYMENT_REMOVED.
   *
   * @generated from field: bool module_removed = 6;
   */
  moduleRemoved = false;

  constructor(data?: PartialMessage<PullSchemaResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1.PullSchemaResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "deployment_key", kind: "scalar", T: 9 /* ScalarType.STRING */, opt: true },
    { no: 2, name: "module_name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 4, name: "schema", kind: "message", T: Module, opt: true },
    { no: 3, name: "more", kind: "scalar", T: 8 /* ScalarType.BOOL */ },
    { no: 5, name: "change_type", kind: "enum", T: proto3.getEnumType(DeploymentChangeType) },
    { no: 6, name: "module_removed", kind: "scalar", T: 8 /* ScalarType.BOOL */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): PullSchemaResponse {
    return new PullSchemaResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): PullSchemaResponse {
    return new PullSchemaResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): PullSchemaResponse {
    return new PullSchemaResponse().fromJsonString(jsonString, options);
  }

  static equals(a: PullSchemaResponse | PlainMessage<PullSchemaResponse> | undefined, b: PullSchemaResponse | PlainMessage<PullSchemaResponse> | undefined): boolean {
    return proto3.util.equals(PullSchemaResponse, a, b);
  }
}

