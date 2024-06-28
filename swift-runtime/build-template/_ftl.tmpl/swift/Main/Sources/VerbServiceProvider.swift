import Foundation

import FTL

import GRPC
import NIOCore
import NIOPosix

struct VerbServiceProviderError: Error {
   let message: String
}

protocol Handler {
   associatedtype Req: FTLType
   associatedtype Resp: FTLType
   
   var name: String { get }
   func execute(_ context:FTL.Context, request:Req) async throws -> (Resp)
   
   var requestType: any FTLType.Type { get }
   var responseType: any FTLType.Type { get }
}

func handlerFor<Req, Resp>(name:String, _ function: @escaping Verb<Req, Resp>) -> any Handler {
   return VerbHandler<Req, Resp>(name: name, handler: function)
}

func handlerFor<Resp>(name:String, _ function: @escaping Source<Resp>) -> any Handler {
   return SourceHandler<Resp>(name: name, handler: function)
}

func handlerFor<Req>(name:String, _ function: @escaping Sink<Req>) -> any Handler {
   return SinkHandler<Req>(name: name, handler: function)
}

func handlerFor(name:String, _ function: @escaping Empty) -> any Handler {
   return EmptyHandler(name: name, handler: function)
}

struct VerbHandler<_Req:FTLType, _Resp:FTLType>: Handler {
   typealias Req = _Req
   typealias Resp = _Resp
   
   let name: String
   let handler: Verb<Req, Resp>
   
   func execute(_ context:FTL.Context, request:Req) async throws -> (Resp) {
      return try await self.handler(context, request)
   }
   
   var requestType: any FTLType.Type {
      return _Req.self
   }
   
   var responseType: any FTLType.Type {
      return _Resp.self
   }
}

struct SourceHandler<_Resp:FTLType>: Handler {
   typealias Req = FTL.Unit
   typealias Resp = _Resp
   
   let name: String
   let handler: Source<Resp>
   
   func execute(_ context:FTL.Context, request:Req) async throws -> (Resp) {
      return try await self.handler(context)
   }
   
   var requestType: any FTLType.Type {
      return FTL.Unit.self
   }
   
   var responseType: any FTLType.Type {
      return _Resp.self
   }
}

struct SinkHandler<_Req:FTLType>: Handler {
   typealias Req = _Req
   typealias Resp = FTL.Unit
   
   let name: String
   let handler: Sink<Req>
   
   func execute(_ context:FTL.Context, request:Req) async throws -> (Resp) {
      try await self.handler(context, request)
      return Unit()
   }
   
   var requestType: any FTLType.Type {
      return _Req.self
   }
   
   var responseType: any FTLType.Type {
      return FTL.Unit.self
   }
}

struct EmptyHandler: Handler {
   typealias Req = FTL.Unit
   typealias Resp = FTL.Unit
   
   let name: String
   let handler: Empty
   
   func execute(_ context:FTL.Context, request:Req) async throws -> (Resp) {
      try await self.handler(context)
      return Unit()
   }
   
   var requestType: any FTLType.Type {
      return FTL.Unit.self
   }
   
   var responseType: any FTLType.Type {
      return FTL.Unit.self
   }
}

final class VerbServiceProvider: Xyz_Block_Ftl_V1_VerbServiceAsyncProvider {
   private let context: FTL.Context
   private let handlers: [String:any Handler]
   
   init(_ context:FTL.Context, handlers: [any Handler]) {
      var mapping = [String: any Handler]()
      for handler in handlers {
         mapping[handler.name] = handler
      }
      self.handlers = mapping
      self.context = context
   }
   
   /// Ping service for readiness.
   func ping(
      request: Xyz_Block_Ftl_V1_PingRequest,
      context: GRPCAsyncServerCallContext
   ) async throws -> Xyz_Block_Ftl_V1_PingResponse {
      return Xyz_Block_Ftl_V1_PingResponse()
   }
   
   /// Get configuration state for the module
   func getModuleContext(
      request: Xyz_Block_Ftl_V1_ModuleContextRequest,
      responseStream: GRPCAsyncResponseStreamWriter<Xyz_Block_Ftl_V1_ModuleContextResponse>,
      context: GRPCAsyncServerCallContext
   ) async throws {
      fatalError("Not implemented")
   }
   
   /// Acquire (and renew) a lease for a deployment.
   ///
   /// Returns ResourceExhausted if the lease is held.
   func acquireLease(
      requestStream: GRPCAsyncRequestStream<Xyz_Block_Ftl_V1_AcquireLeaseRequest>,
      responseStream: GRPCAsyncResponseStreamWriter<Xyz_Block_Ftl_V1_AcquireLeaseResponse>,
      context: GRPCAsyncServerCallContext
   ) async throws {
      fatalError("Not implemented")
   }
   
   /// Send an event to an FSM.
   func sendFSMEvent(
      request: Xyz_Block_Ftl_V1_SendFSMEventRequest,
      context: GRPCAsyncServerCallContext
   ) async throws -> Xyz_Block_Ftl_V1_SendFSMEventResponse {
      fatalError("Not implemented")
   }
   
   /// Publish an event to a topic.
   func publishEvent(
      request: Xyz_Block_Ftl_V1_PublishEventRequest,
      context: GRPCAsyncServerCallContext
   ) async throws -> Xyz_Block_Ftl_V1_PublishEventResponse {
      fatalError("Not implemented")
   }
   
   /// Issue a synchronous call to a Verb.
   func call(
      request: Xyz_Block_Ftl_V1_CallRequest,
      context: GRPCAsyncServerCallContext
   ) async throws -> Xyz_Block_Ftl_V1_CallResponse {
      guard let handler = self.handlers[request.verb.name] else {
         throw VerbServiceProviderError(message: "no handler found for \(request.verb.name)")
      }
      let responseData = try await executeCall(handler, requestData: request.body)
      var response = Xyz_Block_Ftl_V1_CallResponse()
      response.body = responseData
      return response
   }
   
   func executeCall<H: Handler>(_ handler:H, requestData:Data) async throws -> Data {
      let root = try JSONSerialization.jsonObject(with: requestData, options: [])
      do {
         let request = try handler.requestType.ftlDecode(root)
         let response = try await handler.execute(self.context, request:request as! H.Req)
         guard let responseRoot = response.ftlEncode() else {
            throw VerbServiceProviderError(message: "expected non-nil response body")
         }
         return try JSONSerialization.data(withJSONObject: responseRoot)
      }
      catch {
         return try JSONSerialization.data(withJSONObject: ["error": "\(error)"], options: [])
      }
   }
}
