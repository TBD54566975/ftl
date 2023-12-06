// Copyright 2019, OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// @generated by protoc-gen-es v1.5.0 with parameter "target=ts"
// @generated from file opentelemetry/proto/common/v1/common.proto (package opentelemetry.proto.common.v1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Message, proto3 } from "@bufbuild/protobuf";

/**
 * AnyValue is used to represent any type of attribute value. AnyValue may contain a
 * primitive value such as a string or integer or it may contain an arbitrary nested
 * object containing arrays, key-value lists and primitives.
 *
 * @generated from message opentelemetry.proto.common.v1.AnyValue
 */
export class AnyValue extends Message<AnyValue> {
  /**
   * The value is one of the listed fields. It is valid for all values to be unspecified
   * in which case this AnyValue is considered to be "empty".
   *
   * @generated from oneof opentelemetry.proto.common.v1.AnyValue.value
   */
  value: {
    /**
     * @generated from field: string string_value = 1;
     */
    value: string;
    case: "stringValue";
  } | {
    /**
     * @generated from field: bool bool_value = 2;
     */
    value: boolean;
    case: "boolValue";
  } | {
    /**
     * @generated from field: int64 int_value = 3;
     */
    value: bigint;
    case: "intValue";
  } | {
    /**
     * @generated from field: double double_value = 4;
     */
    value: number;
    case: "doubleValue";
  } | {
    /**
     * @generated from field: opentelemetry.proto.common.v1.ArrayValue array_value = 5;
     */
    value: ArrayValue;
    case: "arrayValue";
  } | {
    /**
     * @generated from field: opentelemetry.proto.common.v1.KeyValueList kvlist_value = 6;
     */
    value: KeyValueList;
    case: "kvlistValue";
  } | {
    /**
     * @generated from field: bytes bytes_value = 7;
     */
    value: Uint8Array;
    case: "bytesValue";
  } | { case: undefined; value?: undefined } = { case: undefined };

  constructor(data?: PartialMessage<AnyValue>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "opentelemetry.proto.common.v1.AnyValue";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "string_value", kind: "scalar", T: 9 /* ScalarType.STRING */, oneof: "value" },
    { no: 2, name: "bool_value", kind: "scalar", T: 8 /* ScalarType.BOOL */, oneof: "value" },
    { no: 3, name: "int_value", kind: "scalar", T: 3 /* ScalarType.INT64 */, oneof: "value" },
    { no: 4, name: "double_value", kind: "scalar", T: 1 /* ScalarType.DOUBLE */, oneof: "value" },
    { no: 5, name: "array_value", kind: "message", T: ArrayValue, oneof: "value" },
    { no: 6, name: "kvlist_value", kind: "message", T: KeyValueList, oneof: "value" },
    { no: 7, name: "bytes_value", kind: "scalar", T: 12 /* ScalarType.BYTES */, oneof: "value" },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): AnyValue {
    return new AnyValue().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): AnyValue {
    return new AnyValue().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): AnyValue {
    return new AnyValue().fromJsonString(jsonString, options);
  }

  static equals(a: AnyValue | PlainMessage<AnyValue> | undefined, b: AnyValue | PlainMessage<AnyValue> | undefined): boolean {
    return proto3.util.equals(AnyValue, a, b);
  }
}

/**
 * ArrayValue is a list of AnyValue messages. We need ArrayValue as a message
 * since oneof in AnyValue does not allow repeated fields.
 *
 * @generated from message opentelemetry.proto.common.v1.ArrayValue
 */
export class ArrayValue extends Message<ArrayValue> {
  /**
   * Array of values. The array may be empty (contain 0 elements).
   *
   * @generated from field: repeated opentelemetry.proto.common.v1.AnyValue values = 1;
   */
  values: AnyValue[] = [];

