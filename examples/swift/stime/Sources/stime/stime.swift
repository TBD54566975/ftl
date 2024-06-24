import Foundation
import FTL

@FTLData()
public struct TimeRequest {
    
}

@FTLData()
public struct TimeResponse {
    let time: Date
}

// Time returns the current time.
//
@FTLVerb()
func time(req: TimeRequest) throws -> TimeResponse {
    return TimeResponse(time: Date())
}

