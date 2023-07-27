// @generated by protoc-gen-es v1.3.0 with parameter "target=ts"
// @generated from file xyz/block/ftl/v1/schema/runtime.proto (package xyz.block.ftl.v1.schema, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Message, proto3, Timestamp } from "@bufbuild/protobuf";

/**
 * @generated from enum xyz.block.ftl.v1.schema.Status
 */
export enum Status {
  /**
   * @generated from enum value: OFFLINE = 0;
   */
  OFFLINE = 0,

  /**
   * @generated from enum value: STARTING = 1;
   */
  STARTING = 1,

  /**
   * @generated from enum value: ONLINE = 2;
   */
  ONLINE = 2,

  /**
   * @generated from enum value: STOPPING = 3;
   */
  STOPPING = 3,

  /**
   * @generated from enum value: STOPPED = 4;
   */
  STOPPED = 4,

  /**
   * @generated from enum value: ERRORED = 5;
   */
  ERRORED = 5,
}
// Retrieve enum metadata with: proto3.getEnumType(Status)
proto3.util.setEnumType(Status, "xyz.block.ftl.v1.schema.Status", [
  { no: 0, name: "OFFLINE" },
  { no: 1, name: "STARTING" },
  { no: 2, name: "ONLINE" },
  { no: 3, name: "STOPPING" },
  { no: 4, name: "STOPPED" },
  { no: 5, name: "ERRORED" },
]);

/**
 * @generated from message xyz.block.ftl.v1.schema.ModuleRuntime
 */
export class ModuleRuntime extends Message<ModuleRuntime> {
  /**
   * @generated from field: google.protobuf.Timestamp create_time = 1;
   */
  createTime?: Timestamp;

  /**
   * @generated from field: string language = 2;
   */
  language = "";

  /**
   * @generated from field: int32 min_replicas = 3;
   */
  minReplicas = 0;

  constructor(data?: PartialMessage<ModuleRuntime>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1.schema.ModuleRuntime";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "create_time", kind: "message", T: Timestamp },
    { no: 2, name: "language", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "min_replicas", kind: "scalar", T: 5 /* ScalarType.INT32 */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ModuleRuntime {
    return new ModuleRuntime().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ModuleRuntime {
    return new ModuleRuntime().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ModuleRuntime {
    return new ModuleRuntime().fromJsonString(jsonString, options);
  }

  static equals(a: ModuleRuntime | PlainMessage<ModuleRuntime> | undefined, b: ModuleRuntime | PlainMessage<ModuleRuntime> | undefined): boolean {
    return proto3.util.equals(ModuleRuntime, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1.schema.VerbRuntime
 */
export class VerbRuntime extends Message<VerbRuntime> {
  /**
   * @generated from field: google.protobuf.Timestamp create_time = 1;
   */
  createTime?: Timestamp;

  /**
   * @generated from field: google.protobuf.Timestamp start_time = 2;
   */
  startTime?: Timestamp;

  /**
   * @generated from field: xyz.block.ftl.v1.schema.Status status = 3;
   */
  status = Status.OFFLINE;

  constructor(data?: PartialMessage<VerbRuntime>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1.schema.VerbRuntime";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "create_time", kind: "message", T: Timestamp },
    { no: 2, name: "start_time", kind: "message", T: Timestamp },
    { no: 3, name: "status", kind: "enum", T: proto3.getEnumType(Status) },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): VerbRuntime {
    return new VerbRuntime().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): VerbRuntime {
    return new VerbRuntime().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): VerbRuntime {
    return new VerbRuntime().fromJsonString(jsonString, options);
  }

  static equals(a: VerbRuntime | PlainMessage<VerbRuntime> | undefined, b: VerbRuntime | PlainMessage<VerbRuntime> | undefined): boolean {
    return proto3.util.equals(VerbRuntime, a, b);
  }
}

