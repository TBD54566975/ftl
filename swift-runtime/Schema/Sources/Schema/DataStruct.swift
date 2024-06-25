import Foundation

public struct DataStruct {
    public let name: String
    public var isExported: Bool
    public var fields = [(name:String, type:FTLType)]()
    
    public init(name: String, isExported: Bool, fields: [(name: String, type: FTLType)] = [(name:String, type:FTLType)]()) {
        self.name = name
        self.isExported = isExported
        self.fields = fields
    }
    
    // TODO: associated types
    
    func toProto() -> Xyz_Block_Ftl_V1_Schema_Data {
        var proto = Xyz_Block_Ftl_V1_Schema_Data()
        proto.name = self.name
        proto.export = self.isExported
        for field in fields {
            var protoField = Xyz_Block_Ftl_V1_Schema_Field()
            protoField.name = field.name
            protoField.type = field.type.toProto()
            // TODO: field comments
            proto.fields.append(protoField)
        }
        return proto
    }
}
