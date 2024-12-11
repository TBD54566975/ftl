package xyz.block.ftl.timeline.v1;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 */
@javax.annotation.Generated(
    value = "by gRPC proto compiler (version 1.68.2)",
    comments = "Source: xyz/block/ftl/timeline/v1/timeline.proto")
@io.grpc.stub.annotations.GrpcGenerated
public final class TimelineServiceGrpc {

  private TimelineServiceGrpc() {}

  public static final java.lang.String SERVICE_NAME = "xyz.block.ftl.timeline.v1.TimelineService";

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
    if ((getPingMethod = TimelineServiceGrpc.getPingMethod) == null) {
      synchronized (TimelineServiceGrpc.class) {
        if ((getPingMethod = TimelineServiceGrpc.getPingMethod) == null) {
          TimelineServiceGrpc.getPingMethod = getPingMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.v1.PingRequest, xyz.block.ftl.v1.PingResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "Ping"))
              .setSafe(true)
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.PingRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.PingResponse.getDefaultInstance()))
              .setSchemaDescriptor(new TimelineServiceMethodDescriptorSupplier("Ping"))
              .build();
        }
      }
    }
    return getPingMethod;
  }

  private static volatile io.grpc.MethodDescriptor<xyz.block.ftl.timeline.v1.GetTimelineRequest,
      xyz.block.ftl.timeline.v1.GetTimelineResponse> getGetTimelineMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "GetTimeline",
      requestType = xyz.block.ftl.timeline.v1.GetTimelineRequest.class,
      responseType = xyz.block.ftl.timeline.v1.GetTimelineResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<xyz.block.ftl.timeline.v1.GetTimelineRequest,
      xyz.block.ftl.timeline.v1.GetTimelineResponse> getGetTimelineMethod() {
    io.grpc.MethodDescriptor<xyz.block.ftl.timeline.v1.GetTimelineRequest, xyz.block.ftl.timeline.v1.GetTimelineResponse> getGetTimelineMethod;
    if ((getGetTimelineMethod = TimelineServiceGrpc.getGetTimelineMethod) == null) {
      synchronized (TimelineServiceGrpc.class) {
        if ((getGetTimelineMethod = TimelineServiceGrpc.getGetTimelineMethod) == null) {
          TimelineServiceGrpc.getGetTimelineMethod = getGetTimelineMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.timeline.v1.GetTimelineRequest, xyz.block.ftl.timeline.v1.GetTimelineResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "GetTimeline"))
              .setSafe(true)
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.timeline.v1.GetTimelineRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.timeline.v1.GetTimelineResponse.getDefaultInstance()))
              .setSchemaDescriptor(new TimelineServiceMethodDescriptorSupplier("GetTimeline"))
              .build();
        }
      }
    }
    return getGetTimelineMethod;
  }

  private static volatile io.grpc.MethodDescriptor<xyz.block.ftl.timeline.v1.StreamTimelineRequest,
      xyz.block.ftl.timeline.v1.StreamTimelineResponse> getStreamTimelineMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "StreamTimeline",
      requestType = xyz.block.ftl.timeline.v1.StreamTimelineRequest.class,
      responseType = xyz.block.ftl.timeline.v1.StreamTimelineResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.SERVER_STREAMING)
  public static io.grpc.MethodDescriptor<xyz.block.ftl.timeline.v1.StreamTimelineRequest,
      xyz.block.ftl.timeline.v1.StreamTimelineResponse> getStreamTimelineMethod() {
    io.grpc.MethodDescriptor<xyz.block.ftl.timeline.v1.StreamTimelineRequest, xyz.block.ftl.timeline.v1.StreamTimelineResponse> getStreamTimelineMethod;
    if ((getStreamTimelineMethod = TimelineServiceGrpc.getStreamTimelineMethod) == null) {
      synchronized (TimelineServiceGrpc.class) {
        if ((getStreamTimelineMethod = TimelineServiceGrpc.getStreamTimelineMethod) == null) {
          TimelineServiceGrpc.getStreamTimelineMethod = getStreamTimelineMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.timeline.v1.StreamTimelineRequest, xyz.block.ftl.timeline.v1.StreamTimelineResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.SERVER_STREAMING)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "StreamTimeline"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.timeline.v1.StreamTimelineRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.timeline.v1.StreamTimelineResponse.getDefaultInstance()))
              .setSchemaDescriptor(new TimelineServiceMethodDescriptorSupplier("StreamTimeline"))
              .build();
        }
      }
    }
    return getStreamTimelineMethod;
  }

  private static volatile io.grpc.MethodDescriptor<xyz.block.ftl.timeline.v1.CreateEventsRequest,
      xyz.block.ftl.timeline.v1.CreateEventsResponse> getCreateEventsMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "CreateEvents",
      requestType = xyz.block.ftl.timeline.v1.CreateEventsRequest.class,
      responseType = xyz.block.ftl.timeline.v1.CreateEventsResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<xyz.block.ftl.timeline.v1.CreateEventsRequest,
      xyz.block.ftl.timeline.v1.CreateEventsResponse> getCreateEventsMethod() {
    io.grpc.MethodDescriptor<xyz.block.ftl.timeline.v1.CreateEventsRequest, xyz.block.ftl.timeline.v1.CreateEventsResponse> getCreateEventsMethod;
    if ((getCreateEventsMethod = TimelineServiceGrpc.getCreateEventsMethod) == null) {
      synchronized (TimelineServiceGrpc.class) {
        if ((getCreateEventsMethod = TimelineServiceGrpc.getCreateEventsMethod) == null) {
          TimelineServiceGrpc.getCreateEventsMethod = getCreateEventsMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.timeline.v1.CreateEventsRequest, xyz.block.ftl.timeline.v1.CreateEventsResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "CreateEvents"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.timeline.v1.CreateEventsRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.timeline.v1.CreateEventsResponse.getDefaultInstance()))
              .setSchemaDescriptor(new TimelineServiceMethodDescriptorSupplier("CreateEvents"))
              .build();
        }
      }
    }
    return getCreateEventsMethod;
  }

  private static volatile io.grpc.MethodDescriptor<xyz.block.ftl.timeline.v1.DeleteOldEventsRequest,
      xyz.block.ftl.timeline.v1.DeleteOldEventsResponse> getDeleteOldEventsMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "DeleteOldEvents",
      requestType = xyz.block.ftl.timeline.v1.DeleteOldEventsRequest.class,
      responseType = xyz.block.ftl.timeline.v1.DeleteOldEventsResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<xyz.block.ftl.timeline.v1.DeleteOldEventsRequest,
      xyz.block.ftl.timeline.v1.DeleteOldEventsResponse> getDeleteOldEventsMethod() {
    io.grpc.MethodDescriptor<xyz.block.ftl.timeline.v1.DeleteOldEventsRequest, xyz.block.ftl.timeline.v1.DeleteOldEventsResponse> getDeleteOldEventsMethod;
    if ((getDeleteOldEventsMethod = TimelineServiceGrpc.getDeleteOldEventsMethod) == null) {
      synchronized (TimelineServiceGrpc.class) {
        if ((getDeleteOldEventsMethod = TimelineServiceGrpc.getDeleteOldEventsMethod) == null) {
          TimelineServiceGrpc.getDeleteOldEventsMethod = getDeleteOldEventsMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.timeline.v1.DeleteOldEventsRequest, xyz.block.ftl.timeline.v1.DeleteOldEventsResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "DeleteOldEvents"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.timeline.v1.DeleteOldEventsRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.timeline.v1.DeleteOldEventsResponse.getDefaultInstance()))
              .setSchemaDescriptor(new TimelineServiceMethodDescriptorSupplier("DeleteOldEvents"))
              .build();
        }
      }
    }
    return getDeleteOldEventsMethod;
  }

  /**
   * Creates a new async stub that supports all call types for the service
   */
  public static TimelineServiceStub newStub(io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<TimelineServiceStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<TimelineServiceStub>() {
        @java.lang.Override
        public TimelineServiceStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new TimelineServiceStub(channel, callOptions);
        }
      };
    return TimelineServiceStub.newStub(factory, channel);
  }

  /**
   * Creates a new blocking-style stub that supports unary and streaming output calls on the service
   */
  public static TimelineServiceBlockingStub newBlockingStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<TimelineServiceBlockingStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<TimelineServiceBlockingStub>() {
        @java.lang.Override
        public TimelineServiceBlockingStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new TimelineServiceBlockingStub(channel, callOptions);
        }
      };
    return TimelineServiceBlockingStub.newStub(factory, channel);
  }

  /**
   * Creates a new ListenableFuture-style stub that supports unary calls on the service
   */
  public static TimelineServiceFutureStub newFutureStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<TimelineServiceFutureStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<TimelineServiceFutureStub>() {
        @java.lang.Override
        public TimelineServiceFutureStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new TimelineServiceFutureStub(channel, callOptions);
        }
      };
    return TimelineServiceFutureStub.newStub(factory, channel);
  }

  /**
   */
  public interface AsyncService {

    /**
     * <pre>
     * Ping service for readiness
     * </pre>
     */
    default void ping(xyz.block.ftl.v1.PingRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.PingResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getPingMethod(), responseObserver);
    }

    /**
     * <pre>
     * Get timeline events with filters
     * </pre>
     */
    default void getTimeline(xyz.block.ftl.timeline.v1.GetTimelineRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.timeline.v1.GetTimelineResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getGetTimelineMethod(), responseObserver);
    }

    /**
     * <pre>
     * Stream timeline events with filters
     * </pre>
     */
    default void streamTimeline(xyz.block.ftl.timeline.v1.StreamTimelineRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.timeline.v1.StreamTimelineResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getStreamTimelineMethod(), responseObserver);
    }

    /**
     */
    default void createEvents(xyz.block.ftl.timeline.v1.CreateEventsRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.timeline.v1.CreateEventsResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getCreateEventsMethod(), responseObserver);
    }

    /**
     * <pre>
     * Delete old events of a specific type
     * </pre>
     */
    default void deleteOldEvents(xyz.block.ftl.timeline.v1.DeleteOldEventsRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.timeline.v1.DeleteOldEventsResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getDeleteOldEventsMethod(), responseObserver);
    }
  }

  /**
   * Base class for the server implementation of the service TimelineService.
   */
  public static abstract class TimelineServiceImplBase
      implements io.grpc.BindableService, AsyncService {

    @java.lang.Override public final io.grpc.ServerServiceDefinition bindService() {
      return TimelineServiceGrpc.bindService(this);
    }
  }

  /**
   * A stub to allow clients to do asynchronous rpc calls to service TimelineService.
   */
  public static final class TimelineServiceStub
      extends io.grpc.stub.AbstractAsyncStub<TimelineServiceStub> {
    private TimelineServiceStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected TimelineServiceStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new TimelineServiceStub(channel, callOptions);
    }

    /**
     * <pre>
     * Ping service for readiness
     * </pre>
     */
    public void ping(xyz.block.ftl.v1.PingRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.PingResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getPingMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * Get timeline events with filters
     * </pre>
     */
    public void getTimeline(xyz.block.ftl.timeline.v1.GetTimelineRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.timeline.v1.GetTimelineResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getGetTimelineMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * Stream timeline events with filters
     * </pre>
     */
    public void streamTimeline(xyz.block.ftl.timeline.v1.StreamTimelineRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.timeline.v1.StreamTimelineResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncServerStreamingCall(
          getChannel().newCall(getStreamTimelineMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void createEvents(xyz.block.ftl.timeline.v1.CreateEventsRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.timeline.v1.CreateEventsResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getCreateEventsMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * Delete old events of a specific type
     * </pre>
     */
    public void deleteOldEvents(xyz.block.ftl.timeline.v1.DeleteOldEventsRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.timeline.v1.DeleteOldEventsResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getDeleteOldEventsMethod(), getCallOptions()), request, responseObserver);
    }
  }

  /**
   * A stub to allow clients to do synchronous rpc calls to service TimelineService.
   */
  public static final class TimelineServiceBlockingStub
      extends io.grpc.stub.AbstractBlockingStub<TimelineServiceBlockingStub> {
    private TimelineServiceBlockingStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected TimelineServiceBlockingStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new TimelineServiceBlockingStub(channel, callOptions);
    }

    /**
     * <pre>
     * Ping service for readiness
     * </pre>
     */
    public xyz.block.ftl.v1.PingResponse ping(xyz.block.ftl.v1.PingRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getPingMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Get timeline events with filters
     * </pre>
     */
    public xyz.block.ftl.timeline.v1.GetTimelineResponse getTimeline(xyz.block.ftl.timeline.v1.GetTimelineRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getGetTimelineMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Stream timeline events with filters
     * </pre>
     */
    public java.util.Iterator<xyz.block.ftl.timeline.v1.StreamTimelineResponse> streamTimeline(
        xyz.block.ftl.timeline.v1.StreamTimelineRequest request) {
      return io.grpc.stub.ClientCalls.blockingServerStreamingCall(
          getChannel(), getStreamTimelineMethod(), getCallOptions(), request);
    }

    /**
     */
    public xyz.block.ftl.timeline.v1.CreateEventsResponse createEvents(xyz.block.ftl.timeline.v1.CreateEventsRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getCreateEventsMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Delete old events of a specific type
     * </pre>
     */
    public xyz.block.ftl.timeline.v1.DeleteOldEventsResponse deleteOldEvents(xyz.block.ftl.timeline.v1.DeleteOldEventsRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getDeleteOldEventsMethod(), getCallOptions(), request);
    }
  }

  /**
   * A stub to allow clients to do ListenableFuture-style rpc calls to service TimelineService.
   */
  public static final class TimelineServiceFutureStub
      extends io.grpc.stub.AbstractFutureStub<TimelineServiceFutureStub> {
    private TimelineServiceFutureStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected TimelineServiceFutureStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new TimelineServiceFutureStub(channel, callOptions);
    }

    /**
     * <pre>
     * Ping service for readiness
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<xyz.block.ftl.v1.PingResponse> ping(
        xyz.block.ftl.v1.PingRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getPingMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * Get timeline events with filters
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<xyz.block.ftl.timeline.v1.GetTimelineResponse> getTimeline(
        xyz.block.ftl.timeline.v1.GetTimelineRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getGetTimelineMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<xyz.block.ftl.timeline.v1.CreateEventsResponse> createEvents(
        xyz.block.ftl.timeline.v1.CreateEventsRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getCreateEventsMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * Delete old events of a specific type
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<xyz.block.ftl.timeline.v1.DeleteOldEventsResponse> deleteOldEvents(
        xyz.block.ftl.timeline.v1.DeleteOldEventsRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getDeleteOldEventsMethod(), getCallOptions()), request);
    }
  }

  private static final int METHODID_PING = 0;
  private static final int METHODID_GET_TIMELINE = 1;
  private static final int METHODID_STREAM_TIMELINE = 2;
  private static final int METHODID_CREATE_EVENTS = 3;
  private static final int METHODID_DELETE_OLD_EVENTS = 4;

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
        case METHODID_GET_TIMELINE:
          serviceImpl.getTimeline((xyz.block.ftl.timeline.v1.GetTimelineRequest) request,
              (io.grpc.stub.StreamObserver<xyz.block.ftl.timeline.v1.GetTimelineResponse>) responseObserver);
          break;
        case METHODID_STREAM_TIMELINE:
          serviceImpl.streamTimeline((xyz.block.ftl.timeline.v1.StreamTimelineRequest) request,
              (io.grpc.stub.StreamObserver<xyz.block.ftl.timeline.v1.StreamTimelineResponse>) responseObserver);
          break;
        case METHODID_CREATE_EVENTS:
          serviceImpl.createEvents((xyz.block.ftl.timeline.v1.CreateEventsRequest) request,
              (io.grpc.stub.StreamObserver<xyz.block.ftl.timeline.v1.CreateEventsResponse>) responseObserver);
          break;
        case METHODID_DELETE_OLD_EVENTS:
          serviceImpl.deleteOldEvents((xyz.block.ftl.timeline.v1.DeleteOldEventsRequest) request,
              (io.grpc.stub.StreamObserver<xyz.block.ftl.timeline.v1.DeleteOldEventsResponse>) responseObserver);
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
          getGetTimelineMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              xyz.block.ftl.timeline.v1.GetTimelineRequest,
              xyz.block.ftl.timeline.v1.GetTimelineResponse>(
                service, METHODID_GET_TIMELINE)))
        .addMethod(
          getStreamTimelineMethod(),
          io.grpc.stub.ServerCalls.asyncServerStreamingCall(
            new MethodHandlers<
              xyz.block.ftl.timeline.v1.StreamTimelineRequest,
              xyz.block.ftl.timeline.v1.StreamTimelineResponse>(
                service, METHODID_STREAM_TIMELINE)))
        .addMethod(
          getCreateEventsMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              xyz.block.ftl.timeline.v1.CreateEventsRequest,
              xyz.block.ftl.timeline.v1.CreateEventsResponse>(
                service, METHODID_CREATE_EVENTS)))
        .addMethod(
          getDeleteOldEventsMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              xyz.block.ftl.timeline.v1.DeleteOldEventsRequest,
              xyz.block.ftl.timeline.v1.DeleteOldEventsResponse>(
                service, METHODID_DELETE_OLD_EVENTS)))
        .build();
  }

  private static abstract class TimelineServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoFileDescriptorSupplier, io.grpc.protobuf.ProtoServiceDescriptorSupplier {
    TimelineServiceBaseDescriptorSupplier() {}

    @java.lang.Override
    public com.google.protobuf.Descriptors.FileDescriptor getFileDescriptor() {
      return xyz.block.ftl.timeline.v1.Timeline.getDescriptor();
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.ServiceDescriptor getServiceDescriptor() {
      return getFileDescriptor().findServiceByName("TimelineService");
    }
  }

  private static final class TimelineServiceFileDescriptorSupplier
      extends TimelineServiceBaseDescriptorSupplier {
    TimelineServiceFileDescriptorSupplier() {}
  }

  private static final class TimelineServiceMethodDescriptorSupplier
      extends TimelineServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoMethodDescriptorSupplier {
    private final java.lang.String methodName;

    TimelineServiceMethodDescriptorSupplier(java.lang.String methodName) {
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
      synchronized (TimelineServiceGrpc.class) {
        result = serviceDescriptor;
        if (result == null) {
          serviceDescriptor = result = io.grpc.ServiceDescriptor.newBuilder(SERVICE_NAME)
              .setSchemaDescriptor(new TimelineServiceFileDescriptorSupplier())
              .addMethod(getPingMethod())
              .addMethod(getGetTimelineMethod())
              .addMethod(getStreamTimelineMethod())
              .addMethod(getCreateEventsMethod())
              .addMethod(getDeleteOldEventsMethod())
              .build();
        }
      }
    }
    return result;
  }
}
