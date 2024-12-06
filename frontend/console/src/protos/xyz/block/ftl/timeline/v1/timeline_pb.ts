// @generated by protoc-gen-es v1.10.0 with parameter "target=ts"
// @generated from file xyz/block/ftl/timeline/v1/timeline.proto (package xyz.block.ftl.timeline.v1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Message, proto3, protoInt64, Timestamp } from "@bufbuild/protobuf";
import { AsyncExecuteEvent, CallEvent, CronScheduledEvent, DeploymentCreatedEvent, DeploymentUpdatedEvent, Event, EventType, IngressEvent, LogEvent, LogLevel, PubSubConsumeEvent, PubSubPublishEvent } from "./event_pb.js";

/**
 * @generated from message xyz.block.ftl.timeline.v1.GetTimelineRequest
 */
export class GetTimelineRequest extends Message<GetTimelineRequest> {
  /**
   * @generated from field: repeated xyz.block.ftl.timeline.v1.GetTimelineRequest.Filter filters = 1;
   */
  filters: GetTimelineRequest_Filter[] = [];

  /**
   * @generated from field: int32 limit = 2;
   */
  limit = 0;

  /**
   * @generated from field: xyz.block.ftl.timeline.v1.GetTimelineRequest.Order order = 3;
   */
  order = GetTimelineRequest_Order.UNSPECIFIED;

