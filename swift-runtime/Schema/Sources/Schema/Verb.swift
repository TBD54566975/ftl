import Foundation

public struct Verb {
    public let name: String
    public let isExported: Bool
    public let requestType: FTLType
    public let responseType: FTLType
    
    public init(name: String, isExported: Bool, requestType: FTLType, responseType: FTLType) {
        self.name = name
        self.isExported = isExported
        self.requestType = requestType
        self.responseType = responseType
    }
    
    func toProto() -> Xyz_Block_Ftl_V1_Schema_Verb {
        var proto = Xyz_Block_Ftl_V1_Schema_Verb()
        proto.name = self.name
        proto.export = self.isExported
        proto.request = self.requestType.toProto()
        proto.response = self.responseType.toProto()
        return proto
    }
}
