public protocol FTLType {
   static func ftlDecode(_ json:Any?) throws -> Self
   func ftlEncode() -> Any?
}

// Types of calls
public typealias Verb<Req:FTLType, Resp:FTLType> = (Context, Req) async throws -> (Resp)
public typealias Source<Resp:FTLType> = (Context) async throws -> (Resp)
public typealias Sink<Req:FTLType> = (Context, Req) async throws -> ()
public typealias Empty = (Context) async throws -> ()

public class Context {
   let client:FTLClient
   
   public init(client:FTLClient) {
      self.client = client
   }
   
   public func call<Req:FTLType, Resp:FTLType>(module:String,
                                               name:String,
                                               verb:Verb<Req, Resp>,
                                               request:Req) async throws -> Resp {
      return try await self.client.call(module:module, verb:name, request: request)
   }
   
   public func call<Resp:FTLType>(module:String, name:String, source:Source<Resp>) async throws -> Resp {
      return try await self.call<Unit, Resp>(module:module, name:name, verb:{ ftl, _ in
         return await try source(ftl)
      }, request: Unit())
   }
   
   public func call<Req:FTLType>(module:String, name:String, sink:Sink<Req>, request:Req) async throws {
      let _ = try await self.call(module:module, name:name, verb:{ ftl, req in
         await try sink(ftl, req)
         return Unit()
      }, request: request)
   }
   
   public func call(module:String, name:String, empty:Empty) async throws {
      let _ = try await self.call(module:module, name:name, verb:{ ftl, _ in
         await try empty(ftl)
         return Unit()
      }, request: Unit())
   }
}

public protocol FTLClient {
   func call<Req:FTLType, Resp:FTLType>(module:String, verb:String, request:Req) async throws -> Resp
}
