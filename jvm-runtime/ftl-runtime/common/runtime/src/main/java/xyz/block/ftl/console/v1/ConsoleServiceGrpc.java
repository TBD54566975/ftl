package xyz.block.ftl.console.v1;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 */
@javax.annotation.Generated(
    value = "by gRPC proto compiler (version 1.68.2)",
    comments = "Source: xyz/block/ftl/console/v1/console.proto")
@io.grpc.stub.annotations.GrpcGenerated
public final class ConsoleServiceGrpc {

  private ConsoleServiceGrpc() {}

  public static final java.lang.String SERVICE_NAME = "xyz.block.ftl.console.v1.ConsoleService";

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
    if ((getPingMethod = ConsoleServiceGrpc.getPingMethod) == null) {
      synchronized (ConsoleServiceGrpc.class) {
        if ((getPingMethod = ConsoleServiceGrpc.getPingMethod) == null) {
          ConsoleServiceGrpc.getPingMethod = getPingMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.v1.PingRequest, xyz.block.ftl.v1.PingResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "Ping"))
              .setSafe(true)
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.PingRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.PingResponse.getDefaultInstance()))
              .setSchemaDescriptor(new ConsoleServiceMethodDescriptorSupplier("Ping"))
              .build();
        }
      }
    }
    return getPingMethod;
  }

  private static volatile io.grpc.MethodDescriptor<xyz.block.ftl.console.v1.GetModulesRequest,
      xyz.block.ftl.console.v1.GetModulesResponse> getGetModulesMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "GetModules",
      requestType = xyz.block.ftl.console.v1.GetModulesRequest.class,
      responseType = xyz.block.ftl.console.v1.GetModulesResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<xyz.block.ftl.console.v1.GetModulesRequest,
      xyz.block.ftl.console.v1.GetModulesResponse> getGetModulesMethod() {
    io.grpc.MethodDescriptor<xyz.block.ftl.console.v1.GetModulesRequest, xyz.block.ftl.console.v1.GetModulesResponse> getGetModulesMethod;
    if ((getGetModulesMethod = ConsoleServiceGrpc.getGetModulesMethod) == null) {
      synchronized (ConsoleServiceGrpc.class) {
        if ((getGetModulesMethod = ConsoleServiceGrpc.getGetModulesMethod) == null) {
          ConsoleServiceGrpc.getGetModulesMethod = getGetModulesMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.console.v1.GetModulesRequest, xyz.block.ftl.console.v1.GetModulesResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "GetModules"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.console.v1.GetModulesRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.console.v1.GetModulesResponse.getDefaultInstance()))
              .setSchemaDescriptor(new ConsoleServiceMethodDescriptorSupplier("GetModules"))
              .build();
        }
      }
    }
    return getGetModulesMethod;
  }

  private static volatile io.grpc.MethodDescriptor<xyz.block.ftl.console.v1.StreamModulesRequest,
      xyz.block.ftl.console.v1.StreamModulesResponse> getStreamModulesMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "StreamModules",
      requestType = xyz.block.ftl.console.v1.StreamModulesRequest.class,
      responseType = xyz.block.ftl.console.v1.StreamModulesResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.SERVER_STREAMING)
  public static io.grpc.MethodDescriptor<xyz.block.ftl.console.v1.StreamModulesRequest,
      xyz.block.ftl.console.v1.StreamModulesResponse> getStreamModulesMethod() {
    io.grpc.MethodDescriptor<xyz.block.ftl.console.v1.StreamModulesRequest, xyz.block.ftl.console.v1.StreamModulesResponse> getStreamModulesMethod;
    if ((getStreamModulesMethod = ConsoleServiceGrpc.getStreamModulesMethod) == null) {
      synchronized (ConsoleServiceGrpc.class) {
        if ((getStreamModulesMethod = ConsoleServiceGrpc.getStreamModulesMethod) == null) {
          ConsoleServiceGrpc.getStreamModulesMethod = getStreamModulesMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.console.v1.StreamModulesRequest, xyz.block.ftl.console.v1.StreamModulesResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.SERVER_STREAMING)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "StreamModules"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.console.v1.StreamModulesRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.console.v1.StreamModulesResponse.getDefaultInstance()))
              .setSchemaDescriptor(new ConsoleServiceMethodDescriptorSupplier("StreamModules"))
              .build();
        }
      }
    }
    return getStreamModulesMethod;
  }

  private static volatile io.grpc.MethodDescriptor<xyz.block.ftl.console.v1.GetConfigRequest,
      xyz.block.ftl.console.v1.GetConfigResponse> getGetConfigMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "GetConfig",
      requestType = xyz.block.ftl.console.v1.GetConfigRequest.class,
      responseType = xyz.block.ftl.console.v1.GetConfigResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<xyz.block.ftl.console.v1.GetConfigRequest,
      xyz.block.ftl.console.v1.GetConfigResponse> getGetConfigMethod() {
    io.grpc.MethodDescriptor<xyz.block.ftl.console.v1.GetConfigRequest, xyz.block.ftl.console.v1.GetConfigResponse> getGetConfigMethod;
    if ((getGetConfigMethod = ConsoleServiceGrpc.getGetConfigMethod) == null) {
      synchronized (ConsoleServiceGrpc.class) {
        if ((getGetConfigMethod = ConsoleServiceGrpc.getGetConfigMethod) == null) {
          ConsoleServiceGrpc.getGetConfigMethod = getGetConfigMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.console.v1.GetConfigRequest, xyz.block.ftl.console.v1.GetConfigResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "GetConfig"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.console.v1.GetConfigRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.console.v1.GetConfigResponse.getDefaultInstance()))
              .setSchemaDescriptor(new ConsoleServiceMethodDescriptorSupplier("GetConfig"))
              .build();
        }
      }
    }
    return getGetConfigMethod;
  }

  private static volatile io.grpc.MethodDescriptor<xyz.block.ftl.console.v1.SetConfigRequest,
      xyz.block.ftl.console.v1.SetConfigResponse> getSetConfigMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "SetConfig",
      requestType = xyz.block.ftl.console.v1.SetConfigRequest.class,
      responseType = xyz.block.ftl.console.v1.SetConfigResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<xyz.block.ftl.console.v1.SetConfigRequest,
      xyz.block.ftl.console.v1.SetConfigResponse> getSetConfigMethod() {
    io.grpc.MethodDescriptor<xyz.block.ftl.console.v1.SetConfigRequest, xyz.block.ftl.console.v1.SetConfigResponse> getSetConfigMethod;
    if ((getSetConfigMethod = ConsoleServiceGrpc.getSetConfigMethod) == null) {
      synchronized (ConsoleServiceGrpc.class) {
        if ((getSetConfigMethod = ConsoleServiceGrpc.getSetConfigMethod) == null) {
          ConsoleServiceGrpc.getSetConfigMethod = getSetConfigMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.console.v1.SetConfigRequest, xyz.block.ftl.console.v1.SetConfigResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "SetConfig"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.console.v1.SetConfigRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.console.v1.SetConfigResponse.getDefaultInstance()))
              .setSchemaDescriptor(new ConsoleServiceMethodDescriptorSupplier("SetConfig"))
              .build();
        }
      }
    }
    return getSetConfigMethod;
  }

  private static volatile io.grpc.MethodDescriptor<xyz.block.ftl.console.v1.GetSecretRequest,
      xyz.block.ftl.console.v1.GetSecretResponse> getGetSecretMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "GetSecret",
      requestType = xyz.block.ftl.console.v1.GetSecretRequest.class,
      responseType = xyz.block.ftl.console.v1.GetSecretResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<xyz.block.ftl.console.v1.GetSecretRequest,
      xyz.block.ftl.console.v1.GetSecretResponse> getGetSecretMethod() {
    io.grpc.MethodDescriptor<xyz.block.ftl.console.v1.GetSecretRequest, xyz.block.ftl.console.v1.GetSecretResponse> getGetSecretMethod;
    if ((getGetSecretMethod = ConsoleServiceGrpc.getGetSecretMethod) == null) {
      synchronized (ConsoleServiceGrpc.class) {
        if ((getGetSecretMethod = ConsoleServiceGrpc.getGetSecretMethod) == null) {
          ConsoleServiceGrpc.getGetSecretMethod = getGetSecretMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.console.v1.GetSecretRequest, xyz.block.ftl.console.v1.GetSecretResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "GetSecret"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.console.v1.GetSecretRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.console.v1.GetSecretResponse.getDefaultInstance()))
              .setSchemaDescriptor(new ConsoleServiceMethodDescriptorSupplier("GetSecret"))
              .build();
        }
      }
    }
    return getGetSecretMethod;
  }

  private static volatile io.grpc.MethodDescriptor<xyz.block.ftl.console.v1.SetSecretRequest,
      xyz.block.ftl.console.v1.SetSecretResponse> getSetSecretMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "SetSecret",
      requestType = xyz.block.ftl.console.v1.SetSecretRequest.class,
      responseType = xyz.block.ftl.console.v1.SetSecretResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<xyz.block.ftl.console.v1.SetSecretRequest,
      xyz.block.ftl.console.v1.SetSecretResponse> getSetSecretMethod() {
    io.grpc.MethodDescriptor<xyz.block.ftl.console.v1.SetSecretRequest, xyz.block.ftl.console.v1.SetSecretResponse> getSetSecretMethod;
    if ((getSetSecretMethod = ConsoleServiceGrpc.getSetSecretMethod) == null) {
      synchronized (ConsoleServiceGrpc.class) {
        if ((getSetSecretMethod = ConsoleServiceGrpc.getSetSecretMethod) == null) {
          ConsoleServiceGrpc.getSetSecretMethod = getSetSecretMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.console.v1.SetSecretRequest, xyz.block.ftl.console.v1.SetSecretResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "SetSecret"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.console.v1.SetSecretRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.console.v1.SetSecretResponse.getDefaultInstance()))
              .setSchemaDescriptor(new ConsoleServiceMethodDescriptorSupplier("SetSecret"))
              .build();
        }
      }
    }
    return getSetSecretMethod;
  }

  /**
   * Creates a new async stub that supports all call types for the service
   */
  public static ConsoleServiceStub newStub(io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<ConsoleServiceStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<ConsoleServiceStub>() {
        @java.lang.Override
        public ConsoleServiceStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new ConsoleServiceStub(channel, callOptions);
        }
      };
    return ConsoleServiceStub.newStub(factory, channel);
  }

  /**
   * Creates a new blocking-style stub that supports unary and streaming output calls on the service
   */
  public static ConsoleServiceBlockingStub newBlockingStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<ConsoleServiceBlockingStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<ConsoleServiceBlockingStub>() {
        @java.lang.Override
        public ConsoleServiceBlockingStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new ConsoleServiceBlockingStub(channel, callOptions);
        }
      };
    return ConsoleServiceBlockingStub.newStub(factory, channel);
  }

  /**
   * Creates a new ListenableFuture-style stub that supports unary calls on the service
   */
  public static ConsoleServiceFutureStub newFutureStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<ConsoleServiceFutureStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<ConsoleServiceFutureStub>() {
        @java.lang.Override
        public ConsoleServiceFutureStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new ConsoleServiceFutureStub(channel, callOptions);
        }
      };
    return ConsoleServiceFutureStub.newStub(factory, channel);
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
     */
    default void getModules(xyz.block.ftl.console.v1.GetModulesRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.console.v1.GetModulesResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getGetModulesMethod(), responseObserver);
    }

    /**
     */
    default void streamModules(xyz.block.ftl.console.v1.StreamModulesRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.console.v1.StreamModulesResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getStreamModulesMethod(), responseObserver);
    }

    /**
     */
    default void getConfig(xyz.block.ftl.console.v1.GetConfigRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.console.v1.GetConfigResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getGetConfigMethod(), responseObserver);
    }

    /**
     */
    default void setConfig(xyz.block.ftl.console.v1.SetConfigRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.console.v1.SetConfigResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getSetConfigMethod(), responseObserver);
    }

    /**
     */
    default void getSecret(xyz.block.ftl.console.v1.GetSecretRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.console.v1.GetSecretResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getGetSecretMethod(), responseObserver);
    }

    /**
     */
    default void setSecret(xyz.block.ftl.console.v1.SetSecretRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.console.v1.SetSecretResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getSetSecretMethod(), responseObserver);
    }
  }

  /**
   * Base class for the server implementation of the service ConsoleService.
   */
  public static abstract class ConsoleServiceImplBase
      implements io.grpc.BindableService, AsyncService {

    @java.lang.Override public final io.grpc.ServerServiceDefinition bindService() {
      return ConsoleServiceGrpc.bindService(this);
    }
  }

  /**
   * A stub to allow clients to do asynchronous rpc calls to service ConsoleService.
   */
  public static final class ConsoleServiceStub
      extends io.grpc.stub.AbstractAsyncStub<ConsoleServiceStub> {
    private ConsoleServiceStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected ConsoleServiceStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new ConsoleServiceStub(channel, callOptions);
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
     */
    public void getModules(xyz.block.ftl.console.v1.GetModulesRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.console.v1.GetModulesResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getGetModulesMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void streamModules(xyz.block.ftl.console.v1.StreamModulesRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.console.v1.StreamModulesResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncServerStreamingCall(
          getChannel().newCall(getStreamModulesMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void getConfig(xyz.block.ftl.console.v1.GetConfigRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.console.v1.GetConfigResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getGetConfigMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void setConfig(xyz.block.ftl.console.v1.SetConfigRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.console.v1.SetConfigResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getSetConfigMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void getSecret(xyz.block.ftl.console.v1.GetSecretRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.console.v1.GetSecretResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getGetSecretMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     */
    public void setSecret(xyz.block.ftl.console.v1.SetSecretRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.console.v1.SetSecretResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getSetSecretMethod(), getCallOptions()), request, responseObserver);
    }
  }

  /**
   * A stub to allow clients to do synchronous rpc calls to service ConsoleService.
   */
  public static final class ConsoleServiceBlockingStub
      extends io.grpc.stub.AbstractBlockingStub<ConsoleServiceBlockingStub> {
    private ConsoleServiceBlockingStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected ConsoleServiceBlockingStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new ConsoleServiceBlockingStub(channel, callOptions);
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
     */
    public xyz.block.ftl.console.v1.GetModulesResponse getModules(xyz.block.ftl.console.v1.GetModulesRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getGetModulesMethod(), getCallOptions(), request);
    }

    /**
     */
    public java.util.Iterator<xyz.block.ftl.console.v1.StreamModulesResponse> streamModules(
        xyz.block.ftl.console.v1.StreamModulesRequest request) {
      return io.grpc.stub.ClientCalls.blockingServerStreamingCall(
          getChannel(), getStreamModulesMethod(), getCallOptions(), request);
    }

    /**
     */
    public xyz.block.ftl.console.v1.GetConfigResponse getConfig(xyz.block.ftl.console.v1.GetConfigRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getGetConfigMethod(), getCallOptions(), request);
    }

    /**
     */
    public xyz.block.ftl.console.v1.SetConfigResponse setConfig(xyz.block.ftl.console.v1.SetConfigRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getSetConfigMethod(), getCallOptions(), request);
    }

    /**
     */
    public xyz.block.ftl.console.v1.GetSecretResponse getSecret(xyz.block.ftl.console.v1.GetSecretRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getGetSecretMethod(), getCallOptions(), request);
    }

    /**
     */
    public xyz.block.ftl.console.v1.SetSecretResponse setSecret(xyz.block.ftl.console.v1.SetSecretRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getSetSecretMethod(), getCallOptions(), request);
    }
  }

  /**
   * A stub to allow clients to do ListenableFuture-style rpc calls to service ConsoleService.
   */
  public static final class ConsoleServiceFutureStub
      extends io.grpc.stub.AbstractFutureStub<ConsoleServiceFutureStub> {
    private ConsoleServiceFutureStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected ConsoleServiceFutureStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new ConsoleServiceFutureStub(channel, callOptions);
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
     */
    public com.google.common.util.concurrent.ListenableFuture<xyz.block.ftl.console.v1.GetModulesResponse> getModules(
        xyz.block.ftl.console.v1.GetModulesRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getGetModulesMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<xyz.block.ftl.console.v1.GetConfigResponse> getConfig(
        xyz.block.ftl.console.v1.GetConfigRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getGetConfigMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<xyz.block.ftl.console.v1.SetConfigResponse> setConfig(
        xyz.block.ftl.console.v1.SetConfigRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getSetConfigMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<xyz.block.ftl.console.v1.GetSecretResponse> getSecret(
        xyz.block.ftl.console.v1.GetSecretRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getGetSecretMethod(), getCallOptions()), request);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<xyz.block.ftl.console.v1.SetSecretResponse> setSecret(
        xyz.block.ftl.console.v1.SetSecretRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getSetSecretMethod(), getCallOptions()), request);
    }
  }

  private static final int METHODID_PING = 0;
  private static final int METHODID_GET_MODULES = 1;
  private static final int METHODID_STREAM_MODULES = 2;
  private static final int METHODID_GET_CONFIG = 3;
  private static final int METHODID_SET_CONFIG = 4;
  private static final int METHODID_GET_SECRET = 5;
  private static final int METHODID_SET_SECRET = 6;

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
        case METHODID_GET_MODULES:
          serviceImpl.getModules((xyz.block.ftl.console.v1.GetModulesRequest) request,
              (io.grpc.stub.StreamObserver<xyz.block.ftl.console.v1.GetModulesResponse>) responseObserver);
          break;
        case METHODID_STREAM_MODULES:
          serviceImpl.streamModules((xyz.block.ftl.console.v1.StreamModulesRequest) request,
              (io.grpc.stub.StreamObserver<xyz.block.ftl.console.v1.StreamModulesResponse>) responseObserver);
          break;
        case METHODID_GET_CONFIG:
          serviceImpl.getConfig((xyz.block.ftl.console.v1.GetConfigRequest) request,
              (io.grpc.stub.StreamObserver<xyz.block.ftl.console.v1.GetConfigResponse>) responseObserver);
          break;
        case METHODID_SET_CONFIG:
          serviceImpl.setConfig((xyz.block.ftl.console.v1.SetConfigRequest) request,
              (io.grpc.stub.StreamObserver<xyz.block.ftl.console.v1.SetConfigResponse>) responseObserver);
          break;
        case METHODID_GET_SECRET:
          serviceImpl.getSecret((xyz.block.ftl.console.v1.GetSecretRequest) request,
              (io.grpc.stub.StreamObserver<xyz.block.ftl.console.v1.GetSecretResponse>) responseObserver);
          break;
        case METHODID_SET_SECRET:
          serviceImpl.setSecret((xyz.block.ftl.console.v1.SetSecretRequest) request,
              (io.grpc.stub.StreamObserver<xyz.block.ftl.console.v1.SetSecretResponse>) responseObserver);
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
          getGetModulesMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              xyz.block.ftl.console.v1.GetModulesRequest,
              xyz.block.ftl.console.v1.GetModulesResponse>(
                service, METHODID_GET_MODULES)))
        .addMethod(
          getStreamModulesMethod(),
          io.grpc.stub.ServerCalls.asyncServerStreamingCall(
            new MethodHandlers<
              xyz.block.ftl.console.v1.StreamModulesRequest,
              xyz.block.ftl.console.v1.StreamModulesResponse>(
                service, METHODID_STREAM_MODULES)))
        .addMethod(
          getGetConfigMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              xyz.block.ftl.console.v1.GetConfigRequest,
              xyz.block.ftl.console.v1.GetConfigResponse>(
                service, METHODID_GET_CONFIG)))
        .addMethod(
          getSetConfigMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              xyz.block.ftl.console.v1.SetConfigRequest,
              xyz.block.ftl.console.v1.SetConfigResponse>(
                service, METHODID_SET_CONFIG)))
        .addMethod(
          getGetSecretMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              xyz.block.ftl.console.v1.GetSecretRequest,
              xyz.block.ftl.console.v1.GetSecretResponse>(
                service, METHODID_GET_SECRET)))
        .addMethod(
          getSetSecretMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              xyz.block.ftl.console.v1.SetSecretRequest,
              xyz.block.ftl.console.v1.SetSecretResponse>(
                service, METHODID_SET_SECRET)))
        .build();
  }

  private static abstract class ConsoleServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoFileDescriptorSupplier, io.grpc.protobuf.ProtoServiceDescriptorSupplier {
    ConsoleServiceBaseDescriptorSupplier() {}

    @java.lang.Override
    public com.google.protobuf.Descriptors.FileDescriptor getFileDescriptor() {
      return xyz.block.ftl.console.v1.Console.getDescriptor();
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.ServiceDescriptor getServiceDescriptor() {
      return getFileDescriptor().findServiceByName("ConsoleService");
    }
  }

  private static final class ConsoleServiceFileDescriptorSupplier
      extends ConsoleServiceBaseDescriptorSupplier {
    ConsoleServiceFileDescriptorSupplier() {}
  }

  private static final class ConsoleServiceMethodDescriptorSupplier
      extends ConsoleServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoMethodDescriptorSupplier {
    private final java.lang.String methodName;

    ConsoleServiceMethodDescriptorSupplier(java.lang.String methodName) {
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
      synchronized (ConsoleServiceGrpc.class) {
        result = serviceDescriptor;
        if (result == null) {
          serviceDescriptor = result = io.grpc.ServiceDescriptor.newBuilder(SERVICE_NAME)
              .setSchemaDescriptor(new ConsoleServiceFileDescriptorSupplier())
              .addMethod(getPingMethod())
              .addMethod(getGetModulesMethod())
              .addMethod(getStreamModulesMethod())
              .addMethod(getGetConfigMethod())
              .addMethod(getSetConfigMethod())
              .addMethod(getGetSecretMethod())
              .addMethod(getSetSecretMethod())
              .build();
        }
      }
    }
    return result;
  }
}
