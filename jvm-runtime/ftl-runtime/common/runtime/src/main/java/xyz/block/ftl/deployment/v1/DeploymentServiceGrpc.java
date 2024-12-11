package xyz.block.ftl.deployment.v1;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 * <pre>
 * ModuleService is the service that modules use to interact with the Controller.
 * </pre>
 */
@javax.annotation.Generated(
    value = "by gRPC proto compiler (version 1.68.2)",
    comments = "Source: xyz/block/ftl/deployment/v1/deployment.proto")
@io.grpc.stub.annotations.GrpcGenerated
public final class DeploymentServiceGrpc {

  private DeploymentServiceGrpc() {}

  public static final java.lang.String SERVICE_NAME = "xyz.block.ftl.deployment.v1.DeploymentService";

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
    if ((getPingMethod = DeploymentServiceGrpc.getPingMethod) == null) {
      synchronized (DeploymentServiceGrpc.class) {
        if ((getPingMethod = DeploymentServiceGrpc.getPingMethod) == null) {
          DeploymentServiceGrpc.getPingMethod = getPingMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.v1.PingRequest, xyz.block.ftl.v1.PingResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "Ping"))
              .setSafe(true)
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.PingRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.PingResponse.getDefaultInstance()))
              .setSchemaDescriptor(new DeploymentServiceMethodDescriptorSupplier("Ping"))
              .build();
        }
      }
    }
    return getPingMethod;
  }

  private static volatile io.grpc.MethodDescriptor<xyz.block.ftl.deployment.v1.GetDeploymentContextRequest,
      xyz.block.ftl.deployment.v1.GetDeploymentContextResponse> getGetDeploymentContextMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "GetDeploymentContext",
      requestType = xyz.block.ftl.deployment.v1.GetDeploymentContextRequest.class,
      responseType = xyz.block.ftl.deployment.v1.GetDeploymentContextResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.SERVER_STREAMING)
  public static io.grpc.MethodDescriptor<xyz.block.ftl.deployment.v1.GetDeploymentContextRequest,
      xyz.block.ftl.deployment.v1.GetDeploymentContextResponse> getGetDeploymentContextMethod() {
    io.grpc.MethodDescriptor<xyz.block.ftl.deployment.v1.GetDeploymentContextRequest, xyz.block.ftl.deployment.v1.GetDeploymentContextResponse> getGetDeploymentContextMethod;
    if ((getGetDeploymentContextMethod = DeploymentServiceGrpc.getGetDeploymentContextMethod) == null) {
      synchronized (DeploymentServiceGrpc.class) {
        if ((getGetDeploymentContextMethod = DeploymentServiceGrpc.getGetDeploymentContextMethod) == null) {
          DeploymentServiceGrpc.getGetDeploymentContextMethod = getGetDeploymentContextMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.deployment.v1.GetDeploymentContextRequest, xyz.block.ftl.deployment.v1.GetDeploymentContextResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.SERVER_STREAMING)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "GetDeploymentContext"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.deployment.v1.GetDeploymentContextRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.deployment.v1.GetDeploymentContextResponse.getDefaultInstance()))
              .setSchemaDescriptor(new DeploymentServiceMethodDescriptorSupplier("GetDeploymentContext"))
              .build();
        }
      }
    }
    return getGetDeploymentContextMethod;
  }

  /**
   * Creates a new async stub that supports all call types for the service
   */
  public static DeploymentServiceStub newStub(io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<DeploymentServiceStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<DeploymentServiceStub>() {
        @java.lang.Override
        public DeploymentServiceStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new DeploymentServiceStub(channel, callOptions);
        }
      };
    return DeploymentServiceStub.newStub(factory, channel);
  }

  /**
   * Creates a new blocking-style stub that supports unary and streaming output calls on the service
   */
  public static DeploymentServiceBlockingStub newBlockingStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<DeploymentServiceBlockingStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<DeploymentServiceBlockingStub>() {
        @java.lang.Override
        public DeploymentServiceBlockingStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new DeploymentServiceBlockingStub(channel, callOptions);
        }
      };
    return DeploymentServiceBlockingStub.newStub(factory, channel);
  }

  /**
   * Creates a new ListenableFuture-style stub that supports unary calls on the service
   */
  public static DeploymentServiceFutureStub newFutureStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<DeploymentServiceFutureStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<DeploymentServiceFutureStub>() {
        @java.lang.Override
        public DeploymentServiceFutureStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new DeploymentServiceFutureStub(channel, callOptions);
        }
      };
    return DeploymentServiceFutureStub.newStub(factory, channel);
  }

  /**
   * <pre>
   * ModuleService is the service that modules use to interact with the Controller.
   * </pre>
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
     * Get configuration state for the deployment
     * </pre>
     */
    default void getDeploymentContext(xyz.block.ftl.deployment.v1.GetDeploymentContextRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.deployment.v1.GetDeploymentContextResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getGetDeploymentContextMethod(), responseObserver);
    }
  }

  /**
   * Base class for the server implementation of the service DeploymentService.
   * <pre>
   * ModuleService is the service that modules use to interact with the Controller.
   * </pre>
   */
  public static abstract class DeploymentServiceImplBase
      implements io.grpc.BindableService, AsyncService {

    @java.lang.Override public final io.grpc.ServerServiceDefinition bindService() {
      return DeploymentServiceGrpc.bindService(this);
    }
  }

  /**
   * A stub to allow clients to do asynchronous rpc calls to service DeploymentService.
   * <pre>
   * ModuleService is the service that modules use to interact with the Controller.
   * </pre>
   */
  public static final class DeploymentServiceStub
      extends io.grpc.stub.AbstractAsyncStub<DeploymentServiceStub> {
    private DeploymentServiceStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected DeploymentServiceStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new DeploymentServiceStub(channel, callOptions);
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
     * Get configuration state for the deployment
     * </pre>
     */
    public void getDeploymentContext(xyz.block.ftl.deployment.v1.GetDeploymentContextRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.deployment.v1.GetDeploymentContextResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncServerStreamingCall(
          getChannel().newCall(getGetDeploymentContextMethod(), getCallOptions()), request, responseObserver);
    }
  }

  /**
   * A stub to allow clients to do synchronous rpc calls to service DeploymentService.
   * <pre>
   * ModuleService is the service that modules use to interact with the Controller.
   * </pre>
   */
  public static final class DeploymentServiceBlockingStub
      extends io.grpc.stub.AbstractBlockingStub<DeploymentServiceBlockingStub> {
    private DeploymentServiceBlockingStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected DeploymentServiceBlockingStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new DeploymentServiceBlockingStub(channel, callOptions);
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
     * Get configuration state for the deployment
     * </pre>
     */
    public java.util.Iterator<xyz.block.ftl.deployment.v1.GetDeploymentContextResponse> getDeploymentContext(
        xyz.block.ftl.deployment.v1.GetDeploymentContextRequest request) {
      return io.grpc.stub.ClientCalls.blockingServerStreamingCall(
          getChannel(), getGetDeploymentContextMethod(), getCallOptions(), request);
    }
  }

  /**
   * A stub to allow clients to do ListenableFuture-style rpc calls to service DeploymentService.
   * <pre>
   * ModuleService is the service that modules use to interact with the Controller.
   * </pre>
   */
  public static final class DeploymentServiceFutureStub
      extends io.grpc.stub.AbstractFutureStub<DeploymentServiceFutureStub> {
    private DeploymentServiceFutureStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected DeploymentServiceFutureStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new DeploymentServiceFutureStub(channel, callOptions);
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
  }

  private static final int METHODID_PING = 0;
  private static final int METHODID_GET_DEPLOYMENT_CONTEXT = 1;

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
        case METHODID_GET_DEPLOYMENT_CONTEXT:
          serviceImpl.getDeploymentContext((xyz.block.ftl.deployment.v1.GetDeploymentContextRequest) request,
              (io.grpc.stub.StreamObserver<xyz.block.ftl.deployment.v1.GetDeploymentContextResponse>) responseObserver);
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
          getGetDeploymentContextMethod(),
          io.grpc.stub.ServerCalls.asyncServerStreamingCall(
            new MethodHandlers<
              xyz.block.ftl.deployment.v1.GetDeploymentContextRequest,
              xyz.block.ftl.deployment.v1.GetDeploymentContextResponse>(
                service, METHODID_GET_DEPLOYMENT_CONTEXT)))
        .build();
  }

  private static abstract class DeploymentServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoFileDescriptorSupplier, io.grpc.protobuf.ProtoServiceDescriptorSupplier {
    DeploymentServiceBaseDescriptorSupplier() {}

    @java.lang.Override
    public com.google.protobuf.Descriptors.FileDescriptor getFileDescriptor() {
      return xyz.block.ftl.deployment.v1.Deployment.getDescriptor();
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.ServiceDescriptor getServiceDescriptor() {
      return getFileDescriptor().findServiceByName("DeploymentService");
    }
  }

  private static final class DeploymentServiceFileDescriptorSupplier
      extends DeploymentServiceBaseDescriptorSupplier {
    DeploymentServiceFileDescriptorSupplier() {}
  }

  private static final class DeploymentServiceMethodDescriptorSupplier
      extends DeploymentServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoMethodDescriptorSupplier {
    private final java.lang.String methodName;

    DeploymentServiceMethodDescriptorSupplier(java.lang.String methodName) {
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
      synchronized (DeploymentServiceGrpc.class) {
        result = serviceDescriptor;
        if (result == null) {
          serviceDescriptor = result = io.grpc.ServiceDescriptor.newBuilder(SERVICE_NAME)
              .setSchemaDescriptor(new DeploymentServiceFileDescriptorSupplier())
              .addMethod(getPingMethod())
              .addMethod(getGetDeploymentContextMethod())
              .build();
        }
      }
    }
    return result;
  }
}
