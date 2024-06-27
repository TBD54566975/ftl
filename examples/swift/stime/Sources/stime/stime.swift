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
public func time(req: TimeRequest) throws -> TimeResponse {
   return TimeResponse(time: Date(), name:FTL.Unit())
}