  constructor(data?: PartialMessage<ArrayValue>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "opentelemetry.proto.common.v1.ArrayValue";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "values", kind: "message", T: AnyValue, repeated: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ArrayValue {
    return new ArrayValue().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ArrayValue {
    return new ArrayValue().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ArrayValue {
    return new ArrayValue().fromJsonString(jsonString, options);
  }

  static equals(a: ArrayValue | PlainMessage<ArrayValue> | undefined, b: ArrayValue | PlainMessage<ArrayValue> | undefined): boolean {
    return proto3.util.equals(ArrayValue, a, b);
  }
}

/**
 * KeyValueList is a list of KeyValue messages. We need KeyValueList as a message
 * since `oneof` in AnyValue does not allow repeated fields. Everywhere else where we need
 * a list of KeyValue messages (e.g. in Span) we use `repeated KeyValue` directly to
 * avoid unnecessary extra wrapping (which slows down the protocol). The 2 approaches
 * are semantically equivalent.
 *
 * @generated from message opentelemetry.proto.common.v1.KeyValueList
 */
export class KeyValueList extends Message<KeyValueList> {
  /**
   * A collection of key/value pairs of key-value pairs. The list may be empty (may
   * contain 0 elements).
   * The keys MUST be unique (it is not allowed to have more than one
   * value with the same key).
   *
   * @generated from field: repeated opentelemetry.proto.common.v1.KeyValue values = 1;
   */
  values: KeyValue[] = [];

  constructor(data?: PartialMessage<KeyValueList>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "opentelemetry.proto.common.v1.KeyValueList";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "values", kind: "message", T: KeyValue, repeated: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): KeyValueList {
    return new KeyValueList().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): KeyValueList {
    return new KeyValueList().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): KeyValueList {
    return new KeyValueList().fromJsonString(jsonString, options);
  }

  static equals(a: KeyValueList | PlainMessage<KeyValueList> | undefined, b: KeyValueList | PlainMessage<KeyValueList> | undefined): boolean {
    return proto3.util.equals(KeyValueList, a, b);
  }
}

/**
 * KeyValue is a key-value pair that is used to store Span attributes, Link
 * attributes, etc.
 *
 * @generated from message opentelemetry.proto.common.v1.KeyValue
 */
export class KeyValue extends Message<KeyValue> {
  /**
   * @generated from field: string key = 1;
   */
  key = "";

  /**
   * @generated from field: opentelemetry.proto.common.v1.AnyValue value = 2;
   */
  value?: AnyValue;

  constructor(data?: PartialMessage<KeyValue>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "opentelemetry.proto.common.v1.KeyValue";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "key", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "value", kind: "message", T: AnyValue },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): KeyValue {
    return new KeyValue().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): KeyValue {
    return new KeyValue().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): KeyValue {
    return new KeyValue().fromJsonString(jsonString, options);
  }

  static equals(a: KeyValue | PlainMessage<KeyValue> | undefined, b: KeyValue | PlainMessage<KeyValue> | undefined): boolean {
    return proto3.util.equals(KeyValue, a, b);
  }
}

/**
 * InstrumentationScope is a message representing the instrumentation scope information
 * such as the fully qualified name and version.
 *
 * @generated from message opentelemetry.proto.common.v1.InstrumentationScope
 */
export class InstrumentationScope extends Message<InstrumentationScope> {
  /**
   * An empty instrumentation scope name means the name is unknown.
   *
   * @generated from field: string name = 1;
   */
  name = "";

  /**
   * @generated from field: string version = 2;
   */
  version = "";

  /**
   * Additional attributes that describe the scope. [Optional].
   * Attribute keys MUST be unique (it is not allowed to have more than one
   * attribute with the same key).
   *
   * @generated from field: repeated opentelemetry.proto.common.v1.KeyValue attributes = 3;
   */
  attributes: KeyValue[] = [];

  /**
   * @generated from field: uint32 dropped_attributes_count = 4;
   */
  droppedAttributesCount = 0;

  constructor(data?: PartialMessage<InstrumentationScope>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "opentelemetry.proto.common.v1.InstrumentationScope";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "version", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "attributes", kind: "message", T: KeyValue, repeated: true },
    { no: 4, name: "dropped_attributes_count", kind: "scalar", T: 13 /* ScalarType.UINT32 */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): InstrumentationScope {
    return new InstrumentationScope().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): InstrumentationScope {
    return new InstrumentationScope().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): InstrumentationScope {
    return new InstrumentationScope().fromJsonString(jsonString, options);
  }

  static equals(a: InstrumentationScope | PlainMessage<InstrumentationScope> | undefined, b: InstrumentationScope | PlainMessage<InstrumentationScope> | undefined): boolean {
    return proto3.util.equals(InstrumentationScope, a, b);
  }
}

