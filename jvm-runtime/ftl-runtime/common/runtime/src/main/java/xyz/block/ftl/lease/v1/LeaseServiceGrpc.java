package xyz.block.ftl.lease.v1;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 * <pre>
 * ModuleService is the service that modules use to interact with the Controller.
 * </pre>
 */
@javax.annotation.Generated(
    value = "by gRPC proto compiler (version 1.68.2)",
    comments = "Source: xyz/block/ftl/lease/v1/lease.proto")
@io.grpc.stub.annotations.GrpcGenerated
public final class LeaseServiceGrpc {

  private LeaseServiceGrpc() {}

  public static final java.lang.String SERVICE_NAME = "xyz.block.ftl.lease.v1.LeaseService";

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
    if ((getPingMethod = LeaseServiceGrpc.getPingMethod) == null) {
      synchronized (LeaseServiceGrpc.class) {
        if ((getPingMethod = LeaseServiceGrpc.getPingMethod) == null) {
          LeaseServiceGrpc.getPingMethod = getPingMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.v1.PingRequest, xyz.block.ftl.v1.PingResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "Ping"))
              .setSafe(true)
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.PingRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.PingResponse.getDefaultInstance()))
              .setSchemaDescriptor(new LeaseServiceMethodDescriptorSupplier("Ping"))
              .build();
        }
      }
    }
    return getPingMethod;
  }

  private static volatile io.grpc.MethodDescriptor<xyz.block.ftl.lease.v1.AcquireLeaseRequest,
      xyz.block.ftl.lease.v1.AcquireLeaseResponse> getAcquireLeaseMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "AcquireLease",
      requestType = xyz.block.ftl.lease.v1.AcquireLeaseRequest.class,
      responseType = xyz.block.ftl.lease.v1.AcquireLeaseResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.BIDI_STREAMING)
  public static io.grpc.MethodDescriptor<xyz.block.ftl.lease.v1.AcquireLeaseRequest,
      xyz.block.ftl.lease.v1.AcquireLeaseResponse> getAcquireLeaseMethod() {
    io.grpc.MethodDescriptor<xyz.block.ftl.lease.v1.AcquireLeaseRequest, xyz.block.ftl.lease.v1.AcquireLeaseResponse> getAcquireLeaseMethod;
    if ((getAcquireLeaseMethod = LeaseServiceGrpc.getAcquireLeaseMethod) == null) {
      synchronized (LeaseServiceGrpc.class) {
        if ((getAcquireLeaseMethod = LeaseServiceGrpc.getAcquireLeaseMethod) == null) {
          LeaseServiceGrpc.getAcquireLeaseMethod = getAcquireLeaseMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.lease.v1.AcquireLeaseRequest, xyz.block.ftl.lease.v1.AcquireLeaseResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.BIDI_STREAMING)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "AcquireLease"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.lease.v1.AcquireLeaseRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.lease.v1.AcquireLeaseResponse.getDefaultInstance()))
              .setSchemaDescriptor(new LeaseServiceMethodDescriptorSupplier("AcquireLease"))
              .build();
        }
      }
    }
    return getAcquireLeaseMethod;
  }

  /**
   * Creates a new async stub that supports all call types for the service
   */
  public static LeaseServiceStub newStub(io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<LeaseServiceStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<LeaseServiceStub>() {
        @java.lang.Override
        public LeaseServiceStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new LeaseServiceStub(channel, callOptions);
        }
      };
    return LeaseServiceStub.newStub(factory, channel);
  }

  /**
   * Creates a new blocking-style stub that supports unary and streaming output calls on the service
   */
  public static LeaseServiceBlockingStub newBlockingStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<LeaseServiceBlockingStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<LeaseServiceBlockingStub>() {
        @java.lang.Override
        public LeaseServiceBlockingStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new LeaseServiceBlockingStub(channel, callOptions);
        }
      };
    return LeaseServiceBlockingStub.newStub(factory, channel);
  }

  /**
   * Creates a new ListenableFuture-style stub that supports unary calls on the service
   */
  public static LeaseServiceFutureStub newFutureStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<LeaseServiceFutureStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<LeaseServiceFutureStub>() {
        @java.lang.Override
        public LeaseServiceFutureStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new LeaseServiceFutureStub(channel, callOptions);
        }
      };
    return LeaseServiceFutureStub.newStub(factory, channel);
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
     * Acquire (and renew) a lease for a deployment.
     * Returns ResourceExhausted if the lease is held.
     * </pre>
     */
    default io.grpc.stub.StreamObserver<xyz.block.ftl.lease.v1.AcquireLeaseRequest> acquireLease(
        io.grpc.stub.StreamObserver<xyz.block.ftl.lease.v1.AcquireLeaseResponse> responseObserver) {
      return io.grpc.stub.ServerCalls.asyncUnimplementedStreamingCall(getAcquireLeaseMethod(), responseObserver);
    }
  }

  /**
   * Base class for the server implementation of the service LeaseService.
   * <pre>
   * ModuleService is the service that modules use to interact with the Controller.
   * </pre>
   */
  public static abstract class LeaseServiceImplBase
      implements io.grpc.BindableService, AsyncService {

    @java.lang.Override public final io.grpc.ServerServiceDefinition bindService() {
      return LeaseServiceGrpc.bindService(this);
    }
  }

  /**
   * A stub to allow clients to do asynchronous rpc calls to service LeaseService.
   * <pre>
   * ModuleService is the service that modules use to interact with the Controller.
   * </pre>
   */
  public static final class LeaseServiceStub
      extends io.grpc.stub.AbstractAsyncStub<LeaseServiceStub> {
    private LeaseServiceStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected LeaseServiceStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new LeaseServiceStub(channel, callOptions);
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
     * Acquire (and renew) a lease for a deployment.
     * Returns ResourceExhausted if the lease is held.
     * </pre>
     */
    public io.grpc.stub.StreamObserver<xyz.block.ftl.lease.v1.AcquireLeaseRequest> acquireLease(
        io.grpc.stub.StreamObserver<xyz.block.ftl.lease.v1.AcquireLeaseResponse> responseObserver) {
      return io.grpc.stub.ClientCalls.asyncBidiStreamingCall(
          getChannel().newCall(getAcquireLeaseMethod(), getCallOptions()), responseObserver);
    }
  }

  /**
   * A stub to allow clients to do synchronous rpc calls to service LeaseService.
   * <pre>
   * ModuleService is the service that modules use to interact with the Controller.
   * </pre>
   */
  public static final class LeaseServiceBlockingStub
      extends io.grpc.stub.AbstractBlockingStub<LeaseServiceBlockingStub> {
    private LeaseServiceBlockingStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected LeaseServiceBlockingStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new LeaseServiceBlockingStub(channel, callOptions);
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
  }

  /**
   * A stub to allow clients to do ListenableFuture-style rpc calls to service LeaseService.
   * <pre>
   * ModuleService is the service that modules use to interact with the Controller.
   * </pre>
   */
  public static final class LeaseServiceFutureStub
      extends io.grpc.stub.AbstractFutureStub<LeaseServiceFutureStub> {
    private LeaseServiceFutureStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected LeaseServiceFutureStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new LeaseServiceFutureStub(channel, callOptions);
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
  private static final int METHODID_ACQUIRE_LEASE = 1;

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
        default:
          throw new AssertionError();
      }
    }

    @java.lang.Override
    @java.lang.SuppressWarnings("unchecked")
    public io.grpc.stub.StreamObserver<Req> invoke(
        io.grpc.stub.StreamObserver<Resp> responseObserver) {
      switch (methodId) {
        case METHODID_ACQUIRE_LEASE:
          return (io.grpc.stub.StreamObserver<Req>) serviceImpl.acquireLease(
              (io.grpc.stub.StreamObserver<xyz.block.ftl.lease.v1.AcquireLeaseResponse>) responseObserver);
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
          getAcquireLeaseMethod(),
          io.grpc.stub.ServerCalls.asyncBidiStreamingCall(
            new MethodHandlers<
              xyz.block.ftl.lease.v1.AcquireLeaseRequest,
              xyz.block.ftl.lease.v1.AcquireLeaseResponse>(
                service, METHODID_ACQUIRE_LEASE)))
        .build();
  }

  private static abstract class LeaseServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoFileDescriptorSupplier, io.grpc.protobuf.ProtoServiceDescriptorSupplier {
    LeaseServiceBaseDescriptorSupplier() {}

    @java.lang.Override
    public com.google.protobuf.Descriptors.FileDescriptor getFileDescriptor() {
      return xyz.block.ftl.lease.v1.Lease.getDescriptor();
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.ServiceDescriptor getServiceDescriptor() {
      return getFileDescriptor().findServiceByName("LeaseService");
    }
  }

  private static final class LeaseServiceFileDescriptorSupplier
      extends LeaseServiceBaseDescriptorSupplier {
    LeaseServiceFileDescriptorSupplier() {}
  }

  private static final class LeaseServiceMethodDescriptorSupplier
      extends LeaseServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoMethodDescriptorSupplier {
    private final java.lang.String methodName;

    LeaseServiceMethodDescriptorSupplier(java.lang.String methodName) {
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
      synchronized (LeaseServiceGrpc.class) {
        result = serviceDescriptor;
        if (result == null) {
          serviceDescriptor = result = io.grpc.ServiceDescriptor.newBuilder(SERVICE_NAME)
              .setSchemaDescriptor(new LeaseServiceFileDescriptorSupplier())
              .addMethod(getPingMethod())
              .addMethod(getAcquireLeaseMethod())
              .build();
        }
      }
    }
    return result;
  }
}
