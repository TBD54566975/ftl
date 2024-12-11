package xyz.block.ftl.v1;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 * <pre>
 * AdminService is the service that provides and updates admin data. For example,
 * it is used to encapsulate configuration and secrets.
 * </pre>
 */
@javax.annotation.Generated(
    value = "by gRPC proto compiler (version 1.68.2)",
    comments = "Source: xyz/block/ftl/v1/admin.proto")
@io.grpc.stub.annotations.GrpcGenerated
public final class AdminServiceGrpc {

  private AdminServiceGrpc() {}

  public static final java.lang.String SERVICE_NAME = "xyz.block.ftl.v1.AdminService";

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
    if ((getPingMethod = AdminServiceGrpc.getPingMethod) == null) {
      synchronized (AdminServiceGrpc.class) {
        if ((getPingMethod = AdminServiceGrpc.getPingMethod) == null) {
          AdminServiceGrpc.getPingMethod = getPingMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.v1.PingRequest, xyz.block.ftl.v1.PingResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "Ping"))
              .setSafe(true)
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.PingRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.PingResponse.getDefaultInstance()))
              .setSchemaDescriptor(new AdminServiceMethodDescriptorSupplier("Ping"))
              .build();
        }
      }
    }
    return getPingMethod;
  }

  private static volatile io.grpc.MethodDescriptor<xyz.block.ftl.v1.ConfigListRequest,
      xyz.block.ftl.v1.ConfigListResponse> getConfigListMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "ConfigList",
      requestType = xyz.block.ftl.v1.ConfigListRequest.class,
      responseType = xyz.block.ftl.v1.ConfigListResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<xyz.block.ftl.v1.ConfigListRequest,
      xyz.block.ftl.v1.ConfigListResponse> getConfigListMethod() {
    io.grpc.MethodDescriptor<xyz.block.ftl.v1.ConfigListRequest, xyz.block.ftl.v1.ConfigListResponse> getConfigListMethod;
    if ((getConfigListMethod = AdminServiceGrpc.getConfigListMethod) == null) {
      synchronized (AdminServiceGrpc.class) {
        if ((getConfigListMethod = AdminServiceGrpc.getConfigListMethod) == null) {
          AdminServiceGrpc.getConfigListMethod = getConfigListMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.v1.ConfigListRequest, xyz.block.ftl.v1.ConfigListResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "ConfigList"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.ConfigListRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.ConfigListResponse.getDefaultInstance()))
              .setSchemaDescriptor(new AdminServiceMethodDescriptorSupplier("ConfigList"))
              .build();
        }
      }
    }
    return getConfigListMethod;
  }

  private static volatile io.grpc.MethodDescriptor<xyz.block.ftl.v1.ConfigGetRequest,
      xyz.block.ftl.v1.ConfigGetResponse> getConfigGetMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "ConfigGet",
      requestType = xyz.block.ftl.v1.ConfigGetRequest.class,
      responseType = xyz.block.ftl.v1.ConfigGetResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<xyz.block.ftl.v1.ConfigGetRequest,
      xyz.block.ftl.v1.ConfigGetResponse> getConfigGetMethod() {
    io.grpc.MethodDescriptor<xyz.block.ftl.v1.ConfigGetRequest, xyz.block.ftl.v1.ConfigGetResponse> getConfigGetMethod;
    if ((getConfigGetMethod = AdminServiceGrpc.getConfigGetMethod) == null) {
      synchronized (AdminServiceGrpc.class) {
        if ((getConfigGetMethod = AdminServiceGrpc.getConfigGetMethod) == null) {
          AdminServiceGrpc.getConfigGetMethod = getConfigGetMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.v1.ConfigGetRequest, xyz.block.ftl.v1.ConfigGetResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "ConfigGet"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.ConfigGetRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.ConfigGetResponse.getDefaultInstance()))
              .setSchemaDescriptor(new AdminServiceMethodDescriptorSupplier("ConfigGet"))
              .build();
        }
      }
    }
    return getConfigGetMethod;
  }

  private static volatile io.grpc.MethodDescriptor<xyz.block.ftl.v1.ConfigSetRequest,
      xyz.block.ftl.v1.ConfigSetResponse> getConfigSetMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "ConfigSet",
      requestType = xyz.block.ftl.v1.ConfigSetRequest.class,
      responseType = xyz.block.ftl.v1.ConfigSetResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<xyz.block.ftl.v1.ConfigSetRequest,
      xyz.block.ftl.v1.ConfigSetResponse> getConfigSetMethod() {
    io.grpc.MethodDescriptor<xyz.block.ftl.v1.ConfigSetRequest, xyz.block.ftl.v1.ConfigSetResponse> getConfigSetMethod;
    if ((getConfigSetMethod = AdminServiceGrpc.getConfigSetMethod) == null) {
      synchronized (AdminServiceGrpc.class) {
        if ((getConfigSetMethod = AdminServiceGrpc.getConfigSetMethod) == null) {
          AdminServiceGrpc.getConfigSetMethod = getConfigSetMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.v1.ConfigSetRequest, xyz.block.ftl.v1.ConfigSetResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "ConfigSet"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.ConfigSetRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.ConfigSetResponse.getDefaultInstance()))
              .setSchemaDescriptor(new AdminServiceMethodDescriptorSupplier("ConfigSet"))
              .build();
        }
      }
    }
    return getConfigSetMethod;
  }

  private static volatile io.grpc.MethodDescriptor<xyz.block.ftl.v1.ConfigUnsetRequest,
      xyz.block.ftl.v1.ConfigUnsetResponse> getConfigUnsetMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "ConfigUnset",
      requestType = xyz.block.ftl.v1.ConfigUnsetRequest.class,
      responseType = xyz.block.ftl.v1.ConfigUnsetResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<xyz.block.ftl.v1.ConfigUnsetRequest,
      xyz.block.ftl.v1.ConfigUnsetResponse> getConfigUnsetMethod() {
    io.grpc.MethodDescriptor<xyz.block.ftl.v1.ConfigUnsetRequest, xyz.block.ftl.v1.ConfigUnsetResponse> getConfigUnsetMethod;
    if ((getConfigUnsetMethod = AdminServiceGrpc.getConfigUnsetMethod) == null) {
      synchronized (AdminServiceGrpc.class) {
        if ((getConfigUnsetMethod = AdminServiceGrpc.getConfigUnsetMethod) == null) {
          AdminServiceGrpc.getConfigUnsetMethod = getConfigUnsetMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.v1.ConfigUnsetRequest, xyz.block.ftl.v1.ConfigUnsetResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "ConfigUnset"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.ConfigUnsetRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.ConfigUnsetResponse.getDefaultInstance()))
              .setSchemaDescriptor(new AdminServiceMethodDescriptorSupplier("ConfigUnset"))
              .build();
        }
      }
    }
    return getConfigUnsetMethod;
  }

  private static volatile io.grpc.MethodDescriptor<xyz.block.ftl.v1.SecretsListRequest,
      xyz.block.ftl.v1.SecretsListResponse> getSecretsListMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "SecretsList",
      requestType = xyz.block.ftl.v1.SecretsListRequest.class,
      responseType = xyz.block.ftl.v1.SecretsListResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<xyz.block.ftl.v1.SecretsListRequest,
      xyz.block.ftl.v1.SecretsListResponse> getSecretsListMethod() {
    io.grpc.MethodDescriptor<xyz.block.ftl.v1.SecretsListRequest, xyz.block.ftl.v1.SecretsListResponse> getSecretsListMethod;
    if ((getSecretsListMethod = AdminServiceGrpc.getSecretsListMethod) == null) {
      synchronized (AdminServiceGrpc.class) {
        if ((getSecretsListMethod = AdminServiceGrpc.getSecretsListMethod) == null) {
          AdminServiceGrpc.getSecretsListMethod = getSecretsListMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.v1.SecretsListRequest, xyz.block.ftl.v1.SecretsListResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "SecretsList"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.SecretsListRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.SecretsListResponse.getDefaultInstance()))
              .setSchemaDescriptor(new AdminServiceMethodDescriptorSupplier("SecretsList"))
              .build();
        }
      }
    }
    return getSecretsListMethod;
  }

  private static volatile io.grpc.MethodDescriptor<xyz.block.ftl.v1.SecretGetRequest,
      xyz.block.ftl.v1.SecretGetResponse> getSecretGetMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "SecretGet",
      requestType = xyz.block.ftl.v1.SecretGetRequest.class,
      responseType = xyz.block.ftl.v1.SecretGetResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<xyz.block.ftl.v1.SecretGetRequest,
      xyz.block.ftl.v1.SecretGetResponse> getSecretGetMethod() {
    io.grpc.MethodDescriptor<xyz.block.ftl.v1.SecretGetRequest, xyz.block.ftl.v1.SecretGetResponse> getSecretGetMethod;
    if ((getSecretGetMethod = AdminServiceGrpc.getSecretGetMethod) == null) {
      synchronized (AdminServiceGrpc.class) {
        if ((getSecretGetMethod = AdminServiceGrpc.getSecretGetMethod) == null) {
          AdminServiceGrpc.getSecretGetMethod = getSecretGetMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.v1.SecretGetRequest, xyz.block.ftl.v1.SecretGetResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "SecretGet"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.SecretGetRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.SecretGetResponse.getDefaultInstance()))
              .setSchemaDescriptor(new AdminServiceMethodDescriptorSupplier("SecretGet"))
              .build();
        }
      }
    }
    return getSecretGetMethod;
  }

  private static volatile io.grpc.MethodDescriptor<xyz.block.ftl.v1.SecretSetRequest,
      xyz.block.ftl.v1.SecretSetResponse> getSecretSetMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "SecretSet",
      requestType = xyz.block.ftl.v1.SecretSetRequest.class,
      responseType = xyz.block.ftl.v1.SecretSetResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<xyz.block.ftl.v1.SecretSetRequest,
      xyz.block.ftl.v1.SecretSetResponse> getSecretSetMethod() {
    io.grpc.MethodDescriptor<xyz.block.ftl.v1.SecretSetRequest, xyz.block.ftl.v1.SecretSetResponse> getSecretSetMethod;
    if ((getSecretSetMethod = AdminServiceGrpc.getSecretSetMethod) == null) {
      synchronized (AdminServiceGrpc.class) {
        if ((getSecretSetMethod = AdminServiceGrpc.getSecretSetMethod) == null) {
          AdminServiceGrpc.getSecretSetMethod = getSecretSetMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.v1.SecretSetRequest, xyz.block.ftl.v1.SecretSetResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "SecretSet"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.SecretSetRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.SecretSetResponse.getDefaultInstance()))
              .setSchemaDescriptor(new AdminServiceMethodDescriptorSupplier("SecretSet"))
              .build();
        }
      }
    }
    return getSecretSetMethod;
  }

  private static volatile io.grpc.MethodDescriptor<xyz.block.ftl.v1.SecretUnsetRequest,
      xyz.block.ftl.v1.SecretUnsetResponse> getSecretUnsetMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "SecretUnset",
      requestType = xyz.block.ftl.v1.SecretUnsetRequest.class,
      responseType = xyz.block.ftl.v1.SecretUnsetResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<xyz.block.ftl.v1.SecretUnsetRequest,
      xyz.block.ftl.v1.SecretUnsetResponse> getSecretUnsetMethod() {
    io.grpc.MethodDescriptor<xyz.block.ftl.v1.SecretUnsetRequest, xyz.block.ftl.v1.SecretUnsetResponse> getSecretUnsetMethod;
    if ((getSecretUnsetMethod = AdminServiceGrpc.getSecretUnsetMethod) == null) {
      synchronized (AdminServiceGrpc.class) {
        if ((getSecretUnsetMethod = AdminServiceGrpc.getSecretUnsetMethod) == null) {
          AdminServiceGrpc.getSecretUnsetMethod = getSecretUnsetMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.v1.SecretUnsetRequest, xyz.block.ftl.v1.SecretUnsetResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "SecretUnset"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.SecretUnsetRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.SecretUnsetResponse.getDefaultInstance()))
              .setSchemaDescriptor(new AdminServiceMethodDescriptorSupplier("SecretUnset"))
              .build();
        }
      }
    }
    return getSecretUnsetMethod;
  }

  /**
   * Creates a new async stub that supports all call types for the service
   */
  public static AdminServiceStub newStub(io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<AdminServiceStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<AdminServiceStub>() {
        @java.lang.Override
        public AdminServiceStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new AdminServiceStub(channel, callOptions);
        }
      };
    return AdminServiceStub.newStub(factory, channel);
  }

  /**
   * Creates a new blocking-style stub that supports unary and streaming output calls on the service
   */
  public static AdminServiceBlockingStub newBlockingStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<AdminServiceBlockingStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<AdminServiceBlockingStub>() {
        @java.lang.Override
        public AdminServiceBlockingStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new AdminServiceBlockingStub(channel, callOptions);
        }
      };
    return AdminServiceBlockingStub.newStub(factory, channel);
  }

  /**
   * Creates a new ListenableFuture-style stub that supports unary calls on the service
   */
  public static AdminServiceFutureStub newFutureStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<AdminServiceFutureStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<AdminServiceFutureStub>() {
        @java.lang.Override
        public AdminServiceFutureStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new AdminServiceFutureStub(channel, callOptions);
        }
      };
    return AdminServiceFutureStub.newStub(factory, channel);
  }

  /**
   * <pre>
   * AdminService is the service that provides and updates admin data. For example,
   * it is used to encapsulate configuration and secrets.
   * </pre>
   */
  public interface AsyncService {

    /**
     */
    default void ping(xyz.block.ftl.v1.PingRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.PingResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getPingMethod(), responseObserver);
    }

    /**
     * <pre>
     * List configuration.
     * </pre>
     */
    default void configList(xyz.block.ftl.v1.ConfigListRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.ConfigListResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getConfigListMethod(), responseObserver);
    }

    /**
     * <pre>
     * Get a config value.
     * </pre>
     */
    default void configGet(xyz.block.ftl.v1.ConfigGetRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.ConfigGetResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getConfigGetMethod(), responseObserver);
    }

    /**
     * <pre>
     * Set a config value.
     * </pre>
     */
    default void configSet(xyz.block.ftl.v1.ConfigSetRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.ConfigSetResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getConfigSetMethod(), responseObserver);
    }

    /**
     * <pre>
     * Unset a config value.
     * </pre>
     */
    default void configUnset(xyz.block.ftl.v1.ConfigUnsetRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.ConfigUnsetResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getConfigUnsetMethod(), responseObserver);
    }

    /**
     * <pre>
     * List secrets.
     * </pre>
     */
    default void secretsList(xyz.block.ftl.v1.SecretsListRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.SecretsListResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getSecretsListMethod(), responseObserver);
    }

    /**
     * <pre>
     * Get a secret.
     * </pre>
     */
    default void secretGet(xyz.block.ftl.v1.SecretGetRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.SecretGetResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getSecretGetMethod(), responseObserver);
    }

    /**
     * <pre>
     * Set a secret.
     * </pre>
     */
    default void secretSet(xyz.block.ftl.v1.SecretSetRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.SecretSetResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getSecretSetMethod(), responseObserver);
    }

    /**
     * <pre>
     * Unset a secret.
     * </pre>
     */
    default void secretUnset(xyz.block.ftl.v1.SecretUnsetRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.SecretUnsetResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getSecretUnsetMethod(), responseObserver);
    }
  }

  /**
   * Base class for the server implementation of the service AdminService.
   * <pre>
   * AdminService is the service that provides and updates admin data. For example,
   * it is used to encapsulate configuration and secrets.
   * </pre>
   */
  public static abstract class AdminServiceImplBase
      implements io.grpc.BindableService, AsyncService {

    @java.lang.Override public final io.grpc.ServerServiceDefinition bindService() {
      return AdminServiceGrpc.bindService(this);
    }
  }

  /**
   * A stub to allow clients to do asynchronous rpc calls to service AdminService.
   * <pre>
   * AdminService is the service that provides and updates admin data. For example,
   * it is used to encapsulate configuration and secrets.
   * </pre>
   */
  public static final class AdminServiceStub
      extends io.grpc.stub.AbstractAsyncStub<AdminServiceStub> {
    private AdminServiceStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected AdminServiceStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new AdminServiceStub(channel, callOptions);
    }

    /**
     */
    public void ping(xyz.block.ftl.v1.PingRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.PingResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getPingMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * List configuration.
     * </pre>
     */
    public void configList(xyz.block.ftl.v1.ConfigListRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.ConfigListResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getConfigListMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * Get a config value.
     * </pre>
     */
    public void configGet(xyz.block.ftl.v1.ConfigGetRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.ConfigGetResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getConfigGetMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * Set a config value.
     * </pre>
     */
    public void configSet(xyz.block.ftl.v1.ConfigSetRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.ConfigSetResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getConfigSetMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * Unset a config value.
     * </pre>
     */
    public void configUnset(xyz.block.ftl.v1.ConfigUnsetRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.ConfigUnsetResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getConfigUnsetMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * List secrets.
     * </pre>
     */
    public void secretsList(xyz.block.ftl.v1.SecretsListRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.SecretsListResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getSecretsListMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * Get a secret.
     * </pre>
     */
    public void secretGet(xyz.block.ftl.v1.SecretGetRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.SecretGetResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getSecretGetMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * Set a secret.
     * </pre>
     */
    public void secretSet(xyz.block.ftl.v1.SecretSetRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.SecretSetResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getSecretSetMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * Unset a secret.
     * </pre>
     */
    public void secretUnset(xyz.block.ftl.v1.SecretUnsetRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.v1.SecretUnsetResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getSecretUnsetMethod(), getCallOptions()), request, responseObserver);
    }
  }

  /**
   * A stub to allow clients to do synchronous rpc calls to service AdminService.
   * <pre>
   * AdminService is the service that provides and updates admin data. For example,
   * it is used to encapsulate configuration and secrets.
   * </pre>
   */
  public static final class AdminServiceBlockingStub
      extends io.grpc.stub.AbstractBlockingStub<AdminServiceBlockingStub> {
    private AdminServiceBlockingStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected AdminServiceBlockingStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new AdminServiceBlockingStub(channel, callOptions);
    }

    /**
     */
    public xyz.block.ftl.v1.PingResponse ping(xyz.block.ftl.v1.PingRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getPingMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * List configuration.
     * </pre>
     */
    public xyz.block.ftl.v1.ConfigListResponse configList(xyz.block.ftl.v1.ConfigListRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getConfigListMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Get a config value.
     * </pre>
     */
    public xyz.block.ftl.v1.ConfigGetResponse configGet(xyz.block.ftl.v1.ConfigGetRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getConfigGetMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Set a config value.
     * </pre>
     */
    public xyz.block.ftl.v1.ConfigSetResponse configSet(xyz.block.ftl.v1.ConfigSetRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getConfigSetMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Unset a config value.
     * </pre>
     */
    public xyz.block.ftl.v1.ConfigUnsetResponse configUnset(xyz.block.ftl.v1.ConfigUnsetRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getConfigUnsetMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * List secrets.
     * </pre>
     */
    public xyz.block.ftl.v1.SecretsListResponse secretsList(xyz.block.ftl.v1.SecretsListRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getSecretsListMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Get a secret.
     * </pre>
     */
    public xyz.block.ftl.v1.SecretGetResponse secretGet(xyz.block.ftl.v1.SecretGetRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getSecretGetMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Set a secret.
     * </pre>
     */
    public xyz.block.ftl.v1.SecretSetResponse secretSet(xyz.block.ftl.v1.SecretSetRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getSecretSetMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Unset a secret.
     * </pre>
     */
    public xyz.block.ftl.v1.SecretUnsetResponse secretUnset(xyz.block.ftl.v1.SecretUnsetRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getSecretUnsetMethod(), getCallOptions(), request);
    }
  }

  /**
   * A stub to allow clients to do ListenableFuture-style rpc calls to service AdminService.
   * <pre>
   * AdminService is the service that provides and updates admin data. For example,
   * it is used to encapsulate configuration and secrets.
   * </pre>
   */
  public static final class AdminServiceFutureStub
      extends io.grpc.stub.AbstractFutureStub<AdminServiceFutureStub> {
    private AdminServiceFutureStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected AdminServiceFutureStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new AdminServiceFutureStub(channel, callOptions);
    }

    /**
     */
    public com.google.common.util.concurrent.ListenableFuture<xyz.block.ftl.v1.PingResponse> ping(
        xyz.block.ftl.v1.PingRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getPingMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * List configuration.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<xyz.block.ftl.v1.ConfigListResponse> configList(
        xyz.block.ftl.v1.ConfigListRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getConfigListMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * Get a config value.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<xyz.block.ftl.v1.ConfigGetResponse> configGet(
        xyz.block.ftl.v1.ConfigGetRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getConfigGetMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * Set a config value.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<xyz.block.ftl.v1.ConfigSetResponse> configSet(
        xyz.block.ftl.v1.ConfigSetRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getConfigSetMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * Unset a config value.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<xyz.block.ftl.v1.ConfigUnsetResponse> configUnset(
        xyz.block.ftl.v1.ConfigUnsetRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getConfigUnsetMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * List secrets.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<xyz.block.ftl.v1.SecretsListResponse> secretsList(
        xyz.block.ftl.v1.SecretsListRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getSecretsListMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * Get a secret.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<xyz.block.ftl.v1.SecretGetResponse> secretGet(
        xyz.block.ftl.v1.SecretGetRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getSecretGetMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * Set a secret.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<xyz.block.ftl.v1.SecretSetResponse> secretSet(
        xyz.block.ftl.v1.SecretSetRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getSecretSetMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * Unset a secret.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<xyz.block.ftl.v1.SecretUnsetResponse> secretUnset(
        xyz.block.ftl.v1.SecretUnsetRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getSecretUnsetMethod(), getCallOptions()), request);
    }
  }

  private static final int METHODID_PING = 0;
  private static final int METHODID_CONFIG_LIST = 1;
  private static final int METHODID_CONFIG_GET = 2;
  private static final int METHODID_CONFIG_SET = 3;
  private static final int METHODID_CONFIG_UNSET = 4;
  private static final int METHODID_SECRETS_LIST = 5;
  private static final int METHODID_SECRET_GET = 6;
  private static final int METHODID_SECRET_SET = 7;
  private static final int METHODID_SECRET_UNSET = 8;

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
        case METHODID_CONFIG_LIST:
          serviceImpl.configList((xyz.block.ftl.v1.ConfigListRequest) request,
              (io.grpc.stub.StreamObserver<xyz.block.ftl.v1.ConfigListResponse>) responseObserver);
          break;
        case METHODID_CONFIG_GET:
          serviceImpl.configGet((xyz.block.ftl.v1.ConfigGetRequest) request,
              (io.grpc.stub.StreamObserver<xyz.block.ftl.v1.ConfigGetResponse>) responseObserver);
          break;
        case METHODID_CONFIG_SET:
          serviceImpl.configSet((xyz.block.ftl.v1.ConfigSetRequest) request,
              (io.grpc.stub.StreamObserver<xyz.block.ftl.v1.ConfigSetResponse>) responseObserver);
          break;
        case METHODID_CONFIG_UNSET:
          serviceImpl.configUnset((xyz.block.ftl.v1.ConfigUnsetRequest) request,
              (io.grpc.stub.StreamObserver<xyz.block.ftl.v1.ConfigUnsetResponse>) responseObserver);
          break;
        case METHODID_SECRETS_LIST:
          serviceImpl.secretsList((xyz.block.ftl.v1.SecretsListRequest) request,
              (io.grpc.stub.StreamObserver<xyz.block.ftl.v1.SecretsListResponse>) responseObserver);
          break;
        case METHODID_SECRET_GET:
          serviceImpl.secretGet((xyz.block.ftl.v1.SecretGetRequest) request,
              (io.grpc.stub.StreamObserver<xyz.block.ftl.v1.SecretGetResponse>) responseObserver);
          break;
        case METHODID_SECRET_SET:
          serviceImpl.secretSet((xyz.block.ftl.v1.SecretSetRequest) request,
              (io.grpc.stub.StreamObserver<xyz.block.ftl.v1.SecretSetResponse>) responseObserver);
          break;
        case METHODID_SECRET_UNSET:
          serviceImpl.secretUnset((xyz.block.ftl.v1.SecretUnsetRequest) request,
              (io.grpc.stub.StreamObserver<xyz.block.ftl.v1.SecretUnsetResponse>) responseObserver);
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
          getConfigListMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              xyz.block.ftl.v1.ConfigListRequest,
              xyz.block.ftl.v1.ConfigListResponse>(
                service, METHODID_CONFIG_LIST)))
        .addMethod(
          getConfigGetMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              xyz.block.ftl.v1.ConfigGetRequest,
              xyz.block.ftl.v1.ConfigGetResponse>(
                service, METHODID_CONFIG_GET)))
        .addMethod(
          getConfigSetMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              xyz.block.ftl.v1.ConfigSetRequest,
              xyz.block.ftl.v1.ConfigSetResponse>(
                service, METHODID_CONFIG_SET)))
        .addMethod(
          getConfigUnsetMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              xyz.block.ftl.v1.ConfigUnsetRequest,
              xyz.block.ftl.v1.ConfigUnsetResponse>(
                service, METHODID_CONFIG_UNSET)))
        .addMethod(
          getSecretsListMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              xyz.block.ftl.v1.SecretsListRequest,
              xyz.block.ftl.v1.SecretsListResponse>(
                service, METHODID_SECRETS_LIST)))
        .addMethod(
          getSecretGetMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              xyz.block.ftl.v1.SecretGetRequest,
              xyz.block.ftl.v1.SecretGetResponse>(
                service, METHODID_SECRET_GET)))
        .addMethod(
          getSecretSetMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              xyz.block.ftl.v1.SecretSetRequest,
              xyz.block.ftl.v1.SecretSetResponse>(
                service, METHODID_SECRET_SET)))
        .addMethod(
          getSecretUnsetMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              xyz.block.ftl.v1.SecretUnsetRequest,
              xyz.block.ftl.v1.SecretUnsetResponse>(
                service, METHODID_SECRET_UNSET)))
        .build();
  }

  private static abstract class AdminServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoFileDescriptorSupplier, io.grpc.protobuf.ProtoServiceDescriptorSupplier {
    AdminServiceBaseDescriptorSupplier() {}

    @java.lang.Override
    public com.google.protobuf.Descriptors.FileDescriptor getFileDescriptor() {
      return xyz.block.ftl.v1.Admin.getDescriptor();
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.ServiceDescriptor getServiceDescriptor() {
      return getFileDescriptor().findServiceByName("AdminService");
    }
  }

  private static final class AdminServiceFileDescriptorSupplier
      extends AdminServiceBaseDescriptorSupplier {
    AdminServiceFileDescriptorSupplier() {}
  }

  private static final class AdminServiceMethodDescriptorSupplier
      extends AdminServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoMethodDescriptorSupplier {
    private final java.lang.String methodName;

    AdminServiceMethodDescriptorSupplier(java.lang.String methodName) {
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
      synchronized (AdminServiceGrpc.class) {
        result = serviceDescriptor;
        if (result == null) {
          serviceDescriptor = result = io.grpc.ServiceDescriptor.newBuilder(SERVICE_NAME)
              .setSchemaDescriptor(new AdminServiceFileDescriptorSupplier())
              .addMethod(getPingMethod())
              .addMethod(getConfigListMethod())
              .addMethod(getConfigGetMethod())
              .addMethod(getConfigSetMethod())
              .addMethod(getConfigUnsetMethod())
              .addMethod(getSecretsListMethod())
              .addMethod(getSecretGetMethod())
              .addMethod(getSecretSetMethod())
              .addMethod(getSecretUnsetMethod())
              .build();
        }
      }
    }
    return result;
  }
}
