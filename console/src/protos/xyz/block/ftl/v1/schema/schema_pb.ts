// @generated by protoc-gen-es v1.2.0 with parameter "target=ts"
// @generated from file xyz/block/ftl/v1/schema/schema.proto (package xyz.block.ftl.v1.schema, syntax proto3)
/* eslint-disable */
// @ts-nocheck

// This file is generated by github.com/TBD54566975/ftl/schema/protobuf.go, DO NOT MODIFY

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Message, proto3, protoInt64 } from "@bufbuild/protobuf";
import { ModuleRuntime, VerbRuntime } from "./runtime_pb.js";

/**
 * @generated from message xyz.block.ftl.v1.schema.Array
 */
export class Array extends Message<Array> {
  /**
   * @generated from field: xyz.block.ftl.v1.schema.Type element = 1;
   */
  element?: Type;

  constructor(data?: PartialMessage<Array>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1.schema.Array";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "element", kind: "message", T: Type },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Array {
    return new Array().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Array {
    return new Array().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Array {
    return new Array().fromJsonString(jsonString, options);
  }

  static equals(a: Array | PlainMessage<Array> | undefined, b: Array | PlainMessage<Array> | undefined): boolean {
    return proto3.util.equals(Array, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1.schema.Bool
 */
export class Bool extends Message<Bool> {
  constructor(data?: PartialMessage<Bool>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1.schema.Bool";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Bool {
    return new Bool().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Bool {
    return new Bool().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Bool {
    return new Bool().fromJsonString(jsonString, options);
  }

  static equals(a: Bool | PlainMessage<Bool> | undefined, b: Bool | PlainMessage<Bool> | undefined): boolean {
    return proto3.util.equals(Bool, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1.schema.Data
 */
export class Data extends Message<Data> {
  /**
   * @generated from field: optional xyz.block.ftl.v1.schema.Position pos = 1;
   */
  pos?: Position;

  /**
   * @generated from field: string name = 2;
   */
  name = "";

  /**
   * @generated from field: repeated xyz.block.ftl.v1.schema.Field fields = 3;
   */
  fields: Field[] = [];

  /**
   * @generated from field: repeated xyz.block.ftl.v1.schema.Metadata metadata = 4;
   */
  metadata: Metadata[] = [];

  /**
   * @generated from field: repeated string comments = 5;
   */
  comments: string[] = [];

  constructor(data?: PartialMessage<Data>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1.schema.Data";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "pos", kind: "message", T: Position, opt: true },
    { no: 2, name: "name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "fields", kind: "message", T: Field, repeated: true },
    { no: 4, name: "metadata", kind: "message", T: Metadata, repeated: true },
    { no: 5, name: "comments", kind: "scalar", T: 9 /* ScalarType.STRING */, repeated: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Data {
    return new Data().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Data {
    return new Data().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Data {
    return new Data().fromJsonString(jsonString, options);
  }

  static equals(a: Data | PlainMessage<Data> | undefined, b: Data | PlainMessage<Data> | undefined): boolean {
    return proto3.util.equals(Data, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1.schema.DataRef
 */
export class DataRef extends Message<DataRef> {
  /**
   * @generated from field: optional xyz.block.ftl.v1.schema.Position pos = 1;
   */
  pos?: Position;

  /**
   * @generated from field: string name = 2;
   */
  name = "";

  /**
   * @generated from field: string module = 3;
   */
  module = "";

  constructor(data?: PartialMessage<DataRef>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1.schema.DataRef";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "pos", kind: "message", T: Position, opt: true },
    { no: 2, name: "name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "module", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): DataRef {
    return new DataRef().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): DataRef {
    return new DataRef().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): DataRef {
    return new DataRef().fromJsonString(jsonString, options);
  }

  static equals(a: DataRef | PlainMessage<DataRef> | undefined, b: DataRef | PlainMessage<DataRef> | undefined): boolean {
    return proto3.util.equals(DataRef, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1.schema.Decl
 */
export class Decl extends Message<Decl> {
  /**
   * @generated from oneof xyz.block.ftl.v1.schema.Decl.value
   */
  value: {
    /**
     * @generated from field: xyz.block.ftl.v1.schema.Data data = 1;
     */
    value: Data;
    case: "data";
  } | {
    /**
     * @generated from field: xyz.block.ftl.v1.schema.Verb verb = 2;
     */
    value: Verb;
    case: "verb";
  } | { case: undefined; value?: undefined } = { case: undefined };

  constructor(data?: PartialMessage<Decl>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1.schema.Decl";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "data", kind: "message", T: Data, oneof: "value" },
    { no: 2, name: "verb", kind: "message", T: Verb, oneof: "value" },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Decl {
    return new Decl().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Decl {
    return new Decl().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Decl {
    return new Decl().fromJsonString(jsonString, options);
  }

  static equals(a: Decl | PlainMessage<Decl> | undefined, b: Decl | PlainMessage<Decl> | undefined): boolean {
    return proto3.util.equals(Decl, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1.schema.Field
 */
export class Field extends Message<Field> {
  /**
   * @generated from field: optional xyz.block.ftl.v1.schema.Position pos = 1;
   */
  pos?: Position;

  /**
   * @generated from field: string name = 2;
   */
  name = "";

  /**
   * @generated from field: repeated string comments = 3;
   */
  comments: string[] = [];

  /**
   * @generated from field: xyz.block.ftl.v1.schema.Type type = 4;
   */
  type?: Type;

  constructor(data?: PartialMessage<Field>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1.schema.Field";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "pos", kind: "message", T: Position, opt: true },
    { no: 2, name: "name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "comments", kind: "scalar", T: 9 /* ScalarType.STRING */, repeated: true },
    { no: 4, name: "type", kind: "message", T: Type },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Field {
    return new Field().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Field {
    return new Field().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Field {
    return new Field().fromJsonString(jsonString, options);
  }

  static equals(a: Field | PlainMessage<Field> | undefined, b: Field | PlainMessage<Field> | undefined): boolean {
    return proto3.util.equals(Field, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1.schema.Float
 */
export class Float extends Message<Float> {
  constructor(data?: PartialMessage<Float>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1.schema.Float";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Float {
    return new Float().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Float {
    return new Float().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Float {
    return new Float().fromJsonString(jsonString, options);
  }

  static equals(a: Float | PlainMessage<Float> | undefined, b: Float | PlainMessage<Float> | undefined): boolean {
    return proto3.util.equals(Float, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1.schema.Int
 */
export class Int extends Message<Int> {
  constructor(data?: PartialMessage<Int>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1.schema.Int";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Int {
    return new Int().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Int {
    return new Int().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Int {
    return new Int().fromJsonString(jsonString, options);
  }

  static equals(a: Int | PlainMessage<Int> | undefined, b: Int | PlainMessage<Int> | undefined): boolean {
    return proto3.util.equals(Int, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1.schema.Map
 */
export class Map extends Message<Map> {
  /**
   * @generated from field: xyz.block.ftl.v1.schema.Type key = 1;
   */
  key?: Type;

  /**
   * @generated from field: xyz.block.ftl.v1.schema.Type value = 2;
   */
  value?: Type;

  constructor(data?: PartialMessage<Map>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1.schema.Map";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "key", kind: "message", T: Type },
    { no: 2, name: "value", kind: "message", T: Type },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Map {
    return new Map().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Map {
    return new Map().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Map {
    return new Map().fromJsonString(jsonString, options);
  }

  static equals(a: Map | PlainMessage<Map> | undefined, b: Map | PlainMessage<Map> | undefined): boolean {
    return proto3.util.equals(Map, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1.schema.Metadata
 */
export class Metadata extends Message<Metadata> {
  /**
   * @generated from oneof xyz.block.ftl.v1.schema.Metadata.value
   */
  value: {
    /**
     * @generated from field: xyz.block.ftl.v1.schema.MetadataCalls calls = 1;
     */
    value: MetadataCalls;
    case: "calls";
  } | { case: undefined; value?: undefined } = { case: undefined };

  constructor(data?: PartialMessage<Metadata>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1.schema.Metadata";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "calls", kind: "message", T: MetadataCalls, oneof: "value" },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Metadata {
    return new Metadata().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Metadata {
    return new Metadata().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Metadata {
    return new Metadata().fromJsonString(jsonString, options);
  }

  static equals(a: Metadata | PlainMessage<Metadata> | undefined, b: Metadata | PlainMessage<Metadata> | undefined): boolean {
    return proto3.util.equals(Metadata, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1.schema.MetadataCalls
 */
export class MetadataCalls extends Message<MetadataCalls> {
  /**
   * @generated from field: optional xyz.block.ftl.v1.schema.Position pos = 1;
   */
  pos?: Position;

  /**
   * @generated from field: repeated xyz.block.ftl.v1.schema.VerbRef calls = 2;
   */
  calls: VerbRef[] = [];

  constructor(data?: PartialMessage<MetadataCalls>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1.schema.MetadataCalls";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "pos", kind: "message", T: Position, opt: true },
    { no: 2, name: "calls", kind: "message", T: VerbRef, repeated: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): MetadataCalls {
    return new MetadataCalls().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): MetadataCalls {
    return new MetadataCalls().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): MetadataCalls {
    return new MetadataCalls().fromJsonString(jsonString, options);
  }

  static equals(a: MetadataCalls | PlainMessage<MetadataCalls> | undefined, b: MetadataCalls | PlainMessage<MetadataCalls> | undefined): boolean {
    return proto3.util.equals(MetadataCalls, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1.schema.Module
 */
export class Module extends Message<Module> {
  /**
   * @generated from field: optional xyz.block.ftl.v1.schema.ModuleRuntime runtime = 31634;
   */
  runtime?: ModuleRuntime;

  /**
   * @generated from field: optional xyz.block.ftl.v1.schema.Position pos = 1;
   */
  pos?: Position;

  /**
   * @generated from field: string name = 2;
   */
  name = "";

  /**
   * @generated from field: repeated string comments = 3;
   */
  comments: string[] = [];

  /**
   * @generated from field: repeated xyz.block.ftl.v1.schema.Decl decls = 4;
   */
  decls: Decl[] = [];

  constructor(data?: PartialMessage<Module>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1.schema.Module";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 31634, name: "runtime", kind: "message", T: ModuleRuntime, opt: true },
    { no: 1, name: "pos", kind: "message", T: Position, opt: true },
    { no: 2, name: "name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "comments", kind: "scalar", T: 9 /* ScalarType.STRING */, repeated: true },
    { no: 4, name: "decls", kind: "message", T: Decl, repeated: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Module {
    return new Module().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Module {
    return new Module().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Module {
    return new Module().fromJsonString(jsonString, options);
  }

  static equals(a: Module | PlainMessage<Module> | undefined, b: Module | PlainMessage<Module> | undefined): boolean {
    return proto3.util.equals(Module, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1.schema.Position
 */
export class Position extends Message<Position> {
  /**
   * @generated from field: string filename = 1;
   */
  filename = "";

  /**
   * @generated from field: int64 line = 2;
   */
  line = protoInt64.zero;

  /**
   * @generated from field: int64 column = 3;
   */
  column = protoInt64.zero;

  constructor(data?: PartialMessage<Position>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1.schema.Position";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "filename", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "line", kind: "scalar", T: 3 /* ScalarType.INT64 */ },
    { no: 3, name: "column", kind: "scalar", T: 3 /* ScalarType.INT64 */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Position {
    return new Position().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Position {
    return new Position().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Position {
    return new Position().fromJsonString(jsonString, options);
  }

  static equals(a: Position | PlainMessage<Position> | undefined, b: Position | PlainMessage<Position> | undefined): boolean {
    return proto3.util.equals(Position, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1.schema.Schema
 */
export class Schema extends Message<Schema> {
  /**
   * @generated from field: optional xyz.block.ftl.v1.schema.Position pos = 1;
   */
  pos?: Position;

  /**
   * @generated from field: repeated xyz.block.ftl.v1.schema.Module modules = 2;
   */
  modules: Module[] = [];

  constructor(data?: PartialMessage<Schema>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1.schema.Schema";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "pos", kind: "message", T: Position, opt: true },
    { no: 2, name: "modules", kind: "message", T: Module, repeated: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Schema {
    return new Schema().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Schema {
    return new Schema().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Schema {
    return new Schema().fromJsonString(jsonString, options);
  }

  static equals(a: Schema | PlainMessage<Schema> | undefined, b: Schema | PlainMessage<Schema> | undefined): boolean {
    return proto3.util.equals(Schema, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1.schema.String
 */
export class String extends Message<String> {
  constructor(data?: PartialMessage<String>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1.schema.String";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): String {
    return new String().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): String {
    return new String().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): String {
    return new String().fromJsonString(jsonString, options);
  }

  static equals(a: String | PlainMessage<String> | undefined, b: String | PlainMessage<String> | undefined): boolean {
    return proto3.util.equals(String, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1.schema.Time
 */
export class Time extends Message<Time> {
  constructor(data?: PartialMessage<Time>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1.schema.Time";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Time {
    return new Time().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Time {
    return new Time().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Time {
    return new Time().fromJsonString(jsonString, options);
  }

  static equals(a: Time | PlainMessage<Time> | undefined, b: Time | PlainMessage<Time> | undefined): boolean {
    return proto3.util.equals(Time, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1.schema.Type
 */
export class Type extends Message<Type> {
  /**
   * @generated from oneof xyz.block.ftl.v1.schema.Type.value
   */
  value: {
    /**
     * @generated from field: xyz.block.ftl.v1.schema.Int int = 1;
     */
    value: Int;
    case: "int";
  } | {
    /**
     * @generated from field: xyz.block.ftl.v1.schema.Float float = 2;
     */
    value: Float;
    case: "float";
  } | {
    /**
     * @generated from field: xyz.block.ftl.v1.schema.String string = 3;
     */
    value: String;
    case: "string";
  } | {
    /**
     * @generated from field: xyz.block.ftl.v1.schema.Bool bool = 4;
     */
    value: Bool;
    case: "bool";
  } | {
    /**
     * @generated from field: xyz.block.ftl.v1.schema.Time time = 5;
     */
    value: Time;
    case: "time";
  } | {
    /**
     * @generated from field: xyz.block.ftl.v1.schema.Array array = 6;
     */
    value: Array;
    case: "array";
  } | {
    /**
     * @generated from field: xyz.block.ftl.v1.schema.Map map = 7;
     */
    value: Map;
    case: "map";
  } | {
    /**
     * @generated from field: xyz.block.ftl.v1.schema.VerbRef verbRef = 8;
     */
    value: VerbRef;
    case: "verbRef";
  } | {
    /**
     * @generated from field: xyz.block.ftl.v1.schema.DataRef dataRef = 9;
     */
    value: DataRef;
    case: "dataRef";
  } | { case: undefined; value?: undefined } = { case: undefined };

  constructor(data?: PartialMessage<Type>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1.schema.Type";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "int", kind: "message", T: Int, oneof: "value" },
    { no: 2, name: "float", kind: "message", T: Float, oneof: "value" },
    { no: 3, name: "string", kind: "message", T: String, oneof: "value" },
    { no: 4, name: "bool", kind: "message", T: Bool, oneof: "value" },
    { no: 5, name: "time", kind: "message", T: Time, oneof: "value" },
    { no: 6, name: "array", kind: "message", T: Array, oneof: "value" },
    { no: 7, name: "map", kind: "message", T: Map, oneof: "value" },
    { no: 8, name: "verbRef", kind: "message", T: VerbRef, oneof: "value" },
    { no: 9, name: "dataRef", kind: "message", T: DataRef, oneof: "value" },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Type {
    return new Type().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Type {
    return new Type().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Type {
    return new Type().fromJsonString(jsonString, options);
  }

  static equals(a: Type | PlainMessage<Type> | undefined, b: Type | PlainMessage<Type> | undefined): boolean {
    return proto3.util.equals(Type, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1.schema.Verb
 */
export class Verb extends Message<Verb> {
  /**
   * @generated from field: optional xyz.block.ftl.v1.schema.VerbRuntime runtime = 31634;
   */
  runtime?: VerbRuntime;

  /**
   * @generated from field: optional xyz.block.ftl.v1.schema.Position pos = 1;
   */
  pos?: Position;

  /**
   * @generated from field: string name = 2;
   */
  name = "";

  /**
   * @generated from field: repeated string comments = 3;
   */
  comments: string[] = [];

  /**
   * @generated from field: xyz.block.ftl.v1.schema.DataRef request = 4;
   */
  request?: DataRef;

  /**
   * @generated from field: xyz.block.ftl.v1.schema.DataRef response = 5;
   */
  response?: DataRef;

  /**
   * @generated from field: repeated xyz.block.ftl.v1.schema.Metadata metadata = 6;
   */
  metadata: Metadata[] = [];

  constructor(data?: PartialMessage<Verb>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1.schema.Verb";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 31634, name: "runtime", kind: "message", T: VerbRuntime, opt: true },
    { no: 1, name: "pos", kind: "message", T: Position, opt: true },
    { no: 2, name: "name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "comments", kind: "scalar", T: 9 /* ScalarType.STRING */, repeated: true },
    { no: 4, name: "request", kind: "message", T: DataRef },
    { no: 5, name: "response", kind: "message", T: DataRef },
    { no: 6, name: "metadata", kind: "message", T: Metadata, repeated: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Verb {
    return new Verb().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Verb {
    return new Verb().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Verb {
    return new Verb().fromJsonString(jsonString, options);
  }

  static equals(a: Verb | PlainMessage<Verb> | undefined, b: Verb | PlainMessage<Verb> | undefined): boolean {
    return proto3.util.equals(Verb, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.v1.schema.VerbRef
 */
export class VerbRef extends Message<VerbRef> {
  /**
   * @generated from field: optional xyz.block.ftl.v1.schema.Position pos = 1;
   */
  pos?: Position;

  /**
   * @generated from field: string name = 2;
   */
  name = "";

  /**
   * @generated from field: string module = 3;
   */
  module = "";

  constructor(data?: PartialMessage<VerbRef>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.v1.schema.VerbRef";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "pos", kind: "message", T: Position, opt: true },
    { no: 2, name: "name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "module", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): VerbRef {
    return new VerbRef().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): VerbRef {
    return new VerbRef().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): VerbRef {
    return new VerbRef().fromJsonString(jsonString, options);
  }

  static equals(a: VerbRef | PlainMessage<VerbRef> | undefined, b: VerbRef | PlainMessage<VerbRef> | undefined): boolean {
    return proto3.util.equals(VerbRef, a, b);
  }
}

