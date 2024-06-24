import Foundation

public struct Module {
    //    Pos Position `parser:"" protobuf:"1,optional"`
    public var isBuiltIn = false
    
    public let name: String
    public var decls = [Decl]()
    public var comments = [String]()
    
    public init(name:String) {
        self.name = name
    }
    
    public func serializedBytes() throws -> Data {
        var module = Xyz_Block_Ftl_V1_Schema_Module()
        module.name = self.name
        module.decls = self.decls.map { $0.toProto() }
        module.comments = self.comments
        
        return try module.serializedData()
    }
    
     
}
