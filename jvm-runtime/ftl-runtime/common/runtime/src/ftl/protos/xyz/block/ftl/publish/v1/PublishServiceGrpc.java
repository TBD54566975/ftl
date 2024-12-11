package xyz.block.ftl.publish.v1;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 */
@javax.annotation.Generated(
    value = "by gRPC proto compiler (version 1.68.2)",
    comments = "Source: xyz/block/ftl/publish/v1/publish.proto")
@io.grpc.stub.annotations.GrpcGenerated
public final class PublishServiceGrpc {

  private PublishServiceGrpc() {}

  public static final java.lang.String SERVICE_NAME = "xyz.block.ftl.publish.v1.PublishService";

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
    if ((getPingMethod = PublishServiceGrpc.getPingMethod) == null) {
      synchronized (PublishServiceGrpc.class) {
        if ((getPingMethod = PublishServiceGrpc.getPingMethod) == null) {
          PublishServiceGrpc.getPingMethod = getPingMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.v1.PingRequest, xyz.block.ftl.v1.PingResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "Ping"))
              .setSafe(true)
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.PingRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.PingResponse.getDefaultInstance()))
              .setSchemaDescriptor(new PublishServiceMethodDescriptorSupplier("Ping"))
              .build();
        }
      }
    }
    return getPingMethod;
  }

  private static volatile io.grpc.MethodDescriptor<xyz.block.ftl.publish.v1.PublishEventRequest,
      xyz.block.ftl.publish.v1.PublishEventResponse> getPublishEventMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "PublishEvent",
      requestType = xyz.block.ftl.publish.v1.PublishEventRequest.class,
      responseType = xyz.block.ftl.publish.v1.PublishEventResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<xyz.block.ftl.publish.v1.PublishEventRequest,
      xyz.block.ftl.publish.v1.PublishEventResponse> getPublishEventMethod() {
    io.grpc.MethodDescriptor<xyz.block.ftl.publish.v1.PublishEventRequest, xyz.block.ftl.publish.v1.PublishEventResponse> getPublishEventMethod;
    if ((getPublishEventMethod = PublishServiceGrpc.getPublishEventMethod) == null) {
      synchronized (PublishServiceGrpc.class) {
        if ((getPublishEventMethod = PublishServiceGrpc.getPublishEventMethod) == null) {
          PublishServiceGrpc.getPublishEventMethod = getPublishEventMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.publish.v1.PublishEventRequest, xyz.block.ftl.publish.v1.PublishEventResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "PublishEvent"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.publish.v1.PublishEventRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.publish.v1.PublishEventResponse.getDefaultInstance()))
              .setSchemaDescriptor(new PublishServiceMethodDescriptorSupplier("PublishEvent"))
              .build();
        }
      }
    }
    return getPublishEventMethod;
  }

  /**
   * Creates a new async stub that supports all call types for the service
   */
  public static PublishServiceStub newStub(io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<PublishServiceStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<PublishServiceStub>() {
        @java.lang.Override
        public PublishServiceStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new PublishServiceStub(channel, callOptions);
        }
      };
    return PublishServiceStub.newStub(factory, channel);
  }

  /**
   * Creates a new blocking-style stub that supports unary and streaming output calls on the service
   */
  public static PublishServiceBlockingStub newBlockingStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<PublishServiceBlockingStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<PublishServiceBlockingStub>() {
        @java.lang.Override
        public PublishServiceBlockingStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new PublishServiceBlockingStub(channel, callOptions);
        }
      };
    return PublishServiceBlockingStub.newStub(factory, channel);
  }

  /**
   * Creates a new ListenableFuture-style stub that supports unary calls on the service
   */
  public static PublishServiceFutureStub newFutureStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<PublishServiceFutureStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<PublishServiceFutureStub>() {
        @java.lang.Override
        public PublishServiceFutureStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new PublishServiceFutureStub(channel, callOptions);
        }
      };
    return PublishServiceFutureStub.newStub(factory, channel);
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
     * Publish a message to a topic.
     * </pre>
     */
    default void publishEvent(xyz.block.ftl.publish.v1.PublishEventRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.publish.v1.PublishEventResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getPublishEventMethod(), responseObserver);
    }
  }

  /**
   * Base class for the server implementation of the service PublishService.
   */
  public static abstract class PublishServiceImplBase
      implements io.grpc.BindableService, AsyncService {

    @java.lang.Override public final io.grpc.ServerServiceDefinition bindService() {
      return PublishServiceGrpc.bindService(this);
    }
  }

  /**
   * A stub to allow clients to do asynchronous rpc calls to service PublishService.
   */
  public static final class PublishServiceStub
      extends io.grpc.stub.AbstractAsyncStub<PublishServiceStub> {
    private PublishServiceStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected PublishServiceStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new PublishServiceStub(channel, callOptions);
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
     * Publish a message to a topic.
     * </pre>
     */
    public void publishEvent(xyz.block.ftl.publish.v1.PublishEventRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.publish.v1.PublishEventResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getPublishEventMethod(), getCallOptions()), request, responseObserver);
    }
  }

  /**
   * A stub to allow clients to do synchronous rpc calls to service PublishService.
   */
  public static final class PublishServiceBlockingStub
      extends io.grpc.stub.AbstractBlockingStub<PublishServiceBlockingStub> {
    private PublishServiceBlockingStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected PublishServiceBlockingStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new PublishServiceBlockingStub(channel, callOptions);
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
     * Publish a message to a topic.
     * </pre>
     */
    public xyz.block.ftl.publish.v1.PublishEventResponse publishEvent(xyz.block.ftl.publish.v1.PublishEventRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getPublishEventMethod(), getCallOptions(), request);
    }
  }

  /**
   * A stub to allow clients to do ListenableFuture-style rpc calls to service PublishService.
   */
  public static final class PublishServiceFutureStub
      extends io.grpc.stub.AbstractFutureStub<PublishServiceFutureStub> {
    private PublishServiceFutureStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected PublishServiceFutureStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new PublishServiceFutureStub(channel, callOptions);
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
     * Publish a message to a topic.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<xyz.block.ftl.publish.v1.PublishEventResponse> publishEvent(
        xyz.block.ftl.publish.v1.PublishEventRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getPublishEventMethod(), getCallOptions()), request);
    }
  }

  private static final int METHODID_PING = 0;
  private static final int METHODID_PUBLISH_EVENT = 1;

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
        case METHODID_PUBLISH_EVENT:
          serviceImpl.publishEvent((xyz.block.ftl.publish.v1.PublishEventRequest) request,
              (io.grpc.stub.StreamObserver<xyz.block.ftl.publish.v1.PublishEventResponse>) responseObserver);
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
          getPublishEventMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              xyz.block.ftl.publish.v1.PublishEventRequest,
              xyz.block.ftl.publish.v1.PublishEventResponse>(
                service, METHODID_PUBLISH_EVENT)))
        .build();
  }

  private static abstract class PublishServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoFileDescriptorSupplier, io.grpc.protobuf.ProtoServiceDescriptorSupplier {
    PublishServiceBaseDescriptorSupplier() {}

    @java.lang.Override
    public com.google.protobuf.Descriptors.FileDescriptor getFileDescriptor() {
      return xyz.block.ftl.publish.v1.Publish.getDescriptor();
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.ServiceDescriptor getServiceDescriptor() {
      return getFileDescriptor().findServiceByName("PublishService");
    }
  }

  private static final class PublishServiceFileDescriptorSupplier
      extends PublishServiceBaseDescriptorSupplier {
    PublishServiceFileDescriptorSupplier() {}
  }

  private static final class PublishServiceMethodDescriptorSupplier
      extends PublishServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoMethodDescriptorSupplier {
    private final java.lang.String methodName;

    PublishServiceMethodDescriptorSupplier(java.lang.String methodName) {
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
      synchronized (PublishServiceGrpc.class) {
        result = serviceDescriptor;
        if (result == null) {
          serviceDescriptor = result = io.grpc.ServiceDescriptor.newBuilder(SERVICE_NAME)
              .setSchemaDescriptor(new PublishServiceFileDescriptorSupplier())
              .addMethod(getPingMethod())
              .addMethod(getPublishEventMethod())
              .build();
        }
      }
    }
    return result;
  }
}
