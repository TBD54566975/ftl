import Foundation

public struct Ref {
    public let module: String
    public let name: String
    
    public init(module: String, name: String) {
        self.module = module
        self.name = name
    }
}
