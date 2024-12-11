package xyz.block.ftl.v1;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 */
@javax.annotation.Generated(
    value = "by gRPC proto compiler (version 1.68.2)",
    comments = "Source: xyz/block/ftl/v1/schemaservice.proto")
@io.grpc.stub.annotations.GrpcGenerated
public final class SchemaServiceGrpc {

  private SchemaServiceGrpc() {}

  public static final java.lang.String SERVICE_NAME = "xyz.block.ftl.v1.SchemaService";

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
    if ((getPingMethod = SchemaServiceGrpc.getPingMethod) == null) {
      synchronized (SchemaServiceGrpc.class) {
        if ((getPingMethod = SchemaServiceGrpc.getPingMethod) == null) {
          SchemaServiceGrpc.getPingMethod = getPingMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.v1.PingRequest, xyz.block.ftl.v1.PingResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "Ping"))
              .setSafe(true)
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.PingRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.PingResponse.getDefaultInstance()))
              .setSchemaDescriptor(new SchemaServiceMethodDescriptorSupplier("Ping"))
              .build();
        }
      }
    }
    return getPingMethod;
  }

  private static volatile io.grpc.MethodDescriptor<xyz.block.ftl.v1.GetSchemaRequest,
      xyz.block.ftl.v1.GetSchemaResponse> getGetSchemaMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "GetSchema",
      requestType = xyz.block.ftl.v1.GetSchemaRequest.class,
      responseType = xyz.block.ftl.v1.GetSchemaResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<xyz.block.ftl.v1.GetSchemaRequest,
      xyz.block.ftl.v1.GetSchemaResponse> getGetSchemaMethod() {
    io.grpc.MethodDescriptor<xyz.block.ftl.v1.GetSchemaRequest, xyz.block.ftl.v1.GetSchemaResponse> getGetSchemaMethod;
    if ((getGetSchemaMethod = SchemaServiceGrpc.getGetSchemaMethod) == null) {
      synchronized (SchemaServiceGrpc.class) {
        if ((getGetSchemaMethod = SchemaServiceGrpc.getGetSchemaMethod) == null) {
          SchemaServiceGrpc.getGetSchemaMethod = getGetSchemaMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.v1.GetSchemaRequest, xyz.block.ftl.v1.GetSchemaResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "GetSchema"))
              .setSafe(true)
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.GetSchemaRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.GetSchemaResponse.getDefaultInstance()))
              .setSchemaDescriptor(new SchemaServiceMethodDescriptorSupplier("GetSchema"))
              .build();
        }
      }
    }
    return getGetSchemaMethod;
  }

  private static volatile io.grpc.MethodDescriptor<xyz.block.ftl.v1.PullSchemaRequest,
      xyz.block.ftl.v1.PullSchemaResponse> getPullSchemaMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "PullSchema",
      requestType = xyz.block.ftl.v1.PullSchemaRequest.class,
      responseType = xyz.block.ftl.v1.PullSchemaResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.SERVER_STREAMING)
  public static io.grpc.MethodDescriptor<xyz.block.ftl.v1.PullSchemaRequest,
      xyz.block.ftl.v1.PullSchemaResponse> getPullSchemaMethod() {
    io.grpc.MethodDescriptor<xyz.block.ftl.v1.PullSchemaRequest, xyz.block.ftl.v1.PullSchemaResponse> getPullSchemaMethod;
    if ((getPullSchemaMethod = SchemaServiceGrpc.getPullSchemaMethod) == null) {
      synchronized (SchemaServiceGrpc.class) {
        if ((getPullSchemaMethod = SchemaServiceGrpc.getPullSchemaMethod) == null) {
          SchemaServiceGrpc.getPullSchemaMethod = getPullSchemaMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.v1.PullSchemaRequest, xyz.block.ftl.v1.PullSchemaResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.SERVER_STREAMING)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "PullSchema"))
              .setSafe(true)
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.PullSchemaRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.PullSchemaResponse.getDefaultInstance()))
              .setSchemaDescriptor(new SchemaServiceMethodDescriptorSupplier("PullSchema"))
              .build();
        }
      }
    }
    return getPullSchemaMethod;
  }

  private static volatile io.grpc.MethodDescriptor<xyz.block.ftl.v1.UpdateDeploymentRuntimeRequest,
      xyz.block.ftl.v1.UpdateDeploymentRuntimeResponse> getUpdateDeploymentRuntimeMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "UpdateDeploymentRuntime",
      requestType = xyz.block.ftl.v1.UpdateDeploymentRuntimeRequest.class,
      responseType = xyz.block.ftl.v1.UpdateDeploymentRuntimeResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<xyz.block.ftl.v1.UpdateDeploymentRuntimeRequest,
      xyz.block.ftl.v1.UpdateDeploymentRuntimeResponse> getUpdateDeploymentRuntimeMethod() {
    io.grpc.MethodDescriptor<xyz.block.ftl.v1.UpdateDeploymentRuntimeRequest, xyz.block.ftl.v1.UpdateDeploymentRuntimeResponse> getUpdateDeploymentRuntimeMethod;
    if ((getUpdateDeploymentRuntimeMethod = SchemaServiceGrpc.getUpdateDeploymentRuntimeMethod) == null) {
      synchronized (SchemaServiceGrpc.class) {
        if ((getUpdateDeploymentRuntimeMethod = SchemaServiceGrpc.getUpdateDeploymentRuntimeMethod) == null) {
          SchemaServiceGrpc.getUpdateDeploymentRuntimeMethod = getUpdateDeploymentRuntimeMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.v1.UpdateDeploymentRuntimeRequest, xyz.block.ftl.v1.UpdateDeploymentRuntimeResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "UpdateDeploymentRuntime"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.UpdateDeploymentRuntimeRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.UpdateDeploymentRuntimeResponse.getDefaultInstance()))
              .setSchemaDescriptor(new SchemaServiceMethodDescriptorSupplier("UpdateDeploymentRuntime"))
              .build();
        }
      }
    }
    return getUpdateDeploymentRuntimeMethod;
  }

  /**
   * Creates a new async stub that supports all call types for the service
   */
  public static SchemaServiceStub newStub(io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<SchemaServiceStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<SchemaServiceStub>() {
        @java.lang.Override
        public SchemaServiceStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new SchemaServiceStub(channel, callOptions);
        }
      };
    return SchemaServiceStub.newStub(factory, channel);
  }

  /**
   * Creates a new blocking-style stub that supports unary and streaming output calls on the service
   */
  public static SchemaServiceBlockingStub newBlockingStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<SchemaServiceBlockingStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<SchemaServiceBlockingStub>() {
        @java.lang.Override
        public SchemaServiceBlockingStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new SchemaServiceBlockingStub(channel, callOptions);
        }
      };
    return SchemaServiceBlockingStub.newStub(factory, channel);
  }

  /**
   * Creates a new ListenableFuture-style stub that supports unary calls on the service
   */
  public static SchemaServiceFutureStub newFutureStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<SchemaServiceFutureStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<SchemaServiceFutureStub>() {
        @java.lang.Override
        public SchemaServiceFutureStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new SchemaServiceFutureStub(channel, callOptions);
        }
      };
    return SchemaServiceFutureStub.newStub(factory, channel);
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
     * Get the full schema.
     * </pre>
     */
    default void getSchema(xyz.block.ftl.v1.GetSchemaRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.GetSchemaResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getGetSchemaMethod(), responseObserver);
    }

    /**
     * <pre>
     * Pull schema changes from the Controller.
     * Note that if there are no deployments this will block indefinitely, making it unsuitable for
     * just retrieving the schema. Use GetSchema for that.
     * </pre>
     */
    default void pullSchema(xyz.block.ftl.v1.PullSchemaRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.PullSchemaResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getPullSchemaMethod(), responseObserver);
    }

    /**
     * <pre>
     * UpdateModuleRuntime is used to update the runtime configuration of a module.
     * </pre>
     */
    default void updateDeploymentRuntime(xyz.block.ftl.v1.UpdateDeploymentRuntimeRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.UpdateDeploymentRuntimeResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getUpdateDeploymentRuntimeMethod(), responseObserver);
    }
  }

  /**
   * Base class for the server implementation of the service SchemaService.
   */
  public static abstract class SchemaServiceImplBase
      implements io.grpc.BindableService, AsyncService {

    @java.lang.Override public final io.grpc.ServerServiceDefinition bindService() {
      return SchemaServiceGrpc.bindService(this);
    }
  }

  /**
   * A stub to allow clients to do asynchronous rpc calls to service SchemaService.
   */
  public static final class SchemaServiceStub
      extends io.grpc.stub.AbstractAsyncStub<SchemaServiceStub> {
    private SchemaServiceStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected SchemaServiceStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new SchemaServiceStub(channel, callOptions);
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
     * Get the full schema.
     * </pre>
     */
    public void getSchema(xyz.block.ftl.v1.GetSchemaRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.GetSchemaResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getGetSchemaMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * Pull schema changes from the Controller.
     * Note that if there are no deployments this will block indefinitely, making it unsuitable for
     * just retrieving the schema. Use GetSchema for that.
     * </pre>
     */
    public void pullSchema(xyz.block.ftl.v1.PullSchemaRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.PullSchemaResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncServerStreamingCall(
          getChannel().newCall(getPullSchemaMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * UpdateModuleRuntime is used to update the runtime configuration of a module.
     * </pre>
     */
    public void updateDeploymentRuntime(xyz.block.ftl.v1.UpdateDeploymentRuntimeRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.UpdateDeploymentRuntimeResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getUpdateDeploymentRuntimeMethod(), getCallOptions()), request, responseObserver);
    }
  }

  /**
   * A stub to allow clients to do synchronous rpc calls to service SchemaService.
   */
  public static final class SchemaServiceBlockingStub
      extends io.grpc.stub.AbstractBlockingStub<SchemaServiceBlockingStub> {
    private SchemaServiceBlockingStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected SchemaServiceBlockingStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new SchemaServiceBlockingStub(channel, callOptions);
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
     * Get the full schema.
     * </pre>
     */
    public xyz.block.ftl.v1.GetSchemaResponse getSchema(xyz.block.ftl.v1.GetSchemaRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getGetSchemaMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Pull schema changes from the Controller.
     * Note that if there are no deployments this will block indefinitely, making it unsuitable for
     * just retrieving the schema. Use GetSchema for that.
     * </pre>
     */
    public java.util.Iterator<xyz.block.ftl.v1.PullSchemaResponse> pullSchema(
        xyz.block.ftl.v1.PullSchemaRequest request) {
      return io.grpc.stub.ClientCalls.blockingServerStreamingCall(
          getChannel(), getPullSchemaMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * UpdateModuleRuntime is used to update the runtime configuration of a module.
     * </pre>
     */
    public xyz.block.ftl.v1.UpdateDeploymentRuntimeResponse updateDeploymentRuntime(xyz.block.ftl.v1.UpdateDeploymentRuntimeRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getUpdateDeploymentRuntimeMethod(), getCallOptions(), request);
    }
  }

  /**
   * A stub to allow clients to do ListenableFuture-style rpc calls to service SchemaService.
   */
  public static final class SchemaServiceFutureStub
      extends io.grpc.stub.AbstractFutureStub<SchemaServiceFutureStub> {
    private SchemaServiceFutureStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected SchemaServiceFutureStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new SchemaServiceFutureStub(channel, callOptions);
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
     * Get the full schema.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<xyz.block.ftl.v1.GetSchemaResponse> getSchema(
        xyz.block.ftl.v1.GetSchemaRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getGetSchemaMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * UpdateModuleRuntime is used to update the runtime configuration of a module.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<xyz.block.ftl.v1.UpdateDeploymentRuntimeResponse> updateDeploymentRuntime(
        xyz.block.ftl.v1.UpdateDeploymentRuntimeRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getUpdateDeploymentRuntimeMethod(), getCallOptions()), request);
    }
  }

  private static final int METHODID_PING = 0;
  private static final int METHODID_GET_SCHEMA = 1;
  private static final int METHODID_PULL_SCHEMA = 2;
  private static final int METHODID_UPDATE_DEPLOYMENT_RUNTIME = 3;

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
        case METHODID_GET_SCHEMA:
          serviceImpl.getSchema((xyz.block.ftl.v1.GetSchemaRequest) request,
              (io.grpc.stub.StreamObserver<xyz.block.ftl.v1.GetSchemaResponse>) responseObserver);
          break;
        case METHODID_PULL_SCHEMA:
          serviceImpl.pullSchema((xyz.block.ftl.v1.PullSchemaRequest) request,
              (io.grpc.stub.StreamObserver<xyz.block.ftl.v1.PullSchemaResponse>) responseObserver);
          break;
        case METHODID_UPDATE_DEPLOYMENT_RUNTIME:
          serviceImpl.updateDeploymentRuntime((xyz.block.ftl.v1.UpdateDeploymentRuntimeRequest) request,
              (io.grpc.stub.StreamObserver<xyz.block.ftl.v1.UpdateDeploymentRuntimeResponse>) responseObserver);
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
          getGetSchemaMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              xyz.block.ftl.v1.GetSchemaRequest,
              xyz.block.ftl.v1.GetSchemaResponse>(
                service, METHODID_GET_SCHEMA)))
        .addMethod(
          getPullSchemaMethod(),
          io.grpc.stub.ServerCalls.asyncServerStreamingCall(
            new MethodHandlers<
              xyz.block.ftl.v1.PullSchemaRequest,
              xyz.block.ftl.v1.PullSchemaResponse>(
                service, METHODID_PULL_SCHEMA)))
        .addMethod(
          getUpdateDeploymentRuntimeMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              xyz.block.ftl.v1.UpdateDeploymentRuntimeRequest,
              xyz.block.ftl.v1.UpdateDeploymentRuntimeResponse>(
                service, METHODID_UPDATE_DEPLOYMENT_RUNTIME)))
        .build();
  }

  private static abstract class SchemaServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoFileDescriptorSupplier, io.grpc.protobuf.ProtoServiceDescriptorSupplier {
    SchemaServiceBaseDescriptorSupplier() {}

    @java.lang.Override
    public com.google.protobuf.Descriptors.FileDescriptor getFileDescriptor() {
      return xyz.block.ftl.v1.Schemaservice.getDescriptor();
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.ServiceDescriptor getServiceDescriptor() {
      return getFileDescriptor().findServiceByName("SchemaService");
    }
  }

  private static final class SchemaServiceFileDescriptorSupplier
      extends SchemaServiceBaseDescriptorSupplier {
    SchemaServiceFileDescriptorSupplier() {}
  }

  private static final class SchemaServiceMethodDescriptorSupplier
      extends SchemaServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoMethodDescriptorSupplier {
    private final java.lang.String methodName;

    SchemaServiceMethodDescriptorSupplier(java.lang.String methodName) {
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
      synchronized (SchemaServiceGrpc.class) {
        result = serviceDescriptor;
        if (result == null) {
          serviceDescriptor = result = io.grpc.ServiceDescriptor.newBuilder(SERVICE_NAME)
              .setSchemaDescriptor(new SchemaServiceFileDescriptorSupplier())
              .addMethod(getPingMethod())
              .addMethod(getGetSchemaMethod())
              .addMethod(getPullSchemaMethod())
              .addMethod(getUpdateDeploymentRuntimeMethod())
              .build();
        }
      }
    }
    return result;
  }
}
