package xyz.block.ftl.pubsub.v1;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 * <pre>
 * LegacyPubsubService is the service that modules use to interact with the Controller.
 * </pre>
 */
@javax.annotation.Generated(
    value = "by gRPC proto compiler (version 1.68.2)",
    comments = "Source: xyz/block/ftl/pubsub/v1/pubsub.proto")
@io.grpc.stub.annotations.GrpcGenerated
public final class LegacyPubsubServiceGrpc {

  private LegacyPubsubServiceGrpc() {}

  public static final java.lang.String SERVICE_NAME = "xyz.block.ftl.pubsub.v1.LegacyPubsubService";

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
    if ((getPingMethod = LegacyPubsubServiceGrpc.getPingMethod) == null) {
      synchronized (LegacyPubsubServiceGrpc.class) {
        if ((getPingMethod = LegacyPubsubServiceGrpc.getPingMethod) == null) {
          LegacyPubsubServiceGrpc.getPingMethod = getPingMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.v1.PingRequest, xyz.block.ftl.v1.PingResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "Ping"))
              .setSafe(true)
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.PingRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.PingResponse.getDefaultInstance()))
              .setSchemaDescriptor(new LegacyPubsubServiceMethodDescriptorSupplier("Ping"))
              .build();
        }
      }
    }
    return getPingMethod;
  }

  private static volatile io.grpc.MethodDescriptor<xyz.block.ftl.pubsub.v1.PublishEventRequest,
      xyz.block.ftl.pubsub.v1.PublishEventResponse> getPublishEventMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "PublishEvent",
      requestType = xyz.block.ftl.pubsub.v1.PublishEventRequest.class,
      responseType = xyz.block.ftl.pubsub.v1.PublishEventResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<xyz.block.ftl.pubsub.v1.PublishEventRequest,
      xyz.block.ftl.pubsub.v1.PublishEventResponse> getPublishEventMethod() {
    io.grpc.MethodDescriptor<xyz.block.ftl.pubsub.v1.PublishEventRequest, xyz.block.ftl.pubsub.v1.PublishEventResponse> getPublishEventMethod;
    if ((getPublishEventMethod = LegacyPubsubServiceGrpc.getPublishEventMethod) == null) {
      synchronized (LegacyPubsubServiceGrpc.class) {
        if ((getPublishEventMethod = LegacyPubsubServiceGrpc.getPublishEventMethod) == null) {
          LegacyPubsubServiceGrpc.getPublishEventMethod = getPublishEventMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.pubsub.v1.PublishEventRequest, xyz.block.ftl.pubsub.v1.PublishEventResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "PublishEvent"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.pubsub.v1.PublishEventRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.pubsub.v1.PublishEventResponse.getDefaultInstance()))
              .setSchemaDescriptor(new LegacyPubsubServiceMethodDescriptorSupplier("PublishEvent"))
              .build();
        }
      }
    }
    return getPublishEventMethod;
  }

  private static volatile io.grpc.MethodDescriptor<xyz.block.ftl.pubsub.v1.ResetSubscriptionRequest,
      xyz.block.ftl.pubsub.v1.ResetSubscriptionResponse> getResetSubscriptionMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "ResetSubscription",
      requestType = xyz.block.ftl.pubsub.v1.ResetSubscriptionRequest.class,
      responseType = xyz.block.ftl.pubsub.v1.ResetSubscriptionResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<xyz.block.ftl.pubsub.v1.ResetSubscriptionRequest,
      xyz.block.ftl.pubsub.v1.ResetSubscriptionResponse> getResetSubscriptionMethod() {
    io.grpc.MethodDescriptor<xyz.block.ftl.pubsub.v1.ResetSubscriptionRequest, xyz.block.ftl.pubsub.v1.ResetSubscriptionResponse> getResetSubscriptionMethod;
    if ((getResetSubscriptionMethod = LegacyPubsubServiceGrpc.getResetSubscriptionMethod) == null) {
      synchronized (LegacyPubsubServiceGrpc.class) {
        if ((getResetSubscriptionMethod = LegacyPubsubServiceGrpc.getResetSubscriptionMethod) == null) {
          LegacyPubsubServiceGrpc.getResetSubscriptionMethod = getResetSubscriptionMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.pubsub.v1.ResetSubscriptionRequest, xyz.block.ftl.pubsub.v1.ResetSubscriptionResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "ResetSubscription"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.pubsub.v1.ResetSubscriptionRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.pubsub.v1.ResetSubscriptionResponse.getDefaultInstance()))
              .setSchemaDescriptor(new LegacyPubsubServiceMethodDescriptorSupplier("ResetSubscription"))
              .build();
        }
      }
    }
    return getResetSubscriptionMethod;
  }

  /**
   * Creates a new async stub that supports all call types for the service
   */
  public static LegacyPubsubServiceStub newStub(io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<LegacyPubsubServiceStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<LegacyPubsubServiceStub>() {
        @java.lang.Override
        public LegacyPubsubServiceStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new LegacyPubsubServiceStub(channel, callOptions);
        }
      };
    return LegacyPubsubServiceStub.newStub(factory, channel);
  }

  /**
   * Creates a new blocking-style stub that supports unary and streaming output calls on the service
   */
  public static LegacyPubsubServiceBlockingStub newBlockingStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<LegacyPubsubServiceBlockingStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<LegacyPubsubServiceBlockingStub>() {
        @java.lang.Override
        public LegacyPubsubServiceBlockingStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new LegacyPubsubServiceBlockingStub(channel, callOptions);
        }
      };
    return LegacyPubsubServiceBlockingStub.newStub(factory, channel);
  }

  /**
   * Creates a new ListenableFuture-style stub that supports unary calls on the service
   */
  public static LegacyPubsubServiceFutureStub newFutureStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<LegacyPubsubServiceFutureStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<LegacyPubsubServiceFutureStub>() {
        @java.lang.Override
        public LegacyPubsubServiceFutureStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new LegacyPubsubServiceFutureStub(channel, callOptions);
        }
      };
    return LegacyPubsubServiceFutureStub.newStub(factory, channel);
  }

  /**
   * <pre>
   * LegacyPubsubService is the service that modules use to interact with the Controller.
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
     * Publish an event to a topic.
     * </pre>
     */
    default void publishEvent(xyz.block.ftl.pubsub.v1.PublishEventRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.pubsub.v1.PublishEventResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getPublishEventMethod(), responseObserver);
    }

    /**
     * <pre>
     * Reset the cursor for a subscription to the head of its topic.
     * </pre>
     */
    default void resetSubscription(xyz.block.ftl.pubsub.v1.ResetSubscriptionRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.pubsub.v1.ResetSubscriptionResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getResetSubscriptionMethod(), responseObserver);
    }
  }

  /**
   * Base class for the server implementation of the service LegacyPubsubService.
   * <pre>
   * LegacyPubsubService is the service that modules use to interact with the Controller.
   * </pre>
   */
  public static abstract class LegacyPubsubServiceImplBase
      implements io.grpc.BindableService, AsyncService {

    @java.lang.Override public final io.grpc.ServerServiceDefinition bindService() {
      return LegacyPubsubServiceGrpc.bindService(this);
    }
  }

  /**
   * A stub to allow clients to do asynchronous rpc calls to service LegacyPubsubService.
   * <pre>
   * LegacyPubsubService is the service that modules use to interact with the Controller.
   * </pre>
   */
  public static final class LegacyPubsubServiceStub
      extends io.grpc.stub.AbstractAsyncStub<LegacyPubsubServiceStub> {
    private LegacyPubsubServiceStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected LegacyPubsubServiceStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new LegacyPubsubServiceStub(channel, callOptions);
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
     * Publish an event to a topic.
     * </pre>
     */
    public void publishEvent(xyz.block.ftl.pubsub.v1.PublishEventRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.pubsub.v1.PublishEventResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getPublishEventMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * Reset the cursor for a subscription to the head of its topic.
     * </pre>
     */
    public void resetSubscription(xyz.block.ftl.pubsub.v1.ResetSubscriptionRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.pubsub.v1.ResetSubscriptionResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getResetSubscriptionMethod(), getCallOptions()), request, responseObserver);
    }
  }

  /**
   * A stub to allow clients to do synchronous rpc calls to service LegacyPubsubService.
   * <pre>
   * LegacyPubsubService is the service that modules use to interact with the Controller.
   * </pre>
   */
  public static final class LegacyPubsubServiceBlockingStub
      extends io.grpc.stub.AbstractBlockingStub<LegacyPubsubServiceBlockingStub> {
    private LegacyPubsubServiceBlockingStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected LegacyPubsubServiceBlockingStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new LegacyPubsubServiceBlockingStub(channel, callOptions);
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
     * Publish an event to a topic.
     * </pre>
     */
    public xyz.block.ftl.pubsub.v1.PublishEventResponse publishEvent(xyz.block.ftl.pubsub.v1.PublishEventRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getPublishEventMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Reset the cursor for a subscription to the head of its topic.
     * </pre>
     */
    public xyz.block.ftl.pubsub.v1.ResetSubscriptionResponse resetSubscription(xyz.block.ftl.pubsub.v1.ResetSubscriptionRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getResetSubscriptionMethod(), getCallOptions(), request);
    }
  }

  /**
   * A stub to allow clients to do ListenableFuture-style rpc calls to service LegacyPubsubService.
   * <pre>
   * LegacyPubsubService is the service that modules use to interact with the Controller.
   * </pre>
   */
  public static final class LegacyPubsubServiceFutureStub
      extends io.grpc.stub.AbstractFutureStub<LegacyPubsubServiceFutureStub> {
    private LegacyPubsubServiceFutureStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected LegacyPubsubServiceFutureStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new LegacyPubsubServiceFutureStub(channel, callOptions);
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
     * Publish an event to a topic.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<xyz.block.ftl.pubsub.v1.PublishEventResponse> publishEvent(
        xyz.block.ftl.pubsub.v1.PublishEventRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getPublishEventMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * Reset the cursor for a subscription to the head of its topic.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<xyz.block.ftl.pubsub.v1.ResetSubscriptionResponse> resetSubscription(
        xyz.block.ftl.pubsub.v1.ResetSubscriptionRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getResetSubscriptionMethod(), getCallOptions()), request);
    }
  }

  private static final int METHODID_PING = 0;
  private static final int METHODID_PUBLISH_EVENT = 1;
  private static final int METHODID_RESET_SUBSCRIPTION = 2;

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
          serviceImpl.publishEvent((xyz.block.ftl.pubsub.v1.PublishEventRequest) request,
              (io.grpc.stub.StreamObserver<xyz.block.ftl.pubsub.v1.PublishEventResponse>) responseObserver);
          break;
        case METHODID_RESET_SUBSCRIPTION:
          serviceImpl.resetSubscription((xyz.block.ftl.pubsub.v1.ResetSubscriptionRequest) request,
              (io.grpc.stub.StreamObserver<xyz.block.ftl.pubsub.v1.ResetSubscriptionResponse>) responseObserver);
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
              xyz.block.ftl.pubsub.v1.PublishEventRequest,
              xyz.block.ftl.pubsub.v1.PublishEventResponse>(
                service, METHODID_PUBLISH_EVENT)))
        .addMethod(
          getResetSubscriptionMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              xyz.block.ftl.pubsub.v1.ResetSubscriptionRequest,
              xyz.block.ftl.pubsub.v1.ResetSubscriptionResponse>(
                service, METHODID_RESET_SUBSCRIPTION)))
        .build();
  }

  private static abstract class LegacyPubsubServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoFileDescriptorSupplier, io.grpc.protobuf.ProtoServiceDescriptorSupplier {
    LegacyPubsubServiceBaseDescriptorSupplier() {}

    @java.lang.Override
    public com.google.protobuf.Descriptors.FileDescriptor getFileDescriptor() {
      return xyz.block.ftl.pubsub.v1.Pubsub.getDescriptor();
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.ServiceDescriptor getServiceDescriptor() {
      return getFileDescriptor().findServiceByName("LegacyPubsubService");
    }
  }

  private static final class LegacyPubsubServiceFileDescriptorSupplier
      extends LegacyPubsubServiceBaseDescriptorSupplier {
    LegacyPubsubServiceFileDescriptorSupplier() {}
  }

  private static final class LegacyPubsubServiceMethodDescriptorSupplier
      extends LegacyPubsubServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoMethodDescriptorSupplier {
    private final java.lang.String methodName;

    LegacyPubsubServiceMethodDescriptorSupplier(java.lang.String methodName) {
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
      synchronized (LegacyPubsubServiceGrpc.class) {
        result = serviceDescriptor;
        if (result == null) {
          serviceDescriptor = result = io.grpc.ServiceDescriptor.newBuilder(SERVICE_NAME)
              .setSchemaDescriptor(new LegacyPubsubServiceFileDescriptorSupplier())
              .addMethod(getPingMethod())
              .addMethod(getPublishEventMethod())
              .addMethod(getResetSubscriptionMethod())
              .build();
        }
      }
    }
    return result;
  }
}
