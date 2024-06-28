import Foundation
import FTL

@FTLData()
public struct TimeRequest {
   
}

@FTLData()
public struct TimeResponse {
   let time: Date
   let name: FTL.Unit
}

// Time returns the current time.
@FTLVerb()
public func time(_ context:FTL.Context, req: TimeRequest) async throws -> TimeResponse {
   // call another verb
   let resp = await try #callVerb(innerVerb, request: TimeRequest())
   
   // call a source
   _ = await try #callSource(source)
   
   // call a sink
   await try #callSink(sink, request: TimeRequest())
   
   // call an empty
   await try #callEmpty(empty)
   
   return resp
}

@FTLVerb()
public func innerVerb(_ context:FTL.Context, req: TimeRequest) async throws -> TimeResponse {
   return TimeResponse(time: Date(), name:FTL.Unit())
}

@FTLVerb()
public func source(_ context:FTL.Context) async throws -> TimeResponse {
   return TimeResponse(time: Date(), name:FTL.Unit())
}

@FTLVerb()
public func sink(_ context:FTL.Context, req: TimeRequest) async throws {
   
}

@FTLVerb()
public func empty(_ context:FTL.Context) async throws {
   
}
