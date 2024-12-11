package xyz.block.ftl.provisioner.v1beta1;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 */
@javax.annotation.Generated(
    value = "by gRPC proto compiler (version 1.68.2)",
    comments = "Source: xyz/block/ftl/provisioner/v1beta1/service.proto")
@io.grpc.stub.annotations.GrpcGenerated
public final class ProvisionerServiceGrpc {

  private ProvisionerServiceGrpc() {}

  public static final java.lang.String SERVICE_NAME = "xyz.block.ftl.provisioner.v1beta1.ProvisionerService";

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
    if ((getPingMethod = ProvisionerServiceGrpc.getPingMethod) == null) {
      synchronized (ProvisionerServiceGrpc.class) {
        if ((getPingMethod = ProvisionerServiceGrpc.getPingMethod) == null) {
          ProvisionerServiceGrpc.getPingMethod = getPingMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.v1.PingRequest, xyz.block.ftl.v1.PingResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "Ping"))
              .setSafe(true)
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.PingRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.PingResponse.getDefaultInstance()))
              .setSchemaDescriptor(new ProvisionerServiceMethodDescriptorSupplier("Ping"))
              .build();
        }
      }
    }
    return getPingMethod;
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
    if ((getStatusMethod = ProvisionerServiceGrpc.getStatusMethod) == null) {
      synchronized (ProvisionerServiceGrpc.class) {
        if ((getStatusMethod = ProvisionerServiceGrpc.getStatusMethod) == null) {
          ProvisionerServiceGrpc.getStatusMethod = getStatusMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.v1.StatusRequest, xyz.block.ftl.v1.StatusResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "Status"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.StatusRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.StatusResponse.getDefaultInstance()))
              .setSchemaDescriptor(new ProvisionerServiceMethodDescriptorSupplier("Status"))
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
    if ((getGetArtefactDiffsMethod = ProvisionerServiceGrpc.getGetArtefactDiffsMethod) == null) {
      synchronized (ProvisionerServiceGrpc.class) {
        if ((getGetArtefactDiffsMethod = ProvisionerServiceGrpc.getGetArtefactDiffsMethod) == null) {
          ProvisionerServiceGrpc.getGetArtefactDiffsMethod = getGetArtefactDiffsMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.v1.GetArtefactDiffsRequest, xyz.block.ftl.v1.GetArtefactDiffsResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "GetArtefactDiffs"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.GetArtefactDiffsRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.GetArtefactDiffsResponse.getDefaultInstance()))
              .setSchemaDescriptor(new ProvisionerServiceMethodDescriptorSupplier("GetArtefactDiffs"))
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
    if ((getUploadArtefactMethod = ProvisionerServiceGrpc.getUploadArtefactMethod) == null) {
      synchronized (ProvisionerServiceGrpc.class) {
        if ((getUploadArtefactMethod = ProvisionerServiceGrpc.getUploadArtefactMethod) == null) {
          ProvisionerServiceGrpc.getUploadArtefactMethod = getUploadArtefactMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.v1.UploadArtefactRequest, xyz.block.ftl.v1.UploadArtefactResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "UploadArtefact"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.UploadArtefactRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.UploadArtefactResponse.getDefaultInstance()))
              .setSchemaDescriptor(new ProvisionerServiceMethodDescriptorSupplier("UploadArtefact"))
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
    if ((getCreateDeploymentMethod = ProvisionerServiceGrpc.getCreateDeploymentMethod) == null) {
      synchronized (ProvisionerServiceGrpc.class) {
        if ((getCreateDeploymentMethod = ProvisionerServiceGrpc.getCreateDeploymentMethod) == null) {
          ProvisionerServiceGrpc.getCreateDeploymentMethod = getCreateDeploymentMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.v1.CreateDeploymentRequest, xyz.block.ftl.v1.CreateDeploymentResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "CreateDeployment"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.CreateDeploymentRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.CreateDeploymentResponse.getDefaultInstance()))
              .setSchemaDescriptor(new ProvisionerServiceMethodDescriptorSupplier("CreateDeployment"))
              .build();
        }
      }
    }
    return getCreateDeploymentMethod;
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
    if ((getUpdateDeployMethod = ProvisionerServiceGrpc.getUpdateDeployMethod) == null) {
      synchronized (ProvisionerServiceGrpc.class) {
        if ((getUpdateDeployMethod = ProvisionerServiceGrpc.getUpdateDeployMethod) == null) {
          ProvisionerServiceGrpc.getUpdateDeployMethod = getUpdateDeployMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.v1.UpdateDeployRequest, xyz.block.ftl.v1.UpdateDeployResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "UpdateDeploy"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.UpdateDeployRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.UpdateDeployResponse.getDefaultInstance()))
              .setSchemaDescriptor(new ProvisionerServiceMethodDescriptorSupplier("UpdateDeploy"))
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
    if ((getReplaceDeployMethod = ProvisionerServiceGrpc.getReplaceDeployMethod) == null) {
      synchronized (ProvisionerServiceGrpc.class) {
        if ((getReplaceDeployMethod = ProvisionerServiceGrpc.getReplaceDeployMethod) == null) {
          ProvisionerServiceGrpc.getReplaceDeployMethod = getReplaceDeployMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.v1.ReplaceDeployRequest, xyz.block.ftl.v1.ReplaceDeployResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "ReplaceDeploy"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.ReplaceDeployRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.ReplaceDeployResponse.getDefaultInstance()))
              .setSchemaDescriptor(new ProvisionerServiceMethodDescriptorSupplier("ReplaceDeploy"))
              .build();
        }
      }
    }
    return getReplaceDeployMethod;
  }

  /**
   * Creates a new async stub that supports all call types for the service
   */
  public static ProvisionerServiceStub newStub(io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<ProvisionerServiceStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<ProvisionerServiceStub>() {
        @java.lang.Override
        public ProvisionerServiceStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new ProvisionerServiceStub(channel, callOptions);
        }
      };
    return ProvisionerServiceStub.newStub(factory, channel);
  }

  /**
   * Creates a new blocking-style stub that supports unary and streaming output calls on the service
   */
  public static ProvisionerServiceBlockingStub newBlockingStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<ProvisionerServiceBlockingStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<ProvisionerServiceBlockingStub>() {
        @java.lang.Override
        public ProvisionerServiceBlockingStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new ProvisionerServiceBlockingStub(channel, callOptions);
        }
      };
    return ProvisionerServiceBlockingStub.newStub(factory, channel);
  }

  /**
   * Creates a new ListenableFuture-style stub that supports unary calls on the service
   */
  public static ProvisionerServiceFutureStub newFutureStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<ProvisionerServiceFutureStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<ProvisionerServiceFutureStub>() {
        @java.lang.Override
        public ProvisionerServiceFutureStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new ProvisionerServiceFutureStub(channel, callOptions);
        }
      };
    return ProvisionerServiceFutureStub.newStub(factory, channel);
  }

  /**
   */
  public interface AsyncService {

    /**
     */
    default void ping(xyz.block.ftl.v1.PingRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.PingResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getPingMethod(), responseObserver);
    }

    /**
     */
    default void status(xyz.block.ftl.v1.StatusRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.StatusResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getStatusMethod(), responseObserver);
    }

    /**
     */
    default void getArtefactDiffs(xyz.block.ftl.v1.GetArtefactDiffsRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.GetArtefactDiffsResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getGetArtefactDiffsMethod(), responseObserver);
    }

    /**
     */
    default void uploadArtefact(xyz.block.ftl.v1.UploadArtefactRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.UploadArtefactResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getUploadArtefactMethod(), responseObserver);
    }

    /**
     */
    default void createDeployment(xyz.block.ftl.v1.CreateDeploymentRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.CreateDeploymentResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getCreateDeploymentMethod(), responseObserver);
    }

    /**
     */
    default void updateDeploy(xyz.block.ftl.v1.UpdateDeployRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.UpdateDeployResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getUpdateDeployMethod(), responseObserver);
    }

    /**
     */
    default void replaceDeploy(xyz.block.ftl.v1.ReplaceDeployRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.ReplaceDeployResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getReplaceDeployMethod(), responseObserver);
    }
  }

  /**
   * Base class for the server implementation of the service ProvisionerService.
   */
  public static abstract class ProvisionerServiceImplBase
      implements io.grpc.BindableService, AsyncService {

    @java.lang.Override public final io.grpc.ServerServiceDefinition bindService() {
      return ProvisionerServiceGrpc.bindService(this);
    }
  }

  /**
   * A stub to allow clients to do asynchronous rpc calls to service ProvisionerService.
   */
  public static final class ProvisionerServiceStub
      extends io.grpc.stub.AbstractAsyncStub<ProvisionerServiceStub> {
    private ProvisionerServiceStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected ProvisionerServiceStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new ProvisionerServiceStub(channel, callOptions);
    }

    /**
     */
    public void ping(xyz.block.ftl.v1.PingRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.PingResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getPingMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void status(xyz.block.ftl.v1.StatusRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.StatusResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getStatusMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void getArtefactDiffs(xyz.block.ftl.v1.GetArtefactDiffsRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.GetArtefactDiffsResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getGetArtefactDiffsMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void uploadArtefact(xyz.block.ftl.v1.UploadArtefactRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.UploadArtefactResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getUploadArtefactMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void createDeployment(xyz.block.ftl.v1.CreateDeploymentRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.CreateDeploymentResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getCreateDeploymentMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void updateDeploy(xyz.block.ftl.v1.UpdateDeployRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.UpdateDeployResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getUpdateDeployMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void replaceDeploy(xyz.block.ftl.v1.ReplaceDeployRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.ReplaceDeployResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getReplaceDeployMethod(), getCallOptions()), request, responseObserver);
    }
  }

  /**
   * A stub to allow clients to do synchronous rpc calls to service ProvisionerService.
   */
  public static final class ProvisionerServiceBlockingStub
      extends io.grpc.stub.AbstractBlockingStub<ProvisionerServiceBlockingStub> {
    private ProvisionerServiceBlockingStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected ProvisionerServiceBlockingStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new ProvisionerServiceBlockingStub(channel, callOptions);
    }

    /**
     */
    public xyz.block.ftl.v1.PingResponse ping(xyz.block.ftl.v1.PingRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getPingMethod(), getCallOptions(), request);
    }

    /**
     */
    public xyz.block.ftl.v1.StatusResponse status(xyz.block.ftl.v1.StatusRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getStatusMethod(), getCallOptions(), request);
    }

    /**
     */
    public xyz.block.ftl.v1.GetArtefactDiffsResponse getArtefactDiffs(xyz.block.ftl.v1.GetArtefactDiffsRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getGetArtefactDiffsMethod(), getCallOptions(), request);
    }

    /**
     */
    public xyz.block.ftl.v1.UploadArtefactResponse uploadArtefact(xyz.block.ftl.v1.UploadArtefactRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getUploadArtefactMethod(), getCallOptions(), request);
    }

    /**
     */
    public xyz.block.ftl.v1.CreateDeploymentResponse createDeployment(xyz.block.ftl.v1.CreateDeploymentRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getCreateDeploymentMethod(), getCallOptions(), request);
    }

    /**
     */
    public xyz.block.ftl.v1.UpdateDeployResponse updateDeploy(xyz.block.ftl.v1.UpdateDeployRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getUpdateDeployMethod(), getCallOptions(), request);
    }

    /**
     */
    public xyz.block.ftl.v1.ReplaceDeployResponse replaceDeploy(xyz.block.ftl.v1.ReplaceDeployRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getReplaceDeployMethod(), getCallOptions(), request);
    }
  }

  /**
   * A stub to allow clients to do ListenableFuture-style rpc calls to service ProvisionerService.
   */
  public static final class ProvisionerServiceFutureStub
      extends io.grpc.stub.AbstractFutureStub<ProvisionerServiceFutureStub> {
    private ProvisionerServiceFutureStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected ProvisionerServiceFutureStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new ProvisionerServiceFutureStub(channel, callOptions);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<xyz.block.ftl.v1.PingResponse> ping(
        xyz.block.ftl.v1.PingRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getPingMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<xyz.block.ftl.v1.StatusResponse> status(
        xyz.block.ftl.v1.StatusRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getStatusMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<xyz.block.ftl.v1.GetArtefactDiffsResponse> getArtefactDiffs(
        xyz.block.ftl.v1.GetArtefactDiffsRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getGetArtefactDiffsMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<xyz.block.ftl.v1.UploadArtefactResponse> uploadArtefact(
        xyz.block.ftl.v1.UploadArtefactRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getUploadArtefactMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<xyz.block.ftl.v1.CreateDeploymentResponse> createDeployment(
        xyz.block.ftl.v1.CreateDeploymentRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getCreateDeploymentMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<xyz.block.ftl.v1.UpdateDeployResponse> updateDeploy(
        xyz.block.ftl.v1.UpdateDeployRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getUpdateDeployMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<xyz.block.ftl.v1.ReplaceDeployResponse> replaceDeploy(
        xyz.block.ftl.v1.ReplaceDeployRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getReplaceDeployMethod(), getCallOptions()), request);
    }
  }

  private static final int METHODID_PING = 0;
  private static final int METHODID_STATUS = 1;
  private static final int METHODID_GET_ARTEFACT_DIFFS = 2;
  private static final int METHODID_UPLOAD_ARTEFACT = 3;
  private static final int METHODID_CREATE_DEPLOYMENT = 4;
  private static final int METHODID_UPDATE_DEPLOY = 5;
  private static final int METHODID_REPLACE_DEPLOY = 6;

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
        .build();
  }

  private static abstract class ProvisionerServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoFileDescriptorSupplier, io.grpc.protobuf.ProtoServiceDescriptorSupplier {
    ProvisionerServiceBaseDescriptorSupplier() {}

    @java.lang.Override
    public com.google.protobuf.Descriptors.FileDescriptor getFileDescriptor() {
      return xyz.block.ftl.provisioner.v1beta1.Service.getDescriptor();
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.ServiceDescriptor getServiceDescriptor() {
      return getFileDescriptor().findServiceByName("ProvisionerService");
    }
  }

  private static final class ProvisionerServiceFileDescriptorSupplier
      extends ProvisionerServiceBaseDescriptorSupplier {
    ProvisionerServiceFileDescriptorSupplier() {}
  }

  private static final class ProvisionerServiceMethodDescriptorSupplier
      extends ProvisionerServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoMethodDescriptorSupplier {
    private final java.lang.String methodName;

    ProvisionerServiceMethodDescriptorSupplier(java.lang.String methodName) {
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
      synchronized (ProvisionerServiceGrpc.class) {
        result = serviceDescriptor;
        if (result == null) {
          serviceDescriptor = result = io.grpc.ServiceDescriptor.newBuilder(SERVICE_NAME)
              .setSchemaDescriptor(new ProvisionerServiceFileDescriptorSupplier())
              .addMethod(getPingMethod())
              .addMethod(getStatusMethod())
              .addMethod(getGetArtefactDiffsMethod())
              .addMethod(getUploadArtefactMethod())
              .addMethod(getCreateDeploymentMethod())
              .addMethod(getUpdateDeployMethod())
              .addMethod(getReplaceDeployMethod())
              .build();
        }
      }
    }
    return result;
  }
}
