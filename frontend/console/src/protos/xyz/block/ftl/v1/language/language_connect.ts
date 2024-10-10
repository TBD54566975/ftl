// @generated by protoc-gen-connect-es v1.5.0 with parameter "target=ts"
// @generated from file xyz/block/ftl/v1/language/language.proto (package xyz.block.ftl.v1.language, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import { PingRequest, PingResponse } from "../ftl_pb.js";
import { MethodIdempotency, MethodKind } from "@bufbuild/protobuf";
import { BuildContextUpdatedRequest, BuildContextUpdatedResponse, BuildEvent, BuildRequest, CreateModuleRequest, CreateModuleResponse, DependenciesRequest, DependenciesResponse, GetCreateModuleFlagsRequest, GetCreateModuleFlagsResponse, ModuleConfigDefaultsRequest, ModuleConfigDefaultsResponse } from "./language_pb.js";

/**
 * LanguageService allows a plugin to add support for a programming language.
 *
 * @generated from service xyz.block.ftl.v1.language.LanguageService
 */
export const LanguageService = {
  typeName: "xyz.block.ftl.v1.language.LanguageService",
  methods: {
    /**
     * Ping service for readiness.
     *
     * @generated from rpc xyz.block.ftl.v1.language.LanguageService.Ping
     */
    ping: {
      name: "Ping",
      I: PingRequest,
      O: PingResponse,
      kind: MethodKind.Unary,
      idempotency: MethodIdempotency.NoSideEffects,
    },
    /**
     * Get language specific flags that can be used to create a new module.
     *
     * @generated from rpc xyz.block.ftl.v1.language.LanguageService.GetCreateModuleFlags
     */
    getCreateModuleFlags: {
      name: "GetCreateModuleFlags",
      I: GetCreateModuleFlagsRequest,
      O: GetCreateModuleFlagsResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Generates files for a new module with the requested name
     *
     * @generated from rpc xyz.block.ftl.v1.language.LanguageService.CreateModule
     */
    createModule: {
      name: "CreateModule",
      I: CreateModuleRequest,
      O: CreateModuleResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Provide default values for ModuleConfig for values that are not configured in the ftl.toml file.
     *
     * @generated from rpc xyz.block.ftl.v1.language.LanguageService.ModuleConfigDefaults
     */
    moduleConfigDefaults: {
      name: "ModuleConfigDefaults",
      I: ModuleConfigDefaultsRequest,
      O: ModuleConfigDefaultsResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Extract dependencies for a module
     * FTL will ensure that these dependencies are built before requesting a build for this module.
     *
     * @generated from rpc xyz.block.ftl.v1.language.LanguageService.GetDependencies
     */
    getDependencies: {
      name: "GetDependencies",
      I: DependenciesRequest,
      O: DependenciesResponse,
      kind: MethodKind.Unary,
    },
    /**
     * Build the module and stream back build events.
     *
     * A BuildSuccess or BuildFailure event must be streamed back with the request's context id to indicate the
     * end of the build.
     *
     * The request can include the option to "rebuild_automatically". In this case the plugin should watch for
     * file changes and automatically rebuild as needed as long as this build request is alive. Each automactic
     * rebuild must include the latest build context id provided by the request or subsequent BuildContextUpdated
     * calls.
     *
     * @generated from rpc xyz.block.ftl.v1.language.LanguageService.Build
     */
    build: {
      name: "Build",
      I: BuildRequest,
      O: BuildEvent,
      kind: MethodKind.ServerStreaming,
    },
    /**
     * While a Build call with "rebuild_automatically" set is active, BuildContextUpdated is called whenever the
     * build context is updated.
     *
     * Each time this call is made, the Build call must send back a corresponding BuildSuccess or BuildFailure
     * event with the updated build context id with "is_automatic_rebuild" as false.
     *
     * @generated from rpc xyz.block.ftl.v1.language.LanguageService.BuildContextUpdated
     */
    buildContextUpdated: {
      name: "BuildContextUpdated",
      I: BuildContextUpdatedRequest,
      O: BuildContextUpdatedResponse,
      kind: MethodKind.Unary,
    },
  }
} as const;

