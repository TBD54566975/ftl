import Foundation

import FTL

import GRPC
import NIOCore
import NIOPosix

struct VerbServiceProviderError: Error {
   let message: String
}

typealias Verb<Req, Resp> = (Req) throws -> (Resp)
typealias Source<Resp> = () throws -> (Resp)
typealias Sink<Req> = (Req) throws -> ()
typealias Empty = () throws -> ()

protocol Handler {
   associatedtype Req: FTLType
   associatedtype Resp: FTLType
   
   var name: String { get }
   func execute(_ req:Req) throws -> (Resp)
}

func handlerFor<Req, Resp>(name:String, _ function: @escaping Verb<Req, Resp>) throws -> any Handler {
   return VerbHandler(name: name, handler: function)
}

func handlerFor<Resp>(name:String, _ function: @escaping Source<Resp>) throws -> any Handler {
   return SourceHandler(name: name, handler: function)
}

func handlerFor<Req>(name:String, _ function: @escaping Sink<Req>) throws -> any Handler {
   return SinkHandler(name: name, handler: function)
}

func handlerFor(name:String, _ function: @escaping Empty) throws -> any Handler {
   return EmptyHandler(name: name, handler: function)
}

struct VerbHandler<_Req, _Resp>: Handler {
   typealias Req = _Req
   typealias Resp = _Resp
   
   let name: String
   let handler: Verb<Req, Resp>
   
   func execute(_ req:Req) throws -> (Resp) {
      return try self.handler(req)
   }
}

struct SourceHandler<_Resp>: Handler {
   typealias Req = Any
   typealias Resp = _Resp
   
   let name: String
   let handler: Source<Resp>
   
   func execute(_ req:Req) throws -> (Resp) {
      return try self.handler()
   }
}

struct SinkHandler<_Req>: Handler {
   typealias Req = _Req
   typealias Resp = Any
   
   let name: String
   let handler: Sink<Req>
   
   func execute(_ req:Req) throws -> (Resp) {
      try self.handler(req)
      fatalError("Return unit")
   }
}

struct EmptyHandler: Handler {
   typealias Req = Any
   typealias Resp = Any
   
   let name: String
   let handler: Empty
   
   func execute(_ req:Req) throws -> (Resp) {
      try self.handler()
      fatalError("Return unit")
   }
}

final class VerbServiceProvider: Xyz_Block_Ftl_V1_VerbServiceAsyncProvider {
   private let handlers: [String:any Handler]
   
   init(handlers: [any Handler]) {
      var mapping = [String: any Handler]()
      for handler in handlers {
         mapping[handler.name] = handler
      }
      self.handlers = mapping
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
      request.body
      //        let resp = try handler.execute("need to parse req")
      fatalError("Not implemented: calling handler")
   }
}
