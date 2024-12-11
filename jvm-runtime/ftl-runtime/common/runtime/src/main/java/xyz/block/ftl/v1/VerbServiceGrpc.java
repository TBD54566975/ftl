package xyz.block.ftl.v1;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 * <pre>
 * VerbService is a common interface shared by multiple services for calling Verbs.
 * </pre>
 */
@javax.annotation.Generated(
    value = "by gRPC proto compiler (version 1.68.2)",
    comments = "Source: xyz/block/ftl/v1/verb.proto")
@io.grpc.stub.annotations.GrpcGenerated
public final class VerbServiceGrpc {

  private VerbServiceGrpc() {}

  public static final java.lang.String SERVICE_NAME = "xyz.block.ftl.v1.VerbService";

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
    if ((getPingMethod = VerbServiceGrpc.getPingMethod) == null) {
      synchronized (VerbServiceGrpc.class) {
        if ((getPingMethod = VerbServiceGrpc.getPingMethod) == null) {
          VerbServiceGrpc.getPingMethod = getPingMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.v1.PingRequest, xyz.block.ftl.v1.PingResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "Ping"))
              .setSafe(true)
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.PingRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.PingResponse.getDefaultInstance()))
              .setSchemaDescriptor(new VerbServiceMethodDescriptorSupplier("Ping"))
              .build();
        }
      }
    }
    return getPingMethod;
  }

  private static volatile io.grpc.MethodDescriptor<xyz.block.ftl.v1.CallRequest,
      xyz.block.ftl.v1.CallResponse> getCallMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "Call",
      requestType = xyz.block.ftl.v1.CallRequest.class,
      responseType = xyz.block.ftl.v1.CallResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<xyz.block.ftl.v1.CallRequest,
      xyz.block.ftl.v1.CallResponse> getCallMethod() {
    io.grpc.MethodDescriptor<xyz.block.ftl.v1.CallRequest, xyz.block.ftl.v1.CallResponse> getCallMethod;
    if ((getCallMethod = VerbServiceGrpc.getCallMethod) == null) {
      synchronized (VerbServiceGrpc.class) {
        if ((getCallMethod = VerbServiceGrpc.getCallMethod) == null) {
          VerbServiceGrpc.getCallMethod = getCallMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.v1.CallRequest, xyz.block.ftl.v1.CallResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "Call"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.CallRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.CallResponse.getDefaultInstance()))
              .setSchemaDescriptor(new VerbServiceMethodDescriptorSupplier("Call"))
              .build();
        }
      }
    }
    return getCallMethod;
  }

  /**
   * Creates a new async stub that supports all call types for the service
   */
  public static VerbServiceStub newStub(io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<VerbServiceStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<VerbServiceStub>() {
        @java.lang.Override
        public VerbServiceStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new VerbServiceStub(channel, callOptions);
        }
      };
    return VerbServiceStub.newStub(factory, channel);
  }

  /**
   * Creates a new blocking-style stub that supports unary and streaming output calls on the service
   */
  public static VerbServiceBlockingStub newBlockingStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<VerbServiceBlockingStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<VerbServiceBlockingStub>() {
        @java.lang.Override
        public VerbServiceBlockingStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new VerbServiceBlockingStub(channel, callOptions);
        }
      };
    return VerbServiceBlockingStub.newStub(factory, channel);
  }

  /**
   * Creates a new ListenableFuture-style stub that supports unary calls on the service
   */
  public static VerbServiceFutureStub newFutureStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<VerbServiceFutureStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<VerbServiceFutureStub>() {
        @java.lang.Override
        public VerbServiceFutureStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new VerbServiceFutureStub(channel, callOptions);
        }
      };
    return VerbServiceFutureStub.newStub(factory, channel);
  }

  /**
   * <pre>
   * VerbService is a common interface shared by multiple services for calling Verbs.
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
     * Issue a synchronous call to a Verb.
     * </pre>
     */
    default void call(xyz.block.ftl.v1.CallRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.CallResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getCallMethod(), responseObserver);
    }
  }

  /**
   * Base class for the server implementation of the service VerbService.
   * <pre>
   * VerbService is a common interface shared by multiple services for calling Verbs.
   * </pre>
   */
  public static abstract class VerbServiceImplBase
      implements io.grpc.BindableService, AsyncService {

    @java.lang.Override public final io.grpc.ServerServiceDefinition bindService() {
      return VerbServiceGrpc.bindService(this);
    }
  }

  /**
   * A stub to allow clients to do asynchronous rpc calls to service VerbService.
   * <pre>
   * VerbService is a common interface shared by multiple services for calling Verbs.
   * </pre>
   */
  public static final class VerbServiceStub
      extends io.grpc.stub.AbstractAsyncStub<VerbServiceStub> {
    private VerbServiceStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected VerbServiceStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new VerbServiceStub(channel, callOptions);
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
     * Issue a synchronous call to a Verb.
     * </pre>
     */
    public void call(xyz.block.ftl.v1.CallRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.CallResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getCallMethod(), getCallOptions()), request, responseObserver);
    }
  }

  /**
   * A stub to allow clients to do synchronous rpc calls to service VerbService.
   * <pre>
   * VerbService is a common interface shared by multiple services for calling Verbs.
   * </pre>
   */
  public static final class VerbServiceBlockingStub
      extends io.grpc.stub.AbstractBlockingStub<VerbServiceBlockingStub> {
    private VerbServiceBlockingStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected VerbServiceBlockingStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new VerbServiceBlockingStub(channel, callOptions);
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
     * Issue a synchronous call to a Verb.
     * </pre>
     */
    public xyz.block.ftl.v1.CallResponse call(xyz.block.ftl.v1.CallRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getCallMethod(), getCallOptions(), request);
    }
  }

  /**
   * A stub to allow clients to do ListenableFuture-style rpc calls to service VerbService.
   * <pre>
   * VerbService is a common interface shared by multiple services for calling Verbs.
   * </pre>
   */
  public static final class VerbServiceFutureStub
      extends io.grpc.stub.AbstractFutureStub<VerbServiceFutureStub> {
    private VerbServiceFutureStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected VerbServiceFutureStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new VerbServiceFutureStub(channel, callOptions);
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
     * Issue a synchronous call to a Verb.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<xyz.block.ftl.v1.CallResponse> call(
        xyz.block.ftl.v1.CallRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getCallMethod(), getCallOptions()), request);
    }
  }

  private static final int METHODID_PING = 0;
  private static final int METHODID_CALL = 1;

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
        case METHODID_CALL:
          serviceImpl.call((xyz.block.ftl.v1.CallRequest) request,
              (io.grpc.stub.StreamObserver<xyz.block.ftl.v1.CallResponse>) responseObserver);
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
          getCallMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              xyz.block.ftl.v1.CallRequest,
              xyz.block.ftl.v1.CallResponse>(
                service, METHODID_CALL)))
        .build();
  }

  private static abstract class VerbServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoFileDescriptorSupplier, io.grpc.protobuf.ProtoServiceDescriptorSupplier {
    VerbServiceBaseDescriptorSupplier() {}

    @java.lang.Override
    public com.google.protobuf.Descriptors.FileDescriptor getFileDescriptor() {
      return xyz.block.ftl.v1.Verb.getDescriptor();
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.ServiceDescriptor getServiceDescriptor() {
      return getFileDescriptor().findServiceByName("VerbService");
    }
  }

  private static final class VerbServiceFileDescriptorSupplier
      extends VerbServiceBaseDescriptorSupplier {
    VerbServiceFileDescriptorSupplier() {}
  }

  private static final class VerbServiceMethodDescriptorSupplier
      extends VerbServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoMethodDescriptorSupplier {
    private final java.lang.String methodName;

    VerbServiceMethodDescriptorSupplier(java.lang.String methodName) {
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
      synchronized (VerbServiceGrpc.class) {
        result = serviceDescriptor;
        if (result == null) {
          serviceDescriptor = result = io.grpc.ServiceDescriptor.newBuilder(SERVICE_NAME)
              .setSchemaDescriptor(new VerbServiceFileDescriptorSupplier())
              .addMethod(getPingMethod())
              .addMethod(getCallMethod())
              .build();
        }
      }
    }
    return result;
  }
}
