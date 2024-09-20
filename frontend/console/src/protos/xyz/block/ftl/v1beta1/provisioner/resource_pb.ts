// @generated by protoc-gen-es v1.10.0 with parameter "target=ts"
// @generated from file xyz/block/ftl/v1beta1/provisioner/resource.proto (package xyz.block.ftl.v1beta1.provisioner, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Message, proto3 } from "@bufbuild/protobuf";

/**
 * Resource is an abstract resource extracted from FTL Schema.
 *
 * @generated from message xyz.block.ftl.v1beta1.provisioner.Resource
 */
export class Resource extends Message<Resource> {
  /**
   * id unique within the module
   *
   * @generated from field: string resource_id = 1;
   */
  resourceId = "";

  /**
   * @generated from oneof xyz.block.ftl.v1beta1.provisioner.Resource.resource
   */
  resource: {
    /**
     * @generated from field: xyz.block.ftl.v1beta1.provisioner.PostgresResource postgres = 102;
     */
    value: PostgresResource;
    case: "postgres";
  } | {
    /**
     * @generated from field: xyz.block.ftl.v1beta1.provisioner.MysqlResource mysql = 103;
     */
    value: MysqlResource;
    case: "mysql";
  } | { case: undefined; value?: undefined } = { case: undefined };

  constructor(data?: PartialMessage<Resource>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1beta1.provisioner.Resource";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "resource_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 102, name: "postgres", kind: "message", T: PostgresResource, oneof: "resource" },
    { no: 103, name: "mysql", kind: "message", T: MysqlResource, oneof: "resource" },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Resource {
    return new Resource().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Resource {
    return new Resource().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Resource {
    return new Resource().fromJsonString(jsonString, options);
  }

  static equals(a: Resource | PlainMessage<Resource> | undefined, b: Resource | PlainMessage<Resource> | undefined): boolean {
    return proto3.util.equals(Resource, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1beta1.provisioner.PostgresResource
 */
export class PostgresResource extends Message<PostgresResource> {
  /**
   * @generated from field: xyz.block.ftl.v1beta1.provisioner.PostgresResource.PostgresResourceOutput output = 1;
   */
  output?: PostgresResource_PostgresResourceOutput;

  constructor(data?: PartialMessage<PostgresResource>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1beta1.provisioner.PostgresResource";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "output", kind: "message", T: PostgresResource_PostgresResourceOutput },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): PostgresResource {
    return new PostgresResource().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): PostgresResource {
    return new PostgresResource().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): PostgresResource {
    return new PostgresResource().fromJsonString(jsonString, options);
  }

  static equals(a: PostgresResource | PlainMessage<PostgresResource> | undefined, b: PostgresResource | PlainMessage<PostgresResource> | undefined): boolean {
    return proto3.util.equals(PostgresResource, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1beta1.provisioner.PostgresResource.PostgresResourceOutput
 */
export class PostgresResource_PostgresResourceOutput extends Message<PostgresResource_PostgresResourceOutput> {
  /**
   * @generated from field: string read_endpoint = 1;
   */
  readEndpoint = "";

  /**
   * @generated from field: string write_endpoint = 2;
   */
  writeEndpoint = "";

  constructor(data?: PartialMessage<PostgresResource_PostgresResourceOutput>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1beta1.provisioner.PostgresResource.PostgresResourceOutput";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "read_endpoint", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "write_endpoint", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): PostgresResource_PostgresResourceOutput {
    return new PostgresResource_PostgresResourceOutput().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): PostgresResource_PostgresResourceOutput {
    return new PostgresResource_PostgresResourceOutput().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): PostgresResource_PostgresResourceOutput {
    return new PostgresResource_PostgresResourceOutput().fromJsonString(jsonString, options);
  }

  static equals(a: PostgresResource_PostgresResourceOutput | PlainMessage<PostgresResource_PostgresResourceOutput> | undefined, b: PostgresResource_PostgresResourceOutput | PlainMessage<PostgresResource_PostgresResourceOutput> | undefined): boolean {
    return proto3.util.equals(PostgresResource_PostgresResourceOutput, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1beta1.provisioner.MysqlResource
 */
export class MysqlResource extends Message<MysqlResource> {
  /**
   * @generated from field: xyz.block.ftl.v1beta1.provisioner.MysqlResource.MysqlResourceOutput output = 1;
   */
  output?: MysqlResource_MysqlResourceOutput;

  constructor(data?: PartialMessage<MysqlResource>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1beta1.provisioner.MysqlResource";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "output", kind: "message", T: MysqlResource_MysqlResourceOutput },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): MysqlResource {
    return new MysqlResource().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): MysqlResource {
    return new MysqlResource().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): MysqlResource {
    return new MysqlResource().fromJsonString(jsonString, options);
  }

  static equals(a: MysqlResource | PlainMessage<MysqlResource> | undefined, b: MysqlResource | PlainMessage<MysqlResource> | undefined): boolean {
    return proto3.util.equals(MysqlResource, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1beta1.provisioner.MysqlResource.MysqlResourceOutput
 */
export class MysqlResource_MysqlResourceOutput extends Message<MysqlResource_MysqlResourceOutput> {
  /**
   * @generated from field: string read_endpoint = 1;
   */
  readEndpoint = "";

  /**
   * @generated from field: string write_endpoint = 2;
   */
  writeEndpoint = "";

  constructor(data?: PartialMessage<MysqlResource_MysqlResourceOutput>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1beta1.provisioner.MysqlResource.MysqlResourceOutput";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "read_endpoint", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "write_endpoint", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): MysqlResource_MysqlResourceOutput {
    return new MysqlResource_MysqlResourceOutput().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): MysqlResource_MysqlResourceOutput {
    return new MysqlResource_MysqlResourceOutput().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): MysqlResource_MysqlResourceOutput {
    return new MysqlResource_MysqlResourceOutput().fromJsonString(jsonString, options);
  }

  static equals(a: MysqlResource_MysqlResourceOutput | PlainMessage<MysqlResource_MysqlResourceOutput> | undefined, b: MysqlResource_MysqlResourceOutput | PlainMessage<MysqlResource_MysqlResourceOutput> | undefined): boolean {
    return proto3.util.equals(MysqlResource_MysqlResourceOutput, a, b);
  }
}
