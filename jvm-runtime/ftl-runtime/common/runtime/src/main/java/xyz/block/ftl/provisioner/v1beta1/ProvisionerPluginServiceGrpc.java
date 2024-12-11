package xyz.block.ftl.provisioner.v1beta1;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 */
@javax.annotation.Generated(
    value = "by gRPC proto compiler (version 1.68.2)",
    comments = "Source: xyz/block/ftl/provisioner/v1beta1/plugin.proto")
@io.grpc.stub.annotations.GrpcGenerated
public final class ProvisionerPluginServiceGrpc {

  private ProvisionerPluginServiceGrpc() {}

  public static final java.lang.String SERVICE_NAME = "xyz.block.ftl.provisioner.v1beta1.ProvisionerPluginService";

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
    if ((getPingMethod = ProvisionerPluginServiceGrpc.getPingMethod) == null) {
      synchronized (ProvisionerPluginServiceGrpc.class) {
        if ((getPingMethod = ProvisionerPluginServiceGrpc.getPingMethod) == null) {
          ProvisionerPluginServiceGrpc.getPingMethod = getPingMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.v1.PingRequest, xyz.block.ftl.v1.PingResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "Ping"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.PingRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.PingResponse.getDefaultInstance()))
              .setSchemaDescriptor(new ProvisionerPluginServiceMethodDescriptorSupplier("Ping"))
              .build();
        }
      }
    }
    return getPingMethod;
  }

  private static volatile io.grpc.MethodDescriptor<xyz.block.ftl.provisioner.v1beta1.ProvisionRequest,
      xyz.block.ftl.provisioner.v1beta1.ProvisionResponse> getProvisionMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "Provision",
      requestType = xyz.block.ftl.provisioner.v1beta1.ProvisionRequest.class,
      responseType = xyz.block.ftl.provisioner.v1beta1.ProvisionResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<xyz.block.ftl.provisioner.v1beta1.ProvisionRequest,
      xyz.block.ftl.provisioner.v1beta1.ProvisionResponse> getProvisionMethod() {
    io.grpc.MethodDescriptor<xyz.block.ftl.provisioner.v1beta1.ProvisionRequest, xyz.block.ftl.provisioner.v1beta1.ProvisionResponse> getProvisionMethod;
    if ((getProvisionMethod = ProvisionerPluginServiceGrpc.getProvisionMethod) == null) {
      synchronized (ProvisionerPluginServiceGrpc.class) {
        if ((getProvisionMethod = ProvisionerPluginServiceGrpc.getProvisionMethod) == null) {
          ProvisionerPluginServiceGrpc.getProvisionMethod = getProvisionMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.provisioner.v1beta1.ProvisionRequest, xyz.block.ftl.provisioner.v1beta1.ProvisionResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "Provision"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.provisioner.v1beta1.ProvisionRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.provisioner.v1beta1.ProvisionResponse.getDefaultInstance()))
              .setSchemaDescriptor(new ProvisionerPluginServiceMethodDescriptorSupplier("Provision"))
              .build();
        }
      }
    }
    return getProvisionMethod;
  }

  private static volatile io.grpc.MethodDescriptor<xyz.block.ftl.provisioner.v1beta1.StatusRequest,
      xyz.block.ftl.provisioner.v1beta1.StatusResponse> getStatusMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "Status",
      requestType = xyz.block.ftl.provisioner.v1beta1.StatusRequest.class,
      responseType = xyz.block.ftl.provisioner.v1beta1.StatusResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<xyz.block.ftl.provisioner.v1beta1.StatusRequest,
      xyz.block.ftl.provisioner.v1beta1.StatusResponse> getStatusMethod() {
    io.grpc.MethodDescriptor<xyz.block.ftl.provisioner.v1beta1.StatusRequest, xyz.block.ftl.provisioner.v1beta1.StatusResponse> getStatusMethod;
    if ((getStatusMethod = ProvisionerPluginServiceGrpc.getStatusMethod) == null) {
      synchronized (ProvisionerPluginServiceGrpc.class) {
        if ((getStatusMethod = ProvisionerPluginServiceGrpc.getStatusMethod) == null) {
          ProvisionerPluginServiceGrpc.getStatusMethod = getStatusMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.provisioner.v1beta1.StatusRequest, xyz.block.ftl.provisioner.v1beta1.StatusResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "Status"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.provisioner.v1beta1.StatusRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.provisioner.v1beta1.StatusResponse.getDefaultInstance()))
              .setSchemaDescriptor(new ProvisionerPluginServiceMethodDescriptorSupplier("Status"))
              .build();
        }
      }
    }
    return getStatusMethod;
  }

  /**
   * Creates a new async stub that supports all call types for the service
   */
  public static ProvisionerPluginServiceStub newStub(io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<ProvisionerPluginServiceStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<ProvisionerPluginServiceStub>() {
        @java.lang.Override
        public ProvisionerPluginServiceStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new ProvisionerPluginServiceStub(channel, callOptions);
        }
      };
    return ProvisionerPluginServiceStub.newStub(factory, channel);
  }

  /**
   * Creates a new blocking-style stub that supports unary and streaming output calls on the service
   */
  public static ProvisionerPluginServiceBlockingStub newBlockingStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<ProvisionerPluginServiceBlockingStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<ProvisionerPluginServiceBlockingStub>() {
        @java.lang.Override
        public ProvisionerPluginServiceBlockingStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new ProvisionerPluginServiceBlockingStub(channel, callOptions);
        }
      };
    return ProvisionerPluginServiceBlockingStub.newStub(factory, channel);
  }

  /**
   * Creates a new ListenableFuture-style stub that supports unary calls on the service
   */
  public static ProvisionerPluginServiceFutureStub newFutureStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<ProvisionerPluginServiceFutureStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<ProvisionerPluginServiceFutureStub>() {
        @java.lang.Override
        public ProvisionerPluginServiceFutureStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new ProvisionerPluginServiceFutureStub(channel, callOptions);
        }
      };
    return ProvisionerPluginServiceFutureStub.newStub(factory, channel);
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
    default void provision(xyz.block.ftl.provisioner.v1beta1.ProvisionRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.provisioner.v1beta1.ProvisionResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getProvisionMethod(), responseObserver);
    }

    /**
     */
    default void status(xyz.block.ftl.provisioner.v1beta1.StatusRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.provisioner.v1beta1.StatusResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getStatusMethod(), responseObserver);
    }
  }

  /**
   * Base class for the server implementation of the service ProvisionerPluginService.
   */
  public static abstract class ProvisionerPluginServiceImplBase
      implements io.grpc.BindableService, AsyncService {

    @java.lang.Override public final io.grpc.ServerServiceDefinition bindService() {
      return ProvisionerPluginServiceGrpc.bindService(this);
    }
  }

  /**
   * A stub to allow clients to do asynchronous rpc calls to service ProvisionerPluginService.
   */
  public static final class ProvisionerPluginServiceStub
      extends io.grpc.stub.AbstractAsyncStub<ProvisionerPluginServiceStub> {
    private ProvisionerPluginServiceStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected ProvisionerPluginServiceStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new ProvisionerPluginServiceStub(channel, callOptions);
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
    public void provision(xyz.block.ftl.provisioner.v1beta1.ProvisionRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.provisioner.v1beta1.ProvisionResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getProvisionMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void status(xyz.block.ftl.provisioner.v1beta1.StatusRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.provisioner.v1beta1.StatusResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getStatusMethod(), getCallOptions()), request, responseObserver);
    }
  }

  /**
   * A stub to allow clients to do synchronous rpc calls to service ProvisionerPluginService.
   */
  public static final class ProvisionerPluginServiceBlockingStub
      extends io.grpc.stub.AbstractBlockingStub<ProvisionerPluginServiceBlockingStub> {
    private ProvisionerPluginServiceBlockingStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected ProvisionerPluginServiceBlockingStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new ProvisionerPluginServiceBlockingStub(channel, callOptions);
    }

    /**
     */
    public xyz.block.ftl.v1.PingResponse ping(xyz.block.ftl.v1.PingRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getPingMethod(), getCallOptions(), request);
    }

    /**
     */
    public xyz.block.ftl.provisioner.v1beta1.ProvisionResponse provision(xyz.block.ftl.provisioner.v1beta1.ProvisionRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getProvisionMethod(), getCallOptions(), request);
    }

    /**
     */
    public xyz.block.ftl.provisioner.v1beta1.StatusResponse status(xyz.block.ftl.provisioner.v1beta1.StatusRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getStatusMethod(), getCallOptions(), request);
    }
  }

  /**
   * A stub to allow clients to do ListenableFuture-style rpc calls to service ProvisionerPluginService.
   */
  public static final class ProvisionerPluginServiceFutureStub
      extends io.grpc.stub.AbstractFutureStub<ProvisionerPluginServiceFutureStub> {
    private ProvisionerPluginServiceFutureStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected ProvisionerPluginServiceFutureStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new ProvisionerPluginServiceFutureStub(channel, callOptions);
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
    public com.google.common.util.concurrent.ListenableFuture<xyz.block.ftl.provisioner.v1beta1.ProvisionResponse> provision(
        xyz.block.ftl.provisioner.v1beta1.ProvisionRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getProvisionMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<xyz.block.ftl.provisioner.v1beta1.StatusResponse> status(
        xyz.block.ftl.provisioner.v1beta1.StatusRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getStatusMethod(), getCallOptions()), request);
    }
  }

  private static final int METHODID_PING = 0;
  private static final int METHODID_PROVISION = 1;
  private static final int METHODID_STATUS = 2;

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
        case METHODID_PROVISION:
          serviceImpl.provision((xyz.block.ftl.provisioner.v1beta1.ProvisionRequest) request,
              (io.grpc.stub.StreamObserver<xyz.block.ftl.provisioner.v1beta1.ProvisionResponse>) responseObserver);
          break;
        case METHODID_STATUS:
          serviceImpl.status((xyz.block.ftl.provisioner.v1beta1.StatusRequest) request,
              (io.grpc.stub.StreamObserver<xyz.block.ftl.provisioner.v1beta1.StatusResponse>) responseObserver);
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
          getProvisionMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              xyz.block.ftl.provisioner.v1beta1.ProvisionRequest,
              xyz.block.ftl.provisioner.v1beta1.ProvisionResponse>(
                service, METHODID_PROVISION)))
        .addMethod(
          getStatusMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              xyz.block.ftl.provisioner.v1beta1.StatusRequest,
              xyz.block.ftl.provisioner.v1beta1.StatusResponse>(
                service, METHODID_STATUS)))
        .build();
  }

  private static abstract class ProvisionerPluginServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoFileDescriptorSupplier, io.grpc.protobuf.ProtoServiceDescriptorSupplier {
    ProvisionerPluginServiceBaseDescriptorSupplier() {}

    @java.lang.Override
    public com.google.protobuf.Descriptors.FileDescriptor getFileDescriptor() {
      return xyz.block.ftl.provisioner.v1beta1.Plugin.getDescriptor();
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.ServiceDescriptor getServiceDescriptor() {
      return getFileDescriptor().findServiceByName("ProvisionerPluginService");
    }
  }

  private static final class ProvisionerPluginServiceFileDescriptorSupplier
      extends ProvisionerPluginServiceBaseDescriptorSupplier {
    ProvisionerPluginServiceFileDescriptorSupplier() {}
  }

  private static final class ProvisionerPluginServiceMethodDescriptorSupplier
      extends ProvisionerPluginServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoMethodDescriptorSupplier {
    private final java.lang.String methodName;

    ProvisionerPluginServiceMethodDescriptorSupplier(java.lang.String methodName) {
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
      synchronized (ProvisionerPluginServiceGrpc.class) {
        result = serviceDescriptor;
        if (result == null) {
          serviceDescriptor = result = io.grpc.ServiceDescriptor.newBuilder(SERVICE_NAME)
              .setSchemaDescriptor(new ProvisionerPluginServiceFileDescriptorSupplier())
              .addMethod(getPingMethod())
              .addMethod(getProvisionMethod())
              .addMethod(getStatusMethod())
              .build();
        }
      }
    }
    return result;
  }
}