  constructor(data?: PartialMessage<GetTimelineRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.timeline.v1.GetTimelineRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "filters", kind: "message", T: GetTimelineRequest_Filter, repeated: true },
    { no: 2, name: "limit", kind: "scalar", T: 5 /* ScalarType.INT32 */ },
    { no: 3, name: "order", kind: "enum", T: proto3.getEnumType(GetTimelineRequest_Order) },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetTimelineRequest {
    return new GetTimelineRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetTimelineRequest {
    return new GetTimelineRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetTimelineRequest {
    return new GetTimelineRequest().fromJsonString(jsonString, options);
  }

  static equals(a: GetTimelineRequest | PlainMessage<GetTimelineRequest> | undefined, b: GetTimelineRequest | PlainMessage<GetTimelineRequest> | undefined): boolean {
    return proto3.util.equals(GetTimelineRequest, a, b);
  }
}

/**
 * @generated from enum xyz.block.ftl.timeline.v1.GetTimelineRequest.Order
 */
export enum GetTimelineRequest_Order {
  /**
   * @generated from enum value: ORDER_UNSPECIFIED = 0;
   */
  UNSPECIFIED = 0,

  /**
   * @generated from enum value: ORDER_ASC = 1;
   */
  ASC = 1,

  /**
   * @generated from enum value: ORDER_DESC = 2;
   */
  DESC = 2,
}
// Retrieve enum metadata with: proto3.getEnumType(GetTimelineRequest_Order)
proto3.util.setEnumType(GetTimelineRequest_Order, "xyz.block.ftl.timeline.v1.GetTimelineRequest.Order", [
  { no: 0, name: "ORDER_UNSPECIFIED" },
  { no: 1, name: "ORDER_ASC" },
  { no: 2, name: "ORDER_DESC" },
]);

/**
 * Limit the number of events returned.
 *
 * @generated from message xyz.block.ftl.timeline.v1.GetTimelineRequest.LimitFilter
 */
export class GetTimelineRequest_LimitFilter extends Message<GetTimelineRequest_LimitFilter> {
  /**
   * @generated from field: int32 limit = 1;
   */
  limit = 0;

  constructor(data?: PartialMessage<GetTimelineRequest_LimitFilter>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.timeline.v1.GetTimelineRequest.LimitFilter";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "limit", kind: "scalar", T: 5 /* ScalarType.INT32 */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetTimelineRequest_LimitFilter {
    return new GetTimelineRequest_LimitFilter().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetTimelineRequest_LimitFilter {
    return new GetTimelineRequest_LimitFilter().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetTimelineRequest_LimitFilter {
    return new GetTimelineRequest_LimitFilter().fromJsonString(jsonString, options);
  }

  static equals(a: GetTimelineRequest_LimitFilter | PlainMessage<GetTimelineRequest_LimitFilter> | undefined, b: GetTimelineRequest_LimitFilter | PlainMessage<GetTimelineRequest_LimitFilter> | undefined): boolean {
    return proto3.util.equals(GetTimelineRequest_LimitFilter, a, b);
  }
}

/**
 * Filters events by log level.
 *
 * @generated from message xyz.block.ftl.timeline.v1.GetTimelineRequest.LogLevelFilter
 */
export class GetTimelineRequest_LogLevelFilter extends Message<GetTimelineRequest_LogLevelFilter> {
  /**
   * @generated from field: xyz.block.ftl.timeline.v1.LogLevel log_level = 1;
   */
  logLevel = LogLevel.UNSPECIFIED;

  constructor(data?: PartialMessage<GetTimelineRequest_LogLevelFilter>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.timeline.v1.GetTimelineRequest.LogLevelFilter";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "log_level", kind: "enum", T: proto3.getEnumType(LogLevel) },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetTimelineRequest_LogLevelFilter {
    return new GetTimelineRequest_LogLevelFilter().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetTimelineRequest_LogLevelFilter {
    return new GetTimelineRequest_LogLevelFilter().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetTimelineRequest_LogLevelFilter {
    return new GetTimelineRequest_LogLevelFilter().fromJsonString(jsonString, options);
  }

  static equals(a: GetTimelineRequest_LogLevelFilter | PlainMessage<GetTimelineRequest_LogLevelFilter> | undefined, b: GetTimelineRequest_LogLevelFilter | PlainMessage<GetTimelineRequest_LogLevelFilter> | undefined): boolean {
    return proto3.util.equals(GetTimelineRequest_LogLevelFilter, a, b);
  }
}

/**
 * Filters events by deployment key.
 *
 * @generated from message xyz.block.ftl.timeline.v1.GetTimelineRequest.DeploymentFilter
 */
export class GetTimelineRequest_DeploymentFilter extends Message<GetTimelineRequest_DeploymentFilter> {
  /**
   * @generated from field: repeated string deployments = 1;
   */
  deployments: string[] = [];

  constructor(data?: PartialMessage<GetTimelineRequest_DeploymentFilter>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.timeline.v1.GetTimelineRequest.DeploymentFilter";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "deployments", kind: "scalar", T: 9 /* ScalarType.STRING */, repeated: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetTimelineRequest_DeploymentFilter {
    return new GetTimelineRequest_DeploymentFilter().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetTimelineRequest_DeploymentFilter {
    return new GetTimelineRequest_DeploymentFilter().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetTimelineRequest_DeploymentFilter {
    return new GetTimelineRequest_DeploymentFilter().fromJsonString(jsonString, options);
  }

  static equals(a: GetTimelineRequest_DeploymentFilter | PlainMessage<GetTimelineRequest_DeploymentFilter> | undefined, b: GetTimelineRequest_DeploymentFilter | PlainMessage<GetTimelineRequest_DeploymentFilter> | undefined): boolean {
    return proto3.util.equals(GetTimelineRequest_DeploymentFilter, a, b);
  }
}

/**
 * Filters events by request key.
 *
 * @generated from message xyz.block.ftl.timeline.v1.GetTimelineRequest.RequestFilter
 */
export class GetTimelineRequest_RequestFilter extends Message<GetTimelineRequest_RequestFilter> {
  /**
   * @generated from field: repeated string requests = 1;
   */
  requests: string[] = [];

  constructor(data?: PartialMessage<GetTimelineRequest_RequestFilter>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.timeline.v1.GetTimelineRequest.RequestFilter";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "requests", kind: "scalar", T: 9 /* ScalarType.STRING */, repeated: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetTimelineRequest_RequestFilter {
    return new GetTimelineRequest_RequestFilter().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetTimelineRequest_RequestFilter {
    return new GetTimelineRequest_RequestFilter().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetTimelineRequest_RequestFilter {
    return new GetTimelineRequest_RequestFilter().fromJsonString(jsonString, options);
  }

  static equals(a: GetTimelineRequest_RequestFilter | PlainMessage<GetTimelineRequest_RequestFilter> | undefined, b: GetTimelineRequest_RequestFilter | PlainMessage<GetTimelineRequest_RequestFilter> | undefined): boolean {
    return proto3.util.equals(GetTimelineRequest_RequestFilter, a, b);
  }
}

/**
 * Filters events by event type.
 *
 * @generated from message xyz.block.ftl.timeline.v1.GetTimelineRequest.EventTypeFilter
 */
export class GetTimelineRequest_EventTypeFilter extends Message<GetTimelineRequest_EventTypeFilter> {
  /**
   * @generated from field: repeated xyz.block.ftl.timeline.v1.EventType event_types = 1;
   */
  eventTypes: EventType[] = [];

  constructor(data?: PartialMessage<GetTimelineRequest_EventTypeFilter>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.timeline.v1.GetTimelineRequest.EventTypeFilter";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "event_types", kind: "enum", T: proto3.getEnumType(EventType), repeated: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetTimelineRequest_EventTypeFilter {
    return new GetTimelineRequest_EventTypeFilter().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetTimelineRequest_EventTypeFilter {
    return new GetTimelineRequest_EventTypeFilter().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetTimelineRequest_EventTypeFilter {
    return new GetTimelineRequest_EventTypeFilter().fromJsonString(jsonString, options);
  }

  static equals(a: GetTimelineRequest_EventTypeFilter | PlainMessage<GetTimelineRequest_EventTypeFilter> | undefined, b: GetTimelineRequest_EventTypeFilter | PlainMessage<GetTimelineRequest_EventTypeFilter> | undefined): boolean {
    return proto3.util.equals(GetTimelineRequest_EventTypeFilter, a, b);
  }
}

/**
 * Filters events by time.
 *
 * Either end of the time range can be omitted to indicate no bound.
 *
 * @generated from message xyz.block.ftl.timeline.v1.GetTimelineRequest.TimeFilter
 */
export class GetTimelineRequest_TimeFilter extends Message<GetTimelineRequest_TimeFilter> {
  /**
   * @generated from field: optional google.protobuf.Timestamp older_than = 1;
   */
  olderThan?: Timestamp;

  /**
   * @generated from field: optional google.protobuf.Timestamp newer_than = 2;
   */
  newerThan?: Timestamp;

  constructor(data?: PartialMessage<GetTimelineRequest_TimeFilter>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.timeline.v1.GetTimelineRequest.TimeFilter";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "older_than", kind: "message", T: Timestamp, opt: true },
    { no: 2, name: "newer_than", kind: "message", T: Timestamp, opt: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetTimelineRequest_TimeFilter {
    return new GetTimelineRequest_TimeFilter().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetTimelineRequest_TimeFilter {
    return new GetTimelineRequest_TimeFilter().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetTimelineRequest_TimeFilter {
    return new GetTimelineRequest_TimeFilter().fromJsonString(jsonString, options);
  }

  static equals(a: GetTimelineRequest_TimeFilter | PlainMessage<GetTimelineRequest_TimeFilter> | undefined, b: GetTimelineRequest_TimeFilter | PlainMessage<GetTimelineRequest_TimeFilter> | undefined): boolean {
    return proto3.util.equals(GetTimelineRequest_TimeFilter, a, b);
  }
}

/**
 * Filters events by ID.
 *
 * Either end of the ID range can be omitted to indicate no bound.
 *
 * @generated from message xyz.block.ftl.timeline.v1.GetTimelineRequest.IDFilter
 */
export class GetTimelineRequest_IDFilter extends Message<GetTimelineRequest_IDFilter> {
  /**
   * @generated from field: optional int64 lower_than = 1;
   */
  lowerThan?: bigint;

  /**
   * @generated from field: optional int64 higher_than = 2;
   */
  higherThan?: bigint;

  constructor(data?: PartialMessage<GetTimelineRequest_IDFilter>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.timeline.v1.GetTimelineRequest.IDFilter";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "lower_than", kind: "scalar", T: 3 /* ScalarType.INT64 */, opt: true },
    { no: 2, name: "higher_than", kind: "scalar", T: 3 /* ScalarType.INT64 */, opt: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetTimelineRequest_IDFilter {
    return new GetTimelineRequest_IDFilter().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetTimelineRequest_IDFilter {
    return new GetTimelineRequest_IDFilter().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetTimelineRequest_IDFilter {
    return new GetTimelineRequest_IDFilter().fromJsonString(jsonString, options);
  }

  static equals(a: GetTimelineRequest_IDFilter | PlainMessage<GetTimelineRequest_IDFilter> | undefined, b: GetTimelineRequest_IDFilter | PlainMessage<GetTimelineRequest_IDFilter> | undefined): boolean {
    return proto3.util.equals(GetTimelineRequest_IDFilter, a, b);
  }
}

/**
 * Filters events by call.
 *
 * @generated from message xyz.block.ftl.timeline.v1.GetTimelineRequest.CallFilter
 */
export class GetTimelineRequest_CallFilter extends Message<GetTimelineRequest_CallFilter> {
  /**
   * @generated from field: string dest_module = 1;
   */
  destModule = "";

  /**
   * @generated from field: optional string dest_verb = 2;
   */
  destVerb?: string;

  /**
   * @generated from field: optional string source_module = 3;
   */
  sourceModule?: string;

  constructor(data?: PartialMessage<GetTimelineRequest_CallFilter>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.timeline.v1.GetTimelineRequest.CallFilter";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "dest_module", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "dest_verb", kind: "scalar", T: 9 /* ScalarType.STRING */, opt: true },
    { no: 3, name: "source_module", kind: "scalar", T: 9 /* ScalarType.STRING */, opt: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetTimelineRequest_CallFilter {
    return new GetTimelineRequest_CallFilter().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetTimelineRequest_CallFilter {
    return new GetTimelineRequest_CallFilter().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetTimelineRequest_CallFilter {
    return new GetTimelineRequest_CallFilter().fromJsonString(jsonString, options);
  }

  static equals(a: GetTimelineRequest_CallFilter | PlainMessage<GetTimelineRequest_CallFilter> | undefined, b: GetTimelineRequest_CallFilter | PlainMessage<GetTimelineRequest_CallFilter> | undefined): boolean {
    return proto3.util.equals(GetTimelineRequest_CallFilter, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.timeline.v1.GetTimelineRequest.ModuleFilter
 */
export class GetTimelineRequest_ModuleFilter extends Message<GetTimelineRequest_ModuleFilter> {
  /**
   * @generated from field: string module = 1;
   */
  module = "";

  /**
   * @generated from field: optional string verb = 2;
   */
  verb?: string;

  constructor(data?: PartialMessage<GetTimelineRequest_ModuleFilter>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.timeline.v1.GetTimelineRequest.ModuleFilter";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "module", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "verb", kind: "scalar", T: 9 /* ScalarType.STRING */, opt: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetTimelineRequest_ModuleFilter {
    return new GetTimelineRequest_ModuleFilter().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetTimelineRequest_ModuleFilter {
    return new GetTimelineRequest_ModuleFilter().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetTimelineRequest_ModuleFilter {
    return new GetTimelineRequest_ModuleFilter().fromJsonString(jsonString, options);
  }

  static equals(a: GetTimelineRequest_ModuleFilter | PlainMessage<GetTimelineRequest_ModuleFilter> | undefined, b: GetTimelineRequest_ModuleFilter | PlainMessage<GetTimelineRequest_ModuleFilter> | undefined): boolean {
    return proto3.util.equals(GetTimelineRequest_ModuleFilter, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.timeline.v1.GetTimelineRequest.Filter
 */
export class GetTimelineRequest_Filter extends Message<GetTimelineRequest_Filter> {
  /**
   * These map 1:1 with filters in backend/timeline/filters.go
   *
   * @generated from oneof xyz.block.ftl.timeline.v1.GetTimelineRequest.Filter.filter
   */
  filter: {
    /**
     * @generated from field: xyz.block.ftl.timeline.v1.GetTimelineRequest.LimitFilter limit = 1;
     */
    value: GetTimelineRequest_LimitFilter;
    case: "limit";
  } | {
    /**
     * @generated from field: xyz.block.ftl.timeline.v1.GetTimelineRequest.LogLevelFilter log_level = 2;
     */
    value: GetTimelineRequest_LogLevelFilter;
    case: "logLevel";
  } | {
    /**
     * @generated from field: xyz.block.ftl.timeline.v1.GetTimelineRequest.DeploymentFilter deployments = 3;
     */
    value: GetTimelineRequest_DeploymentFilter;
    case: "deployments";
  } | {
    /**
     * @generated from field: xyz.block.ftl.timeline.v1.GetTimelineRequest.RequestFilter requests = 4;
     */
    value: GetTimelineRequest_RequestFilter;
    case: "requests";
  } | {
    /**
     * @generated from field: xyz.block.ftl.timeline.v1.GetTimelineRequest.EventTypeFilter event_types = 5;
     */
    value: GetTimelineRequest_EventTypeFilter;
    case: "eventTypes";
  } | {
    /**
     * @generated from field: xyz.block.ftl.timeline.v1.GetTimelineRequest.TimeFilter time = 6;
     */
    value: GetTimelineRequest_TimeFilter;
    case: "time";
  } | {
    /**
     * @generated from field: xyz.block.ftl.timeline.v1.GetTimelineRequest.IDFilter id = 7;
     */
    value: GetTimelineRequest_IDFilter;
    case: "id";
  } | {
    /**
     * @generated from field: xyz.block.ftl.timeline.v1.GetTimelineRequest.CallFilter call = 8;
     */
    value: GetTimelineRequest_CallFilter;
    case: "call";
  } | {
    /**
     * @generated from field: xyz.block.ftl.timeline.v1.GetTimelineRequest.ModuleFilter module = 9;
     */
    value: GetTimelineRequest_ModuleFilter;
    case: "module";
  } | { case: undefined; value?: undefined } = { case: undefined };

  constructor(data?: PartialMessage<GetTimelineRequest_Filter>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.timeline.v1.GetTimelineRequest.Filter";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "limit", kind: "message", T: GetTimelineRequest_LimitFilter, oneof: "filter" },
    { no: 2, name: "log_level", kind: "message", T: GetTimelineRequest_LogLevelFilter, oneof: "filter" },
    { no: 3, name: "deployments", kind: "message", T: GetTimelineRequest_DeploymentFilter, oneof: "filter" },
    { no: 4, name: "requests", kind: "message", T: GetTimelineRequest_RequestFilter, oneof: "filter" },
    { no: 5, name: "event_types", kind: "message", T: GetTimelineRequest_EventTypeFilter, oneof: "filter" },
    { no: 6, name: "time", kind: "message", T: GetTimelineRequest_TimeFilter, oneof: "filter" },
    { no: 7, name: "id", kind: "message", T: GetTimelineRequest_IDFilter, oneof: "filter" },
    { no: 8, name: "call", kind: "message", T: GetTimelineRequest_CallFilter, oneof: "filter" },
    { no: 9, name: "module", kind: "message", T: GetTimelineRequest_ModuleFilter, oneof: "filter" },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetTimelineRequest_Filter {
    return new GetTimelineRequest_Filter().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetTimelineRequest_Filter {
    return new GetTimelineRequest_Filter().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetTimelineRequest_Filter {
    return new GetTimelineRequest_Filter().fromJsonString(jsonString, options);
  }

  static equals(a: GetTimelineRequest_Filter | PlainMessage<GetTimelineRequest_Filter> | undefined, b: GetTimelineRequest_Filter | PlainMessage<GetTimelineRequest_Filter> | undefined): boolean {
    return proto3.util.equals(GetTimelineRequest_Filter, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.timeline.v1.GetTimelineResponse
 */
export class GetTimelineResponse extends Message<GetTimelineResponse> {
  /**
   * @generated from field: repeated xyz.block.ftl.timeline.v1.Event events = 1;
   */
  events: Event[] = [];

  constructor(data?: PartialMessage<GetTimelineResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.timeline.v1.GetTimelineResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "events", kind: "message", T: Event, repeated: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetTimelineResponse {
    return new GetTimelineResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetTimelineResponse {
    return new GetTimelineResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetTimelineResponse {
    return new GetTimelineResponse().fromJsonString(jsonString, options);
  }

  static equals(a: GetTimelineResponse | PlainMessage<GetTimelineResponse> | undefined, b: GetTimelineResponse | PlainMessage<GetTimelineResponse> | undefined): boolean {
    return proto3.util.equals(GetTimelineResponse, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.timeline.v1.CreateEventRequest
 */
export class CreateEventRequest extends Message<CreateEventRequest> {
  /**
   * @generated from oneof xyz.block.ftl.timeline.v1.CreateEventRequest.entry
   */
  entry: {
    /**
     * @generated from field: xyz.block.ftl.timeline.v1.LogEvent log = 1;
     */
    value: LogEvent;
    case: "log";
  } | {
    /**
     * @generated from field: xyz.block.ftl.timeline.v1.CallEvent call = 2;
     */
    value: CallEvent;
    case: "call";
  } | {
    /**
     * @generated from field: xyz.block.ftl.timeline.v1.DeploymentCreatedEvent deployment_created = 3;
     */
    value: DeploymentCreatedEvent;
    case: "deploymentCreated";
  } | {
    /**
     * @generated from field: xyz.block.ftl.timeline.v1.DeploymentUpdatedEvent deployment_updated = 4;
     */
    value: DeploymentUpdatedEvent;
    case: "deploymentUpdated";
  } | {
    /**
     * @generated from field: xyz.block.ftl.timeline.v1.IngressEvent ingress = 5;
     */
    value: IngressEvent;
    case: "ingress";
  } | {
    /**
     * @generated from field: xyz.block.ftl.timeline.v1.CronScheduledEvent cron_scheduled = 6;
     */
    value: CronScheduledEvent;
    case: "cronScheduled";
  } | {
    /**
     * @generated from field: xyz.block.ftl.timeline.v1.AsyncExecuteEvent async_execute = 7;
     */
    value: AsyncExecuteEvent;
    case: "asyncExecute";
  } | {
    /**
     * @generated from field: xyz.block.ftl.timeline.v1.PubSubPublishEvent pubsub_publish = 8;
     */
    value: PubSubPublishEvent;
    case: "pubsubPublish";
  } | {
    /**
     * @generated from field: xyz.block.ftl.timeline.v1.PubSubConsumeEvent pubsub_consume = 9;
     */
    value: PubSubConsumeEvent;
    case: "pubsubConsume";
  } | { case: undefined; value?: undefined } = { case: undefined };

  constructor(data?: PartialMessage<CreateEventRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.timeline.v1.CreateEventRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "log", kind: "message", T: LogEvent, oneof: "entry" },
    { no: 2, name: "call", kind: "message", T: CallEvent, oneof: "entry" },
    { no: 3, name: "deployment_created", kind: "message", T: DeploymentCreatedEvent, oneof: "entry" },
    { no: 4, name: "deployment_updated", kind: "message", T: DeploymentUpdatedEvent, oneof: "entry" },
    { no: 5, name: "ingress", kind: "message", T: IngressEvent, oneof: "entry" },
    { no: 6, name: "cron_scheduled", kind: "message", T: CronScheduledEvent, oneof: "entry" },
    { no: 7, name: "async_execute", kind: "message", T: AsyncExecuteEvent, oneof: "entry" },
    { no: 8, name: "pubsub_publish", kind: "message", T: PubSubPublishEvent, oneof: "entry" },
    { no: 9, name: "pubsub_consume", kind: "message", T: PubSubConsumeEvent, oneof: "entry" },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): CreateEventRequest {
    return new CreateEventRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): CreateEventRequest {
    return new CreateEventRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): CreateEventRequest {
    return new CreateEventRequest().fromJsonString(jsonString, options);
  }

  static equals(a: CreateEventRequest | PlainMessage<CreateEventRequest> | undefined, b: CreateEventRequest | PlainMessage<CreateEventRequest> | undefined): boolean {
    return proto3.util.equals(CreateEventRequest, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.timeline.v1.CreateEventResponse
 */
export class CreateEventResponse extends Message<CreateEventResponse> {
  constructor(data?: PartialMessage<CreateEventResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.timeline.v1.CreateEventResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): CreateEventResponse {
    return new CreateEventResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): CreateEventResponse {
    return new CreateEventResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): CreateEventResponse {
    return new CreateEventResponse().fromJsonString(jsonString, options);
  }

  static equals(a: CreateEventResponse | PlainMessage<CreateEventResponse> | undefined, b: CreateEventResponse | PlainMessage<CreateEventResponse> | undefined): boolean {
    return proto3.util.equals(CreateEventResponse, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.timeline.v1.DeleteOldEventsRequest
 */
export class DeleteOldEventsRequest extends Message<DeleteOldEventsRequest> {
  /**
   * @generated from field: xyz.block.ftl.timeline.v1.EventType event_type = 1;
   */
  eventType = EventType.UNSPECIFIED;

  /**
   * @generated from field: int64 age_seconds = 2;
   */
  ageSeconds = protoInt64.zero;

  constructor(data?: PartialMessage<DeleteOldEventsRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.timeline.v1.DeleteOldEventsRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "event_type", kind: "enum", T: proto3.getEnumType(EventType) },
    { no: 2, name: "age_seconds", kind: "scalar", T: 3 /* ScalarType.INT64 */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): DeleteOldEventsRequest {
    return new DeleteOldEventsRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): DeleteOldEventsRequest {
    return new DeleteOldEventsRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): DeleteOldEventsRequest {
    return new DeleteOldEventsRequest().fromJsonString(jsonString, options);
  }

  static equals(a: DeleteOldEventsRequest | PlainMessage<DeleteOldEventsRequest> | undefined, b: DeleteOldEventsRequest | PlainMessage<DeleteOldEventsRequest> | undefined): boolean {
    return proto3.util.equals(DeleteOldEventsRequest, a, b);
  }
}

/**
 * @generated from message xyz.block.ftl.timeline.v1.DeleteOldEventsResponse
 */
export class DeleteOldEventsResponse extends Message<DeleteOldEventsResponse> {
  /**
   * @generated from field: int64 deleted_count = 1;
   */
  deletedCount = protoInt64.zero;

  constructor(data?: PartialMessage<DeleteOldEventsResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "xyz.block.ftl.timeline.v1.DeleteOldEventsResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "deleted_count", kind: "scalar", T: 3 /* ScalarType.INT64 */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): DeleteOldEventsResponse {
    return new DeleteOldEventsResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): DeleteOldEventsResponse {
    return new DeleteOldEventsResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): DeleteOldEventsResponse {
    return new DeleteOldEventsResponse().fromJsonString(jsonString, options);
  }

  static equals(a: DeleteOldEventsResponse | PlainMessage<DeleteOldEventsResponse> | undefined, b: DeleteOldEventsResponse | PlainMessage<DeleteOldEventsResponse> | undefined): boolean {
    return proto3.util.equals(DeleteOldEventsResponse, a, b);
  }
}
