package xyz.block.ftl.v1;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 */
@javax.annotation.Generated(
    value = "by gRPC proto compiler (version 1.68.2)",
    comments = "Source: xyz/block/ftl/v1/controller.proto")
@io.grpc.stub.annotations.GrpcGenerated
public final class ControllerServiceGrpc {

  private ControllerServiceGrpc() {}

  public static final java.lang.String SERVICE_NAME = "xyz.block.ftl.v1.ControllerService";

  // Static method descriptors that strictly reflect the proto.
  private static volatile io.grpc.MethodDescriptor<xyz.block.ftl.v1.PingRequest,
      xyz.block.ftl.v1.PingResponse> getPingMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "Ping",
      requestType = xyz.block.ftl.v1.PingRequest.class,
      responseType = xyz.block.ftl.v1.PingResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<xyz.block.ftl.v1.PingRequest,
      xyz.block.ftl.v1.PingResponse> getPingMethod() {
    io.grpc.MethodDescriptor<xyz.block.ftl.v1.PingRequest, xyz.block.ftl.v1.PingResponse> getPingMethod;
    if ((getPingMethod = ControllerServiceGrpc.getPingMethod) == null) {
      synchronized (ControllerServiceGrpc.class) {
        if ((getPingMethod = ControllerServiceGrpc.getPingMethod) == null) {
          ControllerServiceGrpc.getPingMethod = getPingMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.v1.PingRequest, xyz.block.ftl.v1.PingResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "Ping"))
              .setSafe(true)
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.PingRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.PingResponse.getDefaultInstance()))
              .setSchemaDescriptor(new ControllerServiceMethodDescriptorSupplier("Ping"))
              .build();
        }
      }
    }
    return getPingMethod;
  }

  private static volatile io.grpc.MethodDescriptor<xyz.block.ftl.v1.ProcessListRequest,
      xyz.block.ftl.v1.ProcessListResponse> getProcessListMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "ProcessList",
      requestType = xyz.block.ftl.v1.ProcessListRequest.class,
      responseType = xyz.block.ftl.v1.ProcessListResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<xyz.block.ftl.v1.ProcessListRequest,
      xyz.block.ftl.v1.ProcessListResponse> getProcessListMethod() {
    io.grpc.MethodDescriptor<xyz.block.ftl.v1.ProcessListRequest, xyz.block.ftl.v1.ProcessListResponse> getProcessListMethod;
    if ((getProcessListMethod = ControllerServiceGrpc.getProcessListMethod) == null) {
      synchronized (ControllerServiceGrpc.class) {
        if ((getProcessListMethod = ControllerServiceGrpc.getProcessListMethod) == null) {
          ControllerServiceGrpc.getProcessListMethod = getProcessListMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.v1.ProcessListRequest, xyz.block.ftl.v1.ProcessListResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "ProcessList"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.ProcessListRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.ProcessListResponse.getDefaultInstance()))
              .setSchemaDescriptor(new ControllerServiceMethodDescriptorSupplier("ProcessList"))
              .build();
        }
      }
    }
    return getProcessListMethod;
  }

  private static volatile io.grpc.MethodDescriptor<xyz.block.ftl.v1.StatusRequest,
      xyz.block.ftl.v1.StatusResponse> getStatusMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "Status",
      requestType = xyz.block.ftl.v1.StatusRequest.class,
      responseType = xyz.block.ftl.v1.StatusResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<xyz.block.ftl.v1.StatusRequest,
      xyz.block.ftl.v1.StatusResponse> getStatusMethod() {
    io.grpc.MethodDescriptor<xyz.block.ftl.v1.StatusRequest, xyz.block.ftl.v1.StatusResponse> getStatusMethod;
    if ((getStatusMethod = ControllerServiceGrpc.getStatusMethod) == null) {
      synchronized (ControllerServiceGrpc.class) {
        if ((getStatusMethod = ControllerServiceGrpc.getStatusMethod) == null) {
          ControllerServiceGrpc.getStatusMethod = getStatusMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.v1.StatusRequest, xyz.block.ftl.v1.StatusResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "Status"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.StatusRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.StatusResponse.getDefaultInstance()))
              .setSchemaDescriptor(new ControllerServiceMethodDescriptorSupplier("Status"))
              .build();
        }
      }
    }
    return getStatusMethod;
  }

  private static volatile io.grpc.MethodDescriptor<xyz.block.ftl.v1.GetArtefactDiffsRequest,
      xyz.block.ftl.v1.GetArtefactDiffsResponse> getGetArtefactDiffsMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "GetArtefactDiffs",
      requestType = xyz.block.ftl.v1.GetArtefactDiffsRequest.class,
      responseType = xyz.block.ftl.v1.GetArtefactDiffsResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<xyz.block.ftl.v1.GetArtefactDiffsRequest,
      xyz.block.ftl.v1.GetArtefactDiffsResponse> getGetArtefactDiffsMethod() {
    io.grpc.MethodDescriptor<xyz.block.ftl.v1.GetArtefactDiffsRequest, xyz.block.ftl.v1.GetArtefactDiffsResponse> getGetArtefactDiffsMethod;
    if ((getGetArtefactDiffsMethod = ControllerServiceGrpc.getGetArtefactDiffsMethod) == null) {
      synchronized (ControllerServiceGrpc.class) {
        if ((getGetArtefactDiffsMethod = ControllerServiceGrpc.getGetArtefactDiffsMethod) == null) {
          ControllerServiceGrpc.getGetArtefactDiffsMethod = getGetArtefactDiffsMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.v1.GetArtefactDiffsRequest, xyz.block.ftl.v1.GetArtefactDiffsResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "GetArtefactDiffs"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.GetArtefactDiffsRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.GetArtefactDiffsResponse.getDefaultInstance()))
              .setSchemaDescriptor(new ControllerServiceMethodDescriptorSupplier("GetArtefactDiffs"))
              .build();
        }
      }
    }
    return getGetArtefactDiffsMethod;
  }

  private static volatile io.grpc.MethodDescriptor<xyz.block.ftl.v1.UploadArtefactRequest,
      xyz.block.ftl.v1.UploadArtefactResponse> getUploadArtefactMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "UploadArtefact",
      requestType = xyz.block.ftl.v1.UploadArtefactRequest.class,
      responseType = xyz.block.ftl.v1.UploadArtefactResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<xyz.block.ftl.v1.UploadArtefactRequest,
      xyz.block.ftl.v1.UploadArtefactResponse> getUploadArtefactMethod() {
    io.grpc.MethodDescriptor<xyz.block.ftl.v1.UploadArtefactRequest, xyz.block.ftl.v1.UploadArtefactResponse> getUploadArtefactMethod;
    if ((getUploadArtefactMethod = ControllerServiceGrpc.getUploadArtefactMethod) == null) {
      synchronized (ControllerServiceGrpc.class) {
        if ((getUploadArtefactMethod = ControllerServiceGrpc.getUploadArtefactMethod) == null) {
          ControllerServiceGrpc.getUploadArtefactMethod = getUploadArtefactMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.v1.UploadArtefactRequest, xyz.block.ftl.v1.UploadArtefactResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "UploadArtefact"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.UploadArtefactRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.UploadArtefactResponse.getDefaultInstance()))
              .setSchemaDescriptor(new ControllerServiceMethodDescriptorSupplier("UploadArtefact"))
              .build();
        }
      }
    }
    return getUploadArtefactMethod;
  }

  private static volatile io.grpc.MethodDescriptor<xyz.block.ftl.v1.CreateDeploymentRequest,
      xyz.block.ftl.v1.CreateDeploymentResponse> getCreateDeploymentMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "CreateDeployment",
      requestType = xyz.block.ftl.v1.CreateDeploymentRequest.class,
      responseType = xyz.block.ftl.v1.CreateDeploymentResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<xyz.block.ftl.v1.CreateDeploymentRequest,
      xyz.block.ftl.v1.CreateDeploymentResponse> getCreateDeploymentMethod() {
    io.grpc.MethodDescriptor<xyz.block.ftl.v1.CreateDeploymentRequest, xyz.block.ftl.v1.CreateDeploymentResponse> getCreateDeploymentMethod;
    if ((getCreateDeploymentMethod = ControllerServiceGrpc.getCreateDeploymentMethod) == null) {
      synchronized (ControllerServiceGrpc.class) {
        if ((getCreateDeploymentMethod = ControllerServiceGrpc.getCreateDeploymentMethod) == null) {
          ControllerServiceGrpc.getCreateDeploymentMethod = getCreateDeploymentMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.v1.CreateDeploymentRequest, xyz.block.ftl.v1.CreateDeploymentResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "CreateDeployment"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.CreateDeploymentRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.CreateDeploymentResponse.getDefaultInstance()))
              .setSchemaDescriptor(new ControllerServiceMethodDescriptorSupplier("CreateDeployment"))
              .build();
        }
      }
    }
    return getCreateDeploymentMethod;
  }

  private static volatile io.grpc.MethodDescriptor<xyz.block.ftl.v1.GetDeploymentRequest,
      xyz.block.ftl.v1.GetDeploymentResponse> getGetDeploymentMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "GetDeployment",
      requestType = xyz.block.ftl.v1.GetDeploymentRequest.class,
      responseType = xyz.block.ftl.v1.GetDeploymentResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<xyz.block.ftl.v1.GetDeploymentRequest,
      xyz.block.ftl.v1.GetDeploymentResponse> getGetDeploymentMethod() {
    io.grpc.MethodDescriptor<xyz.block.ftl.v1.GetDeploymentRequest, xyz.block.ftl.v1.GetDeploymentResponse> getGetDeploymentMethod;
    if ((getGetDeploymentMethod = ControllerServiceGrpc.getGetDeploymentMethod) == null) {
      synchronized (ControllerServiceGrpc.class) {
        if ((getGetDeploymentMethod = ControllerServiceGrpc.getGetDeploymentMethod) == null) {
          ControllerServiceGrpc.getGetDeploymentMethod = getGetDeploymentMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.v1.GetDeploymentRequest, xyz.block.ftl.v1.GetDeploymentResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "GetDeployment"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.GetDeploymentRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.GetDeploymentResponse.getDefaultInstance()))
              .setSchemaDescriptor(new ControllerServiceMethodDescriptorSupplier("GetDeployment"))
              .build();
        }
      }
    }
    return getGetDeploymentMethod;
  }

  private static volatile io.grpc.MethodDescriptor<xyz.block.ftl.v1.GetDeploymentArtefactsRequest,
      xyz.block.ftl.v1.GetDeploymentArtefactsResponse> getGetDeploymentArtefactsMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "GetDeploymentArtefacts",
      requestType = xyz.block.ftl.v1.GetDeploymentArtefactsRequest.class,
      responseType = xyz.block.ftl.v1.GetDeploymentArtefactsResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.SERVER_STREAMING)
  public static io.grpc.MethodDescriptor<xyz.block.ftl.v1.GetDeploymentArtefactsRequest,
      xyz.block.ftl.v1.GetDeploymentArtefactsResponse> getGetDeploymentArtefactsMethod() {
    io.grpc.MethodDescriptor<xyz.block.ftl.v1.GetDeploymentArtefactsRequest, xyz.block.ftl.v1.GetDeploymentArtefactsResponse> getGetDeploymentArtefactsMethod;
    if ((getGetDeploymentArtefactsMethod = ControllerServiceGrpc.getGetDeploymentArtefactsMethod) == null) {
      synchronized (ControllerServiceGrpc.class) {
        if ((getGetDeploymentArtefactsMethod = ControllerServiceGrpc.getGetDeploymentArtefactsMethod) == null) {
          ControllerServiceGrpc.getGetDeploymentArtefactsMethod = getGetDeploymentArtefactsMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.v1.GetDeploymentArtefactsRequest, xyz.block.ftl.v1.GetDeploymentArtefactsResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.SERVER_STREAMING)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "GetDeploymentArtefacts"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.GetDeploymentArtefactsRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.GetDeploymentArtefactsResponse.getDefaultInstance()))
              .setSchemaDescriptor(new ControllerServiceMethodDescriptorSupplier("GetDeploymentArtefacts"))
              .build();
        }
      }
    }
    return getGetDeploymentArtefactsMethod;
  }

  private static volatile io.grpc.MethodDescriptor<xyz.block.ftl.v1.RegisterRunnerRequest,
      xyz.block.ftl.v1.RegisterRunnerResponse> getRegisterRunnerMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "RegisterRunner",
      requestType = xyz.block.ftl.v1.RegisterRunnerRequest.class,
      responseType = xyz.block.ftl.v1.RegisterRunnerResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.CLIENT_STREAMING)
  public static io.grpc.MethodDescriptor<xyz.block.ftl.v1.RegisterRunnerRequest,
      xyz.block.ftl.v1.RegisterRunnerResponse> getRegisterRunnerMethod() {
    io.grpc.MethodDescriptor<xyz.block.ftl.v1.RegisterRunnerRequest, xyz.block.ftl.v1.RegisterRunnerResponse> getRegisterRunnerMethod;
    if ((getRegisterRunnerMethod = ControllerServiceGrpc.getRegisterRunnerMethod) == null) {
      synchronized (ControllerServiceGrpc.class) {
        if ((getRegisterRunnerMethod = ControllerServiceGrpc.getRegisterRunnerMethod) == null) {
          ControllerServiceGrpc.getRegisterRunnerMethod = getRegisterRunnerMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.v1.RegisterRunnerRequest, xyz.block.ftl.v1.RegisterRunnerResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.CLIENT_STREAMING)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "RegisterRunner"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.RegisterRunnerRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.RegisterRunnerResponse.getDefaultInstance()))
              .setSchemaDescriptor(new ControllerServiceMethodDescriptorSupplier("RegisterRunner"))
              .build();
        }
      }
    }
    return getRegisterRunnerMethod;
  }

  private static volatile io.grpc.MethodDescriptor<xyz.block.ftl.v1.UpdateDeployRequest,
      xyz.block.ftl.v1.UpdateDeployResponse> getUpdateDeployMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "UpdateDeploy",
      requestType = xyz.block.ftl.v1.UpdateDeployRequest.class,
      responseType = xyz.block.ftl.v1.UpdateDeployResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<xyz.block.ftl.v1.UpdateDeployRequest,
      xyz.block.ftl.v1.UpdateDeployResponse> getUpdateDeployMethod() {
    io.grpc.MethodDescriptor<xyz.block.ftl.v1.UpdateDeployRequest, xyz.block.ftl.v1.UpdateDeployResponse> getUpdateDeployMethod;
    if ((getUpdateDeployMethod = ControllerServiceGrpc.getUpdateDeployMethod) == null) {
      synchronized (ControllerServiceGrpc.class) {
        if ((getUpdateDeployMethod = ControllerServiceGrpc.getUpdateDeployMethod) == null) {
          ControllerServiceGrpc.getUpdateDeployMethod = getUpdateDeployMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.v1.UpdateDeployRequest, xyz.block.ftl.v1.UpdateDeployResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "UpdateDeploy"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.UpdateDeployRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.UpdateDeployResponse.getDefaultInstance()))
              .setSchemaDescriptor(new ControllerServiceMethodDescriptorSupplier("UpdateDeploy"))
              .build();
        }
      }
    }
    return getUpdateDeployMethod;
  }

  private static volatile io.grpc.MethodDescriptor<xyz.block.ftl.v1.ReplaceDeployRequest,
      xyz.block.ftl.v1.ReplaceDeployResponse> getReplaceDeployMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "ReplaceDeploy",
      requestType = xyz.block.ftl.v1.ReplaceDeployRequest.class,
      responseType = xyz.block.ftl.v1.ReplaceDeployResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<xyz.block.ftl.v1.ReplaceDeployRequest,
      xyz.block.ftl.v1.ReplaceDeployResponse> getReplaceDeployMethod() {
    io.grpc.MethodDescriptor<xyz.block.ftl.v1.ReplaceDeployRequest, xyz.block.ftl.v1.ReplaceDeployResponse> getReplaceDeployMethod;
    if ((getReplaceDeployMethod = ControllerServiceGrpc.getReplaceDeployMethod) == null) {
      synchronized (ControllerServiceGrpc.class) {
        if ((getReplaceDeployMethod = ControllerServiceGrpc.getReplaceDeployMethod) == null) {
          ControllerServiceGrpc.getReplaceDeployMethod = getReplaceDeployMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.v1.ReplaceDeployRequest, xyz.block.ftl.v1.ReplaceDeployResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "ReplaceDeploy"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.ReplaceDeployRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.ReplaceDeployResponse.getDefaultInstance()))
              .setSchemaDescriptor(new ControllerServiceMethodDescriptorSupplier("ReplaceDeploy"))
              .build();
        }
      }
    }
    return getReplaceDeployMethod;
  }

  private static volatile io.grpc.MethodDescriptor<xyz.block.ftl.v1.StreamDeploymentLogsRequest,
      xyz.block.ftl.v1.StreamDeploymentLogsResponse> getStreamDeploymentLogsMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "StreamDeploymentLogs",
      requestType = xyz.block.ftl.v1.StreamDeploymentLogsRequest.class,
      responseType = xyz.block.ftl.v1.StreamDeploymentLogsResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.CLIENT_STREAMING)
  public static io.grpc.MethodDescriptor<xyz.block.ftl.v1.StreamDeploymentLogsRequest,
      xyz.block.ftl.v1.StreamDeploymentLogsResponse> getStreamDeploymentLogsMethod() {
    io.grpc.MethodDescriptor<xyz.block.ftl.v1.StreamDeploymentLogsRequest, xyz.block.ftl.v1.StreamDeploymentLogsResponse> getStreamDeploymentLogsMethod;
    if ((getStreamDeploymentLogsMethod = ControllerServiceGrpc.getStreamDeploymentLogsMethod) == null) {
      synchronized (ControllerServiceGrpc.class) {
        if ((getStreamDeploymentLogsMethod = ControllerServiceGrpc.getStreamDeploymentLogsMethod) == null) {
          ControllerServiceGrpc.getStreamDeploymentLogsMethod = getStreamDeploymentLogsMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.v1.StreamDeploymentLogsRequest, xyz.block.ftl.v1.StreamDeploymentLogsResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.CLIENT_STREAMING)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "StreamDeploymentLogs"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.StreamDeploymentLogsRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.StreamDeploymentLogsResponse.getDefaultInstance()))
              .setSchemaDescriptor(new ControllerServiceMethodDescriptorSupplier("StreamDeploymentLogs"))
              .build();
        }
      }
    }
    return getStreamDeploymentLogsMethod;
  }

  /**
   * Creates a new async stub that supports all call types for the service
   */
  public static ControllerServiceStub newStub(io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<ControllerServiceStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<ControllerServiceStub>() {
        @java.lang.Override
        public ControllerServiceStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new ControllerServiceStub(channel, callOptions);
        }
      };
    return ControllerServiceStub.newStub(factory, channel);
  }

  /**
   * Creates a new blocking-style stub that supports unary and streaming output calls on the service
   */
  public static ControllerServiceBlockingStub newBlockingStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<ControllerServiceBlockingStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<ControllerServiceBlockingStub>() {
        @java.lang.Override
        public ControllerServiceBlockingStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new ControllerServiceBlockingStub(channel, callOptions);
        }
      };
    return ControllerServiceBlockingStub.newStub(factory, channel);
  }

  /**
   * Creates a new ListenableFuture-style stub that supports unary calls on the service
   */
  public static ControllerServiceFutureStub newFutureStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<ControllerServiceFutureStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<ControllerServiceFutureStub>() {
        @java.lang.Override
        public ControllerServiceFutureStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new ControllerServiceFutureStub(channel, callOptions);
        }
      };
    return ControllerServiceFutureStub.newStub(factory, channel);
  }

  /**
   */
  public interface AsyncService {

    /**
     * <pre>
     * Ping service for readiness.
     * </pre>
     */
    default void ping(xyz.block.ftl.v1.PingRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.PingResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getPingMethod(), responseObserver);
    }

    /**
     * <pre>
     * List "processes" running on the cluster.
     * </pre>
     */
    default void processList(xyz.block.ftl.v1.ProcessListRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.ProcessListResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getProcessListMethod(), responseObserver);
    }

    /**
     */
    default void status(xyz.block.ftl.v1.StatusRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.StatusResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getStatusMethod(), responseObserver);
    }

    /**
     * <pre>
     * Get list of artefacts that differ between the server and client.
     * </pre>
     */
    default void getArtefactDiffs(xyz.block.ftl.v1.GetArtefactDiffsRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.GetArtefactDiffsResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getGetArtefactDiffsMethod(), responseObserver);
    }

    /**
     * <pre>
     * Upload an artefact to the server.
     * </pre>
     */
    default void uploadArtefact(xyz.block.ftl.v1.UploadArtefactRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.UploadArtefactResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getUploadArtefactMethod(), responseObserver);
    }

    /**
     * <pre>
     * Create a deployment.
     * </pre>
     */
    default void createDeployment(xyz.block.ftl.v1.CreateDeploymentRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.CreateDeploymentResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getCreateDeploymentMethod(), responseObserver);
    }

    /**
     * <pre>
     * Get the schema and artefact metadata for a deployment.
     * </pre>
     */
    default void getDeployment(xyz.block.ftl.v1.GetDeploymentRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.GetDeploymentResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getGetDeploymentMethod(), responseObserver);
    }

    /**
     * <pre>
     * Stream deployment artefacts from the server.
     * Each artefact is streamed one after the other as a sequence of max 1MB
     * chunks.
     * </pre>
     */
    default void getDeploymentArtefacts(xyz.block.ftl.v1.GetDeploymentArtefactsRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.GetDeploymentArtefactsResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getGetDeploymentArtefactsMethod(), responseObserver);
    }

    /**
     * <pre>
     * Register a Runner with the Controller.
     * Each runner issue a RegisterRunnerRequest to the ControllerService
     * every 10 seconds to maintain its heartbeat.
     * </pre>
     */
    default io.grpc.stub.StreamObserver<xyz.block.ftl.v1.RegisterRunnerRequest> registerRunner(
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.RegisterRunnerResponse> responseObserver) {
      return io.grpc.stub.ServerCalls.asyncUnimplementedStreamingCall(getRegisterRunnerMethod(), responseObserver);
    }

    /**
     * <pre>
     * Update an existing deployment.
     * </pre>
     */
    default void updateDeploy(xyz.block.ftl.v1.UpdateDeployRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.UpdateDeployResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getUpdateDeployMethod(), responseObserver);
    }

    /**
     * <pre>
     * Gradually replace an existing deployment with a new one.
     * If a deployment already exists for the module of the new deployment,
     * it will be scaled down and replaced by the new one.
     * </pre>
     */
    default void replaceDeploy(xyz.block.ftl.v1.ReplaceDeployRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.ReplaceDeployResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getReplaceDeployMethod(), responseObserver);
    }

    /**
     * <pre>
     * Stream logs from a deployment
     * </pre>
     */
    default io.grpc.stub.StreamObserver<xyz.block.ftl.v1.StreamDeploymentLogsRequest> streamDeploymentLogs(
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.StreamDeploymentLogsResponse> responseObserver) {
      return io.grpc.stub.ServerCalls.asyncUnimplementedStreamingCall(getStreamDeploymentLogsMethod(), responseObserver);
    }
  }

  /**
   * Base class for the server implementation of the service ControllerService.
   */
  public static abstract class ControllerServiceImplBase
      implements io.grpc.BindableService, AsyncService {

    @java.lang.Override public final io.grpc.ServerServiceDefinition bindService() {
      return ControllerServiceGrpc.bindService(this);
    }
  }

  /**
   * A stub to allow clients to do asynchronous rpc calls to service ControllerService.
   */
  public static final class ControllerServiceStub
      extends io.grpc.stub.AbstractAsyncStub<ControllerServiceStub> {
    private ControllerServiceStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected ControllerServiceStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new ControllerServiceStub(channel, callOptions);
    }

    /**
     * <pre>
     * Ping service for readiness.
     * </pre>
     */
    public void ping(xyz.block.ftl.v1.PingRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.PingResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getPingMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * List "processes" running on the cluster.
     * </pre>
     */
    public void processList(xyz.block.ftl.v1.ProcessListRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.ProcessListResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getProcessListMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void status(xyz.block.ftl.v1.StatusRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.StatusResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getStatusMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * Get list of artefacts that differ between the server and client.
     * </pre>
     */
    public void getArtefactDiffs(xyz.block.ftl.v1.GetArtefactDiffsRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.GetArtefactDiffsResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getGetArtefactDiffsMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * Upload an artefact to the server.
     * </pre>
     */
    public void uploadArtefact(xyz.block.ftl.v1.UploadArtefactRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.UploadArtefactResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getUploadArtefactMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * Create a deployment.
     * </pre>
     */
    public void createDeployment(xyz.block.ftl.v1.CreateDeploymentRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.CreateDeploymentResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getCreateDeploymentMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * Get the schema and artefact metadata for a deployment.
     * </pre>
     */
    public void getDeployment(xyz.block.ftl.v1.GetDeploymentRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.GetDeploymentResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getGetDeploymentMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * Stream deployment artefacts from the server.
     * Each artefact is streamed one after the other as a sequence of max 1MB
     * chunks.
     * </pre>
     */
    public void getDeploymentArtefacts(xyz.block.ftl.v1.GetDeploymentArtefactsRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.GetDeploymentArtefactsResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncServerStreamingCall(
          getChannel().newCall(getGetDeploymentArtefactsMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * Register a Runner with the Controller.
     * Each runner issue a RegisterRunnerRequest to the ControllerService
     * every 10 seconds to maintain its heartbeat.
     * </pre>
     */
    public io.grpc.stub.StreamObserver<xyz.block.ftl.v1.RegisterRunnerRequest> registerRunner(
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.RegisterRunnerResponse> responseObserver) {
      return io.grpc.stub.ClientCalls.asyncClientStreamingCall(
          getChannel().newCall(getRegisterRunnerMethod(), getCallOptions()), responseObserver);
    }

    /**
     * <pre>
     * Update an existing deployment.
     * </pre>
     */
    public void updateDeploy(xyz.block.ftl.v1.UpdateDeployRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.UpdateDeployResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getUpdateDeployMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * Gradually replace an existing deployment with a new one.
     * If a deployment already exists for the module of the new deployment,
     * it will be scaled down and replaced by the new one.
     * </pre>
     */
    public void replaceDeploy(xyz.block.ftl.v1.ReplaceDeployRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.ReplaceDeployResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getReplaceDeployMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * Stream logs from a deployment
     * </pre>
     */
    public io.grpc.stub.StreamObserver<xyz.block.ftl.v1.StreamDeploymentLogsRequest> streamDeploymentLogs(
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.StreamDeploymentLogsResponse> responseObserver) {
      return io.grpc.stub.ClientCalls.asyncClientStreamingCall(
          getChannel().newCall(getStreamDeploymentLogsMethod(), getCallOptions()), responseObserver);
    }
  }

  /**
   * A stub to allow clients to do synchronous rpc calls to service ControllerService.
   */
  public static final class ControllerServiceBlockingStub
      extends io.grpc.stub.AbstractBlockingStub<ControllerServiceBlockingStub> {
    private ControllerServiceBlockingStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected ControllerServiceBlockingStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new ControllerServiceBlockingStub(channel, callOptions);
    }

    /**
     * <pre>
     * Ping service for readiness.
     * </pre>
     */
    public xyz.block.ftl.v1.PingResponse ping(xyz.block.ftl.v1.PingRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getPingMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * List "processes" running on the cluster.
     * </pre>
     */
    public xyz.block.ftl.v1.ProcessListResponse processList(xyz.block.ftl.v1.ProcessListRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getProcessListMethod(), getCallOptions(), request);
    }

    /**
     */
    public xyz.block.ftl.v1.StatusResponse status(xyz.block.ftl.v1.StatusRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getStatusMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Get list of artefacts that differ between the server and client.
     * </pre>
     */
    public xyz.block.ftl.v1.GetArtefactDiffsResponse getArtefactDiffs(xyz.block.ftl.v1.GetArtefactDiffsRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getGetArtefactDiffsMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Upload an artefact to the server.
     * </pre>
     */
    public xyz.block.ftl.v1.UploadArtefactResponse uploadArtefact(xyz.block.ftl.v1.UploadArtefactRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getUploadArtefactMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Create a deployment.
     * </pre>
     */
    public xyz.block.ftl.v1.CreateDeploymentResponse createDeployment(xyz.block.ftl.v1.CreateDeploymentRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getCreateDeploymentMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Get the schema and artefact metadata for a deployment.
     * </pre>
     */
    public xyz.block.ftl.v1.GetDeploymentResponse getDeployment(xyz.block.ftl.v1.GetDeploymentRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getGetDeploymentMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Stream deployment artefacts from the server.
     * Each artefact is streamed one after the other as a sequence of max 1MB
     * chunks.
     * </pre>
     */
    public java.util.Iterator<xyz.block.ftl.v1.GetDeploymentArtefactsResponse> getDeploymentArtefacts(
        xyz.block.ftl.v1.GetDeploymentArtefactsRequest request) {
      return io.grpc.stub.ClientCalls.blockingServerStreamingCall(
          getChannel(), getGetDeploymentArtefactsMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Update an existing deployment.
     * </pre>
     */
    public xyz.block.ftl.v1.UpdateDeployResponse updateDeploy(xyz.block.ftl.v1.UpdateDeployRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getUpdateDeployMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Gradually replace an existing deployment with a new one.
     * If a deployment already exists for the module of the new deployment,
     * it will be scaled down and replaced by the new one.
     * </pre>
     */
    public xyz.block.ftl.v1.ReplaceDeployResponse replaceDeploy(xyz.block.ftl.v1.ReplaceDeployRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getReplaceDeployMethod(), getCallOptions(), request);
    }
  }

  /**
   * A stub to allow clients to do ListenableFuture-style rpc calls to service ControllerService.
   */
  public static final class ControllerServiceFutureStub
      extends io.grpc.stub.AbstractFutureStub<ControllerServiceFutureStub> {
    private ControllerServiceFutureStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected ControllerServiceFutureStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new ControllerServiceFutureStub(channel, callOptions);
    }

    /**
     * <pre>
     * Ping service for readiness.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<xyz.block.ftl.v1.PingResponse> ping(
        xyz.block.ftl.v1.PingRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getPingMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * List "processes" running on the cluster.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<xyz.block.ftl.v1.ProcessListResponse> processList(
        xyz.block.ftl.v1.ProcessListRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getProcessListMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<xyz.block.ftl.v1.StatusResponse> status(
        xyz.block.ftl.v1.StatusRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getStatusMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * Get list of artefacts that differ between the server and client.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<xyz.block.ftl.v1.GetArtefactDiffsResponse> getArtefactDiffs(
        xyz.block.ftl.v1.GetArtefactDiffsRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getGetArtefactDiffsMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * Upload an artefact to the server.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<xyz.block.ftl.v1.UploadArtefactResponse> uploadArtefact(
        xyz.block.ftl.v1.UploadArtefactRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getUploadArtefactMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * Create a deployment.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<xyz.block.ftl.v1.CreateDeploymentResponse> createDeployment(
        xyz.block.ftl.v1.CreateDeploymentRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getCreateDeploymentMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * Get the schema and artefact metadata for a deployment.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<xyz.block.ftl.v1.GetDeploymentResponse> getDeployment(
        xyz.block.ftl.v1.GetDeploymentRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getGetDeploymentMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * Update an existing deployment.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<xyz.block.ftl.v1.UpdateDeployResponse> updateDeploy(
        xyz.block.ftl.v1.UpdateDeployRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getUpdateDeployMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * Gradually replace an existing deployment with a new one.
     * If a deployment already exists for the module of the new deployment,
     * it will be scaled down and replaced by the new one.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<xyz.block.ftl.v1.ReplaceDeployResponse> replaceDeploy(
        xyz.block.ftl.v1.ReplaceDeployRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getReplaceDeployMethod(), getCallOptions()), request);
    }
  }

  private static final int METHODID_PING = 0;
  private static final int METHODID_PROCESS_LIST = 1;
  private static final int METHODID_STATUS = 2;
  private static final int METHODID_GET_ARTEFACT_DIFFS = 3;
  private static final int METHODID_UPLOAD_ARTEFACT = 4;
  private static final int METHODID_CREATE_DEPLOYMENT = 5;
  private static final int METHODID_GET_DEPLOYMENT = 6;
  private static final int METHODID_GET_DEPLOYMENT_ARTEFACTS = 7;
  private static final int METHODID_UPDATE_DEPLOY = 8;
  private static final int METHODID_REPLACE_DEPLOY = 9;
  private static final int METHODID_REGISTER_RUNNER = 10;
  private static final int METHODID_STREAM_DEPLOYMENT_LOGS = 11;

  private static final class MethodHandlers<Req, Resp> implements
      io.grpc.stub.ServerCalls.UnaryMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.ServerStreamingMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.ClientStreamingMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.BidiStreamingMethod<Req, Resp> {
    private final AsyncService serviceImpl;
    private final int methodId;

    MethodHandlers(AsyncService serviceImpl, int methodId) {
      this.serviceImpl = serviceImpl;
      this.methodId = methodId;
    }

    @java.lang.Override
    @java.lang.SuppressWarnings("unchecked")
    public void invoke(Req request, io.grpc.stub.StreamObserver<Resp> responseObserver) {
      switch (methodId) {
        case METHODID_PING:
          serviceImpl.ping((xyz.block.ftl.v1.PingRequest) request,
              (io.grpc.stub.StreamObserver<xyz.block.ftl.v1.PingResponse>) responseObserver);
          break;
        case METHODID_PROCESS_LIST:
          serviceImpl.processList((xyz.block.ftl.v1.ProcessListRequest) request,
              (io.grpc.stub.StreamObserver<xyz.block.ftl.v1.ProcessListResponse>) responseObserver);
          break;
        case METHODID_STATUS:
          serviceImpl.status((xyz.block.ftl.v1.StatusRequest) request,
              (io.grpc.stub.StreamObserver<xyz.block.ftl.v1.StatusResponse>) responseObserver);
          break;
        case METHODID_GET_ARTEFACT_DIFFS:
          serviceImpl.getArtefactDiffs((xyz.block.ftl.v1.GetArtefactDiffsRequest) request,
              (io.grpc.stub.StreamObserver<xyz.block.ftl.v1.GetArtefactDiffsResponse>) responseObserver);
          break;
        case METHODID_UPLOAD_ARTEFACT:
          serviceImpl.uploadArtefact((xyz.block.ftl.v1.UploadArtefactRequest) request,
              (io.grpc.stub.StreamObserver<xyz.block.ftl.v1.UploadArtefactResponse>) responseObserver);
          break;
        case METHODID_CREATE_DEPLOYMENT:
          serviceImpl.createDeployment((xyz.block.ftl.v1.CreateDeploymentRequest) request,
              (io.grpc.stub.StreamObserver<xyz.block.ftl.v1.CreateDeploymentResponse>) responseObserver);
          break;
        case METHODID_GET_DEPLOYMENT:
          serviceImpl.getDeployment((xyz.block.ftl.v1.GetDeploymentRequest) request,
              (io.grpc.stub.StreamObserver<xyz.block.ftl.v1.GetDeploymentResponse>) responseObserver);
          break;
        case METHODID_GET_DEPLOYMENT_ARTEFACTS:
          serviceImpl.getDeploymentArtefacts((xyz.block.ftl.v1.GetDeploymentArtefactsRequest) request,
              (io.grpc.stub.StreamObserver<xyz.block.ftl.v1.GetDeploymentArtefactsResponse>) responseObserver);
          break;
        case METHODID_UPDATE_DEPLOY:
          serviceImpl.updateDeploy((xyz.block.ftl.v1.UpdateDeployRequest) request,
              (io.grpc.stub.StreamObserver<xyz.block.ftl.v1.UpdateDeployResponse>) responseObserver);
          break;
        case METHODID_REPLACE_DEPLOY:
          serviceImpl.replaceDeploy((xyz.block.ftl.v1.ReplaceDeployRequest) request,
              (io.grpc.stub.StreamObserver<xyz.block.ftl.v1.ReplaceDeployResponse>) responseObserver);
          break;
        default:
          throw new AssertionError();
      }
    }

    @java.lang.Override
    @java.lang.SuppressWarnings("unchecked")
    public io.grpc.stub.StreamObserver<Req> invoke(
        io.grpc.stub.StreamObserver<Resp> responseObserver) {
      switch (methodId) {
        case METHODID_REGISTER_RUNNER:
          return (io.grpc.stub.StreamObserver<Req>) serviceImpl.registerRunner(
              (io.grpc.stub.StreamObserver<xyz.block.ftl.v1.RegisterRunnerResponse>) responseObserver);
        case METHODID_STREAM_DEPLOYMENT_LOGS:
          return (io.grpc.stub.StreamObserver<Req>) serviceImpl.streamDeploymentLogs(
              (io.grpc.stub.StreamObserver<xyz.block.ftl.v1.StreamDeploymentLogsResponse>) responseObserver);
        default:
          throw new AssertionError();
      }
    }
  }

  public static final io.grpc.ServerServiceDefinition bindService(AsyncService service) {
    return io.grpc.ServerServiceDefinition.builder(getServiceDescriptor())
        .addMethod(
          getPingMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              xyz.block.ftl.v1.PingRequest,
              xyz.block.ftl.v1.PingResponse>(
                service, METHODID_PING)))
        .addMethod(
          getProcessListMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              xyz.block.ftl.v1.ProcessListRequest,
              xyz.block.ftl.v1.ProcessListResponse>(
                service, METHODID_PROCESS_LIST)))
        .addMethod(
          getStatusMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              xyz.block.ftl.v1.StatusRequest,
              xyz.block.ftl.v1.StatusResponse>(
                service, METHODID_STATUS)))
        .addMethod(
          getGetArtefactDiffsMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              xyz.block.ftl.v1.GetArtefactDiffsRequest,
              xyz.block.ftl.v1.GetArtefactDiffsResponse>(
                service, METHODID_GET_ARTEFACT_DIFFS)))
        .addMethod(
          getUploadArtefactMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              xyz.block.ftl.v1.UploadArtefactRequest,
              xyz.block.ftl.v1.UploadArtefactResponse>(
                service, METHODID_UPLOAD_ARTEFACT)))
        .addMethod(
          getCreateDeploymentMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              xyz.block.ftl.v1.CreateDeploymentRequest,
              xyz.block.ftl.v1.CreateDeploymentResponse>(
                service, METHODID_CREATE_DEPLOYMENT)))
        .addMethod(
          getGetDeploymentMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              xyz.block.ftl.v1.GetDeploymentRequest,
              xyz.block.ftl.v1.GetDeploymentResponse>(
                service, METHODID_GET_DEPLOYMENT)))
        .addMethod(
          getGetDeploymentArtefactsMethod(),
          io.grpc.stub.ServerCalls.asyncServerStreamingCall(
            new MethodHandlers<
              xyz.block.ftl.v1.GetDeploymentArtefactsRequest,
              xyz.block.ftl.v1.GetDeploymentArtefactsResponse>(
                service, METHODID_GET_DEPLOYMENT_ARTEFACTS)))
        .addMethod(
          getRegisterRunnerMethod(),
          io.grpc.stub.ServerCalls.asyncClientStreamingCall(
            new MethodHandlers<
              xyz.block.ftl.v1.RegisterRunnerRequest,
              xyz.block.ftl.v1.RegisterRunnerResponse>(
                service, METHODID_REGISTER_RUNNER)))
        .addMethod(
          getUpdateDeployMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              xyz.block.ftl.v1.UpdateDeployRequest,
              xyz.block.ftl.v1.UpdateDeployResponse>(
                service, METHODID_UPDATE_DEPLOY)))
        .addMethod(
          getReplaceDeployMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              xyz.block.ftl.v1.ReplaceDeployRequest,
              xyz.block.ftl.v1.ReplaceDeployResponse>(
                service, METHODID_REPLACE_DEPLOY)))
        .addMethod(
          getStreamDeploymentLogsMethod(),
          io.grpc.stub.ServerCalls.asyncClientStreamingCall(
            new MethodHandlers<
              xyz.block.ftl.v1.StreamDeploymentLogsRequest,
              xyz.block.ftl.v1.StreamDeploymentLogsResponse>(
                service, METHODID_STREAM_DEPLOYMENT_LOGS)))
        .build();
  }

  private static abstract class ControllerServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoFileDescriptorSupplier, io.grpc.protobuf.ProtoServiceDescriptorSupplier {
    ControllerServiceBaseDescriptorSupplier() {}

    @java.lang.Override
    public com.google.protobuf.Descriptors.FileDescriptor getFileDescriptor() {
      return xyz.block.ftl.v1.ControllerOuterClass.getDescriptor();
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.ServiceDescriptor getServiceDescriptor() {
      return getFileDescriptor().findServiceByName("ControllerService");
    }
  }

  private static final class ControllerServiceFileDescriptorSupplier
      extends ControllerServiceBaseDescriptorSupplier {
    ControllerServiceFileDescriptorSupplier() {}
  }

  private static final class ControllerServiceMethodDescriptorSupplier
      extends ControllerServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoMethodDescriptorSupplier {
    private final java.lang.String methodName;

    ControllerServiceMethodDescriptorSupplier(java.lang.String methodName) {
      this.methodName = methodName;
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.MethodDescriptor getMethodDescriptor() {
      return getServiceDescriptor().findMethodByName(methodName);
    }
  }

  private static volatile io.grpc.ServiceDescriptor serviceDescriptor;

  public static io.grpc.ServiceDescriptor getServiceDescriptor() {
    io.grpc.ServiceDescriptor result = serviceDescriptor;
    if (result == null) {
      synchronized (ControllerServiceGrpc.class) {
        result = serviceDescriptor;
        if (result == null) {
          serviceDescriptor = result = io.grpc.ServiceDescriptor.newBuilder(SERVICE_NAME)
              .setSchemaDescriptor(new ControllerServiceFileDescriptorSupplier())
              .addMethod(getPingMethod())
              .addMethod(getProcessListMethod())
              .addMethod(getStatusMethod())
              .addMethod(getGetArtefactDiffsMethod())
              .addMethod(getUploadArtefactMethod())
              .addMethod(getCreateDeploymentMethod())
              .addMethod(getGetDeploymentMethod())
              .addMethod(getGetDeploymentArtefactsMethod())
              .addMethod(getRegisterRunnerMethod())
              .addMethod(getUpdateDeployMethod())
              .addMethod(getReplaceDeployMethod())
              .addMethod(getStreamDeploymentLogsMethod())
              .build();
        }
      }
    }
    return result;
  }
}
