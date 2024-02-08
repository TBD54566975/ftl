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

// @generated by protoc-gen-es v1.7.2 with parameter "target=ts"
// @generated from file opentelemetry/proto/collector/metrics/v1/metrics_service.proto (package opentelemetry.proto.collector.metrics.v1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { Message, proto3, protoInt64 } from "@bufbuild/protobuf";
import { ResourceMetrics } from "../../../metrics/v1/metrics_pb.js";

/**
 * @generated from message opentelemetry.proto.collector.metrics.v1.ExportMetricsServiceRequest
 */
export class ExportMetricsServiceRequest extends Message<ExportMetricsServiceRequest> {
  /**
   * An array of ResourceMetrics.
   * For data coming from a single resource this array will typically contain one
   * element. Intermediary nodes (such as OpenTelemetry Collector) that receive
   * data from multiple origins typically batch the data before forwarding further and
   * in that case this array will contain multiple elements.
   *
   * @generated from field: repeated opentelemetry.proto.metrics.v1.ResourceMetrics resource_metrics = 1;
   */
  resourceMetrics: ResourceMetrics[] = [];

  constructor(data?: PartialMessage<ExportMetricsServiceRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "opentelemetry.proto.collector.metrics.v1.ExportMetricsServiceRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "resource_metrics", kind: "message", T: ResourceMetrics, repeated: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ExportMetricsServiceRequest {
    return new ExportMetricsServiceRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ExportMetricsServiceRequest {
    return new ExportMetricsServiceRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ExportMetricsServiceRequest {
    return new ExportMetricsServiceRequest().fromJsonString(jsonString, options);
  }

  static equals(a: ExportMetricsServiceRequest | PlainMessage<ExportMetricsServiceRequest> | undefined, b: ExportMetricsServiceRequest | PlainMessage<ExportMetricsServiceRequest> | undefined): boolean {
    return proto3.util.equals(ExportMetricsServiceRequest, a, b);
  }
}

/**
 * @generated from message opentelemetry.proto.collector.metrics.v1.ExportMetricsServiceResponse
 */
export class ExportMetricsServiceResponse extends Message<ExportMetricsServiceResponse> {
  /**
   * The details of a partially successful export request.
   *
   * If the request is only partially accepted
   * (i.e. when the server accepts only parts of the data and rejects the rest)
   * the server MUST initialize the `partial_success` field and MUST
   * set the `rejected_<signal>` with the number of items it rejected.
   *
   * Servers MAY also make use of the `partial_success` field to convey
   * warnings/suggestions to senders even when the request was fully accepted.
   * In such cases, the `rejected_<signal>` MUST have a value of `0` and
   * the `error_message` MUST be non-empty.
   *
   * A `partial_success` message with an empty value (rejected_<signal> = 0 and
   * `error_message` = "") is equivalent to it not being set/present. Senders
   * SHOULD interpret it the same way as in the full success case.
   *
   * @generated from field: opentelemetry.proto.collector.metrics.v1.ExportMetricsPartialSuccess partial_success = 1;
   */
  partialSuccess?: ExportMetricsPartialSuccess;

  constructor(data?: PartialMessage<ExportMetricsServiceResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "opentelemetry.proto.collector.metrics.v1.ExportMetricsServiceResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "partial_success", kind: "message", T: ExportMetricsPartialSuccess },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ExportMetricsServiceResponse {
    return new ExportMetricsServiceResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ExportMetricsServiceResponse {
    return new ExportMetricsServiceResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ExportMetricsServiceResponse {
    return new ExportMetricsServiceResponse().fromJsonString(jsonString, options);
  }

  static equals(a: ExportMetricsServiceResponse | PlainMessage<ExportMetricsServiceResponse> | undefined, b: ExportMetricsServiceResponse | PlainMessage<ExportMetricsServiceResponse> | undefined): boolean {
    return proto3.util.equals(ExportMetricsServiceResponse, a, b);
  }
}

/**
 * @generated from message opentelemetry.proto.collector.metrics.v1.ExportMetricsPartialSuccess
 */
export class ExportMetricsPartialSuccess extends Message<ExportMetricsPartialSuccess> {
  /**
   * The number of rejected data points.
   *
   * A `rejected_<signal>` field holding a `0` value indicates that the
   * request was fully accepted.
   *
   * @generated from field: int64 rejected_data_points = 1;
   */
  rejectedDataPoints = protoInt64.zero;

  /**
   * A developer-facing human-readable message in English. It should be used
   * either to explain why the server rejected parts of the data during a partial
   * success or to convey warnings/suggestions during a full success. The message
   * should offer guidance on how users can address such issues.
   *
   * error_message is an optional field. An error_message with an empty value
   * is equivalent to it not being set.
   *
   * @generated from field: string error_message = 2;
   */
  errorMessage = "";

  constructor(data?: PartialMessage<ExportMetricsPartialSuccess>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "opentelemetry.proto.collector.metrics.v1.ExportMetricsPartialSuccess";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "rejected_data_points", kind: "scalar", T: 3 /* ScalarType.INT64 */ },
    { no: 2, name: "error_message", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ExportMetricsPartialSuccess {
    return new ExportMetricsPartialSuccess().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ExportMetricsPartialSuccess {
    return new ExportMetricsPartialSuccess().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ExportMetricsPartialSuccess {
    return new ExportMetricsPartialSuccess().fromJsonString(jsonString, options);
  }

  static equals(a: ExportMetricsPartialSuccess | PlainMessage<ExportMetricsPartialSuccess> | undefined, b: ExportMetricsPartialSuccess | PlainMessage<ExportMetricsPartialSuccess> | undefined): boolean {
    return proto3.util.equals(ExportMetricsPartialSuccess, a, b);
  }
}

