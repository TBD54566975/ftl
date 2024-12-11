package xyz.block.ftl.language.v1;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 * <pre>
 * LanguageService allows a plugin to add support for a programming language.
 * </pre>
 */
@javax.annotation.Generated(
    value = "by gRPC proto compiler (version 1.68.2)",
    comments = "Source: xyz/block/ftl/language/v1/language.proto")
@io.grpc.stub.annotations.GrpcGenerated
public final class LanguageServiceGrpc {

  private LanguageServiceGrpc() {}

  public static final java.lang.String SERVICE_NAME = "xyz.block.ftl.language.v1.LanguageService";

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
    if ((getPingMethod = LanguageServiceGrpc.getPingMethod) == null) {
      synchronized (LanguageServiceGrpc.class) {
        if ((getPingMethod = LanguageServiceGrpc.getPingMethod) == null) {
          LanguageServiceGrpc.getPingMethod = getPingMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.v1.PingRequest, xyz.block.ftl.v1.PingResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "Ping"))
              .setSafe(true)
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.PingRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.v1.PingResponse.getDefaultInstance()))
              .setSchemaDescriptor(new LanguageServiceMethodDescriptorSupplier("Ping"))
              .build();
        }
      }
    }
    return getPingMethod;
  }

  private static volatile io.grpc.MethodDescriptor<xyz.block.ftl.language.v1.GetCreateModuleFlagsRequest,
      xyz.block.ftl.language.v1.GetCreateModuleFlagsResponse> getGetCreateModuleFlagsMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "GetCreateModuleFlags",
      requestType = xyz.block.ftl.language.v1.GetCreateModuleFlagsRequest.class,
      responseType = xyz.block.ftl.language.v1.GetCreateModuleFlagsResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<xyz.block.ftl.language.v1.GetCreateModuleFlagsRequest,
      xyz.block.ftl.language.v1.GetCreateModuleFlagsResponse> getGetCreateModuleFlagsMethod() {
    io.grpc.MethodDescriptor<xyz.block.ftl.language.v1.GetCreateModuleFlagsRequest, xyz.block.ftl.language.v1.GetCreateModuleFlagsResponse> getGetCreateModuleFlagsMethod;
    if ((getGetCreateModuleFlagsMethod = LanguageServiceGrpc.getGetCreateModuleFlagsMethod) == null) {
      synchronized (LanguageServiceGrpc.class) {
        if ((getGetCreateModuleFlagsMethod = LanguageServiceGrpc.getGetCreateModuleFlagsMethod) == null) {
          LanguageServiceGrpc.getGetCreateModuleFlagsMethod = getGetCreateModuleFlagsMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.language.v1.GetCreateModuleFlagsRequest, xyz.block.ftl.language.v1.GetCreateModuleFlagsResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "GetCreateModuleFlags"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.language.v1.GetCreateModuleFlagsRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.language.v1.GetCreateModuleFlagsResponse.getDefaultInstance()))
              .setSchemaDescriptor(new LanguageServiceMethodDescriptorSupplier("GetCreateModuleFlags"))
              .build();
        }
      }
    }
    return getGetCreateModuleFlagsMethod;
  }

  private static volatile io.grpc.MethodDescriptor<xyz.block.ftl.language.v1.CreateModuleRequest,
      xyz.block.ftl.language.v1.CreateModuleResponse> getCreateModuleMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "CreateModule",
      requestType = xyz.block.ftl.language.v1.CreateModuleRequest.class,
      responseType = xyz.block.ftl.language.v1.CreateModuleResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<xyz.block.ftl.language.v1.CreateModuleRequest,
      xyz.block.ftl.language.v1.CreateModuleResponse> getCreateModuleMethod() {
    io.grpc.MethodDescriptor<xyz.block.ftl.language.v1.CreateModuleRequest, xyz.block.ftl.language.v1.CreateModuleResponse> getCreateModuleMethod;
    if ((getCreateModuleMethod = LanguageServiceGrpc.getCreateModuleMethod) == null) {
      synchronized (LanguageServiceGrpc.class) {
        if ((getCreateModuleMethod = LanguageServiceGrpc.getCreateModuleMethod) == null) {
          LanguageServiceGrpc.getCreateModuleMethod = getCreateModuleMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.language.v1.CreateModuleRequest, xyz.block.ftl.language.v1.CreateModuleResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "CreateModule"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.language.v1.CreateModuleRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.language.v1.CreateModuleResponse.getDefaultInstance()))
              .setSchemaDescriptor(new LanguageServiceMethodDescriptorSupplier("CreateModule"))
              .build();
        }
      }
    }
    return getCreateModuleMethod;
  }

  private static volatile io.grpc.MethodDescriptor<xyz.block.ftl.language.v1.ModuleConfigDefaultsRequest,
      xyz.block.ftl.language.v1.ModuleConfigDefaultsResponse> getModuleConfigDefaultsMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "ModuleConfigDefaults",
      requestType = xyz.block.ftl.language.v1.ModuleConfigDefaultsRequest.class,
      responseType = xyz.block.ftl.language.v1.ModuleConfigDefaultsResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<xyz.block.ftl.language.v1.ModuleConfigDefaultsRequest,
      xyz.block.ftl.language.v1.ModuleConfigDefaultsResponse> getModuleConfigDefaultsMethod() {
    io.grpc.MethodDescriptor<xyz.block.ftl.language.v1.ModuleConfigDefaultsRequest, xyz.block.ftl.language.v1.ModuleConfigDefaultsResponse> getModuleConfigDefaultsMethod;
    if ((getModuleConfigDefaultsMethod = LanguageServiceGrpc.getModuleConfigDefaultsMethod) == null) {
      synchronized (LanguageServiceGrpc.class) {
        if ((getModuleConfigDefaultsMethod = LanguageServiceGrpc.getModuleConfigDefaultsMethod) == null) {
          LanguageServiceGrpc.getModuleConfigDefaultsMethod = getModuleConfigDefaultsMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.language.v1.ModuleConfigDefaultsRequest, xyz.block.ftl.language.v1.ModuleConfigDefaultsResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "ModuleConfigDefaults"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.language.v1.ModuleConfigDefaultsRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.language.v1.ModuleConfigDefaultsResponse.getDefaultInstance()))
              .setSchemaDescriptor(new LanguageServiceMethodDescriptorSupplier("ModuleConfigDefaults"))
              .build();
        }
      }
    }
    return getModuleConfigDefaultsMethod;
  }

  private static volatile io.grpc.MethodDescriptor<xyz.block.ftl.language.v1.GetDependenciesRequest,
      xyz.block.ftl.language.v1.GetDependenciesResponse> getGetDependenciesMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "GetDependencies",
      requestType = xyz.block.ftl.language.v1.GetDependenciesRequest.class,
      responseType = xyz.block.ftl.language.v1.GetDependenciesResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<xyz.block.ftl.language.v1.GetDependenciesRequest,
      xyz.block.ftl.language.v1.GetDependenciesResponse> getGetDependenciesMethod() {
    io.grpc.MethodDescriptor<xyz.block.ftl.language.v1.GetDependenciesRequest, xyz.block.ftl.language.v1.GetDependenciesResponse> getGetDependenciesMethod;
    if ((getGetDependenciesMethod = LanguageServiceGrpc.getGetDependenciesMethod) == null) {
      synchronized (LanguageServiceGrpc.class) {
        if ((getGetDependenciesMethod = LanguageServiceGrpc.getGetDependenciesMethod) == null) {
          LanguageServiceGrpc.getGetDependenciesMethod = getGetDependenciesMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.language.v1.GetDependenciesRequest, xyz.block.ftl.language.v1.GetDependenciesResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "GetDependencies"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.language.v1.GetDependenciesRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.language.v1.GetDependenciesResponse.getDefaultInstance()))
              .setSchemaDescriptor(new LanguageServiceMethodDescriptorSupplier("GetDependencies"))
              .build();
        }
      }
    }
    return getGetDependenciesMethod;
  }

  private static volatile io.grpc.MethodDescriptor<xyz.block.ftl.language.v1.BuildRequest,
      xyz.block.ftl.language.v1.BuildResponse> getBuildMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "Build",
      requestType = xyz.block.ftl.language.v1.BuildRequest.class,
      responseType = xyz.block.ftl.language.v1.BuildResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.SERVER_STREAMING)
  public static io.grpc.MethodDescriptor<xyz.block.ftl.language.v1.BuildRequest,
      xyz.block.ftl.language.v1.BuildResponse> getBuildMethod() {
    io.grpc.MethodDescriptor<xyz.block.ftl.language.v1.BuildRequest, xyz.block.ftl.language.v1.BuildResponse> getBuildMethod;
    if ((getBuildMethod = LanguageServiceGrpc.getBuildMethod) == null) {
      synchronized (LanguageServiceGrpc.class) {
        if ((getBuildMethod = LanguageServiceGrpc.getBuildMethod) == null) {
          LanguageServiceGrpc.getBuildMethod = getBuildMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.language.v1.BuildRequest, xyz.block.ftl.language.v1.BuildResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.SERVER_STREAMING)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "Build"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.language.v1.BuildRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.language.v1.BuildResponse.getDefaultInstance()))
              .setSchemaDescriptor(new LanguageServiceMethodDescriptorSupplier("Build"))
              .build();
        }
      }
    }
    return getBuildMethod;
  }

  private static volatile io.grpc.MethodDescriptor<xyz.block.ftl.language.v1.BuildContextUpdatedRequest,
      xyz.block.ftl.language.v1.BuildContextUpdatedResponse> getBuildContextUpdatedMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "BuildContextUpdated",
      requestType = xyz.block.ftl.language.v1.BuildContextUpdatedRequest.class,
      responseType = xyz.block.ftl.language.v1.BuildContextUpdatedResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<xyz.block.ftl.language.v1.BuildContextUpdatedRequest,
      xyz.block.ftl.language.v1.BuildContextUpdatedResponse> getBuildContextUpdatedMethod() {
    io.grpc.MethodDescriptor<xyz.block.ftl.language.v1.BuildContextUpdatedRequest, xyz.block.ftl.language.v1.BuildContextUpdatedResponse> getBuildContextUpdatedMethod;
    if ((getBuildContextUpdatedMethod = LanguageServiceGrpc.getBuildContextUpdatedMethod) == null) {
      synchronized (LanguageServiceGrpc.class) {
        if ((getBuildContextUpdatedMethod = LanguageServiceGrpc.getBuildContextUpdatedMethod) == null) {
          LanguageServiceGrpc.getBuildContextUpdatedMethod = getBuildContextUpdatedMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.language.v1.BuildContextUpdatedRequest, xyz.block.ftl.language.v1.BuildContextUpdatedResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "BuildContextUpdated"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.language.v1.BuildContextUpdatedRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.language.v1.BuildContextUpdatedResponse.getDefaultInstance()))
              .setSchemaDescriptor(new LanguageServiceMethodDescriptorSupplier("BuildContextUpdated"))
              .build();
        }
      }
    }
    return getBuildContextUpdatedMethod;
  }

  private static volatile io.grpc.MethodDescriptor<xyz.block.ftl.language.v1.GenerateStubsRequest,
      xyz.block.ftl.language.v1.GenerateStubsResponse> getGenerateStubsMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "GenerateStubs",
      requestType = xyz.block.ftl.language.v1.GenerateStubsRequest.class,
      responseType = xyz.block.ftl.language.v1.GenerateStubsResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<xyz.block.ftl.language.v1.GenerateStubsRequest,
      xyz.block.ftl.language.v1.GenerateStubsResponse> getGenerateStubsMethod() {
    io.grpc.MethodDescriptor<xyz.block.ftl.language.v1.GenerateStubsRequest, xyz.block.ftl.language.v1.GenerateStubsResponse> getGenerateStubsMethod;
    if ((getGenerateStubsMethod = LanguageServiceGrpc.getGenerateStubsMethod) == null) {
      synchronized (LanguageServiceGrpc.class) {
        if ((getGenerateStubsMethod = LanguageServiceGrpc.getGenerateStubsMethod) == null) {
          LanguageServiceGrpc.getGenerateStubsMethod = getGenerateStubsMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.language.v1.GenerateStubsRequest, xyz.block.ftl.language.v1.GenerateStubsResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "GenerateStubs"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.language.v1.GenerateStubsRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.language.v1.GenerateStubsResponse.getDefaultInstance()))
              .setSchemaDescriptor(new LanguageServiceMethodDescriptorSupplier("GenerateStubs"))
              .build();
        }
      }
    }
    return getGenerateStubsMethod;
  }

  private static volatile io.grpc.MethodDescriptor<xyz.block.ftl.language.v1.SyncStubReferencesRequest,
      xyz.block.ftl.language.v1.SyncStubReferencesResponse> getSyncStubReferencesMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "SyncStubReferences",
      requestType = xyz.block.ftl.language.v1.SyncStubReferencesRequest.class,
      responseType = xyz.block.ftl.language.v1.SyncStubReferencesResponse.class,
      methodType = io.grpc.MethodDescriptor.MethodType.UNARY)
  public static io.grpc.MethodDescriptor<xyz.block.ftl.language.v1.SyncStubReferencesRequest,
      xyz.block.ftl.language.v1.SyncStubReferencesResponse> getSyncStubReferencesMethod() {
    io.grpc.MethodDescriptor<xyz.block.ftl.language.v1.SyncStubReferencesRequest, xyz.block.ftl.language.v1.SyncStubReferencesResponse> getSyncStubReferencesMethod;
    if ((getSyncStubReferencesMethod = LanguageServiceGrpc.getSyncStubReferencesMethod) == null) {
      synchronized (LanguageServiceGrpc.class) {
        if ((getSyncStubReferencesMethod = LanguageServiceGrpc.getSyncStubReferencesMethod) == null) {
          LanguageServiceGrpc.getSyncStubReferencesMethod = getSyncStubReferencesMethod =
              io.grpc.MethodDescriptor.<xyz.block.ftl.language.v1.SyncStubReferencesRequest, xyz.block.ftl.language.v1.SyncStubReferencesResponse>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.UNARY)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "SyncStubReferences"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.language.v1.SyncStubReferencesRequest.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  xyz.block.ftl.language.v1.SyncStubReferencesResponse.getDefaultInstance()))
              .setSchemaDescriptor(new LanguageServiceMethodDescriptorSupplier("SyncStubReferences"))
              .build();
        }
      }
    }
    return getSyncStubReferencesMethod;
  }

  /**
   * Creates a new async stub that supports all call types for the service
   */
  public static LanguageServiceStub newStub(io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<LanguageServiceStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<LanguageServiceStub>() {
        @java.lang.Override
        public LanguageServiceStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new LanguageServiceStub(channel, callOptions);
        }
      };
    return LanguageServiceStub.newStub(factory, channel);
  }

  /**
   * Creates a new blocking-style stub that supports unary and streaming output calls on the service
   */
  public static LanguageServiceBlockingStub newBlockingStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<LanguageServiceBlockingStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<LanguageServiceBlockingStub>() {
        @java.lang.Override
        public LanguageServiceBlockingStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new LanguageServiceBlockingStub(channel, callOptions);
        }
      };
    return LanguageServiceBlockingStub.newStub(factory, channel);
  }

  /**
   * Creates a new ListenableFuture-style stub that supports unary calls on the service
   */
  public static LanguageServiceFutureStub newFutureStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<LanguageServiceFutureStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<LanguageServiceFutureStub>() {
        @java.lang.Override
        public LanguageServiceFutureStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new LanguageServiceFutureStub(channel, callOptions);
        }
      };
    return LanguageServiceFutureStub.newStub(factory, channel);
  }

  /**
   * <pre>
   * LanguageService allows a plugin to add support for a programming language.
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
     * Get language specific flags that can be used to create a new module.
     * </pre>
     */
    default void getCreateModuleFlags(xyz.block.ftl.language.v1.GetCreateModuleFlagsRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.language.v1.GetCreateModuleFlagsResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getGetCreateModuleFlagsMethod(), responseObserver);
    }

    /**
     * <pre>
     * Generates files for a new module with the requested name
     * </pre>
     */
    default void createModule(xyz.block.ftl.language.v1.CreateModuleRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.language.v1.CreateModuleResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getCreateModuleMethod(), responseObserver);
    }

    /**
     * <pre>
     * Provide default values for ModuleConfig for values that are not configured in the ftl.toml file.
     * </pre>
     */
    default void moduleConfigDefaults(xyz.block.ftl.language.v1.ModuleConfigDefaultsRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.language.v1.ModuleConfigDefaultsResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getModuleConfigDefaultsMethod(), responseObserver);
    }

    /**
     * <pre>
     * Extract dependencies for a module
     * FTL will ensure that these dependencies are built before requesting a build for this module.
     * </pre>
     */
    default void getDependencies(xyz.block.ftl.language.v1.GetDependenciesRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.language.v1.GetDependenciesResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getGetDependenciesMethod(), responseObserver);
    }

    /**
     * <pre>
     * Build the module and stream back build events.
     * A BuildSuccess or BuildFailure event must be streamed back with the request's context id to indicate the
     * end of the build.
     * The request can include the option to "rebuild_automatically". In this case the plugin should watch for
     * file changes and automatically rebuild as needed as long as this build request is alive. Each automactic
     * rebuild must include the latest build context id provided by the request or subsequent BuildContextUpdated
     * calls.
     * </pre>
     */
    default void build(xyz.block.ftl.language.v1.BuildRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.language.v1.BuildResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getBuildMethod(), responseObserver);
    }

    /**
     * <pre>
     * While a Build call with "rebuild_automatically" set is active, BuildContextUpdated is called whenever the
     * build context is updated.
     * Each time this call is made, the Build call must send back a corresponding BuildSuccess or BuildFailure
     * event with the updated build context id with "is_automatic_rebuild" as false.
     * If the plugin will not be able to return a BuildSuccess or BuildFailure, such as when there is no active
     * build stream, it must fail the BuildContextUpdated call.
     * </pre>
     */
    default void buildContextUpdated(xyz.block.ftl.language.v1.BuildContextUpdatedRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.language.v1.BuildContextUpdatedResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getBuildContextUpdatedMethod(), responseObserver);
    }

    /**
     * <pre>
     * Generate stubs for a module.
     * Stubs allow modules to import other module's exported interface. If a language does not need this step,
     * then it is not required to do anything in this call.
     * This call is not tied to the module that this plugin is responsible for. A plugin of each language will
     * be chosen to generate stubs for each module.
     * </pre>
     */
    default void generateStubs(xyz.block.ftl.language.v1.GenerateStubsRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.language.v1.GenerateStubsResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getGenerateStubsMethod(), responseObserver);
    }

    /**
     * <pre>
     * SyncStubReferences is called when module stubs have been updated. This allows the plugin to update
     * references to external modules, regardless of whether they are dependencies.
     * For example, go plugin adds references to all modules into the go.work file so that tools can automatically
     * import the modules when users start reference them.
     * It is optional to do anything with this call.
     * </pre>
     */
    default void syncStubReferences(xyz.block.ftl.language.v1.SyncStubReferencesRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.language.v1.SyncStubReferencesResponse> responseObserver) {
      io.grpc.stub.ServerCalls.asyncUnimplementedUnaryCall(getSyncStubReferencesMethod(), responseObserver);
    }
  }

  /**
   * Base class for the server implementation of the service LanguageService.
   * <pre>
   * LanguageService allows a plugin to add support for a programming language.
   * </pre>
   */
  public static abstract class LanguageServiceImplBase
      implements io.grpc.BindableService, AsyncService {

    @java.lang.Override public final io.grpc.ServerServiceDefinition bindService() {
      return LanguageServiceGrpc.bindService(this);
    }
  }

  /**
   * A stub to allow clients to do asynchronous rpc calls to service LanguageService.
   * <pre>
   * LanguageService allows a plugin to add support for a programming language.
   * </pre>
   */
  public static final class LanguageServiceStub
      extends io.grpc.stub.AbstractAsyncStub<LanguageServiceStub> {
    private LanguageServiceStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected LanguageServiceStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new LanguageServiceStub(channel, callOptions);
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
     * Get language specific flags that can be used to create a new module.
     * </pre>
     */
    public void getCreateModuleFlags(xyz.block.ftl.language.v1.GetCreateModuleFlagsRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.language.v1.GetCreateModuleFlagsResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getGetCreateModuleFlagsMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * Generates files for a new module with the requested name
     * </pre>
     */
    public void createModule(xyz.block.ftl.language.v1.CreateModuleRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.language.v1.CreateModuleResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getCreateModuleMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * Provide default values for ModuleConfig for values that are not configured in the ftl.toml file.
     * </pre>
     */
    public void moduleConfigDefaults(xyz.block.ftl.language.v1.ModuleConfigDefaultsRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.language.v1.ModuleConfigDefaultsResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getModuleConfigDefaultsMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * Extract dependencies for a module
     * FTL will ensure that these dependencies are built before requesting a build for this module.
     * </pre>
     */
    public void getDependencies(xyz.block.ftl.language.v1.GetDependenciesRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.language.v1.GetDependenciesResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getGetDependenciesMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * Build the module and stream back build events.
     * A BuildSuccess or BuildFailure event must be streamed back with the request's context id to indicate the
     * end of the build.
     * The request can include the option to "rebuild_automatically". In this case the plugin should watch for
     * file changes and automatically rebuild as needed as long as this build request is alive. Each automactic
     * rebuild must include the latest build context id provided by the request or subsequent BuildContextUpdated
     * calls.
     * </pre>
     */
    public void build(xyz.block.ftl.language.v1.BuildRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.language.v1.BuildResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncServerStreamingCall(
          getChannel().newCall(getBuildMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * While a Build call with "rebuild_automatically" set is active, BuildContextUpdated is called whenever the
     * build context is updated.
     * Each time this call is made, the Build call must send back a corresponding BuildSuccess or BuildFailure
     * event with the updated build context id with "is_automatic_rebuild" as false.
     * If the plugin will not be able to return a BuildSuccess or BuildFailure, such as when there is no active
     * build stream, it must fail the BuildContextUpdated call.
     * </pre>
     */
    public void buildContextUpdated(xyz.block.ftl.language.v1.BuildContextUpdatedRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.language.v1.BuildContextUpdatedResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getBuildContextUpdatedMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * Generate stubs for a module.
     * Stubs allow modules to import other module's exported interface. If a language does not need this step,
     * then it is not required to do anything in this call.
     * This call is not tied to the module that this plugin is responsible for. A plugin of each language will
     * be chosen to generate stubs for each module.
     * </pre>
     */
    public void generateStubs(xyz.block.ftl.language.v1.GenerateStubsRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.language.v1.GenerateStubsResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getGenerateStubsMethod(), getCallOptions()), request, responseObserver);
    }

    /**
     * <pre>
     * SyncStubReferences is called when module stubs have been updated. This allows the plugin to update
     * references to external modules, regardless of whether they are dependencies.
     * For example, go plugin adds references to all modules into the go.work file so that tools can automatically
     * import the modules when users start reference them.
     * It is optional to do anything with this call.
     * </pre>
     */
    public void syncStubReferences(xyz.block.ftl.language.v1.SyncStubReferencesRequest request,
        io.grpc.stub.StreamObserver<xyz.block.ftl.language.v1.SyncStubReferencesResponse> responseObserver) {
      io.grpc.stub.ClientCalls.asyncUnaryCall(
          getChannel().newCall(getSyncStubReferencesMethod(), getCallOptions()), request, responseObserver);
    }
  }

  /**
   * A stub to allow clients to do synchronous rpc calls to service LanguageService.
   * <pre>
   * LanguageService allows a plugin to add support for a programming language.
   * </pre>
   */
  public static final class LanguageServiceBlockingStub
      extends io.grpc.stub.AbstractBlockingStub<LanguageServiceBlockingStub> {
    private LanguageServiceBlockingStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected LanguageServiceBlockingStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new LanguageServiceBlockingStub(channel, callOptions);
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
     * Get language specific flags that can be used to create a new module.
     * </pre>
     */
    public xyz.block.ftl.language.v1.GetCreateModuleFlagsResponse getCreateModuleFlags(xyz.block.ftl.language.v1.GetCreateModuleFlagsRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getGetCreateModuleFlagsMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Generates files for a new module with the requested name
     * </pre>
     */
    public xyz.block.ftl.language.v1.CreateModuleResponse createModule(xyz.block.ftl.language.v1.CreateModuleRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getCreateModuleMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Provide default values for ModuleConfig for values that are not configured in the ftl.toml file.
     * </pre>
     */
    public xyz.block.ftl.language.v1.ModuleConfigDefaultsResponse moduleConfigDefaults(xyz.block.ftl.language.v1.ModuleConfigDefaultsRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getModuleConfigDefaultsMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Extract dependencies for a module
     * FTL will ensure that these dependencies are built before requesting a build for this module.
     * </pre>
     */
    public xyz.block.ftl.language.v1.GetDependenciesResponse getDependencies(xyz.block.ftl.language.v1.GetDependenciesRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getGetDependenciesMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Build the module and stream back build events.
     * A BuildSuccess or BuildFailure event must be streamed back with the request's context id to indicate the
     * end of the build.
     * The request can include the option to "rebuild_automatically". In this case the plugin should watch for
     * file changes and automatically rebuild as needed as long as this build request is alive. Each automactic
     * rebuild must include the latest build context id provided by the request or subsequent BuildContextUpdated
     * calls.
     * </pre>
     */
    public java.util.Iterator<xyz.block.ftl.language.v1.BuildResponse> build(
        xyz.block.ftl.language.v1.BuildRequest request) {
      return io.grpc.stub.ClientCalls.blockingServerStreamingCall(
          getChannel(), getBuildMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * While a Build call with "rebuild_automatically" set is active, BuildContextUpdated is called whenever the
     * build context is updated.
     * Each time this call is made, the Build call must send back a corresponding BuildSuccess or BuildFailure
     * event with the updated build context id with "is_automatic_rebuild" as false.
     * If the plugin will not be able to return a BuildSuccess or BuildFailure, such as when there is no active
     * build stream, it must fail the BuildContextUpdated call.
     * </pre>
     */
    public xyz.block.ftl.language.v1.BuildContextUpdatedResponse buildContextUpdated(xyz.block.ftl.language.v1.BuildContextUpdatedRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getBuildContextUpdatedMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * Generate stubs for a module.
     * Stubs allow modules to import other module's exported interface. If a language does not need this step,
     * then it is not required to do anything in this call.
     * This call is not tied to the module that this plugin is responsible for. A plugin of each language will
     * be chosen to generate stubs for each module.
     * </pre>
     */
    public xyz.block.ftl.language.v1.GenerateStubsResponse generateStubs(xyz.block.ftl.language.v1.GenerateStubsRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getGenerateStubsMethod(), getCallOptions(), request);
    }

    /**
     * <pre>
     * SyncStubReferences is called when module stubs have been updated. This allows the plugin to update
     * references to external modules, regardless of whether they are dependencies.
     * For example, go plugin adds references to all modules into the go.work file so that tools can automatically
     * import the modules when users start reference them.
     * It is optional to do anything with this call.
     * </pre>
     */
    public xyz.block.ftl.language.v1.SyncStubReferencesResponse syncStubReferences(xyz.block.ftl.language.v1.SyncStubReferencesRequest request) {
      return io.grpc.stub.ClientCalls.blockingUnaryCall(
          getChannel(), getSyncStubReferencesMethod(), getCallOptions(), request);
    }
  }

  /**
   * A stub to allow clients to do ListenableFuture-style rpc calls to service LanguageService.
   * <pre>
   * LanguageService allows a plugin to add support for a programming language.
   * </pre>
   */
  public static final class LanguageServiceFutureStub
      extends io.grpc.stub.AbstractFutureStub<LanguageServiceFutureStub> {
    private LanguageServiceFutureStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected LanguageServiceFutureStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new LanguageServiceFutureStub(channel, callOptions);
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
     * Get language specific flags that can be used to create a new module.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<xyz.block.ftl.language.v1.GetCreateModuleFlagsResponse> getCreateModuleFlags(
        xyz.block.ftl.language.v1.GetCreateModuleFlagsRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getGetCreateModuleFlagsMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * Generates files for a new module with the requested name
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<xyz.block.ftl.language.v1.CreateModuleResponse> createModule(
        xyz.block.ftl.language.v1.CreateModuleRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getCreateModuleMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * Provide default values for ModuleConfig for values that are not configured in the ftl.toml file.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<xyz.block.ftl.language.v1.ModuleConfigDefaultsResponse> moduleConfigDefaults(
        xyz.block.ftl.language.v1.ModuleConfigDefaultsRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getModuleConfigDefaultsMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * Extract dependencies for a module
     * FTL will ensure that these dependencies are built before requesting a build for this module.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<xyz.block.ftl.language.v1.GetDependenciesResponse> getDependencies(
        xyz.block.ftl.language.v1.GetDependenciesRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getGetDependenciesMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * While a Build call with "rebuild_automatically" set is active, BuildContextUpdated is called whenever the
     * build context is updated.
     * Each time this call is made, the Build call must send back a corresponding BuildSuccess or BuildFailure
     * event with the updated build context id with "is_automatic_rebuild" as false.
     * If the plugin will not be able to return a BuildSuccess or BuildFailure, such as when there is no active
     * build stream, it must fail the BuildContextUpdated call.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<xyz.block.ftl.language.v1.BuildContextUpdatedResponse> buildContextUpdated(
        xyz.block.ftl.language.v1.BuildContextUpdatedRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getBuildContextUpdatedMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * Generate stubs for a module.
     * Stubs allow modules to import other module's exported interface. If a language does not need this step,
     * then it is not required to do anything in this call.
     * This call is not tied to the module that this plugin is responsible for. A plugin of each language will
     * be chosen to generate stubs for each module.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<xyz.block.ftl.language.v1.GenerateStubsResponse> generateStubs(
        xyz.block.ftl.language.v1.GenerateStubsRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getGenerateStubsMethod(), getCallOptions()), request);
    }

    /**
     * <pre>
     * SyncStubReferences is called when module stubs have been updated. This allows the plugin to update
     * references to external modules, regardless of whether they are dependencies.
     * For example, go plugin adds references to all modules into the go.work file so that tools can automatically
     * import the modules when users start reference them.
     * It is optional to do anything with this call.
     * </pre>
     */
    public com.google.common.util.concurrent.ListenableFuture<xyz.block.ftl.language.v1.SyncStubReferencesResponse> syncStubReferences(
        xyz.block.ftl.language.v1.SyncStubReferencesRequest request) {
      return io.grpc.stub.ClientCalls.futureUnaryCall(
          getChannel().newCall(getSyncStubReferencesMethod(), getCallOptions()), request);
    }
  }

  private static final int METHODID_PING = 0;
  private static final int METHODID_GET_CREATE_MODULE_FLAGS = 1;
  private static final int METHODID_CREATE_MODULE = 2;
  private static final int METHODID_MODULE_CONFIG_DEFAULTS = 3;
  private static final int METHODID_GET_DEPENDENCIES = 4;
  private static final int METHODID_BUILD = 5;
  private static final int METHODID_BUILD_CONTEXT_UPDATED = 6;
  private static final int METHODID_GENERATE_STUBS = 7;
  private static final int METHODID_SYNC_STUB_REFERENCES = 8;

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
        case METHODID_GET_CREATE_MODULE_FLAGS:
          serviceImpl.getCreateModuleFlags((xyz.block.ftl.language.v1.GetCreateModuleFlagsRequest) request,
              (io.grpc.stub.StreamObserver<xyz.block.ftl.language.v1.GetCreateModuleFlagsResponse>) responseObserver);
          break;
        case METHODID_CREATE_MODULE:
          serviceImpl.createModule((xyz.block.ftl.language.v1.CreateModuleRequest) request,
              (io.grpc.stub.StreamObserver<xyz.block.ftl.language.v1.CreateModuleResponse>) responseObserver);
          break;
        case METHODID_MODULE_CONFIG_DEFAULTS:
          serviceImpl.moduleConfigDefaults((xyz.block.ftl.language.v1.ModuleConfigDefaultsRequest) request,
              (io.grpc.stub.StreamObserver<xyz.block.ftl.language.v1.ModuleConfigDefaultsResponse>) responseObserver);
          break;
        case METHODID_GET_DEPENDENCIES:
          serviceImpl.getDependencies((xyz.block.ftl.language.v1.GetDependenciesRequest) request,
              (io.grpc.stub.StreamObserver<xyz.block.ftl.language.v1.GetDependenciesResponse>) responseObserver);
          break;
        case METHODID_BUILD:
          serviceImpl.build((xyz.block.ftl.language.v1.BuildRequest) request,
              (io.grpc.stub.StreamObserver<xyz.block.ftl.language.v1.BuildResponse>) responseObserver);
          break;
        case METHODID_BUILD_CONTEXT_UPDATED:
          serviceImpl.buildContextUpdated((xyz.block.ftl.language.v1.BuildContextUpdatedRequest) request,
              (io.grpc.stub.StreamObserver<xyz.block.ftl.language.v1.BuildContextUpdatedResponse>) responseObserver);
          break;
        case METHODID_GENERATE_STUBS:
          serviceImpl.generateStubs((xyz.block.ftl.language.v1.GenerateStubsRequest) request,
              (io.grpc.stub.StreamObserver<xyz.block.ftl.language.v1.GenerateStubsResponse>) responseObserver);
          break;
        case METHODID_SYNC_STUB_REFERENCES:
          serviceImpl.syncStubReferences((xyz.block.ftl.language.v1.SyncStubReferencesRequest) request,
              (io.grpc.stub.StreamObserver<xyz.block.ftl.language.v1.SyncStubReferencesResponse>) responseObserver);
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
          getGetCreateModuleFlagsMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              xyz.block.ftl.language.v1.GetCreateModuleFlagsRequest,
              xyz.block.ftl.language.v1.GetCreateModuleFlagsResponse>(
                service, METHODID_GET_CREATE_MODULE_FLAGS)))
        .addMethod(
          getCreateModuleMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              xyz.block.ftl.language.v1.CreateModuleRequest,
              xyz.block.ftl.language.v1.CreateModuleResponse>(
                service, METHODID_CREATE_MODULE)))
        .addMethod(
          getModuleConfigDefaultsMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              xyz.block.ftl.language.v1.ModuleConfigDefaultsRequest,
              xyz.block.ftl.language.v1.ModuleConfigDefaultsResponse>(
                service, METHODID_MODULE_CONFIG_DEFAULTS)))
        .addMethod(
          getGetDependenciesMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              xyz.block.ftl.language.v1.GetDependenciesRequest,
              xyz.block.ftl.language.v1.GetDependenciesResponse>(
                service, METHODID_GET_DEPENDENCIES)))
        .addMethod(
          getBuildMethod(),
          io.grpc.stub.ServerCalls.asyncServerStreamingCall(
            new MethodHandlers<
              xyz.block.ftl.language.v1.BuildRequest,
              xyz.block.ftl.language.v1.BuildResponse>(
                service, METHODID_BUILD)))
        .addMethod(
          getBuildContextUpdatedMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              xyz.block.ftl.language.v1.BuildContextUpdatedRequest,
              xyz.block.ftl.language.v1.BuildContextUpdatedResponse>(
                service, METHODID_BUILD_CONTEXT_UPDATED)))
        .addMethod(
          getGenerateStubsMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              xyz.block.ftl.language.v1.GenerateStubsRequest,
              xyz.block.ftl.language.v1.GenerateStubsResponse>(
                service, METHODID_GENERATE_STUBS)))
        .addMethod(
          getSyncStubReferencesMethod(),
          io.grpc.stub.ServerCalls.asyncUnaryCall(
            new MethodHandlers<
              xyz.block.ftl.language.v1.SyncStubReferencesRequest,
              xyz.block.ftl.language.v1.SyncStubReferencesResponse>(
                service, METHODID_SYNC_STUB_REFERENCES)))
        .build();
  }

  private static abstract class LanguageServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoFileDescriptorSupplier, io.grpc.protobuf.ProtoServiceDescriptorSupplier {
    LanguageServiceBaseDescriptorSupplier() {}

    @java.lang.Override
    public com.google.protobuf.Descriptors.FileDescriptor getFileDescriptor() {
      return xyz.block.ftl.language.v1.Language.getDescriptor();
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.ServiceDescriptor getServiceDescriptor() {
      return getFileDescriptor().findServiceByName("LanguageService");
    }
  }

  private static final class LanguageServiceFileDescriptorSupplier
      extends LanguageServiceBaseDescriptorSupplier {
    LanguageServiceFileDescriptorSupplier() {}
  }

  private static final class LanguageServiceMethodDescriptorSupplier
      extends LanguageServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoMethodDescriptorSupplier {
    private final java.lang.String methodName;

    LanguageServiceMethodDescriptorSupplier(java.lang.String methodName) {
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
      synchronized (LanguageServiceGrpc.class) {
        result = serviceDescriptor;
        if (result == null) {
          serviceDescriptor = result = io.grpc.ServiceDescriptor.newBuilder(SERVICE_NAME)
              .setSchemaDescriptor(new LanguageServiceFileDescriptorSupplier())
              .addMethod(getPingMethod())
              .addMethod(getGetCreateModuleFlagsMethod())
              .addMethod(getCreateModuleMethod())
              .addMethod(getModuleConfigDefaultsMethod())
              .addMethod(getGetDependenciesMethod())
              .addMethod(getBuildMethod())
              .addMethod(getBuildContextUpdatedMethod())
              .addMethod(getGenerateStubsMethod())
              .addMethod(getSyncStubReferencesMethod())
              .build();
        }
      }
    }
    return result;
  }
}
